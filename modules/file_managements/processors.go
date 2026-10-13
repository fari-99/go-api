package file_managements

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"go-api/modules/models"
	fileprocessor "go-api/pkg/file_processor_helper"
)

// Processor runs one action (pdf_summary, audio_transcription, ...) for a file.
// Add a new action by implementing it and registering it in RegistryFromEnv.
type Processor interface {
	Action() string
	// OutputType is the meta type used for group=output
	OutputType() string
	AcceptedMimes() []string
	// AcceptedExtensions is the fallback when the sniffed mime is application/octet-stream
	AcceptedExtensions() []string
	MaxFileBytes() int64
	Params() []ParamSpec
	// ValidateParams rejects unknown or invalid params and returns them normalized to strings
	ValidateParams(params map[string]any) (map[string]string, error)
	// Process may return a non-nil *Result together with an error, to keep the request meta
	Process(ctx context.Context, job Job) (*Result, error)
}

type Job struct {
	FileManagement models.FileManagements
	Storage        models.Storages
	File           io.Reader
	Params         map[string]string
}

type Result struct {
	Output  map[string]string // stored as group=output, type=OutputType()
	Request map[string]string // stored as group=request, type=third_party, never shown to users
}

// ProcessError is a failed job. Message is shown to the user, Internal is stored but never returned.
type ProcessError struct {
	Code     string
	Message  string
	Internal string
}

func (e *ProcessError) Error() string {
	if e.Internal != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Internal)
	}
	return e.Code
}

// ParamSpec describes one accepted param, also returned by GET /file-managements/actions
type ParamSpec struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"` // string | bool
	Options []string `json:"options,omitempty"`
	MinLen  int      `json:"min_length,omitempty"`
	MaxLen  int      `json:"max_length,omitempty"`
}

const (
	paramTypeString = "string"
	paramTypeBool   = "bool"
)

// validateParams checks params against specs and returns them as strings.
func validateParams(specs []ParamSpec, params map[string]any) (map[string]string, error) {
	normalized := make(map[string]string, len(params))
	for name, raw := range params {
		index := slices.IndexFunc(specs, func(spec ParamSpec) bool { return spec.Name == name })
		if index < 0 {
			return nil, fmt.Errorf("param %q is not supported", name)
		}

		spec := specs[index]
		switch spec.Type {
		case paramTypeBool:
			switch value := raw.(type) {
			case bool:
				normalized[name] = strconv.FormatBool(value)
			case string:
				parsed, err := strconv.ParseBool(value)
				if err != nil {
					return nil, fmt.Errorf("param %q must be a boolean", name)
				}
				normalized[name] = strconv.FormatBool(parsed)
			default:
				return nil, fmt.Errorf("param %q must be a boolean", name)
			}
		default:
			value, ok := raw.(string)
			if !ok {
				return nil, fmt.Errorf("param %q must be a string", name)
			}

			value = strings.TrimSpace(value)
			if spec.MinLen > 0 && len(value) < spec.MinLen || spec.MaxLen > 0 && len(value) > spec.MaxLen {
				return nil, fmt.Errorf("param %q must be %d-%d characters", name, spec.MinLen, spec.MaxLen)
			}
			if len(spec.Options) > 0 && !slices.Contains(spec.Options, value) {
				return nil, fmt.Errorf("param %q must be one of %s", name, strings.Join(spec.Options, ", "))
			}
			normalized[name] = value
		}
	}

	return normalized, nil
}

// acceptsFile checks the stored (server-sniffed) mime, falling back to the extension of the
// original filename when sniffing could not tell the type.
func acceptsFile(processor Processor, storage models.Storages) bool {
	mimeType, _, err := mime.ParseMediaType(storage.Mime)
	if err != nil {
		mimeType = strings.ToLower(strings.TrimSpace(storage.Mime))
	}

	if slices.Contains(processor.AcceptedMimes(), mimeType) {
		return true
	}

	if mimeType != "" && mimeType != "application/octet-stream" {
		return false
	}

	extension := strings.ToLower(filepath.Ext(storage.OriginalFilename))
	return extension != "" && slices.Contains(processor.AcceptedExtensions(), extension)
}

// uploader is the part of the third-party client used by processors (faked in tests)
type uploader interface {
	Upload(ctx context.Context, input fileprocessor.UploadRequest) (*fileprocessor.UploadResponse, error)
}

// uploadProcessor sends the file and its params to a third-party endpoint and stores every
// top-level field of the JSON object it answers with as output meta.
type uploadProcessor struct {
	action     string
	outputType string
	endpoint   string
	mimes      []string
	extensions []string
	maxBytes   int64
	params     []ParamSpec
	client     uploader
}

func (p uploadProcessor) Action() string               { return p.action }
func (p uploadProcessor) OutputType() string           { return p.outputType }
func (p uploadProcessor) AcceptedMimes() []string      { return p.mimes }
func (p uploadProcessor) AcceptedExtensions() []string { return p.extensions }
func (p uploadProcessor) MaxFileBytes() int64          { return p.maxBytes }
func (p uploadProcessor) Params() []ParamSpec          { return p.params }

func (p uploadProcessor) ValidateParams(params map[string]any) (map[string]string, error) {
	return validateParams(p.params, params)
}

const maxInternalErrorLength = 2000

func (p uploadProcessor) Process(ctx context.Context, job Job) (*Result, error) {
	resp, err := p.client.Upload(ctx, fileprocessor.UploadRequest{
		Endpoint: p.endpoint,
		FileName: job.Storage.OriginalFilename,
		File:     job.File,
		Fields:   job.Params,
	})

	result := &Result{Request: map[string]string{"endpoint": p.endpoint}}
	if resp != nil {
		result.Request["status_code"] = strconv.Itoa(resp.StatusCode)
		result.Request["latency_ms"] = strconv.FormatInt(resp.Latency.Milliseconds(), 10)
	}

	var apiErr *fileprocessor.Error
	var netErr net.Error
	switch {
	case errors.As(err, &apiErr):
		return result, &ProcessError{
			Code:     "third_party_error",
			Message:  "the processing service could not process this file",
			Internal: truncate(fmt.Sprintf("HTTP %d: %s", apiErr.StatusCode, apiErr.Body), maxInternalErrorLength),
		}
	case errors.Is(err, context.DeadlineExceeded) || errors.As(err, &netErr) && netErr.Timeout():
		return result, &ProcessError{Code: "third_party_timeout", Message: "the processing service took too long", Internal: err.Error()}
	case err != nil:
		return result, &ProcessError{Code: "third_party_unavailable", Message: "the processing service is not reachable", Internal: err.Error()}
	}

	output, err := flattenJSONObject(resp.Body)
	if err != nil {
		return result, &ProcessError{
			Code:     "invalid_response",
			Message:  "the processing service returned an unexpected response",
			Internal: truncate(err.Error()+": "+string(resp.Body), maxInternalErrorLength),
		}
	}

	result.Output = output
	return result, nil
}

// flattenJSONObject turns {"a":"x","b":1,"c":{...}} into {"a":"x","b":"1","c":"{...}"}:
// strings are kept as-is, null becomes "", anything else keeps its JSON encoding.
func flattenJSONObject(body []byte) (map[string]string, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil || fields == nil {
		return nil, errors.New("response is not a JSON object")
	}

	output := make(map[string]string, len(fields))
	for key, raw := range fields {
		var text string
		if err := json.Unmarshal(raw, &text); err == nil {
			output[key] = text
			continue
		}

		var compacted bytes.Buffer
		if err := json.Compact(&compacted, raw); err != nil {
			return nil, err
		}
		output[key] = compacted.String()
	}

	return output, nil
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}

// Registry maps an action name to its processor.
type Registry struct {
	processors map[string]Processor
	order      []string
}

func NewRegistry(processors ...Processor) *Registry {
	registry := &Registry{processors: map[string]Processor{}}
	for _, processor := range processors {
		if _, exists := registry.processors[processor.Action()]; !exists {
			registry.order = append(registry.order, processor.Action())
		}
		registry.processors[processor.Action()] = processor
	}
	return registry
}

func (r *Registry) Get(action string) (Processor, bool) {
	processor, ok := r.processors[action]
	return processor, ok
}

// List returns the processors in registration order.
func (r *Registry) List() []Processor {
	processors := make([]Processor, 0, len(r.order))
	for _, action := range r.order {
		processors = append(processors, r.processors[action])
	}
	return processors
}

// Example actions, see docs/prd/file-managements.md §5.5
const (
	ActionPdfSummary         = "pdf_summary"
	ActionAudioTranscription = "audio_transcription"
)

var languageParam = ParamSpec{Name: "language", Type: paramTypeString, MinLen: 2, MaxLen: 10}

// RegistryFromEnv registers the example actions on the third-party client template. It is empty
// when FILE_PROCESSOR_BASE_URL is not set, so creating a job answers "action not available".
func RegistryFromEnv() *Registry {
	client, err := fileprocessor.New(fileprocessor.ConfigFromEnv())
	if err != nil {
		log.Printf("file managements: no actions available, %s", err.Error())
		return NewRegistry()
	}

	return NewRegistry(
		uploadProcessor{
			action:     ActionPdfSummary,
			outputType: "summary",
			endpoint:   envString("FILE_MANAGEMENT_PDF_SUMMARY_ENDPOINT", "/pdf/summary"),
			mimes:      []string{"application/pdf"},
			extensions: []string{".pdf"},
			maxBytes:   int64(envInt("FILE_MANAGEMENT_PDF_SUMMARY_MAX_MB", 8)) << 20,
			params: []ParamSpec{
				languageParam,
				{Name: "length", Type: paramTypeString, Options: []string{"short", "medium", "long"}},
			},
			client: client,
		},
		uploadProcessor{
			action:     ActionAudioTranscription,
			outputType: "transcription",
			endpoint:   envString("FILE_MANAGEMENT_AUDIO_TRANSCRIPTION_ENDPOINT", "/audio/transcriptions"),
			mimes: []string{
				"audio/mpeg", "audio/wav", "audio/wave", "audio/x-wav", "audio/mp4",
				"audio/ogg", "audio/webm", "audio/flac", "audio/aac",
			},
			extensions: []string{".mp3", ".wav", ".m4a", ".ogg", ".webm", ".flac", ".aac"},
			maxBytes:   int64(envInt("FILE_MANAGEMENT_AUDIO_TRANSCRIPTION_MAX_MB", 8)) << 20,
			params: []ParamSpec{
				languageParam,
				{Name: "timestamps", Type: paramTypeBool},
			},
			client: client,
		},
	)
}

func envString(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func envInt(name string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(name)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func envDuration(name string, fallback time.Duration) time.Duration {
	seconds := envInt(name, 0)
	if seconds == 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}
