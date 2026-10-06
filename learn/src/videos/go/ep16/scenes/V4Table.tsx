import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 4 — a table test, on the one written with keyed rows and t.Run.
 *
 * The rows are keyed by field name on purpose: the first version of this
 * test used positional rows, added a field, and stopped compiling (T-001).
 * That mistake is part of the lesson.
 */
export const V4Table: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 16 · CẢNH 4" source={`internal/adapter/config/ratecard_test.go:96 — repo có ${GO_USAGE.tableTests} bảng test, ${GO_USAGE.subtests} t.Run`}>
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 340, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Bảng test: một slice ca kiểm, một vòng lặp, mỗi hàng một tên
      </Headline>

      <div style={{ marginTop: 22, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={920}
          gap={44}
          left={<CodeCard snippet={GO_SNIPPETS.tableTest} delay={20} accent={COLOR.money} highlight={[7, 10, 11]} lineStagger={2} />}
          right={
            <Bullets
              gap={20}
              items={[
                { at: 150, tag: "vì sao bảng", text: "Mười một cách làm sai một file cấu hình, mỗi cách một hàng. Thêm trường hợp mới là thêm một hàng, không thêm một hàm.", accent: COLOR.go },
                { at: 210, tag: "t.Run", text: "Mỗi hàng thành một subtest có tên. Đỏ thì báo đúng tên hàng; chạy riêng bằng go test -run 'TestParse/no_lanes'.", accent: COLOR.money },
                { at: 270, tag: "hàng theo tên", text: "Viết name:, from:, want: thay vì theo vị trí. Bản đầu viết theo vị trí, thêm một trường là mười hàng cũ không biên dịch — lỗi thật, đợt 22.", accent: COLOR.danger },
                { at: 340, tag: "bên phpunit", text: "@dataProvider trả mảng các ca. Cùng ý; khác ở chỗ Go để dữ liệu ngay trong hàm test, đọc là thấy.", accent: COLOR.php },
              ]}
            />
          }
        />
      </div>
    </div>
  </Stage>
);
