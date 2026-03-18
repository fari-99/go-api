// Package go2rtc provides a Go client for the go2rtc REST API.
// Supports adding, updating, deleting streams and resyncing from a config source.
package go2rtc_helper

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// ─── Types ─────────────────────────────────────────────────────────────────────

// Stream represents a single go2rtc stream entry.
// The Source field is the RTSP URL (or any go2rtc-supported source).
type Stream struct {
	Name   string `json:"name"`
	Source string `json:"source"` // e.g. rtsp://user:pass@192.168.1.100:554/stream1
}

// StreamInfo is the response shape returned by go2rtc's GET /api/streams.
// go2rtc returns a map of stream_name → list of producers/consumers.
type StreamInfo map[string]any

// Client is the go2rtc API client.
type Client struct {
	resty   *resty.Client
	baseURL string
}

// Error wraps a go2rtc API error response.
type Error struct {
	StatusCode int
	Body       string
}

func (e *Error) Error() string {
	return fmt.Sprintf("go2rtc API error %d: %s", e.StatusCode, e.Body)
}

// ─── Constructor ───────────────────────────────────────────────────────────────

// New creates a new go2rtc Client.
func New() (*Client, error) {
	baseUrl := os.Getenv("GO2RTC_BASE_URL")
	if baseUrl == "" {
		return nil, errors.New("GO2RTC_BASE_URL environment variable not set")
	}

	r := resty.New().
		SetBaseURL(baseUrl).
		SetTimeout(10*time.Second).
		SetHeader("Content-Type", "application/json")

	isDebug, _ := strconv.ParseBool(os.Getenv("GO2RTC_DEBUG"))
	r.SetDebug(isDebug)

	c := &Client{resty: r, baseURL: baseUrl}
	return c, nil
}

// ─── Options ───────────────────────────────────────────────────────────────────

// Option is a functional option for Client.
type Option func(*Client)

// WithTimeout overrides the default 10s HTTP timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.resty.SetTimeout(d) }
}

// WithBasicAuth sets HTTP Basic Auth credentials (if go2rtc is behind auth).
func WithBasicAuth(username, password string) Option {
	return func(c *Client) { c.resty.SetBasicAuth(username, password) }
}

// StreamExists checks whether a stream with the given name is loaded in go2rtc.
func (c *Client) StreamExists(name string) (bool, error) {
	streams, err := c.ListStreams()
	if err != nil {
		return false, err
	}
	_, ok := streams[name]
	return ok, nil
}

func checkStatus(resp *resty.Response) error {
	if resp.StatusCode() >= http.StatusBadRequest {
		return &Error{
			StatusCode: resp.StatusCode(),
			Body:       string(resp.Body()),
		}
	}
	return nil
}

type InputGo2RTC struct {
	Name     string
	Url      string
	Username string
	Password string
}

// GenerateUrl generate url for go2rtc stream
func GenerateUrl(input InputGo2RTC) string {
	var url string
	var trimTrueUrl string
	if strings.HasPrefix(input.Url, "rtsp://") {
		trimTrueUrl = strings.TrimPrefix(input.Url, "rtsp://")
		url = "rtsp://"
	} else {
		return ""
	}

	if input.Username != "" && input.Password != "" {
		url += input.Username + ":" + input.Password + "@"
	}

	url += trimTrueUrl
	return url
}
