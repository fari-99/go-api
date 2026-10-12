package activitylog

import (
	"context"
	"database/sql"
	"strings"
	"sync"

	"gorm.io/gorm"
)

// Transaction-aware emission.
//
// gorm has no "after commit" hook, so a callback that fires per statement
// cannot tell whether the surrounding transaction will survive. To get
// "emit only once the commit succeeded" without touching any call site, the
// gorm ConnPool is wrapped (see RegisterCallbacks):
//
//   - hookedPool wraps the root pool. Its BeginTx hands back a bufferedTx.
//   - bufferedTx wraps the real transaction. The gorm callbacks append their
//     entries to it instead of emitting; Commit flushes them (in order, after
//     the underlying commit succeeded), Rollback / a failed Commit drops them.
//
// This covers db.Begin()/Commit()/Rollback(), db.Transaction(...) (including
// nested ones, via SAVEPOINT tracking) and gorm's own per-statement default
// transaction, all with the same code path. Statements that run outside any
// transaction (SkipDefaultTransaction) have nothing to wait for and are emitted
// immediately.

// pendingEntry is a log held back until its transaction commits.
type pendingEntry struct {
	entry ActivityLog
	ctx   interface{}
}

// hookedPool wraps a gorm.ConnPool so transactions started from it are
// bufferedTx values.
type hookedPool struct {
	gorm.ConnPool
}

// BeginTx implements gorm.ConnPoolBeginner.
func (p *hookedPool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	var (
		tx  gorm.ConnPool
		err error
	)

	switch beginner := p.ConnPool.(type) {
	case gorm.TxBeginner:
		var sqlTx *sql.Tx
		sqlTx, err = beginner.BeginTx(ctx, opts)
		if err == nil {
			tx = sqlTx
		}
	case gorm.ConnPoolBeginner:
		tx, err = beginner.BeginTx(ctx, opts)
	default:
		err = gorm.ErrInvalidTransaction
	}
	if err != nil {
		return nil, err
	}

	return &bufferedTx{ConnPool: tx, pool: p}, nil
}

// GetDBConn implements gorm.GetDBConnector so db.DB() keeps working
// (SetMaxOpenConns, Ping, Close, ...).
func (p *hookedPool) GetDBConn() (*sql.DB, error) {
	switch pool := p.ConnPool.(type) {
	case gorm.GetDBConnector:
		return pool.GetDBConn()
	case *sql.DB:
		return pool, nil
	}
	return nil, gorm.ErrInvalidDB
}

// bufferedTx is a transaction that holds activity logs until it ends.
type bufferedTx struct {
	gorm.ConnPool // the real transaction (*sql.Tx or a prepared-stmt tx)
	pool          *hookedPool

	mu         sync.Mutex
	entries    []pendingEntry
	savepoints map[string]int // savepoint name -> len(entries) when it was set
}

// add holds an entry back until Commit.
func (t *bufferedTx) add(e pendingEntry) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries = append(t.entries, e)
}

// Commit commits the real transaction and, only if that succeeded, emits every
// buffered entry in the order the statements ran.
func (t *bufferedTx) Commit() error {
	committer, ok := t.ConnPool.(gorm.TxCommitter)
	if !ok {
		return gorm.ErrInvalidTransaction
	}

	if err := committer.Commit(); err != nil {
		t.drain() // the transaction is gone; nothing it did is real
		return err
	}

	for _, pending := range t.drain() {
		emitter(pending.entry, pending.ctx)
	}
	return nil
}

// Rollback rolls the real transaction back and drops everything buffered.
func (t *bufferedTx) Rollback() error {
	t.drain()

	committer, ok := t.ConnPool.(gorm.TxCommitter)
	if !ok {
		return gorm.ErrInvalidTransaction
	}
	return committer.Rollback()
}

// GetDBConn implements gorm.GetDBConnector so db.DB() works on a transaction.
func (t *bufferedTx) GetDBConn() (*sql.DB, error) {
	return t.pool.GetDBConn()
}

// ExecContext tracks SAVEPOINT / ROLLBACK TO so a rolled-back nested
// transaction (db.Transaction inside db.Transaction) drops the entries it
// produced, while the outer transaction's entries survive.
//
// Only the "SAVEPOINT <name>" / "ROLLBACK TO [SAVEPOINT] <name>" syntax used by
// the MySQL and PostgreSQL dialectors is recognised.
func (t *bufferedTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	result, err := t.ConnPool.ExecContext(ctx, query, args...)
	if err == nil {
		t.trackSavepoint(query)
	}
	return result, err
}

func (t *bufferedTx) trackSavepoint(query string) {
	words := strings.Fields(query)
	if len(words) < 2 {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	switch {
	case strings.EqualFold(words[0], "SAVEPOINT"):
		if t.savepoints == nil {
			t.savepoints = make(map[string]int)
		}
		t.savepoints[cleanSavepointName(words[1])] = len(t.entries)

	case len(words) >= 3 && strings.EqualFold(words[0], "ROLLBACK") && strings.EqualFold(words[1], "TO"):
		name := words[2]
		if strings.EqualFold(name, "SAVEPOINT") && len(words) >= 4 {
			name = words[3]
		}
		if mark, ok := t.savepoints[cleanSavepointName(name)]; ok && mark <= len(t.entries) {
			t.entries = t.entries[:mark]
		}
	}
}

func cleanSavepointName(name string) string {
	return strings.Trim(name, "`\"';")
}

// drain returns the buffered entries and empties the buffer.
func (t *bufferedTx) drain() []pendingEntry {
	t.mu.Lock()
	defer t.mu.Unlock()

	entries := t.entries
	t.entries = nil
	t.savepoints = nil
	return entries
}

// txBufferOf returns the bufferedTx behind a statement's ConnPool, or nil when
// the statement is not running inside a transaction we are tracking.
func txBufferOf(pool gorm.ConnPool) *bufferedTx {
	if buffered, ok := pool.(*bufferedTx); ok {
		return buffered
	}
	return nil
}

// dispatch builds the entry for a callback and either holds it until the
// enclosing transaction commits, or emits it straight away when there is none.
func dispatch(db *gorm.DB, actorUserID interface{}, action, entity string, entityID interface{}, oldData, newData map[string]interface{}, ctx interface{}) {
	entry := buildEntry(actorUserID, action, entity, entityID, oldData, newData)

	if tx := txBufferOf(db.Statement.ConnPool); tx != nil {
		tx.add(pendingEntry{entry: entry, ctx: ctx})
		return
	}
	emitter(entry, ctx)
}
