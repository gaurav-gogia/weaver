# Veccer - Code Embedding Service

A lightweight FastAPI service that converts code snippets into semantic vector embeddings using sentence-transformers.

## What Does This Do?

This service provides a simple HTTP API that takes code (as text) and returns a 768-dimensional vector representation. These vectors capture the **semantic meaning** of code, allowing you to:

- Find similar code snippets
- Search for related vulnerabilities
- Detect code patterns
- Compare code functionality

## How It Works

```
Code String → sentence-transformers → 768-dim Vector
   ↓
"strcpy(buf, input)"
   ↓
[0.23, -0.45, 0.67, ... 768 numbers]
```

The model used is **`intfloat/e5-base-v2`**:
- ✅ Understands code semantics
- ✅ CPU-only (no GPU required)
- ✅ ~400MB model size
- ✅ Fast (~50-100ms per encoding)

## Prerequisites

- **Python 3.12 or higher**
- **Poetry** (Python package manager)

### Installing Poetry

**Windows (PowerShell):**
```powershell
(Invoke-WebRequest -Uri https://install.python-poetry.org -UseBasicParsing).Content | python -
```

**macOS/Linux:**
```bash
curl -sSL https://install.python-poetry.org | python3 -
```

**Or via pip:**
```bash
pip install poetry
```

Verify installation:
```bash
poetry --version
```

## Quick Start

### 1. Install Dependencies

```bash
# Navigate to the veccer directory
cd veccer

# Install all dependencies (this will take a few minutes)
poetry install
```

**What happens:**
- Creates a virtual environment automatically
- Installs FastAPI, uvicorn, sentence-transformers, and dependencies
- Downloads the e5-base-v2 model (~400MB) on first run

### 2. Run the Server

```bash
poetry run python src/veccer/main.py
```

**Expected output:**
```
starting server at: 5005
INFO:     Started server process [12345]
INFO:     Waiting for application startup.
INFO:     Application startup complete.
INFO:     Uvicorn running on http://0.0.0.0:5005
```

The server is now running on **http://localhost:5005**

### 3. Test the API

**Using curl (PowerShell):**
```powershell
curl -X POST http://localhost:5005/embed `
  -H "Content-Type: application/json" `
  -d '{"text": "strcpy(buffer, user_input);"}'
```

**Using curl (bash/Linux/macOS):**
```bash
curl -X POST http://localhost:5005/embed \
  -H "Content-Type: application/json" \
  -d '{"text": "strcpy(buffer, user_input);"}'
```

**Expected response:**
```json
{
  "vector": [0.234, -0.456, 0.678, ... 768 numbers total]
}
```

## API Documentation

Once the server is running, visit:
- **Swagger UI**: http://localhost:5005/docs
- **ReDoc**: http://localhost:5005/redoc

### Endpoint: POST /embed

**Request:**
```json
{
  "text": "your code snippet here"
}
```

**Response:**
```json
{
  "vector": [array of 768 floating-point numbers]
}
```

**Example (Python):**
```python
import requests

response = requests.post(
    "http://localhost:5005/embed",
    json={"text": "strcpy(buffer, user_input);"}
)

vector = response.json()["vector"]
print(f"Vector dimensions: {len(vector)}")  # 768
```

**Example (Go):**
```go
// See ../weaver.go for the full implementation
reqBody := EmbedRequest{Text: "strcpy(buffer, user_input);"}
jsonData, _ := json.Marshal(reqBody)
resp, _ := http.Post("http://localhost:5005/embed", "application/json", bytes.NewBuffer(jsonData))

var result EmbedResponse
json.NewDecoder(resp.Body).Decode(&result)
// result.Vector now contains [768]float64
```

## Development

### Project Structure

```
veccer/
├── pyproject.toml          # Dependencies and project config
├── README.md              # This file
└── src/
    └── veccer/
        └── main.py        # FastAPI server
```

### Running in Development Mode

```bash
# Auto-reload on code changes
poetry run uvicorn veccer.main:app --reload --port 5005
```

### Adding Dependencies

```bash
poetry add package-name
```

### Updating Dependencies

```bash
poetry update
```

## Troubleshooting

### Issue: "poetry: command not found"

**Solution:** Add Poetry to your PATH or use the full path:
```bash
# Windows
%APPDATA%\Python\Scripts\poetry

# macOS/Linux
~/.local/bin/poetry
```

### Issue: "Model download fails or is slow"

**Solution:** The first run downloads ~400MB. Ensure you have:
- Stable internet connection
- Sufficient disk space (~1GB free)
- Wait for download to complete (may take 5-10 minutes)

### Issue: Port 5005 already in use

**Solution:** Change the port in `main.py`:
```python
uvicorn.run(app, host="0.0.0.0", port=5006)  # Use different port
```

Or kill the process using the port:
```powershell
# Windows
netstat -ano | findstr :5005
taskkill /PID <PID> /F

# Linux/macOS
lsof -ti:5005 | xargs kill -9
```

### Issue: Import errors or module not found

**Solution:** Ensure you're running with `poetry run`:
```bash
# ❌ Wrong
python src/veccer/main.py

# ✅ Correct
poetry run python src/veccer/main.py
```

### Issue: Slow embedding performance

**Possible causes:**
- First request loads model into memory (~2-3 seconds)
- Subsequent requests should be fast (50-100ms)
- Very large code snippets (>1000 lines) will be slower

## Performance

- **First request**: 2-3 seconds (model loading)
- **Subsequent requests**: 50-100ms per embedding
- **Concurrent requests**: Supported (FastAPI is async)
- **Memory usage**: ~1GB (model in memory)
- **CPU usage**: 1-2 cores during encoding

## Production Deployment

### Using Docker (Recommended)

Create a `Dockerfile`:
```dockerfile
FROM python:3.12-slim

WORKDIR /app

# Install poetry
RUN pip install poetry

# Copy dependency files
COPY pyproject.toml poetry.lock* ./

# Install dependencies
RUN poetry config virtualenvs.create false \
    && poetry install --no-dev --no-interaction --no-ansi

# Copy application
COPY src/ ./src/

# Expose port
EXPOSE 5005

# Run server
CMD ["python", "src/veccer/main.py"]
```

Build and run:
```bash
docker build -t veccer .
docker run -p 5005:5005 veccer
```

### Using systemd (Linux)

Create `/etc/systemd/system/veccer.service`:
```ini
[Unit]
Description=Veccer Embedding Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/veccer
ExecStart=/usr/local/bin/poetry run python src/veccer/main.py
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable veccer
sudo systemctl start veccer
```

## Why This Model?

**intfloat/e5-base-v2** was chosen because:

1. **Semantic understanding**: Trained on text + code, understands programming concepts
2. **Resource efficient**: Runs on CPU, no GPU needed
3. **Fast**: Optimized for production use
4. **Open source**: Free, no API keys or costs
5. **Well-tested**: Popular in the sentence-transformers ecosystem

**Alternatives considered:**
- ❌ Simple encoding (hash, ASCII): No semantic meaning
- ❌ Large LLMs (GPT, Claude): Too slow and expensive
- ⚠️ CodeBERT: Slightly better but heavier, minimal benefit

See `../VECTORIZATION.md` for detailed comparison.

## Integration with Weaver

This service is used by the Weaver vulnerability database to:

1. Convert code snippets to vectors
2. Store vectors in Weaviate
3. Enable semantic search for similar vulnerabilities

**Workflow:**
```
Code File → Weaver (Go) → Veccer (Python) → Vector → Weaviate DB
```

## Support

- **Issues**: Report at the main Weaver repository
- **Model docs**: https://huggingface.co/intfloat/e5-base-v2
- **sentence-transformers**: https://www.sbert.net/

## License

Same as parent Weaver project.
