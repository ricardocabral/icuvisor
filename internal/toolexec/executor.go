// Package toolexec implements transport-neutral tool registration policy and execution.
package toolexec

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/toolcatalog"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

const (
	// GenericErrorMessage is returned when a handler failure has no safe public detail.
	GenericErrorMessage = "tool failed; try again or check icuvisor logs"
	// InvalidInputErrorMessage is returned for unspecialized argument validation failures.
	InvalidInputErrorMessage = "invalid tool arguments; check the inputs and try again"
	// LocalModeAthleteTargetMessage explains why local calls cannot select another athlete.
	LocalModeAthleteTargetMessage = "athlete_id is only supported when coach mode is enabled"
)

// Outcome contains a handler result and, on failure, both internal and public error forms.
type Outcome struct {
	Result        tools.Result
	Err           error
	PublicMessage string
	Panicked      bool
}

// Execute invokes a tool with local-athlete routing and fail-safe panic recovery.
// An empty localAthleteID means routing has already been resolved by the caller.
func Execute(ctx context.Context, tool tools.Tool, request tools.Request, localAthleteID string) (out Outcome) {
	defer func() {
		if recover() != nil {
			out = Outcome{
				Err:           errors.New("tool handler panicked"),
				PublicMessage: GenericErrorMessage,
				Panicked:      true,
			}
		}
	}()

	if localAthleteID != "" && toolcatalog.IsAthleteScopedTool(tool.Name) {
		if hasArgument(request.Arguments, "athlete_id") {
			err := tools.NewUserError(LocalModeAthleteTargetMessage, nil)
			return Outcome{Err: err, PublicMessage: PublicErrorMessage(err)}
		}
		ctx = intervals.WithTargetAthleteID(ctx, localAthleteID)
	}
	result, err := tool.Handler(ctx, request)
	if err != nil {
		return Outcome{Err: err, PublicMessage: PublicErrorMessage(err)}
	}
	return Outcome{Result: result}
}

// PublicErrorMessage returns the safe, actionable representation of a handler error.
func PublicErrorMessage(err error) string {
	if message, ok := tools.PublicErrorMessage(err); ok {
		return message
	}
	if errors.Is(err, tools.ErrInvalidInput) {
		return InvalidInputErrorMessage
	}
	return GenericErrorMessage
}

// Allows reports whether a tool passes the active capability and toolset gates.
func Allows(capability safety.Capability, active safety.Toolset, tool tools.Tool) bool {
	if capability == nil {
		capability = safety.NewCapability(safety.ModeSafe)
	}
	if tool.RequiresDelete() && !capability.CanDelete() {
		return false
	}
	if tool.RequiresWrite() && !capability.CanWrite() {
		return false
	}
	switch safety.ParseToolset(active.String()) {
	case safety.ToolsetFull:
		return true
	case safety.ToolsetCompact:
		return toolcatalog.IsCompactTool(tool.Name)
	default:
		return tool.EffectiveToolset() == safety.ToolsetCore
	}
}

func hasArgument(arguments json.RawMessage, name string) bool {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(arguments, &object); err != nil {
		return false
	}
	_, ok := object[name]
	return ok
}

// PublicError wraps a sanitized execution failure for non-MCP callers.
func PublicError(toolName, message string) error {
	return fmt.Errorf("calling %s: %s", toolName, message)
}
