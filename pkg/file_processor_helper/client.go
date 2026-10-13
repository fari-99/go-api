// Package file_processor_helper is a template client for a third-party file processing service
// (summaries, transcriptions, ...). It uploads one file as multipart/form-data and returns the raw
// response. It only depends on the standard library and resty so it can be moved to go-helper.
package file_processor_helper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const defaultTimeout = 300 * time.Second
const defaultFieldName = "file"

// Config of the third-party service.
type Config struct {
	BaseURL string
	APIKey  string // optional, sent as "Authorization: Bearer <key>"
	Timeout time.Duration
	Debug   bool // resty debug logs request bodies, dev only
}

// ConfigFromEnv reads FILE_PROCESSOR_BASE_URL, FILE_PROCESSOR_API_KEY,
// FILE_PROCESSOR_TIMEOUT_SECONDS and FILE_PROCESSOR_DEBUG.
func ConfigFromEnv() Config {
	config := Config{
		BaseURL: strings.TrimSpace(os.Getenv("FILE_PROCESSOR_BASE_URL")),
		APIKey:  os.Getenv("FILE_PROCESSOR_API_KEY"),
		Timeout: defaultTimeout,
	}

	if seconds, err := strconv.Atoi(os.Getenv("FILE_PROCESSOR_TIMEOUT_SECONDS")); err == nil && seconds > 0 {
		config.Timeout = time.Duration(seconds) * time.Second
	}

	config.Debug, _ = strconv.ParseBool(os.Getenv("FILE_PROCESSOR_DEBUG"))
	return config
}

// Client is the third-party file processing client.
type Client struct {
	resty *resty.Client
}

// Error is returned when the service answers with HTTP status >= 400.
type Error struct {
	StatusCode int
	Body       string
}

func (e *Error) Error() string {
	return fmt.Sprintf("file processor API error %d: %s", e.StatusCode, e.Body)
}

// New creates a client. BaseURL is required.
func New(config Config) (*Client, error) {
	if config.BaseURL == "" {
		return nil, errors.New("file processor base url is not set")
	}

	if config.Timeout <= 0 {
		config.Timeout = defaultTimeout
	}

	r := resty.New().
		SetBaseURL(strings.TrimRight(config.BaseURL, "/")).
		SetTimeout(config.Timeout).
		SetHeader("Accept", "application/json").
		SetDebug(config.Debug)

	if config.APIKey != "" {
		r.SetAuthToken(config.APIKey)
	}

	return &Client{resty: r}, nil
}

// UploadRequest is one file sent to an endpoint of the service.
type UploadRequest struct {
	Endpoint  string // ex: /pdf/summary
	FieldName string // multipart file field, default "file"
	FileName  string
	File      io.Reader
	Fields    map[string]string // extra form fields, ex: action params
}

// UploadResponse is the raw answer of the service.
type UploadResponse struct {
	StatusCode int
	Body       []byte
	Latency    time.Duration
}

// Upload sends the file as multipart/form-data with POST and returns the response.
// HTTP status >= 400 is returned as *Error, together with the response.
func (c *Client) Upload(ctx context.Context, input UploadRequest) (*UploadResponse, error) {
	if input.File == nil {
		return nil, errors.New("upload file is empty")
	}

	fieldName := input.FieldName
	if fieldName == "" {
		fieldName = defaultFieldName
	}

	endpoint := "/" + strings.TrimLeft(input.Endpoint, "/")

	resp, err := c.resty.R().
		SetContext(ctx).
		SetFileReader(fieldName, input.FileName, input.File).
		SetMultipartFormData(input.Fields).
		Post(endpoint)
	if err != nil {
		return nil, fmt.Errorf("upload to %s: %w", endpoint, err)
	}

	response := &UploadResponse{
		StatusCode: resp.StatusCode(),
		Body:       resp.Body(),
		Latency:    resp.Time(),
	}

	if resp.StatusCode() >= http.StatusBadRequest {
		return response, &Error{StatusCode: resp.StatusCode(), Body: string(resp.Body())}
	}

	return response, nil
}
