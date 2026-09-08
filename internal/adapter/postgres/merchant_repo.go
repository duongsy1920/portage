package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// MerchantRepo implements catalog.MerchantRepository on the merchants table.
type MerchantRepo struct {
	pool *pgxpool.Pool
}

var _ catalog.MerchantRepository = (*MerchantRepo)(nil)

func NewMerchantRepo(pool *pgxpool.Pool) *MerchantRepo {
	return &MerchantRepo{pool: pool}
}

const merchantColumns = `id, name, site, currency, free_ship_kind, free_ship_threshold_minor, sourcing, status, added_at`

// Save writes the snapshot as an upsert: a new merchant inserts, a changed one
// updates every column. Same statement, same code path, no "is it new?" flag.
//
// [PHP] persist() + flush() gộp lại. ON CONFLICT (id) DO UPDATE là "upsert" —
// [PHP] Doctrine không có sẵn, bạn phải find() rồi quyết định persist hay merge.
func (r *MerchantRepo) Save(ctx context.Context, m *catalog.Merchant) error {
	s := m.Snapshot()
	var threshold *int64
	if t, ok := s.FreeShipping.Threshold(); ok {
		v := t.Minor()
		threshold = &v
	}
	modes := make([]string, 0, len(s.Sourcing))
	for _, mode := range s.Sourcing {
		modes = append(modes, string(mode))
	}
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO merchants (`+merchantColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name, site = EXCLUDED.site, currency = EXCLUDED.currency,
			free_ship_kind = EXCLUDED.free_ship_kind, free_ship_threshold_minor = EXCLUDED.free_ship_threshold_minor,
			sourcing = EXCLUDED.sourcing, status = EXCLUDED.status, added_at = EXCLUDED.added_at`,
		s.ID.String(), s.Name, s.Site.String(), s.Currency.Code(),
		string(s.FreeShipping.Kind()), threshold, modes, string(s.Status), s.AddedAt)
	if err != nil {
		return fmt.Errorf("save merchant %s: %w", s.ID, err)
	}
	return nil
}

func (r *MerchantRepo) ByID(ctx context.Context, id catalog.MerchantID) (*catalog.Merchant, error) {
	row := db(ctx, r.pool).QueryRow(ctx, `SELECT `+merchantColumns+` FROM merchants WHERE id = $1`, id.String())
	m, err := scanMerchant(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("merchant %s: %w", id, catalog.ErrMerchantNotFound)
	}
	return m, err
}

func (r *MerchantRepo) BySite(ctx context.Context, site catalog.Hostname) (*catalog.Merchant, error) {
	row := db(ctx, r.pool).QueryRow(ctx, `SELECT `+merchantColumns+` FROM merchants WHERE site = $1`, site.String())
	m, err := scanMerchant(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("merchant at %s: %w", site, catalog.ErrMerchantNotFound)
	}
	return m, err
}

// All is the read side of GET /merchants. ORDER BY id is ORDER BY creation
// time, because the ids are UUIDv7.
func (r *MerchantRepo) All(ctx context.Context) ([]*catalog.Merchant, error) {
	rows, err := db(ctx, r.pool).Query(ctx, `SELECT `+merchantColumns+` FROM merchants ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list merchants: %w", err)
	}
	defer rows.Close()
	out := []*catalog.Merchant{}
	for rows.Next() {
		m, err := scanMerchant(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// scanMerchant reads one row into a snapshot and lets the domain rebuild the
// aggregate. Every column goes back through the domain's own parsers: a row
// the domain would not accept is corruption, reported, never a half-merchant.
func scanMerchant(row pgx.Row) (*catalog.Merchant, error) {
	var (
		id, name, site, currency, kind, status string
		threshold                              *int64
		modes                                  []string
		addedAt                                time.Time
	)
	if err := row.Scan(&id, &name, &site, &currency, &kind, &threshold, &modes, &status, &addedAt); err != nil {
		return nil, err
	}
	mid, err := catalog.ParseMerchantID(id)
	if err != nil {
		return nil, corrupt("merchants", id, err)
	}
	host, err := catalog.ParseHostname(site)
	if err != nil {
		return nil, corrupt("merchants", id, err)
	}
	cur, err := shared.CurrencyFromCode(currency)
	if err != nil {
		return nil, corrupt("merchants", id, err)
	}
	freeShip, err := freeShippingFrom(kind, threshold, cur)
	if err != nil {
		return nil, corrupt("merchants", id, err)
	}
	sourcing := make([]catalog.SourcingMode, 0, len(modes))
	for _, m := range modes {
		sourcing = append(sourcing, catalog.SourcingMode(m))
	}
	m, err := catalog.MerchantFromSnapshot(catalog.MerchantSnapshot{
		ID: mid, Name: name, Site: host, Currency: cur, FreeShipping: freeShip,
		Sourcing: sourcing, Status: catalog.MerchantStatus(status), AddedAt: addedAt.UTC(),
	})
	if err != nil {
		return nil, corrupt("merchants", id, err)
	}
	return m, nil
}

// freeShippingFrom rebuilds the rule from its two columns through the
// constructors — the store never touches FreeShipping's insides.
func freeShippingFrom(kind string, threshold *int64, cur shared.Currency) (catalog.FreeShipping, error) {
	switch catalog.FreeShippingKind(kind) {
	case catalog.FreeShipNever:
		return catalog.NoFreeShipping(), nil
	case catalog.FreeShipAlways:
		return catalog.AlwaysFreeShipping(), nil
	case catalog.FreeShipOver:
		if threshold == nil {
			return catalog.FreeShipping{}, errors.New("free_ship_kind is over but threshold is NULL")
		}
		return catalog.FreeShippingOver(shared.NewMoney(*threshold, cur))
	default:
		return catalog.FreeShipping{}, fmt.Errorf("free_ship_kind %q", kind)
	}
}

// corrupt names the table and row so the log says exactly what to repair.
func corrupt(table, id string, err error) error {
	return fmt.Errorf("%s row %s is corrupt: %w", table, id, err)
}
