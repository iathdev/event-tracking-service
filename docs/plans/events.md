# Event Tracking - Danh sách Event & Properties

## Bảng event theo màn hình

| # | Flow | Event | Event Properties | User Properties |
|---|------|-------|------------------|-----------------|
| 1 | Login | `log_in_success` | `user_id` | `user_name` |
| 2 | Your tests | `click_action_your_test` | `action_your_test`, `batch_id`, `batch_candidate_id` | `user_id` |
| 3 | Test Direction | `click_action_test_direction` | `product_line`, `screen`, `action_test_direction`, `batch_id`, `batch_candidate_id`, `occurred_at` | `user_id` |
| 4 | Regulation | `click_agree_exam_regulation` | `product_line`, `screen`, `action_regulation`, `batch_id`, `batch_candidate_id`, `occurred_at` | `user_id` |
| 5 | Check your Audio | `click_continue_audio` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `occurred_at` | `user_id` |
| 6 | Check your Audio & Microphone | `click_audio_checking` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `occurred_at` | `user_id` |
| 7 | Check your Audio & Microphone | `click_test_mircophone` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `record_error`, `occurred_at` | `user_id` |
| 8 | Check your Audio & Microphone | `click_continue_audio_microphone` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `occurred_at` | `user_id` |
| 9 | Skill test direction | `confirm_skill_direction_toeic` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `occurred_at` | `user_id` |
| 10 | Countdown to test | `view_countdown` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `occurred_at` | `user_id` |
| 11 | Part direction | `view_part_direction_toiec` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `submission_skill_id`, `part_id`, `occurred_at` | `user_id` |
| 12 | Ready to start test skill | `click_start_test_skill_ielts` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `submission_skill_id`, `occurred_at` | `user_id` |
| 13 | Edit volume | `edit_volume_audio_playing_ielts` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `submission_skill_id`, `volume_value`, `occurred_at` | `user_id` |
| 14 | Marking question | `marking_question_ielts` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `submission_skill_id`, `question_id`, `occurred_at` | `user_id` |
| 15 | Điều hướng câu hỏi | `click_question` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `submission_skill_id`, `action_question`, `occurred_at` | `user_id` |
| 16 | Note question | `note_question_ielts` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `submission_skill_id`, `question_id`, `occurred_at` | `user_id` |
| 17 | Highlight question | `highlight_question_ielts` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `submission_skill_id`, `question_id`, `occurred_at` | `user_id` |
| 18 | View note | `view_note_ielts` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `submission_skill_id`, `question_id`, `occurred_at` | `user_id` |
| 19 | Delete note | `delete_note_ielts` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `submission_skill_id`, `question_id`, `occurred_at` | `user_id` |
| 20 | Writing test | `do_writing_test_toiec` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `submission_skill_id`, `action_question`, `occurred_at` | `user_id` |
| 21 | Submit test skill | `submit_test_skill` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `submission_skill_id`, `submit_by`, `occurred_at` | `user_id` |
| 22 | System submit test skill | `system_submit_test_skill` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `submission_skill_id`, `submit_by`, `occurred_at` | `user_id` |
| 23 | Anti cheating | `tracking_cheating` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `submission_skill_id`, `anti_cheating_type`, `time_of_cheating` | `user_id` |
| 24 | Tracking internet | `tracking_internet` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `submission_skill_id`, `network_error`, `occurred_at` | `user_id` |
| 25 | Tracking answer | `tracking_log_answer` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `submission_skill_id`, `candidate_answer`, `error_code`, `occurred_at` | `user_id` |
| 26 | Break time | `break_time` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `submission_skill_id`, `occurred_at` | `user_id` |
| 27 | Microphone not found | `microphone_not_found` | `product_line`, `screen`, `skill`, `batch_id`, `batch_candidate_id`, `submission_skill_id`, `question_id`, `record_error`, `occurred_at` | `user_id` |
| 28 | Complete test | `complete_test` | `product_line`, `screen`, `batch_id`, `batch_candidate_id`, `submission_id`, `submit_by`, `occurred_at` | `user_id` |

---

## Danh sách Properties

| Property name | Description | Data channel | Value |
|---------------|-------------|--------------|-------|
| product_line | Loại chứng chỉ thi | Enum | TOEIC, IELTS |
| batch_id | ID của batch | String | |
| batch_candidate_id | ID của batch candidate (gắn thí sinh với đợt thi) | String | |
| skill | Bài thi kỹ năng | Enum | Listening, Reading, Speaking, Writing |
| submission_id | ID của batch candidate | String | |
| submission_skill_id | ID của bài test skill mà candidate đã tham gia | String | |
| part_id | Id của part thi | String | |
| question_id | Id của câu hỏi | String | |
| user_id | ID của candidate đã tham gia thi | String | |
| occurred_at | Thời gian log event | Datetime | |
| start_time | Thời gian candidate bắt đầu làm skill đầu tiên trong bài test | Datetime | |
| finished_time | Thời gian submit toàn bộ bài test | Datetime | |
| submited_by | Người thực hiện submit bài test | Enum | Candidate, System |
| action_your_test | Hành động tương tác với your test | Enum | Start test, Continue, View result |
| screen | Tên màn hình đang được tracking | Enum | test direction, skill test direction, part direction, regulation, check audio, check audio & microphone, ready to start test, listening test, reading test, speaking test, break time, submit test skill, end test |
| action_test_direction | Hành động tương tác tại màn test direction | Enum | cancel, continue |
| action_regulation | Hành động tương tác tại màn regulation | Enum | check box agree, continue |
| volume_value | Giá trị âm lượng audio được thí sinh chọn | String | |
| action_question | Hành động tương tác với câu hỏi trong bài thi | Enum | previous, next, click part, click question, cut, paste, undo, redo |
| anti_cheating_type | Loại hành vi gian lận | Enum | screen existing, tab switching |
| time_of_cheating | Thời gian thực hiện cheating bắt đầu - kết thúc | Datetime | |
| network_error | Các loại mã lỗi liên quan đến kết nối mạng internet | Enum | Download, Upload, Ping, No internet |
| candidate_answer_error_code | Các loại mã lỗi liên quan đến lưu câu trả lời của thí sinh | Enum | |
| record_error | Các loại mã lỗi liên quan đến lưu file record của thí sinh | Enum | |
