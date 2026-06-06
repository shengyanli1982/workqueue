# pprof-Driven Performance Optimization Plan

> **For agentic workers:** REQUIRED: Use `plan-runbook-execute` for development execution and add `review-spec-implementation` as the post-implementation review gate.

**Goal:** Iteratively optimize workqueue hot paths based on pprof CPU/memory profile data until diminishing returns.

**Architecture:** Small-step optimizations targeting the top-5 CPU hotspots identified via pprof; each round followed by benchmark comparison, correctness verification, and re-profiling.

**Tech Stack:** Go 1.19+, pprof, testing/benchmark

---

## Baseline Data (2026-05-24, i5-12400F, Windows)

| Benchmark | ns/op | B/op | Allocs | Primary Hotspot (pprof) |
|---|---|---|---|---|
| Queue_Put | 86 | 72 | 2 | sync.Pool.Get 38.8% cum |
| Queue_Get | 33 | 22 | 0 | - |
| Queue_PutAndGet | 45 | 8 | 0 | mutex.Lock+Unlock 39.6% cum |
| Idempotent_Put | 425 | 151 | 2 | Set.Add 30.1%, mapassign 28.5% |
| Idempotent_Get | 492 | 112 | 0 | Set.Add 49.6%, map.Delete 18.6% |
| Idempotent_PutGetDone | 177 | 8 | 1 | balanced |
| PriorityQueue_Put | 169 | 71 | 1 | RBTree insert |
| DelayingQueue_PutWithDelay | 232 | 72 | 1 | toDelay + RBTree push |
| DeadLetter_PutGetAck | 247 | 464 | 5 | gcDrain 56.8% |

---

## Round 1: Replace sync.Pool with Mutex-Protected Free List

**Rationale:** `sync.Pool.Get` → `getSlow` → `poolChain.popTail` costs 520ms+500ms cumulative (38.81% of basic Queue Put). For a tightly scoped hot path with predictable allocation patterns, a simple mutex-protected free list eliminates sync.Pool's complex pool management machinery.

**Files:**
- Modify: `internal/container/list/node.go:46-73`

- [ ] **Step 1: Implement free list NodePool**

Replace `sync.Pool` with `sync.Mutex` + `[]*Node` free list with max capacity.

- [ ] **Step 2: Run ALL tests to verify correctness**
Run: `go test -count=1 ./...`
Expected: ALL PASS

- [ ] **Step 3: Run targeted benchmarks**
Run: `go test -bench='BenchmarkQueue_Put$|BenchmarkQueue_Get$|BenchmarkQueue_PutAndGet$' -benchmem -benchtime=3s -count=1 -run='^$' .`

- [ ] **Step 4: Analyze improvement delta**

---

## Round 2: Direct Branching in RBTree Insert

**Rationale:** `ternary.If()` function call in the RBTree insert loop adds call overhead and prevents branch prediction optimization. Direct `if/else` can be inlined and branch-optimized by the compiler.

**Files:**
- Modify: `internal/container/heap/heap.go:109-148` (insert function)

- [ ] **Step 1: Replace ternary.If with direct if/else**
- [ ] **Step 2: Run tests**  
- [ ] **Step 3: Run PriorityQueue benchmarks**

---

## Round 3: Reduce Node.Reset Overhead

**Rationale:** Node.Reset() sets 8 fields including parentRef (unsafe.Pointer). On the Put-back-to-pool path, all fields are zeroed. We can reduce this by only clearing pointer fields (for GC safety) while leaving value fields as-is (they'll be overwritten on next Get).

**Files:**
- Modify: `internal/container/list/node.go:34-42`

- [ ] **Step 1: Minimize Reset to pointer fields only**
- [ ] **Step 2: Run tests**
- [ ] **Step 3: Run queue benchmarks**

---

## Round 4: Pre-allocate Set Map Capacity

**Rationale:** `mapassign` dominates Idempotent_Put (28.5% cum). Pre-allocating the map with an initial capacity avoids repeated grow operations during early usage.

**Files:**
- Modify: `internal/container/set/set.go:10-15`
- Modify: `config.go:31-36` (add default size constant)

- [ ] **Step 1: Add capacity-aware Set constructor**
- [ ] **Step 2: Run tests**
- [ ] **Step 3: Run Idempotent benchmarks**

---

## Round 5: Optimize DeadLetter Allocation (5 allocs → target 3)

**Rationale:** DeadLetter_PutGetAck has 5 allocs/op and 464 B/op, with gcDrain consuming 56.77% of CPU. The Meta map allocation and DeadLetter struct are primary sources.

**Files:**
- Review: `dead_letter_queue.go`
- Review: `interface.go:67-75`

---

## Exit Criteria

- Stop when per-round improvement < 5% relative to previous round
- All existing tests must pass after each round
- Each round must have before/after benchmark comparison
