import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline, Body } from "../../../../components/Text";
import { Anatomy } from "../../../../components/Anatomy";
import { Bullets } from "../../../../components/Bullets";
import { Cues } from "../../../../components/Sfx";
import { COLOR, SAFE } from "../../../../design/tokens";

/**
 * Scene 2 — one field of the contract, taken apart.
 *
 * The thing to land: the tag is a plain string the compiler ignores.
 * encoding/json reads it at run time through reflection. That is why a typo
 * in a tag is not a compile error — and why the codec writes payload keys by
 * hand instead (episode 12).
 */
export const U2Tag: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 15 · CẢNH 2" source="internal/contracts/ordering_v1.go:18">
    <Cues items={[{ at: 20, sound: "appear" }, { at: 80, sound: "reveal" }, { at: 250, sound: "snap", volume: 0.18 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Một tag là một chuỗi đứng sau kiểu — compiler không đọc, thư viện đọc
      </Headline>

      <div style={{ flex: 1, display: "flex", flexDirection: "column", justifyContent: "center", gap: 56 }}>
        <Anatomy
          width={SAFE.contentWidth}
          fontSize={36}
          parts={[
            { text: "Total", note: "tên field, chữ hoa — thấy được từ package khác (tập 2)", at: 80, color: COLOR.go },
            { text: "    " },
            { text: "MoneyV1", note: "kiểu: một struct V1 khác, không phải shared.Money", at: 130, color: COLOR.money },
            { text: "   " },
            { text: '`json:"total"`', note: "struct tag — chuỗi thô, encoding/json đọc bằng reflection để đặt tên key", at: 190, color: COLOR.ok },
          ]}
        />

        <Body delay={250} size={34}>
          Không có tag, key sẽ là <span style={{ color: COLOR.text, fontWeight: 600 }}>"Total"</span> — đúng tên field.
          Có tag, key là <span style={{ color: COLOR.ok, fontWeight: 600 }}>"total"</span>. Gõ sai tag thì vẫn biên dịch; chỉ JSON ra sai.
        </Body>

        <Bullets
          gap={18}
          items={[
            { at: 320, tag: "hai tuỳ chọn hay gặp", text: "`json:\"x,omitempty\"` bỏ key khi giá trị rỗng · `json:\"-\"` không bao giờ xuất field này.", accent: COLOR.money },
            { at: 380, tag: "bên symfony", text: "#[SerializedName('total')] trên property — cũng là metadata đọc lúc chạy, bằng attribute thay vì chuỗi.", accent: COLOR.php },
          ]}
        />
      </div>
    </div>
  </Stage>
);
