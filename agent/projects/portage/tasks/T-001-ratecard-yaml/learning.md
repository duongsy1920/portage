Ngôn ngữ:  `ratecard_test.go:95` — bảng test viết hàng theo tên trường; thêm trường mới không vỡ 10 hàng cũ (đã vỡ một lần khi viết theo vị trí)
Ngôn ngữ:  `ratecard.go:101` — `yaml.v3` `KnownFields(true)`: khoá lạ là lỗi, không rơi về zero value; và scalar `8.81` không ngoặc kép vẫn vào `string` (`ratecard_test.go:138`)
Ngôn ngữ:  `wire.go:195` — nạp file trước `Connect`: thứ rẻ và hay sai kiểm trước, lỗi mang tên file chứ không phải lỗi mạng
Domain:    `ratecard.go:131` — adapter chỉ thêm **tên trường** vào lỗi; luật vẫn ở `pricing.New*` (quy ước 8) — Symfony: TreeBuilder đặt tên node, validator của entity giữ luật
Domain:    `wire.go:284` — cấu hình (đọc từ file, đổi theo môi trường) ≠ value object (bất biến, chụp vào Quote `policy.go:63`); fixture là dữ liệu lập trình viên nên vẫn `Must*`
Domain:    `config/ratecard.yaml` — tỷ giá không vào file: thứ đổi theo ngày có endpoint riêng là trạng thái, không phải cấu hình
Kỹ thuật:  `wire_test.go:423` — hai nguồn sự thật chỉ chấp nhận được khi có một test cột chúng lại
Kỹ thuật:  `cmd/api/main.go:41` — mặc định cwd-relative hợp lệ khi mọi lệnh đã giả định gốc repo; thông báo lỗi phải nói đúng cờ
Agent:     gate đỏ hai lần đều ở test do agent viết; `go build` không biên dịch `_test.go` — kiểm biên dịch phải bằng `go vet` trên package
Agent:     script ghi kết quả phải đọc exit code TRƯỚC khi ghi chữ "passed" (bài học 2 của CLAUDE.md, tái phạm một lần ở verify lần 2)
Sai ở đâu: verify lần 2 ghi "passed" trước khi nhìn exit code — đã sửa record; implement lần 1 kiểm biên dịch bằng `go build`
Nợ mới:    `shared.Rate.String()` in `150.0000%` — định dạng phần trăm cho người đọc · yaml.v3 decode scalar vào string thế nào · khi nào dùng keyed fields trong struct literal
