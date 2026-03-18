package go2rtc_helper

import (
	"fmt"
)

// ─── Resync ────────────────────────────────────────────────────────────────────

// SyncRequest describes the desired state of streams in go2rtc.
type SyncRequest struct {
	// Streams is the full desired list. Resync will add missing ones,
	// update changed ones, and delete any not in this list.
	Streams []Stream
}

// SyncResult reports what Resync did.
type SyncResult struct {
	Added   []string
	Updated []string
	Deleted []string
	Errors  []SyncError
}

// SyncError records a failure for a single stream during resync.
type SyncError struct {
	Name      string
	Operation string // "add" | "update" | "delete"
	Err       error
}

func (e SyncError) Error() string {
	return fmt.Sprintf("[%s] %s: %v", e.Operation, e.Name, e.Err)
}

// Resync reconciles go2rtc's loaded streams with the desired state in req.
//
// Algorithm:
//  1. Fetch current streams from go2rtc.
//  2. For each desired stream:
//     - If missing → Add.
//     - If present  → Update (re-PUT to refresh the source URL).
//  3. For each current stream not in desired list → Delete.
//
// Errors per stream are collected and returned in SyncResult.Errors rather
// than aborting the whole sync. Check len(result.Errors) > 0 to detect partial failures.
func (c *Client) Resync(req SyncRequest) (SyncResult, error) {
	result := SyncResult{}

	current, err := c.ListStreams()
	if err != nil {
		return result, fmt.Errorf("Resync: failed to list current streams: %w", err)
	}

	// Build a lookup map for desired streams
	desired := make(map[string]string, len(req.Streams)) // name → rtspURL
	for _, s := range req.Streams {
		desired[s.Name] = s.Source
	}

	// Add or update desired streams
	for name, rtspURL := range desired {
		if _, exists := current[name]; !exists {
			// Stream is new — add it
			if err := c.AddStream(name, rtspURL); err != nil {
				result.Errors = append(result.Errors, SyncError{Name: name, Operation: "add", Err: err})
			} else {
				result.Added = append(result.Added, name)
			}
		} else {
			// Stream exists — update to ensure source is current
			if err := c.UpdateStream(name, rtspURL); err != nil {
				result.Errors = append(result.Errors, SyncError{Name: name, Operation: "update", Err: err})
			} else {
				result.Updated = append(result.Updated, name)
			}
		}
	}

	// Delete streams that are no longer desired
	for name := range current {
		if _, wanted := desired[name]; !wanted {
			if err := c.DeleteStream(name); err != nil {
				result.Errors = append(result.Errors, SyncError{Name: name, Operation: "delete", Err: err})
			} else {
				result.Deleted = append(result.Deleted, name)
			}
		}
	}

	return result, nil
}
