import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 3 — the second trap, and a real bug it caused here.
 *
 * Go randomises map iteration on purpose. In this repository that surfaced as a
 * select box whose default reshuffled on every page load, because the in-memory
 * adapter returned categories in map order while Postgres said ORDER BY code.
 * A real consequence beats a warning about non-determinism.
 */
export const N3MapOrder: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 8 · CẢNH 3" source="internal/adapter/memory/category_repo.go:38 — bug thật">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 130, sound: "crack", volume: 0.2 }, { at: 300, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Bẫy hai: map duyệt theo thứ tự ngẫu nhiên
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={880}
          gap={52}
          left={<CodeCard snippet={GO_SNIPPETS.mapOrder} delay={20} accent={COLOR.money} highlight={[4, 5]} scale={0.9} />}
          right={
            <Bullets
              gap={22}
              items={[
                {
                  at: 130,
                  tag: "cố ý",
                  text: "Go xáo thứ tự mỗi lần chạy, để không ai lỡ dựa vào một thứ tự mà ngôn ngữ không hứa.",
                  accent: COLOR.money,
                },
                {
                  at: 200,
                  tag: "bug thật ở đây",
                  text: "Bản trong bộ nhớ trả danh mục theo thứ tự map, Postgres trả theo ORDER BY code. Hai bản không khớp nhau.",
                  accent: COLOR.danger,
                },
                {
                  at: 270,
                  tag: "khách thấy gì",
                  text: "Ô chọn danh mục đổi mục mặc định sau mỗi lần tải trang, ngay dưới tay người đang dùng.",
                  accent: COLOR.danger,
                },
                {
                  at: 340,
                  tag: "cách chữa",
                  text: "Cần thứ tự thì phải sort. Không có cách nào bảo map giữ thứ tự chèn.",
                  accent: COLOR.ok,
                },
              ]}
            />
          }
        />
      </div>
    </div>
  </Stage>
);
