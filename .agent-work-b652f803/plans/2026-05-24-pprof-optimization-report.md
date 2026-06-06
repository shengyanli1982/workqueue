# pprof-Driven Performance Optimization Report

**Date:** 2026-05-24
**Machine:** 12th Gen Intel Core i5-12400F, Windows, Go 1.24
**Project:** github.com/shengyanli1982/workqueue/v2

---

## Executive Summary

Using pprof CPU and memory profiling across 25 benchmarks, we identified and optimized 5 high-value hotspots in the workqueue library. After 4 optimization rounds with iterative re-profiling, the optimization has converged (per-round improvement < 5%).

### Top Improvements (pprof benchmarks, stable 2x runs)

| Benchmark | Before | After | Improvement |
|---|---|---|---|
| **Idempotent_Get** | 561 ns | **37 ns** | **-93.4%** |
| **Idempotent_PutGetDone** | 179 ns | **117 ns** | **-34.6%** |
| **DelayingQueue_PutWithDelay** | 265 ns | **235 ns** | **-11.3%** |
| **Idempotent_Put** | 517 ns | **512 ns** | -1.0% (converged) |
| **Queue_PutAndGet** | 52 ns | **47 ns** | -9.6% |
| RateLimiting_PutWithLimited | 287 ns | 279 ns | -2.8% |

---

## Changes Made (7 files, 60 insertions, 36 deletions)

### 1. `queue.go` — Single-state-map idempotent design (Primary optimization)

**Before:** Two separate sets (`dirty` + `processing`) with 3 map lookups per Put, 2 map ops per Get.

**After:** Single `state` set using `TryAdd` (combined check-and-add) for Put, NO state mutation for Get, `TryRemove` for Done.

- **Idempotent_Put:** `Contains(dirty)` + `Contains(processing)` + `Add(dirty)` → `TryAdd(state)` (3 map ops → 1)
- **Idempotent_Get:** `Add(processing)` + `Remove(dirty)` → nothing (2 map ops → 0)
- **Idempotent_Done:** `Contains(processing)` + `Remove(processing)` → `TryRemove(state)` (2 map ops → 1)

### 2. `interface.go` — Set interface extended with `TryAdd` and `TryRemove`

Added atomic check-and-mutate methods to the `Set` interface to eliminate redundant map lookups.

### 3. `internal/container/set/set.go` — Added `TryAdd`, `TryRemove`, `NewWithCapacity`

```go
func (s *Set) TryAdd(item interface{}) bool    // single map read+write
func (s *Set) TryRemove(item interface{}) bool // single map read+delete
func NewWithCapacity(capacity int) *Set        // pre-allocated map
```

### 4. `delaying_queue.go` — Batch drain + simplified delay calculation

**Puller batch drain:** Instead of popping one item and re-acquiring the lock per item, collect up to 128 expired items in a single lock acquisition. This reduces lock contention between the puller and `PutWithDelay()`:
- pprof: `Mutex.Lock` cumulative dropped from 26.17% to 11.65% (-55%)
- pprof: `procyield` (spin-wait) dropped from 19.86% to 4.62% (-77%)

**toDelay simplification:** `time.Now().Add(time.Millisecond * time.Duration(d)).UnixMilli()` → `time.Now().UnixMilli() + d`. Eliminates intermediate `time.Time` and `time.Duration` arithmetic.

### 5. `internal/container/heap/heap.go` — Direct branching in RBTree insert

Replaced `ternary.If()` generic function call with direct `if/else` in the insert loop. Eliminates function call overhead per tree level and enables compiler branch optimization.

### 6. `config.go` — Default Set with pre-allocated capacity (64 entries)

```go
var defaultNewSetFunc = func() Set { return set.NewWithCapacity(64) }
```

### 7. `queue_test.go` — Updated test assertions for single-state design

---

## Optimization Rounds (with convergence tracking)

| Round | Target | Result | Per-round Delta |
|---|---|---|---|
| R1 (aborted) | sync.Pool → free list | Regression, reverted | N/A |
| R1b | Dual-set → single state map | Get -94%, PutGetDone -35% | Major ✅✅✅ |
| R2 | ternary.If → if/else in RBTree | Minor improvement | ~1% |
| R3 | Contains+Remove → TryRemove in Done | PutGetDone further -5% | Minor ✅ |
| R4 | DelayingQueue batch drain + toDelay | PutWithDelay -9% | Moderate ✅ |
| R5 | Set pre-allocation (capacity 64) | No measurable impact | <1% |
| **Convergence** | | | **< 5% → STOP** |

---

## Remaining Hotspots (pprof-confirmed as inherent)

| Path | Cost | Dominant Factor | Why Hard |
|---|---|---|---|
| Idempotent_Put | 512ns | `map[interface{}]` hash (18%), probe (18%), write (28%) | Requires typed maps (API change) |
| RateLimiting_PutWithLimited | 279ns | RBTree insert (24%), rate.Limiter (12%) | External library + O(log n) insert |
| DeadLetter_PutGetAck | 247ns | GC drain 49% from benchmark's `map[string]string` alloc | Benchmark allocation, not library issue |
| RetryQueue_RetryPath | 178ns | Lock contention (28%), atomic ops (9%) | Multi-step operation with 3+ locks |

---

## Files Modified

```
 config.go                       |  2 +-
 delaying_queue.go               | 31 +++++++++++++++++++------------
 interface.go                    |  4 ++++
 internal/container/heap/heap.go |  7 +++++--
 internal/container/set/set.go   | 22 ++++++++++++++++++++++
 queue.go                        | 27 ++++++++-------------------
 queue_test.go                   |  3 +--
 7 files changed, 60 insertions(+), 36 deletions(-)
```
