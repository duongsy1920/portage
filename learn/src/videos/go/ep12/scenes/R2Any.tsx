import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { TermCard } from "../../../../components/TermCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 2 — the empty interface, by its real use in this codebase.
 *
 * `any` is not taught as a feature; it is taught as the price the codec pays
 * to write payload keys by hand. The map's values are `any` so a value can be
 * a string, a number, or another map — and the compiler knows nothing more.
 */
export const R2Any: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 12 · CẢNH 2 · TỪ VỰNG" source={`internal/adapter/eventcodec/codec.go:45 — repo có ${GO_USAGE.anyUses} chỗ dùng any`}>
    <Cues items={[{ at: 20, sound: "appear" }, { at: 140, sound: "reveal" }, { at: 320, sound: "appear", volume: 0.14 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        any: interface không đòi gì, nên nhận mọi thứ
      </Headline>

      <div style={{ marginTop: 26, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={760}
          gap={52}
          left={
            <TermCard
              delay={20}
              accent={COLOR.money}
              term="any"
              plain="Tên mới của interface{}: interface không có method nào, nên kiểu nào cũng thoả. Cái giá: compiler không còn biết đó là gì."
              symfony="mixed — và cũng vì thế mà phải is_string(), instanceof trước khi dùng."
              inRepo="type m = map[string]any — codec.go:45"
            />
          }
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 22 }}>
            <CodeCard snippet={GO_SNIPPETS.payloadMap} delay={140} accent={COLOR.money} />
            <Bullets
              gap={22}
              items={[
                { at: 320, tag: "khi nào dùng", text: "Khi hình dạng chỉ biết lúc chạy: payload JSON, giá trị trong context, dữ liệu từ ngoài vào.", accent: COLOR.go },
                { at: 380, tag: "khi nào không", text: "Khi có thể viết kiểu thật. Repo không có hàm nghiệp vụ nào nhận any — domain chỉ có kiểu cụ thể.", accent: COLOR.danger },
                { at: 440, tag: "cái giá", text: "Mọi thứ đã bỏ vào any phải hỏi lại mới lấy ra được. Hai cách hỏi: type switch và x.(T).", accent: COLOR.money },
              ]}
            />
            </div>
          }
        />
      </div>
    </div>
  </Stage>
);
