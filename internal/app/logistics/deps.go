// Package logisticsapp holds the use cases of logistics: expect a parcel when
// procurement buys, weigh it when it arrives, box parcels into a batch, and
// ship the batch — splitting the carrier's invoice over the orders inside.
//
// [PHP] MessageHandlers của bundle Logistics; ShipBatch là handler duy nhất
// [PHP] đụng hai aggregate (batch + các parcel) trong một transaction — có
// [PHP] lý do, xem ShipBatchHandler.
package logisticsapp

import (
	"context"
	"fmt"

	"github.com/duongsy/portage/internal/app"
	"github.com/duongsy/portage/internal/domain/logistics"
)

type Deps struct {
	Clock   app.Clock
	UoW     app.UnitOfWork
	Outbox  app.Outbox
	Parcels logistics.ParcelRepository
	Batches logistics.BatchRepository
	Lanes   logistics.LaneRuleRepository
}

func mustHave(handler string, deps map[string]any) {
	app.MustHave("logisticsapp: "+handler, deps)
}

func (d Deps) mutateParcel(ctx context.Context, id logistics.ParcelID, fn func(p *logistics.Parcel) error) error {
	return d.UoW.InTx(ctx, func(ctx context.Context) error {
		p, err := d.Parcels.ByID(ctx, id)
		if err != nil {
			return err
		}
		if err := fn(p); err != nil {
			return err
		}
		if err := d.Parcels.Save(ctx, p); err != nil {
			return fmt.Errorf("save parcel: %w", err)
		}
		if err := d.Outbox.Append(ctx, p.PullEvents()); err != nil {
			return fmt.Errorf("outbox: %w", err)
		}
		return nil
	})
}

func (d Deps) mutateBatch(ctx context.Context, id logistics.BatchID, fn func(b *logistics.ConsolidationBatch) error) error {
	return d.UoW.InTx(ctx, func(ctx context.Context) error {
		b, err := d.Batches.ByID(ctx, id)
		if err != nil {
			return err
		}
		if err := fn(b); err != nil {
			return err
		}
		if err := d.Batches.Save(ctx, b); err != nil {
			return fmt.Errorf("save batch: %w", err)
		}
		if err := d.Outbox.Append(ctx, b.PullEvents()); err != nil {
			return fmt.Errorf("outbox: %w", err)
		}
		return nil
	})
}
