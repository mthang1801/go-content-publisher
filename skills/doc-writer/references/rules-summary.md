# Rules Summary — Quick Reference

Condensed version của R1–R6 để tra cứu trong lúc viết.
Đọc full rules khi cần chi tiết: `.agents/workflows/rules/`

---

## R1 — Section Order & Audience Layer

### Section order
```
[0. TEMPLATE] 1. DEFINE → 2. VISUAL → 3. CODE → 4. PITFALLS → 5. REF → 6. RECOMMEND [7. QUICK REF]
```

### Audience layer per section

| Section | Beginner | Experienced | Expert |
|---------|----------|-------------|--------|
| DEFINE | Kịch bản, pain point | Formal def, phân biệt | Invariants, Failure Modes |
| VISUAL | Level 1 simple | Level 2 detailed | Level 3 multi-layer |
| CODE | Basic | Intermediate | Advanced + Expert |
| PITFALLS | Syntax/conceptual | Logic/race condition | Performance, subtle behavior |

### Cross-linking rules

| Link type | Đặt ở | Mục đích |
|-----------|-------|---------|
| Prerequisite | Đầu DEFINE | "Chưa biết X → đọc Y trước" |
| Comparison | Inline DEFINE | "Khác với [X], [concept] không..." |
| Expansion | RECOMMEND | "Bước tiếp theo" |
| Reference | REF | "Nguồn verify nội dung" |

---

## R2 — File Output & Content QA

### Metadata block
```
📅 Ngày tạo: YYYY-MM-DD · 🔄 Cập nhật: YYYY-MM-DD · ⏱️ X phút đọc
```

Thời gian: `(prose words / 200) + (code lines / 400)`, làm tròn lên.

### File naming
`XX-ten-chu-de.md` → `01-bubble-sort.md`, `03-defer-panic-recover.md`

### Navigation cuối file
```
**Liên kết**: [← Previous](./XX-prev.md) · [→ Next](./XX-next.md)
```

### Version bump
Mọi commit → bump `package.json` version theo SemVer.
Chưa bump → không được commit.

---

## R3 — CODE Section

### "Tại sao?" threshold

| Level | Rule |
|-------|------|
| Basic | Optional — chỉ nếu behavior không hiển nhiên |
| Intermediate | BẮT BUỘC |
| Advanced | BẮT BUỘC |
| Expert | BẮT BUỘC |

### Basic — "behavior không hiển nhiên" checklist
```
[ ] Value vs pointer semantics
[ ] Evaluation order (defer, init, goroutine schedule)
[ ] Short-circuit behavior
[ ] Zero value (nil slice vs empty slice)
[ ] Implicit type conversion
[ ] Code có comment "// ⚠️"
```

### Multi-language order
**Go → TypeScript → Java → Rust → C++ → Python**

Java: thêm khi OOP patterns, Design Patterns, Data Structures.
Không thêm Java: Go-specific concurrency, goroutine/channel.

### Excluded (single language only)
`go/`, `drizzle/`, `docker/`, `k8s/`, `elk/`, `linux-command/`, `sql/`, `diagram/`, `glosaries/`, `game-development/`

Tất cả folder khác (`dsa/`, `leet-codes/`, `design-pattern/`, `ood-interview/`, `bit-manipulation/`...) → multi-language bình thường.

### Tab rule
Blocks PHẢI liền nhau — không chèn text giữa.
Dòng đầu mỗi block: `// filename.ext — Pattern Name: Description`

---

## R4 — Writing Style

### Folder → Style mapping

| Folder | Style |
|--------|-------|
| `design-pattern/`, `go/concurrency/`, `go/patterns/`, `go/fundamental/`, `sql/`, `system-design/`, `diagram/`, `glosaries/` | Concept-First |
| `dsa/`, `leet-codes/`, `bit-manipulation/`, `ood-interview/`, `career/` | Problem-Centric |

Default: Concept-First nếu "X là gì", Problem-Centric nếu "giải bài toán X".

### Problem-Centric structure
```
DEFINE: Variant table → Approach comparison → Core insight
VISUAL: ASCII trace (step-by-step), không phải abstract diagram
CODE: ### Problem N: [Name] [LC #ID]
PITFALLS: 5 cột với Severity
RECOMMEND: "Bài liên quan"
QUICK REF: Template + Pattern recognition
```

---

## R5 — Header Template

### Aspect table
```markdown
| Aspect         | Detail                  |
| -------------- | ----------------------- |
| **Complexity** | O(?) / O(? log ?) / ... |
| **Use case**   | Khi nào dùng            |
| **Go stdlib**  | Package liên quan        |
```

### PITFALLS format
```markdown
| # | Severity | Lỗi | Hậu quả | Fix |
|---|----------|-----|---------|-----|
| 1 | 🔴 Fatal  | ... | ... | ... |
| 2 | 🟡 Common | ... | ... | ... |
| 3 | 🔵 Minor  | ... | ... | ... |
```

### REF format
```markdown
| Resource | Loại | Link | Ghi chú |
| -------- | ---- | ---- | ------- |
| ... | Official | ... | ... |
```

### RECOMMEND format
```markdown
| Mở rộng | Khi nào | Lý do | File/Link |
| ------- | ------- | ----- | --------- |
| ... | ... | ... | ... |
```

---

## R6 — VISUAL Section

### Diagram type decision

| Topic | Diagram type |
|-------|-------------|
| DSA/algorithm | ASCII trace bắt buộc |
| Design Pattern | Mermaid class hoặc ASCII box-and-arrow |
| Concurrency | ASCII timeline |
| Request flow | Mermaid sequence |
| State machine | Mermaid stateDiagram |
| Data structure | ASCII tree/grid/heap |
| Infrastructure | ASCII topology |
| Database | Mermaid erDiagram hoặc ASCII table |

### ASCII vs Mermaid

| Khi nào | Dùng |
|---------|------|
| ≤ 5 node, step-by-step trace | ASCII |
| Class diagram, sequence ≥ 4 actor | Mermaid |
| Inline code comment | ASCII only |

### ASCII ký hiệu chuẩn
```
──► →      Flow một chiều
◄──► ↔     Bidirectional
┌─┐ └─┘    Box thường
╔═╗ ╚═╝    Box critical path
⏳ ✅ ❌ ⚠️  Status indicators
```

### Mermaid rules cứng
- Node label ≤ 30 ký tự
- Escape special chars: `()`, `[]`, `{}`
- Test tại mermaid.live trước commit
- Sequence > 6 participant → split 2 diagram

### Caption format
```markdown
*Hình: [Insight cụ thể — consequence nếu có]. Không phải "Hình: Minh họa X."*
```

### Complexity levels
- Level 1 (Simple): BẮT BUỘC — happy path, beginner accessible
- Level 2 (Detailed): BẮT BUỘC nếu bài có Intermediate code
- Level 3 (Multi-layer): Optional
