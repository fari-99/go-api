package go2rtc_helper

import (
	"fmt"
)

// UpdateStream updates an existing stream's source URL.
// go2rtc has no separate PATCH endpoint — update is the same as add (PUT).
// The old stream is deleted first to avoid duplicate producers.
func (c *Client) UpdateStream(name, newRTSPURL string) error {
	// Delete first to cleanly replace
	if err := c.DeleteStream(name); err != nil {
		return fmt.Errorf("UpdateStream %q (delete phase): %w", name, err)
	}
	if err := c.AddStream(name, newRTSPURL); err != nil {
		return fmt.Errorf("UpdateStream %q (add phase): %w", name, err)
	}
	return nil
}
