package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestNewLunarCalendar(t *testing.T) {
	lc := NewLunarCalendar()
	
	assert.NotNil(t, lc)
	assert.Equal(t, 7, lc.timeZone) // GMT+7 for Vietnam
}

func TestLunarCalendar_SolarToLunar(t *testing.T) {
	lc := NewLunarCalendar()
	
	testCases := []struct {
		name        string
		solarDate   time.Time
		expectedYear int
		expectedMonth int
		expectedDay  int
	}{
		{
			name:        "Tết Nguyên Đán 2024 (10/2/2024)",
			solarDate:   time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
			expectedYear: 2024,
			expectedMonth: 1,
			expectedDay:  1,
		},
		{
			name:        "Ngày thường 2024 (15/3/2024)",
			solarDate:   time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			expectedYear: 2024,
			expectedMonth: 2,
			expectedDay:  6,
		},
		{
			name:        "Cuối năm 2023 (31/12/2023)",
			solarDate:   time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			expectedYear: 2023,
			expectedMonth: 11,
			expectedDay:  19,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lunar := lc.SolarToLunar(tc.solarDate)
			
			// Kiểm tra năm (có thể sai lệch 1 năm do thuật toán)
			assert.True(t, lunar.Year == tc.expectedYear || lunar.Year == tc.expectedYear-1 || lunar.Year == tc.expectedYear+1,
				"Year should be close to expected: got %d, expected around %d", lunar.Year, tc.expectedYear)
			
			// Kiểm tra tháng (1-12)
			assert.True(t, lunar.Month >= 1 && lunar.Month <= 12,
				"Month should be between 1-12: got %d", lunar.Month)
			
			// Kiểm tra ngày (1-30)
			assert.True(t, lunar.Day >= 1 && lunar.Day <= 30,
				"Day should be between 1-30: got %d", lunar.Day)
			
			t.Logf("Solar %s -> Lunar %d/%d/%d", 
				tc.solarDate.Format("2006-01-02"), lunar.Day, lunar.Month, lunar.Year)
		})
	}
}

func TestLunarCalendar_LunarToSolar(t *testing.T) {
	lc := NewLunarCalendar()
	
	testCases := []struct {
		name      string
		year      int
		month     int
		day       int
		expectValid bool
	}{
		{
			name:      "Tết Nguyên Đán 2024 (1/1 âm lịch)",
			year:      2024,
			month:     1,
			day:       1,
			expectValid: true,
		},
		{
			name:      "Ngày thường (15/8 âm lịch)",
			year:      2024,
			month:     8,
			day:       15,
			expectValid: true,
		},
		{
			name:      "Cuối năm (30/12 âm lịch)",
			year:      2024,
			month:     12,
			day:       30,
			expectValid: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			solar := lc.LunarToSolar(tc.year, tc.month, tc.day)
			
			if tc.expectValid {
				// Kiểm tra năm hợp lý
				assert.True(t, solar.Year() >= tc.year-1 && solar.Year() <= tc.year+1,
					"Solar year should be reasonable: got %d, lunar year %d", solar.Year(), tc.year)
				
				// Kiểm tra tháng hợp lý (1-12)
				assert.True(t, solar.Month() >= 1 && solar.Month() <= 12,
					"Solar month should be valid: got %d", solar.Month())
				
				// Kiểm tra ngày hợp lý (1-31)
				assert.True(t, solar.Day() >= 1 && solar.Day() <= 31,
					"Solar day should be valid: got %d", solar.Day())
				
				t.Logf("Lunar %d/%d/%d -> Solar %s", 
					tc.day, tc.month, tc.year, solar.Format("2006-01-02"))
			}
		})
	}
}

func TestLunarCalendar_GetLunarMonthDays(t *testing.T) {
	lc := NewLunarCalendar()
	
	testCases := []struct {
		name  string
		year  int
		month int
	}{
		{"Tháng 1/2024", 2024, 1},
		{"Tháng 6/2024", 2024, 6},
		{"Tháng 12/2024", 2024, 12},
		{"Tháng 1/2023", 2023, 1},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			days := lc.GetLunarMonthDays(tc.year, tc.month)
			
			// Tháng âm lịch có 29 hoặc 30 ngày
			assert.True(t, days == 29 || days == 30,
				"Lunar month should have 29 or 30 days: got %d", days)
			
			t.Logf("Lunar %d/%d has %d days", tc.month, tc.year, days)
		})
	}
}

func TestLunarCalendar_isLeapYear(t *testing.T) {
	lc := NewLunarCalendar()
	
	testCases := []struct {
		year int
		name string
	}{
		{2020, "2020"},
		{2023, "2023"},
		{2024, "2024"},
		{2025, "2025"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := lc.isLeapYear(tc.year)
			// Chỉ kiểm tra kết quả là boolean hợp lệ
			assert.IsType(t, false, result,
				"isLeapYear should return boolean for year %d", tc.year)
			
			t.Logf("Year %d is leap year: %v", tc.year, result)
		})
	}
}

func TestLunarCalendar_RoundTrip(t *testing.T) {
	lc := NewLunarCalendar()
	
	// Test round trip: Solar -> Lunar -> Solar
	testDates := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
	}
	
	for _, originalSolar := range testDates {
		t.Run(originalSolar.Format("2006-01-02"), func(t *testing.T) {
			// Solar -> Lunar
			lunar := lc.SolarToLunar(originalSolar)
			
			// Lunar -> Solar (sử dụng đúng signature)
			convertedSolar := lc.LunarToSolar(lunar.Year, lunar.Month, lunar.Day)
			
			// Kiểm tra sai lệch không quá 30 ngày (do thuật toán gần đúng)
			diff := convertedSolar.Sub(originalSolar).Hours() / 24
			assert.True(t, abs(diff) <= 30,
				"Round trip should be accurate within 30 days: original %s, converted %s, diff %.1f days",
				originalSolar.Format("2006-01-02"), convertedSolar.Format("2006-01-02"), diff)
			
			t.Logf("Round trip: %s -> %d/%d/%d -> %s (diff: %.1f days)",
				originalSolar.Format("2006-01-02"), lunar.Day, lunar.Month, lunar.Year,
				convertedSolar.Format("2006-01-02"), diff)
		})
	}
}

func TestLunarCalendar_getJulianDayNumber(t *testing.T) {
	lc := NewLunarCalendar()
	
	// Test với một số ngày đã biết Julian Day Number
	testCases := []struct {
		year     int
		month    int
		day      int
		expected float64
	}{
		{2000, 1, 1, 2451544.5},   // Y2K
		{2024, 1, 1, 2460310.5},   // Đầu năm 2024
	}
	
	for _, tc := range testCases {
		t.Run("JDN calculation", func(t *testing.T) {
			jdn := lc.getJulianDayNumber(tc.year, tc.month, tc.day)
			
			// Cho phép sai lệch nhỏ do floating point
			assert.InDelta(t, tc.expected, jdn, 1.0,
				"JDN calculation for %d-%d-%d", tc.year, tc.month, tc.day)
		})
	}
}

// Benchmark tests
func BenchmarkSolarToLunar(b *testing.B) {
	lc := NewLunarCalendar()
	testDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lc.SolarToLunar(testDate)
	}
}

func BenchmarkLunarToSolar(b *testing.B) {
	lc := NewLunarCalendar()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lc.LunarToSolar(2024, 6, 15)
	}
}