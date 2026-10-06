import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 4 — a named return, and the deferred function that overwrites it.
 *
 * Episode 9 showed this defer for recover; this scene reads the same lines
 * for the other thing they do: `err = fmt.Errorf("commit: %w", cerr)` changes
 * what the function returns AFTER `return fn(...)` has already run. That is
 * only possible because the result has a name.
 */
export const Q4Named: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 11 · CẢNH 4" source={`internal/adapter/postgres/postgres.go:83 — repo có ${GO_USAGE.namedReturns} hàm trả về có tên`}>
    <Cues items={[{ at: 20, sound: "appear" }, { at: 110, sound: "reveal" }, { at: 320, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        (err error): kết quả có tên, nên defer sửa được nó
      </Headline>

      <div style={{ marginTop: 22, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={700}
          gap={44}
          left={
            <div style={{ display: "flex", flexDirection: "column", gap: 22 }}>
              <CodeCard snippet={GO_SNIPPETS.inTxNamed} delay={20} accent={COLOR.go} highlight={[1]} />
              <Bullets
                gap={16}
                items={[
                  {
                    at: 200,
                    tag: "không tên",
                    text: "return fn(…) chạy xong là chốt. Defer chạy sau nhưng không đổi được nó.",
                    accent: COLOR.danger,
                  },
                  {
                    at: 260,
                    tag: "có tên",
                    text: "err là biến thật của hàm. Defer gán err = … là đổi đúng thứ hàm trả.",
                    accent: COLOR.ok,
                  },
                  {
                    at: 320,
                    tag: "bên symfony",
                    text: "Không có. finally không đổi được giá trị return.",
                    accent: COLOR.php,
                  },
                ]}
              />
            </div>
          }
          right={<CodeCard snippet={GO_SNIPPETS.recoverTx} delay={110} accent={COLOR.money} highlight={[9, 10]} />}
        />
      </div>
    </div>
  </Stage>
);
