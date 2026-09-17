import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { CompareTable, Punchline } from "../../../../components/CompareTable";
import { Cues } from "../../../../components/Sfx";
import { SAFE } from "../../../../design/tokens";

/** Scene 5 — the delta, in one table. */
export const J5Table: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 4 · CẢNH 5" source="cùng một ý, hai cách viết">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 300, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Cùng một ý, hai cách viết
      </Headline>

      <div style={{ marginTop: 30 }}>
        <CompareTable
          width={SAFE.contentWidth}
          rows={[
            { about: "Tên của “chính nó”", php: "$this — cố định", go: "đặt tên tuỳ ý, quy ước một chữ cái", at: 40 },
            { about: "Method sửa được bản gốc?", php: "luôn luôn", go: "chỉ khi receiver có dấu *", at: 90 },
            { about: "Value object bất biến", php: "phải tự kỷ luật, tự clone", go: "dùng receiver giá trị là xong", at: 140 },
            { about: "Gọi method trên nil", php: "lỗi ngay", go: "chạy được, nếu method không đụng field", at: 190 },
            { about: "Chỗ viết method", php: "trong ngoặc của class", go: "ngoài struct, cạnh nhau hoặc khác file", at: 240 },
          ]}
        />
      </div>

      <Punchline at={300}>
        PHP cho bạn sửa mọi thứ rồi nhờ bạn đừng sửa. Go bắt bạn nói trước là mình định sửa.
      </Punchline>
    </div>
  </Stage>
);
