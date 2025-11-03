/*
 * Copyright (c) 2006 Ho Ngoc Duc. All Rights Reserved.
 * Astronomical algorithms from the book "Astronomical Algorithms" by Jean Meeus, 1998
 *
 * Permission to use, copy, modify, and redistribute this software and its
 * documentation for personal, non-commercial use is hereby granted provided that
 * this copyright notice and appropriate documentation appears in all copies.
 *
 * ---
 * Ported to Go from the original JavaScript.
 * Wrapped with LunarCalendar struct for API compatibility.
 */

// Package services cung cấp các hàm để chuyển đổi giữa lịch Dương và lịch Âm
// dựa trên thuật toán của Jean Meeus.
package services

import (
	"math"
	"time"
)

const pi = math.Pi

// LunarDate represents a date in lunar calendar
type LunarDate struct {
	Year     int  `json:"year"`
	Month    int  `json:"month"`
	Day      int  `json:"day"`
	IsLeap   bool `json:"is_leap"`   // Tháng nhuận
	LeapYear bool `json:"leap_year"` // Năm nhuận
}

// LunarCalendar handles lunar calendar conversions using Vietnamese lunar calendar algorithm
type LunarCalendar struct {
	timeZone float64 // GMT+7 for Vietnam
}

// NewLunarCalendar creates a new lunar calendar service for Vietnam timezone
func NewLunarCalendar() *LunarCalendar {
	return &LunarCalendar{
		timeZone: 7.0, // GMT+7 for Vietnam
	}
}

// SolarToLunar converts solar date to lunar date using Vietnamese lunar calendar
func (lc *LunarCalendar) SolarToLunar(solar time.Time) LunarDate {
	// Chuyển về múi giờ Việt Nam
	vietnamTime := solar.In(time.FixedZone("ICT", 7*3600))
	
	// Sử dụng hàm ConvertSolar2Lunar từ lunar_date.go
	lunarDay, lunarMonth, lunarYear, lunarLeap := ConvertSolar2Lunar(
		vietnamTime.Day(),
		int(vietnamTime.Month()),
		vietnamTime.Year(),
		lc.timeZone,
	)
	
	return LunarDate{
		Year:     lunarYear,
		Month:    lunarMonth,
		Day:      lunarDay,
		IsLeap:   lunarLeap == 1,
		LeapYear: lc.isLeapYear(lunarYear),
	}
}

// LunarToSolar converts lunar date to solar date
func (lc *LunarCalendar) LunarToSolar(year, month, day int) time.Time {
	return lc.LunarToSolarWithLeap(year, month, day, false)
}

// LunarToSolarWithLeap converts lunar date to solar date with leap month support
func (lc *LunarCalendar) LunarToSolarWithLeap(year, month, day int, isLeap bool) time.Time {
	lunarLeap := 0
	if isLeap {
		lunarLeap = 1
	}
	
	// Sử dụng hàm ConvertLunar2Solar từ lunar_date.go
	solarDay, solarMonth, solarYear := ConvertLunar2Solar(day, month, year, lunarLeap, lc.timeZone)
	
	// Kiểm tra ngày hợp lệ
	if solarDay == 0 && solarMonth == 0 && solarYear == 0 {
		// Ngày không hợp lệ, trả về zero time
		return time.Time{}
	}
	
	return time.Date(solarYear, time.Month(solarMonth), solarDay, 0, 0, 0, 0, time.UTC)
}

// GetLunarMonthDays returns number of days in a lunar month
func (lc *LunarCalendar) GetLunarMonthDays(year, month int) int {
	// Tính ngày đầu tháng và tháng sau
	firstDay := lc.LunarToSolar(year, month, 1)
	
	// Tính tháng sau
	nextMonth := month + 1
	nextYear := year
	if nextMonth > 12 {
		nextMonth = 1
		nextYear++
	}
	
	nextFirstDay := lc.LunarToSolar(nextYear, nextMonth, 1)
	
	// Tính số ngày
	duration := nextFirstDay.Sub(firstDay)
	return int(duration.Hours() / 24)
}

// isLeapYear kiểm tra năm nhuận âm lịch
func (lc *LunarCalendar) isLeapYear(year int) bool {
	// Sử dụng thuật toán kiểm tra năm nhuận dựa trên chu kỳ 19 năm
	// Trong chu kỳ 19 năm có 7 năm nhuận
	return (year*12+17)%19 < 12
}

// === Các hàm từ lunar_date.go ===

// intFunc mô phỏng hàm INT(d) của JS, tức là Math.floor(d)
func intFunc(d float64) int {
	return int(math.Floor(d))
}

// JdFromDate tính toán số ngày Julian (nguyên) của ngày dd/mm/yyyy.
// Công thức từ http://www.tondering.dk/claus/calendar.html
func JdFromDate(dd, mm, yy int) int {
	var a, y, m, jd int
	a = intFunc(float64(14-mm) / 12.0)
	y = yy + 4800 - a
	m = mm + 12*a - 3
	jd = dd + intFunc((153*float64(m)+2.0)/5.0) + 365*y + intFunc(float64(y)/4.0) - intFunc(float64(y)/100.0) + intFunc(float64(y)/400.0) - 32045
	if jd < 2299161 { // Lịch Julian
		jd = dd + intFunc((153*float64(m)+2.0)/5.0) + 365*y + intFunc(float64(y)/4.0) - 32083
	}
	return jd
}

// JdToDate chuyển đổi một số ngày Julian (nguyên) sang ngày/tháng/năm.
func JdToDate(jd int) (int, int, int) {
	var a, b, c, d, e, m, day, month, year int
	if jd > 2299160 { // Sau 5/10/1582, lịch Gregorian
		a = jd + 32044
		b = intFunc((4*float64(a) + 3.0) / 146097.0)
		c = a - intFunc((float64(b)*146097.0)/4.0)
	} else {
		b = 0
		c = jd + 32082
	}
	d = intFunc((4*float64(c) + 3.0) / 1461.0)
	e = c - intFunc((1461.0*float64(d))/4.0)
	m = intFunc((5*float64(e) + 2.0) / 153.0)
	day = e - intFunc((153.0*float64(m)+2.0)/5.0) + 1
	month = m + 3 - 12*intFunc(float64(m)/10.0)
	year = b*100 + d - 4800 + intFunc(float64(m)/10.0)
	return day, month, year
}

// newMoon tính toán thời điểm trăng mới thứ k.
func newMoon(k int) float64 {
	var T, T2, T3, dr, Jd1, M, Mpr, F, C1, deltat, JdNew float64
	kFloat := float64(k)

	T = kFloat / 1236.85 // Thời gian tính bằng thế kỷ Julian từ 1900-01-0.5
	T2 = T * T
	T3 = T2 * T
	dr = pi / 180.0

	Jd1 = 2415020.75933 + 29.53058868*kFloat + 0.0001178*T2 - 0.000000155*T3
	Jd1 = Jd1 + 0.00033*math.Sin((166.56+132.87*T-0.009173*T2)*dr) // Trăng mới trung bình

	M = 359.2242 + 29.10535608*kFloat - 0.0000333*T2 - 0.00000347*T3    // Dị thường trung bình của Mặt trời
	Mpr = 306.0253 + 385.81691806*kFloat + 0.0107306*T2 + 0.00001236*T3 // Dị thường trung bình của Mặt trăng
	F = 21.2964 + 390.67050646*kFloat - 0.0016528*T2 - 0.00000239*T3    // Argument vĩ độ của Mặt trăng

	C1 = (0.1734-0.000393*T)*math.Sin(M*dr) + 0.0021*math.Sin(2*dr*M)
	C1 = C1 - 0.4068*math.Sin(Mpr*dr) + 0.0161*math.Sin(dr*2*Mpr)
	C1 = C1 - 0.0004*math.Sin(dr*3*Mpr)
	C1 = C1 + 0.0104*math.Sin(dr*2*F) - 0.0051*math.Sin(dr*(M+Mpr))
	C1 = C1 - 0.0074*math.Sin(dr*(M-Mpr)) + 0.0004*math.Sin(dr*(2*F+M))
	C1 = C1 - 0.0004*math.Sin(dr*(2*F-M)) - 0.0006*math.Sin(dr*(2*F+Mpr))
	C1 = C1 + 0.0010*math.Sin(dr*(2*F-Mpr)) + 0.0005*math.Sin(dr*(2*Mpr+M))

	if T < -11 {
		deltat = 0.001 + 0.000839*T + 0.0002261*T2 - 0.00000845*T3 - 0.000000081*T*T3
	} else {
		deltat = -0.000278 + 0.000265*T + 0.000262*T2
	}

	JdNew = Jd1 + C1 - deltat
	return JdNew
}

// sunLongitude tính kinh độ của mặt trời tại bất kỳ thời điểm nào.
func sunLongitude(jdn float64) float64 {
	var T, T2, dr, M, L0, DL, L float64
	T = (jdn - 2451545.0) / 36525.0 // Thời gian tính bằng thế kỷ Julian từ 2000-01-01 12:00:00 GMT
	T2 = T * T
	dr = pi / 180.0                                                // độ sang radian
	M = 357.52910 + 35999.05030*T - 0.0001559*T2 - 0.00000048*T*T2 // Dị thường trung bình, độ
	L0 = 280.46645 + 36000.76983*T + 0.0003032*T2                  // Kinh độ trung bình, độ
	DL = (1.914600-0.004817*T-0.000014*T2)*math.Sin(dr*M) +
		(0.019993-0.000101*T)*math.Sin(dr*2*M) +
		0.000290*math.Sin(dr*3*M)
	L = L0 + DL // Kinh độ thực, độ
	L = L * dr
	L = L - pi*2.0*math.Floor(L/(pi*2.0)) // Chuẩn hóa về (0, 2*PI)
	return L
}

// getSunLongitude tính vị trí mặt trời vào nửa đêm của ngày có số Julian đã cho.
func getSunLongitude(dayNumber int, timeZone float64) int {
	jdn := float64(dayNumber) - 0.5 - timeZone/24.0
	sl := sunLongitude(jdn)
	return intFunc(sl / pi * 6.0)
}

// getNewMoonDay tính ngày trăng mới thứ k theo múi giờ đã cho.
func getNewMoonDay(k int, timeZone float64) int {
	nm := newMoon(k)
	return intFunc(nm + 0.5 + timeZone/24.0)
}

// getLunarMonth11 tìm ngày bắt đầu tháng 11 âm lịch của năm đã cho.
func getLunarMonth11(yy int, timeZone float64) int {
	off := JdFromDate(31, 12, yy) - 2415021
	k := intFunc(float64(off) / 29.530588853)
	nm := getNewMoonDay(k, timeZone)
	sunLong := getSunLongitude(nm, timeZone) // Kinh độ mặt trời lúc nửa đêm
	if sunLong >= 9 {
		nm = getNewMoonDay(k-1, timeZone)
	}
	return nm
}

// getLeapMonthOffset tìm chỉ số của tháng nhuận sau tháng bắt đầu vào ngày a11.
func getLeapMonthOffset(a11 int, timeZone float64) int {
	k := intFunc((float64(a11)-2415021.076998695)/29.530588853 + 0.5)
	last := 0
	i := 1 // Bắt đầu với tháng sau tháng 11 âm lịch
	arc := getSunLongitude(getNewMoonDay(k+i, timeZone), timeZone)
	for {
		last = arc
		i++
		arc = getSunLongitude(getNewMoonDay(k+i, timeZone), timeZone)
		if !(arc != last && i < 14) {
			break
		}
	}
	return i - 1
}

// ConvertSolar2Lunar chuyển đổi ngày dương lịch dd/mm/yyyy sang ngày âm lịch tương ứng.
// Trả về (lunarDay, lunarMonth, lunarYear, lunarLeap)
// lunarLeap = 1 nếu là tháng nhuận, 0 nếu không.
func ConvertSolar2Lunar(dd, mm, yy int, timeZone float64) (int, int, int, int) {
	dayNumber := JdFromDate(dd, mm, yy)
	k := intFunc((float64(dayNumber) - 2415021.076998695) / 29.530588853)
	monthStart := getNewMoonDay(k+1, timeZone)
	if monthStart > dayNumber {
		monthStart = getNewMoonDay(k, timeZone)
	}

	a11 := getLunarMonth11(yy, timeZone)
	b11 := a11
	var lunarYear int
	if a11 >= monthStart {
		lunarYear = yy
		a11 = getLunarMonth11(yy-1, timeZone)
	} else {
		lunarYear = yy + 1
		b11 = getLunarMonth11(yy+1, timeZone)
	}

	lunarDay := dayNumber - monthStart + 1
	diff := intFunc(float64(monthStart-a11) / 29.0)
	lunarLeap := 0
	lunarMonth := diff + 11

	if b11-a11 > 365 { // Nếu là năm nhuận
		leapMonthDiff := getLeapMonthOffset(a11, timeZone)
		if diff >= leapMonthDiff {
			lunarMonth = diff + 10
			if diff == leapMonthDiff {
				lunarLeap = 1
			}
		}
	}

	if lunarMonth > 12 {
		lunarMonth = lunarMonth - 12
	}
	if lunarMonth >= 11 && diff < 4 {
		lunarYear -= 1
	}

	return lunarDay, lunarMonth, lunarYear, lunarLeap
}

// ConvertLunar2Solar chuyển đổi ngày âm lịch sang ngày dương lịch tương ứng.
// lunarLeap = 1 nếu là tháng nhuận, 0 nếu không.
// Trả về (day, month, year)
func ConvertLunar2Solar(lunarDay, lunarMonth, lunarYear int, lunarLeap int, timeZone float64) (int, int, int) {
	var a11, b11 int
	if lunarMonth < 11 {
		a11 = getLunarMonth11(lunarYear-1, timeZone)
		b11 = getLunarMonth11(lunarYear, timeZone)
	} else {
		a11 = getLunarMonth11(lunarYear, timeZone)
		b11 = getLunarMonth11(lunarYear+1, timeZone)
	}

	k := intFunc(0.5 + (float64(a11)-2415021.076998695)/29.530588853)
	off := lunarMonth - 11
	if off < 0 {
		off += 12
	}

	if b11-a11 > 365 { // Năm nhuận
		leapOff := getLeapMonthOffset(a11, timeZone)
		leapMonth := leapOff - 2
		if leapMonth < 0 {
			leapMonth += 12
		}
		if lunarLeap != 0 && lunarMonth != leapMonth {
			return 0, 0, 0 // Ngày không hợp lệ
		} else if lunarLeap != 0 || off >= leapOff {
			off += 1
		}
	}

	monthStart := getNewMoonDay(k+off, timeZone)
	return JdToDate(monthStart + lunarDay - 1)
}
