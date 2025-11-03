package services

import (
	"math"
	"time"
)

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
	timeZone int // GMT+7 for Vietnam
}

// NewLunarCalendar creates a new lunar calendar service for Vietnam timezone
func NewLunarCalendar() *LunarCalendar {
	return &LunarCalendar{
		timeZone: 7, // GMT+7 for Vietnam
	}
}

// SolarToLunar converts solar date to lunar date using Vietnamese lunar calendar
func (lc *LunarCalendar) SolarToLunar(solar time.Time) LunarDate {
	// Chuyển về múi giờ Việt Nam
	vietnamTime := solar.In(time.FixedZone("ICT", 7*3600))
	
	// Tính Julian Day Number
	jd := lc.getJulianDayNumber(vietnamTime.Year(), int(vietnamTime.Month()), vietnamTime.Day())
	
	// Tìm tháng 11 âm lịch của năm trước
	k := int(math.Floor(float64(vietnamTime.Year()-1900)*12.3685))
	
	// Tìm new moon gần nhất
	nm := lc.getNewMoonDay(k, lc.timeZone)
	sunLong := lc.getSunLongitude(nm, lc.timeZone)
	
	// Điều chỉnh để tìm đúng tháng 11
	if sunLong >= 9 {
		nm = lc.getNewMoonDay(k-1, lc.timeZone)
	}
	
	// Tìm các new moon trong năm
	for i := 1; i <= 14; i++ {
		lastNm := nm
		nm = lc.getNewMoonDay(k+i, lc.timeZone)
		sunLong = lc.getSunLongitude(nm, lc.timeZone)
		
		if jd < nm {
			// Tính tháng và ngày âm lịch
			lunarMonth := i
			lunarDay := int(jd - lastNm + 1)
			lunarYear := vietnamTime.Year()
			
			// Điều chỉnh năm nếu cần
			if lunarMonth >= 11 {
				lunarYear = vietnamTime.Year() + 1
			}
			
			// Điều chỉnh tháng (tháng 11, 12 thuộc năm sau)
			if lunarMonth >= 11 {
				lunarMonth = lunarMonth - 12
			}
			if lunarMonth <= 0 {
				lunarMonth = lunarMonth + 12
			}
			
			return LunarDate{
				Year:     lunarYear,
				Month:    lunarMonth,
				Day:      lunarDay,
				IsLeap:   false, // Simplified - không tính tháng nhuận trong version này
				LeapYear: lc.isLeapYear(lunarYear),
			}
		}
	}
	
	// Fallback nếu không tìm được
	return LunarDate{
		Year:  vietnamTime.Year(),
		Month: int(vietnamTime.Month()),
		Day:   vietnamTime.Day(),
	}
}

// LunarToSolar converts lunar date to solar date
func (lc *LunarCalendar) LunarToSolar(year, month, day int) time.Time {
	// Tìm new moon của tháng đó
	k := int(math.Floor(float64(year-1900)*12.3685 + float64(month-1)))
	nm := lc.getNewMoonDay(k, lc.timeZone)
	
	// Tính ngày dương lịch
	jd := nm + float64(day-1)
	
	return lc.julianToGregorian(jd)
}

// GetLunarMonthDays returns number of days in a lunar month
func (lc *LunarCalendar) GetLunarMonthDays(year, month int) int {
	// Tính new moon của tháng hiện tại và tháng sau
	k := int(math.Floor(float64(year-1900)*12.3685 + float64(month-1)))
	nm1 := lc.getNewMoonDay(k, lc.timeZone)
	nm2 := lc.getNewMoonDay(k+1, lc.timeZone)
	
	return int(nm2 - nm1)
}

// isLeapYear kiểm tra năm nhuận âm lịch
func (lc *LunarCalendar) isLeapYear(year int) bool {
	// Năm nhuận âm lịch có 13 tháng thay vì 12 tháng
	// Chu kỳ 19 năm có 7 năm nhuận
	return (year*12+17)%19 < 12
}

// getJulianDayNumber tính Julian Day Number
func (lc *LunarCalendar) getJulianDayNumber(year, month, day int) float64 {
	if month <= 2 {
		year--
		month += 12
	}
	
	a := math.Floor(float64(year) / 100)
	b := 2 - a + math.Floor(a/4)
	
	jd := math.Floor(365.25*float64(year+4716)) + math.Floor(30.6001*float64(month+1)) + float64(day) + b - 1524.5
	
	return jd
}

// getNewMoonDay tính ngày new moon
func (lc *LunarCalendar) getNewMoonDay(k int, timeZone int) float64 {
	// Thuật toán tính new moon dựa trên Meeus
	T := float64(k) / 1236.85 // Time in Julian centuries
	T2 := T * T
	T3 := T2 * T
	T4 := T3 * T
	
	// Mean new moon
	Jd1 := 2415020.75933 + 29.53058868*float64(k) + 0.0001178*T2 - 0.000000155*T3 + 0.00000000033*T4
	
	// Sun's mean anomaly
	M := 2.5534 + 29.10535670*float64(k) - 0.0000014*T2 - 0.00000011*T3
	
	// Moon's mean anomaly  
	Mpr := 201.5643 + 385.81693528*float64(k) + 0.0107582*T2 + 0.00001238*T3 - 0.000000058*T4
	
	// Moon's argument of latitude
	F := 160.7108 + 390.67050284*float64(k) - 0.0016118*T2 - 0.00000227*T3 + 0.000000011*T4
	
	// Convert to radians
	M = M * math.Pi / 180
	Mpr = Mpr * math.Pi / 180
	F = F * math.Pi / 180
	
	// Corrections
	C1 := (0.1734-0.000393*T)*math.Sin(M) + 0.0021*math.Sin(2*M) - 0.4068*math.Sin(Mpr) + 0.0161*math.Sin(2*Mpr)
	C1 = C1 - 0.0004*math.Sin(3*Mpr) + 0.0104*math.Sin(2*F) - 0.0051*math.Sin(M+Mpr) - 0.0074*math.Sin(M-Mpr) + 0.0004*math.Sin(2*F+M)
	C1 = C1 - 0.0004*math.Sin(2*F-M) - 0.0006*math.Sin(2*F+Mpr) + 0.0010*math.Sin(2*F-Mpr) + 0.0005*math.Sin(M+2*Mpr)
	
	deltat := 0.0
	if T < -11 {
		deltat = 0.001 + 0.000839*T + 0.0002261*T2 - 0.00000845*T3 - 0.000000081*T4
	} else {
		deltat = -0.000278 + 0.000265*T + 0.000262*T2
	}
	
	JdNew := Jd1 + C1 - float64(timeZone)/24
	
	return math.Floor(JdNew + 0.5 + deltat)
}

// getSunLongitude tính kinh độ mặt trời
func (lc *LunarCalendar) getSunLongitude(jdn float64, timeZone int) float64 {
	T := (jdn - 2451545.0) / 36525 // Time in Julian centuries from J2000.0
	T2 := T * T
	
	// Mean longitude
	dr := math.Pi / 180
	L0 := 280.46645 + 36000.76983*T + 0.0003032*T2
	L0 = L0 * dr
	L0 = L0 - math.Pi*2*(math.Floor(L0/(math.Pi*2)))
	
	// Mean anomaly
	M := 357.52910 + 35999.05030*T - 0.0001559*T2 - 0.00000048*T*T2
	M = M * dr
	
	// Equation of center
	C := (1.914600 - 0.004817*T - 0.000014*T2)*math.Sin(M) + (0.019993-0.000101*T)*math.Sin(2*M) + 0.000290*math.Sin(3*M)
	C = C * dr
	
	// True longitude
	L := L0 + C
	L = L - math.Pi*2*(math.Floor(L/(math.Pi*2)))
	
	return math.Floor(L/math.Pi*6)
}

// julianToGregorian chuyển Julian Day về Gregorian date
func (lc *LunarCalendar) julianToGregorian(jd float64) time.Time {
	jd = jd + 0.5
	z := math.Floor(jd)
	f := jd - z
	
	var a float64
	if z < 2299161 {
		a = z
	} else {
		alpha := math.Floor((z - 1867216.25) / 36524.25)
		a = z + 1 + alpha - math.Floor(alpha/4)
	}
	
	b := a + 1524
	c := math.Floor((b - 122.1) / 365.25)
	d := math.Floor(365.25 * c)
	e := math.Floor((b - d) / 30.6001)
	
	day := int(b - d - math.Floor(30.6001*e) + f)
	
	var month int
	if e < 14 {
		month = int(e - 1)
	} else {
		month = int(e - 13)
	}
	
	var year int
	if month > 2 {
		year = int(c - 4716)
	} else {
		year = int(c - 4715)
	}
	
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
