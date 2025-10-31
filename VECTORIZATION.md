# Vectorization Approach for Code Vulnerability Search

## Executive Summary

**Recommendation: Use `sentence-transformers` (intfloat/e5-base-v2) model**

This document explains the rationale behind our vectorization strategy for the Weaver code vulnerability database.

## Comparison of Approaches

### ✅ RECOMMENDED: Sentence Transformers (intfloat/e5-base-v2)

**What it is:**
- A pre-trained embedding model that converts text/code into 768-dimensional vectors
- Understands semantic similarity, not just keyword matching
- Model: `intfloat/e5-base-v2` (~400MB)

**Pros:**
- ✅ **Semantic understanding**: Captures code meaning and patterns
- ✅ **Resource efficient**: Runs on CPU, no GPU required
- ✅ **Fast**: ~50-100ms per embedding on modern CPUs
- ✅ **Easy to use**: Simple API, well-documented
- ✅ **Code-aware**: Trained on mixed text/code datasets
- ✅ **Similar vulnerabilities**: Finds patterns like `strcpy()` ≈ `memcpy()` ≈ `sprintf()`
- ✅ **Already implemented**: Python service ready to use

**Cons:**
- ⚠️ External service dependency (Python FastAPI)
- ⚠️ ~400MB model download on first run

**Use cases:**
- Finding similar vulnerability patterns across different codebases
- Semantic code search (not just exact matches)
- Identifying related security issues

---

### ❌ NOT RECOMMENDED: Simple String Encoding

**What it is:**
- Convert strings to numbers using hash functions, ASCII values, or character encoding
- Examples: CRC32, MD5 hash as floats, base64 → float arrays

**Pros:**
- ✅ No external dependencies
- ✅ Extremely fast
- ✅ Deterministic

**Cons:**
- ❌ **No semantic similarity**: `strcpy(buf, input)` is completely different from `memcpy(buf, input, len)`
- ❌ **Only exact matches**: Cannot find related vulnerabilities
- ❌ **No pattern recognition**: Misses variations of the same vulnerability
- ❌ **Useless for search**: Random similarity scores

**Example:**
```go
// This approach is NOT suitable for semantic search
func simpleEncode(code string) []float32 {
    vec := make([]float32, 128)
    for i, c := range code {
        if i >= len(vec) { break }
        vec[i] = float32(c) // Just ASCII values - no semantic meaning
    }
    return vec
}
```

**Why this fails:**
These two vulnerable snippets would have completely different vectors despite being similar:
```c
strcpy(buffer, user_input);     // Vector: [115, 116, 114, ...]
memcpy(buffer, user_input, len); // Vector: [109, 101, 109, ...]
```

---

### ❌ NOT RECOMMENDED: Large Language Models (GPT-4, Claude, etc.)

**What it is:**
- Use LLM APIs to generate embeddings
- OpenAI's `text-embedding-ada-002`, Claude embeddings, etc.

**Pros:**
- ✅ State-of-the-art semantic understanding
- ✅ Very high-quality embeddings

**Cons:**
- ❌ **Expensive**: $0.0001-0.0004 per 1K tokens (adds up quickly)
- ❌ **Slow**: 200-500ms per request (network latency)
- ❌ **API dependency**: Requires internet, API keys, rate limits
- ❌ **Privacy concerns**: Sending code to third parties
- ❌ **Overkill**: Similar quality to sentence-transformers for this use case

**Cost comparison:**
- Indexing 10,000 code snippets:
  - LLM API: $5-20 + ongoing costs
  - Sentence-transformers: $0 (one-time 400MB download)

---

### ⚠️ ALTERNATIVE: Code-Specific Models (CodeBERT, GraphCodeBERT)

**What it is:**
- Models specifically trained on code
- Examples: `microsoft/codebert-base`, `microsoft/graphcodebert-base`

**Pros:**
- ✅ Optimized for code understanding
- ✅ Better at code syntax/structure

**Cons:**
- ⚠️ Larger model size (500MB+)
- ⚠️ Marginally better than e5-base-v2 for our use case
- ⚠️ More complex setup

**Verdict:**
Good alternative, but `e5-base-v2` is sufficient and lighter.

---

## Implementation Details

### Current Architecture

```
┌─────────────────┐
│   Go Client     │ ──┐
│  (weaver.go)    │   │
└─────────────────┘   │
                      │ 1. Code snippet
                      ▼
┌─────────────────────────────────┐
│   Python FastAPI Service        │
│   (veccer/main.py)              │
│                                 │
│   Model: intfloat/e5-base-v2   │
│   Output: 768-dim vector       │
└─────────────────────────────────┘
                      │
                      │ 2. Vector embedding
                      ▼
┌─────────────────────────────────┐
│      Weaviate Vector DB         │
│   - Stores code + vector        │
│   - Performs similarity search  │
└─────────────────────────────────┘
```

### Vector Dimensions

- **Size**: 768 floats (3KB per vector)
- **Range**: Typically -1.0 to 1.0 (normalized)
- **Similarity metric**: Cosine similarity (Weaviate default)

### Performance Benchmarks

| Operation | Time | Notes |
|-----------|------|-------|
| Generate embedding | ~50-100ms | CPU-based, no GPU needed |
| Index single file | ~150-200ms | Including file I/O + vectorization |
| Search query | ~10-50ms | Weaviate HNSW index |
| Batch 100 files | ~10-15s | Parallelizable |

## Usage Examples

### Index a single file
```bash
go run . --mode=file --file=examples/vulnerable_code.c
```

### Index entire directory
```bash
go run . --mode=dir --dir=examples
```

### Search for similar vulnerabilities
```bash
go run . --mode=search --search=examples/buffer_overflow.c
```

## Alternative Simple Encoding (NOT RECOMMENDED)

If you absolutely cannot use the Python service, here's a minimal approach:

```go
// WARNING: This provides NO semantic similarity
func simpleHash(code string) []float32 {
    import "hash/fnv"

    h := fnv.New128a()
    h.Write([]byte(code))
    hashBytes := h.Sum(nil)

    vec := make([]float32, 128)
    for i := range vec {
        vec[i] = float32(hashBytes[i%len(hashBytes)])
    }
    return vec
}
```

**Use this ONLY if:**
- You need exact duplicate detection
- Semantic similarity is not important
- You cannot run the Python service

## Recommendation

**Use the existing `sentence-transformers` setup.**

It provides the best balance of:
- ✅ Accuracy (semantic understanding)
- ✅ Performance (fast enough for production)
- ✅ Resource usage (CPU only, ~400MB)
- ✅ Ease of use (simple HTTP API)
- ✅ Cost (free, open-source)

## References

- [sentence-transformers](https://www.sbert.net/)
- [intfloat/e5-base-v2](https://huggingface.co/intfloat/e5-base-v2)
- [Weaviate Vector Database](https://weaviate.io/)
- [Why semantic search matters for code](https://arxiv.org/abs/2002.08155)
