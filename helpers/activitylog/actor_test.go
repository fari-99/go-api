package activitylog

import (
	"testing"

	"gorm.io/gorm"
)

func TestWithActorAttributesStatementsAndTransactions(t *testing.T) {
	db, got := newTestDB(t)
	actor := WithActor(db, uint64(77))

	// reusable across statements
	if err := actor.Create(&txTestUser{Name: "a", Email: "a@x.io"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := actor.Create(&txTestUser{Name: "b", Email: "b@x.io"}).Error; err != nil {
		t.Fatal(err)
	}
	// carried into db.Transaction
	err := actor.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&txTestUser{Name: "c", Email: "c@x.io"}).Error
	})
	if err != nil {
		t.Fatal(err)
	}
	// carried into Begin
	tx := actor.Begin()
	if err := tx.Create(&txTestUser{Name: "d", Email: "d@x.io"}).Error; err != nil {
		t.Fatal(err)
	}
	tx.Commit()

	entries := got.snapshot()
	if len(entries) != 4 {
		t.Fatalf("entries = %d, want 4", len(entries))
	}
	for i, e := range entries {
		if e.ActorUserID != uint64(77) {
			t.Errorf("entry %d actor = %v, want 77", i, e.ActorUserID)
		}
	}
}

func TestWithoutActorDefaultsToZero(t *testing.T) {
	db, got := newTestDB(t)
	if err := db.Create(&txTestUser{Name: "a", Email: "a@x.io"}).Error; err != nil {
		t.Fatal(err)
	}
	if e := got.snapshot()[0]; e.ActorUserID != 0 {
		t.Errorf("actor = %v, want 0", e.ActorUserID)
	}
}

// Documents what is NOT covered: callbacks only see gorm model operations.
func TestWhatIsNotLogged(t *testing.T) {
	db, got := newTestDB(t)

	if err := db.Create(&txTestUser{Name: "a", Email: "a@x.io"}).Error; err != nil {
		t.Fatal(err)
	}
	base := len(got.snapshot())

	db.Exec("UPDATE users SET name = ?", "raw")                                      // raw SQL
	db.Table("users").Create(map[string]interface{}{"name": "m", "email": "m@x.io"}) // map insert, no model
	var users []txTestUser
	db.Find(&users) // reads

	if n := len(got.snapshot()); n != base {
		t.Fatalf("raw SQL / map insert / reads produced %d extra entries, want 0", n-base)
	}
}
