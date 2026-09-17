import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { FlowBoard, type FlowNode, type FlowEdge } from "../../../../components/Diagram";
import { TermCard } from "../../../../components/TermCard";
import { COLOR } from "../../../../design/tokens";
import { Cues } from "../../../../components/Sfx";

/**
 * Scene 2 — the first new word.
 *
 * Nothing later in the episode works if "transaction" stays fuzzy, and the
 * viewer has written `$em->flush()` a thousand times without ever needing to
 * name what it guarantees. So the guarantee gets named first, with a picture.
 *
 * The RIGHT column is the wide one here: on a vocabulary scene the words are
 * the content, and the picture only has to hold four boxes.
 */
const NODES: FlowNode[] = [
  { id: "begin", label: "BEGIN", x: 170, y: 0, w: 220, h: 66, color: COLOR.money, at: 40 },
  {
    id: "w1",
    label: "ghi: đơn = đã cọc",
    sub: "bảng orders",
    x: 100,
    y: 116,
    w: 360,
    h: 96,
    color: COLOR.lineHi,
    at: 70,
  },
  {
    id: "w2",
    label: "ghi: mở việc mua",
    sub: "bảng purchase_tasks",
    x: 100,
    y: 262,
    w: 360,
    h: 96,
    color: COLOR.lineHi,
    at: 100,
  },
  { id: "commit", label: "COMMIT", x: 170, y: 408, w: 220, h: 66, color: COLOR.ok, at: 130 },
];

const EDGES: FlowEdge[] = [
  { from: "begin", to: "w1", at: 62 },
  { from: "w1", to: "w2", at: 92 },
  { from: "w2", to: "commit", at: 122 },
];

export const S2Transaction: React.FC = () => {
  return (
    <Stage eyebrow="TẬP 1 · CẢNH 2 · TỪ VỰNG">
      <Cues
        items={[
          { at: 40, sound: "appear" },
          { at: 130, sound: "land", volume: 0.22 },
          { at: 150, sound: "reveal" },
        ]}
      />
      <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
        <Headline delay={4} size={56}>
          Trước hết, một từ: transaction
        </Headline>

        <div style={{ marginTop: 30, flex: 1 }}>
          <Split
            align="flex-start"
            leftWidth={560}
            gap={60}
            left={<FlowBoard nodes={NODES} edges={EDGES} width={560} height={480} />}
            right={
              <div style={{ display: "flex", flexDirection: "column", gap: 30 }}>
                <TermCard
                  delay={150}
                  accent={COLOR.money}
                  term="transaction"
                  plain="Nhiều lệnh ghi gộp làm một. Database bảo đảm hoặc chạy hết, hoặc không chạy gì."
                  symfony="$em->wrapInTransaction() — bạn đã dùng hàng nghìn lần, chỉ là chưa cần gọi tên điều nó bảo đảm."
                />
                <Bullets
                  gap={22}
                  items={[
                    {
                      at: 300,
                      tag: "nên",
                      text: "Việc thứ hai hỏng thì việc thứ nhất cũng bị xoá, như chưa từng xảy ra.",
                      accent: COLOR.money,
                    },
                    {
                      at: 345,
                      tag: "điều kiện",
                      text: "Hai lệnh ghi phải đi qua cùng một transaction. Portage truyền nó trong ctx, mỗi repository tự lấy ra.",
                      accent: COLOR.money,
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
