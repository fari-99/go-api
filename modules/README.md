# 📦 Project Modules

This directory contains the core business logic of the application, divided into cohesive and independent modules. Each module follows a clean architecture pattern to ensure maintainability and testability.

## 📁 Modules Overview

### 🔐 Auths & Users
- **Auths**: Core authentication logic including JWT session management and login/logout flows.
- **Users**: User profile management and account-related operations.
- **TwoFA (2FA)**: Two-Factor Authentication implementation with OTP support via multiple channels.

### 🔔 Notifications & Communication
- **Notifications**: Centralizing notification delivery for Firebase (FCM), Emails, and Webhooks.
- **WhatsApp**: Native WhatsApp integration for messaging and automation.
- **Telegrams**: Integration with Telegram Bot API for real-time alerts.

### 📹 Specialized Domains
- **Security Cameras**: Managing security camera streams and integrating with `go2rtc`.
- **Calendar Managements**: Handling scheduling and event logic.
- **Locations**: Geolocation and site management.
- **State Machine**: Implementation of Finite State Machines for business process automation.

### 📐 Configuration & Storage
- **Configs**: Application-wide configurations including database, Redis, and service integrations.
- **Storages**: Multi-backend storage system for local and cloud (S3) file management.

### 🛡️ Middleware & RBAC
- **Middleware**: Gin-based middleware for CORS, Logging, and Request Validation.
- **Permissions**: Casbin-based RBAC implementation for resource protection.

---

## 🛠️ Adding a New Module
When creating a new module, ensure you follow these conventions:
1. Create a sub-directory in `modules/`.
2. Define `registrator.go` to handle routing and dependency injection.
3. Separate logic into `controller.go`, `service.go`, and `repository.go` if applicable.
4. Document any new environment variables in `.env.example`.
