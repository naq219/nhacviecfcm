reminders có field mới: isSendedOneTime: boolean 
## 1. Một lần - không CRP
    - điều kiện lấy từ db: 
        - type = one_time
        - max_crp = 0
        - status = active
        - next_action_at <= now()
        - isSendedOneTime = false
        - snooze_until <= now()
    
    - sau khi gửi notifi:
        - set isSendedOneTime = true
        - set status = completed

## 2 Một lần - có CRP - isSendedOneTime = false
     - điều kiện lấy từ db: 
        - type = one_time
        - max_crp > 0
        - status = active
        - next_action_at <= now()
        - isSendedOneTime = false
        - snooze_until <= now()
    - sau khi gửi notifi:
        - set isSendedOneTime = true
        - set next_crp  = now() + crp_interval_sec
## 3 Một lần - có CRP - isSendedOneTime = true
     - điều kiện lấy từ db: 
        - type = one_time
        - max_crp > 0
        - crp_count < max_crp
        - status = active
        - next_crp <= now()
        - isSendedOneTime = true
        - snooze_until <= now()
    - sau khi gửi notifi:
        - set crp_count = crp_count + 1
        - set next_crp = now() + crp_interval_sec
        - if crp_count >= max_crp:
            - set status = completed

## 3. Lặp lại - không CRP


## 4. Lặp lại - có CRP