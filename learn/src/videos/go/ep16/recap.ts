import { GO_USAGE, REPO } from "../../../data/portage";
import type { Recap } from "../../registry";

/**
 * The closing screen's data: what to remember, what an interview asks, what comes next.
 * Read by the scene (to draw it) and by the learning hub (to quiz on it) — one source.
 */
export const RECAP = {
  points: [
    { at: 30, text: `${REPO.testsNoDocker} test chạy dưới hai giây không cần Docker, vì mỗi port có một adapter trong RAM — fake thật, không phải mock kịch bản.` },
    { at: 80, text: `httptest gửi request qua đúng handler thật mà không mở cổng. ${GO_USAGE.httptestFiles} file test HTTP, cùng cái cửa auth với curl.` },
    { at: 130, text: `Bảng test là một slice ca kiểm + một vòng lặp; hàng có tên đi vào t.Run. Repo có ${GO_USAGE.tableTests} bảng, ${GO_USAGE.subtests} t.Run.` },
    { at: 180, text: `${REPO.guards} guard bằng go/ast biến quyết định kiến trúc thành test: vi phạm là build đỏ, không phải một dòng docs bị quên.` },
  ],
  asked: [
    "Fake khác mock ở chỗ nào, và vì sao repo không dùng mock sinh máy?",
    "Table-driven test là gì, và t.Run thêm được gì?",
    "Benchmark trong Go viết thế nào, và b.N nghĩa là gì?",
  ],
} as const satisfies Recap;
