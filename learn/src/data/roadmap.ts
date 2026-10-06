/**
 * roadmap.sh's Go roadmap (the PDF of 06/09/2025, text layer read with
 * pdftotext on 06/10/2026), one row per topic node, each with ONE verdict:
 *
 *   taught    an episode teaches it — `ref` is the episode id, `keywords`
 *             are words that must appear on that episode's screen
 *             (scripts/roadmap-gap.py checks; a "taught" with no hit is a lie)
 *   planned   PLAN-BO-SUNG §5 — an episode T11–T16 not yet built
 *   plate     PLAN-BO-SUNG §6 — a reference plate, not a story
 *   exercise  PLAN-BO-SUNG §7 — must be run to be understood
 *   dropped   PLAN-BO-SUNG §8 — on purpose, with the reason
 *
 * Granularity: roadmap.sh draws some leaves as tiny chips ("if", "if-else",
 * "switch"); those are folded into one row when they are taught or shelved
 * together. PLAN-BO-SUNG counted 84 with a script that was lost; this list is
 * the count that can be re-run, and it is allowed to differ.
 */
export type RoadmapStatus = "taught" | "planned" | "plate" | "exercise" | "dropped";

export type RoadmapItem = {
  group: string;
  name: string;
  status: RoadmapStatus;
  /** episode id · T11–T16 · plate name · exercise name — depends on status */
  ref?: string;
  /** for taught: words to grep on that episode's screen; for dropped: unused */
  keywords?: readonly string[];
  note?: string;
};

export const PLANNED = {
  T11: "Closure: hàm nhận hàm làm tham số",
  T12: "interface rỗng và type switch",
  T13: "Generics: hai chỗ duy nhất trong repo",
  T14: "select, ticker, và dừng một việc đang chạy",
  T15: "struct tag, JSON, và đường ra khỏi domain",
  T16: "Vì sao 278 test chạy dưới hai giây",
} as const;

export const PLATES = {
  syntax: "Cú pháp Go còn lại",
  toolchain: "Toolchain và standard library",
} as const;

export const EXERCISES = {
  escape: "Escape analysis trên code thật",
  trace: "Đọc một stack trace thật",
  pprof: "Cắm pprof vào cmd/api rồi đo",
} as const;

const T = (group: string, name: string, ref: string, keywords: readonly string[], note?: string): RoadmapItem =>
  ({ group, name, status: "taught", ref, keywords, note });
const P = (group: string, name: string, ref: keyof typeof PLANNED, note?: string): RoadmapItem =>
  ({ group, name, status: "planned", ref, note });
const S = (group: string, name: string, ref: keyof typeof PLATES, note?: string): RoadmapItem =>
  ({ group, name, status: "plate", ref, note });
const X = (group: string, name: string, ref: keyof typeof EXERCISES, note?: string): RoadmapItem =>
  ({ group, name, status: "exercise", ref, note });
const D = (group: string, name: string, note: string): RoadmapItem => ({ group, name, status: "dropped", note });

export const ROADMAP: readonly RoadmapItem[] = [
  // ── Introduction ──────────────────────────────────────────────────────────
  T("Introduction", "Why use Go", "Go-Ep01-VongDoi", ["PHP-FPM", "goroutine"], "dạy bằng so sánh vòng đời tiến trình, không bằng danh sách ưu điểm"),
  D("Introduction", "History of Go", "anh đang chạy repo này hằng ngày rồi"),
  D("Introduction", "Setting up the Environment", "đã có; docs/SETUP.md"),
  D("Introduction", "Hello World in Go", "repo thật là hello world đủ rồi"),
  S("Introduction", "`go` command · Commands & Docs", "toolchain"),

  // ── Variables & Constants ─────────────────────────────────────────────────
  T("Variables & Constants", "var vs :=", "Go-Ep02-DocMotDong", [":="]),
  T("Variables & Constants", "Zero Values", "Go-Ep06-ZeroValue", ["zero value", "Zero value"]),
  S("Variables & Constants", "const and iota", "syntax", "repo có 1 iota"),
  S("Variables & Constants", "Scope and Shadowing", "syntax"),

  // ── Data Types ────────────────────────────────────────────────────────────
  S("Data Types", "Boolean", "syntax"),
  T("Data Types", "Numeric Types (int, uint, float, complex)", "Go-Ep02-DocMotDong", ["int64"], "qua `minor int64` của Money; float bị cấm có lý do"),
  S("Data Types", "Runes", "syntax", "repo có 5 chỗ dùng rune"),
  S("Data Types", "Strings · Raw and Interpreted Literals", "syntax"),
  S("Data Types", "Type Conversion", "syntax"),

  // ── Composite Types ───────────────────────────────────────────────────────
  S("Composite Types", "Arrays", "syntax"),
  T("Composite Types", "Slices", "Go-Ep08-SliceMap", ["slice", "append"]),
  T("Composite Types", "Capacity and Growth · make()", "Go-Ep08-SliceMap", ["append"], "đi qua bẫy `append` phải gán lại"),
  S("Composite Types", "Slice ↔ Array Conversion", "syntax"),
  T("Composite Types", "Maps", "Go-Ep08-SliceMap", ["map"]),
  S("Composite Types", "Comma-Ok Idiom", "syntax", "repo dùng ở shared/currency.go, có chú thích [PHP]; chưa tập nào dạy hẳn"),
  T("Composite Types", "Structs", "Go-Ep02-DocMotDong", ["struct"]),
  P("Composite Types", "Struct Tags & JSON", "T15"),
  T("Composite Types", "Embedding Structs", "Go-Ep07-Embedding", ["Embedding", "nhúng"]),

  // ── Control flow ──────────────────────────────────────────────────────────
  S("Conditionals", "if · if-else · switch", "syntax"),
  S("Loops", "for · for range · iterating maps and strings", "syntax"),
  S("Loops", "break · continue · goto", "syntax"),

  // ── Functions ─────────────────────────────────────────────────────────────
  T("Functions", "Functions Basics", "Go-Ep02-DocMotDong", ["func"]),
  S("Functions", "Variadic Functions", "syntax"),
  T("Functions", "Multiple Return Values", "Go-Ep03-LoiLaGiaTri", ["(T, error)", "error)"]),
  P("Functions", "Anonymous Functions", "T11"),
  P("Functions", "Closures", "T11"),
  P("Functions", "Named Return Values", "T11"),
  T("Functions", "Call by Value", "Go-Ep04-Receiver", ["bản sao", "copy"]),

  // ── Pointers ──────────────────────────────────────────────────────────────
  T("Pointers", "Pointers Basics", "Go-Ep04-Receiver", ["*Events", "con trỏ"]),
  T("Pointers", "Pointers with Structs", "Go-Ep04-Receiver", ["*Events", "(e *Events)"]),
  T("Pointers", "With Maps & Slices", "Go-Ep08-SliceMap", ["map", "slice"]),

  // ── Memory ────────────────────────────────────────────────────────────────
  X("Memory Management", "Garbage Collection · overview", "escape"),

  // ── Methods and Interfaces ────────────────────────────────────────────────
  T("Methods and Interfaces", "Methods vs Functions", "Go-Ep04-Receiver", ["receiver", "Receiver"]),
  T("Methods and Interfaces", "Pointer Receivers", "Go-Ep04-Receiver", ["(e *Events)", "*Events"]),
  T("Methods and Interfaces", "Value Receivers", "Go-Ep04-Receiver", ["(m Money)"]),
  T("Methods and Interfaces", "Interfaces Basics", "Go-Ep05-InterfaceNgam", ["interface"]),
  P("Methods and Interfaces", "Empty Interfaces", "T12"),
  S("Methods and Interfaces", "Embedding Interfaces", "syntax", "tập 7 chỉ dạy embedding struct; nhúng interface là một dòng tra cứu"),
  P("Methods and Interfaces", "Type Assertions", "T12"),
  P("Methods and Interfaces", "Type Switch", "T12"),

  // ── Generics ──────────────────────────────────────────────────────────────
  P("Generics", "Why Generics?", "T13"),
  P("Generics", "Generic Functions", "T13"),
  P("Generics", "Generic Types / Interfaces", "T13"),
  P("Generics", "Type Constraints", "T13"),
  P("Generics", "Type Inference", "T13"),

  // ── Error Handling ────────────────────────────────────────────────────────
  T("Error Handling", "Error Handling Basics · `error` interface", "Go-Ep03-LoiLaGiaTri", ["error"]),
  T("Error Handling", "errors.New · Sentinel Errors", "Go-Ep03-LoiLaGiaTri", ["errors.New", "sentinel"]),
  T("Error Handling", "fmt.Errorf · Wrapping/Unwrapping", "Go-Ep03-LoiLaGiaTri", ["%w", "errors.Is"]),
  T("Error Handling", "`panic` and `recover`", "Go-Ep09-DeferPanic", ["panic", "recover"]),
  X("Error Handling", "Stack Traces & Debugging", "trace"),

  // ── Code Organization ─────────────────────────────────────────────────────
  S("Code Organization", "go mod init · tidy · vendor", "toolchain"),
  T("Code Organization", "Packages · Package Import Rules", "Go-Ep02-DocMotDong", ["package"], "chữ HOA/thường thay public/private"),
  S("Code Organization", "Using 3rd Party Packages", "toolchain", "repo có 3 dependency trực tiếp, mỗi cái một lý do"),
  D("Code Organization", "Publishing Modules", "repo là ứng dụng, không phải thư viện"),

  // ── Concurrency ───────────────────────────────────────────────────────────
  T("Concurrency", "Goroutines", "Go-Ep01-VongDoi", ["goroutine"]),
  T("Concurrency", "Channels", "Go-Ep10-DongThoi", ["channel", "chan"]),
  P("Concurrency", "Buffered vs Unbuffered", "T14"),
  P("Concurrency", "Select Statement", "T14", "repo có 4 lần đọc channel, đều trong select — tập 10 đã nói đúng số, T14 dạy cách"),
  T("Concurrency", "Worker Pools", "Go-Ep10-DongThoi", ["worker"], "dạy 'ngoài Portage', dán nhãn rõ"),
  T("Concurrency", "`sync` Package · Mutexes", "Go-Ep10-DongThoi", ["sync.Mutex", "khoá"]),
  P("Concurrency", "WaitGroups", "T14"),
  T("Concurrency", "`context` Package · Deadlines & Cancellations", "Go-Ep01-VongDoi", ["context", "ctx"]),
  P("Concurrency", "Concurrency Patterns (fan-in, fan-out, pipeline)", "T14", "ngoài Portage"),
  T("Concurrency", "Race Detection", "Go-Ep01-VongDoi", ["race", "-race"]),

  // ── Standard Library ──────────────────────────────────────────────────────
  S("Standard Library", "I/O & File Handling · os · bufio", "toolchain"),
  S("Standard Library", "flag · time", "toolchain", "cmd/api dùng flag của stdlib, không CLI framework"),
  P("Standard Library", "encoding/json", "T15"),
  S("Standard Library", "slog · regexp", "toolchain"),
  S("Standard Library", "go:embed", "toolchain", "repo có 1 chỗ dùng"),

  // ── Testing ───────────────────────────────────────────────────────────────
  P("Testing & Benchmarking", "`testing` package basics", "T16"),
  P("Testing & Benchmarking", "Table-driven Tests", "T16", "repo không có khi viết plan; từ 06/10 có 2 (adapter/config, adapter/http — T-001, T-002)"),
  P("Testing & Benchmarking", "Mocks and Stubs", "T16", "repo dùng fake thật (adapter bộ nhớ), không mock sinh máy"),
  P("Testing & Benchmarking", "`httptest` for HTTP Tests", "T16"),
  P("Testing & Benchmarking", "Benchmarks", "T16", "repo không có Benchmark nào — dạy ngoài Portage"),
  P("Testing & Benchmarking", "Coverage", "T16"),

  // ── Ecosystem ─────────────────────────────────────────────────────────────
  D("Ecosystem", "Building CLIs (Cobra, urfave/cli, bubbletea)", "repo không có CLI framework; cmd/api dùng flag"),
  T("Ecosystem", "net/http (standard)", "Go-Ep01-VongDoi", ["ListenAndServe", "http"]),
  D("Ecosystem", "Web Frameworks (gin, echo, fiber, beego)", "repo dùng net/http thuần — một quyết định, không phải thiếu sót"),
  D("Ecosystem", "gRPC & Protocol Buffers", "không có, và một hệ mua hộ một người dùng chưa cần"),
  D("Ecosystem", "ORMs & DB Access (pgx, GORM)", "repo dùng pgx trực tiếp và mapping tay, lý do trong postgres.go; GORM thì không"),
  D("Ecosystem", "Logging (Zerolog, Zap)", "repo dùng log của stdlib"),
  D("Ecosystem", "Realtime (Melody, Centrifugo)", "không có"),

  // ── Toolchain ─────────────────────────────────────────────────────────────
  S("Go Toolchain", "go run · build · install · fmt · test · clean · doc · version", "toolchain"),
  S("Go Toolchain", "go generate · Build Tags", "toolchain"),
  S("Go Toolchain", "go vet · goimports", "toolchain"),
  S("Go Toolchain", "Linters (revive, staticcheck, golangci-lint)", "toolchain"),
  S("Go Toolchain", "govulncheck", "toolchain"),
  X("Go Toolchain", "pprof · trace", "pprof", "repo chưa có pprof — bài tập thêm code"),
  T("Go Toolchain", "Race Detector", "Go-Ep01-VongDoi", ["race"]),
  S("Go Toolchain", "Cross-compilation · Building Executables", "toolchain"),

  // ── Advanced ──────────────────────────────────────────────────────────────
  X("Advanced Topics", "Memory Mgmt. in Depth · Escape Analysis", "escape"),
  D("Advanced Topics", "Reflection", "chỉ chạm ở T15 qua encoding/json; đi sâu là chủ đề của người viết thư viện"),
  D("Advanced Topics", "Unsafe Package", "không có, phỏng vấn backend gần như không hỏi"),
  D("Advanced Topics", "CGO Basics", "không có"),
  D("Advanced Topics", "Compiler & Linker Flags", "không có"),
  D("Advanced Topics", "Plugins & Dynamic Loading", "không có"),
];
