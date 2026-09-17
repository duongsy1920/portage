import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { TimeAxis, type Milestone } from "../../../../components/TimeAxis";
import { COLOR, SAFE } from "../../../../design/tokens";
import { Cues } from "../../../../components/Sfx";

/**
 * Scene 4 — the calendar.
 *
 * Before any argument about transactions can land, the viewer has to feel that
 * one order is not an instant. The marks sit at their real day, so the three
 * empty weeks are visible rather than asserted.
 */
const MILESTONES: Milestone[] = [
  { day: 1, label: "Ngày 1", what: "Khách đặt và trả cọc 50 %", ctx: "ordering", at: 40 },
  { day: 2, label: "Ngày 2", what: "Nhân viên mua ở shop Mỹ", ctx: "procurement", at: 75 },
  { day: 9, label: "Ngày 9", what: "Hàng về kho Denver", ctx: "logistics", at: 110 },
  { day: 20, label: "Ngày 20", what: "Gom lô, bay về", ctx: "logistics", at: 145 },
  { day: 30, label: "Ngày 30", what: "Giao tận nhà ở Việt Nam", ctx: "catalog", at: 180 },
];

export const S4Thirty: React.FC = () => {
  return (
    <Stage eyebrow="TẬP 1 · CẢNH 4" source="Nghiệp vụ thật, xem docs/FLOW-ORDER.md">
      <Cues
        items={[
          { at: 20, sound: "appear" },
          { at: 75, sound: "waiting", volume: 0.22 },
          { at: 320, sound: "reveal" },
        ]}
      />
      <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
        <Headline delay={4} size={56}>
          Một đơn hàng không phải một khoảnh khắc
        </Headline>

        <div style={{ marginTop: 10 }}>
          <TimeAxis milestones={MILESTONES} width={SAFE.contentWidth} />
        </div>

        <div style={{ display: "flex", gap: 70, marginTop: 6 }}>
          <div style={{ width: 810 }}>
            <Bullets
              gap={22}
              items={[
                {
                  at: 220,
                  tag: "chú ý",
                  text: "Mọi việc mình tự làm được đều xong trong hai ngày đầu. Phần còn lại là chờ.",
                  accent: COLOR.money,
                },
                {
                  at: 270,
                  text: "Ba tuần đó mình không điều khiển được ai: shop bên Mỹ, kho ở Denver, hãng bay, hải quan.",
                  accent: COLOR.money,
                },
              ]}
            />
          </div>
          <div style={{ width: 810 }}>
            <Bullets
              gap={22}
              items={[
                {
                  at: 320,
                  tag: "và",
                  text: "Tiền của khách đã vào tài khoản mình từ ngày 1, trước khi bất cứ rủi ro nào lộ ra.",
                  accent: COLOR.danger,
                },
                {
                  at: 370,
                  tag: "câu hỏi",
                  text: "Ba mươi ngày đó có gói vào một transaction được không?",
                  accent: COLOR.danger,
                },
              ]}
            />
          </div>
        </div>
      </div>
    </Stage>
  );
};
