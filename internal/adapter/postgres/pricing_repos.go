package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/duongsy/portage/internal/domain/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
)

// The pricing ports on Postgres. Same construction as the catalog repos: a
// pool, db(ctx) for "transaction if there is one", upserts for value objects
// with a natural key, the domain's sentinel on a miss, and every row that
// fails to rebuild reported as corrupt rather than guessed at.

// ── lanes ────────────────────────────────────────────────────────────────────

type LaneRepo struct {
	pool *pgxpool.Pool
}

var _ pricing.LaneRepository = (*LaneRepo)(nil)

func NewLaneRepo(pool *pgxpool.Pool) *LaneRepo {
	return &LaneRepo{pool: pool}
}

const laneColumns = `code, name, divisor, step_g, currency,
	rate_standard_minor, rate_branded_minor, rate_electronics_minor, rate_sensitive_minor,
	surcharge_battery_minor, duty_itemised, duty_rate_ppm`

func (r *LaneRepo) Save(ctx context.Context, lane pricing.ShippingLane) error {
	rates := lane.Rates()
	var surcharge *int64
	if lane.BatterySurcharge().IsValid() {
		v := lane.BatterySurcharge().Minor()
		surcharge = &v
	}
	duty := lane.DutyPolicy()
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO lanes (`+laneColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (code) DO UPDATE SET
			name = EXCLUDED.name, divisor = EXCLUDED.divisor, step_g = EXCLUDED.step_g, currency = EXCLUDED.currency,
			rate_standard_minor = EXCLUDED.rate_standard_minor, rate_branded_minor = EXCLUDED.rate_branded_minor,
			rate_electronics_minor = EXCLUDED.rate_electronics_minor, rate_sensitive_minor = EXCLUDED.rate_sensitive_minor,
			surcharge_battery_minor = EXCLUDED.surcharge_battery_minor,
			duty_itemised = EXCLUDED.duty_itemised, duty_rate_ppm = EXCLUDED.duty_rate_ppm`,
		lane.Code().String(), lane.Name(), lane.Divisor(), lane.Step().Grams(), lane.Currency().Code(),
		rates.PerKg(pricing.ClassStandard).Minor(), rates.PerKg(pricing.ClassBranded).Minor(),
		rates.PerKg(pricing.ClassElectronics).Minor(), rates.PerKg(pricing.ClassSensitive).Minor(),
		surcharge, duty.Itemised(), duty.Rate().PPM())
	if err != nil {
		return fmt.Errorf("save lane %s: %w", lane.Code(), err)
	}
	return nil
}

func (r *LaneRepo) ByCode(ctx context.Context, code pricing.LaneCode) (pricing.ShippingLane, error) {
	rows, err := db(ctx, r.pool).Query(ctx, `SELECT `+laneColumns+` FROM lanes WHERE code = $1`, code.String())
	if err != nil {
		return pricing.ShippingLane{}, fmt.Errorf("lane %s: %w", code, err)
	}
	all, err := collectLanes(rows)
	if err != nil {
		return pricing.ShippingLane{}, err
	}
	if len(all) == 0 {
		return pricing.ShippingLane{}, fmt.Errorf("lane %s: %w", code, pricing.ErrLaneNotFound)
	}
	return all[0], nil
}

func (r *LaneRepo) All(ctx context.Context) ([]pricing.ShippingLane, error) {
	rows, err := db(ctx, r.pool).Query(ctx, `SELECT `+laneColumns+` FROM lanes ORDER BY code`)
	if err != nil {
		return nil, fmt.Errorf("lanes: %w", err)
	}
	return collectLanes(rows)
}

func collectLanes(rows pgx.Rows) ([]pricing.ShippingLane, error) {
	defer rows.Close()
	var out []pricing.ShippingLane
	for rows.Next() {
		var (
			code, name, currency string
			divisor, stepG       int64
			std, br, el, se      int64
			surcharge            *int64
			itemised             bool
			dutyPPM              int64
		)
		if err := rows.Scan(&code, &name, &divisor, &stepG, &currency, &std, &br, &el, &se, &surcharge, &itemised, &dutyPPM); err != nil {
			return nil, fmt.Errorf("scan lane: %w", err)
		}
		lane, err := laneFrom(code, name, divisor, stepG, currency, [4]int64{std, br, el, se}, surcharge, itemised, dutyPPM)
		if err != nil {
			return nil, corrupt("lanes", code, err)
		}
		out = append(out, lane)
	}
	return out, rows.Err()
}

// laneFrom rebuilds the value object through its validating constructor, so
// a row edited by hand into nonsense is refused on load, not quoted on.
func laneFrom(code, name string, divisor, stepG int64, currency string, perKg [4]int64, surcharge *int64, itemised bool, dutyPPM int64) (pricing.ShippingLane, error) {
	laneCode, err := pricing.ParseLaneCode(code)
	if err != nil {
		return pricing.ShippingLane{}, err
	}
	cur, err := shared.CurrencyFromCode(currency)
	if err != nil {
		return pricing.ShippingLane{}, err
	}
	rates, err := pricing.NewRateCard(
		shared.NewMoney(perKg[0], cur), shared.NewMoney(perKg[1], cur),
		shared.NewMoney(perKg[2], cur), shared.NewMoney(perKg[3], cur))
	if err != nil {
		return pricing.ShippingLane{}, err
	}
	step, err := shared.NewWeight(stepG)
	if err != nil {
		return pricing.ShippingLane{}, err
	}
	d := pricing.LaneDetails{Code: laneCode, Name: name, Divisor: divisor, Step: step, Rates: rates}
	if surcharge != nil {
		d.BatterySurcharge = shared.NewMoney(*surcharge, cur)
	}
	if itemised {
		if d.Duty, err = pricing.DutyItemised(shared.RatePPM(dutyPPM)); err != nil {
			return pricing.ShippingLane{}, err
		}
	}
	return pricing.NewShippingLane(d)
}

// ── quotes ───────────────────────────────────────────────────────────────────

type QuoteRepo struct {
	pool *pgxpool.Pool
}

var _ pricing.QuoteRepository = (*QuoteRepo)(nil)

func NewQuoteRepo(pool *pgxpool.Pool) *QuoteRepo {
	return &QuoteRepo{pool: pool}
}

const quoteColumns = `id, product, lane, status, issued_at, expires_at,
	class, chargeable_g, estimated, lane_currency,
	item_minor, tax_minor, freight_minor, surcharge_minor, duty_minor, subtotal_minor,
	fx_rate, home_currency, subtotal_home_minor, fee_minor, total_minor, deposit_minor`

// Save writes the whole snapshot. Only status ever changes after issue, but
// rewriting every column keeps Save a mirror of the snapshot — one shape to
// read, no "which columns are mutable" knowledge hidden in SQL.
func (r *QuoteRepo) Save(ctx context.Context, q *pricing.Quote) error {
	s := q.Snapshot()
	b := s.Breakdown
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO quotes (`+quoteColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
		ON CONFLICT (id) DO UPDATE SET
			product = EXCLUDED.product, lane = EXCLUDED.lane, status = EXCLUDED.status,
			issued_at = EXCLUDED.issued_at, expires_at = EXCLUDED.expires_at,
			class = EXCLUDED.class, chargeable_g = EXCLUDED.chargeable_g, estimated = EXCLUDED.estimated,
			lane_currency = EXCLUDED.lane_currency, item_minor = EXCLUDED.item_minor, tax_minor = EXCLUDED.tax_minor,
			freight_minor = EXCLUDED.freight_minor, surcharge_minor = EXCLUDED.surcharge_minor, duty_minor = EXCLUDED.duty_minor,
			subtotal_minor = EXCLUDED.subtotal_minor, fx_rate = EXCLUDED.fx_rate, home_currency = EXCLUDED.home_currency,
			subtotal_home_minor = EXCLUDED.subtotal_home_minor, fee_minor = EXCLUDED.fee_minor,
			total_minor = EXCLUDED.total_minor, deposit_minor = EXCLUDED.deposit_minor`,
		s.ID.String(), s.Product.String(), s.Lane.String(), string(s.Status), s.IssuedAt, s.ExpiresAt,
		string(b.Class), b.Chargeable.Grams(), b.Estimated, b.ItemPrice.Currency().Code(),
		b.ItemPrice.Minor(), b.SalesTax.Minor(), b.Freight.Minor(), b.Surcharge.Minor(), b.Duty.Minor(), b.SubtotalUSD.Minor(),
		b.FX.Rate(), b.TotalVND.Currency().Code(), b.SubtotalVND.Minor(), b.ServiceFee.Minor(), b.TotalVND.Minor(), b.Deposit.Minor())
	if err != nil {
		return fmt.Errorf("save quote %s: %w", s.ID, err)
	}
	return nil
}

// scanner is what pgx.Row and pgx.Rows have in common. ByID reads one row and
// IssuedBefore reads many, but the twenty-two columns become a Quote exactly
// once — a second copy of this would be a second place for a column to drift.
type scanner interface {
	Scan(dest ...any) error
}

// scanQuote turns one row into an aggregate through QuoteFromSnapshot, so a
// row that no longer satisfies the domain's invariants is a loud error, not a
// half-built Quote.
func scanQuote(s scanner) (*pricing.Quote, error) {
	var (
		rawID, product, lane, status, class, laneCur, fxRate, homeCur string
		issuedAt, expiresAt                                           time.Time
		chargeable                                                    int64
		estimated                                                     bool
		item, tax, freight, surcharge, duty, subtotal                 int64
		subtotalHome, fee, total, deposit                             int64
	)
	err := s.Scan(&rawID, &product, &lane, &status, &issuedAt, &expiresAt,
		&class, &chargeable, &estimated, &laneCur,
		&item, &tax, &freight, &surcharge, &duty, &subtotal,
		&fxRate, &homeCur, &subtotalHome, &fee, &total, &deposit)
	if err != nil {
		return nil, err // pgx.ErrNoRows included: the caller knows what it asked for
	}

	fail := func(err error) (*pricing.Quote, error) {
		return nil, corrupt("quotes", rawID, err)
	}
	qid, err := pricing.ParseQuoteID(rawID)
	if err != nil {
		return fail(err)
	}
	pid, err := shared.ParseID(product)
	if err != nil {
		return fail(err)
	}
	laneCode, err := pricing.ParseLaneCode(lane)
	if err != nil {
		return fail(err)
	}
	from, err := shared.CurrencyFromCode(laneCur)
	if err != nil {
		return fail(err)
	}
	home, err := shared.CurrencyFromCode(homeCur)
	if err != nil {
		return fail(err)
	}
	fx, err := shared.NewExchangeRate(from, home, fxRate)
	if err != nil {
		return fail(err)
	}
	weight, err := shared.NewWeight(chargeable)
	if err != nil {
		return fail(err)
	}
	q, err := pricing.QuoteFromSnapshot(pricing.QuoteSnapshot{
		ID: qid, Product: pid, Lane: laneCode, Status: pricing.QuoteStatus(status),
		IssuedAt: issuedAt.UTC(), ExpiresAt: expiresAt.UTC(),
		Breakdown: pricing.Breakdown{
			Class: pricing.GoodsClass(class), Chargeable: weight, Estimated: estimated,
			ItemPrice: shared.NewMoney(item, from), SalesTax: shared.NewMoney(tax, from), Freight: shared.NewMoney(freight, from),
			Surcharge: shared.NewMoney(surcharge, from), Duty: shared.NewMoney(duty, from), SubtotalUSD: shared.NewMoney(subtotal, from),
			FX: fx, SubtotalVND: shared.NewMoney(subtotalHome, home), ServiceFee: shared.NewMoney(fee, home),
			TotalVND: shared.NewMoney(total, home), Deposit: shared.NewMoney(deposit, home),
		},
	})
	if err != nil {
		return fail(err)
	}
	return q, nil
}

func (r *QuoteRepo) ByID(ctx context.Context, id pricing.QuoteID) (*pricing.Quote, error) {
	q, err := scanQuote(db(ctx, r.pool).QueryRow(ctx, `SELECT `+quoteColumns+` FROM quotes WHERE id = $1`, id.String()))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("quote %s: %w", id, pricing.ErrQuoteNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("quote %s: %w", id, err)
	}
	return q, nil
}

// IssuedBefore is the expiry sweep's query. quotes_open_idx (0002_pricing.sql)
// is a PARTIAL index on expires_at WHERE status = 'issued', so this reads only
// the quotes that can still expire — it does not grow with the ones that
// already have.
//
// The rows are read fully before any of them is expired: each expiry is its
// own transaction (ExpireQuotesHandler), and holding a cursor open across
// twenty commits would keep this connection busy for all of them.
func (r *QuoteRepo) IssuedBefore(ctx context.Context, t time.Time, limit int) ([]*pricing.Quote, error) {
	rows, err := db(ctx, r.pool).Query(ctx, `
		SELECT `+quoteColumns+` FROM quotes
		WHERE status = 'issued' AND expires_at < $1
		ORDER BY expires_at, id
		LIMIT $2`, t, limit)
	if err != nil {
		return nil, fmt.Errorf("issued quotes before %s: %w", t.Format(time.RFC3339), err)
	}
	defer rows.Close()

	var out []*pricing.Quote
	for rows.Next() {
		q, err := scanQuote(rows)
		if err != nil {
			return nil, fmt.Errorf("issued quotes before %s: %w", t.Format(time.RFC3339), err)
		}
		out = append(out, q)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("issued quotes before %s: %w", t.Format(time.RFC3339), err)
	}
	return out, nil
}

// ── listings (projection) ────────────────────────────────────────────────────

type ListingRepo struct {
	pool *pgxpool.Pool
}

var _ pricing.ListingRepository = (*ListingRepo)(nil)

func NewListingRepo(pool *pgxpool.Pool) *ListingRepo {
	return &ListingRepo{pool: pool}
}

const listingColumns = `product, name, category, price_minor, price_currency,
	parcel_weight_g, parcel_length_mm, parcel_width_mm, parcel_height_mm, measured, active`

func (r *ListingRepo) Save(ctx context.Context, l pricing.Listing) error {
	var w, ln, wd, h *int64
	if !l.Parcel.IsZero() {
		g, d := l.Parcel.Weight().Grams(), l.Parcel.Dimensions()
		a, b, c := d.LengthMM(), d.WidthMM(), d.HeightMM()
		w, ln, wd, h = &g, &a, &b, &c
	}
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO listings (`+listingColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (product) DO UPDATE SET
			name = EXCLUDED.name, category = EXCLUDED.category,
			price_minor = EXCLUDED.price_minor, price_currency = EXCLUDED.price_currency,
			parcel_weight_g = EXCLUDED.parcel_weight_g, parcel_length_mm = EXCLUDED.parcel_length_mm,
			parcel_width_mm = EXCLUDED.parcel_width_mm, parcel_height_mm = EXCLUDED.parcel_height_mm,
			measured = EXCLUDED.measured, active = EXCLUDED.active`,
		l.Product.String(), l.Name, l.Category, l.Price.Minor(), l.Price.Currency().Code(), w, ln, wd, h, l.Measured, l.Active)
	if err != nil {
		return fmt.Errorf("save listing %s: %w", l.Product, err)
	}
	return nil
}

func (r *ListingRepo) ByProduct(ctx context.Context, product shared.ID) (pricing.Listing, error) {
	row := db(ctx, r.pool).QueryRow(ctx, `SELECT `+listingColumns+` FROM listings WHERE product = $1`, product.String())
	var (
		rawID, name, category, currency string
		priceMinor                      int64
		w, ln, wd, h                    *int64
		measured, active                bool
	)
	err := row.Scan(&rawID, &name, &category, &priceMinor, &currency, &w, &ln, &wd, &h, &measured, &active)
	if errors.Is(err, pgx.ErrNoRows) {
		return pricing.Listing{}, fmt.Errorf("listing for product %s: %w", product, pricing.ErrListingNotFound)
	}
	if err != nil {
		return pricing.Listing{}, fmt.Errorf("listing %s: %w", product, err)
	}
	fail := func(err error) (pricing.Listing, error) {
		return pricing.Listing{}, corrupt("listings", rawID, err)
	}
	id, err := shared.ParseID(rawID)
	if err != nil {
		return fail(err)
	}
	cur, err := shared.CurrencyFromCode(currency)
	if err != nil {
		return fail(err)
	}
	l := pricing.Listing{
		Product: id, Name: name, Category: category, Price: shared.NewMoney(priceMinor, cur),
		Measured: measured, Active: active,
	}
	if w != nil && ln != nil && wd != nil && h != nil {
		spec, err := parcelFromColumns(*w, *ln, *wd, *h)
		if err != nil {
			return fail(err)
		}
		l.Parcel = spec
	}
	return l, nil
}

// ── category profiles (projection) ───────────────────────────────────────────

type ProfileRepo struct {
	pool *pgxpool.Pool
}

var _ pricing.CategoryProfileRepository = (*ProfileRepo)(nil)

func NewProfileRepo(pool *pgxpool.Pool) *ProfileRepo {
	return &ProfileRepo{pool: pool}
}

const profileColumns = `code, class, est_weight_g, est_length_mm, est_width_mm, est_height_mm, restrictions`

func (r *ProfileRepo) Save(ctx context.Context, p pricing.CategoryProfile) error {
	d := p.Estimate.Dimensions()
	restrictions := p.Restrictions
	if restrictions == nil {
		restrictions = []string{}
	}
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO category_profiles (`+profileColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (code) DO UPDATE SET
			class = EXCLUDED.class, est_weight_g = EXCLUDED.est_weight_g, est_length_mm = EXCLUDED.est_length_mm,
			est_width_mm = EXCLUDED.est_width_mm, est_height_mm = EXCLUDED.est_height_mm, restrictions = EXCLUDED.restrictions`,
		p.Code, string(p.Class), p.Estimate.Weight().Grams(), d.LengthMM(), d.WidthMM(), d.HeightMM(), restrictions)
	if err != nil {
		return fmt.Errorf("save category profile %s: %w", p.Code, err)
	}
	return nil
}

func (r *ProfileRepo) ByCode(ctx context.Context, code string) (pricing.CategoryProfile, error) {
	row := db(ctx, r.pool).QueryRow(ctx, `SELECT `+profileColumns+` FROM category_profiles WHERE code = $1`, code)
	var (
		rawCode, class string
		g, l, w, h     int64
		restrictions   []string
	)
	err := row.Scan(&rawCode, &class, &g, &l, &w, &h, &restrictions)
	if errors.Is(err, pgx.ErrNoRows) {
		return pricing.CategoryProfile{}, fmt.Errorf("category profile %s: %w", code, pricing.ErrProfileNotFound)
	}
	if err != nil {
		return pricing.CategoryProfile{}, fmt.Errorf("category profile %s: %w", code, err)
	}
	spec, err := parcelFromColumns(g, l, w, h)
	if err != nil {
		return pricing.CategoryProfile{}, corrupt("category_profiles", rawCode, err)
	}
	if len(restrictions) == 0 {
		restrictions = nil // the zero value the domain uses for "none"
	}
	return pricing.CategoryProfile{Code: rawCode, Class: pricing.GoodsClass(class), Estimate: spec, Restrictions: restrictions}, nil
}

// ── exchange rates ───────────────────────────────────────────────────────────

type ExchangeRates struct {
	pool *pgxpool.Pool
}

var _ pricing.ExchangeRates = (*ExchangeRates)(nil)

func NewExchangeRates(pool *pgxpool.Pool) *ExchangeRates {
	return &ExchangeRates{pool: pool}
}

func (r *ExchangeRates) Set(ctx context.Context, rate shared.ExchangeRate) error {
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO fx_rates (from_code, to_code, rate) VALUES ($1, $2, $3)
		ON CONFLICT (from_code, to_code) DO UPDATE SET rate = EXCLUDED.rate, updated_at = now()`,
		rate.From().Code(), rate.To().Code(), rate.Rate())
	if err != nil {
		return fmt.Errorf("set rate %s: %w", rate, err)
	}
	return nil
}

func (r *ExchangeRates) Current(ctx context.Context, from, to shared.Currency) (shared.ExchangeRate, error) {
	var text string
	err := db(ctx, r.pool).QueryRow(ctx, `SELECT rate FROM fx_rates WHERE from_code = $1 AND to_code = $2`, from.Code(), to.Code()).Scan(&text)
	if errors.Is(err, pgx.ErrNoRows) {
		return shared.ExchangeRate{}, fmt.Errorf("%s→%s: %w", from, to, pricing.ErrNoExchangeRate)
	}
	if err != nil {
		return shared.ExchangeRate{}, fmt.Errorf("rate %s→%s: %w", from, to, err)
	}
	rate, err := shared.NewExchangeRate(from, to, text)
	if err != nil {
		return shared.ExchangeRate{}, corrupt("fx_rates", from.Code()+"/"+to.Code(), err)
	}
	return rate, nil
}

// parcelFromColumns is the shared half of "four integers → ParcelSpec".
func parcelFromColumns(g, l, w, h int64) (shared.ParcelSpec, error) {
	weight, err := shared.NewWeight(g)
	if err != nil {
		return shared.ParcelSpec{}, err
	}
	if l < 0 || w < 0 || h < 0 {
		return shared.ParcelSpec{}, errors.New("negative dimension")
	}
	return shared.NewParcelSpec(weight, shared.NewDimensionsMM(l, w, h))
}
