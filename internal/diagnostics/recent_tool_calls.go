package diagnostics

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const defaultRecentToolCallLimit = 50

// RecentToolCall records a redacted MCP tool invocation for diagnostics output.
type RecentToolCall struct {
	Timestamp time.Time `json:"timestamp_utc"`
	Name      string    `json:"name"`
}

// RecentToolCallRecorder stores redacted tool-call metadata.
type RecentToolCallRecorder interface {
	RecordToolCall(context.Context, string, time.Time) error
}

// RecentToolCallStore persists redacted recent tool-call metadata.
type RecentToolCallStore struct {
	path  string
	limit int
	mu    sync.Mutex
}

// DefaultRecentToolCallsPath returns the per-user diagnostics state file path.
func DefaultRecentToolCallsPath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("locating user cache directory: %w", err)
	}
	return filepath.Join(dir, "icuvisor", "recent-tool-calls.jsonl"), nil
}

// NewRecentToolCallStore constructs a redacted recent-tool-call store.
func NewRecentToolCallStore(path string) *RecentToolCallStore {
	return &RecentToolCallStore{path: strings.TrimSpace(path), limit: defaultRecentToolCallLimit}
}

// RecordToolCall appends one redacted tool-call record.
func (s *RecentToolCallStore) RecordToolCall(ctx context.Context, name string, ts time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || strings.TrimSpace(s.path) == "" {
		return nil
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	ts = ts.UTC()
	if ts.IsZero() {
		ts = time.Now().UTC()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	records, err := readRecentToolCallsFile(s.path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	records = append(records, RecentToolCall{Timestamp: ts, Name: name})
	limit := s.limit
	if limit <= 0 {
		limit = defaultRecentToolCallLimit
	}
	if len(records) > limit {
		records = records[len(records)-limit:]
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("creating diagnostics state directory: %w", err)
	}
	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("opening diagnostics state file: %w", err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	for _, record := range records {
		if err := encoder.Encode(record); err != nil {
			return fmt.Errorf("writing diagnostics state file: %w", err)
		}
	}
	return nil
}

// ReadRecentToolCalls returns the most recent redacted tool-call records.
func ReadRecentToolCalls(ctx context.Context, path string, n int) ([]RecentToolCall, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, nil
	}
	records, err := readRecentToolCallsFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if n <= 0 || n > len(records) {
		n = len(records)
	}
	return append([]RecentToolCall(nil), records[len(records)-n:]...), nil
}

func readRecentToolCallsFile(path string) ([]RecentToolCall, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []RecentToolCall
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var record RecentToolCall
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, fmt.Errorf("reading diagnostics state file: %w", err)
		}
		if strings.TrimSpace(record.Name) == "" || record.Timestamp.IsZero() {
			continue
		}
		record.Timestamp = record.Timestamp.UTC()
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading diagnostics state file: %w", err)
	}
	return records, nil
}
