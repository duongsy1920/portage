import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { TermCard } from "../../../../components/TermCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 3 — into[T], and the case where T must be written out.
 *
 * `into` has no parameter of type T, so there is nothing to infer from; the
 * decoder table names the type in brackets. Seeing both forms side by side
 * is what makes "type inference" a thing the viewer can predict rather than a
 * thing that sometimes happens.
 */
export const S3Into: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 13 · CẢNH 3 · TỪ VỰNG" source="internal/adapter/eventcodec/decode.go:66">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 330, sound: "appear", volume: 0.14 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        into[T]: không có gì để suy, nên ghi rõ T
      </Headline>

      <div style={{ marginTop: 22, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={720}
          gap={48}
          left={<CodeCard snippet={GO_SNIPPETS.intoGeneric} delay={20} accent={COLOR.go} highlight={[0, 1]} />}
          right={
            <div style={{ display: "flex", flexDirection: "column", gap: 20 }}>
              <TermCard
                delay={150}
                accent={COLOR.go}
                term="type parameter"
                plain="Một kiểu để trống trong ngoặc vuông, [T any]. Hàm viết một lần; mỗi lần gọi điền một kiểu, compiler kiểm như kiểu thật."
                symfony="@template T trong PHPDoc (Psalm, PHPStan) — nhưng đó là chú thích cho công cụ ngoài; Go thì compiler tự kiểm."
                inRepo="into[contracts.ProductPublishedV1] — ghi rõ, vì tham số chỉ là []byte"
              />
              <Bullets
                gap={16}
                items={[
                  { at: 330, tag: "var v T", text: "Khai một biến kiểu T, zero value của T (tập 6). json.Unmarshal điền vào, rồi trả ra dưới dạng any.", accent: COLOR.money },
                  { at: 390, tag: "vì sao trả any", text: "Bảng decoders phải chứa mọi decoder cùng một kiểu hàm. Kiểu thật được hỏi lại ở on[T] bằng m, ok := msg.(T).", accent: COLOR.ok },
                ]}
              />
            </div>
          }
        />
      </div>
    </div>
  </Stage>
);
