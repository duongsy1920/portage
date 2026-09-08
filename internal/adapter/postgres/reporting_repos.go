package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	reportingapp "github.com/duongsy/portage/internal/app/reporting"
	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The read model on Postgres (0007_reporting.sql).
//
// Every other repository here rebuilds an AGGREGATE through FromSnapshot, so a
// row that no longer satisfies the domain's rules fails loudly. This one does
// not, and the difference is the point: OrderSummary has no invariants to
// check. A half-filled row is a normal state of a projection — the frame an
// event wrote before order_placed arrived — so the scan fills a plain struct
// and every money column is nullable.

type OrderSummaryRepo struct {
	pool *pgxpool.Pool
}

var _ reportingapp.OrderSummaryRepository = (*OrderSummaryRepo)(nil)

func NewOrderSummaryRepo(pool *pgxpool.Pool) *OrderSummaryRepo {
	return &OrderSummaryRepo{pool: pool}
}

const summaryColumns = `"order", customer, product, variant, quote, product_name, status, tracking,
	total_minor, deposit_minor, refund_minor, currency,
	deposit_paid, balance_paid, forfeited, shop_reference,
	placed_at, delivered_at, cancelled_at, updated_at`

func (r *OrderSummaryRepo) Save(ctx context.Context, s reportingapp.OrderSummary) error {
	var total, deposit, refund *int64
	var currency *string
	if s.Total.IsValid() {
		v := s.Total.Minor()
		total = &v
		c := s.Total.Currency().Code()
		currency = &c
	}
	if s.Deposit.IsValid() {
		v := s.Deposit.Minor()
		deposit = &v
	}
	if s.Refund.IsValid() {
		v := s.Refund.Minor()
		refund = &v
	}
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO order_summaries (`+summaryColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		ON CONFLICT ("order") DO UPDATE SET
			customer = EXCLUDED.customer, product = EXCLUDED.product, variant = EXCLUDED.variant, quote = EXCLUDED.quote,
			product_name = EXCLUDED.product_name, status = EXCLUDED.status, tracking = EXCLUDED.tracking,
			total_minor = EXCLUDED.total_minor, deposit_minor = EXCLUDED.deposit_minor,
			refund_minor = EXCLUDED.refund_minor, currency = EXCLUDED.currency,
			deposit_paid = EXCLUDED.deposit_paid, balance_paid = EXCLUDED.balance_paid, forfeited = EXCLUDED.forfeited,
			shop_reference = EXCLUDED.shop_reference,
			placed_at = EXCLUDED.placed_at, delivered_at = EXCLUDED.delivered_at,
			cancelled_at = EXCLUDED.cancelled_at, updated_at = EXCLUDED.updated_at`,
		s.Order.String(), idOrNil(s.Customer), idOrNil(s.Product), idOrNil(s.Variant), idOrNil(s.Quote),
		s.ProductName, string(s.Status), string(s.Tracking),
		total, deposit, refund, currency,
		s.DepositPaid, s.BalancePaid, s.Forfeited, s.ShopReference,
		timeOrNil(s.PlacedAt), timeOrNil(s.DeliveredAt), timeOrNil(s.CancelledAt), s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save order summary %s: %w", s.Order, err)
	}
	return nil
}

func (r *OrderSummaryRepo) ByOrder(ctx context.Context, id ordering.OrderID) (reportingapp.OrderSummary, error) {
	s, err := scanSummary(db(ctx, r.pool).QueryRow(ctx, `SELECT `+summaryColumns+` FROM order_summaries WHERE "order" = $1`, id.String()))
	if errors.Is(err, pgx.ErrNoRows) {
		return reportingapp.OrderSummary{}, fmt.Errorf("order summary %s: %w", id, reportingapp.ErrSummaryNotFound)
	}
	if err != nil {
		return reportingapp.OrderSummary{}, fmt.Errorf("order summary %s: %w", id, err)
	}
	return s, nil
}

// The lists are ordered newest first — what a "my orders" screen shows — and
// the two indexes in 0007 are (customer, placed_at DESC) and (status, …) so
// each of these is one index scan, no sort.
func (r *OrderSummaryRepo) ByCustomer(ctx context.Context, customer shared.ID) ([]reportingapp.OrderSummary, error) {
	return r.list(ctx, `WHERE customer = $1`, customer.String())
}

func (r *OrderSummaryRepo) ByStatus(ctx context.Context, status ordering.OrderStatus) ([]reportingapp.OrderSummary, error) {
	return r.list(ctx, `WHERE status = $1`, string(status))
}

func (r *OrderSummaryRepo) ByProduct(ctx context.Context, product shared.ID) ([]reportingapp.OrderSummary, error) {
	return r.list(ctx, `WHERE product = $1`, product.String())
}

func (r *OrderSummaryRepo) All(ctx context.Context) ([]reportingapp.OrderSummary, error) {
	return r.list(ctx, ``)
}

func (r *OrderSummaryRepo) list(ctx context.Context, where string, args ...any) ([]reportingapp.OrderSummary, error) {
	rows, err := db(ctx, r.pool).Query(ctx,
		`SELECT `+summaryColumns+` FROM order_summaries `+where+` ORDER BY placed_at DESC NULLS LAST, "order"`, args...)
	if err != nil {
		return nil, fmt.Errorf("list order summaries: %w", err)
	}
	defer rows.Close()

	var out []reportingapp.OrderSummary
	for rows.Next() {
		s, err := scanSummary(rows)
		if err != nil {
			return nil, fmt.Errorf("list order summaries: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list order summaries: %w", err)
	}
	return out, nil
}

func scanSummary(sc scanner) (reportingapp.OrderSummary, error) {
	var (
		rawOrder                            string
		customer, product, variant, quote   *string
		name, status, tracking, shopRef     string
		total, deposit, refund              *int64
		currency                            *string
		depositPaid, balancePaid, forfeited bool
		placed, delivered, cancelled        *time.Time
		updated                             time.Time
	)
	if err := sc.Scan(&rawOrder, &customer, &product, &variant, &quote, &name, &status, &tracking,
		&total, &deposit, &refund, &currency,
		&depositPaid, &balancePaid, &forfeited, &shopRef,
		&placed, &delivered, &cancelled, &updated); err != nil {
		return reportingapp.OrderSummary{}, err
	}

	id, err := ordering.ParseOrderID(rawOrder)
	if err != nil {
		return reportingapp.OrderSummary{}, corrupt("order_summaries", rawOrder, err)
	}
	s := reportingapp.OrderSummary{
		Order:       id,
		ProductName: name,
		Status:      ordering.OrderStatus(status),
		Tracking:    reportingapp.Tracking(tracking),
		DepositPaid: depositPaid, BalancePaid: balancePaid, Forfeited: forfeited,
		ShopReference: shopRef,
		PlacedAt:      timeFrom(placed), DeliveredAt: timeFrom(delivered), CancelledAt: timeFrom(cancelled),
		UpdatedAt: updated.UTC(),
	}
	for _, f := range []struct {
		raw  *string
		into *shared.ID
	}{{customer, &s.Customer}, {product, &s.Product}, {variant, &s.Variant}, {quote, &s.Quote}} {
		if f.raw == nil {
			continue
		}
		v, err := shared.ParseID(*f.raw)
		if err != nil {
			return reportingapp.OrderSummary{}, corrupt("order_summaries", rawOrder, err)
		}
		*f.into = v
	}
	if currency != nil {
		cur, err := shared.CurrencyFromCode(*currency)
		if err != nil {
			return reportingapp.OrderSummary{}, corrupt("order_summaries", rawOrder, err)
		}
		for _, f := range []struct {
			raw  *int64
			into *shared.Money
		}{{total, &s.Total}, {deposit, &s.Deposit}, {refund, &s.Refund}} {
			if f.raw != nil {
				*f.into = shared.NewMoney(*f.raw, cur)
			}
		}
	}
	return s, nil
}

// ── product names ────────────────────────────────────────────────────────────

type ProductNameRepo struct {
	pool *pgxpool.Pool
}

var _ reportingapp.ProductNames = (*ProductNameRepo)(nil)

func NewProductNameRepo(pool *pgxpool.Pool) *ProductNameRepo {
	return &ProductNameRepo{pool: pool}
}

func (r *ProductNameRepo) Name(ctx context.Context, product shared.ID) (string, bool, error) {
	var name string
	err := db(ctx, r.pool).QueryRow(ctx, `SELECT name FROM product_names WHERE product = $1`, product.String()).Scan(&name)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil // not an error: the product may simply not be published yet
	}
	if err != nil {
		return "", false, fmt.Errorf("product name %s: %w", product, err)
	}
	return name, true, nil
}

func (r *ProductNameRepo) Save(ctx context.Context, product shared.ID, name string) error {
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO product_names (product, name) VALUES ($1, $2)
		ON CONFLICT (product) DO UPDATE SET name = EXCLUDED.name`, product.String(), name)
	if err != nil {
		return fmt.Errorf("save product name %s: %w", product, err)
	}
	return nil
}

// ── nil helpers ──────────────────────────────────────────────────────────────
//
// A read model's row is half-empty on purpose (the frame), so the zero value
// of an ID, a Money or a time must reach the database as NULL rather than as
// "00000000-0000-0000-0000-000000000000" — a fake id that would then come back
// as a real one.

func idOrNil(id shared.ID) *string {
	if id.IsZero() {
		return nil
	}
	s := id.String()
	return &s
}

func timeOrNil(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func timeFrom(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.UTC()
}
