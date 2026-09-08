package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// CategoryRepo implements catalog.CategoryRepository. CategoryPolicy is a value
// object with a natural key, so Save replaces the whole row under its code.
type CategoryRepo struct {
	pool *pgxpool.Pool
}

var _ catalog.CategoryRepository = (*CategoryRepo)(nil)

func NewCategoryRepo(pool *pgxpool.Pool) *CategoryRepo {
	return &CategoryRepo{pool: pool}
}

const categoryColumns = `code, est_weight_g, est_length_mm, est_width_mm, est_height_mm, restrictions`

func (r *CategoryRepo) Save(ctx context.Context, p catalog.CategoryPolicy) error {
	est := p.DefaultParcelSpec()
	d := est.Dimensions()
	restrictions := make([]string, 0, len(p.Restrictions()))
	for _, x := range p.Restrictions() {
		restrictions = append(restrictions, string(x))
	}
	_, err := db(ctx, r.pool).Exec(ctx, `
		INSERT INTO categories (`+categoryColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (code) DO UPDATE SET
			est_weight_g = EXCLUDED.est_weight_g, est_length_mm = EXCLUDED.est_length_mm,
			est_width_mm = EXCLUDED.est_width_mm, est_height_mm = EXCLUDED.est_height_mm,
			restrictions = EXCLUDED.restrictions`,
		p.Code().String(), est.Weight().Grams(), d.LengthMM(), d.WidthMM(), d.HeightMM(), restrictions)
	if err != nil {
		return fmt.Errorf("save category %s: %w", p.Code(), err)
	}
	return nil
}

func (r *CategoryRepo) ByCode(ctx context.Context, code catalog.CategoryCode) (catalog.CategoryPolicy, error) {
	rows, err := db(ctx, r.pool).Query(ctx, `SELECT `+categoryColumns+` FROM categories WHERE code = $1`, code.String())
	if err != nil {
		return catalog.CategoryPolicy{}, fmt.Errorf("category %s: %w", code, err)
	}
	all, err := collectCategories(rows)
	if err != nil {
		return catalog.CategoryPolicy{}, err
	}
	if len(all) == 0 {
		return catalog.CategoryPolicy{}, fmt.Errorf("category %s: %w", code, catalog.ErrCategoryNotFound)
	}
	return all[0], nil
}

func (r *CategoryRepo) All(ctx context.Context) ([]catalog.CategoryPolicy, error) {
	rows, err := db(ctx, r.pool).Query(ctx, `SELECT `+categoryColumns+` FROM categories ORDER BY code`)
	if err != nil {
		return nil, fmt.Errorf("categories: %w", err)
	}
	return collectCategories(rows)
}

func collectCategories(rows pgx.Rows) ([]catalog.CategoryPolicy, error) {
	defer rows.Close()
	var out []catalog.CategoryPolicy
	for rows.Next() {
		var (
			code         string
			g, l, w, h   int64
			restrictions []string
		)
		if err := rows.Scan(&code, &g, &l, &w, &h, &restrictions); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		p, err := categoryFrom(code, g, l, w, h, restrictions)
		if err != nil {
			return nil, corrupt("categories", code, err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func categoryFrom(code string, g, l, w, h int64, restrictions []string) (catalog.CategoryPolicy, error) {
	c, err := catalog.ParseCategoryCode(code)
	if err != nil {
		return catalog.CategoryPolicy{}, err
	}
	weight, err := shared.NewWeight(g)
	if err != nil {
		return catalog.CategoryPolicy{}, err
	}
	if l < 0 || w < 0 || h < 0 {
		return catalog.CategoryPolicy{}, errors.New("negative dimension")
	}
	spec, err := shared.NewParcelSpec(weight, shared.NewDimensionsMM(l, w, h))
	if err != nil {
		return catalog.CategoryPolicy{}, err
	}
	rs := make([]catalog.Restriction, 0, len(restrictions))
	for _, x := range restrictions {
		rs = append(rs, catalog.Restriction(x))
	}
	return catalog.NewCategoryPolicy(c, spec, rs)
}
