import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { CompareTable, Punchline } from "../../../../components/CompareTable";
import { Cues } from "../../../../components/Sfx";
import { SAFE } from "../../../../design/tokens";

/**
 * Scene 2 — the three conversations, from the repository's own error table.
 *
 * internal/adapter/http/errors.go has 92 rows and exactly three codes for
 * things that are not bugs. Naming what each one asks the client to DO is more
 * useful than naming the code, because that is the part a client can act on.
 */
export const P3S2: React.FC = () => (
  <Stage eyebrow="MÙA 3 · TẬP 3 · CẢNH 2" source="internal/adapter/http/errors.go — 92 dòng, ba câu chuyện">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 300, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Ba mã, ba câu chuyện khác nhau
      </Headline>

      <div style={{ marginTop: 30 }}>
        <CompareTable
          width={SAFE.contentWidth}
          leftTitle="mã"
          rightTitle="khách phải làm gì"
          rows={[
            { about: "Không hiểu được request", php: "400 — 36 dòng", go: "sửa lại request rồi gửi lại", at: 40 },
            { about: "Trỏ vào thứ không có", php: "404 — 13 dòng", go: "đợi relay rồi thử lại, hoặc id sai thật", at: 95 },
            { about: "Hiểu được, nhưng nghiệp vụ nói không", php: "409 — 40 dòng", go: "sửa trạng thái, đừng sửa request", at: 150 },
            { about: "Không nằm trong bảng", php: "500", go: "đó là bug của mình — chỉ ghi log", at: 205 },
          ]}
        />
      </div>

      <Punchline at={300}>
        Mã trạng thái ở đây không phải trang trí. Nó là cách duy nhất để khách biết
        nên thử lại hay nên sửa dữ liệu.
      </Punchline>
    </div>
  </Stage>
);
