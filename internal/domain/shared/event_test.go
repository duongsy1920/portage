package shared_test

import (
	"testing"
	"time"

	"github.com/duongsy/portage/internal/domain/shared"
)

type depositPaid struct {
	at time.Time
}

func (depositPaid) EventName() string {
	return "ordering.deposit_paid"
}
func (e depositPaid) OccurredAt() time.Time {
	return e.at
}

// How an aggregate root will use it: embed Events, call Record from the
// method that changed state, let the application layer Pull after saving.
type fakeOrder struct {
	shared.Events
}

func TestEvents_recordThenPullInOrderAndClear(t *testing.T) {
	var o fakeOrder
	t0 := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)
	o.Record(depositPaid{at: t0})
	o.Record(depositPaid{at: t0.Add(time.Minute)})

	got := o.PullEvents()
	if len(got) != 2 || got[0].OccurredAt() != t0 || got[1].EventName() != "ordering.deposit_paid" {
		t.Fatalf("PullEvents() = %v", got)
	}
	if again := o.PullEvents(); len(again) != 0 {
		t.Fatalf("second PullEvents() must be empty, got %d", len(again))
	}
}

func TestEvents_zeroValuePullsNothing(t *testing.T) {
	var e shared.Events
	if got := e.PullEvents(); len(got) != 0 {
		t.Fatalf("zero Events pulled %d events", len(got))
	}
}
