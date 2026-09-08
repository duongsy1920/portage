package catalogapp

import (
	"context"
	"fmt"

	"github.com/duongsy/portage/internal/domain/catalog"
	"github.com/duongsy/portage/internal/domain/shared"
)

// DefineCategory is the command behind "an operator sets what a kind of goods
// is like": the default parcel a quote uses before the real thing is weighed,
// and what it may not do in transit.
type DefineCategory struct {
	Code         string // as typed; the domain normalises and validates it
	Estimate     shared.ParcelSpec
	Restrictions []catalog.Restriction
}

// DefineCategoryHandler saves the policy and announces it. Categories are
// reference data other contexts copy (pricing needs the goods class and the
// default parcel); saving straight to the repository would tell nobody.
//
// Redefining replaces the value object and announces again. The dev seed in
// platform/wire runs this on every start-up, so the announcement repeats —
// which is fine, and deliberate: consumers must be idempotent anyway
// (at-least-once, DDD.md §27).
type DefineCategoryHandler struct {
	deps Deps
}

func NewDefineCategoryHandler(d Deps) *DefineCategoryHandler {
	mustHave("DefineCategoryHandler", map[string]any{
		"Clock": d.Clock, "UoW": d.UoW, "Categories": d.Categories, "Outbox": d.Outbox,
	})
	return &DefineCategoryHandler{deps: d}
}

func (h *DefineCategoryHandler) Handle(ctx context.Context, cmd DefineCategory) error {
	now := h.deps.Clock.Now()
	code, err := catalog.ParseCategoryCode(cmd.Code)
	if err != nil {
		return err
	}
	policy, err := catalog.NewCategoryPolicy(code, cmd.Estimate, cmd.Restrictions)
	if err != nil {
		return err
	}
	return h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		if err := h.deps.Categories.Save(ctx, policy); err != nil {
			return fmt.Errorf("save category: %w", err)
		}
		if err := h.deps.Outbox.Append(ctx, []shared.Event{policy.Defined(now)}); err != nil {
			return fmt.Errorf("outbox: %w", err)
		}
		return nil
	})
}
