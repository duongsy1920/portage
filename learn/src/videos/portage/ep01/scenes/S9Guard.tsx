import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { SNIPPETS, REPO } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";
import { Cues } from "../../../../components/Sfx";

/**
 * Scene 9 — the rule is enforced, not agreed.
 *
 * This is what separates the repository from a folder convention: the boundary
 * is a failing test, so crossing it turns the build red instead of starting a
 * code-review conversation somebody may lose.
 */
export const S9Guard: React.FC = () => {
  return (
    <Stage
      eyebrow="TẬP 1 · CẢNH 9"
      source={`${REPO.guards} test canh kiến trúc, đọc code bằng go/ast`}
    >
      <Cues
        items={[
          { at: 30, sound: "appear" },
          { at: 150, sound: "snap", volume: 0.2 },
          { at: 330, sound: "land", volume: 0.26 },
        ]}
      />
      <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
        <Headline delay={4} size={56}>
          Và ranh giới do máy canh, không do người nhắc
        </Headline>

        <div style={{ marginTop: 34, flex: 1 }}>
          <Split
            align="flex-start"
            leftWidth={900}
            gap={60}
            left={
              <CodeCard
                snippet={SNIPPETS.guard}
                delay={30}
                accent={COLOR.danger}
                highlight={[0, 1]}
              />
            }
            right={
              <Bullets
                gap={26}
                items={[
                  {
                    at: 150,
                    tag: "thử xem",
                    text: "Cho ordering import procurement. Go vẫn biên dịch được, không báo gì cả.",
                    accent: COLOR.textDim,
                  },
                  {
                    at: 210,
                    tag: "nhưng",
                    text: "Test này đọc mã nguồn của cả năm vùng và tìm import bắc chéo. Thấy là báo đỏ.",
                    accent: COLOR.danger,
                  },
                  {
                    at: 270,
                    tag: "vì sao",
                    text: "Một quy ước chỉ nằm trong tài liệu sẽ bị vi phạm trong sáu tháng, và không ai biết ngày nó bị vi phạm.",
                    accent: COLOR.danger,
                  },
                  {
                    at: 330,
                    tag: "kết",
                    text: "Ranh giới ở đây không phải lời hứa. Nó là một bài test chạy mỗi lần push.",
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
