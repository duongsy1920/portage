import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { TimeAxis } from "../../../../components/TimeAxis";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Season 3, episode 5, scene 1 — where the customer's money stops being safe.
 *
 * Everything before ConfirmPurchase can be undone with a refund. After it, the
 * shop has been paid and the goods are in the air. The axis puts that moment
 * on a real timeline, because "point of no return" is a phrase until you see
 * which day it falls on.
 */
export const P5S1: React.FC = () => (
  <Stage eyebrow="MÙA 3 · TẬP 5" source="internal/domain/ordering/order.go:239 — ConfirmPurchase">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 180, sound: "snap", volume: 0.2 }, { at: 300, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Từ đây trở đi, hoàn tiền không còn là một lựa chọn
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <TimeAxis
          width={SAFE.contentWidth}
          pad={150}
          milestones={[
            { day: 0, label: "đặt đơn", what: "khách chốt báo giá", ctx: "ordering", at: 40 },
            { day: 1, label: "trả cọc 50%", what: "tiền khách vào tài khoản mình", ctx: "ordering", at: 90 },
            { day: 2, label: "mua ở shop Mỹ", what: "điểm không thể quay đầu", ctx: "procurement", at: 180 },
            { day: 9, label: "gom lô ở Denver", what: "chia cước cho từng kiện", ctx: "logistics", at: 240 },
            { day: 30, label: "giao tận nhà", what: "trả nốt số dư", ctx: "ordering", at: 280 },
          ]}
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={300} accent={COLOR.danger}>
          Trước ngày 2, huỷ đơn là hoàn cọc. Sau ngày 2, tiền đã nằm ở shop Mỹ và hàng
          đang bay — luật này có tiền thật của khách đằng sau, không phải một enum.
        </Callout>
      </div>
    </div>
  </Stage>
);
