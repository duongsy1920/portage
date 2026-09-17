import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 4 — why the ports are tiny.
 *
 * Clock has one method. That is not minimalism for its own sake: an interface
 * is a list of demands, and the package that declares it pays for every demand
 * in fake implementations. Small interfaces are what make the fast tests
 * possible, so the two facts are shown together.
 */
export const K4Small: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 5 · CẢNH 4" source="internal/app/ports.go — cả file có vài interface, mỗi cái vài dòng">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 120, sound: "reveal" }, { at: 280, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Interface càng nhỏ càng dễ sống
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={760}
          gap={56}
          left={<CodeCard snippet={GO_SNIPPETS.portInterface} delay={20} accent={COLOR.go} highlight={[5, 6, 7]} />}
          right={
            <Bullets
              gap={22}
              items={[
                {
                  at: 120,
                  tag: "một method",
                  text: "Clock chỉ có Now(). Muốn test “ba ngày sau khi cọc” thì viết một struct trả về ngày cố định là xong.",
                  accent: COLOR.go,
                },
                {
                  at: 190,
                  tag: "vì sao nhỏ",
                  text: "Interface là danh sách yêu cầu. Thêm một method là bắt mọi bản giả phải viết thêm một method.",
                  accent: COLOR.money,
                },
                {
                  at: 260,
                  tag: "hệ quả đo được",
                  text: "278 test chạy dưới 2 giây, không cần Docker, vì mọi thứ bẩn đều nằm sau một interface nhỏ.",
                  accent: COLOR.ok,
                },
                {
                  at: 330,
                  tag: "bên symfony",
                  text: "Cũng khuyên interface nhỏ, nhưng container tự cắm nên interface to vẫn dùng được. Ở đây main() cắm tay, interface to là thấy đau ngay.",
                  accent: COLOR.php,
                },
              ]}
            />
          }
        />
      </div>
    </div>
  </Stage>
);
