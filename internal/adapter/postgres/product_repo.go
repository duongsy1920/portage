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

// ProductRepo implements catalog.ProductRepository over products and
// product_variants. The aggregate is saved and loaded WHOLE: variants are
// rewritten with their parent, never addressed on their own (DDD.md §14).
type ProductRepo struct {
	pool *pgxpool.Pool
}

var _ catalog.ProductRepository = (*ProductRepo)(nil)

func NewProductRepo(pool *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{pool: pool}
}

const productColumns = `id, merchant_id, category, name, source_url, source_host,
	listing_source, listing_at, listing_by,
	price_minor, price_currency, price_source, price_at, price_by,
	parcel_weight_g, parcel_length_mm, parcel_width_mm, parcel_height_mm, parcel_source, parcel_at, parcel_by,
	suspected_duplicate_of, dismissed_duplicates, status, added_at, requested_by, requested_variant`

// Save upserts the product row and rewrites its variants (delete + insert):
// simple, correct for aggregate-owned children, and inside the caller's
// transaction like everything else here.
func (r *ProductRepo) Save(ctx context.Context, p *catalog.Product) error {
	s := p.Snapshot()
	q := db(ctx, r.pool)

	var parcelW, parcelL, parcelWd, parcelH *int64
	var parcelSource *string
	var parcelAt *time.Time
	var parcelBy *string
	if !s.Parcel.IsZero() {
		w, d := s.Parcel.Weight().Grams(), s.Parcel.Dimensions()
		l, wd, h := d.LengthMM(), d.WidthMM(), d.HeightMM()
		parcelW, parcelL, parcelWd, parcelH = &w, &l, &wd, &h
		src, at, by := provenanceColumns(s.ParcelProvenance)
		parcelSource, parcelAt, parcelBy = &src, &at, by
	}
	listingSource, listingAt, listingBy := provenanceColumns(s.ListingProvenance)
	priceSource, priceAt, priceBy := provenanceColumns(s.PriceProvenance)
	var suspected *string
	if !s.SuspectedDuplicateOf.IsZero() {
		v := s.SuspectedDuplicateOf.String()
		suspected = &v
	}
	dismissed := make([]string, 0, len(s.DismissedDuplicates))
	for _, id := range s.DismissedDuplicates {
		dismissed = append(dismissed, id.String())
	}

	_, err := q.Exec(ctx, `
		INSERT INTO products (`+productColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
		        $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27)
		ON CONFLICT (id) DO UPDATE SET
			merchant_id = EXCLUDED.merchant_id, category = EXCLUDED.category, name = EXCLUDED.name,
			source_url = EXCLUDED.source_url, source_host = EXCLUDED.source_host,
			listing_source = EXCLUDED.listing_source, listing_at = EXCLUDED.listing_at, listing_by = EXCLUDED.listing_by,
			price_minor = EXCLUDED.price_minor, price_currency = EXCLUDED.price_currency,
			price_source = EXCLUDED.price_source, price_at = EXCLUDED.price_at, price_by = EXCLUDED.price_by,
			parcel_weight_g = EXCLUDED.parcel_weight_g, parcel_length_mm = EXCLUDED.parcel_length_mm,
			parcel_width_mm = EXCLUDED.parcel_width_mm, parcel_height_mm = EXCLUDED.parcel_height_mm,
			parcel_source = EXCLUDED.parcel_source, parcel_at = EXCLUDED.parcel_at, parcel_by = EXCLUDED.parcel_by,
			suspected_duplicate_of = EXCLUDED.suspected_duplicate_of, dismissed_duplicates = EXCLUDED.dismissed_duplicates,
			status = EXCLUDED.status, added_at = EXCLUDED.added_at,
			requested_by = EXCLUDED.requested_by, requested_variant = EXCLUDED.requested_variant`,
		s.ID.String(), s.Merchant.String(), s.Category.String(), s.Name, s.Source.String(), s.Source.Host().String(),
		listingSource, listingAt, listingBy,
		s.Price.Minor(), s.Price.Currency().Code(), priceSource, priceAt, priceBy,
		parcelW, parcelL, parcelWd, parcelH, parcelSource, parcelAt, parcelBy,
		suspected, dismissed, string(s.Status), s.AddedAt, idOrNil(s.RequestedBy), s.RequestedVariant)
	if err != nil {
		return fmt.Errorf("save product %s: %w", s.ID, err)
	}

	if _, err := q.Exec(ctx, `DELETE FROM product_variants WHERE product_id = $1`, s.ID.String()); err != nil {
		return fmt.Errorf("save product %s variants: %w", s.ID, err)
	}
	for i, v := range s.Variants {
		if _, err := q.Exec(ctx, `
			INSERT INTO product_variants (id, product_id, position, size, color, merchant_ref, added_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			v.ID.String(), s.ID.String(), i, v.Size, v.Color, v.MerchantRef, v.AddedAt); err != nil {
			return fmt.Errorf("save product %s variant %s: %w", s.ID, v.ID, err)
		}
	}
	return nil
}

func (r *ProductRepo) ByID(ctx context.Context, id catalog.ProductID) (*catalog.Product, error) {
	q := db(ctx, r.pool)
	row := q.QueryRow(ctx, `SELECT `+productColumns+` FROM products WHERE id = $1`, id.String())
	snap, err := scanProduct(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("product %s: %w", id, catalog.ErrProductNotFound)
	}
	if err != nil {
		return nil, err
	}
	return r.finish(ctx, q, snap)
}

// BySource matches the URL exactly as stored — the cheap first question of
// duplicate detection. Oldest first, so dups[0] is the original.
func (r *ProductRepo) BySource(ctx context.Context, source catalog.SourceURL) ([]*catalog.Product, error) {
	q := db(ctx, r.pool)
	rows, err := q.Query(ctx, `SELECT `+productColumns+` FROM products WHERE source_url = $1 ORDER BY added_at, id`, source.String())
	if err != nil {
		return nil, fmt.Errorf("products by source: %w", err)
	}
	var snaps []catalog.ProductSnapshot
	for rows.Next() {
		snap, err := scanProduct(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		snaps = append(snaps, snap)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("products by source: %w", err)
	}
	out := make([]*catalog.Product, 0, len(snaps))
	for _, snap := range snaps {
		p, err := r.finish(ctx, q, snap)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// finish loads the variants into the snapshot and asks the domain to rebuild.
func (r *ProductRepo) finish(ctx context.Context, q querier, snap catalog.ProductSnapshot) (*catalog.Product, error) {
	rows, err := q.Query(ctx, `SELECT id, size, color, merchant_ref, added_at FROM product_variants WHERE product_id = $1 ORDER BY position`, snap.ID.String())
	if err != nil {
		return nil, fmt.Errorf("product %s variants: %w", snap.ID, err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id, size, color, ref string
			addedAt              time.Time
		)
		if err := rows.Scan(&id, &size, &color, &ref, &addedAt); err != nil {
			return nil, fmt.Errorf("scan variant: %w", err)
		}
		vid, err := catalog.ParseVariantID(id)
		if err != nil {
			return nil, corrupt("product_variants", id, err)
		}
		snap.Variants = append(snap.Variants, catalog.VariantSnapshot{
			ID: vid, Size: size, Color: color, MerchantRef: ref, AddedAt: addedAt.UTC(),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("product %s variants: %w", snap.ID, err)
	}
	p, err := catalog.ProductFromSnapshot(snap)
	if err != nil {
		return nil, corrupt("products", snap.ID.String(), err)
	}
	return p, nil
}

// scanProduct reads the product row into a snapshot (variants come after).
func scanProduct(row pgx.Row) (catalog.ProductSnapshot, error) {
	var (
		id, merchant, category, name, sourceURL, sourceHost string
		listingSource, priceSource, priceCurrency           string
		listingAt, priceAt, addedAt                         time.Time
		listingBy, priceBy, parcelBy, parcelSource          *string
		priceMinor                                          int64
		parcelW, parcelL, parcelWd, parcelH                 *int64
		parcelAt                                            *time.Time
		suspected                                           *string
		dismissed                                           []string
		status                                              string
		requestedBy                                         *string
		requestedVariant                                    string
	)
	if err := row.Scan(&id, &merchant, &category, &name, &sourceURL, &sourceHost,
		&listingSource, &listingAt, &listingBy,
		&priceMinor, &priceCurrency, &priceSource, &priceAt, &priceBy,
		&parcelW, &parcelL, &parcelWd, &parcelH, &parcelSource, &parcelAt, &parcelBy,
		&suspected, &dismissed, &status, &addedAt, &requestedBy, &requestedVariant); err != nil {
		return catalog.ProductSnapshot{}, err
	}
	_ = sourceHost // derived from source_url on load; stored for indexing/joins only

	fail := func(err error) (catalog.ProductSnapshot, error) {
		return catalog.ProductSnapshot{}, corrupt("products", id, err)
	}
	pid, err := catalog.ParseProductID(id)
	if err != nil {
		return fail(err)
	}
	mid, err := catalog.ParseMerchantID(merchant)
	if err != nil {
		return fail(err)
	}
	code, err := catalog.ParseCategoryCode(category)
	if err != nil {
		return fail(err)
	}
	source, err := catalog.ParseSourceURL(sourceURL)
	if err != nil {
		return fail(err)
	}
	cur, err := shared.CurrencyFromCode(priceCurrency)
	if err != nil {
		return fail(err)
	}
	listingProv, err := provenanceFrom(listingSource, listingAt, listingBy)
	if err != nil {
		return fail(err)
	}
	priceProv, err := provenanceFrom(priceSource, priceAt, priceBy)
	if err != nil {
		return fail(err)
	}
	snap := catalog.ProductSnapshot{
		ID: pid, Merchant: mid, Category: code, Name: name, Source: source,
		ListingProvenance: listingProv,
		Price:             shared.NewMoney(priceMinor, cur), PriceProvenance: priceProv,
		Status: catalog.ProductStatus(status), AddedAt: addedAt.UTC(),
		RequestedVariant: requestedVariant,
	}
	if requestedBy != nil {
		id, err := shared.ParseID(*requestedBy)
		if err != nil {
			return fail(err)
		}
		snap.RequestedBy = id
	}
	if parcelW != nil && parcelL != nil && parcelWd != nil && parcelH != nil && parcelSource != nil && parcelAt != nil {
		weight, err := shared.NewWeight(*parcelW)
		if err != nil {
			return fail(err)
		}
		if *parcelL < 0 || *parcelWd < 0 || *parcelH < 0 {
			return fail(errors.New("negative parcel dimension"))
		}
		spec, err := shared.NewParcelSpec(weight, shared.NewDimensionsMM(*parcelL, *parcelWd, *parcelH))
		if err != nil {
			return fail(err)
		}
		prov, err := provenanceFrom(*parcelSource, *parcelAt, parcelBy)
		if err != nil {
			return fail(err)
		}
		snap.Parcel, snap.ParcelProvenance = spec, prov
	}
	if suspected != nil {
		other, err := catalog.ParseProductID(*suspected)
		if err != nil {
			return fail(err)
		}
		snap.SuspectedDuplicateOf = other
	}
	for _, d := range dismissed {
		other, err := catalog.ParseProductID(d)
		if err != nil {
			return fail(err)
		}
		snap.DismissedDuplicates = append(snap.DismissedDuplicates, other)
	}
	return snap, nil
}

// provenanceColumns / provenanceFrom are the two halves of Provenance ↔ three
// columns. by is NULL when nobody on our side was involved.
func provenanceColumns(p catalog.Provenance) (source string, at time.Time, by *string) {
	source, at = string(p.Source()), p.At()
	if !p.By().IsZero() {
		v := p.By().String()
		by = &v
	}
	return source, at, by
}

func provenanceFrom(source string, at time.Time, by *string) (catalog.Provenance, error) {
	var operator shared.OperatorID
	if by != nil {
		id, err := shared.ParseOperatorID(*by)
		if err != nil {
			return catalog.Provenance{}, err
		}
		operator = id
	}
	return catalog.NewProvenance(catalog.SourcingMode(source), at.UTC(), operator)
}
