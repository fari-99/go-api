package go2rtc_helper

import (
	"fmt"
)

// ListStreams returns all streams currently loaded in go2rtc.
func (c *Client) ListStreams() (StreamInfo, error) {
	resp, err := c.resty.R().
		SetResult(&StreamInfo{}).
		Get("/api/streams")

	if err != nil {
		return nil, fmt.Errorf("ListStreams: %w", err)
	}
	if err := checkStatus(resp); err != nil {
		return nil, fmt.Errorf("ListStreams: %w", err)
	}

	return *resp.Result().(*StreamInfo), nil
}
