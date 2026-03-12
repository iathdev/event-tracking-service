# Curl Examples - Tất cả Events

> Thay `$TOKEN` bằng JWT token hợp lệ.
> Base URL: `http://localhost:8080`

---

### #1 Login — `log_in_success`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "log_in_success",
    "screen": "login",
    "user_id": 456,
    "properties": {
      "user_name": "candidate@example.com"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #2 Your tests — `click_action_your_test`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "click_action_your_test",
    "screen": "your_tests",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "action_your_test": "Start test"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #3 Test Direction — `click_action_test_direction`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "click_action_test_direction",
    "screen": "test_direction",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "action_test_direction": "continue"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #4 Regulation — `click_agree_exam_regulation`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "click_agree_exam_regulation",
    "screen": "regulation",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "action_regulation": "agree"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #5 Check your Audio — `click_continue_audio`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "click_continue_audio",
    "screen": "check_audio",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Listening"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #6 Check your Audio & Microphone — `click_audio_checking`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "click_audio_checking",
    "screen": "check_audio_microphone",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Speaking"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #7 Check your Audio & Microphone — `click_test_mircophone`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "click_test_mircophone",
    "screen": "check_audio_microphone",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Speaking",
      "record_error": ""
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #8 Check your Audio & Microphone — `click_continue_audio_microphone`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "click_continue_audio_microphone",
    "screen": "check_audio_microphone",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Speaking"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #9 Skill test direction — `confirm_skill_direction_toeic`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "confirm_skill_direction_toeic",
    "screen": "skill_test_direction",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "TOEIC",
      "skill": "Listening"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #10 Countdown to test — `view_countdown`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "view_countdown",
    "screen": "ready_to_start_test",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Listening"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #11 Part direction — `view_part_direction_toiec`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "view_part_direction_toiec",
    "screen": "part_direction",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "TOEIC",
      "skill": "Listening",
      "submission_skill_id": "abc-123",
      "part_id": "part-1"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #12 Ready to start test skill — `click_start_test_skill_ielts`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "click_start_test_skill_ielts",
    "screen": "ready_to_start_test",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Reading",
      "submission_skill_id": "abc-123"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #13 Edit volume — `edit_volume_audio_playing_ielts`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "edit_volume_audio_playing_ielts",
    "screen": "listening_test",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Listening",
      "submission_skill_id": "abc-123",
      "volume_value": "75"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #14 Marking question — `marking_question_ielts`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "marking_question_ielts",
    "screen": "reading_test",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Reading",
      "submission_skill_id": "abc-123",
      "question_id": "q-42"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #15 Điều hướng câu hỏi — `click_question`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "click_question",
    "screen": "reading_test",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Reading",
      "submission_skill_id": "abc-123",
      "action_question": "next"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #16 Note question — `note_question_ielts`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "note_question_ielts",
    "screen": "reading_test",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Reading",
      "submission_skill_id": "abc-123",
      "question_id": "q-42"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #17 Highlight question — `highlight_question_ielts`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "highlight_question_ielts",
    "screen": "reading_test",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Reading",
      "submission_skill_id": "abc-123",
      "question_id": "q-42"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #18 View note — `view_note_ielts`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "view_note_ielts",
    "screen": "reading_test",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Reading",
      "submission_skill_id": "abc-123",
      "question_id": "q-42"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #19 Delete note — `delete_note_ielts`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "delete_note_ielts",
    "screen": "reading_test",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Reading",
      "submission_skill_id": "abc-123",
      "question_id": "q-42"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #20 Writing test — `do_writing_test_toiec`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "do_writing_test_toiec",
    "screen": "writing_test",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "TOEIC",
      "skill": "Writing",
      "submission_skill_id": "abc-123",
      "action_question": "paste"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #21 Submit test skill — `submit_test_skill`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "submit_test_skill",
    "screen": "submit_test_skill",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Reading",
      "submission_skill_id": "abc-123",
      "submit_by": "Candidate"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #22 System submit test skill — `system_submit_test_skill`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "system_submit_test_skill",
    "screen": "submit_test_skill",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Listening",
      "submission_skill_id": "abc-123",
      "submit_by": "System"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #23 Anti cheating — `tracking_cheating`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "tracking_cheating",
    "screen": "listening_test",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "TOEIC",
      "skill": "Listening",
      "submission_skill_id": "abc-123",
      "anti_cheating_type": "tab switching",
      "time_of_cheating": "2026-03-16T10:05:00Z"
    }
  }'
```

---

### #24 Tracking internet — `tracking_internet`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "tracking_internet",
    "screen": "reading_test",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Reading",
      "submission_skill_id": "abc-123",
      "network_error": "No internet"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #25 Tracking answer — `tracking_log_answer`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "tracking_log_answer",
    "screen": "reading_test",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Reading",
      "submission_skill_id": "abc-123",
      "candidate_answer": "B",
      "error_code": ""
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #26 Break time — `break_time`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "break_time",
    "screen": "break_time",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Listening",
      "submission_skill_id": "abc-123"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #27 Microphone not found — `microphone_not_found`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "microphone_not_found",
    "screen": "speaking_test",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "skill": "Speaking",
      "submission_skill_id": "abc-123",
      "question_id": "q-10",
      "record_error": "device_not_found"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```

---

### #28 Complete test — `complete_test`

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "complete_test",
    "screen": "end_test",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "batch_id": 123,
      "batch_candidate_id": 789,
      "product_line": "IELTS",
      "submission_id": "sub-456",
      "submit_by": "Candidate"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'
```
