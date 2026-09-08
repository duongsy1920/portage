package reportingapp

import "github.com/duongsy/portage/internal/app"

// Deps wires the read side (convention 10). Note what is NOT here: no Clock
// and no Outbox. A read model does not decide when anything happened — it
// copies the timestamps the write side already put in the events — and it
// announces nothing, because nothing listens to a screen.
type Deps struct {
	UoW       app.UnitOfWork
	Summaries OrderSummaryRepository
	Names     ProductNames
	Worklist  ProductWorklistRepository
}

func mustHave(handler string, deps map[string]any) {
	app.MustHave("reportingapp: "+handler, deps)
}
