package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/duongsy/portage/internal/domain/logistics"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The logistics ports on Postgres.

type ParcelRepo struct {
	pool *pgxpool.Pool
}

var _ logistics.ParcelRepository = (*ParcelRepo)(nil)

func NewParcelRepo(pool *pgxpool.Pool) *ParcelRepo {
	return &ParcelRepo{pool: pool}
}

const parcelColumns = `id, "order", reference, status, actual_weight_g, actual_length_mm, actual_width_mm, actual_height_mm,
	received_by, received_at, batch, expected_at`

func (r *ParcelRepo) Save(ctx context.Context, p *logistics.Parcel) error {
	s := p.Snapshot()
	var w, l, wd, h *int64
	if !s.Actual.IsZero() {
		g, d := s.Actual.Weight().Grams(), s.Actual.Dimensions()
		a, b, c := d.LengthMM(), d.WidthMM(), d.HeightMM()
		w, l, wd, h = &g, &a, &b, &c
	}
	var by, batch *string
	if !s.ReceivedBy.IsZero() {
		v := s.ReceivedBy.String()
		by = &v
	}
	if !s.Batch.IsZero() {
		v := s.Batch.String()
		batch = &v
	}
	var receivedAt *time.Time
	if !s.ReceivedAt.IsZero() {
		v := s.ReceivedAt
		receivedAt = &v
	}
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO parcels (`+parcelColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			"order" = EXCLUDED."order", reference = EXCLUDED.reference, status = EXCLUDED.status,
			actual_weight_g = EXCLUDED.actual_weight_g, actual_length_mm = EXCLUDED.actual_length_mm,
			actual_width_mm = EXCLUDED.actual_width_mm, actual_height_mm = EXCLUDED.actual_height_mm,
			received_by = EXCLUDED.received_by, received_at = EXCLUDED.received_at, batch = EXCLUDED.batch, expected_at = EXCLUDED.expected_at`,
		s.ID.String(), s.Order.String(), s.Reference, string(s.Status), w, l, wd, h, by, receivedAt, batch, s.ExpectedAt)
	if err != nil {
		return fmt.Errorf("save parcel %s: %w", s.ID, err)
	}
	return nil
}

func (r *ParcelRepo) ByID(ctx context.Context, id logistics.ParcelID) (*logistics.Parcel, error) {
	all, err := r.query(ctx, `SELECT `+parcelColumns+` FROM parcels WHERE id = $1`, id.String())
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("parcel %s: %w", id, logistics.ErrParcelNotFound)
	}
	return all[0], nil
}

func (r *ParcelRepo) ByOrder(ctx context.Context, order shared.ID) (*logistics.Parcel, error) {
	all, err := r.query(ctx, `SELECT `+parcelColumns+` FROM parcels WHERE "order" = $1`, order.String())
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("parcel for order %s: %w", order, logistics.ErrParcelNotFound)
	}
	return all[0], nil
}

func (r *ParcelRepo) Pending(ctx context.Context) ([]*logistics.Parcel, error) {
	return r.query(ctx, `SELECT `+parcelColumns+` FROM parcels WHERE status <> 'shipped' ORDER BY expected_at, id`)
}

func (r *ParcelRepo) query(ctx context.Context, sql string, args ...any) ([]*logistics.Parcel, error) {
	rows, err := db(ctx, r.pool).Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("parcels: %w", err)
	}
	defer rows.Close()
	var out []*logistics.Parcel
	for rows.Next() {
		var (
			rawID, order, reference, status string
			w, l, wd, h                     *int64
			by, batch                       *string
			receivedAt                      *time.Time
			expectedAt                      time.Time
		)
		if err := rows.Scan(&rawID, &order, &reference, &status, &w, &l, &wd, &h, &by, &receivedAt, &batch, &expectedAt); err != nil {
			return nil, fmt.Errorf("scan parcel: %w", err)
		}
		p, err := parcelFrom(rawID, order, reference, status, w, l, wd, h, by, receivedAt, batch, expectedAt)
		if err != nil {
			return nil, corrupt("parcels", rawID, err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func parcelFrom(rawID, order, reference, status string, w, l, wd, h *int64, by *string, receivedAt *time.Time, batch *string, expectedAt time.Time) (*logistics.Parcel, error) {
	id, err := logistics.ParseParcelID(rawID)
	if err != nil {
		return nil, err
	}
	orderID, err := shared.ParseID(order)
	if err != nil {
		return nil, err
	}
	snap := logistics.ParcelSnapshot{ID: id, Order: orderID, Reference: reference, Status: logistics.ParcelStatus(status), ExpectedAt: expectedAt.UTC()}
	if w != nil && l != nil && wd != nil && h != nil {
		if snap.Actual, err = parcelFromColumns(*w, *l, *wd, *h); err != nil {
			return nil, err
		}
	}
	if by != nil {
		if snap.ReceivedBy, err = shared.ParseOperatorID(*by); err != nil {
			return nil, err
		}
	}
	if receivedAt != nil {
		snap.ReceivedAt = receivedAt.UTC()
	}
	if batch != nil {
		if snap.Batch, err = logistics.ParseBatchID(*batch); err != nil {
			return nil, err
		}
	}
	return logistics.ParcelFromSnapshot(snap)
}

type BatchRepo struct {
	pool *pgxpool.Pool
}

var _ logistics.BatchRepository = (*BatchRepo)(nil)

func NewBatchRepo(pool *pgxpool.Pool) *BatchRepo {
	return &BatchRepo{pool: pool}
}

// Save rewrites the batch row and its two child tables — the aggregate is
// saved whole, like a product with its variants.
func (r *BatchRepo) Save(ctx context.Context, b *logistics.ConsolidationBatch) error {
	s := b.Snapshot()
	q := db(ctx, r.pool)
	var freight *int64
	var currency *string
	if s.Freight.IsValid() {
		v, c := s.Freight.Minor(), s.Freight.Currency().Code()
		freight, currency = &v, &c
	}
	var closedAt, shippedAt *time.Time
	if !s.ClosedAt.IsZero() {
		v := s.ClosedAt
		closedAt = &v
	}
	if !s.ShippedAt.IsZero() {
		v := s.ShippedAt
		shippedAt = &v
	}
	if _, err := q.Exec(ctx, `
		INSERT INTO batches (id, lane, status, freight_minor, freight_currency, opened_at, closed_at, shipped_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET lane = EXCLUDED.lane, status = EXCLUDED.status, freight_minor = EXCLUDED.freight_minor,
			freight_currency = EXCLUDED.freight_currency, opened_at = EXCLUDED.opened_at, closed_at = EXCLUDED.closed_at, shipped_at = EXCLUDED.shipped_at`,
		s.ID.String(), s.Lane, string(s.Status), freight, currency, s.OpenedAt, closedAt, shippedAt); err != nil {
		return fmt.Errorf("save batch %s: %w", s.ID, err)
	}
	for _, table := range []string{"batch_items", "batch_allocations"} {
		if _, err := q.Exec(ctx, `DELETE FROM `+table+` WHERE batch = $1`, s.ID.String()); err != nil {
			return fmt.Errorf("save batch %s: %w", s.ID, err)
		}
	}
	for i, it := range s.Items {
		d := it.Actual.Dimensions()
		if _, err := q.Exec(ctx, `INSERT INTO batch_items (batch, position, parcel, "order", weight_g, length_mm, width_mm, height_mm) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			s.ID.String(), i, it.Parcel.String(), it.Order.String(), it.Actual.Weight().Grams(), d.LengthMM(), d.WidthMM(), d.HeightMM()); err != nil {
			return fmt.Errorf("save batch %s item %d: %w", s.ID, i, err)
		}
	}
	for i, a := range s.Allocations {
		if _, err := q.Exec(ctx, `INSERT INTO batch_allocations (batch, position, parcel, "order", chargeable_g, freight_minor) VALUES ($1, $2, $3, $4, $5, $6)`,
			s.ID.String(), i, a.Parcel.String(), a.Order.String(), a.Chargeable.Grams(), a.Freight.Minor()); err != nil {
			return fmt.Errorf("save batch %s allocation %d: %w", s.ID, i, err)
		}
	}
	return nil
}

func (r *BatchRepo) ByID(ctx context.Context, id logistics.BatchID) (*logistics.ConsolidationBatch, error) {
	q := db(ctx, r.pool)
	var (
		rawID, lane, status string
		freight             *int64
		currency            *string
		openedAt            time.Time
		closedAt, shippedAt *time.Time
	)
	err := q.QueryRow(ctx, `SELECT id, lane, status, freight_minor, freight_currency, opened_at, closed_at, shipped_at FROM batches WHERE id = $1`, id.String()).
		Scan(&rawID, &lane, &status, &freight, &currency, &openedAt, &closedAt, &shippedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("batch %s: %w", id, logistics.ErrBatchNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("batch %s: %w", id, err)
	}
	fail := func(err error) (*logistics.ConsolidationBatch, error) {
		return nil, corrupt("batches", rawID, err)
	}
	snap := logistics.BatchSnapshot{ID: id, Lane: lane, Status: logistics.BatchStatus(status), OpenedAt: openedAt.UTC()}
	var cur shared.Currency
	if freight != nil && currency != nil {
		c, err := shared.CurrencyFromCode(*currency)
		if err != nil {
			return fail(err)
		}
		cur = c
		snap.Freight = shared.NewMoney(*freight, cur)
	}
	if closedAt != nil {
		snap.ClosedAt = closedAt.UTC()
	}
	if shippedAt != nil {
		snap.ShippedAt = shippedAt.UTC()
	}

	rows, err := q.Query(ctx, `SELECT parcel, "order", weight_g, length_mm, width_mm, height_mm FROM batch_items WHERE batch = $1 ORDER BY position`, rawID)
	if err != nil {
		return nil, fmt.Errorf("batch %s items: %w", id, err)
	}
	for rows.Next() {
		var (
			parcel, order string
			g, l, w, h    int64
		)
		if err := rows.Scan(&parcel, &order, &g, &l, &w, &h); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan batch item: %w", err)
		}
		item, err := batchItemFrom(parcel, order, g, l, w, h)
		if err != nil {
			rows.Close()
			return fail(err)
		}
		snap.Items = append(snap.Items, item)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("batch %s items: %w", id, err)
	}

	rows, err = q.Query(ctx, `SELECT parcel, "order", chargeable_g, freight_minor FROM batch_allocations WHERE batch = $1 ORDER BY position`, rawID)
	if err != nil {
		return nil, fmt.Errorf("batch %s allocations: %w", id, err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			parcel, order string
			g, minor      int64
		)
		if err := rows.Scan(&parcel, &order, &g, &minor); err != nil {
			return nil, fmt.Errorf("scan allocation: %w", err)
		}
		p, err := shared.ParseID(parcel)
		if err != nil {
			return fail(err)
		}
		o, err := shared.ParseID(order)
		if err != nil {
			return fail(err)
		}
		if cur.IsZero() {
			return fail(errors.New("allocations without an invoice currency"))
		}
		snap.Allocations = append(snap.Allocations, logistics.Allocation{Parcel: p, Order: o, Chargeable: shared.Grams(g), Freight: shared.NewMoney(minor, cur)})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("batch %s allocations: %w", id, err)
	}
	b, err := logistics.BatchFromSnapshot(snap)
	if err != nil {
		return fail(err)
	}
	return b, nil
}

func batchItemFrom(parcel, order string, g, l, w, h int64) (logistics.BatchItem, error) {
	p, err := shared.ParseID(parcel)
	if err != nil {
		return logistics.BatchItem{}, err
	}
	o, err := shared.ParseID(order)
	if err != nil {
		return logistics.BatchItem{}, err
	}
	spec, err := parcelFromColumns(g, l, w, h)
	if err != nil {
		return logistics.BatchItem{}, err
	}
	return logistics.BatchItem{Parcel: p, Order: o, Actual: spec}, nil
}

type LaneRuleRepo struct {
	pool *pgxpool.Pool
}

var _ logistics.LaneRuleRepository = (*LaneRuleRepo)(nil)

func NewLaneRuleRepo(pool *pgxpool.Pool) *LaneRuleRepo {
	return &LaneRuleRepo{pool: pool}
}

func (r *LaneRuleRepo) Save(ctx context.Context, rule logistics.LaneRule) error {
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO lane_rules (code, divisor, step_g) VALUES ($1, $2, $3)
		ON CONFLICT (code) DO UPDATE SET divisor = EXCLUDED.divisor, step_g = EXCLUDED.step_g`,
		rule.Code, rule.Divisor, rule.Step.Grams())
	if err != nil {
		return fmt.Errorf("save lane rule %s: %w", rule.Code, err)
	}
	return nil
}

func (r *LaneRuleRepo) ByCode(ctx context.Context, code string) (logistics.LaneRule, error) {
	var (
		raw            string
		divisor, stepG int64
	)
	err := db(ctx, r.pool).QueryRow(ctx, `SELECT code, divisor, step_g FROM lane_rules WHERE code = $1`, code).Scan(&raw, &divisor, &stepG)
	if errors.Is(err, pgx.ErrNoRows) {
		return logistics.LaneRule{}, fmt.Errorf("lane rule %s: %w", code, logistics.ErrLaneRuleNotFound)
	}
	if err != nil {
		return logistics.LaneRule{}, fmt.Errorf("lane rule %s: %w", code, err)
	}
	step, err := shared.NewWeight(stepG)
	if err != nil {
		return logistics.LaneRule{}, corrupt("lane_rules", raw, err)
	}
	return logistics.LaneRule{Code: raw, Divisor: divisor, Step: step}, nil
}
