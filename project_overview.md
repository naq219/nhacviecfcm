# Báo cáo tổng quan dự án RemiAq

## 1. Cấu trúc thư mục

Đây là cấu trúc thư mục chính của dự án, tập trung vào các phần logic nghiệp vụ.

```
internal/
├── db/
│   ├── db_utils.go
│   ├── mapper.go
│   └── transaction.go
├── handlers/
│   ├── query_handler.go
│   ├── reminder_handler.go
│   └── system_status_handler.go
├── middleware/
│   ├── cors.go
│   └── validation.go
├── models/
│   └── reminder.go
├── repository/
│   ├── interface.go
│   └── pocketbase/
│       ├── query_repo.go
│       ├── reminder_repo.go
│       ├── system_status_repo.go
│       └── user_repo.go
├── services/
│   ├── fcm_service.go
│   ├── lunar_calendar.go
│   ├── reminder_service.go
│   └── schedule_calculator.go
├── utils/
│   └── response.go
└── worker/
    └── worker.go
```

## 2. Chi tiết các thành phần

### `internal/db`

Cung cấp các tiện ích để tương tác với cơ sở dữ liệu PocketBase, bao gồm các hàm generic để truy vấn và ánh xạ dữ liệu an toàn.

#### `db/db_utils.go`

- **Interface: `DBHelperInterface`**: Định nghĩa các phương thức cơ bản để tương tác với CSDL, cho phép mocking trong unit test.
- **Struct: `DBHelper`**: Triển khai `DBHelperInterface`, chứa instance của `pocketbase.PocketBase`.
- **Func: `NewDBHelper(...)`**: Tạo một `DBHelper` mới.
- **Func: `GetOne[T](...)`**: Lấy một bản ghi và tự động ánh xạ vào struct `T`.
- **Func: `GetAll[T](...)`**: Lấy danh sách bản ghi và ánh xạ vào slice `[]T`.
- **Func: `Exec(...)`**: Thực thi các câu lệnh `INSERT`, `UPDATE`, `DELETE`.
- **Func: `Count(...)`**: Đếm số lượng bản ghi.
- **Func: `Exists(...)`**: Kiểm tra sự tồn tại của bản ghi một cách hiệu quả.

#### `db/mapper.go`

- **Interface: `CustomMapper`**: Cho phép tùy chỉnh logic ánh xạ cho các trường dữ liệu đặc biệt.
- **Struct: `MapperConfig`**: Cấu hình cho quá trình ánh xạ (ví dụ: yêu cầu các trường bắt buộc).
- **Func: `MapNullStringMapToStruct[T](...)`**: Hàm generic chính để ánh xạ một `dbx.NullStringMap` (dữ liệu thô từ DB) sang một struct `T` dựa trên `db` tag. Hỗ trợ cache để tăng hiệu năng.

#### `db/transaction.go`

- **Func: `InTransaction(...)`**: Bọc một hoặc nhiều thao tác CSDL trong một transaction. Tự động `commit` nếu không có lỗi và `rollback` khi có lỗi.

### `internal/repository`

Lớp trừu tượng hóa việc truy cập dữ liệu. Định nghĩa các `interface` và các triển khai cụ thể sử dụng PocketBase.

#### `repository/interface.go`

- **Interface: `ReminderRepository`**: Định nghĩa các thao tác CRUD và truy vấn cho `Reminder`.
- **Interface: `UserRepository`**: Định nghĩa các thao tác cho `User`, bao gồm quản lý token FCM.
- **Interface: `SystemStatusRepository`**: Định nghĩa các thao tác để quản lý trạng thái của hệ thống (ví dụ: bật/tắt worker).
- **Interface: `QueryRepository`**: Định nghĩa các thao tác cho các truy vấn SQL thô.

#### `repository/pocketbase/reminder_repo.go`

- **Struct: `ReminderRepo`**: Triển khai `ReminderRepository` sử dụng `DBHelper`.
- **Func: `NewReminderRepo(...)`**: Tạo một `ReminderRepo` mới.
- **Func: `Create(...)`, `GetByID(...)`, `Update(...)`, `Delete(...)`**: Các hàm CRUD cơ bản.
- **Func: `GetDueReminders(...)`**: Lấy các nhắc nhở đã đến hạn để worker xử lý.

#### `repository/pocketbase/system_status_repo.go`

- **Struct: `SystemStatusRepo`**: Triển khai `SystemStatusRepository`.
- **Func: `NewSystemStatusRepo(...)`**: Tạo một `SystemStatusRepo` mới.
- **Func: `Get(...)`**: Lấy thông tin trạng thái hệ thống (là một singleton).
- **Func: `IsWorkerEnabled(...)`**: Kiểm tra xem worker có đang được cho phép chạy hay không.
- **Func: `EnableWorker(...)`, `DisableWorker(...)`**: Bật hoặc tắt worker.

### `internal/services`

Chứa logic nghiệp vụ của ứng dụng, điều phối hoạt động giữa các repository và các dịch vụ bên ngoài.

#### `services/reminder_service.go`

- **Struct: `ReminderService`**: Điều phối logic liên quan đến `Reminder`.
- **Func: `NewReminderService(...)`**: Tạo một `ReminderService` mới.
- **Func: `CreateReminder(...)`, `UpdateReminder(...)`**: Tạo/cập nhật nhắc nhở, tính toán thời gian kích hoạt tiếp theo.
- **Func: `ProcessDueReminders(...)`**: Logic cốt lõi được worker gọi. Lấy các nhắc nhở đến hạn, gửi thông báo qua FCM, và cập nhật lại lịch trình cho các nhắc nhở lặp lại.
- **Func: `CompleteReminder(...)`**: Xử lý khi người dùng hoàn thành một nhắc nhở.

#### `services/fcm_service.go`

- **Struct: `FCMService`**: Gửi thông báo đẩy qua Firebase Cloud Messaging.
- **Func: `NewFCMService(...)`**: Khởi tạo service với credentials.
- **Func: `SendNotification(...)`**: Gửi một thông báo đơn giản đến một thiết bị.

#### `services/schedule_calculator.go`

- **Struct: `ScheduleCalculator`**: Tính toán thời điểm kích hoạt tiếp theo cho các nhắc nhở.
- **Func: `NewScheduleCalculator(...)`**: Tạo một `ScheduleCalculator` mới.
- **Func: `CalculateNextTrigger(...)`**: Logic chính để tính toán, hỗ trợ cả lịch dương và lịch âm, các kiểu lặp lại (hàng ngày, hàng tuần, hàng tháng) và lặp lại dựa trên khoảng thời gian.

### `internal/handlers`

Lớp xử lý các yêu cầu HTTP, chuyển đổi dữ liệu từ request, gọi các service tương ứng và trả về response.

- **Struct: `ReminderHandler`**: Xử lý các endpoint liên quan đến `Reminder` (CRUD, snooze, complete).
- **Func: `NewReminderHandler(...)`**: Tạo một `ReminderHandler` mới.

### `internal/models`

Định nghĩa các cấu trúc dữ liệu chính của ứng dụng.

- **Struct: `Reminder`**: Đại diện cho một nhắc nhở, chứa tất cả thông tin về lịch trình, trạng thái, và nội dung.
- **Struct: `RecurrencePattern`**: Định nghĩa quy tắc lặp lại cho nhắc nhở.
- **Struct: `User`**: Đại diện cho người dùng, chứa token FCM.
- **Struct: `SystemStatus`**: Lưu trữ trạng thái toàn cục của hệ thống (singleton).
- **Constants**: Định nghĩa các giá trị hằng số (ví dụ: `ReminderStatusActive`, `CalendarTypeLunar`) để đảm bảo tính nhất quán.

### `internal/worker`

Thành phần chạy nền để xử lý các tác vụ định kỳ.

- **Struct: `Worker`**: Vòng lặp chính của worker, chạy theo một khoảng thời gian (`interval`) được cấu hình.
- **Func: `NewWorker(...)`**: Tạo một `Worker` mới.
- **Func: `Start(...)`**: Bắt đầu vòng lặp của worker trong một goroutine.
- **Func: `runOnce(...)`**: Logic thực thi trong mỗi chu kỳ: kiểm tra xem worker có được bật không, sau đó gọi `ReminderService.ProcessDueReminders()` để xử lý các nhắc nhở đến hạn. Tự động tắt worker nếu có lỗi hệ thống.

## 📌 Tổng kết

Dự án **RemiAq** đã có một cấu trúc code rõ ràng, tuân thủ nguyên tắc **Clean Architecture** với các lớp được tách biệt rõ ràng:

- **Database Layer (`internal/db`)**: Cung cấp các hàm generic để thao tác với cơ sở dữ liệu, hỗ trợ transaction, mapping dữ liệu, và xử lý lỗi.
- **Repository Layer (`internal/repository`)**: Định nghĩa interface và triển khai PocketBase cho việc truy xuất dữ liệu.
- **Service Layer (`internal/services`)**: Chứa logic nghiệp vụ chính như tính toán lịch trình, gửi thông báo FCM, và quản lý nhắc nhở.
- **Handler Layer (`internal/handlers`)**: Tiếp nhận và xử lý các yêu cầu HTTP từ client.
- **Model Layer (`internal/models`)**: Định nghĩa các cấu trúc dữ liệu cốt lõi của hệ thống.
- **Worker Layer (`internal/worker`)**: Thực thi các tác vụ nền theo định kỳ.

### Các tính năng nổi bật đã được triển khai:
- ✅ Hệ thống nhắc nhở linh hoạt với hỗ trợ lịch âm và lịch dương.
- ✅ Gửi thông báo qua Firebase Cloud Messaging (FCM).
- ✅ Tính toán thời gian kích hoạt tiếp theo cho nhắc nhở định kỳ.
- ✅ Worker tự động xử lý và gửi thông báo đúng hạn.
- ✅ API RESTful đầy đủ cho việc quản lý nhắc nhở.

## 🔄 Đề xuất cập nhật tài liệu

Để đảm bảo tài liệu luôn phản ánh đúng mã nguồn hiện tại, tôi đề xuất:

1. **Tự động hóa việc tạo tài liệu từ GoDoc:** Sử dụng công cụ như `godoc` hoặc `pkgsite` để tạo tài liệu trực tiếp từ các comment trong mã nguồn. Điều này giúp giảm thiểu công sức duy trì và đảm bảo tính nhất quán.

2. **Cập nhật `README.md`:** Bổ sung phần "Cấu trúc dự án" hoặc "Tổng quan kiến trúc" với nội dung tương tự như báo cáo này để người mới có thể nhanh chóng hiểu được hệ thống.

3. **Cập nhật `SRS.md` (nếu có):** Đảm bảo các chức năng đã triển khai được đánh dấu là hoàn thành và mô tả chi tiết cách thức hoạt động của các tính năng phức tạp như tính toán lịch âm hoặc gửi thông báo đa thiết bị.

4. **Tạo `CHANGELOG.md`:** Ghi lại lịch sử thay đổi của dự án theo từng phiên bản, giúp theo dõi tiến độ và các cải tiến đã thực hiện.

5. **Thêm diagram kiến trúc:** Sử dụng công cụ vẽ sơ đồ (ví dụ: draw.io, PlantUML) để tạo sơ đồ kiến trúc hệ thống, giúp trực quan hóa cách các thành phần tương tác với nhau.

Việc duy trì tài liệu cập nhật không chỉ giúp sếp và team hiểu rõ tiến độ mà còn là cơ sở quan trọng cho việc onboard thành viên mới và đảm bảo chất lượng dự án trong dài hạn.