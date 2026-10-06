import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { CompareTable, Punchline } from "../../../../components/CompareTable";
import { Cues } from "../../../../components/Sfx";
import { SAFE } from "../../../../design/tokens";

/** Scene 5 — the delta, in one table. */
export const S5Table: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 13 · CẢNH 5" source="cùng một ý, hai cách viết">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 330, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Cùng một ý, hai cách viết
      </Headline>

      <div style={{ marginTop: 30 }}>
        <CompareTable
          width={SAFE.contentWidth}
          rows={[
            { about: "Khai kiểu để trống", php: "/** @template T */ — chú thích", go: "func on[T any](…) — cú pháp thật", at: 40 },
            { about: "Ai kiểm", php: "Psalm / PHPStan, nếu có chạy", go: "compiler, luôn luôn", at: 90 },
            { about: "Điền kiểu lúc gọi", php: "không có; chỉ suy từ docblock", go: "suy từ tham số, hoặc ghi rõ into[T]", at: 140 },
            { about: "Ràng buộc T", php: "@template T of Foo", go: "[T any] · [T comparable] · [T Foo]", at: 190 },
            { about: "Không có generic thì", php: "nhận mixed, trả mixed", go: "nhận any, trả any — rồi x.(T) khắp nơi", at: 240 },
            { about: "Trong repo", php: "—", go: "2 hàm: on[T], into[T]; cả hai [T any]", at: 290 },
          ]}
        />
      </div>

      <Punchline at={330}>
        PHP không có generics. Go có, và một hệ thật dùng chúng đúng hai lần. Cả hai sự thật đều đáng nói khi phỏng vấn.
      </Punchline>
    </div>
  </Stage>
);
