# Hướng Dẫn Thêm Field Mới Vào Model Trong RemiAq

Dưới đây là hướng dẫn chi tiết để thêm một field mới vào model (ví dụ: thêm field `tag` hoặc `origin_time` vào model `Reminder`). Quy trình này đảm bảo tính nhất quán và tránh lỗi như field không xuất hiện trong API response hoặc không được lưu vào DB.

## Bước 1: Cập Nhật Model
- Mở file: `internal/models/reminder.go` (hoặc model tương ứng).
- Thêm field mới vào struct với tags phù hợp (json và db).
- Ví dụ:
  ```go
  type Reminder struct {
      // Các field cũ...
      Tag        string    `json:"tag" db:"tag"`
      OriginTime time.Time `json:"origin_time" db:"origin_time"` // Chú ý kiểu dữ liệu
      // Các field khác...
  }
  ```
- Thêm validation nếu cần trong method `Validate()`.

## Bước 2: Cập Nhật Schema Database
- Mở file migration (ví dụ `migrations/1671631110_init_schema.go`) hoặc tạo migration mới.
- Thêm cột vào definition của collection.
- **Lưu ý kiểu dữ liệu**: `Text`, `Int`, `Bool`, `Date`, `Json`, v.v.

## Bước 3: Cập Nhật Repository (QUAN TRỌNG)
Đây là bước dễ bị thiếu sót nhất. Mở file: `internal/repository/pocketbase/reminder_orm_repo.go`.

### 3.1. Cập nhật Conversion Functions
- **`recordToReminder`**: Đọc từ Record ra Model.
  - String: `record.GetString("field")`
  - Time: `record.GetDateTime("field").Time()`
  - Int: `record.GetInt("field")`
- **`reminderToRecord`**: Ghi từ Model vào Record.
  - String/Int: `record.Set("field", val)`
  - Time: Cần format chuẩn PocketBase (RFC3339Nano) hoặc gán `nil` nếu zero.
    ```go
    if !reminder.OriginTime.IsZero() {
        record.Set("origin_time", reminder.OriginTime.Format(time.RFC3339Nano))
    } else {
        record.Set("origin_time", nil)
    }
    ```

### 3.2. Cập nhật Raw SQL Queries (Dễ Quên!)
Các hàm như `GetDueReminders`, `GetByUserID` thường dùng câu lệnh SQL trực tiếp (`db.Select(...)`) và map vào một **struct nội bộ** (thường tên là `ReminderRecord`).
- **Phải thêm field vào struct nội bộ này**.
- **Phải thêm field vào danh sách cột trong câu lệnh `Select`** (nếu không select `*`).
- **Phải cập nhật logic mapping** từ `ReminderRecord` sang `models.Reminder` ở cuối hàm.
  - Với field `time.Time`, dùng hàm helper `parseTimeDB()`.

## Bước 4: Cập Nhật Services Và Handlers
- **Handler (`internal/handlers/reminder_handler.go`)**:
  - Cập nhật hàm validate (ví dụ `validateReminderForCreate`) nếu field là bắt buộc.
  - Kiểm tra xem field có cần xử lý đặc biệt khi nhận từ request body không.
- **Service (`internal/services/reminder_service.go`)**:
  - Nếu field ảnh hưởng đến logic (ví dụ tính toán thời gian), cập nhật các hàm liên quan (như `CalculateNext...`).

## Checklist Kiểm Tra
- [ ] Đã thêm vào Model struct?
- [ ] Đã thêm vào DB Schema?
- [ ] Đã cập nhật `recordToReminder`?
- [ ] Đã cập nhật `reminderToRecord`?
- [ ] **Đã cập nhật struct nội bộ trong `GetDueReminders`?**
- [ ] **Đã cập nhật struct nội bộ trong `GetByUserID`?**
- [ ] Đã cập nhật Validate function?

Hướng dẫn này được cập nhật dựa trên kinh nghiệm fix bug field `origin_time` không được lưu.