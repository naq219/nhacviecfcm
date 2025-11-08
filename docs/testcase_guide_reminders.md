# RemiAq - Complete Test Cases & Regression Checklist

## 📋 Test Cases Matrix

### Group 1: ONE-TIME Reminders

#### Test 1.1: One-time, no CRP (max_crp=0)
**Scenario:**
- Create: `type=one_time, max_crp=0`
- Expected: Send once, mark completed immediately

**Check after each fix:**
- ✅ Notification sent 1 time only
- ✅ Status → "completed"
- ✅ next_action_at → empty
- ✅ No CRP retry

**Code affected:** `processCRP()` → if one_time AND crp_count >= max_crp

---

#### Test 1.2: One-time with CRP (max_crp=3, interval=20s)
**Scenario:**
- Create: `type=one_time, max_crp=3, crp_interval_sec=20`
- Send at: 12:00, 12:20, 12:40
- Expected: After 3 sends → completed

**Check after each fix:**
- ✅ CRP 1: Send, crp_count=1, next_crp=12:20
- ✅ CRP 2: Send, crp_count=2, next_crp=12:40
- ✅ CRP 3: Send, crp_count=3, status="completed"
- ✅ next_action_at → empty

**Code affected:** `processCRP()` → one_time quota check

---

#### Test 1.3: One-time, user complete at CRP 2 (before quota)
**Scenario:**
- Create: `max_crp=3, crp_interval_sec=20`
- CRP 1: 12:00 ✅
- CRP 2: 12:20 ✅
- User complete: 12:20 (after CRP 2)
- Expected: Stop immediately, no CRP 3

**Check after each fix:**
- ✅ Status → "completed"
- ✅ CRP 3 NOT sent ❌
- ✅ next_action_at → empty

**Code affected:** `OnUserComplete()` → one_time case

---

### Group 2: RECURRING + repeat_strategy = "none"

#### Test 2.1: Recurring none, auto-repeat (interval_seconds=180)
**Scenario:**
- Create: `type=recurring, repeat_strategy=none, interval_seconds=180`
- FRP 1: 12:00
- FRP 2: 12:03
- FRP 3: 12:06
- Expected: Auto-repeat every 3 minutes forever

**Check after each fix:**
- ✅ FRP 1: Send, next_recurring=12:03
- ✅ FRP 2: Send, next_recurring=12:06
- ✅ FRP 3: Send, next_recurring=12:09
- ✅ No user complete needed

**Code affected:** `processFRP()` → repeat_strategy=none case

---

#### Test 2.2: Recurring none with CRP (max_crp=2, interval=20s)
**Scenario:**
- Create: `repeat_strategy=none, max_crp=2, crp_interval_sec=20`
- FRP 1: 12:00
  - CRP 1: 12:20
  - CRP 2: 12:40 (quota reached)
- FRP 2: 12:03
  - CRP 1: 12:23
  - CRP 2: 12:43 (quota reached)
- Expected: CRP resets every FRP cycle

**Check after each fix:**
- ✅ FRP trigger resets crp_count=0
- ✅ CRP 1,2 per FRP cycle
- ✅ When CRP quota reached → next_action_at = next_recurring ✅

**Code affected:** `processFRP()` + `processCRP()`

---

#### Test 2.3: Recurring none, user complete (should continue auto-repeat)
**Scenario:**
- FRP 1: 12:00
- User bấm complete: 12:05
- Expected: FRP 2 vẫn trigger lúc 12:03 (theo lịch cũ)

**Check after each fix:**
- ✅ User complete → crp_count=0, next_recurring not changed (tính từ old base)
- ✅ FRP 2 trigger at 12:03
- ⚠️ **REGRESSION RISK**: Đừng tính lại next_recurring từ completion time!

**Code affected:** `OnUserComplete()` → repeat_strategy=none case

---

### Group 3: RECURRING + repeat_strategy = "crp_until_complete" (NEW)

#### Test 3.1: Recurring crp_until_complete, quota not reached
**Scenario:**
- Create: `repeat_strategy=crp_until_complete, max_crp=3, interval=180s`
- FRP 1: 12:00
  - CRP 1: 12:20
  - CRP 2: 12:40
  - CRP 3: 13:00 (quota reached)
- Expected: WAIT FOR USER, don't trigger FRP 2

**Check after each fix:**
- ✅ FRP 1: Send, next_recurring=12:03 (calculated)
- ✅ CRP 1,2,3: Send with 20s interval
- ✅ After CRP quota: next_action_at = EMPTY (not next_recurring!) ❌ IMPORTANT!
- ✅ Worker skip this reminder

**Code affected:** `processCRP()` → recurring + crp_until_complete + quota reached

---

#### Test 3.2: Recurring crp_until_complete, user complete at CRP 2
**Scenario:**
- FRP 1: 12:00
- CRP 1: 12:20
- CRP 2: 12:40
- User complete: 12:40 (at CRP 2, before CRP 3)
- Expected: STOP CRP immediately, wait for FRP from 12:40

**Check after each fix:**
- ✅ crp_count → 0 (reset)
- ✅ next_recurring → 12:43 (3 min from 12:40) ✅ NEW CALC!
- ✅ next_action_at → 12:43
- ✅ CRP 3 NOT sent ❌
- ✅ FRP 2 will trigger at 12:43

**Code affected:** `OnUserComplete()` → recurring + crp_until_complete

---

#### Test 3.3: Recurring crp_until_complete, user complete after quota
**Scenario:**
- FRP 1: 12:00
- CRP 1,2,3: 12:20, 12:40, 13:00 (quota reached)
- next_action_at = EMPTY (waiting)
- User complete: 13:05
- Expected: Calculate FRP from 13:05

**Check after each fix:**
- ✅ crp_count → 0
- ✅ next_recurring → 13:08 (3 min from 13:05) ✅
- ✅ next_action_at → 13:08
- ✅ FRP 2 trigger at 13:08

**Code affected:** `OnUserComplete()` → same case as 3.2

---

### Group 4: Snooze (both types)

#### Test 4.1: One-time snoozed
**Scenario:**
- Create: `one_time, max_crp=3`
- FRP 1: 12:00
- User snooze: 12:05 for 10 minutes
- Expected: Skip until 12:15

**Check after each fix:**
- ✅ snooze_until = 12:15
- ✅ next_action_at = 12:15
- ✅ Worker skip from 12:05 to 12:15
- ✅ CRP resume at 12:15

**Code affected:** `processReminder()` → snooze check

---

#### Test 4.2: Recurring crp_until_complete snoozed during CRP
**Scenario:**
- FRP 1: 12:00
- CRP 1: 12:20
- User snooze: 12:25 for 5 minutes
- Expected: next_action_at = 12:30 (snooze time)

**Check after each fix:**
- ✅ snooze_until = 12:30
- ✅ next_action_at = 12:30 (not next_crp)
- ✅ CRP 2 resume at 12:30

**Code affected:** `processReminder()` + `CalculateNextActionAt()`

---

### Group 5: Edge Cases

#### Test 5.1: Recurring none, FRP trigger multiple times rapidly
**Scenario:**
- Worker down 10 minutes
- Create: `interval_seconds=180` (3 min)
- Worker restart
- Expected: Catch up, send all missed FRP

**Check after each fix:**
- ✅ next_recurring auto-jump forward
- ✅ Send current FRP only (not all missed)
- ✅ next_recurring set to next cycle

**Code affected:** `CalculateNextRecurring()` → calculateNextIntervalSeconds()

---

#### Test 5.2: Recurring crp_until_complete, user complete then complete again immediately
**Scenario:**
- FRP 1: 12:00
- User complete: 12:01
- next_recurring = 12:04
- User complete AGAIN: 12:04 (exactly at next FRP time)
- Expected: next_recurring = 12:07

**Check after each fix:**
- ✅ First complete: next_recurring = 12:04
- ✅ Second complete: next_recurring = 12:07 (not 12:04 again)

**Code affected:** `OnUserComplete()` → CalculateNextRecurring() from now

---

#### Test 5.3: Recurring none with very short CRP (1 second)
**Scenario:**
- Create: `crp_interval_sec=1, max_crp=100`
- Expected: Send 100 times with 1 second interval

**Check after each fix:**
- ✅ next_crp calculated correctly each time
- ✅ Doesn't spam faster than worker cycle (10s)
- ✅ All 100 sent eventually

**Code affected:** `processCRP()` → next_crp calculation

---

## ✅ Regression Checklist (Run After Each Fix)

### Before making ANY change:
1. **Run all 12 unit tests**
   ```bash
   go test -v ./internal/worker/
   ```
   Expected: 12/12 PASS

2. **Manual integration test** (10 minutes)
   - Create reminder each type
   - Watch worker logs
   - Verify each phase

---

### After fix to `processFRP()`:
- [ ] Test 2.1: Recurring none auto-repeat
- [ ] Test 2.2: Recurring none with CRP
- [ ] Test 3.1: Recurring crp_until_complete

---

### After fix to `processCRP()`:
- [ ] Test 1.2: One-time CRP quota
- [ ] Test 2.2: Recurring none with CRP
- [ ] Test 3.1: Recurring crp_until_complete quota

---

### After fix to `OnUserComplete()`:
- [ ] Test 1.3: One-time complete at CRP 2
- [ ] Test 2.3: Recurring none user complete
- [ ] Test 3.2: Recurring crp_until_complete user complete
- [ ] Test 3.3: Recurring crp_until_complete complete after quota

---

### After fix to `CalculateNextActionAt()`:
- [ ] Test 4.1: One-time snoozed
- [ ] Test 4.2: Recurring snoozed during CRP

---

## 🚨 Critical Conflict Points

### ⚠️ Conflict 1: repeat_strategy check location
**Issue:** If check `repeat_strategy` in WRONG place, breaks both types

```go
// ❌ WRONG: In processCRP()
if repeat_strategy == "crp_until_complete" {
    next_action_at = time.Time{}
}
// This breaks repeat_strategy="none" case!

// ✅ CORRECT: Check BOTH types separately
if repeat_strategy == "crp_until_complete" {
    next_action_at = time.Time{}  // Wait for user
} else {
    next_action_at = next_recurring  // Auto-repeat
}
```

---

### ⚠️ Conflict 2: One-time vs Recurring in OnUserComplete()
**Issue:** If apply recurring logic to one-time, breaks it

```go
// ❌ WRONG: Apply to all
reminder.NextRecurring = calculateNext()  // one-time doesn't have this!

// ✅ CORRECT: Type check first
if type == "one_time" {
    status = "completed"
    next_action_at = empty
} else if type == "recurring" {
    next_recurring = calculateNext()
}
```

---

### ⚠️ Conflict 3: next_action_at = empty breaks queries
**Issue:** If next_action_at stays empty too long, worker skips forever

```sql
-- Worker query
SELECT * FROM reminders 
WHERE next_action_at <= NOW()  -- Empty times filtered out!

-- If next_action_at = empty, query doesn't return it
-- User complete must SET it to next_recurring
```

---

## 📊 Test Coverage Summary

| Test # | Type | repeat_strategy | With CRP | User Complete | Status |
|--------|------|-----------------|----------|---------------|--------|
| 1.1 | one_time | N/A | No | N/A | ✅ |
| 1.2 | one_time | N/A | Yes (3) | No | ✅ |
| 1.3 | one_time | N/A | Yes (3) | Yes (CRP 2) | ✅ |
| 2.1 | recurring | none | No | No | ✅ |
| 2.2 | recurring | none | Yes | No | ✅ |
| 2.3 | recurring | none | No | Yes | ⚠️ NEW |
| 3.1 | recurring | crp_until_complete | Yes | No | 🔴 PRIORITY |
| 3.2 | recurring | crp_until_complete | Yes | Yes (CRP 2) | 🔴 PRIORITY |
| 3.3 | recurring | crp_until_complete | Yes | Yes (after quota) | 🔴 PRIORITY |
| 4.1 | one_time | N/A | Yes | Snooze | ✅ |
| 4.2 | recurring | crp_until_complete | Yes | Snooze | ⚠️ CHECK |
| 5.1 | recurring | none | N/A | No | ✅ |
| 5.2 | recurring | crp_until_complete | No | Yes (2x) | ⚠️ CHECK |
| 5.3 | recurring | none | Yes (100x) | No | ✅ |

---

## 🎯 Priority Order (Fix & Test)

1. **PRIORITY 1** (Core fixes)
   - [ ] Test 3.1 - crp_until_complete quota behavior
   - [ ] Test 3.2 - User complete before quota
   - [ ] Test 1.2 - One-time CRP works

2. **PRIORITY 2** (Regressions)
   - [ ] Test 2.1 - Recurring none doesn't break
   - [ ] Test 2.2 - Recurring none with CRP
   - [ ] Test 2.3 - User complete doesn't mess with none strategy

3. **PRIORITY 3** (Edge cases)
   - [ ] Test 4.1 - Snooze one-time
   - [ ] Test 4.2 - Snooze during CRP
   - [ ] Test 5.2 - Double complete