# 📦 Packages (Pkg)

These internal packages are shared across the application to provide common functionality and helpers. They are designed to be decoupled and reusable across various modules.

## 📁 Packages Overview

### 🔐 OTP Helper (`otp_helper`)
- **Generation**: Secure generation of numeric and alphanumeric one-time passwords.
- **Validation**: Time-based validation logic with support for multiple delivery channels.
- **Integration**: Works seamlessly with `twoFA` and `notifications` modules.

### 💾 Redis Helpers (`redis_helpers`)
- **Counters**: Generic Redis-based counters for rate limiting and frequency tracking.
- **Distributed Locking**: Wrappers for `redsync` to ensure high availability and race-condition prevention.
- **Distributed Lock Test**: Comprehensive tests for concurrent locking scenarios.

### 🎥 GO2RTC Helper (`go2rtc_helper`)
- **Streaming Integration**: Communication with `go2rtc` for streaming video and managing camera feeds.
- **Synchronization**: Syncing database camera entities with real-time stream servers.

### ⚙️ Utilities
- **Once**: Utilities to ensure a block of code runs only once (e.g., initialization logic).
- **Pagination**: Standardized pagination structure for consistent API responses.

---

## 🛠️ Usage Guidelines
1. **Importing**: Use `go-api/pkg/<package_name>` to import these helpers.
2. **Statelessness**: Keep packages as stateless as possible, preferring configuration-based initialization.
3. **Tests**: Each package should have its own unit tests within the same directory.
