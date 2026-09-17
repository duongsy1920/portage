import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { TermCard } from "../../../../components/TermCard";
import { FlowBoard, type FlowNode } from "../../../../components/Diagram";
import { COLOR, CONTEXT } from "../../../../design/tokens";
import { Cues } from "../../../../components/Sfx";

/**
 * Scene 8 — the third new word, taught by collision rather than definition.
 *
 * Four boxes all called "sản phẩm", each meaning something different and each
 * correct in its own place. Once that is on screen, "bounded context" stops
 * being a folder convention and becomes the obvious answer to a real problem.
 */
const NODES: FlowNode[] = [
  {
    id: "cat",
    label: '"sản phẩm"',
    sub: "tên, shop, các size",
    x: 0,
    y: 0,
    w: 290,
    h: 130,
    color: CONTEXT.catalog.color,
    at: 40,
  },
  {
    id: "pri",
    label: '"sản phẩm"',
    sub: "giá, loại hàng, hộp để tính cước",
    x: 330,
    y: 0,
    w: 290,
    h: 130,
    color: CONTEXT.pricing.color,
    at: 70,
  },
  {
    id: "pro",
    label: '"sản phẩm"',
    sub: "link, size US 9, màu đen",
    x: 0,
    y: 170,
    w: 290,
    h: 130,
    color: CONTEXT.procurement.color,
    at: 100,
  },
  {
    id: "log",
    label: '"sản phẩm"',
    sub: "một vật nặng 1250 g",
    x: 330,
    y: 170,
    w: 290,
    h: 130,
    color: CONTEXT.logistics.color,
    at: 130,
  },
];

export const S8Context: React.FC = () => {
  return (
    <Stage eyebrow="TẬP 1 · CẢNH 8 · TỪ VỰNG">
      <Cues
        items={[
          { at: 40, sound: "appear" },
          { at: 160, sound: "reveal" },
          { at: 320, sound: "appear", volume: 0.14 },
        ]}
      />
      <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
        <Headline delay={4} size={56}>
          Từ thứ ba: bounded context
        </Headline>

        <div style={{ marginTop: 30, flex: 1 }}>
          <Split
            align="flex-start"
            leftWidth={620}
            gap={60}
            left={<FlowBoard nodes={NODES} width={620} height={310} />}
            right={
              <div style={{ display: "flex", flexDirection: "column", gap: 30 }}>
                <TermCard
                  delay={160}
                  accent={COLOR.go}
                  term="bounded context"
                  plain="Một vùng mà trong đó mỗi từ chỉ có một nghĩa."
                  symfony="Như tách thành các Bundle riêng, mỗi cái một schema Doctrine, và cấm quan hệ ManyToOne bắc chéo giữa chúng."
                />
                <Bullets
                  gap={22}
                  items={[
                    {
                      at: 320,
                      tag: "vấn đề",
                      text: "Bốn vùng đều nói 'sản phẩm' và đều đúng, nhưng ý bốn thứ khác nhau.",
                      accent: COLOR.go,
                    },
                    {
                      at: 370,
                      tag: "nếu gộp",
                      text: "Ép chung một struct thì được một bảng bốn mươi cột mà ai cũng sửa và không ai dám xoá.",
                      accent: COLOR.danger,
                    },
                    {
                      at: 420,
                      tag: "mẹo",
                      text: "Hai bên cãi nhau về nghĩa một từ, và cả hai đều đúng: đó chính là chỗ nên tách.",
                      accent: COLOR.ok,
                    },
                  ]}
                />
              </div>
            }
          />
        </div>
      </div>
    </Stage>
  );
};
