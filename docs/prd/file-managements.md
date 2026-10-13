# PRD: File Managements (post-upload file processing)

- Status: Implemented v1 (see §12), owner decisions in §9
- Owner: TBD
- Related modules: `modules/storages` (input), new `modules/file_managements`, `modules/tasks` (worker),
  new `pkg/file_processor_helper` (third-party client template)
- Author: generated with Claude Code from codebase research on 2026-10-13, revised the same day with owner decisions

## 1. Problem / Motivation

Some uploaded files need more than storage. A PDF may need a summary, an audio file may need a
transcript. Both jobs go to a third-party service, take seconds to minutes, and can fail.

Today `modules/storages` only stores and serves bytes. Nothing tracks "do X with file Y", its
progress, or its result.

This PRD adds a **file management** module. It is a job layer on top of storages:

1. The client uploads through the existing storages endpoints (`POST /storages/upload` or the WS
   upload) and gets a `storage_id`.
2. The client sends `storage_id` + an `action` to `POST /file-managements`.
3. A background worker sends the file to a third-party service and records status and results.
4. The client polls the file management record and reads the result.

Results and parameters differ a lot per action, so they go into a `file_management_meta` table
(`group`, `type`, `key`, `value`) instead of new columns for each one.

The third-party service is not known yet. It runs "locally" (self-hosted, reachable over HTTP).
v1 ships a **generic upload client template** built on resty. The actions `pdf_summary` and
`audio_transcription` are examples of how to plug an action in, not integrations with a specific
provider.

## 2. Current State (verified in code)

| Area | Finding |
|---|---|
| Storage record | `models.Storages` (`modules/models/storages.go`): `type`, `path`, `filename`, `mime` (server-sniffed), `original_filename`, `status` (99 active / 97 deleted), `created_by`. |
| Ownership check | `storages.service.OpenOwned` treats records owned by another user, or not active, as "not found". It is tied to `*gin.Context`, so the worker cannot call it. The worker can call the vendored `GetFiles(type, path, filename)` directly, which works for local/S3/GCS through the same `STORAGE_DRIVER` switch. |
| Upload size | REST upload hard limit is 8 MB. The WS default is `WS_UPLOAD_MAX_FILE_MB=8`. Accepted for v1 (§9 Q10). |
| Meta table precedent | `tax_configuration_meta` (`group`, `key`, `value`) and `user_metas` (`groups`, `key`, `value`). |
| Status constants | Integer constants in `constant/constants.go` with label maps in `constant/constant_models/`. |
| State machine | `looplab/fsm v1.0.4` is already a direct, vendored dependency, used by `bussiness_flow/` for transactions. This module uses the package directly with its own small FSM and does not reuse `bussiness_flow` (§9 Q5). |
| HTTP client | `go-resty/resty/v2 v2.17.2` is already direct and vendored. `pkg/go2rtc_helper` is the house pattern for a resty client (env-configured `New()`, typed `Error`, `checkStatus`). |
| Async jobs | RabbitMQ through vendored `go-helper/rabbitmq`. Consumers are CLI commands in `modules/tasks/*_commands.go` run by `cmd/tasks/commands.go`. The handler signature is `func(rabbitmq.ConsumerHandlerData)` with no error return. Messages are acked after the handler returns; only a panic triggers the library retry. `NewBaseQueue` blocks until connected (it sleeps 5 s before each dial). `Publish` mutates a shared message struct, so concurrent publishes on one `QueueSetup` must be serialized. |
| Activity log | `activitylog.RegisterCallbacks(db)` logs GORM writes automatically. The actor comes from `activitylog.WithActor(db, userID)`; the worker uses actor `0` (system). |
| RBAC | Casbin checks `(role, URL path, method)`. New routes need policy rows for non-admin roles. Admin has allow-all. |
| Migrations | Separate repo `../database-migration`, Flyway-style `mysql/V{n}__*.sql` + `postgres/V{n}__*.sql`. Latest is `V4`. IDs are `bigint unsigned AUTO_INCREMENT` (`models.IDType` = `uint64`). |

## 3. Goals

1. Create a processing job for an owned, active storage file with a given `action`.
2. Track the job through the statuses **new → on progress → completed / failed / cancelled**,
   enforced by a `looplab/fsm` state machine.
3. Run jobs asynchronously in a RabbitMQ worker, never inside the HTTP request.
4. Provide a reusable, self-contained resty client template that uploads a file to a third-party
   endpoint and returns its JSON response.
5. Ship two example actions, `pdf_summary` and `audio_transcription`, built on that template.
6. Store all per-action inputs, outputs, request details and errors in `file_management_meta`.
7. Let users list, view, cancel, retry and delete their jobs.
8. Make adding a new action mean writing one processor and registering it, with no schema change.

## 4. Non-Goals (v1)

- Integrating a specific provider (OpenAI, Anthropic, Deepgram, …).
- Changing the upload flow, `models.Storages`, or the 8 MB upload limit.
- Sweeper / stuck-job recovery and automatic retries (§10).
- Completion notifications (§10, optional).
- Chaining actions or real-time push of progress.

## 5. Proposed Design

### 5.1 High-level flow

```
Client                     go-api (HTTP)                    RabbitMQ           Worker (tasks CLI)        3rd party (local)
  | POST /storages/upload ->  storages row (active)
  |<- storage_id
  | POST /file-managements {storage_id, action, params}
  |                         validate owner/mime/size/params
  |                         INSERT file_managements (status=new) + input meta   (one tx)
  |                         publish {file_management_id} ---->  file-management-queue
  |<- 201 {id, status:new}                                            | consume
  |                                                                   | fsm "process": new -> on_progress
  |                                                                   |   UPDATE ... WHERE id=? AND status=new
  |                                                                   | open file (GetFiles)
  |                                                                   | resty multipart POST ------------> /endpoint
  |                                                                   |                       <------------ JSON
  |                                                                   | fsm "complete": on_progress -> completed
  |                                                                   |   + output/request meta   (one tx)
  | GET /file-managements/:id  -> status + grouped meta
```

### 5.2 Data model

#### `file_managements`

| Column | Type | Notes |
|---|---|---|
| `id` | bigint unsigned PK AI | |
| `storage_id` | bigint unsigned NOT NULL | References `storages.id` by convention (no hard FK, matches the existing schema). |
| `action` | varchar(50) NOT NULL | `pdf_summary`, `audio_transcription`, … |
| `status` | int(3) unsigned NOT NULL | See §5.3. |
| `attempt` | int unsigned NOT NULL DEFAULT 0 | Incremented on every `process` transition (first run = 1). |
| `started_at` | datetime NULL | Set on `process`. |
| `finished_at` | datetime NULL | Set on completed / failed / cancelled, cleared on retry. |
| `created_by` | bigint unsigned NOT NULL | Owner, from the session. |
| `created_at` / `updated_at` / `deleted_at` | datetime | Standard `models.Base`. |

Indexes: `(created_by, status)`, `(storage_id, action)`.

#### `file_management_meta`

| Column | Type | Notes |
|---|---|---|
| `id` | bigint unsigned PK AI | |
| `file_management_id` | bigint unsigned NOT NULL | |
| `group` | varchar(50) NOT NULL | Top-level category. Quoted in SQL (reserved word). |
| `type` | varchar(50) NOT NULL | **Sub-category under `group`** (§9 Q1). |
| `key` | varchar(100) NOT NULL | Quoted in SQL. |
| `value` | LONGTEXT NULL | **Always a string** (§9 Q1). Non-string data (numbers, objects, arrays) is stored as its JSON encoding. |
| `created_by` | bigint unsigned NULL | 0 when written by the worker. |
| `created_at` / `updated_at` / `deleted_at` | datetime | |

Unique index: `(file_management_id, group, type, key)`. Writes are upserts, so a retry overwrites
earlier output instead of adding rows.

**Groups, types and keys (v1):**

| group | type | keys | Written by | Shown to owner |
|---|---|---|---|---|
| `input` | `params` | one per validated param (`language`, `length`, …) | API on create | yes |
| `output` | the action's output type (`summary`, `transcription`) | one per top-level field of the third-party JSON response | worker on success | yes |
| `request` | `third_party` | `endpoint`, `status_code`, `latency_ms` | worker | no |
| `error` | `attempt_<n>` | `code`, `message`, `internal_message` | worker on failure | `code` and `message` only |

Keeping errors per attempt (`attempt_1`, `attempt_2`, …) keeps the retry history. The detail
endpoint shows the error for the current `attempt` only.

### 5.3 Status lifecycle (`looplab/fsm`)

Constants in `constant/constants.go`, labels in `constant/constant_models/file_managements.go`:

| Const | Value | Label | FSM state |
|---|---|---|---|
| `FileManagementNew` | 1 | New | `new` |
| `FileManagementOnProgress` | 2 | On Progress | `on_progress` |
| `FileManagementFailed` | 3 | Failed | `failed` |
| `FileManagementCancelled` | 4 | Cancelled | `cancelled` |
| `FileManagementCompleted` | 5 | Completed | `completed` |

No extra statuses are needed in v1. Deletion uses `deleted_at`, not a status.

FSM events:

| Event | From | To | Trigger |
|---|---|---|---|
| `process` | new | on_progress | worker picks the job |
| `complete` | on_progress | completed | third party success |
| `fail` | on_progress | failed | third party error or timeout; also from `new` when the publish fails |
| `cancel` | new, on_progress | cancelled | user |
| `retry` | failed, cancelled | new | user |

`completed` is terminal. To re-run, create a new job.

**Implementation:** `modules/file_managements/state.go` builds an `fsm.FSM` at the record's current
state and fires the event. An invalid event returns `fsm.InvalidEventError`, which becomes a 409
for API calls. The new state is then written with a **conditional update**
(`UPDATE … SET status=? WHERE id=? AND status=<from>`). If 0 rows are affected, the transition
lost a race and is treated as a conflict. This is what keeps it correct:

- A redelivered RabbitMQ message can't run a job twice (`process` needs `status=new`).
- A cancel during processing wins. The worker's `complete` update finds `status=cancelled`,
  affects 0 rows, and the result is dropped.

### 5.4 Third-party client template (`pkg/file_processor_helper`)

The package has no go-api imports (only stdlib + resty), so it can move into go-helper later
(§9 Q6). It follows `pkg/go2rtc_helper`:

```go
type Config struct {
    BaseURL string
    APIKey  string        // optional, sent as "Authorization: Bearer <key>"
    Timeout time.Duration
    Debug   bool
}

func ConfigFromEnv() Config            // FILE_PROCESSOR_BASE_URL, _API_KEY, _TIMEOUT_SECONDS, _DEBUG
func New(cfg Config) (*Client, error)  // error when BaseURL is empty

type UploadRequest struct {
    Endpoint  string            // e.g. "/pdf/summary"
    FieldName string            // multipart file field, default "file"
    FileName  string
    File      io.Reader
    Fields    map[string]string // extra form fields (action params)
}

type UploadResponse struct {
    StatusCode int
    Body       []byte
    Latency    time.Duration
}

func (c *Client) Upload(ctx context.Context, req UploadRequest) (*UploadResponse, error)

type Error struct { StatusCode int; Body string }   // returned for HTTP >= 400
```

- Sends `POST {base_url}{endpoint}` as `multipart/form-data` with the file + fields, `Accept: application/json`.
- Network errors and timeouts are returned wrapped. HTTP ≥ 400 returns `*Error`.
- The template assumes a synchronous JSON response. Async APIs (job id + polling/webhook) are
  an enhancement (§10).

### 5.5 Actions & processors

```go
type Processor interface {
    Action() string                    // "pdf_summary"
    OutputType() string                // meta type for group=output, e.g. "summary"
    AcceptedMimes() []string
    AcceptedExtensions() []string      // fallback when sniffing gives application/octet-stream
    MaxFileBytes() int64
    ValidateParams(params map[string]any) (map[string]string, error)
    Process(ctx context.Context, job Job) (*Result, error)
}
```

`Process` errors are `*ProcessError{Code, Message, Internal}`. `Message` is safe to show the user;
`Internal` is stored but never returned.

A shared `uploadProcessor` implements `Process` with the template client. It uploads the file
with the params as form fields, decodes the JSON object response, and turns each top-level field
into an `output` meta row (strings as-is, `null` as `""`, everything else JSON-encoded). Each example action is a
small config on top of it:

| Action | Output type | Accepted MIME (ext fallback) | Params | Endpoint env (default) |
|---|---|---|---|---|
| `pdf_summary` | `summary` | `application/pdf` (`.pdf`) | `language` (2–10 chars), `length` (`short`/`medium`/`long`) | `FILE_MANAGEMENT_PDF_SUMMARY_ENDPOINT` (`/pdf/summary`) |
| `audio_transcription` | `transcription` | `audio/mpeg`, `audio/wav`, `audio/wave`, `audio/x-wav`, `audio/mp4`, `audio/ogg`, `audio/webm`, `audio/flac`, `audio/aac` (`.mp3`, `.wav`, `.m4a`, `.ogg`, `.webm`, `.flac`, `.aac`) | `language` (2–10 chars), `timestamps` (bool) | `FILE_MANAGEMENT_AUDIO_TRANSCRIPTION_ENDPOINT` (`/audio/transcriptions`) |

Unknown params are rejected. The registry registers the example actions only when
`FILE_PROCESSOR_BASE_URL` is set. The API then reports a missing action as `action_not_available`
instead of failing jobs later.

### 5.6 API

All routes use the `authorized` middleware (session + RBAC). Users only see their own jobs.
Another user's id returns **404**, the same anti-probing rule as `OpenOwned`.

| Method | Path | Purpose |
|---|---|---|
| GET | `/file-managements/actions` | Available actions with accepted MIME types, extensions, max size, params. |
| POST | `/file-managements` | Create a job. |
| GET | `/file-managements` | List: `page`, `limit` (≤100), filters `status`, `action`, `storage_id`. No meta. |
| GET | `/file-managements/:fileManagementID` | Detail + storage summary + visible meta. |
| POST | `/file-managements/:fileManagementID/cancel` | `cancel` event. |
| POST | `/file-managements/:fileManagementID/retry` | `retry` event + republish. |
| DELETE | `/file-managements/:fileManagementID` | Soft delete, only when completed / failed / cancelled. |

**Create request:**

```json
{ "storage_id": 123, "action": "pdf_summary", "params": { "length": "short", "language": "en" } }
```

**Create validation, in order:**
1. Body is valid (`storage_id`, `action` required), else 400.
2. `action` is registered, else 422 `action_not_available`.
3. Storage exists, is owned by the user and is active, else 404 `storage_not_found`.
4. MIME (or extension fallback) is accepted, else 422 `unsupported_file_type`.
5. `params` pass `ValidateParams`, else 422 `invalid_params`.
6. **Idempotency (§9 Q4):** if a `new` or `on_progress` job exists for the same user, storage,
   action and normalized params, return it with **200** instead of creating a new one.
7. Active-job cap: the user's `new` + `on_progress` jobs must be below
   `FILE_MANAGEMENT_MAX_ACTIVE_PER_USER` (default 5), else 429 `too_many_active_jobs`.

File size is **not** checked on create: for S3/GCS that would mean downloading the file in the
request. The worker checks it after opening the file and fails the job with `file_too_large`.
Uploads are capped at 8 MB anyway, the same as the default action limit.

On success, insert the job and its `input` meta in one transaction, then publish after commit.
If the publish fails, fire `fail` with error code `queue_publish_failed`. The user can then retry.
Response is **201**.

**Detail response:**

```json
{
  "id": 42, "storage_id": 123, "action": "pdf_summary",
  "status": 5, "status_label": "Completed", "attempt": 1,
  "started_at": "...", "finished_at": "...", "created_at": "...", "updated_at": "...",
  "storage": { "id": 123, "original_filename": "report.pdf", "mime": "application/pdf" },
  "meta": {
    "input":  { "params":  { "length": "short", "language": "en" } },
    "output": { "summary": { "summary": "...", "pages": "12" } }
  },
  "error": null
}
```

`error` is `{ "code": "...", "message": "..." }` for the current attempt when the job failed.

### 5.7 Worker

- Queue `file-management-queue` (`constant/queue_name.go`), event `file-management-process`
  (`constant/event_name.go`). Payload data: `{"file_management_id": 42}`. The worker always
  reloads state from the DB. The queue is durable and messages are persistent, so jobs survive a
  broker restart.
- CLI command `handler-file-management-event` (alias `hfme`) in
  `modules/tasks/file_management_commands.go`. The handler logic lives in
  `modules/file_managements/worker.go` (no gin), so the module owns its own state handling.
- Per message:
  1. Load the job. If it is missing, or `process` fails (not `new`, e.g. cancelled or already
     taken), skip it.
  2. `process`: conditional update to `on_progress`, `attempt+1`, `started_at=now`.
  3. Load the storage. If it is missing or not active, `fail` with `storage_unavailable`.
  4. Open the file with vendored `GetFiles` and run the processor with a context timeout of
     `FILE_MANAGEMENT_PROCESS_TIMEOUT_SECONDS` (default 300).
  5. Success: in one transaction, `complete` (conditional) + upsert `output` and `request` meta.
     If the conditional update affects 0 rows (cancelled meanwhile), roll back and drop the result.
  6. Failure: `fail` (conditional) + upsert `error/attempt_<n>` and `request` meta.
- Scaling: one consumer per process (the library runs one goroutine with prefetch 1). Run more
  worker processes to scale.
- Cancel while on_progress is best effort. The HTTP call already in flight is not interrupted,
  and its result is dropped.
- Known v1 gap: a worker crash mid-job leaves the row `on_progress`. The user can `cancel` and
  then `retry`. Automatic recovery is the sweeper enhancement (§10).

### 5.8 Configuration (env)

| Env | Default | Meaning |
|---|---|---|
| `FILE_MANAGEMENT_ENABLED` | `false` | Feature flag. When off, routes are not registered and no RabbitMQ publisher is opened. |
| `FILE_MANAGEMENT_MAX_ACTIVE_PER_USER` | `5` | Cap on new + on_progress jobs per user. |
| `FILE_MANAGEMENT_PROCESS_TIMEOUT_SECONDS` | `300` | Per-job processing timeout. |
| `FILE_MANAGEMENT_PDF_SUMMARY_ENDPOINT` | `/pdf/summary` | Third-party path for `pdf_summary`. |
| `FILE_MANAGEMENT_PDF_SUMMARY_MAX_MB` | `8` | Size limit for `pdf_summary`. |
| `FILE_MANAGEMENT_AUDIO_TRANSCRIPTION_ENDPOINT` | `/audio/transcriptions` | Third-party path for `audio_transcription`. |
| `FILE_MANAGEMENT_AUDIO_TRANSCRIPTION_MAX_MB` | `8` | Size limit for `audio_transcription`. |
| `FILE_PROCESSOR_BASE_URL` | — | Third-party base URL. When empty, no actions are available. |
| `FILE_PROCESSOR_API_KEY` | — | Optional bearer token. |
| `FILE_PROCESSOR_TIMEOUT_SECONDS` | `300` | HTTP timeout. |
| `FILE_PROCESSOR_DEBUG` | `false` | resty debug logging (logs request bodies, so dev only). |

## 6. Code Layout

```
pkg/file_processor_helper/        client.go (+ tests)
modules/file_managements/
  registrator.go  controller.go  service.go  repository.go  requests.go
  state.go        (looplab/fsm)
  processors.go   (Processor, ProcessError, registry, uploadProcessor, example actions)
  meta.go         (meta row builders, grouping for responses)
  publisher.go    (RabbitMQ publisher)
  worker.go       (queue handler)
modules/models/file_managements.go, file_management_meta.go
constant/constants.go, constant/constant_models/file_managements.go,
constant/queue_name.go, constant/event_name.go
modules/tasks/file_management_commands.go (+ registered in base.go)
cmd/servers/main/main.go (wired behind FILE_MANAGEMENT_ENABLED)
../database-migration/{mysql,postgres}/V5__file_managements.sql
.env.example
```

## 7. Security

- Ownership is checked for both the job and the storage on every endpoint. Other users' ids return 404.
- No user consent flow is needed for now (§9 Q8).
- Never return `request` meta or `error.internal_message` to users. Third-party error bodies may
  contain internal details.
- The third-party response is stored as plain strings. The frontend must render it escaped, not as HTML.
- The API key lives only in env and is never logged or stored in meta. `FILE_PROCESSOR_DEBUG`
  must stay off outside development.
- `FILE_MANAGEMENT_MAX_ACTIVE_PER_USER` bounds how much third-party load a single user can cause.
- Activity log covers both tables automatically. The worker writes with actor 0 (system).
- RBAC: add Casbin policies for the new routes for the non-admin roles that should use them.

## 8. Testing

- `pkg/file_processor_helper`: `httptest` server checks multipart file + fields, auth header,
  HTTP error → `*Error`, timeout.
- FSM: every allowed transition, invalid events, unknown status.
- Repository (sqlite): conditional update race (double `process`, cancel vs complete), meta upsert.
- Service: create validation order; idempotent create; active cap; cancel/retry/delete rules;
  publish failure → failed.
- Worker: fake processor success / failure; job cancelled before or during processing;
  redelivered message is a no-op; storage deleted.
- Processors: param validation; JSON response → output meta; non-JSON / non-object response → error.

## 9. Decisions Log

| # | Question | Decision |
|---|---|---|
| Q1 | Meaning of meta `type` | Sub-category under `group`. `value` is always a string. |
| Q2 | Transcription provider | Unknown third party running locally. Build a resty upload client template. |
| Q3 | PDF summary provider | None. `pdf_summary` / `audio_transcription` are examples on the template. |
| Q4 | Duplicate create while one is running | Return the existing job. |
| Q5 | State machine | Use `looplab/fsm`, separate from `bussiness_flow`. |
| Q6 | Where provider client code lives | Self-contained package, ready to move into go-helper. |
| Q7 | Sweeper | On hold, listed as an enhancement (§10). |
| Q8 | User consent for sending files to third parties | Not needed currently. |
| Q9 | Completion notification | Optional, enhancement (§10). |
| Q10 | 8 MB upload limit | Accepted for v1, enhancement later (§10). |
| Q11 | Status numbering | 1–5. No extra statuses needed for v1. |

## 10. Enhancements (nice to have later)

- **Sweeper via cron** (`file-management-sweep`): fail `on_progress` jobs older than the process
  timeout (worker crash), republish `new` jobs whose publish was lost, and run delayed automatic
  retries for retryable errors with backoff and a max-attempts env.
- **Completion notification** through the existing `send-notifications` event (e.g. action
  `file-management-completed`) with a notification template.
- **Larger uploads for audio**: raise the storage upload limit (REST and WS) per file type.
- **Async third-party APIs**: job id + polling or a webhook callback, stored as `request` meta.
- **Large outputs**: when output exceeds N MB, write it to storages and keep only the storage id in meta.
- **Chaining**: `source_file_management_id` (e.g. summarize a transcript).
- **Progress**: `progress` meta group and WebSocket / SSE push.
- More actions: OCR, translation, video → audio extraction, document classification.

## 11. Rollout / Milestones

1. Migration `V5` + models + constants.
2. `pkg/file_processor_helper` client template.
3. Module: FSM, repository, service, controller, registrator, processors.
4. Worker command.
5. RBAC policies, `.env.example`, Swagger, then enable `FILE_MANAGEMENT_ENABLED` per environment.

## 12. Implementation Status (2026-10-13)

All milestones in §11 are implemented, except RBAC policy rows and regenerating Swagger (run
`make swagger`). Both are left to the owner.

| Area | Where |
|---|---|
| Migration `V5` | `../database-migration/{mysql,postgres}/V5__file_managements.sql` |
| Models, constants | `modules/models/file_managements.go`, `file_management_meta.go`, `constant/constants.go`, `constant/constant_models/file_managements.go`, `constant/{queue,event}_name.go` |
| Client template | `pkg/file_processor_helper/client.go` |
| Module | `modules/file_managements/` |
| Worker command | `modules/tasks/file_management_commands.go` (`go run cmd/tasks/commands.go handler-file-management-event`) |
| Storage access for the worker | `storages.OpenStoredFile` (`modules/storages/service.go`) |
| Wiring | `cmd/servers/main/main.go`, behind `FILE_MANAGEMENT_ENABLED` |
| Tests | `pkg/file_processor_helper/client_test.go`, `modules/file_managements/*_test.go` (sqlite with activity-log callbacks, passes with `-race`) |

Notes / deviations:
- The service passes `ctx.Request.Context()` to the repository, not `*gin.Context`. gin reuses
  `*gin.Context` after the request ends while `database/sql` may still read it (caught by `-race`).
- The detail response includes `available_events` (`cancel` / `retry`) so the frontend knows which
  buttons to show.
- Duplicate detection is check-then-insert, so two concurrent identical creates can still both
  succeed. The active-job cap bounds this. A unique constraint isn't possible because params live in meta.
- Not verified against a real third-party service or RabbitMQ (none in this environment). The
  queue code follows the existing notification consumers.
