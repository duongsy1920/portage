package worker

import (
	"context"
	"log"
	"time"
)

// Pass is one run of periodic work: it does what it can and reports how much.
// pricingapp.ExpireQuotesHandler.Handle is the first one.
type Pass func(ctx context.Context) (int, error)

// Sweeper runs a Pass on a ticker until the context ends — the same shape as
// Relay.Run, for work that no event triggers.
//
// The Relay reacts to rows somebody wrote; a Sweeper reacts to TIME PASSING.
// That is the whole difference, and it is why the expiry of a quote is not an
// event anyone publishes: nothing happened, except that nothing happened.
//
// A failing pass is logged and retried on the next tick, never fatal: the
// process exists to relay the outbox, and a sweep that cannot reach the
// database this second must not take that down with it.
//
// [PHP] Một cron job (`* * * * * bin/console app:…`), nhưng trong process sẵn
// [PHP] có: không cần cài crontab lên máy chủ để tính năng chạy được.
type Sweeper struct {
	name     string
	pass     Pass
	interval time.Duration
	log      *log.Logger
}

// NewSweeper builds one. name appears in every log line, so a process running
// two sweeps stays readable.
func NewSweeper(name string, p Pass, interval time.Duration, l *log.Logger) *Sweeper {
	if p == nil {
		panic("worker: Sweeper wired without a Pass")
	}
	if interval <= 0 {
		interval = time.Minute
	}
	if l == nil {
		l = log.Default()
	}
	return &Sweeper{name: name, pass: p, interval: interval, log: l}
}

// Run WAITS one interval before the first pass, unlike Relay.Run which runs
// immediately. The relay has rows waiting for it the moment it starts; a sweep
// has only a clock, and doing a scan in the same breath as start-up buys
// nothing while making a boot loop scan the table over and over.
//
// Run ticks until ctx ends. It logs only when a pass did something or failed:
// a sweep that finds nothing every minute for a week must not fill the log
// with a week of "nothing".
func (s *Sweeper) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
		n, err := s.pass(ctx)
		if n > 0 {
			s.log.Printf("worker: sweep %s: %d", s.name, n)
		}
		if err != nil && ctx.Err() == nil {
			s.log.Printf("worker: sweep %s: %v", s.name, err)
		}
	}
}
