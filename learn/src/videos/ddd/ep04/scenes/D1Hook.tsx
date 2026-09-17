import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline, Callout } from "../../../../components/Text";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { DDD_SNIPPETS } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Season 2, episode 4, scene 1 — who declares the repository.
 *
 * In Symfony the repository extends a Doctrine base class, so the domain knows
 * Doctrine exists. Here the domain writes an interface and nothing more. The
 * single most useful sentence of the episode is on the card: this file imports
 * only the standard library.
 */
export const D1Hook: React.FC = () => (
  <Stage eyebrow="MÙA 2 · TẬP 4" source="internal/domain/catalog/repository.go:34 — domain khai, không cài">
    <Cues items={[{ at: 16, sound: "appear" }, { at: 160, sound: "land", volume: 0.26 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={58}>
        Domain viết interface, và dừng ở đó
      </Headline>

      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <Split
          leftWidth={1020}
          gap={56}
          align="center"
          left={<CodeCard snippet={DDD_SNIPPETS.repoPort} delay={20} accent={COLOR.go} scale={0.94} />}
          right={
            <div style={{ fontSize: 34, lineHeight: 1.45, color: COLOR.textDim }}>
              File này import đúng{" "}
              <span style={{ color: COLOR.ok, fontWeight: 700 }}>stdlib</span>.
              <br />
              Không Doctrine, không pgx,
              <br />
              không biết database nào tồn tại.
            </div>
          }
        />
      </div>

      <div style={{ marginBottom: 26 }}>
        <Callout delay={160} accent={COLOR.php}>
          Bên Symfony, <span style={{ color: COLOR.php }}>MerchantRepository extends ServiceEntityRepository</span> —
          domain biết Doctrine tồn tại. Ở đây chiều phụ thuộc bị đảo lại.
        </Callout>
      </div>
    </div>
  </Stage>
);
