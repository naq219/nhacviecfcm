## cần phải thêm trường, xoá và thay đổi 1 số :
- next_mercurring: thời gian tiếp theo khi lặp lại theo recurrence_pattern.
- next_crp: thay cho next_trigger_at (xoá bỏ) 
- max_retries đổi thành max_crp 
- retry_count đổi thành crp_count 
- retry_interval_sec đổi thành crp_interval_sec
- retry_until_complete đổi thành crp_until_complete
 
## TRƯỜNG HỢP 1: type = none (one-time reminder)
- 1.1 max_crp = 0 : chỉ notifi 1 lần khi next_trigger_at <= now 
- 1.2 max_crp > 0 : notifi cho đến khi crp_count >= max_crp hoặc user complete . 

## TRƯỜNG HỢP 2: type = recurring (định kỳ)
- crp_strategy chỉ áp dụng với trường hợp 2 type = recurring
  - crp_strategy = none : next_mercurring = next_mercurring + recurrence_pattern  , không quan tâm khi nào bấm complete. Nhưng nếu next_mercurring + recurrence_pattern < now thì phải làm sao để tính next_mercurring tiếp theo?  next_mercurring phải luôn là bôi số của recurrence_pattern và lớn hơn now 
  - crp_strategy = crp_until_complete : next_mercurring =  last_completed_at + recurrence_pattern.

- 2.1 max_crp = 0 :
  
    
- 2.2 max_crp > 0 : 
  - lặp notifi cho đến khi crpe_count >= max_crp hoặc user complete hoặc next_mercurring + recurrence_pattern < now.

last_sent_at là chung cho crp và frp

mỗi lần muốn crp thì kiểm tra 

nếu type =none và  crp_count < max_crp và now - last_sent_at > crp_interval_sec   là có thể gửi crp

nếu type =recurring : vì frp sẽ check trước, frp đến hạn thì luôn gửi trước và reset crp 

  `last_completed_at` TEXT DEFAULT '' NOT NULL, 
  `last_sent_at` TEXT DEFAULT '' NOT NULL, 
  `max_retries` NUMERIC DEFAULT 0 NOT NULL, 
  `next_trigger_at` TEXT DEFAULT '' NOT NULL, 
  `recurrence_pattern` JSON DEFAULT NULL, 
  `repeat_strategy` TEXT DEFAULT '' NOT NULL, 
  `retry_count` NUMERIC DEFAULT 0 NOT NULL, 
  `retry_interval_sec` NUMERIC DEFAULT 0 NOT NULL, 
  `snooze_until` TEXT DEFAULT '' NOT NULL, 
  `status` TEXT DEFAULT '' NOT NULL, 
  `title` TEXT DEFAULT '' NOT NULL, 
  `trigger_time_of_day` TEXT DEFAULT '' NOT NULL, 


if type == "one_time" {
    status = "completed"
} else {
    if repeat_strategy == "crp_until_complete" {
        last_completed_at = now
        next_recurring = last_completed_at + pattern
    }
    else {
        // thì làm sao 
    }
}




--------------------------------------
CƠ CHẾ WORKER MỚI (Ngôn ngữ tự nhiên)

KHÁI NIỆM CƠ BẢN:

FRP (Father Recurrence Pattern) - Vòng lặp cha:

Chỉ có khi type = "recurring"
Lặp theo lịch: ngày/tuần/tháng/năm (dương/âm lịch)
Thời điểm kích hoạt: next_recurring
Chiến lược: repeat_strategy = "none" (theo lịch) hoặc "crp_until_complete" (chờ hoàn thành)


CRP (Child Repeat Pattern) - Vòng lặp con:

Có cho cả one_time và recurring
Lặp theo giây: crp_interval_sec (có thể rất lớn, vd: 15 ngày = 1,296,000 giây)
Giới hạn: max_crp (0 nghĩa là chỉ gửi 1 lần)
Đếm: crp_count (số lần đã gửi notification)


Các mốc thời gian:

last_sent_at: Lần cuối gửi notification (dùng chung cho CRP và FRP)
last_crp_completed_at: User đánh dấu hoàn thành lần notification hiện tại
last_completed_at: Dùng để tính chu kỳ FRP tiếp theo (chỉ khi repeat_strategy = "crp_until_complete")




QUY TRÌNH WORKER:
Bước 1: Lấy danh sách reminders

Lấy tất cả reminders có status = 'active'
Bỏ qua những cái đang bị snooze (snooze_until > now)

Bước 2: Xử lý từng reminder
2.1. Nếu type = "recurring" → Kiểm tra FRP trước:

Điều kiện: now >= next_recurring → Chu kỳ FRP đến hạn
Hành động:

Gửi notification ngay lập tức
Cập nhật last_sent_at = now
Reset crp_count = 0 (bắt đầu đếm CRP mới)
Tính next_recurring tiếp theo:

Nếu repeat_strategy = "none": Tính theo pattern (ví dụ: mỗi tháng)
Nếu repeat_strategy = "crp_until_complete": Chờ user complete, không tính


Cập nhật next_crp = next_recurring
Dừng xử lý, chờ chu kỳ worker tiếp theo



2.2. Kiểm tra CRP (áp dụng cho cả one_time và recurring):

Điều kiện gửi CRP:

crp_count < max_crp (chưa đạt giới hạn)
VÀ (last_sent_at chưa có HOẶC now - last_sent_at > crp_interval_sec)


Hành động:

Gửi notification
Cập nhật last_sent_at = now
Tăng crp_count++
Nếu type = "one_time" VÀ crp_count >= max_crp:

Đánh dấu status = "completed" (kết thúc reminder)






KHI USER BÁO COMPLETED:
Nếu type = "one_time":

Đánh dấu status = "completed" → Kết thúc reminder

Nếu type = "recurring":

Cập nhật last_crp_completed_at = now
Reset crp_count = 0 (cho phép nhận CRP lại trong chu kỳ hiện tại)
Nếu repeat_strategy = "crp_until_complete":

Cập nhật last_completed_at = now
Tính next_recurring = last_completed_at + pattern
Cập nhật next_crp = next_recurring


Nếu repeat_strategy = "none":

Không làm gì với next_recurring (vì nó chạy theo lịch cố định)




HÀM TÍNH NEXT_RECURRING:
Tính thời điểm FRP tiếp theo dựa trên recurrence_pattern:

Lấy next_recurring hiện tại
Cộng thêm pattern (ngày/tuần/tháng dương/âm lịch)
Nếu kết quả <= now: Tiếp tục cộng cho đến khi tìm được bội số đầu tiên > now
Trả về thời điểm tìm được

Ví dụ:

Pattern: Mỗi tháng
next_recurring cũ: 01/01/2025
Hôm nay: 15/03/2025
Kết quả: 01/04/2025 (bội số gần nhất lớn hơn hôm nay)


ƯU ĐIỂM CỦA CƠ CHẾ NÀY:

Đơn giản: Không cần field next_crp trong query, chỉ kiểm tra runtime
Chính xác: FRP luôn được ưu tiên, CRP tự động reset khi sang chu kỳ mới
Linh hoạt: crp_count được lưu DB, không sợ mất data khi restart
Rõ ràng: Phân biệt rõ "complete CRP" vs "complete FRP"