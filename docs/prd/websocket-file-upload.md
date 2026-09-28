# PRD: WebSocket File Upload (S3/GCS) with Short-Lived Token Auth

- Status: Draft
- Owner: TBD
- Related module: `modules/storages`
- Author: generated with Claude Code from codebase research on 2026-09-28

## 1. Problem / Motivation

Uploads currently go through `POST /storages/upload` (`modules/storages/controller.go`), a
standard multipart REST endpoint. The user wants a WebSocket-based upload path instead —
presumably for progress streaming / long-lived connections / chat-attachment style flows —
that still ends up in S3 or GCS like the REST path is *supposed* to.

Because a WebSocket upload session is authenticated differently than a normal REST session
(the connection is long-lived, and re-authenticating per message is undesirable), this PRD
also proposes a short-lived token pair (access ~5 min / refresh ~30 min) stored in its own
Redis logical DB, separate from the existing session-based auth.

## 2. Current State (as of this PRD, verified in code)

This section documents what exists today, so design decisions below aren't made blind.

### 2.1 REST upload endpoint

- Route: `POST /storages/upload` → `UploadAction` (`modules/storages/controller.go`),
  protected by `authHandler` middleware. Routes registered in `modules/storages/registrator.go:9-25`.
- `UploadAction` calls `ctx.Request.ParseMultipartForm(8 << 20)` — **hardcoded 8 MB limit**,
  no other size/type validation at the controller layer.
- `Uploads` (`modules/storages/service.go:32-72`) resolves the current user, then for each
  file calls `storages.NewStorageBase(file, fileType[0])` (from the vendored
  `github.com/fari-99/go-helper/storages` package) `.UploadFiles()`, and batch-inserts
  `models.Storages` rows.
- `POST /storages/s3-policy` (`S3Policy`) exists as a route but is an unimplemented stub
  (`controller.go:55-58`).

### 2.2 S3/GCS storage backend — important gap

All storage-backend logic lives in the **vendored** `vendor/github.com/fari-99/go-helper/storages`
package (`base.go`, `s3.go`, `gcs.go`, `local.go`). Backend selection is driven by which
setter the caller invokes on `StorageBase`: `SetAwsS3(...)`, `SetGoogleGCS(...)`, or neither
(falls back to local disk, `LOCAL_STORAGE_PATH`).

**`modules/storages/service.go` never calls `SetAwsS3` or `SetGoogleGCS`.** A repo-wide grep
outside `vendor/` confirms zero call sites. This means **the current REST upload endpoint only
ever writes to local disk**, even though:
- `S3_ENABLE="false"` and `GCS_ENABLED=false` exist in `.env`, but neither is read anywhere
  in the Go code — they're vestigial.
- The vendored library fully implements S3 (aws-sdk-go-v2 + `transfermanager`) and GCS
  (`cloud.google.com/go/storage`) upload/download/presign paths.

**Implication for this PRD:** "upload to S3/GCS like the REST upload does" is not actually
true today for the REST endpoint. Wiring real S3/GCS support (env-driven backend selection)
is a prerequisite shared by both the REST and the new WebSocket path, not just a WS-specific
task.

### 2.3 Auth / session (existing REST auth)

- Framework: Gin (`github.com/gin-gonic/gin`).
- JWT via `github.com/golang-jwt/jwt/v5`, wrapped by vendored `github.com/fari-99/go-helper/token_generator`.
  **Expiry granularity is whole days, hardcoded** (`getExpiredDate`, `jwt.go:257-269` uses
  `AddDate(0,0,n)`); the only override (`setExpired`) is unexported and unused. The `.env`
  vars `JWT_ACCESS_TOKEN_EXPIRED`/`_TYPE`, `JWT_REFRESH_TOKEN_EXPIRED`/`_TYPE` are dead —
  never read in Go code. **The existing JWT wrapper cannot issue a 5-minute or 30-minute
  token as-is.**
- Sessions are JWT + Redis-backed (`helpers/sessions.go`, `modules/middleware/middleware_auth.go`):
  `authServe` validates the JWT *and* checks a live Redis session (`GET {uuid}:access_token`),
  so revoking the Redis key invalidates the token before JWT expiry. Refresh flow additionally
  does reuse detection via a `{uuid}:family` record (`SetFamily`/`CheckFamily`).
- Redis: `modules/configs/redis.go` already has a **generic, prefix-driven** multi-DB
  convention — `GetRedis(prefix)` reads `{PREFIX}_HOST/PORT/PASSWORD/DB/...` and returns a
  singleton client per prefix. Currently configured prefixes (`global.env`):
  `REDIS_CACHE_DB=1`, `REDIS_SESSION_DB=2`, `REDIS_COUNTING_DB=3`, `REDIS_LOCK_DB=4` — all on
  the same Redis instance/host, just different logical DB indices. **DB 0 and DB 5+ are free.**

### 2.4 WebSocket / library availability

- **No WebSocket dependency exists today.** `github.com/coder/websocket` is present only as
  an indirect dependency, pulled in transitively by `go.mau.fi/whatsmeow` (WhatsApp Web
  client) as an *outbound* client to WhatsApp's servers — not usable as a server-side pattern
  and not exposed as an endpoint of this API.
- Gin has no built-in WebSocket support; a server-side WS endpoint needs a new direct
  dependency (`gorilla/websocket` is the most common Gin pairing) and manual
  `http.Hijacker`-based upgrade from `*gin.Context`.

### 2.5 Module conventions

Standard module layout (`controller.go`, `service.go`, `repository.go`, `requests.go`,
`registrator.go`), wired centrally in `cmd/servers/main/main.go`. A new capability would
either extend `modules/storages` or become a new module (e.g. `modules/ws_uploads` or
`modules/auth_tokens`) — see Open Questions.

## 3. Goals

1. Add a WebSocket endpoint that accepts file uploads and stores them in S3 or GCS (backend
   selectable via config), producing the same `models.Storages` records the REST path
   produces today.
2. Add a short-lived token auth scheme for WS upload sessions:
   - Access token: ~5 minute TTL.
   - Refresh token: ~30 minute TTL.
   - Stored in a **new, separate Redis logical DB** from the existing session DB, following
     the existing `GetRedis(prefix)` convention.
3. Fix the underlying gap in §2.2 so S3/GCS is actually wired up (config-driven backend
   selection), since the WS path depends on it and the REST path currently doesn't get it
   either.
4. Keep behavior/response shape compatible enough that existing consumers of `models.Storages`
   (image resize endpoint, other modules referencing `storages` records) keep working.

## 4. Non-Goals

- Replacing or modifying the existing REST session auth (`helpers/sessions.go`,
  `middleware_auth.go`) used by the rest of the API. The new token scheme is additive and
  scoped to WS upload only, unless the user later decides to generalize it.
- Resumable/chunked upload with server-side persistence across reconnects (out of scope for
  v1 — see Open Questions on whether basic chunking-for-progress is still wanted).
- Implementing the `S3Policy` presigned-POST stub (`controller.go:55-58`) — unrelated to this
  effort unless we decide the WS flow should use presigned uploads instead of proxying bytes
  through the API process (see §6.4).
- Virus/malware scanning pipeline (flagging as a possible follow-up, not in scope here).

## 5. Proposed Design

### 5.1 High-level flow

1. Client authenticates normally (existing REST login) and additionally requests a WS
   upload token pair from a new endpoint, e.g. `POST /storages/ws-token`
   (auth-protected by the existing session middleware).
2. Server issues `{ ws_access_token (5m), ws_refresh_token (30m) }`, persists them in the new
   Redis DB keyed by a session id, mirroring the existing key pattern
   (`{id}:ws_access_token`, `{id}:ws_refresh_token`) with `EX` set to the TTL.
3. Client opens `GET /ws/storages/upload?token=<ws_access_token>` (or passes the token as the
   first WS message — see Open Questions on transport). Server validates the token against
   Redis (not just JWT signature/exp) before completing the upgrade, same "JWT + live Redis
   check" pattern as `authServe`.
4. Client streams file(s) over the socket using a small framed protocol (see §5.3).
5. Server buffers/streams each file to the configured backend (S3, GCS, or local — same
   `StorageBase` abstraction, now actually wired), inserts `models.Storages` rows the same
   way `Uploads` does today, and sends an ack message per file with the resulting storage
   record(s).
6. If the access token expires mid-session (5 min may be short for large/slow uploads), the
   server closes the socket with a defined close code; the client uses the refresh token
   against a REST refresh endpoint to get a new access token and reconnects.

### 5.2 Auth token design

- New Redis prefix, e.g. `REDIS_WS_AUTH` (or `REDIS_AUTH_TOKEN` if we want it reusable beyond
  WS), added to `global.env` with `_DB=5` (next free index) and the same
  `_HOST/_PORT/_PASSWORD/_TIMEOUT/_MIN_IDLE` fields as the other prefixes, consumed via the
  existing `configs.GetRedis("REDIS_WS_AUTH")`.
- Token issuance cannot reuse `token_generator.NewJwt` as-is (day-only granularity, §2.3).
  Options:
  a. Extend the vendored wrapper to accept a duration instead of day-count (touches a
     third-party vendored module — invasive, but consistent long-term).
  b. Issue these tokens with `golang-jwt/jwt/v5` directly (already a direct dependency),
     bypassing the `token_generator` wrapper entirely for this token type, with explicit
     `ExpiresAt` in minutes.
  Recommendation: (b) — isolates the new short-lived scheme from the existing wrapper's
  assumptions instead of risking regressions in the day-granularity path used everywhere
  else. Decision needed from the team (see Open Questions).
- Redis value: minimal JSON blob (uuid, user id, issued-at) — same shape philosophy as
  `SessionRedisData`, TTL = `EX 300` (access) / `EX 1800` (refresh).
- Refresh should rotate the access token and, following the existing reuse-detection
  precedent (`SetFamily`/`CheckFamily`), optionally rotate the refresh token too — worth
  deciding whether that complexity is justified for a 30-minute-lifetime token.

### 5.3 WebSocket wire protocol (proposed, for review)

Text/JSON control frames + binary data frames over the same connection:

```
→ {"type":"start","file_name":"...","file_type":"...","size":12345,"mime":"..."}
→ <binary chunk>
→ <binary chunk>
→ {"type":"end"}
← {"type":"progress","received":8192}
← {"type":"ack","storage":{...models.Storages...}}
← {"type":"error","message":"..."}
```

Multiple files per connection = repeat `start`/binary/`end`. Connection-level max message
size, max total bytes, and idle timeout need concrete numbers (see Open Questions).

### 5.4 Storage backend wiring (prerequisite fix)

Introduce a config-driven backend selector (e.g. `STORAGE_DRIVER=local|s3|gcs` in `.env`),
read once at startup/DI (`modules/configs`), and have both the REST `Uploads` service and the
new WS upload service call the appropriate `SetAwsS3`/`SetGoogleGCS` on `StorageBase` based on
that driver, instead of silently defaulting to local disk. This is required for the WS path
to actually reach S3/GCS, and incidentally fixes the same gap for the existing REST endpoint.

### 5.5 Library choice

No WS library is currently a direct dependency. Recommend adding `gorilla/websocket` (most
common, best-documented pairing with Gin's `http.Hijacker`-based upgrade) as a new direct
dependency, rather than promoting the indirect `coder/websocket` (pulled in for an unrelated
outbound WhatsApp client use case).

## 6. Open Questions (need decisions before implementation)

1. **Token transport on WS handshake**: query param (`?token=`, simplest, but tokens can leak
   into access logs/proxies) vs. `Sec-WebSocket-Protocol` header trick vs. first-message auth
   after an unauthenticated upgrade (delays rejection but avoids token-in-URL). Which is
   acceptable given existing infra (reverse proxy logging, etc.)?
2. **New Redis prefix name**: `REDIS_WS_AUTH` (scoped to WS) vs. `REDIS_AUTH_TOKEN` (generic,
   reusable if short-lived tokens are wanted elsewhere later)?
3. **JWT issuance approach**: extend vendored `token_generator` for minute-granularity vs.
   bypass it with direct `golang-jwt/jwt/v5` calls for this token type only (recommended in
   §5.2, but changes precedent — confirm)?
4. **New module vs. extending `modules/storages`**: does this live inside `storages` (new
   `ws_controller.go`/`ws_registrator.go`) or as a new `modules/ws_uploads` +
   `modules/auth_tokens` pair? Affects DI wiring in `cmd/servers/main/main.go`.
5. **Proxy vs. presigned upload**: should the server proxy raw bytes to S3/GCS (simpler,
   matches current REST behavior, but ties API process resources to upload duration), or
   should the WS flow hand back a presigned S3/GCS URL and only use the socket for
   progress/metadata? (This overlaps with the unimplemented `S3Policy` stub.)
6. **Access token TTL vs. upload duration**: 5 minutes may be too short for large files on
   slow connections — do we want a grace period, a keepalive/renew-in-place message type, or
   is "reconnect with refresh token" an acceptable UX for that case?
7. **Refresh token rotation**: rotate-on-use with reuse detection (like the existing session
   `family` mechanism) or a simpler non-rotating refresh token given the already-short
   30-minute lifetime?
8. **File size / type limits for WS uploads**: reuse the REST path's ad-hoc 8 MB multipart
   cap, or define new explicit limits (max file size, max total bytes per connection, allowed
   MIME types) now that this is a new endpoint anyway?
9. **Concurrency**: max concurrent WS upload connections per user, and per-connection max
   concurrent in-flight files?

## 7. Security Considerations

- WSS (TLS) only in production; document this as a hard requirement, not just a
  recommendation.
- Validate the access token against Redis on both handshake **and** periodically during long
  connections (mirrors `authServe`'s "JWT + live Redis" double-check so a revoked/killed
  session can't keep uploading).
- Origin/CSRF-equivalent checks on the WS upgrade (Gin/gorilla don't check `Origin` by
  default — must be explicit).
- Enforce message/frame size limits at the WS layer, independent of any S3/GCS-side limits,
  to prevent memory exhaustion from a malicious `start` message with an unbounded `size`.
- Keep the same content-type sniffing behavior already present in the vendored storage lib
  (`http.DetectContentType`) rather than trusting the client-supplied `mime` field for
  anything security-relevant.
- Rate-limit token issuance (`POST /storages/ws-token`) to prevent Redis exhaustion via mass
  token creation.

## 8. Rollout / Milestones (proposed)

1. Wire real S3/GCS backend selection for the existing REST upload path (§5.4) — independently
   valuable, de-risks the WS work, and is currently silently broken.
2. Add the new Redis prefix + short-lived token issuance/validation (REST endpoints only:
   issue + refresh), no WS yet — testable in isolation.
3. Add the WS upgrade endpoint with token validation, no file handling yet (connect/auth
   round-trip only).
4. Add the upload wire protocol + S3/GCS write path + `models.Storages` insertion.
5. Load/size-limit testing, then rollout behind a feature flag if the project has one.

## 9. Appendix — Key File References

| Area | File |
|---|---|
| REST upload controller | `modules/storages/controller.go` |
| REST upload service | `modules/storages/service.go` |
| Storage routes | `modules/storages/registrator.go` |
| Storage model | `modules/models/storages.go` |
| Vendored storage backends | `vendor/github.com/fari-99/go-helper/storages/{base,s3,gcs,local}.go` |
| JWT wrapper | `vendor/github.com/fari-99/go-helper/token_generator/{base,jwt}.go` |
| Session helpers | `helpers/sessions.go` |
| Auth middleware | `modules/middleware/middleware_auth.go` |
| Redis client factory | `modules/configs/redis.go` |
| Redis config (env) | `global.env` (repo root) |
| S3/GCS/JWT env vars | `.env`, `.env.example` |
| Composition root | `cmd/servers/main/main.go` |
