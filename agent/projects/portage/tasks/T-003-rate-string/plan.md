# plan — T-003 rate-string (chế độ học: anh viết code)

## Mục tiêu
`Rate.String()` in `8.81%`, `30%`, `-2.5%`, `0.0001%`, `108.81%`, `0%` — bỏ số 0 thừa, không float.

## Gợi ý cho người
- **Bắt đầu ở test, không ở code.** Mở `internal/domain/shared/rate_test.go:16–20`, đổi cột kỳ vọng
  (`"8.8100%"` → `"8.81%"`, `"30.0000%"` → `"30%"`, …), thêm hai dòng: `Rate{}` → `"0%"` và
  `MustParsePercent("8.81").Plus1()` → `"108.81%"`. Chạy `go test ./internal/domain/shared -run ParsePercent`
  — thấy **đỏ đúng chỗ** trước đã.
- **Rồi mới sửa `rate.go:65–72`.** Hướng rẻ nhất: giữ nguyên `fmt.Sprintf("%s%d.%04d", …)`, rồi
  `strings.TrimRight(s, "0")` và `strings.TrimSuffix(s, ".")` trước khi nối `%`. Cẩn thận thứ tự hai lần
  cắt, và cẩn thận số âm (đã có nhánh `sign`).
- **Câu hỏi để tự kiểm** trước khi gọi review: (1) `0.0001%` còn nguyên không — `TrimRight` có cắt quá
  tay không? (2) `Rate{}` in gì? (3) có chỗ nào import `strings` chưa? (4) `go vet ./...` — có test nào ngoài
  `rate_test.go` đỏ không? (câu trả lời mong đợi: không, đã grep).
- **Đừng** đổi `PPM()`, đừng đổi `ParsePercent`, đừng thêm tham số cho `String()`.

## File sẽ đụng — danh sách ĐÓNG
- `internal/domain/shared/rate.go` — `String()`
- `internal/domain/shared/rate_test.go` — kỳ vọng + 2 dòng
- `internal/adapter/config/ratecard_test.go` — CHỈ nếu anh muốn khẳng định `"deposit 150%"` trong lỗi
  (không bắt buộc; hàng `deposit over 100 %` hiện chỉ đòi chứa `"quote"`)

## Test sẽ thêm hoặc đổi
- `rate_test.go`: đổi 5 kỳ vọng, +2 dòng (`Rate{}`, `Plus1`).

## Không làm
- Không dùng `float64`/`%g`. Không đổi `PPM()`. Không đổi JSON hay DB — không chỗ nào ghi chuỗi này.
- Không làm `Money.String()` cùng lúc, dù ngứa tay — task khác.

## Tự phản biện
- **Cách hiển nhiên hơn:** `strconv.FormatFloat(float64(ppm)/10_000, 'f', -1, 64)`. Không chọn vì quy ước 2
  và vì `-1` precision của float in `8.81` hôm nay nhưng `0.30000000000000004` một ngày nào đó.
- **Chỗ có thể sai:** `TrimRight("30.0000", "0")` → `"30."` → `TrimSuffix` → `"30"` đúng; nhưng
  `TrimRight("0.0000", "0")` → `"0."` → `"0"` đúng; và `"100.0000"` → cắt `0` thừa → `"100."`? Không:
  `TrimRight` cắt tới dấu `.` rồi dừng vì `.` không phải `0` → `"100."` → `"100"`. Đúng, nhưng anh nên
  đưa `100` vào bảng để chắc.
- **Đề nghị: làm theo gợi ý này**, agent không viết code ở task này.

## Rủi ro
- Thông báo lỗi trong `config` thay đổi chữ — test của `config` chỉ so chứa `"quote"`, không đỏ.

## Acceptance (chép từ TASK.md)
- [ ] `String()` ra `8.81%`, `30%`, `0.0001%`, `-2.5%`, `108.81%`, `0%`
- [ ] Không float
- [ ] `go test ./...` xanh, không test nào khác đổi kỳ vọng
- [ ] Lỗi của `config.Parse` với cọc 150 in `deposit 150%`
