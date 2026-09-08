package catalogapp

import (
	"context"
	"fmt"

	"github.com/duongsy/portage/internal/domain/catalog"
)

// PublishProductHandler is the plainest use case shape: load one aggregate,
// ask it to do one thing, save it, hand over its events. It adds no rule of
// its own — every refusal is Product.Publish's, passed through so the adapter
// can tell the operator exactly what is missing.
//
// [PHP] Đây là cái "load-edit-flush" quen thuộc của Symfony:
// [PHP]   $p = $repo->find($id); $p->publish($now); $em->flush();
// [PHP] — chỉ khác thứ tự tường minh: Save rồi mới PullEvents, và $now đến từ
// [PHP] Clock của tầng app, không phải new \DateTimeImmutable() trong entity.
type PublishProductHandler struct {
	deps Deps
}

func NewPublishProductHandler(d Deps) *PublishProductHandler {
	mustHave("PublishProductHandler", map[string]any{
		"Clock": d.Clock, "UoW": d.UoW, "Products": d.Products, "Outbox": d.Outbox,
	})
	return &PublishProductHandler{deps: d}
}

func (h *PublishProductHandler) Handle(ctx context.Context, id catalog.ProductID) error {
	now := h.deps.Clock.Now()

	return h.deps.UoW.InTx(ctx, func(ctx context.Context) error {
		p, err := h.deps.Products.ByID(ctx, id)
		if err != nil {
			return err
		}
		if err := p.Publish(now); err != nil {
			return err // the domain's refusal, with its reason
		}
		if err := h.deps.Products.Save(ctx, p); err != nil {
			return fmt.Errorf("save product: %w", err)
		}
		if err := h.deps.Outbox.Append(ctx, p.PullEvents()); err != nil {
			return fmt.Errorf("outbox: %w", err)
		}
		return nil
	})
}
