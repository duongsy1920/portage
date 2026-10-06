/**
 * FACTS ABOUT THE PORTAGE REPOSITORY.
 *
 * This file is the single source of truth for every episode. Nothing here is
 * invented: every snippet was copied out of the Go repository, and every
 * snippet carries the path and line it came from so a viewer can open it.
 *
 * Rule for maintaining this file: if a number or a snippet changes in the Go
 * project, change it HERE, not inside a scene. A scene must never hard-code a
 * fact about the codebase.
 *
 * Counts verified on 2026-09-17 with grep and `go test`.
 */

export const REPO = {
  name: "Portage",
  oneLine: "Mua hộ xuyên biên giới, Mỹ → Việt Nam",
  module: "github.com/duongsy/portage",

  /** Đếm mux.HandleFunc trong adapter/http/server.go. */
  routes: 41,
  /** Đếm bus.Subscribe trong platform/wire/subscribe.go. */
  subscribeLines: 35,
  /** Đếm trong mọi file events.go của năm context. */
  events: 37,
  /** aggregate roots, counted by hand across the five contexts */
  aggregateRoots: 7,
  /** Đếm method Handle trong internal/app. */
  useCases: 30,
  /** Đếm dòng === RUN của go test -v. */
  /** Đếm lại 06/10: `go test ./... -count=1 -v` có PORTAGE_TEST_DSN → 297 PASS + 1 SKIP. */
  tests: 298,
  /** the subset that needs no Docker. Đếm lại 06/10: `--- PASS` cấp cao nhất, không DSN; verify-numbers.sh kiểm. */
  testsNoDocker: 266,
  /** Đếm func TestDecision_ trong domain/decisions_test.go. */
  guards: 7,
  migrations: 13,
  /** direct dependencies in go.mod */
  directDeps: 2,
} as const;

/** The five write-side contexts plus the read side, in lifecycle order. */
export const CONTEXTS = [
  {
    key: "catalog",
    short: "món gì, shop nào",
    dir: "internal/domain/catalog",
    roots: "Merchant · Product ⊃ Variant",
    answers: "bán món gì, của shop nào, size nào",
  },
  {
    key: "pricing",
    short: "giá bao nhiêu",
    dir: "internal/domain/pricing",
    roots: "Quote · Reconciliation",
    answers: "giá bao nhiêu, cuối cùng lời hay lỗ",
  },
  {
    key: "ordering",
    short: "khách cam kết gì",
    dir: "internal/domain/ordering",
    roots: "CustomerOrder",
    answers: "khách cam kết gì, huỷ thì lấy lại gì",
  },
  {
    key: "procurement",
    short: "ai đi mua",
    dir: "internal/domain/procurement",
    roots: "PurchaseTask",
    answers: "ai đi mua, trả thật bao nhiêu",
  },
  {
    key: "logistics",
    short: "thùng, cước",
    dir: "internal/domain/logistics",
    roots: "Parcel · ConsolidationBatch",
    answers: "thùng cân bao nhiêu, cước chia sao",
  },
] as const;

/**
 * The real timeline of one order, from the business rules in CLAUDE.md and
 * the flow in docs/FLOW-ORDER.md. Days are the realistic shape of the
 * business, not timestamps from a database row.
 */
export const ORDER_TIMELINE = [
  { day: "Ngày 1", what: "Khách dán link", ctx: "catalog" },
  { day: "Ngày 1", what: "Báo giá, khách đồng ý", ctx: "pricing" },
  { day: "Ngày 1", what: "Cọc 50 %", ctx: "ordering" },
  { day: "Ngày 2", what: "Nhân viên mua ở shop Mỹ", ctx: "procurement" },
  { day: "Ngày 9", what: "Hàng về kho Denver", ctx: "logistics" },
  { day: "Ngày 20", what: "Gom lô, bay về", ctx: "logistics" },
  { day: "Ngày 30", what: "Giao tận nhà", ctx: "ordering" },
] as const;

/** The golden numbers. scripts/smoke.sh asserts these on every run. */
export const GOLDEN = {
  itemUSD: "150.00",
  taxUSD: "13.22",
  freightQuotedUSD: "25.00",
  subtotalUSD: "188.22",
  // Dấu chấm phân cách nghìn: đúng dạng API nhận và trả với Accept-Language: vi.
  fx: "26.000",
  serviceFeeVND: "500.000",
  totalVND: "5.393.720",
  depositVND: "2.696.860",
  actualGoodsUSD: "163.22",
  actualFreightUSD: "27.50",
  varianceUSD: "−2.50",
} as const;

export type Snippet = {
  path: string;
  /** Line in that file. Omitted for an illustrative PHP snippet, which has no file. */
  line?: number;
  /**
   * "go" means copied verbatim out of the repository.
   * "php" means written for the comparison — the viewer's own world, never a
   * claim about this codebase. Those snippets say so in their header.
   */
  lang: "go" | "php";
  /** Real lines from the file. `…` marks fields elided to fit the frame. */
  code: string[];
};

/**
 * Snippets used by episode 1. Copied verbatim; where a struct was too tall for
 * a vertical frame, fields were REMOVED (never renamed) and the cut marked.
 */
export const SNIPPETS = {
  /** Season 3: the outbox write, in the same transaction as the aggregate. */
  outboxAppend: {
    path: "internal/adapter/postgres/outbox.go",
    line: 35,
    lang: "go",
    code: [
      "func (o *Outbox) Append(ctx context.Context,",
      "    events []shared.Event) error {",
      "    q := db(ctx, o.pool)",
      "    for _, ev := range events {",
      "        …",
      "        if _, err := q.Exec(ctx, `INSERT INTO outbox",
      "            (event_name, occurred_at, payload)",
      "            VALUES ($1, $2, $3)`,",
      "            …",
      "    }",
      "    return nil",
      "}",
    ],
  },
  /** Season 3: quote versus actual, and the number that answers the business. */
  reconciliation: {
    path: "internal/domain/pricing/reconciliation.go",
    line: 21,
    lang: "go",
    code: [
      "type Reconciliation struct {",
      "    Order shared.ID",
      "    Quote shared.ID",
      "",
      "    QuotedGoods      shared.Money",
      "    QuotedFreight    shared.Money",
      "    …",
      "    ActualGoods      shared.Money",
      "    ActualFreight    shared.Money",
      "    …",
      "}",
    ],
  },
  customerOrder: {
    path: "internal/domain/ordering/order.go",
    line: 110,
    lang: "go",
    code: [
      "type CustomerOrder struct {",
      "    shared.Events",
      "    id       OrderID",
      "    …",
      "    customer shared.ID",
      "    …",
      "    deposit  shared.Money",
      "    status   OrderStatus",
      "    …",
      "}",
    ],
  },

  purchaseTask: {
    path: "internal/domain/procurement/task.go",
    line: 98,
    lang: "go",
    code: [
      "type PurchaseTask struct {",
      "    shared.Events",
      "    id       TaskID",
      "    order    shared.ID",
      "    …",
      "    status   TaskStatus",
      "    receipt  PurchaseReceipt",
      "    …",
      "}",
    ],
  },

  guard: {
    path: "internal/domain/decisions_test.go",
    line: 260,
    lang: "go",
    code: [
      "func TestDecision_boundedContexts",
      "         DoNotImportEachOther(t *testing.T) {",
      "    …",
      "    t.Errorf(",
      '      "%s (context %q) imports context %q.\\n"+',
      "      …)",
      "}",
    ],
  },

  subscribe: {
    path: "internal/platform/wire/subscribe.go",
    line: 57,
    lang: "go",
    code: [
      'bus.Subscribe("ordering.deposit_paid",',
      "    on(buyer.OnDepositPaid))",
    ],
  },
} as const satisfies Record<string, Snippet>;

/**
 * Snippets for SEASON 1 — the Go language itself.
 *
 * Each one is real code from this repository, so the language lesson and the
 * project lesson never drift apart: the same file a learner reads in episode 1
 * is the file they will open again in season 3.
 */
export const GO_SNIPPETS = {
  /** Episode 3: two return values, and the wrap that keeps the cause alive. */
  parseMoney: {
    path: "internal/domain/shared/money.go",
    line: 102,
    lang: "go",
    code: [
      "func ParseMoney(s string, c Currency) (Money, error) {",
      "    minor, err := parseDecimal(s, int(c.exponent))",
      "    if err != nil {",
      '        return Money{}, fmt.Errorf("parse %q as %s: %w",',
      "            s, c.code, err)",
      "    }",
      "    return NewMoney(minor, c), nil",
      "}",
    ],
  },
  /** One sentinel per reason: a refusal names its cause. */
  sentinels: {
    path: "internal/domain/ordering/errors.go",
    line: 7,
    lang: "go",
    code: [
      "var (",
      '    ErrInvalidOrder     = errors.New("invalid order")',
      '    ErrQuoteNotAccepted = errors.New("quote is not accepted")',
      '    ErrQuoteAlreadyUsed = errors.New("quote already has an order")',
      "    …",
      '    ErrVariantUnknown = errors.New("variant is unknown here")',
      "    …",
      ")",
    ],
  },
  /** Reading the cause back out, through however many wraps. */
  errorsIs: {
    path: "internal/app/pricing/accept_quote.go",
    line: 35,
    lang: "go",
    code: [
      "if err := q.Accept(now); err != nil {",
      "    if !errors.Is(err, pricing.ErrQuoteExpired) {",
      "        return err // không đổi gì; quay đầu, báo lên",
      "    }",
      "    refused = err // đã đổi: lưu lại, rồi mới báo",
      "}",
    ],
  },
  /** Episode 4: a value receiver — Go copies the struct, so the original is safe. */
  valueReceiver: {
    path: "internal/domain/shared/money.go",
    line: 201,
    lang: "go",
    code: [
      "func (m Money) Add(o Money) (Money, error) {",
      '    if err := m.combinable(o, "add"); err != nil {',
      "        return Money{}, err",
      "    }",
      "    return Money{",
      "        minor:    addExact(m.minor, o.minor),",
      "        currency: m.currency,",
      "    }, nil",
      "}",
    ],
  },
  /** Episode 4: a pointer receiver — the only way to change the original. */
  pointerReceiver: {
    path: "internal/domain/shared/event.go",
    line: 78,
    lang: "go",
    code: [
      "func (e *Events) Record(ev Event) {",
      "    e.pending = append(e.pending, ev)",
      "}",
      "",
      "func (e *Events) PullEvents() []Event {",
      "    out := e.pending",
      "    e.pending = nil",
      "    return out",
      "}",
    ],
  },
  /** Episode 5: a port, declared where it is needed. */
  portInterface: {
    path: "internal/app/ports.go",
    line: 46,
    lang: "go",
    code: [
      "type Clock interface {",
      "    Now() time.Time",
      "}",
      "",
      "…",
      "",
      "type UnitOfWork interface {",
      "    InTx(ctx context.Context,",
      "        fn func(ctx context.Context) error) error",
      "}",
    ],
  },
  /** Episode 5: the one line that makes an implicit interface explicit. */
  assertion: {
    path: "internal/adapter/postgres/product_repo.go",
    line: 23,
    lang: "go",
    code: [
      "type ProductRepo struct {",
      "    pool *pgxpool.Pool",
      "}",
      "",
      "var _ catalog.ProductRepository = (*ProductRepo)(nil)",
    ],
  },
  /** Episode 6: the zero value is legal, so the constructor has to refuse it. */
  zeroGuard: {
    path: "internal/domain/shared/money.go",
    line: 71,
    lang: "go",
    code: [
      "func NewMoney(minor int64, c Currency) Money {",
      "    if c.IsZero() {",
      '        panic("shared: NewMoney called with the zero Currency")',
      "    }",
      "    return Money{minor: minor, currency: c}",
      "}",
    ],
  },
  /** The other half of the zero-value guard, in the file it really lives in. */
  isZero: {
    path: "internal/domain/shared/currency.go",
    line: 87,
    lang: "go",
    code: [
      "func (c Currency) IsZero() bool {",
      "    return c == Currency{}",
      "}",
    ],
  },
  /** Episode 7: a typed id, built by embedding rather than by copying methods. */
  typedID: {
    path: "internal/domain/ordering/order.go",
    line: 25,
    lang: "go",
    code: [
      "type OrderID struct {",
      "    shared.ID",
      "}",
      "",
      "func NewOrderID() OrderID {",
      "    return OrderID{shared.NewID()}",
      "}",
    ],
  },
  /** Episode 7: the aggregate root embeds its event recorder. */
  embedEvents: {
    path: "internal/domain/ordering/order.go",
    line: 109,
    lang: "go",
    code: [
      "type CustomerOrder struct {",
      "    shared.Events",
      "",
      "    id       OrderID",
      "    quote    shared.ID",
      "    …",
      "    total    shared.Money",
      "    …",
      "}",
    ],
  },
  /** Episode 8: map iteration is randomised on purpose, so this sorts. */
  mapOrder: {
    path: "internal/adapter/memory/category_repo.go",
    line: 38,
    lang: "go",
    code: [
      "out := make([]catalog.CategoryPolicy, 0, len(r.byCode))",
      "for _, p := range r.byCode {",
      "    out = append(out, p)",
      "}",
      "// Go map iteration is deliberately randomised",
      "sort.Slice(out, func(i, j int) bool {",
      "    return out[i].Code().String() < out[j].Code().String()",
      "})",
    ],
  },
  /** Episode 9: defer, recover, and a panic put back where it was. */
  recoverTx: {
    path: "internal/adapter/postgres/postgres.go",
    line: 91,
    lang: "go",
    code: [
      "defer func() {",
      "    if p := recover(); p != nil {",
      "        _ = tx.Rollback(ctx)",
      "        panic(p)",
      "    }",
      "    if err != nil {",
      "        _ = tx.Rollback(ctx)",
      "        return",
      "    }",
      "    if cerr := tx.Commit(ctx); cerr != nil {",
      '        err = fmt.Errorf("commit: %w", cerr)',
      "    }",
      "}()",
    ],
  },
  /** Episode 2 reads this struct one line at a time. Type follows name. */
  moneyStruct: {
    path: "internal/domain/shared/money.go",
    line: 57,
    lang: "go",
    code: [
      "type Money struct {",
      "    minor    int64",
      "    currency Currency",
      "}",
    ],
  },
  /** A constructor is an ordinary function. The return type sits at the end. */
  newMoney: {
    path: "internal/domain/shared/money.go",
    line: 71,
    lang: "go",
    code: [
      "func NewMoney(minor int64, c Currency) Money {",
      "    if c.IsZero() {",
      '        panic("shared: NewMoney called with the zero Currency")',
      "    }",
      "    return Money{minor: minor, currency: c}",
      "}",
    ],
  },
  /** Methods live outside the struct, attached by the bit before the name. */
  minorMethod: {
    path: "internal/domain/shared/money.go",
    line: 145,
    lang: "go",
    code: [
      "func (m Money) Minor() int64 {",
      "    return m.minor",
      "}",
    ],
  },
  /** main() builds the world ONCE. This is the shape of every Go server. */
  mainOnce: {
    path: "cmd/api/main.go",
    line: 38,
    lang: "go",
    code: [
      "func main() {",
      "    …",
      "    ctx, cancel := context.WithCancel(context.Background())",
      "    defer cancel()",
      "    …",
      "    g = wire.Memory(clock.System{})",
      "    …",
      "    go func() {",
      "        _ = relay.Run(ctx)",
      "    }()",
      "    …",
      "    log.Fatal(srv.ListenAndServe())",
      "}",
    ],
  },


  /** A shared map behind a lock: the cost of a process that stays alive. */
  mutex: {
    path: "internal/adapter/memory/merchant_repo.go",
    line: 24,
    lang: "go",
    code: [
      "type MerchantRepo struct {",
      "    mu           sync.Mutex",
      "    byID         map[catalog.MerchantID]*catalog.Merchant",
      "    …",
      "}",
      "",
      "…",
      "",
      "func (r *MerchantRepo) ByID(ctx context.Context,",
      "    id catalog.MerchantID) (*catalog.Merchant, error) {",
      "    r.mu.Lock()",
      "    defer r.mu.Unlock()",
      "    …",
      "}",
    ],
  },

  /** ctx as the first parameter of everything that can block. */
  ctx: {
    path: "internal/domain/catalog/repository.go",
    line: 35,
    lang: "go",
    code: [
      "type MerchantRepository interface {",
      "    ByID(ctx context.Context, id MerchantID)",
      "        (*Merchant, error)",
      "",
      "    …",
      "",
      "    Save(ctx context.Context, m *Merchant) error",
      "}",
    ],
  },
  // ── Season 1, episodes 11–16 (PLAN-BO-SUNG §5) ────────────────────────────

  /** T11 — the port takes a function; commit and rollback are InTx's business. */
  inTxPort: {
    path: "internal/app/ports.go",
    line: 46,
    lang: "go",
    code: [
      "type UnitOfWork interface {",
      "    InTx(ctx context.Context,",
      "         fn func(ctx context.Context) error) error",
      "}",
    ],
  },
  /** T11 — a closure capturing `id` from the enclosing scope: how a result escapes InTx. */
  closureResult: {
    path: "internal/app/catalog/register_merchant.go",
    line: 31,
    lang: "go",
    code: [
      "func (h *RegisterMerchantHandler) Handle(ctx context.Context,",
      "    d catalog.MerchantDetails) (catalog.MerchantID, error) {",
      "    now := h.deps.Clock.Now()",
      "",
      "    var id catalog.MerchantID",
      "    err := h.deps.UoW.InTx(ctx, func(ctx context.Context) error {",
      "        m, err := catalog.RegisterMerchant(d, now)",
      "        …",
      "        id = m.ID()",
      "        return nil",
      "    })",
      "    return id, err",
      "}",
    ],
  },
  /** T11 — a named return: the deferred function below can overwrite err. */
  inTxNamed: {
    path: "internal/adapter/postgres/postgres.go",
    line: 83,
    lang: "go",
    code: [
      "func (u *UnitOfWork) InTx(ctx context.Context,",
      "    fn func(ctx context.Context) error) (err error) {",
      "    …",
      "    return fn(context.WithValue(ctx, txKey{}, tx))",
      "}",
    ],
  },
  /** T11 — a closure over `currency` and `at`, written in T-001. */
  moneyHelper: {
    path: "internal/adapter/config/ratecard.go",
    line: 200,
    lang: "go",
    code: [
      "money := func(name, s string) (shared.Money, error) {",
      "    m, err := shared.ParseMoney(s, currency)",
      "    if err != nil {",
      "        return shared.Money{}, field(at(name), err)",
      "    }",
      "    return m, nil",
      "}",
      'standard, err := money("rates.standard", l.Rates.Standard)',
    ],
  },

  /** T12 — two methods make an Event; every event in five contexts satisfies it. */
  eventInterface: {
    path: "internal/domain/shared/event.go",
    line: 26,
    lang: "go",
    code: [
      "type Event interface {",
      "    EventName() string",
      "    OccurredAt() time.Time",
      "}",
    ],
  },
  /** T12 — the payload is a map of any: keys written by hand, never a domain type. */
  payloadMap: {
    path: "internal/adapter/eventcodec/codec.go",
    line: 45,
    lang: "go",
    code: ["type m = map[string]any"],
  },
  /** T12 — one type switch, 37 cases: the only place an event becomes bytes. */
  typeSwitch: {
    path: "internal/adapter/eventcodec/codec.go",
    line: 70,
    lang: "go",
    code: [
      "func payload(ev shared.Event) (m, error) {",
      "    switch e := ev.(type) {",
      "    case catalog.MerchantRegistered:",
      '        return m{"id": e.ID.String(), "name": e.Name,',
      '            "site": e.Site.String(), "currency": e.Currency.Code(),',
      '            "at": ts(e.At)}, nil',
      "    case catalog.MerchantRenamed:",
      '        return m{"id": e.ID.String(), "from": e.From, "to": e.To, "at": ts(e.At)}, nil',
      "    …",
      "    }",
      "}",
    ],
  },
  /** T12 — a type assertion with ok: wrong type is a false, not a crash. */
  assertOk: {
    path: "internal/platform/wire/subscribe.go",
    line: 111,
    lang: "go",
    code: [
      "m, ok := msg.(T)",
      "if !ok {",
      '    return fmt.Errorf("%s decodes to %T, handler wants %T",',
      "        e.Name, msg, *new(T))",
      "}",
      "return handle(ctx, m)",
    ],
  },
  /** T12/T13 — Decode hands back `any`; the caller must assert. */
  decodeAny: {
    path: "internal/adapter/eventcodec/decode.go",
    line: 55,
    lang: "go",
    code: [
      "func Decode(name string, payload []byte) (any, error) {",
      "    dec, ok := decoders[name]",
      "    if !ok {",
      '        return nil, fmt.Errorf("decode %s: %w", name, ErrNoDecoder)',
      "    }",
      "    …",
      "}",
    ],
  },

  /** T13 — the first generic function: a typed handler becomes an untyped listener. */
  onGeneric: {
    path: "internal/platform/wire/subscribe.go",
    line: 105,
    lang: "go",
    code: [
      "func on[T any](handle func(context.Context, T) error) worker.Handler {",
      "    return func(ctx context.Context, e worker.Entry) error {",
      "        msg, err := eventcodec.Decode(e.Name, e.Payload)",
      "        if err != nil {",
      "            return err",
      "        }",
      "        m, ok := msg.(T)",
      "        if !ok {",
      '            return fmt.Errorf("%s decodes to %T, handler wants %T", e.Name, msg, *new(T))',
      "        }",
      "        return handle(ctx, m)",
      "    }",
      "}",
    ],
  },
  /** T13 — the second: JSON into whichever V1 struct T is. */
  intoGeneric: {
    path: "internal/adapter/eventcodec/decode.go",
    line: 66,
    lang: "go",
    code: [
      "func into[T any](payload []byte) (any, error) {",
      "    var v T",
      "    if err := json.Unmarshal(payload, &v); err != nil {",
      "        return nil, err",
      "    }",
      "    return v, nil",
      "}",
    ],
  },
  /** T13 — the type argument written out, because nothing lets Go infer it here. */
  decoders: {
    path: "internal/adapter/eventcodec/decode.go",
    line: 19,
    lang: "go",
    code: [
      "var decoders = map[string]func([]byte) (any, error){",
      '    "catalog.product_published": into[contracts.ProductPublishedV1],',
      '    "catalog.product_measured":  into[contracts.ProductMeasuredV1],',
      "    …",
      "}",
    ],
  },

  /** T14 — the four channel reads of episode 10, in their select. */
  selectLoop: {
    path: "internal/worker/worker.go",
    line: 127,
    lang: "go",
    code: [
      "func (r *Relay) Run(ctx context.Context) error {",
      "    ticker := time.NewTicker(r.deps.Interval)",
      "    defer ticker.Stop()",
      "    for {",
      "        if _, err := r.RunOnce(ctx); err != nil && ctx.Err() == nil {",
      '            r.deps.Log.Printf("worker: %v", err)',
      "        }",
      "        select {",
      "        case <-ctx.Done():",
      "            return ctx.Err()",
      "        case <-ticker.C:",
      "        }",
      "    }",
      "}",
    ],
  },
  /** T14 — WaitGroup + a closed channel as a starting gun: four goroutines collide on purpose. */
  raceStart: {
    path: "internal/adapter/postgres/postgres_test.go",
    line: 103,
    lang: "go",
    code: [
      "const starters = 4",
      "var (",
      "    wg    sync.WaitGroup",
      "    start = make(chan struct{})",
      "    errs  = make([]error, starters)",
      ")",
      "for i := range starters {",
      "    wg.Add(1)",
      "    go func(i int) {",
      "        defer wg.Done()",
      "        …",
      "        <-start",
      "        errs[i] = postgres.Migrate(ctx, p)",
      "    }(i)",
      "}",
      "close(start)",
      "wg.Wait()",
    ],
  },
  /** T14 — a buffered channel of one, so the goroutine can finish even if nobody reads. */
  bufferedDone: {
    path: "internal/adapter/openai/openai_test.go",
    line: 120,
    lang: "go",
    code: [
      "done := make(chan error, 1)",
      "go func() {",
      "    _, err := c.Extract(context.Background(), url(t))",
      "    done <- err",
      "}()",
      "select {",
      "case err := <-done:",
      "    …",
      "case <-time.After(3 * time.Second):",
      '    t.Fatal("Extract did not give up — a request with no timeout can hang forever")',
      "}",
    ],
  },

  /** T15 — the contract: tags, primitives, a version in the name. */
  contractV1: {
    path: "internal/contracts/ordering_v1.go",
    line: 11,
    lang: "go",
    code: [
      "type OrderPlacedV1 struct {",
      '    ID       string    `json:"id"`',
      '    Quote    string    `json:"quote"`',
      '    Product  string    `json:"product"`',
      '    Variant  string    `json:"variant"`',
      '    Customer string    `json:"customer"`',
      '    PlacedBy string    `json:"placed_by"`',
      '    Total    MoneyV1   `json:"total"`',
      '    Deposit  MoneyV1   `json:"deposit"`',
      '    At       time.Time `json:"at"`',
      "}",
    ],
  },
  /** T15 — the same fact in the domain: typed, and not one tag. */
  domainEvent: {
    path: "internal/domain/ordering/events.go",
    line: 12,
    lang: "go",
    code: [
      "type OrderPlaced struct {",
      "    ID       OrderID",
      "    Quote    shared.ID",
      "    Product  shared.ID",
      "    Variant  shared.ID",
      "    Customer shared.ID",
      "    PlacedBy shared.OperatorID",
      "    Total    shared.Money",
      "    Deposit  shared.Money",
      "    At       time.Time",
      "}",
    ],
  },
  /** T15 — the other exit: a screen's money is text, never a float. */
  moneyView: {
    path: "internal/adapter/http/quotes.go",
    line: 64,
    lang: "go",
    code: [
      "type moneyView struct {",
      '    Amount   string `json:"amount"`',
      '    Currency string `json:"currency"`',
      "}",
    ],
  },
  /** T15 — a consumer reads the contract and parses it back into domain types. */
  projectorReads: {
    path: "internal/app/reporting/projector.go",
    line: 81,
    lang: "go",
    code: [
      "func (p *Projector) OnOrderPlaced(ctx context.Context,",
      "    m contracts.OrderPlacedV1) error {",
      "    id, err := ordering.ParseOrderID(m.ID)",
      "    …",
      "    total, err := moneyFrom(m.Total)",
      "    …",
    ],
  },

  /** T16 — the fake world: in-memory adapters behind the same ports Postgres implements. */
  worldFake: {
    path: "internal/app/pricing/pricing_test.go",
    line: 19,
    lang: "go",
    code: [
      "type world struct {",
      "    clock    *clock.Fixed",
      "    lanes    *memory.LaneRepo",
      "    quotes   *memory.QuoteRepo",
      "    listings *memory.ListingRepo",
      "    profiles *memory.ProfileRepo",
      "    rates    *memory.ExchangeRates",
      "    outbox   *memory.Outbox",
      "    deps     pricingapp.Deps",
      "}",
    ],
  },
  /** T16 — httptest: a request through the real handler, no port opened. */
  callHelper: {
    path: "internal/adapter/http/server_test.go",
    line: 166,
    lang: "go",
    code: [
      "func (a *api) call(t *testing.T, method, path, body string,",
      "    headers map[string]string) *httptest.ResponseRecorder {",
      "    t.Helper()",
      "    …",
      "    return a.callAnon(t, method, path, body, headers)",
      "}",
      "…",
      "req := httptest.NewRequest(method, path, strings.NewReader(body))",
    ],
  },
  /** T16 — a table test with keyed rows and t.Run: one row, one name, one failure. */
  tableTest: {
    path: "internal/adapter/config/ratecard_test.go",
    line: 96,
    lang: "go",
    code: [
      "cases := []struct {",
      "    name string",
      "    from string",
      "    to   string",
      "    file string",
      "    want string",
      "}{",
      '    {name: "deposit over 100 %", from: `deposit: "50"`, to: `deposit: "150"`, want: "quote"},',
      "    …",
      "}",
      "for _, tc := range cases {",
      "    t.Run(tc.name, func(t *testing.T) {",
    ],
  },
  /** T16 — a decision guard: the rule lives in a test, so breaking it breaks the build. */
  guardAllowlist: {
    path: "internal/domain/decisions_test.go",
    line: 225,
    lang: "go",
    code: [
      "func TestDecision_domainImportsOnlyStdlibAndAllowlist(t *testing.T) {",
      "    allowed := map[string]bool{",
      '        "github.com/google/uuid": true,',
      "    }",
      "    for _, f := range domainSources(t) {",
      "        for _, imp := range f.ast.Imports {",
      "            …",
      '            t.Errorf("%s imports %q.\\n"+',
      "                …",
      "        }",
      "    }",
      "}",
    ],
  },
} as const satisfies Record<string, Snippet>;

/** Counts that make the Go lessons concrete rather than abstract. */
/**
 * Season 2 (DDD) works from the same repository, so its snippets live here too.
 * The three Variant types are the heart of it: one word, three shapes, three
 * packages — and nobody had to agree on a single definition.
 */
export const DDD_SNIPPETS = {
  variantCatalog: {
    path: "internal/domain/catalog/variant.go",
    line: 49,
    lang: "go",
    code: [
      "type Variant struct {",
      "    id          VariantID",
      "    size        string",
      "    color       string",
      "    merchantRef string",
      "    addedAt     time.Time",
      "}",
    ],
  },
  variantOrdering: {
    path: "internal/domain/ordering/variant.go",
    line: 18,
    lang: "go",
    code: [
      "type Variant struct {",
      "    Variant shared.ID",
      "    Product shared.ID",
      "}",
    ],
  },
  variantProcurement: {
    path: "internal/domain/procurement/repository.go",
    line: 33,
    lang: "go",
    code: [
      "type Variant struct {",
      "    Variant     shared.ID",
      "    Product     shared.ID",
      "    Size        string",
      "    Color       string",
      "    MerchantRef string",
      "}",
    ],
  },
  /** The aggregate's only door: a command, a rule, an event — never a setter. */
  oneDoor: {
    path: "internal/domain/ordering/order.go",
    line: 225,
    lang: "go",
    code: [
      "func (o *CustomerOrder) PayDeposit(",
      "    amount shared.Money, now time.Time) error {",
      "    if o.status != StatusAwaitingDeposit {",
      '        return fmt.Errorf("pay deposit on %s: %w",',
      "            o.id, o.refuse(ErrNotAwaitingDeposit))",
      "    }",
      "    …",
      "    o.status = StatusDeposited",
      "    o.Record(DepositPaid{ID: o.id, …",
      "    return nil",
      "}",
    ],
  },
  /** A repository interface declared by the domain, implemented elsewhere. */
  repoPort: {
    path: "internal/domain/catalog/repository.go",
    line: 34,
    lang: "go",
    code: [
      "type MerchantRepository interface {",
      "    ByID(ctx context.Context, id MerchantID) (*Merchant, error)",
      "    BySite(ctx context.Context, site Hostname) (*Merchant, error)",
      "    …",
      "    Save(ctx context.Context, m *Merchant) error",
      "}",
    ],
  },
} as const satisfies Record<string, Snippet>;

/**
 * PHP written for contrast, not copied from anywhere. Marked `lang: "php"` so
 * the card can label it, because a viewer must never mistake an illustration
 * for something that exists in this repository.
 */
export const PHP_SNIPPETS = {
  /** The anemic model season 2 opens with — recognisable, and nobody's fault. */
  anemic: {
    path: "Symfony — minh hoạ, không phải code repo",
    lang: "php",
    code: [
      "class Order {",
      "    private int $status;",
      "    public function getStatus(): int { … }",
      "    public function setStatus(int $s): void { … }",
      "    public function setDeposit(Money $m): void { … }",
      "}",
      "",
      "// luật nghiệp vụ nằm ở đâu?",
      "$order->setStatus(Order::DEPOSIT_PAID);",
    ],
  },
  tryCatch: {
    path: "Symfony — minh hoạ, không phải code repo",
    lang: "php",
    code: [
      "try {",
      "    $money = Money::parse($raw, $currency);",
      "} catch (MalformedAmount $e) {",
      "    // dễ quên khối này; code vẫn chạy",
      "    $this->logger->error($e->getMessage());",
      "}",
    ],
  },
} as const satisfies Record<string, Snippet>;

export const GO_USAGE = {
  /** grep -c 'context.Context' over internal/ and cmd/, tests excluded */
  ctx: 361,
  /** sync.Mutex / RWMutex */
  mutex: 25,
  /**
   * defer statements. Đếm lại 17/09 bằng grep -rnE '^\s*defer ' (không tính
   * test): 95. Con số 96 cũ đếm nhầm một chữ `defer` nằm trong comment.
   * Đếm lại 06/10: 96 — T-001 thêm `defer f.Close()` trong adapter/config.
   */
  defers: 96,
  /** fmt.Errorf with %w. 456 ngày 17/09; 462 ngày 06/10 (T-001, adapter/config). */
  wrap: 462,
  /** errors.Is */
  errorsIs: 44,
  /**
   * var _ Port = (*Adapter)(nil) compile-time assertions.
   * Đếm lại 17/09: grep -rn 'var _ .* = (\*' --include='*.go' internal/ cmd/
   * → 48, kể cả file test. Con số 53 trong bản kế hoạch cũ là sai.
   */
  assertions: 48,
  /** value receivers `func (x T)` across internal/domain */
  valueReceivers: 176,
  /** pointer receivers `func (x *T)` across internal/domain */
  pointerReceivers: 120,
  /** goroutines started with `go ...` */
  goroutines: 3,
  /**
   * recover() call sites in code that ships — exactly one, in UnitOfWork.InTx.
   * Đếm: grep -rn 'recover()' ... | grep -v _test | grep -v '//'
   * (grep thô ra 7 vì tính cả 5 chỗ trong test và 1 dòng comment.)
   */
  recovers: 1,
  /**
   * Channels the shipping code CREATES: none — `chan` appears only in tests.
   * Đếm: lines matching \bchan\b in internal/ and cmd/, tests and comment-only
   * lines excluded — the `code` helper in scripts/verify-numbers.sh.
   */
  channelsDeclared: 0,
  /**
   * Channel RECEIVES in shipping code: four, all inside a `select` —
   * <-ctx.Done() and <-ticker.C in worker/worker.go and worker/sweep.go.
   * The episode used to say "0 channel". True of declarations, false of what a
   * reader meets on day one in worker.go — so both numbers are shown now.
   * Đếm: lines containing <- , same helper, same exclusions.
   */
  channelsRead: 4,

  // ── Episodes 11–16. Mỗi số kèm lệnh đếm trong scripts/verify-numbers.sh. ──
  /** T11 — `.InTx(` call sites in shipping code. */
  inTxCalls: 38,
  /** T11 — function literals: a `func(` that opens a body on the same line and is not a declaration. */
  funcLiterals: 114,
  /** T11 — functions whose last result is a named `err error`. */
  namedReturns: 4,
  /** T12 — `case` arms of the one type switch in eventcodec.payload. */
  typeSwitchCases: 37,
  /** T12 — `any` / `interface{}` in shipping code. */
  anyUses: 54,
  /** T12 — type assertions written with the ok form. */
  typeAssertOk: 2,
  /** T13 — generic functions. Two. Said out loud. */
  generics: 2,
  /** T14 — `select {` blocks in shipping code. */
  selects: 2,
  /** T14 — sync.WaitGroup, tests included (shipping code has none). */
  waitGroupsInTests: 1,
  /** T14 — `make(chan` in tests (shipping code declares no channel — channelsDeclared). */
  chansInTests: 5,
  /** T15 — `json:"` struct tags in shipping code; zero of them in internal/domain. */
  jsonTags: 364,
  /** T15 — shipping files that import encoding/json. */
  jsonFiles: 6,
  /** T15 — `…V1 struct` types in internal/contracts. */
  contractTypes: 26,
  /** T16 — _test.go files. */
  testFiles: 67,
  /** T16 — `func Test…` functions. */
  testFuncs: 299,
  /** T16 — t.Run calls (subtests). */
  subtests: 5,
  /** T16 — table-driven loops: `for _, tc|c|tt := range`. */
  tableTests: 16,
  /** T16 — `func Benchmark…`: none. Said out loud. */
  benchmarks: 0,
  /** T16 — test files using httptest. */
  httptestFiles: 3,
} as const;
