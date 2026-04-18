# Narrative Patterns Reference

Tất cả patterns kể chuyện, transition bridges, curiosity gap templates, và payoff structures.

---

## Story Arc Patterns

### 3-Act Structure theo từng Style

**Concept-First Arc**:
```
ACT 1 — TENSION (DEFINE):
  "Bạn đang làm X. Một lúc nào đó Y xảy ra. Không có Z, bạn sẽ phải làm A, B, C lặp đi lặp lại.
  Đây là lúc [concept] trở nên cần thiết."

ACT 2 — BUILD (VISUAL + CODE):
  Diagram cho thấy flow → Basic cho thấy concept hoạt động →
  Intermediate cho thấy real-world application → Advanced cho thấy edge cases

ACT 3 — PAYOFF (PITFALLS + RECOMMEND):
  "Bây giờ bạn biết cách dùng đúng. Đây là những gì có thể đi sai."
  "Bạn vừa master điều này. Thế giới rộng hơn bạn nghĩ."
```

**Problem-Centric Arc**:
```
ACT 1 — TENSION (DEFINE):
  Bảng variant → "Tất cả những bài toán này có điểm chung nào?"
  → Core insight dẫn ra pattern

ACT 2 — BUILD (VISUAL trace + CODE):
  Step-by-step trace → Problem 1 → Problem 2 → Problem 3 (leo thang)

ACT 3 — PAYOFF:
  "Khi nào nhận ra bài toán thuộc pattern này?"
  → Pattern recognition table trong Quick Reference
```

---

## Tension Templates

### Kịch bản mở đầu — Concept-First

**Pattern "Bạn vừa..."** (incident-based):
```
"Bạn vừa [hành động thường làm]. [N] giây / phút sau, [điều bất ngờ xảy ra].
[Metric bình thường]. [Metric bình thường]. Chỉ có một thay đổi: [điều duy nhất khác].
Bạn cần quyết định trong [thời gian ngắn]."
```

Ví dụ:
```
"Bạn vừa deploy tính năng mới. 30 giây sau, Slack báo: response time tăng từ 50ms
lên 4s. CPU bình thường. Memory bình thường. Chỉ có một thay đổi: goroutine count
tăng từ 200 lên 18,000. Bạn cần quyết định trong 5 phút."
```

**Pattern "Hàm này trông đúng..."** (code-based):
```
"Hàm này trông hoàn toàn bình thường:

[code snippet ngắn, có vấn đề ẩn]

Nó compile. Nó pass test. Nó chạy trong production suốt 6 tháng.
Cho đến khi [edge case / load condition] xảy ra."
```

**Pattern "Bạn đang làm X, cần Y..."** (task-based):
```
"Bạn cần [mô tả task]. Cách hiển nhiên nhất:

[approach hiển nhiên]

Với N=100, ổn. Với N=10,000, [vấn đề cụ thể — thời gian, memory, etc.].
[Concept] giải quyết bài này bằng cách [1 câu mô tả approach]."
```

### Kịch bản mở đầu — Problem-Centric

**Pattern "Đề cho bạn..."** (competitive programming):
```
"Đề cho bạn [input]. Brute force [approach và complexity].
Với n=[số], [hệ quả cụ thể — TLE, MLE, etc.].

Bạn biết [property của input]. Có thông tin nào bạn chưa khai thác?
→ [Insight dẫn đến pattern]
→ [Insight dẫn đến pattern]

Đây là intuition của [pattern]: [1 câu giải thích]."
```

---

## Curiosity Gap Templates

### Tạo gap

```
Type 1 — Câu hỏi treo:
  "Nhưng tại sao approach B lại tệ hơn trong trường hợp này?
  Câu trả lời liên quan đến cách Go scheduler thực sự hoạt động — và nó khác
  với những gì hầu hết mọi người nghĩ."

Type 2 — Term chưa giải thích:
  "sync.Pool nghe có vẻ là giải pháp hoàn hảo. Nhưng có một catch —
  và Go docs cố tình để nó ở cuối trang."

Type 3 — Partial reveal:
  "Điều này giải thích được 80% trường hợp. 20% còn lại là lý do
  mà production systems thường implement thêm một layer."

Type 4 — Escalation signal:
  "Code này hoạt động tốt với 10 goroutines. Nhưng khi scale lên 10,000,
  có một behavior mà không ai expect."
```

### Đóng gap

```
Pattern "Đây chính là lý do":
  "[Reveal insight]. Đây chính là lý do [behavior đã thấy] xảy ra."

Pattern "Bây giờ thì rõ rồi":
  "Nhìn lại code ban đầu — bây giờ bạn thấy vấn đề ngay. [Explain với kiến thức mới]."

Pattern "Echo về opening":
  "Quay lại tình huống lúc đầu: goroutine count nhảy lên 18,000.
  Bây giờ bạn biết chính xác điều gì đã xảy ra — và cách fix nó trong 3 dòng."
```

---

## Transition Bridge Patterns

### DEFINE → VISUAL

```
Loại "Lý thuyết → Thực tế":
  "Cơ chế nghe đơn giản trên giấy — nhưng timing trong thực tế quyết định tất cả.
  Hãy xem nó diễn ra như thế nào."

Loại "Khái niệm → Hình dung":
  "Bạn vừa hiểu lý thuyết. Hãy hình dung nó trong thực tế trước khi bắt tay implement."

Loại "Abstract → Concrete":
  "Định nghĩa vẫn còn trừu tượng — một diagram sẽ làm rõ tất cả trong 30 giây."
```

### VISUAL → CODE

```
Loại "Flow → Implementation":
  "Diagram cho bạn thấy flow tổng quan. Bây giờ hãy implement từng bước —
  bắt đầu từ version đơn giản nhất."

Loại "Lý thuyết đủ → Thực hành":
  "Bạn đã thấy cách nó hoạt động. Đây là cách bạn build nó."
```

### CODE Basic → Intermediate

```
Loại "Pain point escalation":
  "[Kết luận của Basic — nêu giới hạn]. Trong production, bạn sẽ cần [điều gì đó mạnh hơn]."

Loại "Happy path → Edge case":
  "Version basic hoạt động tốt cho happy path. Khi [condition], cần thêm một layer."

Loại "Single → Concurrent":
  "Single-goroutine version đơn giản và dễ hiểu. Nhưng khi cần xử lý đồng thời..."
```

### CODE → PITFALLS

```
"Bạn vừa implement được [pattern]. Bây giờ là phần quan trọng nhất:
những cách mà ngay cả người biết rồi vẫn dùng sai."

"Code đúng chưa đủ — bạn cần biết code nào trông đúng nhưng sai.
Đây là những bẫy phổ biến nhất:"
```

### PITFALLS → RECOMMEND

```
"Bạn vừa tránh được những bẫy phổ biến nhất. Bây giờ, câu hỏi tiếp theo:
khi nào [concept này] không còn là công cụ phù hợp?"

"[Concept] giải quyết được [vấn đề]. Nhưng nó có điểm mù: [limitation].
Đó là lý do các pattern dưới đây tồn tại:"
```

---

## Voice & Pacing Patterns

### Expert-to-colleague voice examples

```
❌ Paper-style:
   "The implementation necessitates careful consideration of the relationship
   between buffer capacity and concurrent worker utilization."

❌ Chat-style:
   "ok vậy là mình cần buffer bằng số worker nha, nếu nhiều hơn thì hơi tệ"

✅ Expert-to-colleague:
   "Buffer size của jobs channel nên bằng pool size — không nhiều hơn.
   Lớn hơn thì mất back-pressure: sender vẫn push được khi worker bận,
   queue tích lại không kiểm soát."
```

### Pacing — "Breath" sentence

Sau đoạn dày:
```
"Nói ngắn gọn: Go scheduler quyết định khi nào goroutine chạy — không phải bạn."

"Tóm lại: X đảm bảo A, Y đảm bảo B. Điều quan trọng là thứ tự."

"Đơn giản hơn: đây là mutex, nhưng thông minh hơn."
```

### Câu hỏi tu từ — dẫn dắt suy nghĩ

```
"Bạn biết array đã sorted. Có thông tin nào bạn chưa khai thác?"

"Tại sao Mutex không phải lúc nào cũng là câu trả lời đúng?"

"Điều gì xảy ra khi 1,000 goroutine cùng ghi vào một map?"
```

---

## Payoff Engineering

### "Aha Moment" Structure

```
SETUP (đầu DEFINE):
  Đặt ra câu hỏi/vấn đề đủ cụ thể để ám ảnh người đọc

BUILD (giữa bài):
  Cung cấp đủ context, nhưng chưa giải quyết tension chính

REVEAL (cuối DEFINE hoặc đầu CODE):
  Đưa ra insight theo cách "tất nhiên là vậy" — không phải "đây là đáp án"

ECHO (kết luận CODE hoặc RECOMMEND):
  Kết nối lại với pain point ban đầu: "Quay lại tình huống lúc đầu..."
```

### RECOMMEND — Narrative paragraph template

```
"[Tên bài] vừa cho bạn [điều cụ thể đã học được] — [1 câu về giá trị của nó].

Nhưng [limitation của concept vừa học]. Đó là lý do [concept tiếp theo] tồn tại.
Và khi bạn hiểu [concept tiếp theo], bạn sẽ hiểu tại sao [concept cao hơn] ra đời."
```

### PITFALLS — Fatal narrative template

```
"### 🔴 Pitfall #N — [Tên gợi hình, không phải mô tả kỹ thuật]

Code này compile, chạy, và pass test trong 99% trường hợp:

[code snippet]

[Giải thích ngắn tại sao nó trông đúng]

Nhưng khi [specific condition — load cao, timing cụ thể, edge case]:
[Mô tả failure mode — cụ thể, không abstract]

Bug này [đặc điểm — chỉ xuất hiện ở production / không stable reproduce / etc.]

**Fix**: [Cách đúng — 1-2 câu, có lý do]"
```
