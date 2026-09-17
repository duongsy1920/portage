import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { CompareTable, Punchline } from "../../../../components/CompareTable";
import { Cues } from "../../../../components/Sfx";
import { SAFE } from "../../../../design/tokens";

/**
 * Scene 7 — the delta, in one table.
 *
 * Six years of Symfony is not a gap to be filled, it is a map to be redrawn.
 * Every row names something the viewer already does daily, so nothing here is
 * new information — it is the same information, re-indexed.
 */
export const H7Table: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 2 · CẢNH 7" source="cùng một ý, hai cách viết">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 330, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Cùng một ý, hai cách viết
      </Headline>

      <div style={{ marginTop: 30 }}>
        <CompareTable
          width={SAFE.contentWidth}
          rows={[
            { about: "Khai báo một field", php: "private int $minor;", go: "minor int64", at: 40 },
            { about: "Kiểu trả về của hàm", php: "function f(): Money", go: "func f() Money", at: 90 },
            { about: "Cho package khác dùng", php: "từ khoá public", go: "viết hoa chữ đầu", at: 140 },
            { about: "Giấu đi", php: "private — chỉ class này", go: "viết thường — cả package đọc được", at: 190 },
            { about: "Method nằm ở đâu", php: "trong ngoặc của class", go: "ngoài struct, gắn bằng receiver", at: 240 },
            { about: "Tạo một cái mới", php: "new Money(...)", go: "Money{...} — không có new", at: 290 },
          ]}
        />
      </div>

      <Punchline at={330}>
        Cú pháp khác, ý niệm gần như y hệt. Chỗ thật sự khác là dòng thứ tư.
      </Punchline>
    </div>
  </Stage>
);
