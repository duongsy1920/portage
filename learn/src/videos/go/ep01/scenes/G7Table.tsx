import React from "react";
import { Stage } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { CompareTable, Punchline, type CompareRow } from "../../../../components/CompareTable";
import { Cues } from "../../../../components/Sfx";
import { SAFE } from "../../../../design/tokens";

/**
 * Season 1, scene 7 — the delta, in one table.
 *
 * Everything in this table was derived on screen in the previous scenes, so
 * the table is a place to consolidate rather than new information. That is
 * what makes it worth pausing on.
 */
const ROWS: CompareRow[] = [
  {
    about: "Vòng đời",
    php: "Dựng lại cả thế giới mỗi request",
    go: "Dựng một lần, phục vụ mọi request",
    at: 40,
  },
  {
    about: "Kết nối database",
    php: "Mở lại mỗi request, hoặc pool nằm ngoài",
    go: "Một pool duy nhất, dựng ở main()",
    at: 80,
  },
  {
    about: "Biến toàn cục",
    php: "Chết theo request, vô hại",
    go: "Sống mãi, phải khoá bằng sync.Mutex",
    at: 120,
  },
  {
    about: "Đồng thời",
    php: "Nhiều process PHP-FPM",
    go: "Nhiều goroutine trong MỘT process",
    at: 160,
  },
  {
    about: "Huỷ giữa chừng",
    php: "Không có",
    go: "context.Context",
    at: 200,
  },
  {
    about: "Ráp phụ thuộc",
    php: "services.yaml, autowire",
    go: "Viết tay trong main(), sai là không compile",
    at: 240,
  },
  {
    about: "Rò rỉ bộ nhớ",
    php: "Gần như không gặp",
    go: "Có thật, tích luỹ theo ngày",
    at: 280,
  },
];

export const G7Table: React.FC = () => {
  return (
    <Stage eyebrow="MÙA 1 · TẬP 1 · CẢNH 7">
      <Cues
        items={[
          { at: 40, sound: "appear", volume: 0.12 },
          { at: 120, sound: "appear", volume: 0.12 },
          { at: 200, sound: "appear", volume: 0.12 },
          { at: 280, sound: "appear", volume: 0.12 },
          { at: 330, sound: "land" },
        ]}
      />
      <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
        <Headline delay={4} size={56}>
          Bảy dòng, và tất cả đều từ một khác biệt
        </Headline>

        <div style={{ marginTop: 26 }}>
          <CompareTable rows={ROWS} width={SAFE.contentWidth} />
          <Punchline at={330}>
            Không phải bảy điều phải học thuộc. Là một điều, nhìn từ bảy phía.
          </Punchline>
        </div>
      </div>
    </Stage>
  );
};
