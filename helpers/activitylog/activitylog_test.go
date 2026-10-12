package activitylog

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

// TestBuildEntry verifies the payload carries every required field, stamps the
// timestamp as RFC3339 UTC, strips sensitive columns, and — for UPDATE — emits
// a per-field diff instead of separate old/new maps.
func TestBuildEntry(t *testing.T) {
	entry := buildEntry(
		int64(42),
		ActionUpdate,
		"users",
		7,
		map[string]interface{}{"email": "old@mbiz.co.id", "status": 1, "password": "should-be-removed"},
		map[string]interface{}{"email": "new@mbiz.co.id", "password_hash": "also-removed"},
	)

	if entry.ActorUserID != int64(42) {
		t.Errorf("actor_user_id = %v, want 42", entry.ActorUserID)
	}
	if entry.Action != ActionUpdate {
		t.Errorf("action = %q, want %q", entry.Action, ActionUpdate)
	}
	if entry.Entity != "users" {
		t.Errorf("entity = %q, want %q", entry.Entity, "users")
	}
	if entry.EntityID != 7 {
		t.Errorf("entity_id = %v, want 7", entry.EntityID)
	}

	// Timestamp must be RFC3339 and in UTC (trailing Z).
	if _, err := time.Parse(time.RFC3339, entry.Timestamp); err != nil {
		t.Errorf("timestamp %q not RFC3339: %v", entry.Timestamp, err)
	}
	if !strings.HasSuffix(entry.Timestamp, "Z") {
		t.Errorf("timestamp %q is not UTC (no trailing Z)", entry.Timestamp)
	}

	// UPDATE populates diff only — old_data/new_data must stay unset.
	if entry.OldData != nil {
		t.Errorf("old_data = %v, want nil for UPDATE", entry.OldData)
	}
	if entry.NewData != nil {
		t.Errorf("new_data = %v, want nil for UPDATE", entry.NewData)
	}

	// Sensitive columns must be gone from the diff.
	if _, ok := entry.Diff["password"]; ok {
		t.Error("diff still contains sensitive column 'password'")
	}
	if _, ok := entry.Diff["password_hash"]; ok {
		t.Error("diff still contains sensitive column 'password_hash'")
	}
	// "email" changed, so it must appear in the diff with both values.
	if got := entry.Diff["email"]; got.Old != "old@mbiz.co.id" || got.New != "new@mbiz.co.id" {
		t.Errorf("diff email = %+v, want {old@mbiz.co.id new@mbiz.co.id}", got)
	}
	// "status" only existed in oldData (unchanged, not part of the update), so
	// it must not appear in the diff.
	if _, ok := entry.Diff["status"]; ok {
		t.Error("diff contains 'status', which was not part of newData")
	}
}

// TestBuildEntryCreate verifies CREATE emits new_data only (no old_data, no
// diff), and marshals cleanly with the expected keys.
func TestBuildEntryCreate(t *testing.T) {
	entry := buildEntry(1, ActionCreate, "companies", 99, nil, map[string]interface{}{"name": "PT Mbiz"})

	if entry.OldData != nil {
		t.Errorf("old_data = %v, want nil for CREATE", entry.OldData)
	}
	if entry.Diff != nil {
		t.Errorf("diff = %v, want nil for CREATE", entry.Diff)
	}
	if entry.NewData["name"] != "PT Mbiz" {
		t.Errorf("new_data name = %v, want PT Mbiz", entry.NewData["name"])
	}

	b, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	for _, key := range []string{"timestamp", "actor_user_id", "action", "entity", "entity_id", "new_data"} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("marshaled log missing required key %q", key)
		}
	}
	// old_data/diff are omitempty — they must not appear at all for CREATE.
	if _, ok := decoded["old_data"]; ok {
		t.Errorf("old_data present in CREATE payload, want omitted")
	}
	if _, ok := decoded["diff"]; ok {
		t.Errorf("diff present in CREATE payload, want omitted")
	}
}

// TestBuildEntryDelete verifies DELETE emits old_data only (no new_data, no diff).
func TestBuildEntryDelete(t *testing.T) {
	entry := buildEntry(1, ActionDelete, "users", 5, map[string]interface{}{"email": "gone@mbiz.co.id"}, nil)

	if entry.NewData != nil {
		t.Errorf("new_data = %v, want nil for DELETE", entry.NewData)
	}
	if entry.Diff != nil {
		t.Errorf("diff = %v, want nil for DELETE", entry.Diff)
	}
	if entry.OldData["email"] != "gone@mbiz.co.id" {
		t.Errorf("old_data email = %v, want gone@mbiz.co.id", entry.OldData["email"])
	}
}

// TestRemoveSensitiveNilSafe verifies nil/empty maps do not cause runtime errors
// (acceptance criterion: partial/empty maps must be safe).
func TestRemoveSensitiveNilSafe(t *testing.T) {
	if got := RemoveSensitive(nil); got != nil {
		t.Errorf("RemoveSensitive(nil) = %v, want nil", got)
	}
	if got := RemoveSensitive(map[string]interface{}{}); got == nil || len(got) != 0 {
		t.Errorf("RemoveSensitive(empty) = %v, want empty map", got)
	}

	// A map of only sensitive keys reduces to empty, not nil.
	got := RemoveSensitive(map[string]interface{}{"password": "x", "user_token": "y"})
	if len(got) != 0 {
		t.Errorf("RemoveSensitive(all-sensitive) = %v, want empty", got)
	}
}

// TestLogActivityNilMapsNoPanic asserts that passing nil for oldData and newData
// does not panic and still emits a valid line.
func TestLogActivityNilMapsNoPanic(t *testing.T) {
	line := captureStdout(t, func() {
		LogActivity(int64(0), ActionDelete, "users", 5, nil, nil)
	})

	var decoded ActivityLog
	if err := json.Unmarshal([]byte(line), &decoded); err != nil {
		t.Fatalf("emitted line is not valid JSON: %v (line=%q)", err, line)
	}
	if decoded.Action != ActionDelete || decoded.Entity != "users" {
		t.Errorf("unexpected payload: %+v", decoded)
	}
}

// TestLogActivityStdout verifies the helper writes a single JSON line to stdout
// with an UPDATE diff reflecting old vs new.
func TestLogActivityStdout(t *testing.T) {
	line := captureStdout(t, func() {
		LogActivity(int64(10), ActionUpdate, "users", 3,
			map[string]interface{}{"status": 1},
			map[string]interface{}{"status": 99})
	})

	var decoded ActivityLog
	if err := json.Unmarshal([]byte(line), &decoded); err != nil {
		t.Fatalf("emitted line is not valid JSON: %v (line=%q)", err, line)
	}
	if decoded.Diff["status"].Old != float64(1) { // JSON numbers decode to float64
		t.Errorf("diff status old = %v, want 1", decoded.Diff["status"].Old)
	}
	if decoded.Diff["status"].New != float64(99) {
		t.Errorf("diff status new = %v, want 99", decoded.Diff["status"].New)
	}
}

// captureStdout runs fn with os.Stdout redirected to a pipe and returns the
// first line written.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()
	_ = w.Close()

	line, err := bufio.NewReader(r).ReadString('\n')
	if err != nil && line == "" {
		t.Fatalf("failed to read captured stdout: %v", err)
	}
	return strings.TrimSpace(line)
}
