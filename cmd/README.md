# 🖥️ Entry Points (Cmd)

This directory contains the main application entry points, including HTTP servers, task runners, and CLI commands.

## 📁 Command Overview

### 🌐 Servers (`servers/`)
- **Main Server**: The primary HTTP server entry point in `cmd/servers/main/main.go`.
- **Server Tests**: Integrated tests for server initialization and routing in `main_test`.

### ⚡ Tasks (`tasks/`)
- **Background Workers**: Entry points for background task processing (e.g., consumer for message queues).
- **Scheduled Jobs**: Commands intended to be run by a scheduler (like cron).

---

## 🛠️ Running the Commands
To run the main server:
```bash
go run cmd/servers/main/main.go
```

To run a specific background task:
```bash
go run cmd/tasks/<task_name>/main.go
```
