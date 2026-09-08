package app

import "fmt"

// MustHave panics, naming the handler and the dependency, when a required
// dependency is nil. Every use-case constructor calls it once: wiring is
// programmer input, and a missing dependency is a start-up bug, not a
// request-time error (convention 1).
//
// [PHP] Bằng với lỗi compile container của Symfony khi thiếu argument — ở đây
// [PHP] nổ lúc dựng handler trong main(), cùng thời điểm, cùng mức độ.
func MustHave(handler string, deps map[string]any) {
	for name, v := range deps {
		if v == nil {
			panic(fmt.Sprintf("%s wired without %s", handler, name))
		}
	}
}
