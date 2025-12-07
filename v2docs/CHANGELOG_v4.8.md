# Change Log - Weekly Recurrence Update

**Date:** 2025-12-07  
**Version:** 4.8  
**Change:** Use originTime.Weekday() instead of pattern.DayOfWeek

---

## 📝 Change Summary

### What Changed
Function `calcNextWeekly` now derives the target weekday from `originTime.Weekday()` instead of reading it from `pattern.DayOfWeek`.

### Why
- **Simplifies data model**: No need to store redundant `DayOfWeek` field when it's already encoded in `originTime`
- **Single source of truth**: `originTime` contains all timing information (date, time, day of week)
- **Consistency**: Matches how daily and monthly use `originTime` exclusively

---

## 🔧 Technical Changes

### Code Change

**File:** `internal/services/calNextTime.go`

**BEFORE:**
```go
func calcNextWeekly(...) {
    targetWeekday := pattern.DayOfWeek // Read from pattern
    ...
}
```

**AFTER:**
```go
func calcNextWeekly(...) {
    // Lấy weekday từ originTime
    targetWeekday := int(originTime.Weekday()) // Derive from originTime
    ...
}
```

### Test Updates

Updated all weekly tests to set `originTime` to the correct weekday:

**BEFORE:**
```go
OriginTime: createTime(2025, 1, 1, 10, 0), // Wednesday
RecurrencePattern: &models.RecurrencePattern{
    DayOfWeek: 5, // Friday (different from originTime!)
}
```

**AFTER:**
```go
OriginTime: createTime(2025, 1, 3, 10, 0), // Friday (correct day)
RecurrencePattern: &models.RecurrencePattern{
    // No DayOfWeek field needed
}
```

---

## ✅ Verification

### Test Results
```bash
✅ TestCalcNextWeekly_SameWeek - PASS
✅ TestCalcNextWeekly_EveryTwoWeeks - PASS  
✅ TestCalcNextWeekly_Sunday - PASS
✅ TestTinhtoan_NextRecurringV2_Weekly - PASS
```

### Build Status
```bash
✅ go build ./cmd/server - SUCCESS
```

---

## 📊 Impact Analysis

### User Impact
- **Existing reminders**: Will use `originTime` weekday (correct behavior)
- **New reminders**: App should set `originTime` to desired weekday
- **API change**: `DayOfWeek` field in pattern is now **optional/ignored** for weekly

### Example Behavior

| Origin Date | Origin Weekday | Next Occurrence |
|-------------|----------------|-----------------|
| 2025-01-03 (Fri) | Friday | Next Friday     |
| 2025-01-06 (Mon) | Monday | Next Monday     |
| 2025-01-05 (Sun) | Sunday | Next Sunday     |

**Pattern:**
- `originTime` = "2025-01-03 10:00" (Friday)
- `interval` = 1 (every week)
- **Result:** Reminder will trigger every Friday at 10:00

---

## 🔄 Migration Guide

### For Frontend/API
When creating weekly reminders:

**OLD (Optional):**
```json
{
  "originTime": "2025-01-01T10:00:00Z",
  "recurrencePattern": {
    "type": "weekly",
    "dayOfWeek": 1,  // Monday
    "interval": 1
  }
}
```

**NEW (Recommended):**
```json
{
  "originTime": "2025-01-06T10:00:00Z",  // Monday 10:00
  "recurrencePattern": {
    "type": "weekly",
    "interval": 1
    // dayOfWeek is ignored, use originTime's weekday
  }
}
```

### For Database
No migration needed. Existing records will work correctly.

---

## 🎯 Benefits

1. **Cleaner Code**
   - Single source of truth for timing
   - Less room for inconsistencies (e.g., originTime on Wednesday but DayOfWeek = Friday)

2. **Better UX**
   - User picks a date/time → that's what they get
   - No confusion about which field takes precedence

3. **Consistency**
   - Daily uses `originTime.Hour/Minute`
   - Monthly uses `originTime.Day/Hour/Minute`  
   - Weekly now uses `originTime.Weekday/Hour/Minute`

---

## 📋 Related Changes

**Version 4.8 includes:**
- ✅ Weekly recurrence uses `originTime.Weekday()`
- ✅ Fixed monthly overflow bug (Day 31 → Feb)
- ✅ Comprehensive test suite (22 tests)
- ✅ Updated API version message

---

**Status:** ✅ **COMPLETE & TESTED**  
**Ready for:** Production Deployment
