import React from "react";
import { Stage, Split } from "../../../../components/Stage";
import { Headline } from "../../../../components/Text";
import { Bullets } from "../../../../components/Bullets";
import { TermCard } from "../../../../components/TermCard";
import { Cues } from "../../../../components/Sfx";
import { GO_USAGE } from "../../../../data/portage";
import { COLOR } from "../../../../design/tokens";

/**
 * Scene 2 — the word "select", and the one design choice around it in
 * the repository: the sweeper waits one tick before its first pass, the relay
 * does not. The reason is in sweep.go's own comment and is worth repeating:
 * a job that only has a clock gains nothing by scanning at boot.
 */
export const T2Select: React.FC = () => (
  <Stage eyebrow="MÙA 1 · TẬP 14 · CẢNH 2 · TỪ VỰNG" source={`internal/worker/sweep.go:56 — repo có ${GO_USAGE.selects} khối select`}>
    <Cues items={[{ at: 20, sound: "appear" }, { at: 150, sound: "reveal" }, { at: 340, sound: "appear", volume: 0.14 }]} />
    <div style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Headline delay={4} size={56}>
        select: đứng chờ nhiều channel, cái nào tới trước thì đi
      </Headline>

      <div style={{ marginTop: 22, flex: 1 }}>
        <Split
          align="flex-start"
          leftWidth={820}
          gap={52}
          left={
            <TermCard
              delay={20}
              accent={COLOR.go}
              term="select"
              plain="Một khối chờ. Mỗi case là một lần đọc hoặc ghi channel; nhánh nào sẵn sàng trước thì chạy, các nhánh kia bỏ qua. Nhiều nhánh cùng sẵn sàng: chọn ngẫu nhiên."
              symfony="Không có. Gần nhất là stream_select() hoặc một event loop — nhưng ở Go nó là cú pháp của ngôn ngữ."
              inRepo="worker.go:134 và sweep.go:60 — cùng hai case: dừng, hoặc nhịp kế"
            />
          }
          right={
            <Bullets
              gap={20}
              items={[
                { at: 150, tag: "dừng thế nào", text: "Không ai gọi hàm Stop(). Người giữ ctx gọi cancel(), ctx.Done() đóng, select chạy nhánh return ctx.Err().", accent: COLOR.danger },
                { at: 220, tag: "ctx.Err()", text: "Trả về lý do dừng: Canceled hay DeadlineExceeded. Người gọi Run biết mình dừng vì gì.", accent: COLOR.money },
                { at: 290, tag: "một khác biệt nhỏ", text: "Relay chạy một lượt rồi mới chờ. Sweeper chờ một nhịp rồi mới chạy — vì lúc khởi động, relay có hàng đợi sẵn, sweeper chỉ có đồng hồ.", accent: COLOR.go },
                { at: 360, tag: "default", text: "select có thể có nhánh default: không channel nào sẵn thì chạy nó ngay, không chờ. Repo không dùng — cả hai chỗ đều muốn chờ.", accent: COLOR.ok },
              ]}
            />
          }
        />
      </div>
    </div>
  </Stage>
);
