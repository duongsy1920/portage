// Command worker relays the outbox (DDD.md §27, §30 layer 5): it reads rows
// the API wrote, publishes them to the subscribers wire.Subscribe installs
// (today: pricing's projector), marks them sent.
//
//	go run ./cmd/worker -dsn postgres://…
//
// -dsn is REQUIRED. There is no in-memory mode on purpose: an in-memory outbox
// lives inside the api process, so a separate worker would have nothing to
// read — and a worker that starts, does nothing and exits 0 is worse than one
// that refuses to start (SETUP.md §9 đợt 8).
//
// The graph comes from internal/platform/wire, the same function cmd/api
// uses, so the two binaries cannot disagree about how the pieces fit.
//
// [PHP] `bin/console messenger:consume` — nhưng cho outbox.
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	pricingapp "github.com/duongsy/portage/internal/app/pricing"
	"github.com/duongsy/portage/internal/platform/clock"
	"github.com/duongsy/portage/internal/platform/wire"
	"github.com/duongsy/portage/internal/worker"
)

func main() {
	dsn := flag.String("dsn", "", "Postgres DSN (required: the worker reads the shared outbox)")
	every := flag.Duration("every", time.Second, "pause between passes")
	batch := flag.Int("batch", 100, "rows per pass")
	sweep := flag.Duration("sweep", time.Minute, "pause between quote-expiry sweeps; 0 turns the sweep off")
	sweepBatch := flag.Int("sweep-batch", 100, "quotes expired per sweep")
	flag.Parse()
	if *dsn == "" {
		log.Fatal("cmd/worker: -dsn is required; there is no in-memory outbox to read from another process")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, closeFn, err := wire.Postgres(ctx, clock.System{}, *dsn)
	if err != nil {
		log.Fatalf("cmd/worker: %v", err)
	}
	defer closeFn()

	// Subscribers: every consumer in the system (wire.Subscribe is the routing
	// table), plus one log line per event so "another context hears it" is
	// visible in the terminal.
	bus := worker.NewBus()
	wire.Subscribe(bus, g)
	bus.Subscribe("*", func(_ context.Context, e worker.Entry) error {
		log.Printf("event #%d %s at %s %s", e.ID, e.Name, e.OccurredAt.Format(time.RFC3339), e.Payload)
		return nil
	})

	// The expiry sweep. Nothing publishes "this quote got old", so unlike
	// every other consumer in the system this one is driven by the clock
	// (pricingapp.ExpireQuotesHandler). It runs beside the relay rather than
	// in its own binary: it is a few queries a minute, and a second process
	// would be a second thing to deploy for no gain.
	if *sweep > 0 {
		expire := pricingapp.NewExpireQuotesHandler(g.Pricing, *sweepBatch)
		go func() {
			_ = worker.NewSweeper("quote-expiry", expire.Handle, *sweep, nil).Run(ctx) // ends with ctx
		}()
		log.Printf("portage worker: quote-expiry sweep every %s, %d quotes per pass", *sweep, *sweepBatch)
	}

	relay := worker.New(worker.Deps{
		Source: g.Source, Publisher: bus, UoW: g.Catalog.UoW, Clock: g.Catalog.Clock,
		Batch: *batch, Interval: *every,
	})
	log.Printf("portage worker: relaying outbox every %s, %d rows per pass", *every, *batch)
	if err := relay.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("cmd/worker: %v", err)
	}
	log.Print("portage worker: stopped")
}
