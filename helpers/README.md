# 🛠️ Helpers

This directory contains logic that supports the API's global functionality, including response handling, session management, and common utilities.

## 📁 Helpers Overview

### 📨 Response Handling (`response.go`)
- **Standardized Response Format**: Unified success/error JSON structures.
- **Error Codes**: Mapping internal errors to HTTP status codes.
- **Pagination Helper**: Logic to wrap results with current page, total records, and other metadata.

### 🔑 Session Management (`sessions.go`)
- **JWT Middleware**: Token extraction, verification, and claims handling.
- **Session Store**: Logic to manage active sessions and integrate with Redis for token blacklisting/refreshing.
- **Auth Context**: Storing authenticated user data in the Gin request context.

### 📄 Document & Data Utilities
- **PDF Generation (`generate_pdf`)**: Helper to convert dynamic HTML templates into PDF documents.
- **Utilities (`utilities.go`)**: Small, reusable functions for string manipulation, data casting, and validation.

---

## 🛠️ Usage Guidelines
- These helpers are designed to be used globally across all modules.
- Prefer using `helpers.SuccessResponse(c, data)` and `helpers.ErrorResponse(c, err)` for consistent API interaction.
