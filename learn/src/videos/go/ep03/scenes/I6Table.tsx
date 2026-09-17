import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { CompareTable, Punchline } from "../../../../components/CompareTable";
import { Cues } from "../../../../components/Sfx";
import { SAFE } from "../../../../design/tokens";

/** Scene 6 — the delta, in one table. */
export const I6Table: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 3 · CẢNH 6" source="cùng một ý, hai cách viết">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 330, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Cùng một ý, hai cách viết
      </Headline>

      <div style={{ marginTop: 30 }}>
        <CompareTable
          width={SAFE.contentWidth}
          rows={[
            { about: "Báo hỏng", php: "throw — nhảy khỏi hàm", go: "return ..., err — hàm vẫn kết thúc bình thường", at: 40 },
            { about: "Bắt lỗi", php: "catch (X $e)", go: "if err != nil { ... }", at: 90 },
            { about: "Phân biệt loại lỗi", php: "class exception riêng", go: "biến ErrXxx + errors.Is", at: 140 },
            { about: "Giữ nguyên nhân gốc", php: "previous: $e", go: "%w trong fmt.Errorf", at: 190 },
            { about: "Quên xử lý", php: "chương trình vẫn chạy", go: "không build được", at: 240 },
            { about: "Cố tình bỏ qua", php: "catch rỗng, khó thấy", go: "dấu _ , nhìn là thấy", at: 290 },
          ]}
        />
      </div>

      <Punchline at={330}>
        Không có gì Go làm được mà PHP không làm được. Khác ở chỗ Go không cho quên.
      </Punchline>
    </div>
  </Stage>
);
