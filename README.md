# Weaver — Code Vulnerability Vector Database

Weaver is a semantic search system for code vulnerabilities. It indexes code snippets as vector embeddings and enables similarity-based search to find related security issues across codebases.

## 🎯 What Does It Do?

- **Index code vulnerabilities** with rich metadata (CWE, CVE, CVSS scores, etc.)
- **Semantic search** — find similar vulnerabilities based on code patterns, not just keywords
- **Multi-language support** — C, C++, Python, Go, Java, JavaScript, Rust, and more
- **Bulk processing** — index entire directories of code automatically

**Example:** Search for "buffer overflow" patterns and find `strcpy()`, `memcpy()`, `sprintf()`, and similar unsafe operations — even if they use different function names.

## 🏗️ Architecture

```
┌─────────────────┐
│  Weaver (Go)    │  Reads code files, manages indexing/search
│  - weaver.go    │
│  - bulk_indexer │
└────────┬────────┘
         │ HTTP POST /embed
         ▼
┌─────────────────────────────┐
│  Veccer (Python FastAPI)    │  Converts code → vectors
│  sentence-transformers      │
│  intfloat/e5-base-v2        │
└────────┬────────────────────┘
         │ 768-dim vector
         ▼
┌─────────────────────────────┐
│  Weaviate (Docker)          │  Stores vectors, performs search
│  Vector Database            │
└─────────────────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.24+**
- **Python 3.12+** with Poetry
- **Docker** and Docker Compose
- **~1GB disk space** (for model and dependencies)

### 1. Start the Python Embedding Service

```powershell
cd veccer
poetry install
poetry run python src/veccer/main.py
```

The service starts on `http://localhost:5005`. First run downloads the e5-base-v2 model (~400MB).

### 2. Start Weaviate Vector Database

```powershell
# From project root
docker-compose up -d
```

Weaviate runs on `http://localhost:8080`.

### 3. Build and Run Weaver

```powershell
# Build
go build -o weaver.exe .

# Index example vulnerabilities
.\weaver.exe --mode=dir --dir=examples

# Index a single file
.\weaver.exe --mode=file --file=examples\vulnerable_code.c

# Search for similar code patterns
.\weaver.exe --mode=search --search=examples\vulnerable_code.c
```

## 📖 Usage

### Command-Line Modes

**Index a directory** (recursively processes all code files):
```powershell
.\weaver.exe --mode=dir --dir=path\to\code
```

**Index a single file** (with metadata):
```powershell
.\weaver.exe --mode=file --file=path\to\vulnerable.c
```

**Search for similar vulnerabilities** (semantic similarity):
```powershell
.\weaver.exe --mode=search --search=path\to\query.c
```

### What Gets Indexed?

Each code snippet is stored with:
- ✅ Source code text
- ✅ Programming language
- ✅ Vulnerability type (Buffer Overflow, SQL Injection, etc.)
- ✅ CWE, CVE identifiers
- ✅ CVSS score and vector
- ✅ File path, function name, library
- ✅ Severity, exploit/patch availability
- ✅ Version information
- ✅ Audit tool and auditor

## 🧪 Examples

Sample vulnerable code is provided in `examples/`:

- **`vulnerable_code.c`** — Buffer Overflow (CWE-120)
  ```c
  strcpy(buffer, user_input); // No bounds checking!
  ```

- **`sql_injection.py`** — SQL Injection (CWE-89)
  ```python
  query = f"SELECT * FROM users WHERE username = '{username}'"
  ```

- **`path_traversal.go`** — Path Traversal (CWE-22)
  ```go
  fullPath := filepath.Join(baseDir, filename) // No validation!
  ```

## 🔍 How Semantic Search Works

Traditional keyword search would miss these as "similar":

```c
strcpy(buffer, input);    // Version 1
memcpy(dest, src, len);   // Version 2
sprintf(buf, "%s", str);  // Version 3
```

**Weaver's semantic search** understands these are all buffer operations and finds them as related (similarity ~0.75-0.85).

## 🛠️ Development

### Project Structure

```
weaver/
├── weaver.go              # Main CLI application
├── bulk_indexer.go        # Directory indexing logic
├── alternative_vectors.go # Educational examples (not for production)
├── .vscode/
│   └── launch.json       # VS Code debug configurations
├── examples/             # Sample vulnerable code
│   ├── vulnerable_code.c
│   ├── sql_injection.py
│   └── path_traversal.go
├── veccer/               # Python embedding service
│   ├── pyproject.toml
│   ├── README.md
│   └── src/veccer/main.py
├── VECTORIZATION.md      # Why we chose this model
├── USAGE.md             # Detailed usage guide
└── docker-compose.yml   # Weaviate setup
```

### Debugging in VS Code

1. Open the project in VS Code
2. Go to **Run and Debug** (Ctrl+Shift+D)
3. Select a configuration:
   - **Debug Weaver - Index Directory**
   - **Debug Weaver - Index Single File**
   - **Debug Weaver - Search**
   - **Debug Current Go File**
4. Set breakpoints and press **F5**

### Running Tests

```powershell
# Go tests
go test ./...

# Python service in dev mode (auto-reload)
cd veccer
poetry run uvicorn veccer.main:app --reload --port 5005
```

## 🤔 Why This Approach?

### Vectorization Model: `intfloat/e5-base-v2`

**Why we chose it:**
- ✅ **Semantic understanding** — captures code meaning, not just syntax
- ✅ **Resource efficient** — runs on CPU, ~400MB model size
- ✅ **Fast** — 50-100ms per embedding
- ✅ **Free & open-source** — no API costs
- ✅ **Battle-tested** — popular in production systems

**Alternatives we rejected:**
- ❌ **Simple hash/encoding** — no semantic similarity, only exact matches
- ❌ **Large LLMs (GPT/Claude)** — too slow and expensive for this use case
- ⚠️ **CodeBERT** — slightly better but heavier, minimal benefit

See `VECTORIZATION.md` for detailed comparison.

## 📊 Performance

| Operation | Time | Notes |
|-----------|------|-------|
| First embedding | 2-3s | Model loading |
| Subsequent embeddings | 50-100ms | CPU-based |
| Index single file | ~150-200ms | Including I/O |
| Search query | 10-50ms | Weaviate HNSW index |
| Batch 100 files | ~10-15s | Parallelizable |

**Resource usage:**
- Python service: ~1GB RAM (model loaded)
- Weaviate: ~200MB base + indexed data
- Go CLI: Minimal (<50MB)

## 🐛 Troubleshooting

### Python service won't start

**Issue:** `poetry: command not found`

**Solution:** Install Poetry:
```powershell
(Invoke-WebRequest -Uri https://install.python-poetry.org -UseBasicParsing).Content | python -
```

### Model download is slow

**Issue:** First run takes 5-10 minutes

**Solution:** This is normal. The e5-base-v2 model (~400MB) downloads once. Ensure stable internet and ~1GB free disk space.

### Port already in use

**Issue:** `Error: port 5005 already in use`

**Solution:** Kill the process or change the port in `veccer/src/veccer/main.py`:
```python
uvicorn.run(app, host="0.0.0.0", port=5006)  # Different port
```

### Weaviate connection error

**Issue:** `failed to connect to localhost:8080`

**Solution:** Ensure Docker is running and Weaviate started:
```powershell
docker-compose up -d
docker-compose logs weaviate
```

### No results in search

**Issue:** Search returns empty results

**Solution:** Ensure you've indexed some code first:
```powershell
.\weaver.exe --mode=dir --dir=examples
```

## 📚 Documentation

- **`README.md`** (this file) — Project overview and quick start
- **`USAGE.md`** — Detailed CLI usage and examples
- **`VECTORIZATION.md`** — Deep dive on model selection
- **`veccer/README.md`** — Python embedding service documentation
- **`QUICKSTART.md`** — Ultra-condensed reference card
- **`IMPLEMENTATION_SUMMARY.md`** — Technical implementation notes

## 🤝 Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Commit your changes (`git commit -m 'Add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

**What we're looking for:**
- Bug fixes
- Performance improvements
- Additional language support
- Better metadata extraction
- Documentation improvements

## 🔐 Security

This tool is designed to **index** vulnerabilities for research and analysis. If you find a security issue in Weaver itself, please report it privately.

## 📝 License

See `LICENSE` file.

## 🙏 Acknowledgments

- **sentence-transformers** — embedding framework
- **Weaviate** — vector database
- **FastAPI** — Python web framework
- **intfloat/e5-base-v2** — embedding model

## 🔗 Related Projects

- [Weaviate](https://weaviate.io/) — Open-source vector database
- [sentence-transformers](https://www.sbert.net/) — State-of-the-art text embeddings
- [Semgrep](https://semgrep.dev/) — Static analysis tool for finding bugs

---

**Need help?** Check out `USAGE.md` for detailed examples or open an issue!