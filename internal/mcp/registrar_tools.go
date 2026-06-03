package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/diagnostics"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/toolcatalog"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

const genericToolErrorMessage = "tool failed; try again or check icuvisor logs"

const invalidInputToolErrorMessage = "invalid tool arguments; check the inputs and try again"

const invalidTargetAthleteFormatMessage = "invalid athlete_id; use format i12345 or 12345"

const unauthorizedTargetAthleteMessage = "athlete_id is not authorized for this icuvisor coach roster"

const toolNotAllowedForAthleteMessage = "tool is not allowed for the selected athlete"

const localModeAthleteTargetMessage = "athlete_id is only supported when coach mode is enabled"

func capabilityOrSafe(capability safety.Capability) safety.Capability {
	if capability != nil {
		return capability
	}
	return safety.NewCapability(safety.ModeSafe)
}

func toolsetOrCore(opts Options) safety.Toolset {
	toolset := opts.Toolset
	if toolset == "" {
		toolset = opts.Config.Toolset
	}
	return safety.ParseToolset(string(toolset))
}

type safeRegistrar struct {
	server                 *sdkmcp.Server
	logger                 *slog.Logger
	config                 config.Config
	coachFilter            coach.ToolFilter
	selectionStore         *coach.SelectionStore
	capability             safety.Capability
	toolset                safety.Toolset
	names                  map[string]struct{}
	registeredTools        []tools.Tool
	coachVisibleCatalog    []tools.Tool
	registeredCount        int
	skippedToolsetCount    int
	skippedCapabilityCount int
	skippedCoachCount      int
	recentToolCalls        diagnostics.RecentToolCallRecorder
}

func (r *safeRegistrar) AddTool(tool tools.Tool) error {
	tool = r.prepareTool(tool)
	if err := r.validateTool(tool); err != nil {
		return err
	}
	r.names[tool.Name] = struct{}{}
	if tool.Name == toolcatalog.ICUvisorListAdvancedCapabilities && r.config.CoachModeEnabled() {
		tool.Handler = tools.NewFilteredAdvancedCapabilitiesHandler(r.coachVisibleCatalog, r.toolset, func(ctx context.Context, candidate tools.Tool) bool {
			return r.visibleForAthlete(r.selectedAthleteID(ctx), candidate.Name)
		})
	}
	if !r.capabilityAllows(tool) {
		r.skippedCapabilityCount++
		return nil
	}
	if tool.Name != toolcatalog.ICUvisorListAdvancedCapabilities && r.coachAllows(tool) {
		r.coachVisibleCatalog = append(r.coachVisibleCatalog, tool)
	}
	if !r.toolsetAllows(tool) {
		r.skippedToolsetCount++
		return nil
	}
	if !r.coachAllows(tool) {
		r.skippedCoachCount++
		return nil
	}

	return withPanicRecovery(fmt.Sprintf("registering tool %q", tool.Name), func() error {
		if r.server == nil {
			r.registeredTools = append(r.registeredTools, tool)
			r.registeredCount++
			return nil
		}
		r.server.AddTool(&sdkmcp.Tool{
			Name:         tool.Name,
			Description:  tool.Description,
			InputSchema:  tool.InputSchema,
			OutputSchema: tool.OutputSchema,
		}, func(ctx context.Context, req *sdkmcp.CallToolRequest) (res *sdkmcp.CallToolResult, err error) {
			started := time.Now()
			argumentBytes := len(req.Params.Arguments)
			logToolCallStarted(r.logger, tool.Name, argumentBytes)
			status := "ok"
			defer func() {
				logToolCallCompleted(r.logger, tool.Name, status, time.Since(started), argumentBytes, res)
			}()

			if r.recentToolCalls != nil {
				if err := r.recentToolCalls.RecordToolCall(ctx, tool.Name, started.UTC()); err != nil {
					r.logger.Warn("recording diagnostics tool call failed", "tool", tool.Name, "error", err)
				}
			}
			callCtx := r.withSelection(ctx, req.Session)
			callCtx, arguments, err := r.resolveToolTarget(callCtx, tool.Name, req.Params.Arguments)
			if err != nil {
				status = "target_error"
				res = toolErrorResult(publicToolErrorMessage(err))
				return res, nil
			}
			result, err := tool.Handler(callCtx, tools.Request{
				Name:      req.Params.Name,
				Arguments: arguments,
			})
			if err != nil {
				status = "tool_error"
				logToolHandlerError(r.logger, tool.Name, err)
				res = toolErrorResult(publicToolErrorMessage(err))
				return res, nil
			}
			converted, err := convertResult(result)
			if err != nil {
				status = "conversion_error"
				r.logger.Error("tool result conversion failed", "tool", tool.Name, "error", err)
				res = toolErrorResult(genericToolErrorMessage)
				return res, nil
			}
			if converted.IsError {
				status = "tool_error"
			}
			res = converted
			return res, nil
		})
		r.registeredTools = append(r.registeredTools, tool)
		r.registeredCount++
		return nil
	})
}

func (r *safeRegistrar) visibilityMiddleware() sdkmcp.Middleware {
	return func(next sdkmcp.MethodHandler) sdkmcp.MethodHandler {
		return func(ctx context.Context, method string, req sdkmcp.Request) (sdkmcp.Result, error) {
			result, err := next(ctx, method, req)
			if err != nil || method != "tools/list" {
				return result, err
			}
			toolsResult, ok := result.(*sdkmcp.ListToolsResult)
			if !ok {
				return result, err
			}
			selectionCtx := r.withSelection(ctx, req.GetSession().(*sdkmcp.ServerSession))
			athleteID := r.config.Coach.DefaultAthleteID
			if selection, ok := coach.SelectionContextFromContext(selectionCtx); ok && selection.Store != nil {
				athleteID = selection.Store.Selected(selection.Key)
			}
			filtered := toolsResult.Tools[:0]
			for _, tool := range toolsResult.Tools {
				if r.visibleForAthlete(athleteID, tool.Name) {
					filtered = append(filtered, tool)
				}
			}
			toolsResult.Tools = filtered
			return toolsResult, nil
		}
	}
}

func (r *safeRegistrar) visibleToolNamesForAthlete(athleteID string) []string {
	toolNames := make([]string, 0, len(r.registeredTools))
	for _, tool := range r.registeredTools {
		toolNames = append(toolNames, tool.Name)
	}
	return r.coachFilter.VisibleToolNamesForAthlete(athleteID, toolNames)
}

func (r *safeRegistrar) visibleForAthlete(athleteID string, toolName string) bool {
	return r.coachFilter.VisibleForAthlete(athleteID, toolName)
}

func (r *safeRegistrar) prepareTool(tool tools.Tool) tools.Tool {
	if !r.config.CoachModeEnabled() || !toolcatalog.IsAthleteScopedTool(tool.Name) {
		return tool
	}
	tool.InputSchema = SchemaWithAthleteID(tool.InputSchema)
	return tool
}

func (r *safeRegistrar) coachAllows(tool tools.Tool) bool {
	return r.coachFilter.AllowedForAny(tool.Name)
}

func (r *safeRegistrar) withSelection(ctx context.Context, session *sdkmcp.ServerSession) context.Context {
	if r.selectionStore == nil {
		return ctx
	}
	sessionID := ""
	if session != nil {
		sessionID = session.ID()
	}
	key, scope := r.selectionStore.Key(sessionID)
	return coach.WithSelectionContext(ctx, coach.SelectionContext{Store: r.selectionStore, Key: key, Scope: scope, VisibleTools: r.visibleToolNamesForAthlete})
}

func (r *safeRegistrar) resolveToolTarget(ctx context.Context, toolName string, raw json.RawMessage) (context.Context, json.RawMessage, error) {
	if !toolcatalog.IsAthleteScopedTool(toolName) {
		return ctx, raw, nil
	}
	arguments, suppliedAthleteID, suppliedAthleteIDPresent, err := stripAthleteID(raw)
	if err != nil {
		return ctx, nil, tools.NewUserError(invalidTargetAthleteFormatMessage, err)
	}
	if !r.config.CoachModeEnabled() {
		if suppliedAthleteIDPresent {
			return ctx, nil, tools.NewUserError(localModeAthleteTargetMessage, nil)
		}
		return intervals.WithTargetAthleteID(ctx, r.config.AthleteID), raw, nil
	}
	targetAthleteID, err := r.resolveAthleteID(ctx, toolName, suppliedAthleteID)
	if err != nil {
		return ctx, nil, tools.NewUserError(publicAthleteRoutingMessage(err), err)
	}
	return intervals.WithTargetAthleteID(ctx, targetAthleteID), arguments, nil
}

func (r *safeRegistrar) resolveAthleteID(ctx context.Context, toolName string, suppliedAthleteID string) (string, error) {
	return r.coachFilter.ResolveTarget(suppliedAthleteID, r.config.Coach.DefaultAthleteID, r.selectedAthleteID(ctx), toolName, config.NormalizeAthleteID)
}

func publicAthleteRoutingMessage(err error) string {
	switch {
	case errors.Is(err, coach.ErrInvalidAthleteID):
		return invalidTargetAthleteFormatMessage
	case errors.Is(err, coach.ErrToolNotAllowed):
		return toolNotAllowedForAthleteMessage
	case errors.Is(err, coach.ErrAthleteNotAuthorized):
		return unauthorizedTargetAthleteMessage
	default:
		return unauthorizedTargetAthleteMessage
	}
}

func (r *safeRegistrar) selectedAthleteID(ctx context.Context) string {
	athleteID := r.config.Coach.DefaultAthleteID
	if selection, ok := coach.SelectionContextFromContext(ctx); ok && selection.Store != nil {
		athleteID = selection.Store.Selected(selection.Key)
	}
	return athleteID
}

func (r *safeRegistrar) toolsetAllows(tool tools.Tool) bool {
	active := safety.ParseToolset(string(r.toolset))
	if active == safety.ToolsetFull {
		return true
	}
	return tool.EffectiveToolset() == safety.ToolsetCore
}

func (r *safeRegistrar) capabilityAllows(tool tools.Tool) bool {
	capability := r.capability
	if capability == nil {
		capability = safety.NewCapability(safety.ModeSafe)
	}
	if tool.RequiresDelete() && !capability.CanDelete() {
		return false
	}
	return !tool.RequiresWrite() || capability.CanWrite()
}

func (r *safeRegistrar) validateTool(tool tools.Tool) error {
	if !snakeCaseToolName.MatchString(tool.Name) {
		return fmt.Errorf("invalid tool name %q; use snake_case", tool.Name)
	}
	if _, exists := r.names[tool.Name]; exists {
		return fmt.Errorf("duplicate tool name %q", tool.Name)
	}
	if tool.Description == "" {
		return fmt.Errorf("tool %q is missing a description", tool.Name)
	}
	if err := validateToolset(tool); err != nil {
		return err
	}
	if err := validateObjectSchema("input", tool.Name, tool.InputSchema, true); err != nil {
		return err
	}
	if err := validateObjectSchema("output", tool.Name, tool.OutputSchema, false); err != nil {
		return err
	}
	if tool.Handler == nil {
		return fmt.Errorf("tool %q is missing a handler", tool.Name)
	}
	return nil
}

func logToolCallStarted(logger *slog.Logger, toolName string, argumentBytes int) {
	if logger == nil {
		logger = slog.Default()
	}
	logger.Info("tool call started", "tool", toolName, "argument_bytes", argumentBytes, "estimated_argument_tokens", estimateTokenCount(argumentBytes), "token_estimate_method", "bytes_div_4")
}

func logToolCallCompleted(logger *slog.Logger, toolName string, status string, duration time.Duration, argumentBytes int, result *sdkmcp.CallToolResult) {
	if logger == nil {
		logger = slog.Default()
	}
	responseBytes := marshaledJSONSize(result)
	estimatedArgumentTokens := estimateTokenCount(argumentBytes)
	estimatedResponseTokens := estimateTokenCount(responseBytes)
	estimatedMCPTokens := estimatedArgumentTokens + estimatedResponseTokens
	if estimatedResponseTokens < 0 {
		estimatedMCPTokens = -1
	}
	logger.Info("tool call completed", "tool", toolName, "status", status, "duration_ms", duration.Milliseconds(), "argument_bytes", argumentBytes, "response_bytes", responseBytes, "estimated_argument_tokens", estimatedArgumentTokens, "estimated_response_tokens", estimatedResponseTokens, "estimated_mcp_tokens", estimatedMCPTokens, "token_estimate_method", "bytes_div_4")
}

func marshaledJSONSize(value any) int {
	if value == nil {
		return 0
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return -1
	}
	return len(payload)
}

func estimateTokenCount(byteCount int) int {
	if byteCount < 0 {
		return -1
	}
	if byteCount == 0 {
		return 0
	}
	return (byteCount + 3) / 4
}

func logToolHandlerError(logger *slog.Logger, toolName string, err error) {
	if logger == nil {
		logger = slog.Default()
	}
	attrs := []any{"tool", toolName, "error", err}
	var userErr *tools.UserError
	if errors.As(err, &userErr) && userErr.Unwrap() != nil {
		attrs = append(attrs, "cause", userErr.Unwrap())
	}
	if errors.Is(err, tools.ErrInvalidInput) {
		logger.Warn("tool handler rejected invalid input", attrs...)
		return
	}
	logger.Error("tool handler failed", attrs...)
}

func publicToolErrorMessage(err error) string {
	if message, ok := tools.PublicErrorMessage(err); ok {
		return message
	}
	if errors.Is(err, tools.ErrInvalidInput) {
		return invalidInputToolErrorMessage
	}
	return genericToolErrorMessage
}

func toolErrorResult(message string) *sdkmcp.CallToolResult {
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{
			&sdkmcp.TextContent{Text: message},
		},
		IsError: true,
	}
}
