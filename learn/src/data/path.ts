/**
 * docs/HOC.md's ten sessions, as data: which episodes belong to which session,
 * and what to read and do besides watching. Until now this mapping existed
 * only in prose — HOC.md never names an episode — so the hub is the first
 * place a learner sees "buổi 1 = Mùa 1, tập 1–10" written down.
 *
 * The mapping is a judgement, stated here once: an episode goes with the
 * session whose reading it anchors. Season 1 is all of session 1 because
 * session 1 is "read Go"; the DDD and Portage episodes are spread over
 * sessions 2–7 following the concepts each session introduces. Sessions 0, 8
 * and 9 are things to DO, not to watch.
 */
export type Session = {
  n: number;
  title: string;
  /** One line: what this session is for. */
  goal: string;
  /** Episode ids in SEASONS order. */
  episodes: readonly string[];
  /** What to read, as the doc says it — file and section. */
  read: readonly string[];
  /** What to do before calling the session done. */
  do: readonly string[];
};

export const PATH: readonly Session[] = [
  {
    n: 0,
    title: "Bấm trước, đọc sau",
    goal: "Có bản đồ trong đầu trước khi đọc dòng code nào.",
    episodes: [],
    read: ["docs/UI-GUIDE.md"],
    do: ["Mở web/flow.html, bấm qua 26 bước của một đơn.", "Chạy go run ./cmd/api -web ./web, đi hết luồng trên hai màn hình thật."],
  },
  {
    n: 1,
    title: "Go không có gì lạ, chỉ khác chỗ để lỗi",
    goal: "Đọc được mọi file .go trong repo.",
    episodes: ["Go-Ep01-VongDoi", "Go-Ep02-DocMotDong", "Go-Ep03-LoiLaGiaTri", "Go-Ep04-Receiver", "Go-Ep05-InterfaceNgam", "Go-Ep06-ZeroValue", "Go-Ep07-Embedding", "Go-Ep08-SliceMap", "Go-Ep09-DeferPanic", "Go-Ep10-DongThoi"],
    read: ["docs/GO-CHO-PHP.md §1–§9, theo thứ tự", "internal/domain/shared/money.go · decimal.go · id.go"],
    do: ["Sửa một dòng trong money.go cho test đỏ, rồi sửa lại cho xanh."],
  },
  {
    n: 2,
    title: "Value Object và Entity: cái gì có ID, cái gì không",
    goal: "Phân biệt được bằng một câu hỏi: hai cái giống hệt có phải một không?",
    episodes: ["Ddd-Ep01-ThieuMau", "Ddd-Ep02-ValueObject"],
    read: ["docs/DDD.md §10–§13"],
    do: ["Đọc shared/weight.go, catalog/freeshipping.go, catalog/merchant.go; nói to cái nào là VO, cái nào là Entity."],
  },
  {
    n: 3,
    title: "Aggregate: ranh giới của một transaction",
    goal: "Chỉ ra được vì sao Product.Publish() có năm điều kiện và chúng nằm ở đó.",
    episodes: ["Ddd-Ep03-Aggregate", "Ddd-Ep06-DomainEvent", "Ep01-NamVung"],
    read: ["docs/DDD.md §14, §15, §29"],
    do: ["Đọc catalog/product.go (Publish) và catalog/events.go; với từng điều kiện hỏi: ở Symfony cái này nằm đâu?"],
  },
  {
    n: 4,
    title: "Port & Adapter: cái làm project này khác một CRUD",
    goal: "Thấy một port có hai adapter, và vì sao test không cần Docker.",
    episodes: ["Ddd-Ep04-Repository", "Ep04-PortAdapter"],
    read: ["docs/DDD.md §16, §20, §21", "docs/WALKTHROUGH.md §4, §6"],
    do: ["So domain/catalog/repository.go với adapter/memory và adapter/postgres: cùng một interface, hai cài đặt."],
  },
  {
    n: 5,
    title: "Một đơn hàng đi hết năm context",
    goal: "Kể lại được đường đi của một đơn qua event, không qua lời gọi hàm.",
    episodes: ["Ddd-Ep05-BoundedContext", "Ep02-Outbox", "Ep03-NhatQuan", "Ep05-KhongQuayDau"],
    read: ["docs/FLOW-ORDER.md cả file", "docs/DDD.md §6, §7, §25, §27"],
    do: ["Chạy smoke thật (scripts/smoke.sh) và đọc log của worker: đếm event mỗi lượt relay."],
  },
  {
    n: 6,
    title: "Tiền: Quote vs Actual",
    goal: "Giải thích được variance −2.50 USD từ số nào trừ số nào.",
    episodes: ["Ep06-BaoGia"],
    read: ["docs/DDD.md §28", "docs/WALKTHROUGH.md §14, §17"],
    do: ["Đọc pricing/calc.go (hàm thuần), quote.go, reconciliation.go, shared/allocate.go."],
  },
  {
    n: 7,
    title: "Ai đang gọi, việc không ai gọi, và màn hình cần bốn context",
    goal: "Biết vì sao auth ở biên, sweep chạy theo đồng hồ, và read model không JOIN.",
    episodes: ["Ep07-ReadModel", "Ep08-ACL"],
    read: ["docs/WALKTHROUGH.md §18–§21"],
    do: ["Tắt OPENAI_API_KEY rồi gọi POST /products/from-url: xem API trả gì."],
  },
  {
    n: 8,
    title: "Đọc hết repo",
    goal: "101 file theo thứ tự WALKTHROUGH §10, chia bốn ngày.",
    episodes: [],
    read: ["docs/WALKTHROUGH.md §10"],
    do: ["Ngày 1: file 1–35 · ngày 2: 36–56 · ngày 3: 57–86 · ngày 4: 87–101."],
  },
  {
    n: 9,
    title: "Ba thứ chỉ học được khi CHẠY THẬT",
    goal: "Ba lần code đang có hoá ra sai, và chỉ lộ ra khi chạy.",
    episodes: [],
    read: ["docs/WALKTHROUGH.md §22", "docs/SETUP.md §9 đợt 15, 18, 20"],
    do: ["Đọc ba đợt review đó như đọc nhật ký của người đi trước."],
  },
];
