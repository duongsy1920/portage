import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { CompareTable, Punchline } from "../../../../components/CompareTable";
import { Cues } from "../../../../components/Sfx";
import { SAFE } from "../../../../design/tokens";

/** Scene 4 — when to reach for which, as a table you can answer from. */
export const P4Table: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 10 · CẢNH 4" source="chọn theo hình dạng bài toán, không theo mốt">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 300, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Khi nào khoá, khi nào channel
      </Headline>

      <div style={{ marginTop: 30 }}>
        <CompareTable
          width={SAFE.contentWidth}
          leftTitle="dùng sync.Mutex"
          rightTitle="dùng channel"
          rows={[
            { about: "Hình dạng bài toán", php: "nhiều goroutine cùng chạm một dữ liệu", go: "chuyển việc từ goroutine này sang cái kia", at: 40 },
            { about: "Ví dụ", php: "một map trong bộ nhớ, một bộ đếm", go: "hàng đợi việc, kết quả trả về, tín hiệu dừng", at: 90 },
            { about: "Trong repo này", php: "25 chỗ", go: "0 chỗ", at: 140 },
            { about: "Hỏng thì hỏng kiểu gì", php: "quên mở khoá → treo; quên khoá → dữ liệu hỏng", go: "không ai lấy → treo; đóng hai lần → panic", at: 190 },
            { about: "Công cụ dò", php: "go test -race", go: "go test -race, và pprof xem goroutine treo", at: 240 },
          ]}
        />
      </div>

      <Punchline at={300}>
        Câu trả lời đúng khi phỏng vấn không phải “channel hay hơn”. Là: dữ liệu dùng chung
        thì khoá, chuyển giao việc thì channel.
      </Punchline>
    </div>
  </Stage>
);
