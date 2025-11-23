# Hướng Dẫn Thêm Field Mới Vào Model Trong RemiAq

Dưới đây là hướng dẫn chi tiết để thêm một field mới vào model (ví dụ: thêm field `tag` vào model `Reminder`). Quy trình này đảm bảo tính nhất quán và tránh lỗi như field không xuất hiện trong API response.

## Bước 1: Cập Nhật Model
- Mở file: `internal/models/reminder.go` (hoặc model tương ứng).
- Thêm field mới vào struct với tags phù hợp (json và db).
- Ví dụ:
  ```go
  type Reminder struct {
      // Các field cũ...
      Tag string `json:"tag" db:"tag"`
      // Các field khác...
  }
  ```
- Thêm validation nếu cần trong method `Validate()`.

## Bước 2: Cập Nhật Schema Database (Nếu Cần)
- Nếu field mới yêu cầu thay đổi schema, tạo migration mới trong folder `migrations/`.
- Ví dụ: Thêm cột `tag` vào table `reminders`.
- Chạy migration để áp dụng thay đổi.

## Bước 3: Cập Nhật Repository
- Mở file: `internal/repository/pocketbase/reminder_orm_repo.go` (hoặc repo tương ứng).
- Thêm field vào tất cả các struct nội bộ (như `ReminderRecord` trong `GetByUserID`, `GetDueReminders`, v.v.).
- Cập nhật functions conversion:
  - Trong `recordToReminder`: Thêm mapping `Tag: record.GetString("tag"),`.
  - Trong `reminderToRecord`: Thêm `record.Set("tag", reminder.Tag)`.
- Kiểm tra tất cả methods như `GetByID`, `GetByUserID`, `Create`, `Update` để đảm bảo field được xử lý.

## Bước 4: Cập Nhật Services Và Handlers
- Kiểm tra `internal/services/reminder_service.go`: Đảm bảo field được truyền qua khi gọi repo.
- Kiểm tra `internal/handlers/reminder_handler.go`: Field sẽ tự động xuất hiện trong JSON response nếu đã mapping đúng.

## Bước 5: Viết Và Chạy Tests
- Thêm tests cho field mới trong file `_test.go` tương ứng (repo, service).
- Chạy `go test -v ./...` để xác nhận không có lỗi.

## Bước 6: Kiểm Tra API
- Chạy server: `go run ./cmd/server serve`.
- Test API (ví dụ: `GET /api/reminders/mine`) để xác nhận field mới xuất hiện trong response.

## Lưu Ý
- Luôn kiểm tra tất cả functions conversion để tránh field bị thiếu.
- Nếu field là optional, sử dụng pointer (`*string`).
- Commit thay đổi và test đầy đủ trước khi deploy.

Hướng dẫn này dựa trên kinh nghiệm fix bug field `tag` bị rỗng.