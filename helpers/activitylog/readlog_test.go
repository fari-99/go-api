package activitylog

import (
	"strings"
	"testing"

	"gorm.io/gorm"
)

// newReadTestDB is newTestDB plus read logging; reads() returns only READ entries.
func newReadTestDB(t *testing.T, exclude ...string) (db *gorm.DB, got *collector, reads func() []ActivityLog) {
	t.Helper()

	db, got = newTestDB(t)
	if err := RegisterReadCallbacks(db, ReadLogConfig{ExcludeTables: exclude}); err != nil {
		t.Fatalf("register read callbacks: %v", err)
	}

	reads = func() []ActivityLog {
		var out []ActivityLog
		for _, e := range got.snapshot() {
			if e.Action == ActionRead {
				out = append(out, e)
			}
		}
		return out
	}
	return db, got, reads
}

func TestReadLogEmitsQueryNotRows(t *testing.T) {
	db, _, reads := newReadTestDB(t)

	if err := db.Create(&txTestUser{Name: "alice", Email: "a@x.io"}).Error; err != nil {
		t.Fatal(err)
	}

	var users []txTestUser
	if err := db.Where("status = ?", 0).Find(&users).Error; err != nil {
		t.Fatal(err)
	}

	got := reads()
	if len(got) != 1 {
		t.Fatalf("reads = %d, want 1: %+v", len(got), got)
	}
	e := got[0]
	if e.Entity != "users" {
		t.Errorf("entity = %q, want users", e.Entity)
	}
	if !strings.Contains(strings.ToUpper(e.Query), "SELECT") || !strings.Contains(e.Query, "status = ?") {
		t.Errorf("query = %q, want the SELECT with placeholders", e.Query)
	}
	if strings.Contains(e.Query, "alice") {
		t.Errorf("query leaked data: %q", e.Query)
	}
	if e.Rows == nil || *e.Rows != 1 {
		t.Errorf("rows = %v, want 1", e.Rows)
	}
	if e.OldData != nil || e.NewData != nil || e.Diff != nil {
		t.Errorf("READ must not carry row data: %+v", e)
	}
}

func TestReadLogCoversScanAndCount(t *testing.T) {
	db, _, reads := newReadTestDB(t)

	var count int64
	if err := db.Model(&txTestUser{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	var names []string
	if err := db.Raw("SELECT name FROM users").Scan(&names).Error; err != nil {
		t.Fatal(err)
	}

	if n := len(reads()); n != 2 {
		t.Fatalf("reads = %d, want 2 (Count + Raw.Scan): %+v", n, reads())
	}
}

// Writes must not produce READ noise: the old-row lookup before an update or
// delete is internal and has to be skipped.
func TestReadLogIgnoresInternalLookups(t *testing.T) {
	db, got, reads := newReadTestDB(t)

	user := txTestUser{Name: "before", Email: "u@x.io"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	user.Name = "after"
	if err := db.Save(&user).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Delete(&user).Error; err != nil {
		t.Fatal(err)
	}

	if n := len(reads()); n != 0 {
		t.Fatalf("internal lookups logged as reads: %+v", reads())
	}
	if len(got.snapshot()) != 3 {
		t.Fatalf("entries = %v, want CREATE UPDATE DELETE", got.actions())
	}
}

func TestReadLogSkipsFailedAndExcluded(t *testing.T) {
	db, _, reads := newReadTestDB(t, "USERS")

	var u txTestUser
	_ = db.First(&u, 999).Error // not found + excluded
	var users []txTestUser
	_ = db.Find(&users).Error // excluded

	if n := len(reads()); n != 0 {
		t.Fatalf("reads = %+v, want none", reads())
	}
}

func TestReadLogRedactsSensitiveQueries(t *testing.T) {
	db, _, reads := newReadTestDB(t)

	var users []txTestUser
	_ = db.Where("password = ?", "x").Find(&users).Error

	got := reads()
	if len(got) != 1 || got[0].Query != "[redacted]" {
		t.Fatalf("reads = %+v, want one redacted query", got)
	}
}

// Reads are emitted at once, not buffered with the transaction: the data was
// read whether or not the transaction commits.
func TestReadLogInTransactionEmitsImmediately(t *testing.T) {
	db, _, reads := newReadTestDB(t)

	tx := db.Begin()
	var users []txTestUser
	if err := tx.Find(&users).Error; err != nil {
		t.Fatal(err)
	}
	if n := len(reads()); n != 1 {
		t.Fatalf("reads before commit = %d, want 1", n)
	}
	tx.Rollback()
	if n := len(reads()); n != 1 {
		t.Fatalf("reads after rollback = %d, want 1", n)
	}
}

func TestReadLogOffByDefault(t *testing.T) {
	db, got := newTestDB(t) // no RegisterReadCallbacks
	var users []txTestUser
	_ = db.Find(&users).Error
	if len(got.snapshot()) != 0 {
		t.Fatalf("entries = %v, want none", got.actions())
	}
}

func TestReadLogDoesNotWarnWithoutActor(t *testing.T) {
	db, _, reads := newReadTestDB(t)
	var users []txTestUser
	_ = db.Find(&users).Error
	if e := reads()[0]; e.ActorUserID != 0 {
		t.Errorf("actor = %v, want 0", e.ActorUserID)
	}
}
