import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { FlowBoard, type FlowNode, type FlowEdge } from "../../../../components/Diagram";
import { Cues } from "../../../../components/Sfx";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 3 — capital letter instead of a `public` keyword.
 *
 * The rule takes ten seconds. The part worth a scene is the part that bites: the
 * unit of privacy in Go is the PACKAGE, not the type. `minor` is lowercase, so
 * every file in package shared can touch it — currency.go included. A Symfony
 * developer reads `private` as "only this class" and will be wrong.
 */
const NODES: FlowNode[] = [
  {
    id: "pkg",
    label: "package shared",
    kind: "boundary",
    x: 0,
    y: 16,
    w: 452,
    h: 246,
    color: COLOR.go,
    at: 30,
  },
  { id: "money", label: "money.go", sub: "minor, currency", x: 26, y: 66, w: 190, h: 76, at: 50 },
  { id: "curr", label: "currency.go", sub: "cùng package", x: 240, y: 66, w: 188, h: 76, at: 70 },
  {
    id: "ord",
    label: "package ordering",
    sub: "khác package",
    x: 0,
    y: 322,
    w: 452,
    h: 80,
    color: COLOR.textFaint,
    at: 150,
  },
];

const EDGES: FlowEdge[] = [
  { from: "curr", to: "money", at: 110, color: COLOR.ok },
  { from: "ord", to: "money", at: 190, color: COLOR.danger, dashed: true },
];

export const H3Case: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 2 · CẢNH 3" source="internal/domain/shared/ — một package, nhiều file">
    <Cues
      items={[
        { at: 20, sound: "appear" },
        { at: 110, sound: "travel", volume: 0.16 },
        { at: 190, sound: "crack", volume: 0.22 },
      ]}
    />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Luật hai: chữ HOA là public, chữ thường là private
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={452}
          gap={64}
          left={<FlowBoard nodes={NODES} edges={EDGES} width={452} height={412} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 20 }}>
              <Bullets
                gap={20}
                items={[
                  {
                    at: 60,
                    tag: "luật",
                    text: "Tên bắt đầu bằng chữ HOA thì package khác gọi được. Chữ thường thì không.",
                    accent: COLOR.go,
                  },
                  {
                    at: 130,
                    tag: "ví dụ",
                    text: "Minor() gọi được từ mọi nơi. minor thì không — cùng một chữ, khác mỗi chữ M.",
                    accent: COLOR.go,
                  },
                  {
                    at: 220,
                    tag: "bẫy lớn nhất",
                    text: "Riêng tư ở đây tính theo package, không theo struct. currency.go đọc thẳng m.minor được, vì nó cùng package shared.",
                    accent: COLOR.danger,
                  },
                  {
                    at: 300,
                    tag: "bên symfony",
                    text: "private nghĩa là chỉ class đó. Go không có mức ấy — muốn giấu khỏi file bên cạnh thì phải tách package.",
                    accent: COLOR.php,
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
