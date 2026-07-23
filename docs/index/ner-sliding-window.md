---
id: okf-013
feature: ner-sliding-window
branch: feature/ner-sliding-window
status: done
files:
  - nlp/bertner/window.go (Windows + MergeEntities)
  - nlp/bertner/tokenizer.go (Aggregate: never splits mid-word)
  - nlp/onnx/engine.go (256 windows/64 overlap, parallel inference, RWMutex)
  - examples/ner/main.go (E2E long text, entities near end of document)
tests:
  - nlp/bertner/window_test.go + tokenizer_test.go (92.5%)
decisions:
  - "2026-07-17: 256-token windows (not 512) — labels degrade near the end of long sequences AND O(n²) attention makes 2×256 cheaper than 1×512"
  - "2026-07-17: 64-token overlap; MergeEntities fusion = longest span then score"
  - "2026-07-17: ORT sessions thread-safe for Run → RWMutex (Load/Destroy as writers), windows inferred in parallel"
  - "2026-07-17: an entity never ends mid-word (## sub-word tagged O gets reattached)"
---

**What**: NER now sees the ENTIRE text — overlapping windows inferred in
parallel, cross-window duplicate merging, words never split.

**Measurement/E2E**: entities at position 4700+ (600+ tokens): before =
invisible; after = "Marie Curie" PERSON 0.96, "University of Warsaw" ORG 0.95.

**Pitfalls**:
- In a full 512 window, the model's labels degrade near the end of the
  sequence ("Marie" → ORG 0.64): a model symptom, not a code bug — hence
  the 256-token windows.
- The last sub-word of a name can be labeled O by the model ("Curi/##e"):
  reattach ## continuations to the open entity.
