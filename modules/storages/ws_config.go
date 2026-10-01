package storages

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// wsConfig holds the WS upload limits, all read from env (sizes in MB/KB as named).
type wsConfig struct {
	MaxFileBytes       int64         // WS_UPLOAD_MAX_FILE_MB, default 8
	MaxConnectionBytes int64         // WS_UPLOAD_MAX_CONNECTION_MB, default 0 (no cap)
	MaxChunkBytes      int64         // WS_UPLOAD_MAX_CHUNK_KB, default 1024
	IdleTimeout        time.Duration // WS_UPLOAD_IDLE_TIMEOUT_SECONDS, default 120
	StallTimeout       time.Duration // WS_UPLOAD_STALL_SECONDS, default 30
	OriginPatterns     []string      // WS_ALLOWED_ORIGINS, comma separated host patterns; empty = same host only
}

func wsConfigFromEnv() wsConfig {
	cfg := wsConfig{
		MaxFileBytes:       envInt64("WS_UPLOAD_MAX_FILE_MB", 8) * 1024 * 1024,
		MaxConnectionBytes: envInt64("WS_UPLOAD_MAX_CONNECTION_MB", 0) * 1024 * 1024,
		MaxChunkBytes:      envInt64("WS_UPLOAD_MAX_CHUNK_KB", 1024) * 1024,
		IdleTimeout:        time.Duration(envInt64("WS_UPLOAD_IDLE_TIMEOUT_SECONDS", 120)) * time.Second,
		StallTimeout:       time.Duration(envInt64("WS_UPLOAD_STALL_SECONDS", 30)) * time.Second,
	}

	for _, origin := range strings.Split(os.Getenv("WS_ALLOWED_ORIGINS"), ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			cfg.OriginPatterns = append(cfg.OriginPatterns, origin)
		}
	}

	return cfg
}

func envInt64(key string, def int64) int64 {
	value, err := strconv.ParseInt(strings.TrimSpace(os.Getenv(key)), 10, 64)
	if err != nil || value < 0 {
		return def
	}

	if value == 0 && def != 0 {
		return def
	}

	return value
}
