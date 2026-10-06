import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { CompareTable, Punchline } from "../../../../components/CompareTable";
import { Cues } from "../../../../components/Sfx";
import { GO_USAGE } from "../../../../data/portage";
import { SAFE } from "../../../../design/tokens";

/** Scene 5 — the delta, in one table. */
export const Q5Table: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 11 · CẢNH 5" source="cùng một ý, hai cách viết">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 330, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Cùng một ý, hai cách viết
      </Headline>

      <div style={{ marginTop: 30 }}>
        <CompareTable
          width={SAFE.contentWidth}
          rows={[
            { about: "Hàm không tên", php: "function () { … }  /  fn () => …", go: "func() { … } — không có dạng rút gọn", at: 40 },
            { about: "Bắt biến ngoài", php: "phải khai use ($x), mặc định là bản sao", go: "tự động, luôn là chính biến đó", at: 90 },
            { about: "Sửa biến ngoài", php: "use (&$x) — thêm dấu &", go: "gán bình thường: id = m.ID()", at: 140 },
            { about: "Trả về có tên", php: "không có", go: "(err error) — defer ghi đè được", at: 190 },
            { about: "Chạy song song", php: "không có", go: "go func() { … }() — tập 1 và tập 10", at: 240 },
            { about: "Trong repo", php: "$em->wrapInTransaction(fn)", go: `InTx(ctx, fn) — ${GO_USAGE.inTxCalls} chỗ`, at: 290 },
          ]}
        />
      </div>

      <Punchline at={330}>
        PHP cũng có closure. Khác ở chỗ Go bắt biến mặc định, nên một hàm nhận hàm là cách tự nhiên để ép một đoạn code chạy trong khung của hàm khác.
      </Punchline>
    </div>
  </Stage>
);
