package go2rtc_helper

import (
	"fmt"
)

// AddStream adds a new stream to go2rtc.
// go2rtc uses PUT /api/streams?name=<name>&src=<rtsp_url>
// If the stream name already exists, go2rtc will overwrite it.
func (c *Client) AddStream(name, rtspURL string) error {
	resp, err := c.resty.R().
		SetQueryParams(map[string]string{
			"name": name,
			"src":  rtspURL,
		}).
		Put("/api/streams")

	if err != nil {
		return fmt.Errorf("AddStream %q: %w", name, err)
	}
	if err := checkStatus(resp); err != nil {
		return fmt.Errorf("AddStream %q: %w", name, err)
	}
	return nil
}
