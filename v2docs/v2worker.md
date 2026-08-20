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

## 4. Lặp lại - không có CRP - không có crp_until_complete
     - điều kiện lấy từ db: 
        - type = recurring
        - status = active
        - snooze_until <= now()
	
        - max_crp = 0
        - repeat_strategy = none 
        - next_recurring <= now()
    - sau khi gửi notifi:
        - set next_recurring = calculate_next_recurringTH4NoCrpNoUT() 

## 5.1 Lặp lại - có CRP - không có crp_until_complete - next_recurring < now()
     - điều kiện lấy từ db: 
        - type = recurring
        - status = active
        - snooze_until <= now()

        - max_crp > 0
        - repeat_strategy = none 
        - next_recurring < now() 
    - sau khi gửi notifi:
        - set next_recurring = calculate_next_recurringTH4YesCrpNoUT() 
        - set crp_count = 0
        - set next_crp = now() + crp_interval_sec
    
 ## 5.2 Lặp lại - có CRP - không có crp_until_complete - next_recurring > now()
    - điều kiện lấy từ db: 
        - type = recurring
        - status = active
        - snooze_until <= now()

        - max_crp > 0
        - repeat_strategy = none 
        - next_recurring > now()
        - crp_count < max_crp
        - next_crp <= now()



## 6.1 Lặp lại - không có CRP - có crp_until_complete - lần đầu 
     - điều kiện lấy từ db: 
        - type = recurring
        - status = active
        - snooze_until <= now()

	    - maxt_crp=0
        - repeat_strategy = crp_until_complete  
        - next_recurring <= now()
        - isSendedOneTime = false 
    - sau khi gửi notifi:
        - set isSendedOneTime = true
        - last_sent_at = now()

  ## 6.2 Lặp lại - không có CRP - có crp_until_complete - đã complete
     - điều kiện lấy từ db: 
        - type = recurring
        - status = active
        - snooze_until <= now()

	    - maxt_crp=0
        - repeat_strategy = crp_until_complete  
        - next_recurring <= now()
        - isSendedOneTime = true
        - last_completed_at is valid 
        - last_completed_at > last_sent_at 
    - sau khi gửi notifi:
        - set last_sent_at = now()
        -       


## 7. Lặp lại - có CRP - có crp_until_complete
    - điều kiện lấy từ db: 
        - type = recurring
        - status = active
        - snooze_until <= now()
        - max_crp > 0
        - repeat_strategy = crp_until_complete  

        - next_recurring <= now()
        <!-- - isSendedOneTime = true
        - last_completed_at is valid 
        - last_completed_at > last_sent_at  -->


	-
    - sau khi gửi notifi:
        


        

#