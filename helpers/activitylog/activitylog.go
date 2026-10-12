// Package activitylog emits user-activity audit logs for the Radiance service.
//
// Every audited data mutation is rendered as a single line of JSON on stdout
// (consumed by Fluentd per the "User Activity Logs" PRD). Two entry points are
// provided:
//
//   - LogActivity — a low-coupling helper any controller/service can call
//     directly with pre-computed field maps.
//   - RegisterCallbacks (see callback.go) — wires gorm Create/Update/Delete
//     callbacks so mutations are logged automatically, without every call site
//     remembering to log.
//
// Immutability (PRD NFR-03): these logs are strictly WRITE-ONLY. They are
// emitted to stdout only and are NEVER persisted to an application table, so no
// HTTP controller or other application code path can update or delete a past
// entry. Do not add a database destination for these logs.
package activitylog

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Action names emitted in the "action" field of an ActivityLog. The BULK_*
// variants are used by the gorm callbacks when an operation targets rows via a
// WHERE condition rather than a single primary key (see callback.go).
const (
	ActionCreate     = "CREATE"
	ActionUpdate     = "UPDATE"
	ActionDelete     = "DELETE"
	ActionBulkCreate = "BULK_CREATE"
	ActionBulkUpdate = "BULK_UPDATE"
	ActionBulkDelete = "BULK_DELETE"
)

// ActivityLog is the JSON payload emitted (one line) per audited mutation.
//
// Only the fields relevant to the action are populated, to keep the emitted
// line small:
//
//   - CREATE / BULK_CREATE: new_data only (there is no prior state).
//   - UPDATE: diff only — one entry per changed field, old vs new.
//   - DELETE / BULK_DELETE: old_data only (there is no new state).
//   - BULK_UPDATE: old_data/new_data, unchanged — old_data carries the WHERE
//     condition (not real previous values), so a per-field diff would be
//     meaningless.
//
// old_data / new_data / diff are intentionally map-typed to avoid coupling the
// logger to any concrete model. Unpopulated fields are omitted from the JSON
// line entirely (omitempty).
type ActivityLog struct {
	Timestamp   string                 `json:"timestamp"`
	ActorUserID interface{}            `json:"actor_user_id"`
	Action      string                 `json:"action"`
	Entity      string                 `json:"entity"`
	EntityID    interface{}            `json:"entity_id"`
	OldData     map[string]interface{} `json:"old_data,omitempty"`
	NewData     map[string]interface{} `json:"new_data,omitempty"`
	Diff        map[string]FieldDiff   `json:"diff,omitempty"`
}

// FieldDiff is a single changed field emitted under ActivityLog.Diff for
// UPDATE actions.
type FieldDiff struct {
	Old interface{} `json:"old"`
	New interface{} `json:"new"`
}

// sensitiveColumns lists column/field-name fragments whose values must never
// reach the log sink. Matching is case-insensitive and substring-based, so
// "password" also hides "password_hash". Extend as new sensitive fields appear.
var sensitiveColumns = []string{
	"password",
	"salt",
	"token",
	"secret",
	"otp",
	"pin",
	"credential",
	"private_key",
}

// LogActivity builds an ActivityLog from the supplied fields and emits it as a
// single line of JSON to stdout. The timestamp is generated here in UTC and
// formatted as RFC3339 (e.g. 2026-06-23T10:00:00Z) — callers never pass it.
//
// Passing nil (or a partial map) for oldData/newData is safe: nil becomes a JSON
// null, and only the fields provided are emitted. Sensitive columns (see
// sensitiveColumns) are stripped from both maps.
//
// It is defensive by contract: any marshal/write failure — or panic — is
// reported to Sentry (using the optional ctx args, which helpers/sentry accepts
// as iris.Context / context.Context) and swallowed, so activity logging can
// never return an error that would block the caller's business operation.
//
// Usage:
//
//	activitylog.LogActivity(actorUserID, activitylog.ActionUpdate, "users", user.ID,
//	    map[string]interface{}{"email": oldEmail},
//	    map[string]interface{}{"email": newEmail},
//	    ctx)
func LogActivity(actorUserID interface{}, action, entity string, entityID interface{}, oldData, newData map[string]interface{}, ctx ...interface{}) {
	emitter(buildEntry(actorUserID, action, entity, entityID, oldData, newData), ctx...)
}

// emitter is the single sink every log goes through (direct LogActivity calls,
// immediate gorm callbacks, and transaction flushes). It is a variable only so
// tests can capture entries without parsing the output.
var emitter = emit

// buildEntry assembles an ActivityLog, stamping the current UTC time (RFC3339)
// and sanitizing the data maps. Split out from LogActivity so it can be unit
// tested without capturing stdout.
func buildEntry(actorUserID interface{}, action, entity string, entityID interface{}, oldData, newData map[string]interface{}) ActivityLog {
	entry := ActivityLog{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		ActorUserID: actorUserID,
		Action:      action,
		Entity:      entity,
		EntityID:    entityID,
	}

	sanitizedOld := RemoveSensitive(oldData)
	sanitizedNew := RemoveSensitive(newData)

	switch action {
	case ActionCreate, ActionBulkCreate:
		entry.NewData = sanitizedNew
	case ActionDelete, ActionBulkDelete:
		entry.OldData = sanitizedOld
	case ActionUpdate:
		entry.Diff = buildDiff(sanitizedOld, sanitizedNew)
	default:
		// ActionBulkUpdate and any caller-supplied custom action: keep both maps
		// as-is (see the BULK_UPDATE note on ActivityLog).
		entry.OldData = sanitizedOld
		entry.NewData = sanitizedNew
	}

	return entry
}

// buildDiff pairs each changed field in newData with its previous value from
// oldData, dropping fields whose value didn't actually change. Only newData's
// keys are considered: oldData may carry the full previous row (untouched
// columns included), and those must not show up as spurious diffs.
func buildDiff(oldData, newData map[string]interface{}) map[string]FieldDiff {
	if len(newData) == 0 {
		return nil
	}

	diff := make(map[string]FieldDiff, len(newData))
	for field, newVal := range newData {
		oldVal, hadOld := oldData[field]
		if hadOld && reflect.DeepEqual(oldVal, newVal) {
			continue
		}
		diff[field] = FieldDiff{Old: oldVal, New: newVal}
	}

	if len(diff) == 0 {
		return nil
	}
	return diff
}

// emit marshals an ActivityLog to one JSON line and writes it to stdout. On any
// failure (marshal, write, or panic) it falls back to Sentry and returns — it
// never panics and never returns an error.
func emit(entry ActivityLog, ctx ...interface{}) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("activitylog: recovered from panic while emitting log: %v", r)
		}
	}()

	body, err := marshalStructuredLog(entry.Action, entry)
	if err != nil {
		return
	}

	key := entry.Entity
	if entry.EntityID != nil {
		key = fmt.Sprintf("%s:%v", entry.Entity, entry.EntityID)
	}
	deliver(Message{Key: key, Body: body})
}

// marshalStructuredLog renders the {"log_name","data"} envelope as one compact
// JSON line.
func marshalStructuredLog(logName string, data interface{}) ([]byte, error) {
	now := time.Now().UTC()

	dataPayload := map[string]interface{}{
		"created_at": now,
		"updated_at": now,
	}

	if payloadData, ok := data.(map[string]interface{}); ok {
		for k, v := range payloadData {
			dataPayload[k] = v
		}

		for k, v := range dataPayload {
			dataPayload[k] = normalizeStructuredLogDataField(v)
		}
	} else {
		dataPayload["payload"] = data
	}

	payload := struct {
		LogName string                 `json:"log_name"`
		Data    map[string]interface{} `json:"data"`
	}{
		LogName: logName,
		Data:    dataPayload,
	}

	// Compact JSON keeps each log event on a single line.
	return json.Marshal(payload)
}

// normalizeStructuredLogDataField attempts to decode JSON-encoded strings into objects/arrays.
func normalizeStructuredLogDataField(value interface{}) interface{} {
	s, ok := value.(string)
	if !ok {
		return value
	}

	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return value
	}

	switch v.(type) {
	case map[string]interface{}, []interface{}:
		return v
	default:
		b, err := json.Marshal(value)
		if err != nil {
			return value
		}

		var nestedMap map[string]interface{}
		if err := json.Unmarshal(b, &nestedMap); err == nil && nestedMap != nil {
			return RemoveSensitive(nestedMap)
		}

		var nestedSlice []interface{}
		if err := json.Unmarshal(b, &nestedSlice); err == nil {
			return sanitizeValue(nestedSlice)
		}

		return value
	}
}

// RemoveSensitive returns a shallow copy of data with any sensitive columns
// (see sensitiveColumns) dropped, so secrets like passwords never reach the log
// sink. It is nil-safe: a nil map returns nil.
func RemoveSensitive(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}
	cleaned := make(map[string]interface{}, len(data))
	for key, value := range data {
		if isSensitive(key) {
			continue
		}
		cleaned[key] = sanitizeValue(value)
	}
	return cleaned
}

func sanitizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		return RemoveSensitive(v)
	case []interface{}:
		cleaned := make([]interface{}, len(v))
		for i, item := range v {
			cleaned[i] = sanitizeValue(item)
		}
		return cleaned
	default:
		return value
	}
}

// isSensitive reports whether a column/field name matches any sensitiveColumns
// fragment (case-insensitive, substring).
func isSensitive(column string) bool {
	lower := strings.ToLower(column)
	for _, sensitive := range sensitiveColumns {
		if strings.Contains(lower, sensitive) {
			return true
		}
	}
	return false
}
