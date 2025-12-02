1. lặp theo X ngày : cố định giờ . tương đương x*24*60*60 giây.
1.1 nếu noUT  next_recurring = next_recurring + x*24*60*60 giây
1.2 nếu YesUT : không có trường hợp này vì next_recurring luôn là quá khứ 
* nếu user complete:
   -- set crp_count=max_crp 

    ** và YesUT thì next_recurring = now + x*24*60*60 giây
* lưu ý: nếu new next_recurring có hour:minute != old next_recurring thì set lại cho đúng 


{
    "type": "daily",
   
    "interval": 1,
    "origin_time": "2025-12-01T09:47:44"

}




// lặp theo x tháng : cố định ngày và giờ . 

2. lặp hàng tháng dương lịch: cố định ngày và giờ
{
    "type": "monthly",
    "origin_time": "2025-12-01T09:47:44"
    "interval": 1,
    "calendar_type": "solar",

}






3. lặp hàng tháng âm lịch: cố định ngày và giờ
{
    "type": "monthly",
    "origin_time": "2025-12-01T09:47:44",
    "calendar_type": "lunar",
    "day_of_month": 1,
}


<!-- {
    "type": "monthly",
    "trigger_time_of_day": "09:00",
    "day_of_month": 1,
    "calendar_type": "lunar",
} -->



lặp cuối tháng âm lịch: không cố định ngày (chọn cuối tháng ), cố định giờ 
{
    "type": "monthly",
    "trigger_time_of_day": "09:00",
    "day_of_month": 1,
    "calendar_type": "lunar",
    "last_day_of_month": true,
}

lặp cuối tháng dương lịch: không cố định ngày (chọn cuối tháng ), cố định giờ 
{
    "type": "monthly",
    "trigger_time_of_day": "09:00",
    "day_of_month": 1,
    "calendar_type": "solar",
    "last_day_of_month": true,
}

lặp hàng năm dương lịch: cố định ngày và giờ
{
    "type": "yearly",
    "trigger_time_of_day": "09:00",
    "day_of_month": 1,
    "calendar_type": "solar",
}

lặp hàng năm âm lịch: cố định ngày và giờ
{
    "type": "yearly",
    "trigger_time_of_day": "09:00",
    "day_of_month": 1,
    "calendar_type": "lunar",
}

lặp ngày trong tuần 


