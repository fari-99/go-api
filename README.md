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

## 📚 Recommended Features to Learn (TODO)
To further elevate this project, here are the recommended features and patterns to explore:

- [ ] **Circuit Breaker Pattern**: Implement `gobreaker` or similar to handle external service failures gracefully.
- [ ] **OpenTelemetry**: Integrate tracing and metrics for better observability (Prometheus/Grafana).
- [ ] **gRPC Support**: Implement high-performance RPC communication using Protocol Buffers.
- [ ] **API Documentation**: Auto-generate Swagger/OpenAPI documentation using `swag`.

---

## 📄 License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
