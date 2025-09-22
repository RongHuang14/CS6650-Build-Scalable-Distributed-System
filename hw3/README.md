## HW3: Threading and Load Tests

### Prerequisites
- Go 1.21+ installed (`go version`)
- Python 3.9+ with Locust for local runs: `pip install locust`
- Docker + Docker Compose (optional, for distributed Locust)

### Repo Layout
- `part2/`: Go experiments (atomicity, collections, file I/O, context switching)
- `part3/`: Simple Go HTTP server and Locust test files
- `screenshot/` and `report.pdf`: screenshot and brief report

## Part II: Thread Experiments (Go)
All examples live in `part2/`. Run each with `go run`, and repeat runs to gather averages. Use `-race` where relevant.

- Atomicity
```bash
go run -race /Users/ronghuang/MyCScode/NEU/cs6650/hw3/part2/atomic_counter.go
go run -race /Users/ronghuang/MyCScode/NEU/cs6650/hw3/part2/non_atomic.go
```

- Collections (plain map, mutex, RWMutex, sync.Map)
```bash
go run -race /Users/ronghuang/MyCScode/NEU/cs6650/hw3/part2/collections_plain.go
go run -race /Users/ronghuang/MyCScode/NEU/cs6650/hw3/part2/collections_mutex.go
go run -race /Users/ronghuang/MyCScode/NEU/cs6650/hw3/part2/collections_rwmutex.go
go run -race /Users/ronghuang/MyCScode/NEU/cs6650/hw3/part2/collections_syncmap.go
```

- File access (buffered vs unbuffered)
```bash
go run /Users/ronghuang/MyCScode/NEU/cs6650/hw3/part2/file_access.go
```

- Context switching (GOMAXPROCS=1 vs default)
```bash
go run /Users/ronghuang/MyCScode/NEU/cs6650/hw3/part2/context_switching.go
```

Capture runtimes/lengths and summarize tradeoffs in your report.

## Part III: Load Testing with Locust
### 1) Start the server
```bash
cd /Users/ronghuang/MyCScode/NEU/cs6650/hw3/part3
go run server.go
# Server listens on http://localhost:8080 with /get and /post
```

### 2) Run Locust locally (single process)
- Basic 1:1 GET/POST
```bash
locust -f /Users/ronghuang/MyCScode/NEU/cs6650/hw3/part3/locustfile.py --host http://localhost:8080
# Open http://localhost:8089 and start users
```

- Local 3:1 GET/POST and short wait
```bash
locust -f /Users/ronghuang/MyCScode/NEU/cs6650/hw3/part3/locustfile_local.py --host http://localhost:8080
```

- High-throughput client (FastHttpUser)
```bash
locust -f /Users/ronghuang/MyCScode/NEU/cs6650/hw3/part3/locustfile_fast.py --host http://localhost:8080
```

Headless example (matches suggested params):
```bash
locust -f /Users/ronghuang/MyCScode/NEU/cs6650/hw3/part3/locustfile_local.py \
  --host http://localhost:8080 --headless -u 50 -r 10 -t 30s
```

### 3) Run Locust with Docker (master/worker)
```bash
cd /Users/ronghuang/MyCScode/NEU/cs6650/hw3/part3
docker compose up --pull always
# Web UI: http://localhost:8089 (uses locustfile_local.py by default)
# In the UI set Host to http://host.docker.internal:8080
# Scale workers if desired:
# docker compose up --scale worker=4
```

### Notes
- Ensure Host is a full URL with scheme (e.g., `http://localhost:8080`).
- Keep an eye on CPU usage; add `wait_time` or reduce users/workers if your machine saturates.
- Compare GET vs POST latency/throughput; then switch to `FastHttpUser` to observe client-side impact.

### What to Hand In
- Short PDF report with screenshots and brief observations (already placed as `report.pdf`).
- Updated code in this repo so you can re-run during the mock interview. 