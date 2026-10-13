package activitylog

import (
	"strings"

	"gorm.io/gorm"
)

// internalQueryKey marks statements the activity log itself issues (the old-row
// lookup before an update/delete) so they are never logged as reads.
const internalQueryKey = "activitylog:internal_query"

// ReadLogConfig controls read (SELECT) logging.
type ReadLogConfig struct {
	// ExcludeTables are never logged (case-insensitive), e.g. high-traffic
	// lookup tables such as sessions or casbin rules.
	ExcludeTables []string
}

// systemQueryMarkers identify catalog / migration queries gorm issues itself.
var systemQueryMarkers = []string{
	"information_schema",
	"pg_catalog",
	"sqlite_master",
}

// RegisterReadCallbacks logs every SELECT as a READ entry holding the SQL
// statement (placeholders only, no argument values) and, where known, the row
// count — never the rows themselves. A "SELECT * FROM users" has no identifier
// to log, so the statement is the useful part.
//
// It is opt-in: nothing is registered, and so nothing costs anything, unless
// this is called (configs.SetConnection does so when ACTIVITY_LOG_READ_ENABLED
// is true). Reads are very high volume; prefer a broker driver over stdout and
// exclude noisy tables.
//
// Covers Find/First/Take/Count/Pluck (Query callbacks) and Scan/Rows/Row (Row
// callbacks), including db.Raw(...).Find/Scan. Failed queries (including
// "record not found") are not logged. Reads are emitted immediately, even
// inside a transaction that later rolls back: the data was read either way.
func RegisterReadCallbacks(db *gorm.DB, cfg ReadLogConfig) error {
	exclude := make(map[string]struct{}, len(cfg.ExcludeTables))
	for _, table := range cfg.ExcludeTables {
		if table = strings.ToLower(strings.TrimSpace(table)); table != "" {
			exclude[table] = struct{}{}
		}
	}

	if err := db.Callback().Query().After("gorm:query").Register("activitylog:read", logRead(exclude, true)); err != nil {
		return err
	}
	return db.Callback().Row().After("gorm:row").Register("activitylog:read_row", logRead(exclude, false))
}

// logRead builds the callback. knownRows is true for Query callbacks, where
// db.RowsAffected is the number of rows returned.
func logRead(exclude map[string]struct{}, knownRows bool) func(*gorm.DB) {
	return func(db *gorm.DB) {
		guard(db, func() {
			if db.Error != nil || db.Statement == nil {
				return
			}
			if _, internal := db.Statement.Settings.Load(internalQueryKey); internal {
				return
			}

			table := db.Statement.Table
			if _, skip := exclude[strings.ToLower(table)]; skip {
				return
			}

			query := db.Statement.SQL.String()
			if query == "" || isSystemQuery(query) {
				return
			}
			if isSensitive(query) {
				query = "[redacted]"
			}

			actor, ctx := resolveActor(db, false)
			entry := ActivityLog{
				Timestamp:   nowRFC3339(),
				ActorUserID: actor,
				Action:      ActionRead,
				Entity:      table,
				Query:       query,
			}
			if knownRows {
				rows := db.RowsAffected
				entry.Rows = &rows
			}

			emitter(entry, ctx)
		})
	}
}

func isSystemQuery(query string) bool {
	lower := strings.ToLower(query)
	for _, marker := range systemQueryMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}
