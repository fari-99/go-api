package activitylog

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type txTestUser struct {
	ID       uint64 `gorm:"primaryKey"`
	Name     string
	Email    string `gorm:"uniqueIndex"`
	Status   int
	Password string
}

func (txTestUser) TableName() string { return "users" }

// collector replaces the package sink for the duration of a test.
type collector struct {
	mu      sync.Mutex
	entries []ActivityLog
}

func (c *collector) add(entry ActivityLog, _ ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, entry)
}

func (c *collector) snapshot() []ActivityLog {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]ActivityLog(nil), c.entries...)
}

func (c *collector) actions() []string {
	var actions []string
	for _, e := range c.snapshot() {
		actions = append(actions, e.Action)
	}
	return actions
}

// newTestDB opens a throw-away sqlite file (a file, not :memory:, so every pool
// connection sees the same data) with the activity-log callbacks registered and
// the sink replaced by a collector.
func newTestDB(t *testing.T) (*gorm.DB, *collector) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&txTestUser{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	got := &collector{}
	orig := emitter
	emitter = got.add
	t.Cleanup(func() { emitter = orig })

	if err := RegisterCallbacks(db); err != nil {
		t.Fatalf("register callbacks: %v", err)
	}
	return db, got
}

func assertActions(t *testing.T, got *collector, want ...string) {
	t.Helper()
	actual := got.actions()
	if len(actual) != len(want) {
		t.Fatalf("emitted actions = %v, want %v", actual, want)
	}
	for i := range want {
		if actual[i] != want[i] {
			t.Fatalf("emitted actions = %v, want %v", actual, want)
		}
	}
}

// A statement outside an explicit transaction runs in gorm's own default
// transaction; its log must be emitted once that commits.
func TestAutoCommitStatementEmits(t *testing.T) {
	db, got := newTestDB(t)

	user := txTestUser{Name: "a", Email: "a@x.io", Password: "secret"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create: %v", err)
	}

	assertActions(t, got, ActionCreate)
	entry := got.snapshot()[0]
	if entry.Entity != "users" || entry.EntityID != user.ID {
		t.Errorf("entry = %+v, want entity users / id %d", entry, user.ID)
	}
	if entry.NewData["email"] != "a@x.io" {
		t.Errorf("new_data = %v, want email a@x.io", entry.NewData)
	}
	if _, ok := entry.NewData["password"]; ok {
		t.Error("new_data leaked the password column")
	}
}

// Nothing is emitted while the transaction is open; everything is emitted, in
// statement order, once Commit succeeds.
func TestExplicitTransactionCommitEmitsInBulk(t *testing.T) {
	db, got := newTestDB(t)

	tx := db.Begin()
	a := txTestUser{Name: "a", Email: "a@x.io"}
	b := txTestUser{Name: "b", Email: "b@x.io"}
	for _, u := range []*txTestUser{&a, &b} {
		if err := tx.Create(u).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
	}
	if err := tx.Model(&a).Update("name", "a2").Error; err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := tx.Delete(&b).Error; err != nil {
		t.Fatalf("delete: %v", err)
	}

	if n := len(got.snapshot()); n != 0 {
		t.Fatalf("emitted %d entries before commit, want 0", n)
	}

	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}
	assertActions(t, got, ActionCreate, ActionCreate, ActionUpdate, ActionDelete)
}

func TestExplicitTransactionRollbackEmitsNothing(t *testing.T) {
	db, got := newTestDB(t)

	tx := db.Begin()
	if err := tx.Create(&txTestUser{Name: "a", Email: "a@x.io"}).Error; err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := tx.Rollback().Error; err != nil {
		t.Fatalf("rollback: %v", err)
	}

	assertActions(t, got)

	var count int64
	db.Model(&txTestUser{}).Count(&count)
	if count != 0 {
		t.Fatalf("row survived the rollback (count=%d)", count)
	}
}

func TestTransactionHelperErrorEmitsNothing(t *testing.T) {
	db, got := newTestDB(t)

	boom := errors.New("boom")
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&txTestUser{Name: "a", Email: "a@x.io"}).Error; err != nil {
			return err
		}
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want boom", err)
	}

	assertActions(t, got)
}

func TestTransactionHelperSuccessEmits(t *testing.T) {
	db, got := newTestDB(t)

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&txTestUser{Name: "a", Email: "a@x.io"}).Error; err != nil {
			return err
		}
		return tx.Create(&txTestUser{Name: "b", Email: "b@x.io"}).Error
	})
	if err != nil {
		t.Fatalf("transaction: %v", err)
	}

	assertActions(t, got, ActionCreate, ActionCreate)
}

// A statement that fails (here: unique violation) must not be logged, and the
// failing transaction must not take earlier, rolled-back work with it.
func TestFailedStatementEmitsNothing(t *testing.T) {
	db, got := newTestDB(t)

	if err := db.Create(&txTestUser{Name: "a", Email: "dup@x.io"}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	assertActions(t, got, ActionCreate)

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&txTestUser{Name: "b", Email: "b@x.io"}).Error; err != nil {
			return err
		}
		return tx.Create(&txTestUser{Name: "c", Email: "dup@x.io"}).Error // fails
	})
	if err == nil {
		t.Fatal("expected unique violation")
	}

	assertActions(t, got, ActionCreate) // still only the seed
}

// A rolled-back nested transaction (SAVEPOINT) drops only its own entries.
func TestNestedTransactionRollbackDropsOnlyInnerEntries(t *testing.T) {
	db, got := newTestDB(t)

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&txTestUser{Name: "outer", Email: "outer@x.io"}).Error; err != nil {
			return err
		}

		_ = tx.Transaction(func(inner *gorm.DB) error {
			if err := inner.Create(&txTestUser{Name: "inner", Email: "inner@x.io"}).Error; err != nil {
				return err
			}
			return errors.New("inner failed")
		})

		return tx.Create(&txTestUser{Name: "outer2", Email: "outer2@x.io"}).Error
	})
	if err != nil {
		t.Fatalf("transaction: %v", err)
	}

	assertActions(t, got, ActionCreate, ActionCreate)
	for _, e := range got.snapshot() {
		if e.NewData["name"] == "inner" {
			t.Fatalf("rolled-back inner entry was emitted: %+v", e)
		}
	}

	var count int64
	db.Model(&txTestUser{}).Count(&count)
	if count != 2 {
		t.Fatalf("row count = %d, want 2 (inner row rolled back)", count)
	}
}

// old_data must come from the database row, not from the in-memory model that
// is about to be written, otherwise Save() produces an empty diff.
func TestUpdateCapturesOldValuesFromDatabase(t *testing.T) {
	db, got := newTestDB(t)

	user := txTestUser{Name: "before", Email: "u@x.io", Status: 1}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create: %v", err)
	}

	user.Name = "after"
	if err := db.Save(&user).Error; err != nil {
		t.Fatalf("save: %v", err)
	}

	assertActions(t, got, ActionCreate, ActionUpdate)
	diff := got.snapshot()[1].Diff
	if d, ok := diff["name"]; !ok || d.Old != "before" || d.New != "after" {
		t.Fatalf("diff = %+v, want name before -> after", diff)
	}
	if _, ok := diff["email"]; ok {
		t.Errorf("diff contains unchanged column email: %+v", diff)
	}
}

func TestDeleteCapturesOldData(t *testing.T) {
	db, got := newTestDB(t)

	user := txTestUser{Name: "gone", Email: "g@x.io", Password: "secret"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := db.Delete(&user).Error; err != nil {
		t.Fatalf("delete: %v", err)
	}

	assertActions(t, got, ActionCreate, ActionDelete)
	entry := got.snapshot()[1]
	if entry.OldData["name"] != "gone" {
		t.Errorf("old_data = %v, want name gone", entry.OldData)
	}
	if _, ok := entry.OldData["password"]; ok {
		t.Error("old_data leaked the password column")
	}
}

func TestBulkOperationsAreLoggedAsBulk(t *testing.T) {
	db, got := newTestDB(t)

	batch := []txTestUser{{Name: "a", Email: "a@x.io"}, {Name: "b", Email: "b@x.io"}}
	if err := db.Create(&batch).Error; err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := db.Model(&txTestUser{}).Where("status = ?", 0).Update("status", 5).Error; err != nil {
		t.Fatalf("bulk update: %v", err)
	}
	if err := db.Where("status = ?", 5).Delete(&txTestUser{}).Error; err != nil {
		t.Fatalf("bulk delete: %v", err)
	}

	assertActions(t, got, ActionBulkCreate, ActionBulkUpdate, ActionBulkDelete)
}

// Wrapping the pool must not break db.DB() — on the root and inside a tx.
func TestWrappedPoolStillExposesSQLDB(t *testing.T) {
	db, _ := newTestDB(t)

	if sqlDB, err := db.DB(); err != nil || sqlDB == nil {
		t.Fatalf("db.DB() = %v, %v", sqlDB, err)
	}

	tx := db.Begin()
	defer tx.Rollback()
	if sqlDB, err := tx.DB(); err != nil || sqlDB == nil {
		t.Fatalf("tx.DB() = %v, %v", sqlDB, err)
	}
}

func TestRegisterCallbacksTwiceDoesNotDoubleWrap(t *testing.T) {
	db, _ := newTestDB(t)

	wrapped := db.ConnPool
	// Second registration fails on the duplicate callback name; the pool must
	// not have been wrapped again on the way there.
	_ = RegisterCallbacks(db)
	if db.ConnPool != wrapped {
		t.Fatal("pool was wrapped twice")
	}
}
