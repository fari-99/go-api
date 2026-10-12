package activitylog

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Keys used to pass the acting user into a gorm operation so the callbacks can
// attribute the resulting ActivityLog. Set them per-call with gorm's Set:
//
//	db.Set(activitylog.CurrentUserIDKey, userID).Create(&model)  // explicit actor
//	db.Set(activitylog.CurrentUserCtxKey, ctx).Save(&model)      // system/context flow
//
// Actor precedence (see resolveActor):
//  1. an explicit CurrentUserIDKey value wins;
//  2. else, if only a context is present (system flows that carry no user
//     header) the actor defaults to 0;
//  3. else a WARNING is logged and the actor still defaults to 0 — logging must
//     never block the operation.
const (
	CurrentUserIDKey  = "activitylog:current_user_id"
	CurrentUserCtxKey = "activitylog:current_user_ctx"

	// instanceOldDataKey stashes the pre-mutation row between the Before and
	// After callbacks of a single update/delete (instance-scoped).
	instanceOldDataKey = "activitylog:old_data"
)

// systemActorUserID is the fallback actor for system-triggered or unattributed
// operations (note #2/#3: default to 0 rather than failing the operation).
const systemActorUserID = 0

// RegisterCallbacks wires the activity-log callbacks into a gorm DB so every
// Create/Update/Delete automatically emits an ActivityLog without call sites
// having to remember to log. Call once, right after opening the connection —
// utils.GetDBConnection does this for the shared singleton:
//
//	db := utils.GetDBConnection()
//	activitylog.RegisterCallbacks(db)
//
// Create is logged AFTER the row is written (so the generated primary key is
// available). Update/Delete capture the previous row state in a BEFORE callback
// (read within the same transaction) and log in an AFTER callback, so only
// operations that actually executed are logged and bulk conditions can be read
// from the executed SQL.
//
// Inside a transaction (db.Begin, db.Transaction, or gorm's own per-statement
// transaction) the logs are held in a buffer and emitted in bulk only when the
// commit succeeds; a rollback or failed commit drops them (see transaction.go).
// That is why the log callbacks run BEFORE gorm:commit_or_rollback_transaction:
// they must reach the buffer before the commit flushes it.
func RegisterCallbacks(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("RegisterCallbacks: db is nil")
	}

	// Wrap the pool once so transactions report their commit / rollback.
	if _, wrapped := db.ConnPool.(*hookedPool); !wrapped {
		db.ConnPool = &hookedPool{ConnPool: db.ConnPool}
		db.Statement.ConnPool = db.ConnPool
	}

	err := db.Callback().Create().Before("gorm:commit_or_rollback_transaction").Register("activitylog:create", logCreate)
	if err != nil {
		return err
	}

	err = db.Callback().Update().Before("gorm:update").Register("activitylog:before_update", captureOldDataForUpdate)
	if err != nil {
		return err
	}

	err = db.Callback().Update().Before("gorm:commit_or_rollback_transaction").Register("activitylog:update", logUpdate)
	if err != nil {
		return err
	}

	err = db.Callback().Delete().Before("gorm:delete").Register("activitylog:before_delete", captureOldDataForDelete)
	if err != nil {
		return err
	}

	err = db.Callback().Delete().Before("gorm:commit_or_rollback_transaction").Register("activitylog:delete", logDelete)
	if err != nil {
		return err
	}

	return nil
}

// logCreate emits a CREATE (or BULK_CREATE) log after a successful insert.
func logCreate(db *gorm.DB) {
	guard(db, func() {
		if db.Error != nil || db.Statement == nil || db.Statement.Schema == nil {
			return
		}
		actor, ctx := resolveActor(db)
		entity := db.Statement.Table

		// A batch insert (slice value) gives us no per-row primary keys, so it
		// is recorded as a bulk operation with no entity id (note #1).
		if db.Statement.ReflectValue.Kind() == reflect.Slice {
			dispatch(db, actor, ActionBulkCreate, entity, nil, nil, nil, ctx)
			return
		}

		dispatch(db, actor, ActionCreate, entity, resolvePK(db), nil, currentDataMap(db), ctx)
	})
}

// captureOldDataForUpdate reads the pre-update row (single-record updates only)
// and stashes it for the matching After callback.
func captureOldDataForUpdate(db *gorm.DB) {
	guard(db, func() {
		if db.Error != nil || db.Statement == nil || db.Statement.Schema == nil {
			return
		}
		if isBulkCondition(db) {
			return
		}
		if old := oldDataByPrimaryKey(db); old != nil {
			db.InstanceSet(instanceOldDataKey, old)
		}
	})
}

// logUpdate emits an UPDATE (or BULK_UPDATE) log after a successful update.
func logUpdate(db *gorm.DB) {
	guard(db, func() {
		if db.Error != nil || db.RowsAffected == 0 || db.Statement == nil || db.Statement.Schema == nil {
			return
		}
		actor, ctx := resolveActor(db)
		entity := db.Statement.Table
		newData := updatedDataMap(db)

		// Bulk update via WHERE condition — no single primary key. Record the
		// executed statement as the condition instead of a per-row old/id (note #1).
		if isBulkCondition(db) {
			dispatch(db, actor, ActionBulkUpdate, entity, nil, conditionData(db), newData, ctx)
			return
		}

		old, _ := db.InstanceGet(instanceOldDataKey)
		dispatch(db, actor, ActionUpdate, entity, resolvePK(db), toDataMap(old), newData, ctx)
	})
}

// captureOldDataForDelete reads the pre-delete row (single-record deletes only)
// and stashes it for the matching After callback.
func captureOldDataForDelete(db *gorm.DB) {
	guard(db, func() {
		if db.Error != nil || db.Statement == nil || db.Statement.Schema == nil {
			return
		}
		if isBulkCondition(db) {
			return
		}
		if old := oldDataByPrimaryKey(db); old != nil {
			db.InstanceSet(instanceOldDataKey, old)
		}
	})
}

// logDelete emits a DELETE (or BULK_DELETE) log after a successful delete.
func logDelete(db *gorm.DB) {
	guard(db, func() {
		if db.Error != nil || db.RowsAffected == 0 || db.Statement == nil || db.Statement.Schema == nil {
			return
		}
		actor, ctx := resolveActor(db)
		entity := db.Statement.Table

		// Bulk delete via WHERE condition — record the executed statement as the
		// condition, no per-row entity id (note #1).
		if isBulkCondition(db) {
			dispatch(db, actor, ActionBulkDelete, entity, nil, conditionData(db), nil, ctx)
			return
		}

		old, _ := db.InstanceGet(instanceOldDataKey)
		dispatch(db, actor, ActionDelete, entity, resolvePK(db), toDataMap(old), nil, ctx)
	})
}

// resolveActor determines the acting user id (and any context for Sentry) for a
// db statement, applying the precedence documented on CurrentUserIDKey.
//
// In gorm.io/gorm, db.Set() stores values in db.Statement.Settings (a sync.Map).
// We read them back with Statement.Settings.Load() rather than jinzhu's scope.Get().
func resolveActor(db *gorm.DB) (actorUserID interface{}, ctx interface{}) {
	if value, ok := db.Statement.Settings.Load(CurrentUserIDKey); ok && value != nil {
		ctxValue, _ := db.Statement.Settings.Load(CurrentUserCtxKey)
		return value, ctxValue
	}

	if ctxValue, ok := db.Statement.Settings.Load(CurrentUserCtxKey); ok {
		// System / context-driven flow with no explicit user header: default the
		// actor to 0 rather than failing the operation (note #2).
		return systemActorUserID, ctxValue
	}

	// Neither an actor id nor a context was provided: surface a WARNING so the
	// gap is visible, and fall back to actor 0 (note #3). Never block the op.
	log.Printf("activitylog: WARNING no actor set for operation on %q; defaulting actor_user_id to %d", db.Statement.Table, systemActorUserID)
	return systemActorUserID, nil
}

// isBulkCondition reports whether an update/delete targets rows by a WHERE
// condition (no single primary key) or a slice value, and so must be logged as
// a BULK_* action.
func isBulkCondition(db *gorm.DB) bool {
	if db.Statement.ReflectValue.Kind() == reflect.Slice {
		return true
	}
	return isZeroPK(resolvePK(db))
}

// resolvePK extracts the primary key value from the statement's model/dest.
// Returns nil when no primary key can be resolved (bulk conditions, empty structs).
func resolvePK(db *gorm.DB) interface{} {
	if db.Statement == nil || db.Statement.Schema == nil {
		return nil
	}
	pf := db.Statement.Schema.PrioritizedPrimaryField
	if pf == nil {
		return nil
	}
	val, isZero := pf.ValueOf(db.Statement.Context, db.Statement.ReflectValue)
	if isZero {
		return nil
	}
	return val
}

// isZeroPK reports whether a resolved primary key is nil or a zero value.
func isZeroPK(pk interface{}) bool {
	if pk == nil {
		return true
	}
	rv := reflect.ValueOf(pk)
	if !rv.IsValid() {
		return true
	}
	return rv.IsZero()
}

// oldDataByPrimaryKey loads the current (pre-mutation) row by primary key so an
// update/delete log can carry its old_data.
//
// Session{NewDB: true} drops the pending WHERE / Select / Omit state of the
// statement being logged (which would otherwise leak into this lookup) while
// keeping the same ConnPool, so the read still happens inside an open
// transaction. The returned map is built from the row loaded here — NOT from
// db.Statement.ReflectValue, which holds the in-memory (new) values.
func oldDataByPrimaryKey(db *gorm.DB) map[string]interface{} {
	if db.Statement.ReflectValue.Kind() != reflect.Struct {
		return nil
	}
	pk := resolvePK(db)
	if isZeroPK(pk) {
		return nil
	}

	// Allocate a fresh struct of the same type to receive the old row.
	oldRecord := reflect.New(db.Statement.ReflectValue.Type())
	if err := db.Session(&gorm.Session{NewDB: true}).
		Table(db.Statement.Table).
		First(oldRecord.Interface(), pk).Error; err != nil {
		// Row already gone / not found / query error — best-effort, skip old_data.
		return nil
	}

	return schemaDataMap(db.Statement.Context, db.Statement.Schema, oldRecord.Elem())
}

// currentDataMap flattens a statement's schema fields into a map keyed by db
// column name, skipping associations, ignored/unexported fields, and sensitive
// columns.
//
// In gorm.io/gorm, field metadata lives on schema.Field (db.Statement.Schema.Fields)
// rather than jinzhu's scope.Fields(). We use field.ValueOf to read the current
// value from the statement's reflect value.
func currentDataMap(db *gorm.DB) map[string]interface{} {
	if db.Statement == nil {
		return nil
	}
	return schemaDataMap(db.Statement.Context, db.Statement.Schema, db.Statement.ReflectValue)
}

// schemaDataMap reads every loggable column of rv (a struct of the given
// schema) into a map keyed by db column name.
func schemaDataMap(ctx context.Context, sch *schema.Schema, rv reflect.Value) map[string]interface{} {
	if sch == nil || rv.Kind() != reflect.Struct {
		return nil
	}

	data := make(map[string]interface{})
	for _, field := range sch.Fields {
		// Skip: associations, ignored/unexported, and sensitive columns.
		if skipField(field) {
			continue
		}
		if isSensitive(field.DBName) || isSensitive(field.Name) {
			continue
		}
		val, isZero := field.ValueOf(ctx, rv)
		if !isZero || val != nil {
			data[field.DBName] = val
		}
	}
	return data
}

// updatedDataMap returns the changed columns for an update. When the caller
// used a map or struct with db.Updates(), gorm.io/gorm stores the resolved
// destination in db.Statement.Dest. If Dest is a map we use it directly
// (it contains only the columns being changed); otherwise we fall back to the
// full model's current fields. Sensitive columns are stripped either way.
//
// Note: in jinzhu/gorm this was read from scope.InstanceGet("gorm:update_attrs").
// In gorm.io/gorm the resolved update target is db.Statement.Dest.
func updatedDataMap(db *gorm.DB) map[string]interface{} {
	if db.Statement == nil {
		return nil
	}
	if destMap, ok := db.Statement.Dest.(map[string]interface{}); ok {
		data := make(map[string]interface{}, len(destMap))
		for column, value := range destMap {
			if isSensitive(column) {
				continue
			}
			data[column] = value
		}
		return data
	}
	return currentDataMap(db)
}

// conditionData records the executed SQL statement for a bulk operation, since
// there is no single primary key to identify affected rows. Read from
// db.Statement.SQL in an After callback so the builder has already finished.
// Args are always redacted because they may contain user data.
func conditionData(db *gorm.DB) map[string]interface{} {
	if db.Statement == nil {
		return nil
	}

	sql := db.Statement.SQL.String()
	if sql == "" {
		return nil
	}

	condition := sql
	if isSensitive(condition) {
		condition = "[redacted]"
	}

	return map[string]interface{}{
		"condition": condition,
		"args":      "[redacted]",
	}
}

// skipField reports whether a schema field should be excluded from data maps.
// Mirrors jinzhu's field.IsIgnored || !field.IsNormal || field.Relationship != nil.
func skipField(f *schema.Field) bool {
	// Ignore fields that are not read/written to the DB column directly.
	if f.IgnoreMigration {
		return true
	}
	// Ignore fields with no DB column (e.g. has-many, belongs-to associations).
	if f.DataType == "" {
		return true
	}
	// Ignore fields tagged with gorm:"-".
	if !f.Readable {
		return true
	}
	return false
}

// toDataMap coerces an interface{} stashed via InstanceSet back to a data map.
func toDataMap(value interface{}) map[string]interface{} {
	if data, ok := value.(map[string]interface{}); ok {
		return data
	}
	return nil
}

// guard runs a callback body while guaranteeing the activity-log layer can never
// break the host database operation: any panic is captured to Sentry and
// swallowed.
func guard(db *gorm.DB, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("activitylog: recovered from panic in gorm callback: %v", r)
		}
	}()
	fn()
}
