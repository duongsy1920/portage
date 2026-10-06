# learn/ — loạt video học Go và DDD

30 tập video, mỗi tập kèm một bản in A3 để ôn lại, cộng hai bản in tra cứu không có video. Dựng bằng
[Remotion](https://remotion.dev) (React → mp4). **Không dính gì tới code Go** của
Portage: thư mục này không import gì từ repo, và repo không import gì từ đây.

Nội dung, thứ tự tập và lý do của thứ tự đó: [`CURRICULUM.md`](CURRICULUM.md).

```
out/mua1-go/       16 tập — Go cho người đã viết PHP sáu năm (10 + 6 bổ sung theo roadmap.sh)
out/mua2-ddd/       6 tập — DDD, giải thích bằng ví dụ Symfony đã quen
out/mua3-portage/   8 tập — vì sao project thật này trông như vậy
```

Mỗi thư mục tập có: `video.mp4` · `cheatsheet.png` · `cheatsheet.pdf` · `canh/`
(ảnh từng cảnh, để soi nhanh không cần mở video). `out/` không vào git.

## Trang học — `hub/`

Một trang web để **đi hết lộ trình** thay vì mở Studio. Bố cục kiểu khoá học: **danh sách
bài bên trái** (ba mùa, tiến độ, hoặc nhóm theo buổi của HOC.md), **bài đang xem bên phải**:
player `@remotion/player` chạy thẳng composition (không cần mp4), rồi bốn tab — mục lục các
đoạn (bấm là tua), cheatsheet A3 trong trang, code trong repo link tới đúng commit đã kiểm,
tự kiểm tra. Cộng trang Ôn tập theo lịch 1/3/7 ngày và bảng so với roadmap.sh. Tiến độ lưu
trong trình duyệt, tải về / nạp từ file JSON. Giao diện: **shadcn/ui + Tailwind v4 +
radix-ui** (`components.json` ở `learn/`, component trong `hub/src/components/ui/`), sáng/tối
theo hệ điều hành, font Be Vietnam Pro tự host. Plan gốc: `HUB-PLAN.md`; làm lại 06/10 theo
`HUB-REDESIGN-PLAN.md` (bố cục hai cột, shadcn, và bảng từ: thuật ngữ giữ tiếng Anh).

```bash
npm run hub            # http://localhost:5173/  (dev, sửa cảnh là thấy ngay)
npm run hub:build      # hub/dist/ — tĩnh, mở ở bất kỳ đường dẫn nào
npm run hub:preview    # xem bản build ở http://localhost:4173/
```

Dữ liệu của trang là dữ liệu của video, không chép: mỗi tập có `recap.ts` (4 ý · 3 câu phỏng
vấn · tập sau) mà **cả cảnh "Nhớ lại" lẫn trang học cùng đọc**, và `anchors` trong `spec.ts`
sinh từ chính các cảnh. Cửa kiểm thứ năm, `scripts/verify-hub.py`, canh việc đó.

## Lệnh

```bash
npm i                              # lần đầu
npx remotion studio --no-open      # xem và chỉnh từng cảnh

./scripts/render-plates.sh         # 32 bản in (30 tập + 2 tra cứu) → PNG @2x + PDF A3 ngang
npx remotion render Go-Ep01-VongDoi out/mua1-go/ep01-vong-doi/video.mp4
python3 scripts/stills.py          # ảnh từng cảnh cho mọi tập
```

## Bốn cửa kiểm — chạy trước khi tin là xong

`tsc` xanh và preview trong Studio **không** bắt được lỗi hình. Hơn mười lỗi thật
đã lọt qua cả hai. Bốn lệnh này thì bắt được, và mỗi lệnh sinh ra từ một lỗi đã
xảy ra thật:

```bash
./scripts/verify-numbers.sh        # 32 con số trong data/portage.ts vs repo Go
python3 scripts/verify-snippets.py # mọi dòng code lên hình có thật trong file nó khai
python3 scripts/check-overflow.py  # 147 cảnh, không cảnh nào tràn khung
python3 scripts/verify-hub.py      # trang học không nói sai về video: recap, anchors, roadmap, HOC
node scripts/hub-check.mjs <url>   # trình duyệt thật: 30 tập phát tới cuối, 3 bề rộng, ngăn kéo, dark mode, tiến độ, chữ cũ
```

Và sau khi render, mọi mp4 phải có đúng một track âm thanh:

```bash
for f in out/*/*/video.mp4; do
  echo "$f $(node_modules/@remotion/compositor-linux-x64-gnu/ffmpeg -i "$f" 2>&1 | grep -c 'Audio: aac')"
done
```

## Thêm một tập

Một `spec.ts`, vài cảnh, và **một dòng** trong `src/videos/index.ts`. Không sửa
`Root.tsx` (nó duyệt danh sách), không sửa script nào (`scripts/episodes.py` tự
đọc mọi `spec.ts`).

Mọi con số và mọi đoạn code lên hình lấy từ `src/data/portage.ts`, mỗi mảnh kèm
đường dẫn và số dòng. Cảnh không được tự viết ra một sự thật về codebase.

## Vì sao có `go.mod` trong một project TypeScript

`go test ./...` ở gốc repo đi vào `node_modules` và nhặt phải một package Go nằm
lẫn trong một thư viện JavaScript. Khai một module lồng làm module gốc dừng lại ở
đây. Không có dòng Go nào trong thư mục này được build.
