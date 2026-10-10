# 🚀 Advanced Go Backend API

## 📝 Description
This repository serves as a robust, production-ready foundation for a Go-based API service. It is designed with a focus on clean architecture, scalability, and modern backend practices. The project integrates multiple third-party services, implementing advanced features like distributed locking, multi-channel notifications, and complex state machine logic.

Built with **Go 1.25.0**, this project is an ongoing journey in mastering backend engineering, demonstrating best practices in API design, security, and infrastructure integration.

---

## ✨ Features

### 🔐 Authentication & Security
- **JWT-based Authentication**: Secure session management using JSON Web Tokens.
- **Two-Factor Authentication (2FA)**: OTP implementation via Email, SMS (Twilio), WhatsApp, and Telegram.
- **Recovery Codes**: Secure generation and management of 2FA recovery codes.
- **Casbin RBAC**: Robust role-based access control for fine-grained permissions.
- **Distributed Locking**: Using `redsync` (Redis-based) to ensure consistency in concurrent environments.
- **Rate Limiting**: Intelligent rate limiting using Redis counters for OTP and login protection.

### 📡 Communication & Notifications
- **Multi-channel OTP**: Send OTPs via WhatsApp (WhatsMeow), Telegram, Email (Gomail), and SMS (Twilio).
- **Firebase Cloud Messaging (FCM)**: Push notification support for mobile and web clients.
- **Webhook Integration**: Flexible webhook system for external service communication.

### 🏗️ Architecture & Logic
- **Finite State Machine (FSM)**: Managing complex business flows (e.g., transaction states) using `looplab/fsm`.
- **Clean Module Structure**: Decoupled modules for `auths`, `users`, `notifications`, `security_cameras`, and more.
- **Helper Packages**: Native implementations for Redis helpers, OTP generation, and pagination.

### 💾 Data & Infrastructure
- **MySQL & PostgreSQL**: Using GORM with support for advanced Percona configurations.
- **Elasticsearch**: Built-in integration for high-performance searching.
- **Redis & Redis Cache**: Integrated caching layer and state management.
- **RabbitMQ**: Message queueing for asynchronous task processing.
- **AWS S3 & Local Storage**: Flexible file storage abstractions.
- **Database Migrations**: Streamlined schema management.

### 🎥 Specialized Modules
- **Security Camera Streaming**: Integration with `go2rtc` for real-time video stream synchronization.
- **WhatsApp Integration**: Native WhatsApp automation using `whatsmeow`.
- **PDF Generation**: HTML-to-PDF conversion using `wkhtmltopdf`.

---

## 🛠️ Installation & Setup

### Prerequisites
- **Go**: Version 1.25.0 or later ([Download](https://golang.org/dl/))
- **Databases**: MySQL, Redis, RabbitMQ, Elasticsearch.
- **Local Proxy** (Optional): `go-api.fadhlan.loc` for local development.

### Steps
1. **Clone the repository**:
   ```bash
   git clone <repo-url>
   cd go-api
   ```
2. **Install dependencies**:
   ```bash
   go mod vendor
   ```
3. **Environment Configuration**:
   ```bash
   cp .env.example .env
   # Edit .env with your local credentials
   ```
4. **Run the Application**:
   Using `fresh` for hot-reloading:
   ```bash
   go get github.com/pilu/fresh
   fresh
   ```
   Or traditionally:
   ```bash
   go run cmd/servers/main/main.go
   ```

---

## 🛡️ RBAC / Permissions

Authorization uses Casbin (`api_rule_access` table, model in `modules/configs/rbac_model.conf`). A request passes only if one of the user's roles has a policy matching the route and method.

### How it is wired
- RBAC is **opt-in per module**, not global. `cmd/servers/main/main.go` builds an `authorized` handler (authentication, then RBAC) and passes it to each registrator. A new module must be given `authorized` to be enforced.
- Gated: users, roles, permissions (all routes), locations, notifications, state_machine, storages, security_cameras, twoFA, whatsapp and ws_auth token issuing.
- Auth-only, no RBAC: login (`/users/auth`), `/users/sessions/*` (list, sign-out, delete, refresh), hasura and xendit.
- A denied request returns `403`; a missing or invalid session returns `401`.
- The permissions UI lists **every** registered route (`gin.Routes()`), including routes that are not RBAC-gated. Assigning a policy to an ungated route has no effect.

### First-time setup
A fresh install has no users, roles or policies, so every gated route returns 403 until they exist. Seed them with the CLI:

```bash
# inside the go-api container (loads the same env the app uses)
docker exec projects-go-api sh -c 'cd /go/src/go-api \
  && export $(cat /env_files/global.env | sed "s/#.*//g" | xargs) \
  && export $(cat /env_files/app.env | sed "s/#.*//g" | xargs) \
  && RBAC_SEED_PASSWORD="<password>" go run cmd/tasks/commands.go rbac-seed --email admin@example.com'
```

Flags: `--role` (default `admin`), `--role-type` (1 Customer, 2 Seller), `--username`, `--password` (or env `RBAC_SEED_PASSWORD`).

`rbac-seed` is idempotent. It creates the role, creates the user if the email is not found (an existing user's password is left untouched), assigns the role, and adds an allow-all policy `<role>-<UserType> /.+ .*`. Then restart the API so it picks up the new policy (see below).

To turn permission checks off during setup, set `RBAC_ENABLED=false` in `.env` and restart. Login is still required, and a `WARNING: RBAC_ENABLED=false` line is logged on boot. Set it back to `true` (or remove it) and restart when done. It is read once at startup. Never leave it `false` in a shared or production environment.

### Things to know
- **Policies are cached in memory.** The API loads policies once at startup. Policies created through the permissions API/UI go through the same enforcer and take effect immediately. Changes made out of process (the `rbac-seed` command, direct SQL) need an API restart.
- **Roles are fixed at login.** A session stores the user's roles when they log in. Changing a user's roles takes effect after they log in again.
- **Subject format is `{RoleName}-{UserType}`**, for example `admin-Customer`. Policies must use this subject. `role_name` is not unique, so check for duplicate role rows before assigning policies.
- **Users with no roles are always denied** on gated routes, even if the policy table is empty.
- **Password helper rejects weak passwords**, including ones that contain the username or email parts (for example `Rbac-Admin-1` for `rbac-admin@...`). The seed command goes through the same check.
- **Malformed `Authorization` header returns 500.** A header without a space (for example just `Bearer`) panics at `token[1]` in `checkAuthHeader` (`modules/middleware/middleware_auth.go`). This is a known issue that predates RBAC.

---

## 📚 Recommended Features to Learn (TODO)
To further elevate this project, here are the recommended features and patterns to explore:

- [ ] **Circuit Breaker Pattern**: Implement `gobreaker` or similar to handle external service failures gracefully.
- [ ] **OpenTelemetry**: Integrate tracing and metrics for better observability (Prometheus/Grafana).
- [ ] **gRPC Support**: Implement high-performance RPC communication using Protocol Buffers.
- [ ] **API Documentation**: Auto-generate Swagger/OpenAPI documentation using `swag`.

---

## 📄 License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
