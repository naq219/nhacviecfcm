# CalNextTime Test Suite

## Mô tả

Test suite toàn diện cho module `calNextTime.go` - module tính toán thời điểm lặp lại tiếp theo cho reminders.

## Cấu trúc Tests

### 1. **Daily Recurrence Tests** (`TestCalcNextDailyTime_*`)
- ✅ `SimpleDaily`: Test lặp hàng ngày cơ bản
- ✅ `EveryTwoDays`: Test lặp mỗi 2 ngày
- ✅ `CrossMonth`: Test qua tháng mới
- ✅ `ZeroInterval`: Edge case interval = 0

**Coverage:**
- Daily interval (1, 2, N ngày)
- Timezone handling
- Month boundary

### 2. **Weekly Recurrence Tests** (`TestCalcNextWeekly_*`)
- ✅ `SameWeek`: Tìm ngày trong tuần hiện tại
- ✅ `EveryTwoWeeks`: Lặp mỗi 2 tuần
- ✅ `Sunday`: Test với Sunday (weekday = 0)
- ✅ `NoOriginTime`: Edge case thiếu originTime

**Coverage:**
- DayOfWeek (0-6: Sunday-Saturday)
- Interval (1, 2, N tuần)
- Week boundary

### 3. **Monthly Solar Recurrence Tests** (`TestCalcNextSolarMonthly_*`)
- ✅ `SimpleMonthly`: Lặp hàng tháng cơ bản
- ✅ `Day31ToFebruary`: Ngày 31 -> tháng 2 (chỉ có 28 ngày)
- ✅ `EveryTwoMonths`: Lặp mỗi 2 tháng
- ✅ `LeapYear`: Test năm nhuận (Feb có 29 ngày)

**Coverage:**
- Monthly interval
- Month với số ngày khác nhau (28, 29, 30, 31)
- Leap year handling

### 4. **Interval Seconds Tests** (`TestCalcNextIntervalSeconds_*`)
- ✅ `OneHour`: Interval 1 giờ (3600s)
- ✅ `SystemDowntime`: Hệ thống down 2 giờ, catch-up logic

**Coverage:**
- Various intervals (seconds, minutes, hours)
- System downtime recovery

### 5. **Last Day of Month Tests** (`TestCalcNextSolarLastDayOfMonth_*`)
- ✅ `SameMonth`: Tìm ngày cuối tháng hiện tại
- ✅ `AlreadyLastDay`: Đã là ngày cuối -> tháng sau
- ✅ `February`: Tháng 2 (28 ngày)
- ✅ `LeapYearFebruary`: Tháng 2 năm nhuận (29 ngày)

**Coverage:**
- Các tháng có số ngày khác nhau
- Leap year

### 6. **Integration Tests** (`TestTinhtoan_NextRecurringV2_*`)
- ✅ `Daily`: Integration test cho daily
- ✅ `Weekly`: Integration test cho weekly
- ✅ `NilPattern`: Test error handling
- ✅ `FromAPIComplete_ShouldError`: Test validation

### 7. **Benchmark Tests**
- `BenchmarkCalcNextDailyTime`
- `BenchmarkCalcNextWeekly`
- `BenchmarkCalcNextSolarMonthly`

## Chạy Tests

### Chạy tất cả tests:
```bash
go test ./internal/services -v
```

### Chạy test specific pattern:
```bash
go test ./internal/services -run TestCalcNextDaily -v
go test ./internal/services -run TestCalcNextWeekly -v
go test ./internal/services -run TestCalcNextSolarMonthly -v
```

### Chạy single test:
```bash
go test ./internal/services -run TestCalcNextDailyTime_SimpleDaily -v
```

### Chạy benchmarks:
```bash
go test ./internal/services -bench=. -benchmem
```

### Chạy với coverage:
```bash
go test ./internal/services -cover
go test ./internal/services -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Cases Matrix

| Function | Test Cases | Edge Cases Covered |
|----------|------------|-------------------|
| `calcNextDailyTime` | 4 | Interval, Month boundary, Zero interval |
| `calcNextWeekly` | 4 | Weekday 0-6, Multi-week, Missing origin |
| `calcNextSolarMonthly` | 4 | 28-31 days, Leap year, Interval |
| `calcNextIntervalSeconds` | 2 | Downtime recovery |
| `calcNextSolarLastDayOfMonth` | 4 | Leap year, Different months |
| Integration | 4 | Error validation, API complete check |

## Expected Results

### All Tests Should Pass
```
PASS: TestCalcNextDailyTime_SimpleDaily
PASS: TestCalcNextDailyTime_EveryTwoDays
PASS: TestCalcNextDailyTime_CrossMonth
PASS: TestCalcNextWeekly_SameWeek
PASS: TestCalcNextWeekly_EveryTwoWeeks
PASS: TestCalcNextWeekly_Sunday
PASS: TestCalcNextSolarMonthly_SimpleMonthly
PASS: TestCalcNextSolarMonthly_Day31ToFebruary
PASS: TestCalcNextSolarMonthly_EveryTwoMonths
PASS: TestCalcNextSolarMonthly_LeapYear
PASS: TestCalcNextIntervalSeconds_OneHour
PASS: TestCalcNextIntervalSeconds_SystemDowntime
PASS: TestCalcNextSolarLastDayOfMonth_SameMonth
PASS: TestCalcNextSolarLastDayOfMonth_AlreadyLastDay
PASS: TestCalcNextSolarLastDayOfMonth_February
PASS: TestCalcNextSolarLastDayOfMonth_LeapYearFebruary
PASS: TestTinhtoan_NextRecurringV2_Daily
PASS: TestTinhtoan_NextRecurringV2_Weekly
PASS: TestTinhtoan_NextRecurringV2_NilPattern
PASS: TestTinhtoan_NextRecurringV2_FromAPIComplete_ShouldError
PASS: TestCalcNextDaily_ZeroInterval
PASS: TestCalcNextWeekly_NoOriginTime
```

## Known Limitations

1. **Lunar Calendar Tests**: Chưa test `calcNextLunarMonthly` vì phụ thuộc vào `SolarToLunarMap`
2. **Timezone Tests**: Hiện tại tests sử dụng UTC, chưa test với Vietnam timezone
3. **Long-term Recurrence**: Chưa test các recurrence span nhiều năm

## Future Enhancements

- [ ] Add tests for `calcNextLunarMonthly`
- [ ] Add timezone-specific tests (VN, US, EU)
- [ ] Add property-based testing (generative tests)
- [ ] Add fuzz testing for edge cases
- [ ] Add performance regression tests

## Troubleshooting

### Test Failures

**Nếu test fail với time mismatch:**
- Check timezone: Ensure all times use UTC
- Check daylight saving: STD/DST changes
- Check leap year logic

**Nếu test timeout:**
```bash
go test ./internal/services -timeout 30s
```

**Nếu muốn xem chi tiết:**
```bash
go test ./internal/services -v -count=1
```

## Contributing

Khi thêm feature mới vào `calNextTime.go`:
1. Thêm ít nhất 3 test cases (happy path, edge case, error case)
2. Chạy `go test` để verify
3. Chạy `go test -race` để check race conditions
4. Update test matrix trong README này

---

**Last Updated:** 2025-12-07  
**Test Count:** 22 tests  
**Coverage Target:** >85%
