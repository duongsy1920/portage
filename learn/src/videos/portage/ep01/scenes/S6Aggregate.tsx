import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { TermCard } from "../../../../components/TermCard";
import { FlowBoard, type FlowNode, type FlowEdge } from "../../../../components/Diagram";
import { COLOR, CONTEXT } from "../../../../design/tokens";
import { Cues } from "../../../../components/Sfx";

/**
 * Scene 6 — the second new word.
 *
 * "Aggregate" is the term most likely to be learned as a ritual ("put things
 * in a folder") rather than as a rule, so it is defined by what it DECIDES:
 * one aggregate is one thing a single transaction is allowed to change.
 *
 * The picture shows the boundary, not a class diagram: one door in, and the
 * rules living behind it.
 */
/**
 * The boundary is listed FIRST so it paints behind everything else: it is the
 * container, and the whole point of the picture is that the fields are inside
 * it and the outside world is not.
 */
const NODES: FlowNode[] = [
  {
    id: "wall",
    label: "RANH GIỚI AGGREGATE",
    x: 240,
    y: 10,
    w: 370,
    h: 370,
    color: CONTEXT.ordering.color,
    kind: "boundary",
    at: 90,
  },
  {
    id: "outside",
    label: "tầng ngoài",
    sub: "HTTP, worker",
    x: 0,
    y: 155,
    w: 200,
    h: 90,
    color: COLOR.textFaint,
    at: 30,
  },
  {
    id: "root",
    label: "CustomerOrder",
    sub: "cửa duy nhất vào trong",
    x: 265,
    y: 150,
    w: 320,
    h: 110,
    color: CONTEXT.ordering.color,
    at: 55,
  },
  { id: "f1", label: "status", x: 270, y: 62, w: 150, h: 62, kind: "ghost", at: 115 },
  { id: "f2", label: "deposit", x: 435, y: 62, w: 150, h: 62, kind: "ghost", at: 125 },
  { id: "f3", label: "refund", x: 350, y: 292, w: 150, h: 62, kind: "ghost", at: 135 },
];

const EDGES: FlowEdge[] = [{ from: "outside", to: "root", at: 75 }];

export const S6Aggregate: React.FC = () => {
  return (
    <Stage eyebrow="TẬP 1 · CẢNH 6 · TỪ VỰNG">
      <Cues
        items={[
          { at: 30, sound: "appear" },
          { at: 150, sound: "reveal" },
          { at: 300, sound: "appear", volume: 0.14 },
        ]}
      />
      <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
        <Headline delay={4} size={56}>
          Từ thứ hai: aggregate
        </Headline>

        <div style={{ marginTop: 30, flex: 1 }}>
          <Split
            align="flex-start"
            leftWidth={620}
            gap={60}
            left={<FlowBoard nodes={NODES} edges={EDGES} width={620} height={400} />}
            right={
              <div style={{ display: "flex", flexDirection: "column", gap: 30 }}>
                <TermCard
                  delay={140}
                  accent={CONTEXT.ordering.color}
                  term="aggregate"
                  plain="Một cụm dữ liệu phải luôn đúng cùng nhau, chỉ sửa được qua một cửa."
                  symfony="Gần nhất là một Entity gốc mà bạn không cho ai gọi setter — mọi thay đổi phải đi qua method có tên nghiệp vụ."
                />
                <Bullets
                  gap={22}
                  items={[
                    {
                      at: 300,
                      tag: "luật",
                      text: "Một transaction chỉ được sửa MỘT aggregate. Nhiều hơn là dấu hiệu chia sai.",
                      accent: CONTEXT.ordering.color,
                    },
                    {
                      at: 350,
                      tag: "nên",
                      text: "Câu hỏi để chia: cái gì phải đúng ngay lập tức? Cái đó ở chung. Cái gì đúng sau vài giây cũng được? Tách ra.",
                      accent: CONTEXT.ordering.color,
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
