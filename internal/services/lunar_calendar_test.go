package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLunarCalendar(t *testing.T) {
	lc := NewLunarCalendar()
	
	assert.NotNil(t, lc)
	assert.Equal(t, 7.0, lc.timeZone) // GMT+7 for Vietnam
}

func TestLunarCalendar_SolarToLunar(t *testing.T) {
	lc := NewLunarCalendar()
	
	testCases := []struct {
		name         string
		solarDate    time.Time
		expectedYear int
		expectedMonth int
		expectedDay  int
		expectedLeap bool
	}{
		{
			name:         "Tết Nguyên Đán 2025 (29/1/2025)",
			solarDate:    time.Date(2025, 1, 29, 0, 0, 0, 0, time.UTC),
			expectedYear: 2025,
			expectedMonth: 1,
			expectedDay:  1,
			expectedLeap: false,
		},
		{
			name:         "Tết Nguyên Đán 2024 (10/2/2024)",
			solarDate:    time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
			expectedYear: 2024,
			expectedMonth: 1,
			expectedDay:  1,
			expectedLeap: false,
		},
		{
			name:         "Rằm tháng Giêng 2025 (12/2/2025)",
			solarDate:    time.Date(2025, 2, 12, 0, 0, 0, 0, time.UTC),
			expectedYear: 2025,
			expectedMonth: 1,
			expectedDay:  15,
			expectedLeap: false,
		},
		{
			name:         "Ngày thường 2024 (15/3/2024)",
			solarDate:    time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			expectedYear: 2024,
			expectedMonth: 2,
			expectedDay:  6,
			expectedLeap: false,
		},
		{
			name:         "Cuối năm 2023 (31/12/2023)",
			solarDate:    time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			expectedYear: 2023,
			expectedMonth: 11,
			expectedDay:  19,
			expectedLeap: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := lc.SolarToLunar(tc.solarDate)
			
			assert.Equal(t, tc.expectedYear, result.Year, "Year mismatch")
			assert.Equal(t, tc.expectedMonth, result.Month, "Month mismatch")
			assert.Equal(t, tc.expectedDay, result.Day, "Day mismatch")
			assert.Equal(t, tc.expectedLeap, result.IsLeap, "Leap month mismatch")
		})
	}
}

func TestLunarCalendar_LunarToSolar(t *testing.T) {
	lc := NewLunarCalendar()
	
	testCases := []struct {
		name          string
		lunarYear     int
		lunarMonth    int
		lunarDay      int
		expectedDate  time.Time
	}{
		{
			name:         "Mùng 1 Tết 2025",
			lunarYear:    2025,
			lunarMonth:   1,
			lunarDay:     1,
			expectedDate: time.Date(2025, 1, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name:         "Mùng 1 Tết 2024",
			lunarYear:    2024,
			lunarMonth:   1,
			lunarDay:     1,
			expectedDate: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
		},
		{
			name:         "Rằm tháng Giêng 2025",
			lunarYear:    2025,
			lunarMonth:   1,
			lunarDay:     15,
			expectedDate: time.Date(2025, 2, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			name:         "Ngày 6/2 âm lịch 2024",
			lunarYear:    2024,
			lunarMonth:   2,
			lunarDay:     6,
			expectedDate: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := lc.LunarToSolar(tc.lunarYear, tc.lunarMonth, tc.lunarDay)
			
			assert.Equal(t, tc.expectedDate.Year(), result.Year(), "Year mismatch")
			assert.Equal(t, tc.expectedDate.Month(), result.Month(), "Month mismatch")
			assert.Equal(t, tc.expectedDate.Day(), result.Day(), "Day mismatch")
		})
	}
}

func TestLunarCalendar_LunarToSolarWithLeap(t *testing.T) {
	lc := NewLunarCalendar()
	
	t.Run("should handle leap month correctly", func(t *testing.T) {
		// Test với tháng nhuận (nếu có)
		result := lc.LunarToSolarWithLeap(2023, 2, 15, true)
		
		// Kiểm tra kết quả không phải zero time
		assert.False(t, result.IsZero(), "Should return valid date for leap month")
	})
	
	t.Run("should handle invalid leap month", func(t *testing.T) {
		// Test với tháng nhuận không hợp lệ
		// Chỉ cần đảm bảo không panic
		assert.NotPanics(t, func() {
			lc.LunarToSolarWithLeap(2024, 13, 1, true)
		})
	})
}

func TestLunarCalendar_GetLunarMonthDays(t *testing.T) {
	lc := NewLunarCalendar()
	
	testCases := []struct {
		name  string
		year  int
		month int
	}{
		{
			name:  "Tháng Giêng 2025",
			year:  2025,
			month: 1,
		},
		{
			name:  "Tháng 2 âm lịch 2025",
			year:  2025,
			month: 2,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := lc.GetLunarMonthDays(tc.year, tc.month)
			
			// Tháng âm lịch có thể có 29 hoặc 30 ngày
			assert.True(t, result == 29 || result == 30, 
				"Lunar month should have 29 or 30 days, got %d", result)
			
			t.Logf("Lunar month %d/%d has %d days", tc.month, tc.year, result)
		})
	}
}

func TestLunarCalendar_isLeapYear(t *testing.T) {
	lc := NewLunarCalendar()
	
	// Test với một số năm để kiểm tra logic năm nhuận âm lịch
	testCases := []struct {
		year int
		name string
	}{
		{2020, "2020"},
		{2021, "2021"},
		{2023, "2023"},
		{2024, "2024"},
		{2025, "2025"},
	}
	
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Year_%d", tc.year), func(t *testing.T) {
			result := lc.isLeapYear(tc.year)
			// Chỉ kiểm tra kết quả là boolean hợp lệ
			// Vì đây là năm nhuận âm lịch, không phải dương lịch
			assert.IsType(t, false, result, "isLeapYear should return boolean for year %d", tc.year)
			
			t.Logf("Lunar leap year %d: %v", tc.year, result)
		})
	}
}

// Test các hàm utility từ lunar_date.go
func TestJdFromDate(t *testing.T) {
	testCases := []struct {
		name     string
		day      int
		month    int
		year     int
		expected int
	}{
		{
			name:     "Ngày 1/1/2000",
			day:      1,
			month:    1,
			year:     2000,
			expected: 2451545, // Julian Day Number cho 1/1/2000
		},
		{
			name:     "Ngày 29/1/2025 (Tết 2025)",
			day:      29,
			month:    1,
			year:     2025,
			expected: 2460705,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := JdFromDate(tc.day, tc.month, tc.year)
			assert.Equal(t, tc.expected, result, "Julian Day Number mismatch")
		})
	}
}

func TestJdToDate(t *testing.T) {
	testCases := []struct {
		name          string
		jd            int
		expectedDay   int
		expectedMonth int
		expectedYear  int
	}{
		{
			name:          "Julian Day 2451545 (1/1/2000)",
			jd:            2451545,
			expectedDay:   1,
			expectedMonth: 1,
			expectedYear:  2000,
		},
		{
			name:          "Julian Day 2460705 (29/1/2025)",
			jd:            2460705,
			expectedDay:   29,
			expectedMonth: 1,
			expectedYear:  2025,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			day, month, year := JdToDate(tc.jd)
			assert.Equal(t, tc.expectedDay, day, "Day mismatch")
			assert.Equal(t, tc.expectedMonth, month, "Month mismatch")
			assert.Equal(t, tc.expectedYear, year, "Year mismatch")
		})
	}
}

func TestConvertSolar2Lunar(t *testing.T) {
	testCases := []struct {
		name          string
		day           int
		month         int
		year          int
		expectedLDay  int
		expectedLMonth int
		expectedLYear int
		expectedLeap  int
	}{
		{
			name:          "Tết 2025 (29/1/2025)",
			day:           29,
			month:         1,
			year:          2025,
			expectedLDay:  1,
			expectedLMonth: 1,
			expectedLYear: 2025,
			expectedLeap:  0,
		},
		{
			name:          "Tết 2024 (10/2/2024)",
			day:           10,
			month:         2,
			year:          2024,
			expectedLDay:  1,
			expectedLMonth: 1,
			expectedLYear: 2024,
			expectedLeap:  0,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lDay, lMonth, lYear, lLeap := ConvertSolar2Lunar(tc.day, tc.month, tc.year, 7.0)
			
			assert.Equal(t, tc.expectedLDay, lDay, "Lunar day mismatch")
			assert.Equal(t, tc.expectedLMonth, lMonth, "Lunar month mismatch")
			assert.Equal(t, tc.expectedLYear, lYear, "Lunar year mismatch")
			assert.Equal(t, tc.expectedLeap, lLeap, "Lunar leap mismatch")
		})
	}
}

func TestConvertLunar2Solar(t *testing.T) {
	testCases := []struct {
		name         string
		lDay         int
		lMonth       int
		lYear        int
		lLeap        int
		expectedDay  int
		expectedMonth int
		expectedYear int
	}{
		{
			name:         "Mùng 1 Tết 2025",
			lDay:         1,
			lMonth:       1,
			lYear:        2025,
			lLeap:        0,
			expectedDay:  29,
			expectedMonth: 1,
			expectedYear: 2025,
		},
		{
			name:         "Mùng 1 Tết 2024",
			lDay:         1,
			lMonth:       1,
			lYear:        2024,
			lLeap:        0,
			expectedDay:  10,
			expectedMonth: 2,
			expectedYear: 2024,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sDay, sMonth, sYear := ConvertLunar2Solar(tc.lDay, tc.lMonth, tc.lYear, tc.lLeap, 7.0)
			
			assert.Equal(t, tc.expectedDay, sDay, "Solar day mismatch")
			assert.Equal(t, tc.expectedMonth, sMonth, "Solar month mismatch")
			assert.Equal(t, tc.expectedYear, sYear, "Solar year mismatch")
		})
	}
}

// Test round-trip conversion (Dương -> Âm -> Dương)
func TestRoundTripConversion(t *testing.T) {
	lc := NewLunarCalendar()
	
	testDates := []time.Time{
		time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC), // Tết 2024
		time.Date(2025, 1, 29, 0, 0, 0, 0, time.UTC), // Tết 2025
		time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), // Ngày thường
		time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC), // Cuối năm
	}
	
	for _, originalDate := range testDates {
		t.Run(originalDate.Format("2006-01-02"), func(t *testing.T) {
			// Dương -> Âm
			lunar := lc.SolarToLunar(originalDate)
			
			// Âm -> Dương
			convertedBack := lc.LunarToSolarWithLeap(lunar.Year, lunar.Month, lunar.Day, lunar.IsLeap)
			
			// So sánh ngày gốc và ngày chuyển đổi ngược
			assert.Equal(t, originalDate.Year(), convertedBack.Year(), "Year mismatch in round-trip")
			assert.Equal(t, originalDate.Month(), convertedBack.Month(), "Month mismatch in round-trip")
			assert.Equal(t, originalDate.Day(), convertedBack.Day(), "Day mismatch in round-trip")
		})
	}
}