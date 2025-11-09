NextCRP: Thời điểm CRP tiếp theo sẽ trigger
  Ví dụ: now=12:00, interval=30s → NextCRP=12:00:30
  Dùng để: CanSendCRP() check (now >= NextCRP?)

NextActionAt: Thời điểm worker sẽ check reminder này lại
  = MIN(snooze_until, next_recurring, next_crp)
  Dùng để: Worker query (WHERE next_action_at <= NOW)


  
# 1 ONE TIME
  
    - NextActionAt > now
        - MaxCRP = > 0
        - MaxCRP = 0  
    - NextActionAt < now
        - MaxCRP = > 0
        - MaxCRP = 0  
    - chưa có âm lịch nha
    - user bấm hoàn thành giữa chừng
    {
        "title": "One-Time Now - khong CRP",
        "description": "Send immediately",
        "type": "one_time",
        "calendar_type": "solar",
        "max_crp": 0,
        "crp_interval_sec": 40,
        "status": "active"
    }

 #2 LẶP FRP  
 
