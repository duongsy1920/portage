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
  tests: 312,
  /** the subset that needs no Docker */
  testsNoDocker: 278,
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
   */
  defers: 95,
  /** fmt.Errorf with %w */
  wrap: 456,
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
  /** channels — NONE. Said out loud in the episode, because an interview will ask. */
  channels: 0,
} as const;
