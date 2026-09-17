import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { CompareTable, Punchline } from "../../../../components/CompareTable";
import { Cues } from "../../../../components/Sfx";
import { SAFE } from "../../../../design/tokens";

/**
 * Scene 4 — the third trap, plus the whole delta in one table.
 *
 * PHP has one `array` for three different things. Naming which Go type replaces
 * which use is more useful than a warning, because the mistake a Symfony
 * developer makes is not misuse — it is not realising a choice was required.
 */
export const N4Compare: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 8 · CẢNH 4" source="PHP có một array; Go có ba thứ khác nhau">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 330, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Một array của PHP, ba thứ của Go
      </Headline>

      <div style={{ marginTop: 30 }}>
        <CompareTable
          width={SAFE.contentWidth}
          rows={[
            { about: "Danh sách có thứ tự", php: "[1, 2, 3]", go: "[]Event — slice, thêm bằng append", at: 40 },
            { about: "Mảng kích thước cố định", php: "không phân biệt", go: "[5]Event — hiếm dùng", at: 90 },
            { about: "Mảng kết hợp", php: "['a' => 1]", go: "map[string]int", at: 140 },
            { about: "Thêm phần tử", php: "$a[] = $x;", go: "a = append(a, x) — phải gán lại", at: 190 },
            { about: "Duyệt theo thứ tự chèn", php: "có, luôn luôn", go: "slice thì có, map thì không", at: 240 },
            { about: "So sánh hai cái", php: "== so từng phần tử", go: "struct so được, slice và map thì không", at: 290 },
          ]}
        />
      </div>

      <Punchline at={330}>
        Bẫy ba nằm ở dòng cuối: so hai struct bằng == là hợp lệ, nhưng struct nào có
        field là slice thì không biên dịch được.
      </Punchline>
    </div>
  </Stage>
);
