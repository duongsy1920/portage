Ngôn ngữ:  `summaries.go:71` — `if v, err := …; err == nil` để nuốt lỗi CÓ chủ ý: hợp lệ khi lỗi chỉ xảy ra với dữ liệu hỏng và hậu quả thấy được (trường thiếu)
Ngôn ngữ:  `wire_test.go:297` — tìm dòng theo id thay vì `[0]`: danh sách "mới nhất trước" đổi thứ tự khi test thêm đơn thứ hai
Ngôn ngữ:  `summaries_test.go:57` — bảng test có hàm trong hàng (`as func()`), một mẫu Go cho "cùng khẳng định, hai route"
Domain:    `summaries.go:43` — giá trị dẫn xuất trên bảng đọc: tính lúc đọc khi hai vế cùng dòng; chỉ lưu khi cần lọc/sắp — Symfony: một getter trên DTO, không phải cột
Domain:    `order.go:196` ↔ `summaries.go:71` — cùng một phép trừ ở hai nơi là chấp nhận được khi cả hai gọi cùng `Money.Sub`; chép sang JS thì không
Kỹ thuật:  `words.js:233` — bỏ `BigInt` khỏi màn hình: số tiền không bao giờ tính ở client, chỉ hiển thị
Kỹ thuật:  `api.js:124` — `money(undefined)` in "—": chỗ gọi không cần guard vì helper đã quyết cách nói "không có"
Agent:     kiểm biên dịch bằng `go vet` trên package (kể cả `_test.go`) bắt 2 lỗi trước gate — bài học T-001 dùng được ngay
Agent:     vòng ADR-007 đầu tiên: 0 điểm dừng trước cửa commit, gate xanh lần đầu
Sai ở đâu: dòng `commands.smoke` tôi ghi hộ hôm sáng thiếu `bash` (file không có bit thực thi) — thấy khi đọc lại PROJECT.md trước verify
Nợ mới:    test `node --test` cho `words.js` (`journeyOf`, giờ không test trong repo) · vì sao `go vet` biên dịch test còn `go build` thì không
