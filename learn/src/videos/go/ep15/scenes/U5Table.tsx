import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { CompareTable, Punchline } from "../../../../components/CompareTable";
import { Cues } from "../../../../components/Sfx";
import { SAFE } from "../../../../design/tokens";

/** Scene 5 — the delta, in one table. */
export const U5Table: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 15 · CẢNH 5" source="cùng một ý, hai cách viết">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 330, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Cùng một ý, hai cách viết
      </Headline>

      <div style={{ marginTop: 30 }}>
        <CompareTable
          width={SAFE.contentWidth}
          rows={[
            { about: "Đặt tên key", php: "#[SerializedName('total')]", go: "`json:\"total\"` — chuỗi sau kiểu", at: 40 },
            { about: "Ai đọc metadata", php: "Serializer, lúc chạy", go: "encoding/json, lúc chạy, bằng reflection", at: 90 },
            { about: "Bỏ field", php: "#[Ignore] / groups", go: "`json:\"-\"` · `omitempty`", at: 140 },
            { about: "Entity → JSON", php: "normalize thẳng Entity (hay gặp)", go: "domain không có tag; qua contract hoặc view", at: 190 },
            { about: "Đổi contract", php: "đổi attribute, khó thấy", go: "kiểu V2 mới; codec viết tay không biên dịch", at: 240 },
            { about: "Trong repo", php: "—", go: "364 tag, 26 kiểu V1, 0 tag trong domain", at: 290 },
          ]}
        />
      </div>

      <Punchline at={330}>
        Tag là cơ chế. Ranh giới mới là bài học: thứ đi ra khỏi domain luôn là một struct khác, viết cho người nhận.
      </Punchline>
    </div>
  </Stage>
);
