# analysis — T-003 rate-string (chế độ học)

## Đụng vào cái gì
- **aggregate / module:** value object `shared.Rate` (`internal/domain/shared/rate.go`), chỉ method
  `String()` (`:65–72`). `PPM()` (`:56`) và `Plus1()` (`:61`) không đổi — giá trị không đổi, chỉ lời.
- **invariant:** không có invariant nghiệp vụ nào ở đây; bất biến kỹ thuật là **không float** (quy ước 2)
  và **zero value in được** (quy ước 9): `Rate{}` → `0%`.
- **context khác bị chạm:** không. Đã kiểm ai dùng chuỗi này: chỉ thông báo lỗi (`pricing/policy.go`,
  `lane.go`, `shared/rate.go`, 4 chỗ `%s`) và test `rate_test.go:16–20`. Không view HTTP, không cột DB
  (`pricing_repos.go:57` ghi `PPM()`), không event.

## Luật của project áp vào đây
- Quy ước 2: giữ số nguyên. Cách gợi ý: in như cũ rồi cắt `0` thừa và dấu `.` thừa ở phần thập phân.
- Quy ước 9: thử `Rate{}.String()` trong bảng.

## Điều ngoài luật, hoặc quyết định đã đóng bị chạm
- Không có.

## Câu trả lời của người
- Không hỏi. Người chọn task khác thay cho task này thì ghi ở `## Corrections` của TASK.md.

## Gợi ý kiến thức riêng của project
- `Rate.String()` là "lời", `PPM()` là "sự thật"; chỗ nào cần so sánh trong test thì so `PPM()`.
