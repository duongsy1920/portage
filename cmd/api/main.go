// Command api runs the Portage HTTP API.
//
//	go run ./cmd/api                      in-memory: everything works, nothing survives a restart
//	go run ./cmd/api -dsn postgres://…    Postgres: migrates on start, outbox rows for cmd/worker
//	go run ./cmd/api -web ./web           …and serves the browser console at /ui/
//
// main() does two things only: parse flags and serve. How the pieces fit is
// internal/platform/wire, shared with cmd/worker so the two binaries cannot
// drift apart.
//
// In memory there is no second process to relay the outbox — it lives inside
// this one — so the relay runs here, in a goroutine, with the same
// subscriptions cmd/worker installs. On Postgres it does NOT: that is the
// worker's job, and running both would make every event arrive twice
// (harmlessly, but pointlessly).
//
// [PHP] public/index.php: nhận request, chọn kernel, chạy. Config nằm chỗ khác.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	httpapi "github.com/duongsy/portage/internal/adapter/http"
	"github.com/duongsy/portage/internal/adapter/postgres"
	pricingapp "github.com/duongsy/portage/internal/app/pricing"
	"github.com/duongsy/portage/internal/domain/shared"
	"github.com/duongsy/portage/internal/platform/auth"
	"github.com/duongsy/portage/internal/platform/clock"
	"github.com/duongsy/portage/internal/platform/wire"
	"github.com/duongsy/portage/internal/worker"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dsn := flag.String("dsn", "", "Postgres DSN; empty runs in memory")
	bootstrap := flag.String("bootstrap-operator-token", "",
		"Postgres only: create the FIRST operator token, and only while api_tokens is empty")
	web := flag.String("web", "", "serve the browser console from this directory at /ui/ (dev)")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var g wire.Graph
	if *dsn == "" {
		g = wire.Memory(clock.System{}) // the ONE place the real clock is chosen
		log.Printf("portage api: in-memory adapters (dev); data is lost on exit")

		bus := worker.NewBus()
		wire.Subscribe(bus, g)
		relay := worker.New(worker.Deps{
			Source: g.Source, Publisher: bus, UoW: g.Catalog.UoW, Clock: g.Catalog.Clock, Interval: 200 * time.Millisecond,
		})
		go func() {
			_ = relay.Run(ctx) // ends with ctx; errors are logged by Run itself
		}()
		log.Printf("portage api: in-process outbox relay every 200ms (no cmd/worker in memory mode)")

		// Same reason: with no cmd/worker there is nothing else to expire
		// stale quotes, and a dev run where a quote is valid forever would
		// behave differently from production in exactly the rule the quote
		// exists to enforce.
		expire := pricingapp.NewExpireQuotesHandler(g.Pricing, 100)
		go func() {
			_ = worker.NewSweeper("quote-expiry", expire.Handle, time.Minute, nil).Run(ctx)
		}()
	} else {
		connectCtx, cancelConnect := context.WithTimeout(ctx, 10*time.Second)
		defer cancelConnect()
		var closeFn func()
		var err error
		g, closeFn, err = wire.Postgres(connectCtx, clock.System{}, *dsn)
		if err != nil {
			log.Fatalf("portage api: %v", err)
		}
		defer closeFn()
		log.Printf("portage api: postgres, migrated; run cmd/worker to relay the outbox")

		if *bootstrap != "" {
			if err := bootstrapOperator(ctx, *dsn, *bootstrap); err != nil {
				log.Fatalf("portage api: %v", err)
			}
		}
	}
	if *bootstrap != "" && *dsn == "" {
		log.Printf("portage api: -bootstrap-operator-token ignored in memory mode; use %q", wire.DevOperatorToken)
	}

	var handler http.Handler = httpapi.NewHandler(httpapi.Deps{Catalog: g.Catalog, Pricing: g.Pricing, Ordering: g.Ordering, Procurement: g.Procurement, Logistics: g.Logistics, Reporting: g.Reporting, Extractor: g.Extractor, Auth: g.Auth, Tokens: g.Tokens, Registry: g.Registry})
	if *web != "" {
		handler = withConsole(handler, *web, *addr)
	}

	srv := &http.Server{
		Addr:              *addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("portage api listening on %s", *addr)
	log.Fatal(srv.ListenAndServe())
}

// withConsole puts the browser console next to the API, on the SAME ORIGIN.
//
// That is the whole reason it lives in this binary instead of a second dev
// server: same origin means the browser sends no CORS preflight, so the API
// needs no cross-origin middleware for a front end to exist, and a bearer
// token in fetch() reaches exactly the authenticate() that curl reaches.
//
// The API keeps its own mux and its own auth — this only mounts a file server
// beside it. Those files are public on purpose: the console ships no secret,
// it asks for the token at run time.
//
// [PHP] Tương đương việc để `public/` và route API cùng một Kernel: cùng
// [PHP] origin nên không phải cấu hình nelmio/cors cho môi trường dev.
func withConsole(api http.Handler, dir, addr string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir(dir))))
	// "/{$}" matches the root and NOTHING else, so it cannot shadow an API
	// route the way a bare "/" handler would.
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/", http.StatusFound)
	})
	mux.Handle("/", api)
	log.Printf("portage api: screens at http://localhost%s/ui/ (files from %s)", addr, dir)
	return mux
}

// bootstrapOperator creates the first operator token — and ONLY while there is
// no token at all.
//
// The guard is the whole point. Seeding a fixed credential into a database
// that already has users is a back door: anyone who reads a deployment script,
// a shell history or this source could then call the API as staff. On an empty
// database there is nobody to impersonate yet, so the same act is just the
// first key being cut.
//
// It opens its own pool: this runs once at start-up, and threading a pool out
// of wire.Postgres only for this would widen that function's contract for a
// caller it does not otherwise have.
func bootstrapOperator(ctx context.Context, dsn, token string) error {
	pool, err := postgres.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	defer pool.Close()

	repo := postgres.NewTokenRepo(pool)
	empty, err := repo.IsEmpty(ctx)
	if err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	if !empty {
		log.Printf("portage api: -bootstrap-operator-token ignored; api_tokens is not empty")
		return nil
	}
	operator := auth.MustPrincipal(auth.Operator, shared.NewID())
	if err := repo.Issue(ctx, token, operator, "bootstrap", time.Now()); err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}
	// The id, not the token: the token is already in the operator's hands, and
	// a log file is the last place it should also be.
	log.Printf("portage api: first operator created, id %s", operator.ID())
	return nil
}
