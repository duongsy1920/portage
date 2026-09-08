package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/duongsy/portage/internal/domain/procurement"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The procurement ports on Postgres.

type TaskRepo struct {
	pool *pgxpool.Pool
}

var _ procurement.TaskRepository = (*TaskRepo)(nil)

func NewTaskRepo(pool *pgxpool.Pool) *TaskRepo {
	return &TaskRepo{pool: pool}
}

const taskColumns = `id, "order", product, variant, currency, status, reference, paid_minor, paid_by, reason,
	product_name, variant_label, variant_ref, source, opened_at, closed_at`

func (r *TaskRepo) Save(ctx context.Context, t *procurement.PurchaseTask) error {
	s := t.Snapshot()
	var paid *int64
	if s.Paid.IsValid() {
		v := s.Paid.Minor()
		paid = &v
	}
	var paidBy *string
	if !s.PaidBy.IsZero() {
		v := s.PaidBy.String()
		paidBy = &v
	}
	var closedAt *time.Time
	if !s.ClosedAt.IsZero() {
		v := s.ClosedAt
		closedAt = &v
	}
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO purchase_tasks (`+taskColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (id) DO UPDATE SET
			"order" = EXCLUDED."order", product = EXCLUDED.product, variant = EXCLUDED.variant, currency = EXCLUDED.currency,
			status = EXCLUDED.status, reference = EXCLUDED.reference, paid_minor = EXCLUDED.paid_minor, paid_by = EXCLUDED.paid_by,
			reason = EXCLUDED.reason, product_name = EXCLUDED.product_name, variant_label = EXCLUDED.variant_label,
			variant_ref = EXCLUDED.variant_ref, source = EXCLUDED.source,
			opened_at = EXCLUDED.opened_at, closed_at = EXCLUDED.closed_at`,
		s.ID.String(), s.Order.String(), s.Product.String(), s.Variant.String(), s.Currency.Code(), string(s.Status),
		s.Reference, paid, paidBy, s.Reason,
		s.Subject.ProductName, s.Subject.VariantLabel, s.Subject.VariantRef, s.Subject.Source,
		s.OpenedAt, closedAt)
	if err != nil {
		return fmt.Errorf("save purchase task %s: %w", s.ID, err)
	}
	return nil
}

func (r *TaskRepo) ByID(ctx context.Context, id procurement.TaskID) (*procurement.PurchaseTask, error) {
	all, err := r.query(ctx, `SELECT `+taskColumns+` FROM purchase_tasks WHERE id = $1`, id.String())
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("purchase task %s: %w", id, procurement.ErrTaskNotFound)
	}
	return all[0], nil
}

func (r *TaskRepo) ByOrder(ctx context.Context, order shared.ID) (*procurement.PurchaseTask, error) {
	all, err := r.query(ctx, `SELECT `+taskColumns+` FROM purchase_tasks WHERE "order" = $1`, order.String())
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("purchase task for order %s: %w", order, procurement.ErrTaskNotFound)
	}
	return all[0], nil
}

func (r *TaskRepo) Open(ctx context.Context) ([]*procurement.PurchaseTask, error) {
	return r.query(ctx, `SELECT `+taskColumns+` FROM purchase_tasks WHERE status = 'open' ORDER BY opened_at, id`)
}

func (r *TaskRepo) query(ctx context.Context, sql string, args ...any) ([]*procurement.PurchaseTask, error) {
	rows, err := db(ctx, r.pool).Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("purchase tasks: %w", err)
	}
	defer rows.Close()
	var out []*procurement.PurchaseTask
	for rows.Next() {
		var (
			rawID, order, product, variant, currency, status, reference, reason string
			subject                                                             procurement.Subject
			paid                                                                *int64
			paidBy                                                              *string
			openedAt                                                            time.Time
			closedAt                                                            *time.Time
		)
		if err := rows.Scan(&rawID, &order, &product, &variant, &currency, &status, &reference, &paid, &paidBy, &reason,
			&subject.ProductName, &subject.VariantLabel, &subject.VariantRef, &subject.Source, &openedAt, &closedAt); err != nil {
			return nil, fmt.Errorf("scan purchase task: %w", err)
		}
		t, err := taskFrom(rawID, order, product, variant, currency, status, reference, paid, paidBy, reason, subject, openedAt, closedAt)
		if err != nil {
			return nil, corrupt("purchase_tasks", rawID, err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func taskFrom(rawID, order, product, variant, currency, status, reference string, paid *int64, paidBy *string,
	reason string, subject procurement.Subject, openedAt time.Time, closedAt *time.Time) (*procurement.PurchaseTask, error) {
	id, err := procurement.ParseTaskID(rawID)
	if err != nil {
		return nil, err
	}
	ids := make([]shared.ID, 3)
	for i, raw := range []string{order, product, variant} {
		if ids[i], err = shared.ParseID(raw); err != nil {
			return nil, err
		}
	}
	cur, err := shared.CurrencyFromCode(currency)
	if err != nil {
		return nil, err
	}
	snap := procurement.TaskSnapshot{
		ID: id, Order: ids[0], Product: ids[1], Variant: ids[2], Currency: cur, Status: procurement.TaskStatus(status),
		Subject: subject, Reference: reference, Reason: reason, OpenedAt: openedAt.UTC(),
	}
	if paid != nil {
		snap.Paid = shared.NewMoney(*paid, cur)
	}
	if paidBy != nil {
		if snap.PaidBy, err = shared.ParseOperatorID(*paidBy); err != nil {
			return nil, err
		}
	}
	if closedAt != nil {
		snap.ClosedAt = closedAt.UTC()
	}
	return procurement.TaskFromSnapshot(snap)
}

type ShopRepo struct {
	pool *pgxpool.Pool
}

var _ procurement.ShopRepository = (*ShopRepo)(nil)

func NewShopRepo(pool *pgxpool.Pool) *ShopRepo {
	return &ShopRepo{pool: pool}
}

func (r *ShopRepo) Save(ctx context.Context, s procurement.Shop) error {
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO procurement_shops (merchant, name, site, currency) VALUES ($1, $2, $3, $4)
		ON CONFLICT (merchant) DO UPDATE SET name = EXCLUDED.name, site = EXCLUDED.site, currency = EXCLUDED.currency`,
		s.Merchant.String(), s.Name, s.Site, s.Currency.Code())
	if err != nil {
		return fmt.Errorf("save shop %s: %w", s.Merchant, err)
	}
	return nil
}

func (r *ShopRepo) ByID(ctx context.Context, merchant shared.ID) (procurement.Shop, error) {
	var raw, name, site, currency string
	err := db(ctx, r.pool).QueryRow(ctx, `SELECT merchant, name, site, currency FROM procurement_shops WHERE merchant = $1`, merchant.String()).Scan(&raw, &name, &site, &currency)
	if errors.Is(err, pgx.ErrNoRows) {
		return procurement.Shop{}, fmt.Errorf("shop %s: %w", merchant, procurement.ErrShopNotFound)
	}
	if err != nil {
		return procurement.Shop{}, fmt.Errorf("shop %s: %w", merchant, err)
	}
	id, err := shared.ParseID(raw)
	if err != nil {
		return procurement.Shop{}, corrupt("procurement_shops", raw, err)
	}
	cur, err := shared.CurrencyFromCode(currency)
	if err != nil {
		return procurement.Shop{}, corrupt("procurement_shops", raw, err)
	}
	return procurement.Shop{Merchant: id, Name: name, Site: site, Currency: cur}, nil
}

type ItemRepo struct {
	pool *pgxpool.Pool
}

var _ procurement.ItemRepository = (*ItemRepo)(nil)

func NewItemRepo(pool *pgxpool.Pool) *ItemRepo {
	return &ItemRepo{pool: pool}
}

func (r *ItemRepo) Save(ctx context.Context, i procurement.Item) error {
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO procurement_items (product, merchant, name, source) VALUES ($1, $2, $3, $4)
		ON CONFLICT (product) DO UPDATE SET merchant = EXCLUDED.merchant, name = EXCLUDED.name, source = EXCLUDED.source`,
		i.Product.String(), i.Merchant.String(), i.Name, i.Source)
	if err != nil {
		return fmt.Errorf("save item %s: %w", i.Product, err)
	}
	return nil
}

func (r *ItemRepo) ByProduct(ctx context.Context, product shared.ID) (procurement.Item, error) {
	var raw, merchant, name, source string
	err := db(ctx, r.pool).QueryRow(ctx, `SELECT product, merchant, name, source FROM procurement_items WHERE product = $1`, product.String()).
		Scan(&raw, &merchant, &name, &source)
	if errors.Is(err, pgx.ErrNoRows) {
		return procurement.Item{}, fmt.Errorf("item %s: %w", product, procurement.ErrItemNotFound)
	}
	if err != nil {
		return procurement.Item{}, fmt.Errorf("item %s: %w", product, err)
	}
	p, err := shared.ParseID(raw)
	if err != nil {
		return procurement.Item{}, corrupt("procurement_items", raw, err)
	}
	m, err := shared.ParseID(merchant)
	if err != nil {
		return procurement.Item{}, corrupt("procurement_items", raw, err)
	}
	return procurement.Item{Product: p, Merchant: m, Name: name, Source: source}, nil
}

// VariantRepo is the dictionary "this variant id means this size", kept from
// catalog.variant_added so the buyer's screen never shows a bare uuid.
type VariantRepo struct {
	pool *pgxpool.Pool
}

var _ procurement.VariantRepository = (*VariantRepo)(nil)

func NewVariantRepo(pool *pgxpool.Pool) *VariantRepo {
	return &VariantRepo{pool: pool}
}

func (r *VariantRepo) Save(ctx context.Context, v procurement.Variant) error {
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO procurement_variants (variant, product, size, color, merchant_ref)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (variant) DO UPDATE SET product = EXCLUDED.product, size = EXCLUDED.size,
			color = EXCLUDED.color, merchant_ref = EXCLUDED.merchant_ref`,
		v.Variant.String(), v.Product.String(), v.Size, v.Color, v.MerchantRef)
	if err != nil {
		return fmt.Errorf("save variant %s: %w", v.Variant, err)
	}
	return nil
}

func (r *VariantRepo) ByID(ctx context.Context, variant shared.ID) (procurement.Variant, error) {
	var raw, product, size, color, ref string
	err := db(ctx, r.pool).QueryRow(ctx, `
		SELECT variant, product, size, color, merchant_ref FROM procurement_variants WHERE variant = $1`,
		variant.String()).Scan(&raw, &product, &size, &color, &ref)
	if errors.Is(err, pgx.ErrNoRows) {
		return procurement.Variant{}, fmt.Errorf("variant %s: %w", variant, procurement.ErrVariantNotFound)
	}
	if err != nil {
		return procurement.Variant{}, fmt.Errorf("variant %s: %w", variant, err)
	}
	v, err := shared.ParseID(raw)
	if err != nil {
		return procurement.Variant{}, corrupt("procurement_variants", raw, err)
	}
	p, err := shared.ParseID(product)
	if err != nil {
		return procurement.Variant{}, corrupt("procurement_variants", raw, err)
	}
	return procurement.Variant{Variant: v, Product: p, Size: size, Color: color, MerchantRef: ref}, nil
}
