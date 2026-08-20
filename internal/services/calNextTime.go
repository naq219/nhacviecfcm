package services

import (
	"errors"
	"fmt"
	"time"

	"remiaq/internal/models"
)

func Tinhtoan_NextRecurringV2_fromapi(reminder models.Reminder, now time.Time) (time.Time, error) {

	// user click complete đi chỗ khác chơi
	if reminder.From != "api_complete" {
		return time.Time{}, errors.New("không phải gọi từ api!!")
	}

	pattern := reminder.RecurrencePattern

	switch pattern.Type {
	case models.RecurrenceTypeDaily:
		return calcNextDailyTime(reminder, now, now)
	case models.RecurrenceTypeMonthly:
		return calcNextSolarMonthly(reminder, now, now)
	case models.RecurrenceTypeSolarLastDayOfMonth:
		return calcNextSolarLastDayOfMonth(now)
	case models.RecurrenceTypeIntervalSeconds:
		return calcNextIntervalSeconds(reminder, now, now)
	default:
		return time.Time{}, errors.New("unsupported recurrence type31235")
	}

}
func Tinhtoan_NextRecurringV2(reminder models.Reminder, now time.Time) (time.Time, error) {
	if reminder.RecurrencePattern == nil {

		return time.Time{}, errors.New("recurrence_pattern required for recurring reminder")
	}

	pattern := reminder.RecurrencePattern

	// user click complete đi chỗ khác chơi
	if reminder.From == "api_complete" {
		return time.Time{}, errors.New("gọi sai chỗ rồi")
	}

	// hàng x tháng âm lịch
	if reminder.CalendarType == models.CalendarTypeLunar {
		return calcNextLunarMonthly(reminder, now)
	}

	// chỉ trường hợp lặp không UT tại vì có UT thì không cần tính toán
	//

	// mỗi X seconds
	if pattern.Type == models.RecurrenceTypeIntervalSeconds {
		return calcNextIntervalSeconds(reminder, reminder.OriginTime, now)
	}

	// Existing logic (daily, weekly, monthly, lunar)
	switch pattern.Type {
	case models.RecurrenceTypeDaily:
		return calcNextDailyTime(reminder, reminder.OriginTime, now)
	case models.RecurrenceTypeWeekly:
		return calcNextWeekly(reminder, reminder.OriginTime, now)
	case models.RecurrenceTypeMonthly:
		return calcNextSolarMonthly(reminder, reminder.OriginTime, now)
	case models.RecurrenceTypeSolarLastDayOfMonth:
		return calcNextSolarLastDayOfMonth(now)
	default:
		return time.Time{}, errors.New("unsupported recurrence type")
	}
}

func calcNextIntervalSeconds(reminder models.Reminder, lastTime time.Time, now time.Time) (time.Time, error) {
	intervalSeconds := reminder.RecurrencePattern.IntervalSeconds
	if intervalSeconds <= 0 {
		return time.Time{}, fmt.Errorf("intervalSeconds must be > 0")
	}

	nextTime := lastTime.Add(time.Duration(intervalSeconds) * time.Second)

	// trường hợp hệ thống bị stop quá lâu, khi quay lại thì dù có cộng thêm vẫn chưa quá now
	for !nextTime.After(now) {
		nextTime = nextTime.Add(time.Duration(intervalSeconds) * time.Second)
	}

	return nextTime, nil
}
func calcNextSolarLastDayOfMonth(now time.Time) (time.Time, error) {
	location := now.Location()

	// ngày cuối tháng hiện tại
	lastDayThisMonth := time.Date(
		now.Year(),
		now.Month()+1,
		0,
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond(),
		location,
	)

	// nếu hôm nay chưa phải ngày cuối tháng → trả về cuối tháng này
	if now.Day() < lastDayThisMonth.Day() {
		return lastDayThisMonth, nil
	}

	// nếu hôm nay là ngày cuối tháng → trả về ngày cuối tháng sau
	lastDayNextMonth := time.Date(
		now.Year(),
		now.Month()+2, // tháng sau + 1
		0,
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond(),
		location,
	)

	return lastDayNextMonth, nil
}

func calcNextDailyTime(reminder models.Reminder, lastTime time.Time, now time.Time) (time.Time, error) {
	pattern := reminder.RecurrencePattern
	interval := pattern.Interval      // số ngày lặp
	originTime := reminder.OriginTime // thời gian bắt đầu
	if originTime.IsZero() {
		return time.Time{}, errors.New("originTime required for daily reminder")
	}
	originHour := originTime.Hour()
	originMinute := originTime.Minute()

	// tạo thời điểm lặp đầu tiên từ baseTime với giờ/phút chuẩn
	nextTime := time.Date(
		lastTime.Year(), lastTime.Month(), lastTime.Day(),
		originHour, originMinute, 0, 0,
		lastTime.Location(),
	)

	// cộng ít nhất interval ngày
	nextTime = nextTime.AddDate(0, 0, interval)

	// nếu vẫn chưa vượt qua `now` thì cộng tiếp theo interval
	for !nextTime.After(now) {
		nextTime = nextTime.AddDate(0, 0, interval)
	}

	return nextTime, nil

}

func calcNextSolarMonthly(reminder models.Reminder, lastTime time.Time, now time.Time) (time.Time, error) {
	pattern := reminder.RecurrencePattern
	interval := pattern.Interval      // số tháng lặp
	originTime := reminder.OriginTime // thời gian gốc (lấy ngày + giờ/phút)
	location := lastTime.Location()

	originDay := originTime.Day()
	originHour := originTime.Hour()
	originMinute := originTime.Minute()

	// 1) Bắt đầu từ lastTime (normalize về ngày 1 để tránh overflow)
	nextTime := time.Date(
		lastTime.Year(),
		lastTime.Month(),
		1, // START AT DAY 1
		originHour, originMinute, 0, 0,
		location,
	)

	// 2) Cộng interval tháng
	nextTime = nextTime.AddDate(0, interval, 0)

	// 3) Set ngày theo originDay, điều chỉnh nếu tháng không đủ ngày
	lastDay := time.Date(nextTime.Year(), nextTime.Month()+1, 0, 0, 0, 0, 0, location).Day()

	if originDay > lastDay {
		// Tháng này không có ngày originDay → dùng ngày cuối tháng
		nextTime = time.Date(
			nextTime.Year(),
			nextTime.Month(),
			lastDay,
			originHour, originMinute, 0, 0,
			location,
		)
	} else {
		// Tháng này có đủ ngày → set originDay
		nextTime = time.Date(
			nextTime.Year(),
			nextTime.Month(),
			originDay,
			originHour, originMinute, 0, 0,
			location,
		)
	}

	// 4) Nếu vẫn chưa vượt now → cộng thêm interval tháng cho tới khi > now
	for !nextTime.After(now) {
		// Normalize về ngày 1 trước khi add để tránh overflow
		tmpTime := time.Date(nextTime.Year(), nextTime.Month(), 1, originHour, originMinute, 0, 0, location)
		tmpTime = tmpTime.AddDate(0, interval, 0)

		lastDay = time.Date(tmpTime.Year(), tmpTime.Month()+1, 0, 0, 0, 0, 0, location).Day()

		if originDay > lastDay {
			nextTime = time.Date(tmpTime.Year(), tmpTime.Month(), lastDay, originHour, originMinute, 0, 0, location)
		} else {
			nextTime = time.Date(tmpTime.Year(), tmpTime.Month(), originDay, originHour, originMinute, 0, 0, location)
		}
	}

	return nextTime, nil
}

// calcNextWeekly tính thời điểm lặp tiếp theo cho weekly pattern
// Lấy weekday từ originTime.Weekday() thay vì pattern.DayOfWeek
func calcNextWeekly(reminder models.Reminder, lastTime time.Time, now time.Time) (time.Time, error) {
	pattern := reminder.RecurrencePattern
	originTime := reminder.OriginTime
	if originTime.IsZero() {
		return time.Time{}, errors.New("originTime required for weekly reminder")
	}

	// Lấy weekday từ originTime
	targetWeekday := int(originTime.Weekday()) // 0=Sunday, 1=Monday, ..., 6=Saturday
	originHour := originTime.Hour()
	originMinute := originTime.Minute()
	interval := pattern.Interval // số tuần lặp
	if interval <= 0 {
		interval = 1
	}

	// Bắt đầu từ lastTime, set giờ/phút theo origin
	nextTime := time.Date(
		lastTime.Year(), lastTime.Month(), lastTime.Day(),
		originHour, originMinute, 0, 0,
		lastTime.Location(),
	)

	// Tính số ngày cần thêm để đến targetWeekday
	currentWeekday := int(nextTime.Weekday())
	daysAhead := (targetWeekday - currentWeekday + 7) % 7

	if daysAhead == 0 {
		// Hôm nay đúng weekday, check xem đã qua now chưa
		if nextTime.After(now) {
			return nextTime, nil
		}
		// Chưa đến now, lấy tuần tiếp theo
		daysAhead = 7 * interval
	}

	nextTime = nextTime.AddDate(0, 0, daysAhead)

	// Nếu vẫn chưa qua now, cộng thêm interval weeks
	for !nextTime.After(now) {
		nextTime = nextTime.AddDate(0, 0, 7*interval)
	}

	return nextTime, nil
}

// calcNextLunarMonthly tính thời điểm lặp tiếp theo cho lunar monthly
func calcNextLunarMonthly(reminder models.Reminder, now time.Time) (time.Time, error) {
	originTime := reminder.OriginTime
	if originTime.IsZero() {
		return time.Time{}, errors.New("originTime required for lunar monthly")
	}

	// Lấy ngày âm lịch từ originTime
	vnOrigin := originTime.In(time.FixedZone("VN", 7*3600))
	key := vnOrigin.Format("2006-01-02")
	lunarOrigin, ok := SolarToLunarMap[key]
	if !ok {
		return time.Time{}, errors.New("origin date not in lunar map")
	}

	lunarDay := lunarOrigin.Day
	originHour := originTime.Hour()
	originMinute := originTime.Minute()

	// Tìm ngày âm lịch tiếp theo
	nextSolar, err := FindNextLunarMonthly(now, lunarDay)
	if err != nil {
		return time.Time{}, err
	}

	// Set giờ/phút theo originTime
	nextTime := time.Date(
		nextSolar.Year(), nextSolar.Month(), nextSolar.Day(),
		originHour, originMinute, 0, 0,
		now.Location(),
	)

	return nextTime, nil
}
