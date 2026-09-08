// Package clock implements app.Clock: the real one for production, a fixed one
// for tests. It lives in platform/, not domain/ — the domain never sees a clock.
package clock

import "time"

// System is the wall clock. Use it in main(); nowhere else.
type System struct{}

func (System) Now() time.Time {
	return time.Now()
}

// Fixed is a clock that says what the test tells it to. Advance moves it.
//
// [PHP] Symfony\Component\Clock\MockClock — cùng vai.
type Fixed struct {
	at time.Time
}

func FixedAt(t time.Time) *Fixed {
	return &Fixed{at: t}
}

func (f *Fixed) Now() time.Time {
	return f.at
}

func (f *Fixed) Advance(d time.Duration) {
	f.at = f.at.Add(d)
}
