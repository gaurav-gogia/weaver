# Weaver - Code Vulnerability Vector Database

A semantic search system for finding similar code vulnerabilities using vector embeddings.

## Quick Start

### 1. Start the Python Embedding Service

```bash
cd veccer
poetry install
poetry run python src/veccer/main.py
```

The service will start on `http://localhost:5005`

### 2. Start Weaviate (using Docker)

```bash
docker-compose up -d
```

### 3. Run Weaver

**Index a single file:**
```bash
go run . --mode=file --file=examples/vulnerable_code.c
```

**Index entire directory:**
```bash
go run . --mode=dir --dir=examples
```

**Search for similar vulnerabilities:**
```bash
go run . --mode=search --search=examples/vulnerable_code.c
```

## Features

✅ **File-based indexing**: Read code from files instead of hardcoding
✅ **Semantic vectorization**: Uses `intfloat/e5-base-v2` model for understanding code
✅ **Bulk indexing**: Process entire directories of code
✅ **Similarity search**: Find vulnerabilities with similar patterns
✅ **Multi-language**: Supports C, C++, Python, Go, Java, JavaScript, and more

## How It Works

1. **Read**: Load code files from disk
2. **Vectorize**: Convert code to 768-dimensional vectors using sentence-transformers
3. **Index**: Store in Weaviate vector database with metadata
4. **Search**: Find similar vulnerabilities using semantic similarity

## Why Sentence Transformers?

See [VECTORIZATION.md](./VECTORIZATION.md) for detailed comparison of approaches.

**TL;DR**: Sentence transformers provide semantic understanding of code (finding similar patterns) without the cost and complexity of large language models, while being infinitely better than simple string encoding.

## Example Vulnerabilities

The `examples/` directory contains sample vulnerable code:

- `vulnerable_code.c` - Buffer overflow (CWE-120)
- `sql_injection.py` - SQL injection (CWE-89)
- `path_traversal.go` - Path traversal (CWE-22)

## Architecture

```
┌──────────────┐     HTTP      ┌───────────────────┐
│   weaver.go  │ ────────────> │  Python FastAPI   │
│              │               │  (veccer service) │
│  Read files  │               │                   │
│  Index data  │               │  sentence-trans.  │
│  Search DB   │               │  intfloat/e5-v2   │
└──────────────┘               └───────────────────┘
       │                                │
       │                                │ 768-dim vector
       ▼                                ▼
┌────────────────────────────────────────────────────┐
│              Weaviate Vector Database              │
│  - HNSW index for fast similarity search          │
│  - Stores code + metadata + vector embeddings     │
└────────────────────────────────────────────────────┘
```

## Command-Line Options

```
--mode     Operation mode: 'file', 'dir', or 'search' (default: 'dir')
--file     Path to single code file (for mode=file)
--dir      Path to directory (default: 'examples')
--search   Path to code file to use as search query (for mode=search)
```

## Development

Build and run:
```bash
go build -o weaver.exe
./weaver.exe --mode=dir --dir=examples
```

## License

See LICENSE file.
