package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

// ReconciliationRepo stores pricing's Quote vs Actual rows. Every amount is
// nullable: the row grows as events arrive in whatever order they come.
type ReconciliationRepo struct {
	pool *pgxpool.Pool
}

var _ pricing.ReconciliationRepository = (*ReconciliationRepo)(nil)

func NewReconciliationRepo(pool *pgxpool.Pool) *ReconciliationRepo {
	return &ReconciliationRepo{pool: pool}
}

func (r *ReconciliationRepo) Save(ctx context.Context, rec pricing.Reconciliation) error {
	var quote, currency *string
	if !rec.Quote.IsZero() {
		v := rec.Quote.String()
		quote = &v
	}
	minor := func(m shared.Money) *int64 {
		if !m.IsValid() {
			return nil
		}
		v := m.Minor()
		c := m.Currency().Code()
		currency = &c
		return &v
	}
	grams := func(w shared.Weight) *int64 {
		if w.IsZero() {
			return nil
		}
		v := w.Grams()
		return &v
	}
	qg, qf, ag, af := minor(rec.QuotedGoods), minor(rec.QuotedFreight), minor(rec.ActualGoods), minor(rec.ActualFreight)
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO reconciliations ("order", quote, currency, quoted_goods_minor, quoted_freight_minor, quoted_chargeable_g, actual_goods_minor, actual_freight_minor, actual_chargeable_g)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT ("order") DO UPDATE SET quote = EXCLUDED.quote, currency = EXCLUDED.currency,
			quoted_goods_minor = EXCLUDED.quoted_goods_minor, quoted_freight_minor = EXCLUDED.quoted_freight_minor, quoted_chargeable_g = EXCLUDED.quoted_chargeable_g,
			actual_goods_minor = EXCLUDED.actual_goods_minor, actual_freight_minor = EXCLUDED.actual_freight_minor, actual_chargeable_g = EXCLUDED.actual_chargeable_g`,
		rec.Order.String(), quote, currency, qg, qf, grams(rec.QuotedChargeable), ag, af, grams(rec.ActualChargeable))
	if err != nil {
		return fmt.Errorf("save reconciliation %s: %w", rec.Order, err)
	}
	return nil
}

func (r *ReconciliationRepo) ByOrder(ctx context.Context, order shared.ID) (pricing.Reconciliation, error) {
	var (
		rawOrder               string
		quote, currency        *string
		qg, qf, qc, ag, af, ac *int64
	)
	err := db(ctx, r.pool).QueryRow(ctx, `SELECT "order", quote, currency, quoted_goods_minor, quoted_freight_minor, quoted_chargeable_g, actual_goods_minor, actual_freight_minor, actual_chargeable_g FROM reconciliations WHERE "order" = $1`, order.String()).
		Scan(&rawOrder, &quote, &currency, &qg, &qf, &qc, &ag, &af, &ac)
	if errors.Is(err, pgx.ErrNoRows) {
		return pricing.Reconciliation{}, fmt.Errorf("reconciliation for order %s: %w", order, pricing.ErrReconciliationNotFound)
	}
	if err != nil {
		return pricing.Reconciliation{}, fmt.Errorf("reconciliation %s: %w", order, err)
	}
	fail := func(err error) (pricing.Reconciliation, error) {
		return pricing.Reconciliation{}, corrupt("reconciliations", rawOrder, err)
	}
	id, err := shared.ParseID(rawOrder)
	if err != nil {
		return fail(err)
	}
	rec := pricing.Reconciliation{Order: id}
	if quote != nil {
		if rec.Quote, err = shared.ParseID(*quote); err != nil {
			return fail(err)
		}
	}
	var cur shared.Currency
	if currency != nil {
		if cur, err = shared.CurrencyFromCode(*currency); err != nil {
			return fail(err)
		}
	}
	money := func(v *int64) shared.Money {
		if v == nil || cur.IsZero() {
			return shared.Money{}
		}
		return shared.NewMoney(*v, cur)
	}
	grams := func(v *int64) shared.Weight {
		if v == nil {
			return shared.Weight{}
		}
		return shared.Grams(*v)
	}
	rec.QuotedGoods, rec.QuotedFreight, rec.QuotedChargeable = money(qg), money(qf), grams(qc)
	rec.ActualGoods, rec.ActualFreight, rec.ActualChargeable = money(ag), money(af), grams(ac)
	return rec, nil
}
