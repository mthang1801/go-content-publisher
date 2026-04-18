---
name: technical-doc-writer
description: >
    Skill viết tài liệu kỹ thuật chuyên nghiệp theo hệ thống workflow R1–R7.
    Kích hoạt khi người dùng yêu cầu viết, tạo, hoặc soạn thảo tài liệu kỹ thuật
    về bất kỳ chủ đề nào: Go, NestJS, design patterns, DSA, DevOps, system design,
    database, concurrency, microservices, hoặc bất kỳ technical topic nào khác.
    Cũng kích hoạt khi người dùng nói "viết doc về X", "tạo tài liệu cho Y",
    "document cái này", hoặc yêu cầu giải thích kỹ thuật ở dạng có cấu trúc.
    Skill này đảm bảo output đạt chuẩn: có story arc, voice nhất quán, audience layering,
    code examples đúng nhịp 4 bước, và narrative transitions giữa các section.
---

# Technical Doc Writer Skill

Skill này hướng dẫn agent tạo tài liệu kỹ thuật theo hệ thống workflow 7 rules —
từ triết lý kể chuyện đến cấu trúc chi tiết, từ voice đến pacing, từ code example đến transition.

**Mục tiêu**: Không chỉ là tài liệu — là một cuốn sách biết nói, biết kể chuyện,
khiến người đọc cảm thấy "nghiện", hứng thú.

---

## BƯỚC 0: Trước khi viết bất cứ điều gì — Pre-flight

### 0.1 Xác định topic và context

Thu thập đủ thông tin trước khi bắt đầu:

```
- Topic là gì? (tên concept, pattern, hoặc bài toán cụ thể)
- Folder/category nào? (go/, dsa/, design-pattern/, sql/...)
- Audience chính? (beginner-friendly hay experienced-first?)
- Có gold standard reference chưa? (file nào trong repo đang là ví dụ tốt nhất?)
```

Nếu thiếu thông tin → hỏi trước, không đoán.

### 0.2 Internalize 3 câu hỏi nền tảng

Trả lời 3 câu hỏi này TRƯỚC KHI viết — không phải trong lúc viết:

```
1. Người đọc đang đứng ở đâu khi bắt đầu bài này?
   → Họ biết gì, không biết gì, đang gặp vấn đề gì trong thực tế

2. Tôi muốn họ cảm thấy gì ở từng điểm trong bài?
   → Tò mò → khó chịu với vấn đề → nhẹ nhõm khi hiểu → tự tin khi apply

3. Họ rời bài này với gì?
   → Không chỉ "biết thêm" — mà "hiểu sâu hơn" hoặc "có thể làm điều trước đây không làm được"
```

### 0.3 Lên kế hoạch Story Arc

Trước khi viết, phác thảo arc trong 3 dòng:

```
TENSION (DEFINE):  [Pain point cụ thể — người đọc cảm thấy vấn đề gì?]
BUILD (CODE):      [Từ basic đến advanced — mỗi example giải quyết tension nào, tạo tension mới nào?]
PAYOFF (PITFALLS + RECOMMEND): [Người đọc kết thúc với insight gì mà họ không có lúc đầu?]
```

---

## BƯỚC 1: Chọn Style và Voice (R4)

### Style decision

```
Concept-First → Design Patterns, Go stdlib, SQL, system concepts, concurrency
Problem-Centric → DSA, LeetCode, OOD interview, bit manipulation, career

Folder mới chưa có trong mapping:
  → Concept nếu topic là "X là gì / hoạt động thế nào"
  → Problem-Centric nếu topic là "giải bài toán X / implement Y"
```

→ Chi tiết mapping đầy đủ: `references/rules-summary.md` § R4

### Voice declaration

Chọn 1 trong 2 và giữ xuyên suốt:

- **"bạn"** — gần gũi, direct: "Khi bạn gặp vấn đề này..."
- **"ta"** — collegial: "Khi ta implement pattern này..."

Voice tone: **expert-to-colleague** — không phải paper, không phải chat.
Có quan điểm, đồng cảm, trực tiếp. Không neutral vô vị.

---

## BƯỚC 2: Scaffold Section Order và Transition Plan (R1)

### Section order

```
[0. TEMPLATE]  ← Optional: chỉ khi có boilerplate người đọc cần thường xuyên
 1. DEFINE     ← Tension setup: kịch bản → pain point → concept
 2. VISUAL     ← Build: diagram tối thiểu 2 level
 3. CODE       ← Build: Basic → Intermediate → Advanced → Expert
 4. PITFALLS   ← Payoff: câu chuyện cảnh báo, không phải checklist
 5. REF        ← Nguồn verify — không phải "bài liên quan"
 6. RECOMMEND  ← Payoff: cánh cửa mở ra horizon mới
[7. QUICK REF] ← Optional: cheatsheet nếu topic có nhiều API cần tra cứu
```

### Transition plan — lên kế hoạch NGAY TẠI ĐÂY

Viết bridge sentence cho 4 transition bắt buộc:

| Transition                | Bridge sentence (draft) |
| ------------------------- | ----------------------- |
| DEFINE → VISUAL           |                         |
| VISUAL → CODE             |                         |
| CODE Basic → Intermediate |                         |
| PITFALLS → RECOMMEND      |                         |

Quy tắc vận hành:

- Bridge phải trả lời: `Vì sao section tiếp theo phải đến ngay bây giờ?`
- Chọn đúng nhịp chuyển: `zoom in`, `contrast`, `escalate`, hoặc `payoff`
- Không copy nguyên mẫu bridge giữa hai file khác nhau; mỗi bridge phải bám đúng tension vừa mở ra trong bài đó
- Nếu bỏ bridge đi mà section sau đọc như bị "teleport", draft hiện tại chưa đủ lực

Mini cheat sheet:

- `DEFINE → VISUAL`: `Khái niệm đã rõ trên giấy. Phần còn dễ hiểu sai nhất nằm ở cách nó vận hành trong thực tế.`
- `VISUAL → CODE`: `Flow đã hiện ra. Giờ ta hạ nó xuống artifact mà team có thể review, debug hoặc áp dụng ngay.`
- `CODE Basic → Intermediate`: `Basic case đã chạy được. Production bắt đầu khó từ đúng chỗ ví dụ vừa bỏ qua.`
- `CODE → PITFALLS`: `Biết cách làm đúng mới chỉ là một nửa câu chuyện; phần còn lại là những chỗ rất dễ làm gần đúng rồi vẫn hỏng.`
- `PITFALLS → RECOMMEND`: `Khi đã thấy term này thường gãy ở đâu, bước kế là mở đúng concept lân cận để tránh sửa sai lớp vấn đề.`

→ Bộ mẫu đầy đủ 20 bridge sentence: `../../workflows/rules/r7-narrative-architecture.md` § Transition / cheat sheet

---

## BƯỚC 3: Chọn Diagram Type (R6)

Cho mỗi diagram cần vẽ, trả lời: **"Diagram này trả lời câu hỏi nào?"**

```
Quick decision:
  DSA/algorithm    → ASCII trace bắt buộc + optional Mermaid
  Design Pattern   → Mermaid class diagram hoặc ASCII box-and-arrow
  Concurrency      → ASCII timeline (goroutine A | goroutine B)
  Request flow     → Mermaid sequence diagram
  State machine    → Mermaid stateDiagram
  Data structure   → ASCII tree/grid

ASCII khi: ≤ 5 node, step-by-step trace, inline code comment
Mermaid khi: class diagram, sequence ≥ 4 actor, complex flowchart

Width limit: ASCII ≤ 80 chars
```

Mỗi diagram PHẢI có caption 1 dòng italic ngay bên dưới.
Caption tốt = người đọc hiểu insight chính ngay cả khi không nhìn diagram.

---

## BƯỚC 4: Viết Section CODE (R3)

### Nhịp 4 bước cho mỗi example

**Bước 1 — INTRODUCE (câu chuyện dẫn vào):**

- Bối cảnh: đang làm gì, vấn đề xuất hiện như thế nào
- Pain point đủ cụ thể để gây khó chịu
- Approach được chọn + lý do thay vì cách khác
- Input/Output: `f([1,2,3], 2) → 4`

**Bước 2 — CODE (annotated comments = reasoning):**

- Comments giải thích _tại sao_, không phải _cái gì_
- `// ⚠️` cho warning không hiển nhiên
- `// ✅` cho best practice cần chú ý
- Tối thiểu 30-50 dòng có ý nghĩa

**Bước 3 — "TẠI SAO?" (mở hộp đen):**

| Complexity   | Rule                                              |
| ------------ | ------------------------------------------------- |
| Basic        | Optional — chỉ thêm nếu behavior không hiển nhiên |
| Intermediate | **BẮT BUỘC**                                      |
| Advanced     | **BẮT BUỘC**                                      |
| Expert       | **BẮT BUỘC**                                      |

"Tại sao?" giải thích **cơ chế nội tại** — không phải mô tả output.

**Bước 4 — KẾT LUẬN (đóng vòng câu chuyện):**

- Đã đạt được gì (1 dòng)
- Caveat / giới hạn (khi nào KHÔNG dùng)
- Signal để nhận ra khi nào nên dùng pattern này

### Multi-language tabs

Thứ tự: **Go → TypeScript → Java → Rust → C++ → Python**

Java chỉ thêm khi: OOP patterns rõ ràng, Design Patterns, Data Structures, OOD Interview.
Không thêm Java cho: Go-specific concurrency, goroutine/channel, Go stdlib.

Blocks PHẢI liền nhau — không chèn text giữa các block (tab sẽ vỡ).

Excluded folders (single language): `go/`, `drizzle/`, `docker/`, `k8s/`, `sql/`, `diagram/`, `game-development/`

→ Chi tiết đầy đủ: `references/rules-summary.md` § R3

---

## BƯỚC 5: Scaffold File (R5)

Header chuẩn:

```markdown
<!-- tags: topic, subtopic -->

# 🔄 [Tên chủ đề]

> Mô tả ngắn 1-2 câu — giải thích file này về cái gì và tại sao cần đọc.

📅 Ngày tạo: YYYY-MM-DD · 🔄 Cập nhật: YYYY-MM-DD · ⏱️ X phút đọc

| Aspect         | Detail                   |
| -------------- | ------------------------ |
| **Complexity** | O(?) / O(? log ?) / ...  |
| **Use case**   | Khi nào dùng pattern này |
| **Go stdlib**  | Package liên quan        |
```

Thời gian đọc: `(số từ prose / 200) + (số dòng code / 400)`, làm tròn lên.

DEFINE template với audience markers:

```markdown
## 1. DEFINE

_(Prerequisite: [tên] → [link])_ ← xóa nếu không cần

<!-- [Beginner layer] -->

[Kịch bản mở đầu — quyết định maker, pain point cụ thể]

<!-- [Experienced layer] -->

[Định nghĩa formal — sau kịch bản]

<!-- [Expert layer] -->

### 1.x Invariants & Failure Modes

[Edge cases, ràng buộc bất biến]
```

Xóa audience markers trước khi commit.

---

## BƯỚC 6: Viết nội dung — Trong khi viết

### DEFINE — Vivid Scenario Opening + Tension setup

> **Rule**: DEFINE PHẢI mở bằng **vivid scenario** — tình huống cụ thể, có thể hình dung được, khiến người đọc gật đầu "đúng, tôi đã gặp cái này." Không phải định nghĩa, không phải lịch sử — là **khoảnh khắc thật** mà concept đang giải quyết.

**Cấu trúc vivid scenario**:

1. Đặt người đọc vào tình huống cụ thể — dùng "bạn", "hình dung"
2. Mô tả pain point bằng chi tiết kỹ thuật chính xác — không chung chung
3. Đưa ra consequence — điều gì xảy ra nếu không giải quyết
4. Kết nối concept — concept xuất hiện như "lời giải cho đúng bài toán này"

```
❌ Mở bằng định nghĩa:

   "Defer là một keyword trong Go cho phép bạn lên lịch thực thi..."

✅ Mở bằng vivid scenario:
   "Hình dung bạn đang viết một hàm mở file, kết nối database, tạo lock —
   rồi có đến 5 điểm return khác nhau bên trong. Bỏ sót một chỗ Close()
   là resource leak, đôi khi là deadlock.
   Đó là bài toán defer sinh ra để giải quyết."
```

Sau vivid scenario → kịch bản mở đầu PHẢI pass 4 self-check:

```
[ ] Người đọc vào vị trí ra quyết định (không phải observer)?
[ ] Pain point cụ thể, cảm nhận được sự khó chịu?
[ ] Concept là lời giải tự nhiên, không phải arbitrary?
[ ] Không thể dùng kịch bản này cho topic khác?
```

Sau kịch bản → định nghĩa formal → audience layering:

- Paragraph 1-2: beginner accessible
- Paragraph 3-4: experienced (formal + phân biệt)
- Sub-section Invariants/Failure Modes: expert territory

### VISUAL — Build layer 1

Không vẽ "minh họa tổng quát". Mỗi diagram = 1 câu hỏi cụ thể.

Level 1 (Simple) BẮT BUỘC → Level 2 (Detailed) BẮT BUỘC nếu có Intermediate code.

### CODE — Build layer 2 (escalating)

Examples KHÔNG phải 3 bài độc lập — chúng là chapters của một câu chuyện leo thang.

Kết luận của example N dẫn vào tension cho example N+1.

### PITFALLS — Payoff layer 1

Fatal pitfall = narrative block (câu chuyện cảnh báo) + bảng row.
Common/Minor = bảng row là đủ.

PITFALLS mở đầu bằng bridge: "Bây giờ bạn biết cách dùng đúng — đây là những cách dùng sai mà ngay cả người biết rồi vẫn rơi vào."

### RECOMMEND — Payoff layer 2

Mở đầu bằng narrative paragraph:

1. Tổng kết 1 dòng những gì người đọc vừa đạt được
2. Mở curiosity gap mới → dẫn vào bảng

Bảng: `Mở rộng | Khi nào | Lý do | File/Link`

---

## BƯỚC 7: Addiction Loop — Kiểm tra trước khi viết từng section

Tự hỏi sau mỗi section:

- Tôi vừa tạo ra curiosity gap nào?
- Curiosity gap từ section trước đã được đóng chưa?
- Người đọc có lý do để đọc tiếp không?

Kỹ thuật tạo gap:

```
- Đặt câu hỏi cuối đoạn chưa trả lời ngay
- Giới thiệu term chưa giải thích ngay
- "Nhưng điều này dẫn đến một vấn đề khác..."
- "...nhưng có một catch mà hầu hết mọi người bỏ qua"
```

---

## BƯỚC 8: QA Checklist trước khi hoàn thành

Chạy theo thứ tự: **Narrative QA → Content QA → File QA**

### Narrative QA (R7)

```
STORY ARC
[ ] DEFINE mở bằng vivid scenario — người đọc thấy mình trong tình huống đó?
[ ] DEFINE có tension rõ — người đọc cảm thấy vấn đề trước khi thấy solution?
[ ] CODE examples escalate từng bước — không phải 3 bài độc lập?
[ ] PITFALLS có câu chuyện ở Fatal pitfalls, không chỉ bảng?
[ ] RECOMMEND mở ra horizon mới, không chỉ list link?

VOICE & PACING
[ ] Voice nhất quán từ đầu đến cuối?
[ ] Expert-to-colleague tone — có quan điểm, không neutral vô vị?
[ ] Điều hiển nhiên không bị giải thích dư thừa?
[ ] Concept phức tạp được giải thích đủ chậm?

TRANSITION
[ ] Có bridge sentence giữa DEFINE→VISUAL, VISUAL→CODE, CODE→PITFALLS, PITFALLS→RECOMMEND?
[ ] Mỗi CODE example dẫn vào example tiếp theo?
[ ] Bridge có bám đúng tension của file hiện tại, hay chỉ là câu có thể copy sang bài khác?

ADDICTION LOOP
[ ] Có ít nhất 1 curiosity gap được tạo ra và đóng lại?
[ ] "Aha moment" được dẫn dắt (setup → build → reveal → echo)?
```

### Content QA (R2)

```
DEFINE
[ ] Mở bằng vivid scenario — tình huống cụ thể, relatable, không phải định nghĩa?
[ ] Kịch bản đặt người đọc vào vị trí ra quyết định?
[ ] Định nghĩa formal có mặt SAU kịch bản?
[ ] Audience layering đúng thứ tự: beginner → experienced → expert?

VISUAL
[ ] Mỗi diagram trả lời đúng 1 câu hỏi?
[ ] Có caption 1 dòng dưới mỗi diagram?
[ ] Mermaid đã test render tại mermaid.live?

CODE
[ ] Thứ tự complexity đúng: Basic → Intermediate → Advanced → Expert?
[ ] Intermediate+ có "Tại sao?" block giải thích cơ chế?
[ ] Code chạy được?
[ ] Multi-language tabs liền nhau?

PITFALLS
[ ] ≥ 3 pitfalls, có cột Severity, Fatal xếp trước?
[ ] Fatal pitfalls có narrative block?

LINKS
[ ] Prerequisite link đầu DEFINE?
[ ] Expansion link trong RECOMMEND?
[ ] Không link đến file chưa tồn tại?
```

---

## Tham chiếu nhanh

Đọc thêm khi cần chi tiết:

| Cần gì                                        | Đọc file                                         |
| --------------------------------------------- | ------------------------------------------------ |
| Story arc, addiction loop, payoff engineering | `references/narrative-patterns.md`               |
| Bridge sentence patterns cho transitions      | `references/narrative-patterns.md` § Transitions |
| Audience layering rules đầy đủ                | `references/rules-summary.md` § R1               |
| Mapping folder → style + multi-lang tabs      | `references/rules-summary.md` § R3, R4           |
| ASCII/Mermaid format chuẩn                    | `references/rules-summary.md` § R6               |
| File naming, version bump, sync taxonomy      | `references/rules-summary.md` § R2               |
| Ví dụ ❌/✅ cho từng rule                     | `references/examples.md`                         |
