# Examples Reference — ❌/✅ cho từng rule

---

## DEFINE Opening

### Kịch bản mở đầu

❌ **Observer (không có tension)**:
```
"Hãy tưởng tượng có một hệ thống e-commerce đang xử lý hàng ngàn đơn hàng mỗi ngày.
Hệ thống cần đảm bảo concurrent access an toàn..."
```

✅ **Decision maker (có tension)**:
```
"Bạn vừa deploy tính năng flash sale. 30 giây sau, Slack báo:
response time nhảy từ 40ms lên 3.2s. CPU 20%. Memory bình thường.
Chỉ có một điểm khác biệt so với tuần trước: 50x traffic.
Counter số lượng hàng đang bị oversell — khách mua được hàng đã hết."
```

❌ **Pain point mơ hồ**:
```
"Nếu không xử lý đúng, hệ thống sẽ gặp vấn đề về hiệu năng và tính nhất quán..."
```

✅ **Pain point cụ thể + concept là lời giải tự nhiên**:
```
"Hai goroutine đọc stock=1, cùng check 'còn hàng', cùng trừ về 0, cùng confirm order.
Stock là -1. Hai khách đều nhận email xác nhận. Kho chỉ có 1 sản phẩm.
Mutex giải quyết bài này: chỉ 1 goroutine được vào critical section tại một thời điểm."
```

---

## CODE — Annotated Comments

❌ **Comment lặp lại statement**:
```go
wg.Add(1)        // add 1 to waitgroup
go worker(jobs)  // start goroutine
wg.Wait()        // wait for completion
```

✅ **Comment giải thích reasoning**:
```go
// wg.Add() phải gọi TRƯỚC go func() — nếu gọi bên trong goroutine,
// scheduler có thể chạy wg.Wait() trước khi Add() kịp thực thi → race condition.
wg.Add(1)
go worker(jobs, &wg)

// Wait() block main goroutine đến khi tất cả worker gọi Done().
// Bỏ dòng này → main exit sớm, kill toàn bộ goroutine đang chạy → data loss.
wg.Wait()
```

---

## CODE — Pre-context

❌ **Liệt kê kỹ thuật, không có bối cảnh**:
```markdown
> **Mục tiêu**: Implement Worker Pool với 5 goroutines.
> **Input**: jobs channel. **Output**: results channel.
```

✅ **Câu chuyện dẫn vào**:
```markdown
Bạn có 50,000 ảnh cần resize. Spawn 50,000 goroutine cùng lúc: memory spike,
OS file descriptor exhausted sau 2 giây. Xử lý tuần tự: mất 45 phút.

Worker Pool giải quyết: giữ đúng N goroutine sống, mỗi goroutine tự lấy job tiếp
khi xong — throughput ổn định, resource có giới hạn.

Input: `[]string{"img1.jpg", "img2.jpg", ...}`, N=10 workers
Output: `[]Result` theo thứ tự hoàn thành (không đảm bảo thứ tự gốc)
```

---

## CODE — "Tại sao?" block

❌ **Mô tả output, không phải cơ chế**:
```markdown
> **Tại sao?** Đoạn code trên xử lý được 50,000 ảnh mà không bị OOM.
```

✅ **Giải thích cơ chế nội tại**:
```markdown
> **Tại sao Worker Pool tránh được memory spike?**
> Channel `jobs` buffered với capacity = pool size. Khi tất cả worker bận,
> sender block tại `jobs <- task` — không allocate thêm goroutine, không mở
> thêm file descriptor. Back-pressure được tạo ra từ channel semantics,
> không cần rate limiter bên ngoài hay semaphore.
```

---

## CODE — Kết luận

❌ **Kết luận rỗng**:
```markdown
> **Kết luận**: Đây là cách implement Worker Pool trong Go.
```

✅ **Kết luận đóng vòng câu chuyện**:
```markdown
> **Kết luận**: Worker Pool với buffered channel kiểm soát concurrency mà không cần
> external rate limiter — Go channel semantics tự tạo back-pressure.
>
> **Caveat**: Pool size cố định không tự điều chỉnh theo load. Job có thời gian
> xử lý rất không đồng đều (1ms vs 30s) → một số worker idle trong khi queue đầy.
>
> **Dùng khi**: batch processing với tập lớn, cần giới hạn resource —
> image processors, crawlers, report generators.
> Cần dynamic scaling → xem [semaphore pattern](../patterns/semaphore.md).
```

---

## Transition Bridges

❌ **Không có bridge — teleport**:
```markdown
[...định nghĩa Mutex xong...]

## 2. VISUAL

[diagram xuất hiện đột ngột, không có context]
```

✅ **Có bridge**:
```markdown
[...định nghĩa Mutex xong...]

Cơ chế Lock/Unlock đơn giản trên giấy — nhưng timing trong thực tế
quyết định tất cả. Hãy xem nó diễn ra như thế nào.

## 2. VISUAL
```

❌ **CODE examples rời rạc**:
```markdown
### Ví dụ 1: Basic — Goroutine đơn giản
[code]

### Ví dụ 2: Intermediate — WaitGroup
[code đột ngột phức tạp hơn]
```

✅ **CODE examples leo thang có cầu nối**:
```markdown
### Ví dụ 1: Basic — Goroutine đơn giản
[code]

> **Kết luận**: Goroutine chạy được. Nhưng main() exit trước khi goroutine
> kịp in kết quả — đang race với scheduler. Cần một cách để đợi goroutine hoàn thành.

### Ví dụ 2: Intermediate — WaitGroup: đồng bộ hóa có kiểm soát
```

---

## PITFALLS

❌ **Checklist rỗng**:
```markdown
| 1 | Gọi wg.Add() trong goroutine | Race condition | Gọi trước go func() |
```

✅ **Narrative pitfall (Fatal)**:
```markdown
### 🔴 Pitfall #1 — Cái bẫy "trông đúng nhưng sai"

Code này compile, chạy, pass unit test trong môi trường dev:

```go
go func() {
    wg.Add(1) // trông hợp lý: goroutine tự declare mình
    defer wg.Done()
    process(job)
}()
wg.Wait()
```

Scheduler có thể execute wg.Wait() TRƯỚC khi goroutine kịp gọi wg.Add(1).
Wait() thấy counter = 0, return. Goroutine vẫn đang chạy, process() vẫn đang
ghi vào shared state. Main goroutine tiếp tục — race condition trên production.

Bug này không reproducible trong test vì scheduling không deterministic ở low load.
Chỉ xuất hiện khi system busy, là loại bug tệ nhất: intermittent, production-only.

**Fix**: `wg.Add(1)` gọi TRƯỚC `go func()`, trong calling goroutine, không trong goroutine con.
```

---

## RECOMMEND

❌ **Danh sách link**:
```markdown
## 6. RECOMMEND

| sync.RWMutex | Read-heavy | Concurrent reads | [link] |
| sync.Map | Concurrent map | No manual lock | [link] |
```

✅ **Narrative + table**:
```markdown
## 6. RECOMMEND

Bạn vừa nắm được Mutex — cơ chế đảm bảo mutual exclusion đơn giản và đáng tin cậy nhất.
Với Mutex, bạn protect được bất kỳ shared state nào trong concurrent Go code.

Nhưng Mutex có điểm mù: nó không phân biệt reader và writer.
100 reader đồng thời vẫn phải chờ nhau — dù họ không conflict với nhau.
Đó là lý do `sync.RWMutex` tồn tại. Và khi concurrent map trở thành pattern
phổ biến, `sync.Map` giải quyết cả lock lẫn boilerplate trong một bước.

| Mở rộng | Khi nào | Lý do | File/Link |
|---------|---------|-------|-----------|
| sync.RWMutex | Read >> Write ratio | Concurrent reads OK — Mutex block không cần | [08-rwmutex.md] |
| sync.Map | Concurrent map phổ biến | Built-in, optimized high contention | [09-sync-map.md] |
```

---

## Voice

❌ **Neutral, không có quan điểm**:
```
"Cả hai approach đều có ưu và nhược điểm tùy vào use case của từng dự án."
```

✅ **Có quan điểm, direct**:
```
"Approach B phức tạp hơn 3 lần nhưng chỉ nhanh hơn 8% trong benchmark.
Đừng dùng B trừ khi bạn đã profile và 8% đó thực sự quan trọng."
```

❌ **Audience collapse — 1 giọng cho tất cả**:
```
"Goroutine là lightweight thread được Go runtime quản lý.
Dùng `go` keyword để spawn. WaitGroup để sync."
```

✅ **Layered — phục vụ cả 3 nhóm**:
```
[Beginner]: "Bạn cần xử lý 10,000 request đồng thời. Thread OS tốn ~1MB stack mỗi cái.
10,000 threads = 10GB RAM chỉ cho stack. Không khả thi.
Go giải quyết bài này với goroutine — cùng concurrency, nhưng chỉ ~2KB mỗi cái."

[Experienced]: "Goroutine là coroutine được Go runtime schedule theo mô hình M:N —
M goroutines chạy trên N OS threads (N = GOMAXPROCS, mặc định = số CPU core)."

[Expert]: "Implication khi deploy trên container: GOMAXPROCS đọc CPU quota từ
/proc/cpuinfo, không phải container limit. Dùng `uber-go/automaxprocs` để fix."
```

---

## Curiosity Gap

❌ **Reveal ngay, không có gap**:
```
"sync.Pool giúp tái sử dụng object, giảm GC pressure.
Object không dùng được giữ lại giữa GC cycles."
```

✅ **Tạo gap trước, reveal sau**:
```
"sync.Pool nghe có vẻ là silver bullet cho GC pressure.
Nhưng có một catch mà Go docs để cuối trang — và nó ảnh hưởng đến
cách bạn design API của Pool.

Trước khi reveal catch đó, hãy hiểu tại sao Pool hoạt động được: [code]

Bây giờ về cái catch: sync.Pool không đảm bảo object còn đó sau GC cycle.
Runtime có thể evict bất kỳ lúc nào. Điều này có nghĩa Pool chỉ safe
cho stateless buffers — không safe cho object có lifecycle hoặc state."
```
