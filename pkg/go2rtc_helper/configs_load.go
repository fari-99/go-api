package go2rtc_helper

import (
	"fmt"
)

// ─── Config Reload ─────────────────────────────────────────────────────────────

// ReloadConfig triggers go2rtc to reload its config file / HTTP config source.
// Use this after bulk changes if go2rtc is configured with a dynamic config URL.
func (c *Client) ReloadConfig() error {
	resp, err := c.resty.R().Post("/api/config/reload")
	if err != nil {
		return fmt.Errorf("ReloadConfig: %w", err)
	}
	if err := checkStatus(resp); err != nil {
		return fmt.Errorf("ReloadConfig: %w", err)
	}
	return nil
}
