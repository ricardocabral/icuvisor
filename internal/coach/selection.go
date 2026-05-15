package coach

import (
	"context"
	"sync"
)

const processSelectionKey = "process"

type selectionContextKey struct{}

// SelectionStore tracks the active coach athlete per MCP session.
type SelectionStore struct {
	mu        sync.RWMutex
	defaultID string
	selected  map[string]string
}

// NewSelectionStore creates a session-safe selection store.
func NewSelectionStore(defaultAthleteID string) *SelectionStore {
	return &SelectionStore{defaultID: defaultAthleteID, selected: make(map[string]string)}
}

// Key returns a stable selection key and scope for an SDK session ID.
func (s *SelectionStore) Key(sessionID string) (string, string) {
	if sessionID == "" {
		return processSelectionKey, "process"
	}
	return sessionID, "session"
}

// Selected returns the active athlete for key, falling back to the default.
func (s *SelectionStore) Selected(key string) string {
	if s == nil {
		return ""
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if athleteID := s.selected[key]; athleteID != "" {
		return athleteID
	}
	return s.defaultID
}

// Select updates key's active athlete and returns the previous value.
func (s *SelectionStore) Select(key string, athleteID string) string {
	if s == nil {
		return ""
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	previous := s.selected[key]
	if previous == "" {
		previous = s.defaultID
	}
	s.selected[key] = athleteID
	return previous
}

// SelectionContext exposes session selection state to SDK-agnostic tools.
type SelectionContext struct {
	Store        *SelectionStore
	Key          string
	Scope        string
	VisibleTools func(athleteID string) []string
}

// WithSelectionContext stores selection context for a tool call.
func WithSelectionContext(ctx context.Context, selection SelectionContext) context.Context {
	return context.WithValue(ctx, selectionContextKey{}, selection)
}

// SelectionContextFromContext returns session selection state from ctx.
func SelectionContextFromContext(ctx context.Context) (SelectionContext, bool) {
	selection, ok := ctx.Value(selectionContextKey{}).(SelectionContext)
	return selection, ok
}
