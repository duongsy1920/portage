import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { CompareTable, Punchline } from "../../../../components/CompareTable";
import { Cues } from "../../../../components/Sfx";
import { SAFE } from "../../../../design/tokens";

/** Scene 5 — the delta, in one table. */
export const K5Table: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 5 · CẢNH 5" source="cùng một ý, hai cách viết">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 300, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Cùng một ý, hai cách viết
      </Headline>

      <div style={{ marginTop: 30 }}>
        <CompareTable
          width={SAFE.contentWidth}
          rows={[
            { about: "Khai báo là mình theo interface", php: "implements XInterface", go: "không khai báo gì, cứ đủ method là được", at: 40 },
            { about: "Interface khai ở đâu", php: "cạnh class, trong bundle của nó", go: "ở package cần nó, không ở package làm ra nó", at: 90 },
            { about: "Cắm bản cài đặt vào", php: "container tự autowire", go: "main() cắm tay, đọc là thấy", at: 140 },
            { about: "Đổi Postgres sang bộ nhớ", php: "sửa services.yaml", go: "sửa một dòng trong main()", at: 190 },
            { about: "Xoá nhầm một method", php: "lỗi ngay ở class đó", go: "lỗi ở nơi gọi — trừ khi có dòng var _", at: 240 },
          ]}
        />
      </div>

      <Punchline at={300}>
        Đảo chiều phụ thuộc ở đây không phải là một khuôn mẫu phải nhớ. Nó là hệ quả
        của việc không có từ khoá implements.
      </Punchline>
    </div>
  </Stage>
);
