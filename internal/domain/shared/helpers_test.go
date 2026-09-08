package shared_test

import "testing"

// mustPanic asserts that fn panics. Symfony: the closest thing is
// $this->expectException(\LogicException::class) — a panic here marks a
// programming error, not a business rule violation.
func mustPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic")
		}
	}()
	fn()
}
