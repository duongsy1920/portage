import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { CodeCard } from "../../../../components/CodeCard";
import { Cues } from "../../../../components/Sfx";
import { GO_SNIPPETS, GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 3 — httptest, through the helper every HTTP test in the repo calls.
 *
 * No port is opened: a request object goes straight into the real handler
 * and a recorder catches the response. The helper adds the operator token
 * by default so three hundred tests about routes do not each spell out a
 * header — and a test about the door uses callAnon.
 */
export const V3Httptest: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 16 · CẢNH 3" source={`internal/adapter/http/server_test.go:166 — httptest ở ${GO_USAGE.httptestFiles} file`}>
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 340, sound: "land", volume: 0.24 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        httptest: một request đi qua handler thật, không mở cổng nào
      </Headline>

      <div style={{ marginTop: 22, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={880}
          gap={48}
          left={<CodeCard snippet={GO_SNIPPETS.callHelper} delay={20} accent={COLOR.go} highlight={[1, 7]} />}
          right={
            <Bullets
              gap={20}
              items={[
                { at: 150, tag: "NewRequest", text: "Dựng một *http.Request trong bộ nhớ: method, path, body. Không socket, không cổng, không mạng.", accent: COLOR.go },
                { at: 210, tag: "ResponseRecorder", text: "Một http.ResponseWriter ghi lại status, header, body để test đọc. Handler không biết nó không phải client thật.", accent: COLOR.money },
                { at: 270, tag: "cùng cái cửa", text: "Request đi qua đúng authenticate() mà curl đi qua. Test về 401 dùng callAnon — không có token nào được thêm hộ.", accent: COLOR.ok },
                { at: 340, tag: "bên symfony", text: "WebTestCase + static::createClient() — cũng không mở cổng, nhưng dựng cả kernel. Ở đây chỉ dựng handler và graph bộ nhớ.", accent: COLOR.php },
              ]}
            />
          }
        />
      </div>
    </div>
  </Stage>
);
