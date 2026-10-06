import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 4 — the two exits and the one entrance.
 *
 * JSON leaves the domain through the outbox (contracts) and through HTTP
 * (views). It comes back in through a consumer that PARSES the contract into
 * domain types — ParseOrderID, moneyFrom — so a bad message is refused at the
 * door, by the domain's own constructors (convention 8).
 */
export const U4Exit: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 15 · CẢNH 4" source={`encoding/json được import ở ${GO_USAGE.jsonFiles} file chạy thật — tất cả nằm ngoài domain`}>
    <Cues items={[{ at: 20, sound: "appear" }, { at: 130, sound: "reveal" }, { at: 350, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Hai cửa ra khỏi domain, và một cửa vào
      </Headline>

      <div style={{ marginTop: 22, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={820}
          gap={48}
          left={
            <div style={{ display: "flex", flexDirection: "column", gap: 22 }}>
              <CodeCard snippet={GO_SNIPPETS.moneyView} delay={20} accent={COLOR.ok} highlight={[1]} />
              <CodeCard snippet={GO_SNIPPETS.projectorReads} delay={130} accent={COLOR.go} highlight={[2, 4]} />
            </div>
          }
          right={
            <Bullets
              gap={20}
              items={[
                { at: 230, tag: "cửa ra 1: outbox", text: "Event → contract V1 → JSON, viết tay trong codec. Người nhận là context khác.", accent: COLOR.money },
                { at: 290, tag: "cửa ra 2: http", text: "Aggregate → view → JSON. Tiền là chuỗi \"150.00\", không bao giờ float — màn hình hiển thị, không cộng trừ.", accent: COLOR.ok },
                { at: 350, tag: "cửa vào", text: "Consumer nhận contract rồi Parse ngược về kiểu domain. Chuỗi sai là lỗi ở đây, không phải sâu trong aggregate.", accent: COLOR.go },
                { at: 410, tag: "quy ước 8", text: "Adapter không tự kiểm. Nó gọi ParseOrderID, ParseMoney — luật vẫn ở domain, adapter chỉ chuyển dạng.", accent: COLOR.php },
              ]}
            />
          }
        />
      </div>
    </div>
  </Stage>
);
