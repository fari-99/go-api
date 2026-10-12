# activitylog

Audit log of data changes. Every create / update / delete made through gorm is
turned into one JSON log entry (who, what, old vs new), without call sites
having to remember to log.

## What is logged automatically

Everything that goes through the shared gorm connection
(`configs.SetConnection`, which calls `activitylog.RegisterCallbacks`):

| Operation                                        | Action        | Contents                                   |
| ------------------------------------------------ | ------------- | ------------------------------------------ |
| `db.Create(&model)`                              | `CREATE`      | `new_data`, `entity_id`                    |
| `db.Create(&[]model{...})`                       | `BULK_CREATE` | no per-row data / id                       |
| `db.Save(&model)`, `db.Model(&m).Update(s)(...)` | `UPDATE`      | `diff` of changed fields (old → new)       |
| `db.Where(...).Updates(...)` (no primary key)    | `BULK_UPDATE` | executed SQL as `condition`, args redacted |
| `db.Delete(&model)` (incl. soft delete)          | `DELETE`      | `old_data`                                 |
| `db.Where(...).Delete(&Model{})`                 | `BULK_DELETE` | executed SQL as `condition`                |

You do not need to "enable" anything per model or per call. Sensitive columns
(`password`, `token`, `secret`, `otp`, `pin`, ... see `sensitiveColumns`) are
never written.

### What is NOT logged

- Reads (`Find`, `First`, ...).
- Raw SQL: `db.Exec(...)`, `db.Raw(...)`.
- Inserts without a model: `db.Table("x").Create(map[string]interface{}{...})`.
- Anything on a connection that did not go through `RegisterCallbacks`.

If one of those needs an audit trail, call `LogActivity` yourself (below).

## Setting the actor (do this)

The callbacks cannot know who the user is. Without it the entry is still
emitted, but with `actor_user_id: 0` and a `WARNING no actor set` line.
Attribute the statements with `WithActor`:

```go
func (r repository) Create(ctx *gin.Context, user models.Users, actorID uint64) error {
    db := activitylog.WithActor(r.DB.WithContext(ctx), actorID)

    return db.Create(&user).Error
}
```

The handle returned by `WithActor` is reusable and carries into transactions:

```go
db := activitylog.WithActor(r.DB.WithContext(ctx), currentUser.ID)

return db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&a).Error; err != nil { return err }  // actor = currentUser.ID
    return tx.Save(&b).Error                                   // actor = currentUser.ID
})
```

Plain gorm works too, per statement:
`db.Set(activitylog.CurrentUserIDKey, id).Create(&m)`.

Precedence: explicit `CurrentUserIDKey` > context-only flow (system job,
actor `0`, no warning) > nothing set (actor `0` + warning). Logging never fails
the database operation.

## Transactions

Inside a transaction (`db.Begin()`, `db.Transaction(...)`, or gorm's default
per-statement one) entries are **held back** and emitted together, in order,
only after the commit succeeds. A rollback, an error returned from
`Transaction`, or a failed commit emits nothing. A rolled-back nested
`Transaction` (savepoint) drops only its own entries.

This is done by wrapping the gorm connection pool, so existing
`Begin`/`Commit`/`Transaction` code needs no change.

## Where the logs go

Set `ACTIVITY_LOG_DRIVER` (see `.env.example`):

| Driver              | Behaviour                                                       |
| ------------------- | --------------------------------------------------------------- |
| `stdout` (default)  | one JSON line per entry via `log.Print`                         |
| `rabbitmq` / `kafka`| published asynchronously, in batches, off the request path      |

For `rabbitmq` / `kafka` an in-memory queue sits in front of the broker, so a
slow or down broker never slows a database call. Anything that cannot be
delivered (queue full, retries exhausted, shutting down) is written to stdout
instead of being dropped. Delivery is at-least-once; consumers should
tolerate duplicates. A hard crash loses what is still queued — a normal
SIGINT/SIGTERM flushes first.

Tuning: `ACTIVITY_LOG_BUFFER_SIZE`, `ACTIVITY_LOG_BATCH_SIZE`,
`ACTIVITY_LOG_FLUSH_INTERVAL_MS`, `ACTIVITY_LOG_MAX_RETRIES`.

Startup / shutdown (already wired in `cmd/servers/main`):

```go
configs.InitActivityLog()                 // reads ACTIVITY_LOG_* and configures the sink
defer activitylog.Shutdown(ctx)           // flush queued entries before exit
```

## Logging by hand

For changes the callbacks cannot see:

```go
activitylog.LogActivity(actorID, activitylog.ActionUpdate, "users", user.ID,
    map[string]interface{}{"email": oldEmail},
    map[string]interface{}{"email": newEmail})
```

Note: a direct `LogActivity` call has no transaction, so it is emitted
immediately even if you are inside one.

## Entry shape

```json
{"log_name":"UPDATE","data":{"created_at":"...","updated_at":"...","payload":{
  "timestamp":"2026-06-23T10:00:00Z","actor_user_id":42,"action":"UPDATE",
  "entity":"users","entity_id":7,
  "diff":{"email":{"old":"a@x.io","new":"b@x.io"}}}}}
```

`old_data` / `new_data` / `diff` are only present for the actions listed above.
These logs are write-only by design: never store them in an application table.
