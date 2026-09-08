package postgres

import (
	"context"
	"encoding/json"
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

// ── the worklist (0009_product_worklist.sql) ─────────────────────────────────

const worklistColumns = `product, merchant, category, name, source_url,
	price_minor, price_currency, sourced_by, requested_by,
	listing_confirmed, measured, published, variants,
	added_at, updated_at`

type ProductWorklistRepo struct {
	pool *pgxpool.Pool
}

var _ reportingapp.ProductWorklistRepository = (*ProductWorklistRepo)(nil)

func NewProductWorklistRepo(pool *pgxpool.Pool) *ProductWorklistRepo {
	return &ProductWorklistRepo{pool: pool}
}

func (r *ProductWorklistRepo) ByProduct(ctx context.Context, product shared.ID) (reportingapp.WorklistItem, error) {
	row := db(ctx, r.pool).QueryRow(ctx,
		`SELECT `+worklistColumns+` FROM product_worklist WHERE product = $1`, product.String())
	w, err := scanWorklistItem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return reportingapp.WorklistItem{}, fmt.Errorf("worklist item %s: %w", product, reportingapp.ErrWorklistItemNotFound)
	}
	return w, err
}

func (r *ProductWorklistRepo) Open(ctx context.Context) ([]reportingapp.WorklistItem, error) {
	return r.listWorklist(ctx, `WHERE published = false`)
}

func (r *ProductWorklistRepo) ByRequester(ctx context.Context, customer shared.ID) ([]reportingapp.WorklistItem, error) {
	if customer.IsZero() {
		// requested_by IS NULL would match every row an operator added on
		// spec, and hand one person a list that is not theirs.
		return []reportingapp.WorklistItem{}, nil
	}
	return r.listWorklist(ctx, `WHERE requested_by = $1`, customer.String())
}

func (r *ProductWorklistRepo) Save(ctx context.Context, w reportingapp.WorklistItem) error {
	// The list goes down as jsonb. An empty list is "[]" and not NULL, so a
	// reader never has to tell "no sizes yet" from "column missing".
	rows := make([]worklistVariantRow, 0, len(w.Variants))
	for _, v := range w.Variants {
		rows = append(rows, worklistVariantRow{ID: v.ID.String(), Label: v.Label})
	}
	variants, err := json.Marshal(rows)
	if err != nil {
		return fmt.Errorf("save worklist item %s: %w", w.Product, err)
	}
	_, err = db(ctx, r.pool).Exec(ctx, `
		INSERT INTO product_worklist (`+worklistColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (product) DO UPDATE SET
			merchant = EXCLUDED.merchant, category = EXCLUDED.category, name = EXCLUDED.name,
			source_url = EXCLUDED.source_url,
			price_minor = EXCLUDED.price_minor, price_currency = EXCLUDED.price_currency,
			sourced_by = EXCLUDED.sourced_by, requested_by = EXCLUDED.requested_by,
			listing_confirmed = EXCLUDED.listing_confirmed,
			measured = EXCLUDED.measured, published = EXCLUDED.published,
			variants = EXCLUDED.variants,
			added_at = EXCLUDED.added_at, updated_at = EXCLUDED.updated_at`,
		w.Product.String(), w.Merchant.String(), w.Category, w.Name, w.Source,
		w.Price.Minor(), w.Price.Currency().Code(), w.SourcedBy, idOrNil(w.RequestedBy),
		w.ListingConfirmed, w.Measured, w.Published, variants,
		w.AddedAt, w.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save worklist item %s: %w", w.Product, err)
	}
	return nil
}

func (r *ProductWorklistRepo) listWorklist(ctx context.Context, where string, args ...any) ([]reportingapp.WorklistItem, error) {
	rows, err := db(ctx, r.pool).Query(ctx,
		`SELECT `+worklistColumns+` FROM product_worklist `+where+` ORDER BY added_at, product`, args...)
	if err != nil {
		return nil, fmt.Errorf("list worklist: %w", err)
	}
	defer rows.Close()
	out := []reportingapp.WorklistItem{}
	for rows.Next() {
		w, err := scanWorklistItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// worklistVariantRow is the jsonb shape. Named keys, written by hand, for the
// same reason event payloads are: the column IS a contract once a row exists.
type worklistVariantRow struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// scanWorklistItem fills a plain struct, like the summary scan above: a
// projection has no invariant to re-check, only ids and money to parse.
func scanWorklistItem(row pgx.Row) (reportingapp.WorklistItem, error) {
	var (
		product, merchant, category, name, source string
		priceCurrency, sourcedBy                  string
		priceMinor                                int64
		requestedBy                               *string
		confirmed, measured, published            bool
		variantsJSON                              []byte
		addedAt, updatedAt                        time.Time
	)
	if err := row.Scan(&product, &merchant, &category, &name, &source,
		&priceMinor, &priceCurrency, &sourcedBy, &requestedBy,
		&confirmed, &measured, &published, &variantsJSON,
		&addedAt, &updatedAt); err != nil {
		return reportingapp.WorklistItem{}, err
	}
	pid, err := shared.ParseID(product)
	if err != nil {
		return reportingapp.WorklistItem{}, corrupt("product_worklist", product, err)
	}
	mid, err := shared.ParseID(merchant)
	if err != nil {
		return reportingapp.WorklistItem{}, corrupt("product_worklist", product, err)
	}
	cur, err := shared.CurrencyFromCode(priceCurrency)
	if err != nil {
		return reportingapp.WorklistItem{}, corrupt("product_worklist", product, err)
	}
	w := reportingapp.WorklistItem{
		Product: pid, Merchant: mid, Category: category, Name: name, Source: source,
		Price: shared.NewMoney(priceMinor, cur), SourcedBy: sourcedBy,
		ListingConfirmed: confirmed, Measured: measured, Published: published,
		AddedAt: addedAt.UTC(), UpdatedAt: updatedAt.UTC(),
	}
	var rows []worklistVariantRow
	if len(variantsJSON) > 0 {
		if err := json.Unmarshal(variantsJSON, &rows); err != nil {
			return reportingapp.WorklistItem{}, corrupt("product_worklist", product, err)
		}
	}
	for _, v := range rows {
		id, err := shared.ParseID(v.ID)
		if err != nil {
			return reportingapp.WorklistItem{}, corrupt("product_worklist", product, err)
		}
		w.Variants = append(w.Variants, reportingapp.WorklistVariant{ID: id, Label: v.Label})
	}
	if requestedBy != nil {
		id, err := shared.ParseID(*requestedBy)
		if err != nil {
			return reportingapp.WorklistItem{}, corrupt("product_worklist", product, err)
		}
		w.RequestedBy = id
	}
	return w, nil
}
