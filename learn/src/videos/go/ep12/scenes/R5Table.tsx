import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { CompareTable, Punchline } from "../../../../components/CompareTable";
import { Cues } from "../../../../components/Sfx";
import { SAFE } from "../../../../design/tokens";

/** Scene 5 — the delta, in one table. */
export const R5Table: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 12 · CẢNH 5" source="cùng một ý, hai cách viết">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 330, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Cùng một ý, hai cách viết
      </Headline>

      <div style={{ marginTop: 30 }}>
        <CompareTable
          width={SAFE.contentWidth}
          rows={[
            { about: "Nhận mọi kiểu", php: "mixed", go: "any (= interface{})", at: 40 },
            { about: "Hỏi lại kiểu", php: "$x instanceof T", go: "v, ok := x.(T)", at: 90 },
            { about: "Rẽ nhánh theo kiểu", php: "chuỗi if … instanceof", go: "switch v := x.(type) { case T: … }", at: 140 },
            { about: "Sai kiểu thì", php: "TypeError lúc chạy", go: "panic nếu không ok; false nếu có ok", at: 190 },
            { about: "Ép kiểu số", php: "(int) $x — đổi giá trị", go: "int(x) — chỉ khi x là số; không ép từ any", at: 240 },
            { about: "Trong repo", php: "—", go: "1 type switch (codec), 2 assertion, cả hai có ok", at: 290 },
          ]}
        />
      </div>

      <Punchline at={330}>
        any là cửa thoát khỏi kiểu tĩnh. Repo đi qua cửa đó đúng hai chỗ: mã hoá event ra, và giải mã về.
      </Punchline>
    </div>
  </Stage>
);
