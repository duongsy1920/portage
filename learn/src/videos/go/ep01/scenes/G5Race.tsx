import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Season 1, scene 5 — the bill for staying alive.
 *
 * This is the first consequence a PHP developer has genuinely never had to
 * think about, so it gets its own scene and a real lock from the repository
 * rather than a toy example.
 */
export const G5Race: React.FC = () => {
  return (
    <Stage
      eyebrow="MÙA 1 · TẬP 1 · CẢNH 5"
      source={`internal/adapter/memory — repo có ${GO_USAGE.mutex} chỗ khoá như thế này`}
    >
      <Cues
        items={[
          { at: 20, sound: "appear" },
          { at: 150, sound: "snap", volume: 0.2 },
          { at: 300, sound: "land", volume: 0.3 },
        ]}
      />
      <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
        <Headline delay={4} size={56}>
          Cái giá của việc sống lâu: hai request giẫm lên nhau
        </Headline>

        <div style={{ marginTop: 30, flex: 1 }}>
          <Split
            align="flex-start"
            leftWidth={840}
            gap={60}
            left={
              <CodeCard
                snippet={GO_SNIPPETS.mutex}
                delay={20}
                accent={COLOR.danger}
                highlight={[1, 2, 10, 11]}
              />
            }
            right={
              <Bullets
                gap={26}
                items={[
                  {
                    at: 110,
                    tag: "bên php",
                    text: "Hai request là hai process khác nhau, không chia sẻ gì. Bạn chưa từng phải nghĩ tới chuyện này.",
                    accent: COLOR.php,
                  },
                  {
                    at: 170,
                    tag: "bên go",
                    text: "byID là một map sống trong bộ nhớ của process. Hai goroutine ghi vào cùng lúc thì dữ liệu hỏng, và Go báo lỗi ngay khi phát hiện.",
                    accent: COLOR.danger,
                  },
                  {
                    at: 240,
                    tag: "cách chữa",
                    text: "sync.Mutex: ai vào thì khoá, xong thì mở. defer bảo đảm luôn mở kể cả khi hàm thoát giữa chừng.",
                    accent: COLOR.go,
                  },
                  {
                    at: 310,
                    tag: "phỏng vấn",
                    text: "Câu hay hỏi: chạy go test -race để làm gì? Để Go dò xem có hai goroutine chạm cùng một ô nhớ mà không khoá.",
                    accent: COLOR.ok,
                  },
                ]}
              />
            }
          />
        </div>
      </div>
    </Stage>
  );
};
