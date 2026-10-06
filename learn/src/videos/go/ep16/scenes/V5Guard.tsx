import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, REPO } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 5 — a decision guard: an architecture rule that fails the build.
 *
 * Season 3 showed guard 7 for bounded contexts; this scene reads guard 6,
 * the allowlist, as a TEST — go/ast parses every domain file and checks each
 * import. The point for a test episode: a rule in a document is forgotten, a
 * rule in a test is a red build.
 */
export const V5Guard: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 16 · CẢNH 5" source={`internal/domain/decisions_test.go:225 — ${REPO.guards} guard, cùng một file`}>
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 340, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        Một luật kiến trúc viết thành test: vi phạm là build đỏ
      </Headline>

      <div style={{ marginTop: 22, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={900}
          gap={48}
          left={<CodeCard snippet={GO_SNIPPETS.guardAllowlist} delay={20} accent={COLOR.danger} highlight={[1, 2, 7]} />}
          right={
            <Bullets
              gap={20}
              items={[
                { at: 150, tag: "go/ast", text: "Thư viện chuẩn đọc được chính code Go: parse từng file domain, duyệt danh sách import. Test không chạy code, nó đọc code.", accent: COLOR.go },
                { at: 210, tag: "luật gì", text: "internal/domain chỉ được import thư viện chuẩn và một allowlist (uuid). Thêm yaml vào domain là test đỏ — T-001 đã đi vòng qua đúng luật này.", accent: COLOR.money },
                { at: 270, tag: "vì sao có", text: "Lỗi thật: CategoryPolicy từng được thêm hsCode sau khi đã quyết hai lần là không. Mọi unit test xanh, thiết kế vẫn sai.", accent: COLOR.danger },
                { at: 340, tag: "khi đỏ", text: "Hoặc sửa code, hoặc đổi quyết định và sửa guard — nhưng phải cố ý. Đó là khác biệt giữa tripwire và điều răn.", accent: COLOR.ok },
              ]}
            />
          }
        />
      </div>
    </div>
  </Stage>
);
