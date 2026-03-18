package go2rtc_helper

import (
	"encoding/json"
	"fmt"
)

// Ping checks if the go2rtc instance is reachable and returns its version string.
func (c *Client) Ping() (string, error) {
	resp, err := c.resty.R().Get("/api")
	if err != nil {
		return "", fmt.Errorf("Ping: %w", err)
	}
	if err := checkStatus(resp); err != nil {
		return "", fmt.Errorf("Ping: %w", err)
	}

	// go2rtc /api returns a JSON object with version info
	var info map[string]any
	if err := json.Unmarshal(resp.Body(), &info); err != nil {
		return string(resp.Body()), nil
	}
	if v, ok := info["version"].(string); ok {
		return v, nil
	}
	return string(resp.Body()), nil
}
