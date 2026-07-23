---
id: okf-012
feature: perf-parallel-analyze
branch: feature/perf-parallel-analyze
status: done
files:
  - contextaware/enhancer.go (bounded backward/forward scan window — no more slicing the whole prefix)
  - analyzer/analyzer.go (goroutine fan-out for recognizers, reassembled in registry order)
tests:
  - analyzer/analyzer_bench_test.go (30 KB document, full en+fr registry)
  - analyzer + contextaware suites unchanged, passing with -race
decisions:
  - "2026-07-17: the benchmark showed the REAL bottleneck was the enhancer (O(results × text length)), not sequentiality — always measure before optimizing"
  - "2026-07-17: fan-out determinism guaranteed by indexed reassembly in registry order"
---

**What**: the two pipeline optimizations — context window scanning bounded
to the entity's neighborhood, and parallel execution of recognizers
(independent and thread-safe).

**Measurement** (30 KB document, 28 recognizers): 550 ms/op → 13.3 ms/op
(×41), 534 MB → 3.8 MB allocated per analysis (÷139).
