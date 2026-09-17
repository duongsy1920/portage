import React from "react";
import { Interactive, interpolate, useCurrentFrame } from "remotion";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { SketchCard } from "../../../../components/SketchCard";
import { TxBracket } from "../../../../components/Timeline";
import { COLOR, RADIUS, TYPE } from "../../../../design/tokens";
import { MONO } from "../../../../design/fonts";
import { EASE } from "../../../../design/motion";
import { GOLDEN } from "../../../../data/portage";
import { Cues } from "../../../../components/Sfx";

/** The design a Doctrine reflex reaches for first. Deliberately wrong. */
const NAIVE = [
  "CREATE TABLE orders (",
  "  id, customer, product,",
  "  deposit_paid,      -- tiền khách",
  "  shop_order_ref,    -- mã ở shop",
  "  purchase_status,   -- mua chưa",
  "  parcel_weight, batch_id, freight,",
  "  tracking, …",
  ");",
] as const;

/**
 * Scene 5 — the reflex, and where it breaks.
 *
 * The naive design is shown fairly before it is criticised: it is not stupid,
 * it is what everyone writes first, and it works right up until the day it
 * does not. The break has to cost money on screen rather than be declared,
 * so the deposit is the real number from the golden path.
 */
export const S5Break: React.FC = () => {
  const frame = useCurrentFrame();

  return (
    <Stage eyebrow="TẬP 1 · CẢNH 5">
      <Cues
        items={[
          { at: 20, sound: "appear" },
          { at: 60, sound: "reveal" },
          { at: 300, sound: "crack", volume: 0.24 },
        ]}
      />
      <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
        <Headline delay={4} size={56}>
          Phản xạ đầu tiên: nhét tất cả vào một bảng
        </Headline>

        <div style={{ marginTop: 30, flex: 1 }}>
          <Split
            align="flex-start"
            leftWidth={840}
            gap={60}
            left={
              <div style={{ display: "flex", gap: 30, alignItems: "flex-start" }}>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <SketchCard label="cách quen thuộc" lines={NAIVE} delay={20} />
                </div>
                <TxBracket height={300} delay={60} breakAt={260} />
              </div>
            }
            right={
              <div style={{ display: "flex", flexDirection: "column", gap: 28 }}>
                <Bullets
                  gap={22}
                  items={[
                    {
                      at: 120,
                      tag: "ngày 1",
                      text: "Khách trả cọc. Ghi một dòng. Mọi thứ đúng.",
                      accent: COLOR.ok,
                    },
                    {
                      at: 200,
                      tag: "ngày 2",
                      text: "Shop báo hết size US 9. Giờ phải sửa cả trạng thái đơn lẫn trạng thái việc mua cùng lúc.",
                      accent: COLOR.danger,
                    },
                  ]}
                />

                {/* The money already moved. That is what makes this expensive. */}
                <Interactive.Div
                  name="Deposit already paid"
                  style={{
                    display: "flex",
                    alignItems: "baseline",
                    gap: 18,
                    padding: "24px 30px",
                    borderRadius: RADIUS.md,
                    backgroundColor: COLOR.surface,
                    border: `1px solid ${COLOR.line}`,
                    opacity: interpolate(frame, [230, 256], [0, 1], {
                      extrapolateLeft: "clamp",
                      extrapolateRight: "clamp",
                      easing: EASE,
                    }),
                  }}
                >
                  <span style={{ fontSize: TYPE.label, color: COLOR.textDim }}>
                    Khách đã trả
                  </span>
                  <span
                    style={{
                      fontFamily: MONO,
                      fontSize: TYPE.sub,
                      fontWeight: 700,
                      color: COLOR.money,
                    }}
                  >
                    {GOLDEN.depositVND} ₫
                  </span>
                </Interactive.Div>

                <Bullets
                  gap={22}
                  items={[
                    {
                      at: 300,
                      tag: "chỗ vỡ",
                      text: "Shop ở Mỹ không biết transaction của mình tồn tại, và không chờ nó.",
                      accent: COLOR.danger,
                    },
                    {
                      at: 350,
                      text: "Một transaction chỉ giữ được thứ database giữ được. Nó không giữ được ba mươi ngày và một công ty khác.",
                      accent: COLOR.danger,
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
