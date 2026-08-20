# Viết tắt:
- Nhắc nhở: nn . Nhắc nhở lặp: nnl. Nhắc Nhở 1 lần: nn1

# Danh sách lời nhắc nhỏ
- [ ] Hiển thị danh sách lời nhắc 
- mỗi row: 
    - hiển thị nội dung nhắc, thời gian nhắc, trạng thái (chưa hoàn thành/hoàn thành), kiểu lặp (1 lần/lặp x time)
- có bộ lọc 7 ngày, 1 tháng, toàn bộ cho lời nhắc sắp tới (chưa hoàn thành), 
- vuốt sang để hiển thị nút chức năng 
# Tạo nhắc nhở
- Phải có thời gian bắt đầu gồm ngày tháng năm , giờ phút (Khi nào bắt đầu nhắc nhở , kể cả nnl)
- Với nnl phải có thời gian lặp (Mỗi bao nhiêu phút, giờ, ngày, tuần, tháng)
- Tuỳ chọn lặp có: Mỗi ngày, Mỗi tuần, Mỗi tháng, Mỗi năm, Tuỳ chọn
    - Nếu tuỳ chọn: nhập vào số + ký tự, ví dụ: 1p, 1g, 3n, 16t (1 phút, 1 giờ, 3 ngày, 1 tuần, 6 tháng)
- Mức độ quan trọng: cao, bình thường. có 1 checkbox chọn mức độ quan trọng.không chọn là thông thường
# Yêu cầu chung
- Không cần phải hiển thị chữ quá rõ ràng nếu người dùng vẫn hiểu, thay thế bằng icon nếu hợp lý cho gọn
- Những tính năng thường dùng phải ít thao tác nhất
- Giao diện hiện đại, không màu mè, không đơn giản
- các thành phần sắp xếp gọn gàng, để tránh hạn chế scroll .
- Giao diện nên phân biệt nhóm nếu nhiều thao tác, thông tin. ví dụ tuỳ chọn lặp lại nếu chưa hoàn thành thì có (text giới thiệu, edittext số lần nhắc, edittext nhập thời gian lặp lại) thì phải làm sao để người dùng biết là  nhưng field đó là của 1 chức năng. 

dùng floating hint cho edittext 
