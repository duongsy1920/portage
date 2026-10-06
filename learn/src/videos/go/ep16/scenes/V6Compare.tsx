import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { CompareTable, Punchline } from "../../../../components/CompareTable";
import { Cues } from "../../../../components/Sfx";
import { GO_USAGE } from "../../../../data/portage";
import { SAFE } from "../../../../design/tokens";

/**
 * Scene 6 — PHPUnit against the testing package, including the two rows
 * where the repository has nothing: benchmarks and generated mocks. Said in
 * the table, not hidden.
 */
export const V6Compare: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 16 · CẢNH 6" source="cùng một ý, hai cách viết — và hai hàng repo chưa có">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 360, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        PHPUnit và gói testing
      </Headline>

      <div style={{ marginTop: 26 }}>
        <CompareTable
          width={SAFE.contentWidth}
          leftTitle="PHPUnit"
          rightTitle="testing"
          rows={[
            { about: "Khẳng định", php: "$this->assertSame($want, $got)", go: "if got != want { t.Errorf(…) } — không có thư viện assert", at: 40 },
            { about: "Nhiều ca", php: "@dataProvider", go: `bảng + t.Run — ${GO_USAGE.tableTests} bảng trong repo`, at: 85 },
            { about: "Thay phụ thuộc", php: "createMock(), Prophecy", go: "fake trong adapter/memory; không mock sinh máy", at: 130 },
            { about: "HTTP", php: "WebTestCase, dựng kernel", go: "httptest, chỉ dựng handler", at: 175 },
            { about: "Đo tốc độ", php: "—", go: `func BenchmarkX(b *testing.B), b.N, go test -bench — repo có ${GO_USAGE.benchmarks}`, at: 220 },
            { about: "Độ phủ · đua", php: "--coverage (xdebug) · —", go: "go test -cover · go test -race — CI bật cả hai", at: 265 },
          ]}
        />
      </div>

      <Punchline at={360}>
        Hai hàng trống của repo là hai câu phỏng vấn: benchmark viết thế nào, và vì sao không mock. Câu thứ hai trả lời được bằng tập này.
      </Punchline>
    </div>
  </Stage>
);
