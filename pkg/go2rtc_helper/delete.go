package go2rtc_helper

import (
	"fmt"
)

// DeleteStream removes a stream from go2rtc by name.
// go2rtc uses DELETE /api/streams?name=<name>
func (c *Client) DeleteStream(name string) error {
	resp, err := c.resty.R().
		SetQueryParam("src", name).
		Delete("/api/streams")

	if err != nil {
		return fmt.Errorf("DeleteStream %q: %w", name, err)
	}
	// go2rtc returns 200 even if stream didn't exist — that's fine
	if err := checkStatus(resp); err != nil {
		return fmt.Errorf("DeleteStream %q: %w", name, err)
	}
	return nil
}
