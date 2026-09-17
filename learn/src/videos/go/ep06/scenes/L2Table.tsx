import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { CompareTable, Punchline } from "../../../../components/CompareTable";
import { Cues } from "../../../../components/Sfx";
import { SAFE } from "../../../../design/tokens";

/**
 * Scene 2 — the zero of every type the repository actually uses.
 *
 * Learned as a list once, this stops being surprising forever. The last row is
 * the one that bites: a nil slice is not an error, it appends and ranges like
 * an empty one, which is exactly where a PHP habit (`if ($x === null)`) becomes
 * unnecessary noise.
 */
export const L2Table: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 6 · CẢNH 2" source="mỗi kiểu một giá trị mặc định, không có ngoại lệ">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 320, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Zero value của từng kiểu
      </Headline>

      <div style={{ marginTop: 30 }}>
        <CompareTable
          width={SAFE.contentWidth}
          leftTitle="chưa gán thì PHP là"
          rightTitle="chưa gán thì Go là"
          rows={[
            { about: "int64 · int32", php: "null", go: "0", at: 40 },
            { about: "string", php: "null", go: '"" — chuỗi rỗng, không phải null', at: 85 },
            { about: "bool", php: "null", go: "false", at: 130 },
            { about: "struct (Money, Currency…)", php: "null", go: "struct với mọi field bằng zero của nó", at: 175 },
            { about: "con trỏ · interface", php: "null", go: "nil", at: 220 },
            { about: "slice []Event", php: "null", go: "nil — nhưng append và range vẫn chạy", at: 265 },
            { about: "map", php: "null", go: "nil — đọc được, ghi vào thì panic", at: 310 },
          ]}
        />
      </div>

      <Punchline at={360}>
        Hai dòng cuối là chỗ hay sai nhất: slice nil dùng gần như bình thường, map nil
        thì đọc được mà ghi vào là chết.
      </Punchline>
    </div>
  </Stage>
);
