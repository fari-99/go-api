# PRD: WebSocket File Upload (S3/GCS) with Short-Lived Token Auth

- Status: Implemented v1 (see §11 for what was built and how it was verified)
- Owner: TBD
- Related module: `modules/storages`
- Author: generated with Claude Code from codebase research on 2026-09-28, revised 2026-09-30 with owner decisions

## 1. Problem / Motivation

Uploads currently go through `POST /storages/upload` (`modules/storages/controller.go`), a
standard multipart REST endpoint. The user wants a WebSocket-based upload path instead —
presumably for progress streaming / long-lived connections / chat-attachment style flows —
that ends up in S3 or GCS like the REST path is *supposed* to.

Because a WebSocket upload session is authenticated differently than a normal REST session
(the connection is long-lived, and re-authenticating per message is undesirable), this PRD
also proposes a short-lived token pair (access ~5 min / refresh ~30 min) stored in its own
Redis (separate prefix, and optionally a separate Redis instance) from the existing
session-based auth.

## 2. Current State (verified in code)

### 2.1 REST upload endpoint

- Route: `POST /storages/upload` → `UploadAction` (`modules/storages/controller.go`),
  protected by `authHandler` middleware. Routes registered in `modules/storages/registrator.go:9-25`.
- `UploadAction` calls `ctx.Request.ParseMultipartForm(8 << 20)` — **hardcoded 8 MB limit**,
  no other size/type validation at the controller layer.
- `Uploads` (`modules/storages/service.go:32-72`) resolves the current user, then for each
  file calls `storages.NewStorageBase(file, fileType[0])` (vendored
  `github.com/fari-99/go-helper/storages`) `.UploadFiles()`, and batch-inserts
  `models.Storages` rows.
- `POST /storages/s3-policy` (`S3Policy`) exists as a route but is an unimplemented stub
  (`controller.go:55-58`).

### 2.2 S3/GCS storage backend — important gap

All storage-backend logic lives in the vendored `vendor/github.com/fari-99/go-helper/storages`
package (`base.go`, `s3.go`, `gcs.go`, `local.go`). Backend selection is driven by which
setter the caller invokes on `StorageBase`: `SetAwsS3(...)`, `SetGoogleGCS(...)`, or neither
(falls back to local disk, `LOCAL_STORAGE_PATH`).

**`modules/storages/service.go` never calls `SetAwsS3` or `SetGoogleGCS`.** A repo-wide grep
outside `vendor/` confirms zero call sites, so **the current REST upload endpoint only ever
writes to local disk**. `S3_ENABLE="false"` and `GCS_ENABLED=false` in `.env` are not read
anywhere in Go code.

**Implication:** wiring env-driven backend selection is a prerequisite shared by REST and WS.

**Vendored API shape constraint (relevant to the "no new wheel" rule, see §5.4):**
`NewStorageBase` takes a `*multipart.FileHeader`, and `s3Upload`/`gcsUpload` take a
`multipart.File`. Both first copy the whole body into a local temp file, then upload the temp
file. A WebSocket stream is not a `multipart.FileHeader`, so the vendored `UploadFiles()`
cannot be called directly with WS data without faking a multipart envelope and double-buffering
to disk.

### 2.3 Auth / session (existing REST auth)

- Framework: Gin. JWT via `github.com/golang-jwt/jwt/v5`, wrapped by vendored
  `github.com/fari-99/go-helper/token_generator`.
- **Update (go-helper v1.6.3, verified in vendor):** `BaseJwt.SetExpired(expiredAccess, accessType,
  expiredRefresh, refreshType)` is now exported, and `getExpiredDate` now computes
  `time.Duration(n) * type` instead of whole-day `AddDate`. Minute-granularity is therefore
  supported: `SetExpired(5, time.Minute, 30, time.Minute)`. Constraint: it returns an error if
  access TTL > refresh TTL. The old "cannot issue 5/30-minute tokens" limitation no longer applies.
- Sessions are JWT + Redis-backed (`helpers/sessions.go`, `modules/middleware/middleware_auth.go`):
  `authServe` validates the JWT *and* a live Redis session (`GET {uuid}:access_token`), so
  revoking the Redis key invalidates the token before JWT expiry. Refresh does reuse detection
  via a `{uuid}:family` record (`SetFamily`/`CheckFamily`, `helpers/sessions.go:398-420`).
- Redis: `modules/configs/redis.go` has a generic prefix-driven convention —
  `GetRedis(prefix)` reads `{PREFIX}_HOST/PORT/PASSWORD/DB/...` and returns a singleton per
  prefix. Existing prefixes: `REDIS_CACHE_DB=1`, `REDIS_SESSION_DB=2`, `REDIS_COUNTING_DB=3`,
  `REDIS_LOCK_DB=4`, all on one instance. Because host/port are per-prefix, a new prefix can
  point at a different Redis instance with no code change.

### 2.4 WebSocket / library availability

- No direct WebSocket dependency today. `github.com/coder/websocket v1.8.15` is already an
  indirect dependency **and is already vendored** (`vendor/modules.txt`), pulled in by
  `go.mau.fi/whatsmeow`.
- Gin has no built-in WebSocket support; the endpoint needs a library that upgrades from
  `*gin.Context`'s `Writer`/`Request`.

### 2.5 Encryption helper (already available)

Vendored `go-helper/crypts` already provides AES-GCM `Encrypt`/`Decrypt` with passphrase and
URL-safe base64 output (`SetPassphrase`, `SetEncodeBase64(true)`, `SetUseRandomness(true, ...)`).
No new crypto code is needed for token encryption.

### 2.6 Module conventions

Standard module layout (`controller.go`, `service.go`, `repository.go`, `requests.go`,
`registrator.go`), wired in `cmd/servers/main/main.go`.

## 3. Goals

1. WebSocket endpoint that accepts file uploads and stores them in S3 or GCS (backend
   selectable via config), producing the same `models.Storages` records as the REST path.
2. Short-lived token auth for WS upload sessions (access ~5 min, refresh ~30 min), stored under
   its own Redis prefix, with tokens **encrypted** before being handed to the client.
3. Fix the §2.2 gap: config-driven S3/GCS/local selection shared by REST and WS.
4. Keep `models.Storages` output compatible with existing consumers (image resize, etc.).
5. Env-configurable size limit (MB) and per-user concurrent upload cap.

## 4. Non-Goals

- Replacing/modifying the existing REST session auth. The new scheme is additive, WS-upload only.
- Resumable/chunked upload persistence across reconnects.
- Presigned direct-to-bucket uploads and the `S3Policy` stub (deferred, see §6.5).
- Virus/malware scanning.

## 5. Proposed Design

### 5.1 High-level flow

1. Client logs in via the existing REST flow, then calls `POST /storages/ws-token`
   (protected by existing session middleware).
2. Server issues `{ws_access_token (5m), ws_refresh_token (30m)}`, both **encrypted opaque
   strings** (§5.2), and stores the state in the new auth Redis.
3. Client opens `GET /ws/storages/upload` and presents the access token (§5.3 transport). A
   WS-specific middleware decrypts it, verifies JWT signature/exp, and checks the live Redis
   record before the upgrade completes.
4. Client streams files using the protocol in §5.6.
5. Server streams each file straight to the configured backend (§5.4), inserts
   `models.Storages` rows the same way `Uploads` does, and acks each file.
6. Before the access token expires, the client sends a `renew` message on the same socket
   (§5.5) and receives a fresh pair without reconnecting.

### 5.2 Auth token design

- **Issuance:** use vendored `token_generator` with `SetExpired(5, time.Minute, 30, time.Minute)`
  (now possible, §2.3). Prefer TTLs from env (`WS_ACCESS_TOKEN_TTL_MINUTES=5`,
  `WS_REFRESH_TOKEN_TTL_MINUTES=30`). No hand-rolled JWT code.
- **Encryption (decision: encrypt the whole signed JWT):** after signing, encrypt the entire JWT
  string with `crypts` (AES-GCM, URL-safe base64) using a dedicated key
  `WS_TOKEN_ENCRYPTION_KEY` (separate from JWT signing secrets). The client only ever sees
  an opaque blob.
  - Why whole-JWT instead of only the `data` claim: the client can't read *or* fingerprint any
    claim (uuid, user id, exp), a tampered/forged blob fails at GCM auth before any JWT parsing,
    and it's one encrypt/decrypt call at each boundary. Trade-off: the token is bigger and not
    parseable by standard JWT tooling — acceptable since it's a private WS-only scheme.
  - **Decrypt middleware** (`modules/middleware`, e.g. `wsAuthHandler`): read token → `Decrypt`
    → `token_generator` parse/verify → check Redis record → set user in Gin context. Any
    failure returns a generic 401 (no oracle on which step failed).
  - Key rotation: support an optional `WS_TOKEN_ENCRYPTION_KEY_PREVIOUS` for decrypt-only
    fallback so rotating the key doesn't invalidate live 30-minute sessions.
- **Redis (decision: separate prefix and, ideally, separate environment):** new prefix
  `REDIS_WS_AUTH` with its own `_HOST/_PORT/_PASSWORD/_DB/_TIMEOUT/_MIN_IDLE`, consumed via
  `configs.GetRedis("REDIS_WS_AUTH")`. Because host/port are per-prefix, it can point at a
  dedicated Redis instance in production and fall back to the shared instance (`_DB=5`) in
  local/dev. Keys:
  - `{uuid}:ws_access_token`, `{uuid}:ws_refresh_token` → JSON `{uuid, user_id, issued_at}`,
    `EX` = TTL.
  - `{uuid}:ws_family` → refresh reuse detection (§5.5).
  - Concurrency counters (§5.7) live here too (or in `REDIS_COUNTING_DB` if the team prefers
    to reuse the existing counting DB; leaning to keep everything WS-related in one place).

### 5.3 Token transport on handshake

**Decision: browser clients are the priority.** The browser `WebSocket` API cannot set an
`Authorization` header, so the encrypted access token is passed as a query param on the
handshake: `wss://.../ws/storages/upload?token=<encrypted_access_token>`.

- The token is only used for the handshake. After that, renewal happens in-band (§5.5), so a
  long session never re-sends a token in a URL.
- Mitigations for URL exposure: the token is encrypted (opaque), expires in 5 minutes, and is
  backed by live Redis state (revocable). Reverse proxy / app access logs must redact the
  `token` query param.
- Non-browser clients may additionally send the token via the `Authorization` header; the
  middleware accepts either, header first.

### 5.4 Storage backend wiring and the new streaming package

- **Backend selector:** `STORAGE_DRIVER=local|s3|gcs` in `.env`, read once in `modules/configs`.
  The REST `Uploads` service calls `SetAwsS3(nil)` / `SetGoogleGCS(nil)` accordingly (the
  vendored setters already read `S3_*` / `GCS_*` env when passed `nil`). This fixes the REST gap.
- **Reuse rule (decision):** create a new package only where the vendored function can't do
  the same job; otherwise call the vendored code.
  - *REST path:* keep using vendored `NewStorageBase(...).UploadFiles()` — same function, no
    new code.
  - *WS path:* vendored `UploadFiles()` requires a `*multipart.FileHeader` and always spools to
    a temp file (§2.2), so it does **not** fit a byte stream. Add a new package
    (e.g. `pkg/wsstorage` or `modules/storages/wsupload`) that:
    - takes an `io.Reader` (fed from the socket via `io.Pipe`) plus filename/type/size metadata;
    - for S3, calls the `transfermanager`/`manager.UploadObject` directly with `Body` as the
      reader (no temp file);
    - for GCS, `client.Bucket(...).Object(...).NewWriter(ctx)` + `io.Copy` from the reader;
    - for local, streams to `LOCAL_STORAGE_PATH` with the same path/filename scheme;
    - **is self-contained and upstream-ready (decision):** the S3/GCS client init
      (`s3Config()`/`gcsInit()` equivalents), streaming upload, and setup structs are written
      as a standalone package with no imports from this repo's modules (only stdlib + the
      AWS/GCS SDKs already vendored), so the owner can move it into go-helper later with no
      refactor. Anything reusable and *exported* from the vendored `storages` package
      (`S3Setup`, `GCSSetup`, `StorageData`) is reused as-is; unexported logic (client init,
      storage path/filename generation) is re-implemented in the new package, keeping the
      same path/filename scheme so records stay compatible with REST-created ones.
    - MIME: sniff the first 512 bytes with `http.DetectContentType` (same as vendored
      behavior), do not trust the client `mime`.
  - **Image handling (decision): store as-is.** The vendored REST path decodes images and
    re-encodes to JPEG (`IsImage` branch); the WS path does not. Images are streamed to the
    bucket unchanged, keeping the sniffed MIME. Consequence: WS-uploaded images are not
    JPEG-normalized/scaled like REST ones, so consumers such as the resize endpoint must not
    assume JPEG for these records.

### 5.5 Renewal, keepalive and refresh rotation

- **Renew in place (decision):** client sends
  `{"type":"renew","refresh_token":"<encrypted>"}` on the open socket. Server decrypts,
  validates the refresh record, rotates, and replies
  `{"type":"renewed","access_token":"...","refresh_token":"...","access_expires_at":"..."}`.
  The socket and in-flight uploads are untouched.
- **Keepalive:** server sends WS ping frames on an interval; client may also send
  `{"type":"ping"}`. Ping alone does **not** extend auth — only `renew` does.
- **In-flight uploads vs. expiry:** access token is checked at handshake, on each `start`,
  and on a periodic timer (also re-checking that the Redis record still exists so a revoked
  session stops uploading). A file already in progress is allowed to finish (bounded by the
  size cap); a new `start` with an expired access token gets
  `{"type":"error","code":"token_expired"}` and the client must `renew`. If neither renew nor a
  valid token arrives within a grace period (env, default 30 s) after expiry, the server
  closes the socket with a defined close code.
- **Refresh rotation (decision: keep reuse detection):** every `renew` (and the REST refresh
  endpoint) issues a new refresh token and marks the old one used in `{uuid}:ws_family`,
  mirroring `SetFamily`/`CheckFamily`. Presenting an already-used refresh token is treated as
  theft: revoke the whole family (all WS tokens for that chain), close any open sockets of that
  family, return 401.

### 5.6 WebSocket wire protocol

Text frames are JSON control messages; binary frames are data. Every upload has a client-chosen
numeric `id` (uint32) so several files can be in flight on one socket (up to the §5.7 cap).

**Binary frame layout:** `[4 bytes big-endian upload id][payload]`. Max frame size is
`WS_UPLOAD_MAX_CHUNK_KB` + 4 bytes.

```
→ {"type":"start","id":1,"file_name":"a.png","file_type":"avatars","size":12345}
→ <binary: id=1 + chunk>   (repeat)
→ {"type":"end","id":1}
→ {"type":"abort","id":1}
→ {"type":"renew","refresh_token":"<encrypted>"}
→ {"type":"ping"}
← {"type":"started","id":1}
← {"type":"progress","id":1,"received":8192}
← {"type":"ack","id":1,"storage":{...models.Storages...}}
← {"type":"renewed","access_token":"...","refresh_token":"...","access_expires_at":"...","refresh_expires_at":"..."}
← {"type":"pong"}
← {"type":"error","id":1,"code":"...","message":"..."}
```

`size` is optional (0 = unknown); when given it must match the bytes received at `end`.
`file_type` is a storage folder name and must match `[A-Za-z0-9_-]{1,64}`. The client `mime` is
ignored; MIME is sniffed server-side.

Error codes: `token_expired`, `too_many_uploads`, `file_too_large`, `size_mismatch`,
`invalid_file_type`, `invalid_file_name`, `duplicate_id`, `unknown_upload`, `empty_file`,
`upload_stalled`, `connection_limit`, `bad_message`, `unknown_type`, `upload_failed`,
`internal_error`, `unauthorized`.

Close codes: `4401` session revoked / refresh token reused / bad renew, `4408` access token
expired and not renewed (only closed when no upload is in flight), `1009` connection byte cap.

Backpressure: the read loop is single-threaded, so a slow storage backend slows the client; if a
chunk can't be queued within `WS_UPLOAD_STALL_SECONDS` that upload is aborted (`upload_stalled`).

### 5.7 Limits and concurrency (decision: env-driven)

| Env var | Default | Meaning |
|---|---|---|
| `WS_UPLOAD_MAX_FILE_MB` | `8` (matches REST today) | Max size per file, in MB. Compared against **bytes actually received**, not the declared `size`; upload aborted and partial object deleted on breach. |
| `WS_UPLOAD_MAX_CONNECTION_MB` | unset (no cap) | Optional total bytes per connection; not required for v1. |
| `WS_UPLOAD_MAX_CONCURRENT` | `3` | Max simultaneous in-flight uploads **per user**. |
| `WS_ACCESS_TOKEN_TTL_MINUTES` / `WS_REFRESH_TOKEN_TTL_MINUTES` | `5` / `30` | Token lifetimes. |
| `WS_TOKEN_ENCRYPTION_KEY` (+ `_PREVIOUS`) | — | Token encryption (§5.2). |
| `WS_RENEW_GRACE_SECONDS` | `30` | Grace after access expiry before close. |

The size limit is read from env at startup and converted to bytes (`MB * 1024 * 1024`); the WS
library's per-message read limit is set from the same value's chunk size, not the file size.

**Concurrency counter:** Redis key `ws_upload:active:{user_id}` in the WS auth Redis.
- On `start`: `INCR`; if the result exceeds `WS_UPLOAD_MAX_CONCURRENT`, `DECR` and reply
  `error: too_many_uploads`.
- On finish/abort/socket close: `DECR` in a `defer` (guarded so a file is decremented once).
- **Decision: simple `INCR`/`DECR`.** Crash safety: set a TTL on the key (refreshed on each
  `start`) so a killed process can't leave a user permanently locked out. Never let the counter
  go below zero (clamp/reset on negative).
- Keyed by **user id**, not connection or token uuid, so opening more sockets doesn't bypass
  the cap.

### 5.8 Library choice (decision: `github.com/coder/websocket`)

Searched for the simplest well-maintained option:

| Library | Notes |
|---|---|
| `github.com/coder/websocket` (formerly `nhooyr.io/websocket`) | Minimal API, `context`-aware read/write, built-in `SetReadLimit`, concurrent-safe writes, first-class `net/http` (works from Gin via `c.Writer, c.Request`), actively maintained by Coder, has proper close-code handling and permessage-deflate. **Already in `go.mod` (indirect, v1.8.15) and vendored.** |
| `github.com/gorilla/websocket` | Most widely known, revived maintenance, but older callback/deadline-style API, no context support, one concurrent writer rule to manage manually. |

**Recommendation: `coder/websocket`.** Promote it from indirect to direct in `go.mod`
(`go mod tidy` + `go mod vendor`); no new module to vendor and no extra supply-chain surface.
Use `websocket.Accept` with explicit `OriginPatterns` (§8).

## 6. Decisions Log (previously open questions)

| # | Question | Decision |
|---|---|---|
| 1 | Token transport | Browsers are the priority: encrypted token in `?token=` on handshake, header also accepted for non-browser clients (§5.3). |
| 2 | Redis prefix | Separate `REDIS_WS_AUTH` prefix, separate Redis env/instance where possible (§5.2). |
| 3 | JWT issuance | Use vendored `token_generator` — `SetExpired` is now public in go-helper v1.6.3. Direct `golang-jwt` no longer needed. |
| 4 | Module placement | Extend `modules/storages` for WS controller/service; token issuance + middleware in a small `modules/ws_auth` (or under `middleware`). Final layout at implementation time. |
| 5 | Proxy vs presigned | **Proxy** (stream through the API to S3/GCS). See rationale below. |
| 6 | Token TTL vs upload duration | `renew` message in place (§5.5); in-flight file may finish after expiry. |
| 7 | Refresh rotation | Rotate on use **with** reuse detection (§5.5). |
| 8 | Limits | Env-driven, in MB (§5.7). |
| 9 | Concurrency | Env-driven, default 3 per user, Redis counter by user id (§5.7). |

**Proxy vs presigned rationale (owner was unsure):** proxy is recommended for v1 because:
(a) size limit, MIME sniffing and the concurrency cap are enforced server-side on real bytes —
with presigned URLs the client uploads directly and the server can only constrain via policy
conditions, then must trust a "done" callback to create the `models.Storages` row;
(b) the point of the feature is a WS upload path with progress, which presigned URLs would
reduce to a metadata channel; (c) the cost (API process handles bytes for the upload duration)
is bounded by the 3-per-user cap and the MB limit, and streaming avoids temp files/RAM
spikes. Revisit presigned if uploads move to very large files (hundreds of MB+) or API
bandwidth becomes the bottleneck; it would live alongside the existing `S3Policy` stub.

## 7. Resolved Follow-ups

- Image handling: store as-is (§5.4).
- Default `WS_UPLOAD_MAX_FILE_MB`: 8; no per-connection cap in v1 (§5.7).
- Concurrency counter: simple `INCR`/`DECR` with TTL (§5.7).
- Browser clients: priority; token via query param (§5.3).
- S3/GCS streaming code: new self-contained package, to be moved into go-helper later (§5.4).

No open questions remain blocking implementation.

## 8. Security Considerations

- WSS (TLS) only in production — hard requirement.
- Tokens are encrypted (AES-GCM) and signed (JWT) and backed by live Redis state; all failures
  return a generic 401.
- Re-validate token + Redis record periodically on long connections so a revoked session stops
  uploading.
- Explicit `Origin` allow-list on the upgrade (`OriginPatterns`); no wildcard in production.
- Enforce per-message read limit at the WS layer and count real bytes against the env size
  cap; never trust the `start.size` field.
- Sniff MIME server-side (`http.DetectContentType`); do not trust client `mime`.
- Sanitize `file_name` (path traversal, control chars) — storage path/filename is
  server-generated, client name stored only as `original_filename`.
- Refresh-token reuse ⇒ revoke the whole family and close sockets.
- Rate-limit `POST /storages/ws-token` to prevent Redis exhaustion via mass token creation.
- On abort/limit breach, delete the partial S3/GCS object and never insert the `models.Storages` row.

## 9. Rollout / Milestones

1. Add `STORAGE_DRIVER` selection for the existing REST upload path (§5.4) — independently
   valuable and currently silently broken.
2. Add `REDIS_WS_AUTH` prefix, token issue/refresh REST endpoints, encryption + decrypt
   middleware, env config — testable in isolation.
3. Promote `coder/websocket` to a direct dependency; add WS upgrade endpoint with auth,
   ping/keepalive and `renew`, no file handling yet.
4. Add the streaming storage package, wire protocol, size limit and concurrency cap, and
   `models.Storages` insertion.
5. Load/size-limit/concurrency testing (including double-`DECR` and crash cases), then rollout
   behind a feature flag if the project has one.

## 10. Appendix — Key File References

| Area | File |
|---|---|
| REST upload controller | `modules/storages/controller.go` |
| REST upload service | `modules/storages/service.go` |
| Storage routes | `modules/storages/registrator.go` |
| Storage model | `modules/models/storages.go` |
| Vendored storage backends | `vendor/github.com/fari-99/go-helper/storages/{base,s3,gcs,local}.go` |
| JWT wrapper (`SetExpired` now public) | `vendor/github.com/fari-99/go-helper/token_generator/{base,jwt}.go` |
| Encryption helper | `vendor/github.com/fari-99/go-helper/crypts/{base,default}.go` |
| WebSocket library (vendored, indirect today) | `vendor/github.com/coder/websocket` |
| Session helpers | `helpers/sessions.go` |
| Auth middleware | `modules/middleware/middleware_auth.go` |
| Redis client factory | `modules/configs/redis.go` |
| Redis config (env) | `global.env` (repo root) |
| S3/GCS/JWT env vars | `.env`, `.env.example` |
| Composition root | `cmd/servers/main/main.go` |

## 11. Implementation Status (2026-09-30)

All milestones in §9 are implemented.

| Milestone | Where |
|---|---|
| 1. `STORAGE_DRIVER` for REST | `modules/storages/service.go` (`newStorageBase`) |
| 2. Token issue/refresh, encryption, decrypt middleware, `REDIS_WS_AUTH` | `modules/ws_auth/`, `modules/configs/redis.go` |
| 3. WS endpoint `GET /ws/storages/upload`, ping/renew/watchdog | `modules/storages/ws_controller.go`, `registrator.go` |
| 4. Streaming storage package, protocol, limits, concurrency, `models.Storages` insert | `pkg/wsstorage/`, `modules/storages/ws_*.go` |
| 5. Tests | `pkg/wsstorage/*_test.go`, `modules/ws_auth/service_test.go`, `modules/storages/ws_controller_test.go` |

Endpoints: `POST /storages/ws-token` (REST session auth, rate limited),
`POST /storages/ws-token/refresh` (`{"refresh_token": "..."}`), `GET /ws/storages/upload?token=...`.

Notes / deviations from the design above:
- Tests use `miniredis` (in-memory Redis), a new test-only dependency.
- `wsstorage` and the `ws_auth` service/config now live in go-helper v1.7.0 (`wsstorage`, `ws_auth`);
  go-api keeps only the gin controller, middleware and registrator. `crypts.Decrypt` no longer
  panics on short input (fixed in the same release).
- Feature flag: `WS_UPLOAD_ENABLED` (default `false`). When off, the WS routes are not registered and
  no `WS_*` secrets or `REDIS_WS_AUTH_*` are needed. When on, `wsauth.NewService` still panics at
  startup if the `WS_*` secrets are missing.
- Renewal keeps a family alive for as long as the client keeps renewing before the 30 minute
  refresh token expires (sliding session). Cap the total lifetime later if that is unwanted.
- Not verified against real S3/GCS (no credentials in this environment): only the local driver is
  covered by tests. The S3/GCS code paths compile and follow the vendored implementation.
