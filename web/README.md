# RemiAq Vue.js Client

Website client cho hệ thống quản lý nhắc nhở RemiAq, được xây dựng với Vue.js 3 và các thư viện UI hiện đại.

## Tính năng

### ✅ Đã hoàn thành
- **Đăng nhập/Đăng ký**: Giao diện đẹp với form chuyển đổi mượt mà
- **Dashboard thống kê**: Hiển thị số liệu tổng quan về nhắc nhở
- **Danh sách nhắc nhở**: Giao diện card-based hiện đại với trạng thái màu sắc
- **Tạo nhắc nhở mới**: Form với các tùy chọn cơ bản và nâng cao
- **Chức năng CRUD**: Thêm, sửa, xóa, hoàn thành, trì hoãn nhắc nhở
- **Tùy chọn nâng cao**: Form expandable với ngày đến hạn, âm lịch, cron expression
- **Thiết kế responsive**: Hoạt động tốt trên desktop và mobile
- **Giao diện hiện đại**: Tông màu sáng, thiết kế mềm mại với glass morphism

### 🎨 Thiết kế
- **Màu sắc**: Tông màu tím-indigo pastel kết hợp gradient
- **Hiệu ứng**: Hover animations, smooth transitions, glass morphism
- **Typography**: Font hiện đại, dễ đọc với hierarchy rõ ràng
- **Icons**: Font Awesome 6.4.0 với biểu tượng trực quan

## Cấu trúc file

```
web/
├── index.html      # File HTML chính với Vue.js app
├── styles.css      # CSS tùy chỉnh và variables
└── README.md       # Hướng dẫn này
```

## Cách sử dụng

### 1. Mở website
- Mở file `index.html` trong trình duyệt
- Hoặc serve qua local server (đề xuất)

### 2. Đăng nhập thử nghiệm
- Hiện tại đang ở chế độ demo với data mẫu
- Form đăng nhập/đăng ký đã có sẵn nhưng chưa kết nối API

### 3. Quản lý nhắc nhở
- **Thêm mới**: Click nút "Thêm nhắc nhở" và điền form
- **Tùy chọn nâng cao**: Click "Tùy chọn nâng cao" để mở rộng form
- **Chỉnh sửa**: Click icon bút chì trên card nhắc nhở
- **Hoàn thành**: Click dấu check để đánh dấu hoàn thành
- **Trì hoãn**: Click đồng hồ để trì hoãn 1 giờ
- **Xóa**: Click thùng rác và xác nhận

## Tích hợp API

Website đã chuẩn bị sẵn các hàm để tích hợp với backend:

### Authentication API
```javascript
// Đăng nhập
async login() {
    const response = await fetch('/api/collections/musers/auth-with-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            identity: this.loginForm.email,
            password: this.loginForm.password
        })
    });
    const data = await response.json();
    localStorage.setItem('token', data.token);
    this.user = data.record;
}
```

### Reminder API
```javascript
// Lấy danh sách nhắc nhở
async loadReminders() {
    const response = await fetch('/api/reminders', {
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
    });
    this.reminders = await response.json();
}

// Tạo nhắc nhở mới
async addReminder() {
    const response = await fetch('/api/reminders', {
        method: 'POST',
        headers: { 
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify(this.newReminder)
    });
    const reminder = await response.json();
    this.reminders.unshift(reminder);
}
```

## Tùy chỉnh

### Thay đổi màu sắc
Chỉnh sửa CSS variables trong `styles.css`:
```css
:root {
    --primary-color: #6366f1;     /* Màu chính */
    --secondary-color: #f59e0b;   /* Màu phụ */
    --background: #f8fafc;       /* Nền */
}
```

### Thêm field mới
1. Thêm field vào `newReminder` object trong Vue data
2. Thêm input tương ứng trong form
3. Cập nhật API integration

### Thay đổi ngôn ngữ
Tất cả text đều nằm trong HTML, có thể dễ dàng dịch sang ngôn ngữ khác.

## Browser Support
- Chrome 80+
- Firefox 75+
- Safari 13+
- Edge 80+

## Performance
- Vue.js 3 với Composition API
- Tailwind CSS utility-first
- Minimal custom CSS
- No build step required

## Next Steps
1. Kết nối với API backend thực tế
2. Thêm validation form
3. Implement real-time updates
4. Add offline support
5. Mobile app wrapper