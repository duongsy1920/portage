/**
 * The 17 self-test questions of docs/HOC.md ("Tự kiểm tra — trả lời không
 * nhìn code"), as data the hub can turn into review cards. The wording is the
 * document's, verbatim; scripts/verify-hub.py checks that it still is.
 */
export type QuizGroup = "Go" | "DDD" | "Hệ thống";

export type QuizItem = {
  n: number;
  group: QuizGroup;
  q: string;
  /** Episodes that answer it, in SEASONS order; empty when only the docs do. */
  episodes: readonly string[];
};

export const HOC_QUIZ: readonly QuizItem[] = [
  { n: 1, group: "Go", q: "Vì sao tiền là `int64` chứ không `float64`?", episodes: ["Go-Ep02-DocMotDong"] },
  { n: 2, group: "Go", q: "`error` là giá trị trả về. Điều đó đổi cách viết code thế nào so với exception?", episodes: ["Go-Ep03-LoiLaGiaTri"] },
  { n: 3, group: "Go", q: "`interface` ở Go được khai báo ở đâu: nơi CẦN hay nơi CÀI? Vì sao?", episodes: ["Go-Ep05-InterfaceNgam"] },
  { n: 4, group: "Go", q: "Vì sao thư mục `internal/` có ý nghĩa với compiler?", episodes: ["Ep04-PortAdapter"] },
  { n: 5, group: "DDD", q: "Aggregate root là gì, và ranh giới của nó quyết định điều gì về transaction?", episodes: ["Ddd-Ep03-Aggregate", "Ep01-NamVung"] },
  { n: 6, group: "DDD", q: "Value Object khác Entity ở đúng một điểm. Điểm nào?", episodes: ["Ddd-Ep02-ValueObject"] },
  { n: 7, group: "DDD", q: "Bounded Context: hai context cùng có \"Product\". Chúng có phải cùng một thứ?", episodes: ["Ddd-Ep05-BoundedContext", "Ep01-NamVung"] },
  { n: 8, group: "DDD", q: "Domain Event khác một message queue thường ở chỗ nào?", episodes: ["Ddd-Ep06-DomainEvent", "Ep02-Outbox"] },
  { n: 9, group: "DDD", q: "Anti-Corruption Layer bảo vệ khỏi cái gì? Kể hai cái trong repo này.", episodes: ["Ep08-ACL"] },
  { n: 10, group: "DDD", q: "CQRS: vì sao read model **không có** invariant?", episodes: ["Ep07-ReadModel"] },
  { n: 11, group: "Hệ thống", q: "Outbox giải quyết vấn đề gì mà \"lưu xong rồi publish\" không giải quyết được?", episodes: ["Ep02-Outbox"] },
  { n: 12, group: "Hệ thống", q: "At-least-once nghĩa là subscriber phải có tính chất gì?", episodes: ["Ep03-NhatQuan"] },
  { n: 13, group: "Hệ thống", q: "Vì sao khách hỏi đơn người khác trả **404** chứ không phải 403?", episodes: [] },
  { n: 14, group: "Hệ thống", q: "`variance` âm nghĩa là gì, và bao nhiêu đơn âm liên tiếp thì phải xem lại bảng giá?", episodes: ["Ep06-BaoGia"] },
  { n: 15, group: "Hệ thống", q: "Tính năng AI tắt (không có API key) thì API trả gì, và vì sao **không** được rơi về Fake?", episodes: ["Ep08-ACL"] },
  { n: 16, group: "Hệ thống", q: "Hai context nghe **cùng một event**: vì sao một bên chép 5 field mà bên kia chỉ chép 2? Chép thừa thì hỏng chỗ nào?", episodes: ["Ep03-NhatQuan"] },
  { n: 17, group: "Hệ thống", q: "Hai `CREATE TABLE IF NOT EXISTS` chạy song song trong PostgreSQL thì ra gì, và vì sao lỗi đó **chỉ** hiện trên database trống?", episodes: [] },
];
