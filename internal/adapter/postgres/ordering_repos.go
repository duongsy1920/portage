package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/duongsy/portage/internal/domain/ordering"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The ordering ports on Postgres.

type OrderRepo struct {
	pool *pgxpool.Pool
}

var _ ordering.OrderRepository = (*OrderRepo)(nil)

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{pool: pool}
}

const orderColumns = `id, quote, product, variant, customer, total_minor, deposit_minor, currency,
	status, balance_paid, refund_minor, forfeited, cancel_reason, placed_at`

func (r *OrderRepo) Save(ctx context.Context, o *ordering.CustomerOrder) error {
	s := o.Snapshot()
	var refund *int64
	if s.Refund.IsValid() {
		v := s.Refund.Minor()
		refund = &v
	}
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO orders (`+orderColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO UPDATE SET
			quote = EXCLUDED.quote, product = EXCLUDED.product, variant = EXCLUDED.variant, customer = EXCLUDED.customer,
			total_minor = EXCLUDED.total_minor, deposit_minor = EXCLUDED.deposit_minor, currency = EXCLUDED.currency,
			status = EXCLUDED.status, balance_paid = EXCLUDED.balance_paid, refund_minor = EXCLUDED.refund_minor,
			forfeited = EXCLUDED.forfeited, cancel_reason = EXCLUDED.cancel_reason, placed_at = EXCLUDED.placed_at`,
		s.ID.String(), s.Quote.String(), s.Product.String(), s.Variant.String(), s.Customer.String(),
		s.Total.Minor(), s.Deposit.Minor(), s.Total.Currency().Code(),
		string(s.Status), s.BalancePaid, refund, s.Forfeited, s.CancelReason, s.PlacedAt)
	if err != nil {
		return fmt.Errorf("save order %s: %w", s.ID, err)
	}
	return nil
}

func (r *OrderRepo) ByID(ctx context.Context, id ordering.OrderID) (*ordering.CustomerOrder, error) {
	return r.one(ctx, `SELECT `+orderColumns+` FROM orders WHERE id = $1`, id.String(), fmt.Sprintf("order %s", id))
}

func (r *OrderRepo) ByQuote(ctx context.Context, quote shared.ID) (*ordering.CustomerOrder, error) {
	return r.one(ctx, `SELECT `+orderColumns+` FROM orders WHERE quote = $1`, quote.String(), fmt.Sprintf("order for quote %s", quote))
}

func (r *OrderRepo) one(ctx context.Context, sql, arg, what string) (*ordering.CustomerOrder, error) {
	row := db(ctx, r.pool).QueryRow(ctx, sql, arg)
	var (
		rawID, quote, product, variant, customer, currency, status, reason string
		total, deposit                                                     int64
		balancePaid, forfeited                                             bool
		refund                                                             *int64
		placedAt                                                           time.Time
	)
	err := row.Scan(&rawID, &quote, &product, &variant, &customer, &total, &deposit, &currency,
		&status, &balancePaid, &refund, &forfeited, &reason, &placedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", what, ordering.ErrOrderNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", what, err)
	}
	fail := func(err error) (*ordering.CustomerOrder, error) {
		return nil, corrupt("orders", rawID, err)
	}
	id, err := ordering.ParseOrderID(rawID)
	if err != nil {
		return fail(err)
	}
	ids := make([]shared.ID, 4)
	for i, raw := range []string{quote, product, variant, customer} {
		if ids[i], err = shared.ParseID(raw); err != nil {
			return fail(err)
		}
	}
	cur, err := shared.CurrencyFromCode(currency)
	if err != nil {
		return fail(err)
	}
	snap := ordering.OrderSnapshot{
		ID: id, Quote: ids[0], Product: ids[1], Variant: ids[2], Customer: ids[3],
		Total: shared.NewMoney(total, cur), Deposit: shared.NewMoney(deposit, cur),
		Status: ordering.OrderStatus(status), BalancePaid: balancePaid, Forfeited: forfeited, CancelReason: reason, PlacedAt: placedAt.UTC(),
	}
	if refund != nil {
		snap.Refund = shared.NewMoney(*refund, cur)
	}
	o, err := ordering.OrderFromSnapshot(snap)
	if err != nil {
		return fail(err)
	}
	return o, nil
}

type AcceptedQuoteRepo struct {
	pool *pgxpool.Pool
}

var _ ordering.AcceptedQuoteRepository = (*AcceptedQuoteRepo)(nil)

func NewAcceptedQuoteRepo(pool *pgxpool.Pool) *AcceptedQuoteRepo {
	return &AcceptedQuoteRepo{pool: pool}
}

func (r *AcceptedQuoteRepo) Save(ctx context.Context, q ordering.AcceptedQuote) error {
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO accepted_quotes (quote, product, total_minor, deposit_minor, currency)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (quote) DO UPDATE SET
			product = EXCLUDED.product, total_minor = EXCLUDED.total_minor, deposit_minor = EXCLUDED.deposit_minor, currency = EXCLUDED.currency`,
		q.Quote.String(), q.Product.String(), q.Total.Minor(), q.Deposit.Minor(), q.Total.Currency().Code())
	if err != nil {
		return fmt.Errorf("save accepted quote %s: %w", q.Quote, err)
	}
	return nil
}

func (r *AcceptedQuoteRepo) ByID(ctx context.Context, quote shared.ID) (ordering.AcceptedQuote, error) {
	var (
		rawQuote, product, currency string
		total, deposit              int64
	)
	err := db(ctx, r.pool).QueryRow(ctx, `SELECT quote, product, total_minor, deposit_minor, currency FROM accepted_quotes WHERE quote = $1`, quote.String()).
		Scan(&rawQuote, &product, &total, &deposit, &currency)
	if errors.Is(err, pgx.ErrNoRows) {
		return ordering.AcceptedQuote{}, fmt.Errorf("accepted quote %s: %w", quote, ordering.ErrAcceptedQuoteNotFound)
	}
	if err != nil {
		return ordering.AcceptedQuote{}, fmt.Errorf("accepted quote %s: %w", quote, err)
	}
	fail := func(err error) (ordering.AcceptedQuote, error) {
		return ordering.AcceptedQuote{}, corrupt("accepted_quotes", rawQuote, err)
	}
	q, err := shared.ParseID(rawQuote)
	if err != nil {
		return fail(err)
	}
	p, err := shared.ParseID(product)
	if err != nil {
		return fail(err)
	}
	cur, err := shared.CurrencyFromCode(currency)
	if err != nil {
		return fail(err)
	}
	return ordering.AcceptedQuote{Quote: q, Product: p, Total: shared.NewMoney(total, cur), Deposit: shared.NewMoney(deposit, cur)}, nil
}

// VariantRepo is ordering's projection of catalog's variants: existence and
// which product, nothing more (see ordering.Variant).
type OrderingVariantRepo struct {
	pool *pgxpool.Pool
}

var _ ordering.VariantRepository = (*OrderingVariantRepo)(nil)

func NewOrderingVariantRepo(pool *pgxpool.Pool) *OrderingVariantRepo {
	return &OrderingVariantRepo{pool: pool}
}

func (r *OrderingVariantRepo) Save(ctx context.Context, v ordering.Variant) error {
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO ordering_variants (variant, product) VALUES ($1, $2)
		ON CONFLICT (variant) DO UPDATE SET product = EXCLUDED.product`,
		v.Variant.String(), v.Product.String())
	if err != nil {
		return fmt.Errorf("save variant %s: %w", v.Variant, err)
	}
	return nil
}

func (r *OrderingVariantRepo) ByID(ctx context.Context, variant shared.ID) (ordering.Variant, error) {
	var raw, product string
	err := db(ctx, r.pool).QueryRow(ctx,
		`SELECT variant, product FROM ordering_variants WHERE variant = $1`, variant.String()).Scan(&raw, &product)
	if errors.Is(err, pgx.ErrNoRows) {
		return ordering.Variant{}, fmt.Errorf("variant %s: %w", variant, ordering.ErrVariantNotFound)
	}
	if err != nil {
		return ordering.Variant{}, fmt.Errorf("variant %s: %w", variant, err)
	}
	v, err := shared.ParseID(raw)
	if err != nil {
		return ordering.Variant{}, corrupt("ordering_variants", raw, err)
	}
	p, err := shared.ParseID(product)
	if err != nil {
		return ordering.Variant{}, corrupt("ordering_variants", raw, err)
	}
	return ordering.Variant{Variant: v, Product: p}, nil
}
