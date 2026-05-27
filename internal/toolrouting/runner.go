package toolrouting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	icumcp "github.com/ricardocabral/icuvisor/internal/mcp"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

const (
	EnvRoutingEvalProvider = "ICUVISOR_ROUTING_EVAL_PROVIDER"
	EnvRoutingEvalModel    = "ICUVISOR_ROUTING_EVAL_MODEL"
	EnvAnthropicAPIKey     = "ANTHROPIC_API_KEY" // #nosec G101 -- environment variable name, not a credential value.
	EnvAnthropicURL        = "ICUVISOR_ROUTING_EVAL_ANTHROPIC_URL"
	DefaultAnthropicURL    = "https://api.anthropic.com/v1/messages"
	DefaultAnthropicModel  = "claude-sonnet-4-20250514"
)

// ToolDefinition is the provider-facing MCP tool description used by the smoke eval.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// Provider chooses the first tool it would call for a prompt and catalog.
type Provider interface {
	FirstTool(context.Context, PromptRequest) (Selection, error)
}

// PromptRequest is one provider first-tool-choice request.
type PromptRequest struct {
	Case  Case
	Tools []ToolDefinition
}

// Selection is the provider's first tool call, or no-tool text.
type Selection struct {
	ToolName   string
	RawMessage string
}

// EvalSummary summarizes a routing smoke-eval run.
type EvalSummary struct {
	Total   int      `json:"total"`
	Passed  int      `json:"passed"`
	Failed  int      `json:"failed"`
	Skipped int      `json:"skipped"`
	Results []Result `json:"results"`
}

// EnvProvider returns a provider only when explicit opt-in environment is configured.
func EnvProvider(getenv func(string) string, client *http.Client) (Provider, string, bool, error) {
	providerName := strings.ToLower(strings.TrimSpace(getenv(EnvRoutingEvalProvider)))
	if providerName == "" {
		return nil, "", false, nil
	}
	if providerName != "anthropic" {
		return nil, providerName, true, fmt.Errorf("unsupported %s=%q (supported: anthropic)", EnvRoutingEvalProvider, providerName)
	}
	apiKey := strings.TrimSpace(getenv(EnvAnthropicAPIKey))
	if apiKey == "" {
		return nil, providerName, true, fmt.Errorf("%s=anthropic requires %s", EnvRoutingEvalProvider, EnvAnthropicAPIKey)
	}
	model := strings.TrimSpace(getenv(EnvRoutingEvalModel))
	if model == "" {
		model = DefaultAnthropicModel
	}
	endpoint := strings.TrimSpace(getenv(EnvAnthropicURL))
	if endpoint == "" {
		endpoint = DefaultAnthropicURL
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return AnthropicProvider{APIKey: apiKey, Model: model, Endpoint: endpoint, Client: client}, providerName, true, nil
}

// Run evaluates all fixture cases. A nil provider validates cases/catalogs and reports skips without network access.
func Run(ctx context.Context, fixture Fixture, provider Provider, out io.Writer) (EvalSummary, error) {
	catalogs := make(map[string][]ToolDefinition)
	loadCatalog := func(c Case) ([]ToolDefinition, error) {
		if tools, ok := catalogs[c.CatalogMode]; ok {
			return tools, nil
		}
		defs, err := LoadToolDefinitions(ctx, safety.ParseToolset(c.Toolset), safety.ParseMode(c.DeleteMode))
		if err != nil {
			return nil, err
		}
		catalogs[c.CatalogMode] = defs
		return defs, nil
	}

	summary := EvalSummary{Total: len(fixture.Cases)}
	for _, c := range fixture.Cases {
		defs, err := loadCatalog(c)
		if err != nil {
			return summary, fmt.Errorf("loading catalog for %s: %w", c.ID, err)
		}
		if !expectedToolAvailable(c, defs) {
			return summary, fmt.Errorf("case %s expects %s, which is not exposed in %s", c.ID, expectedName(c), c.CatalogMode)
		}
		if provider == nil {
			summary.Skipped++
			result := CompareResult(c, "", "provider not configured")
			result.Pass = false
			result.Detail = "skipped: set ICUVISOR_ROUTING_EVAL_PROVIDER=anthropic and ANTHROPIC_API_KEY to run provider-backed routing"
			summary.Results = append(summary.Results, result)
			fmt.Fprintf(out, "SKIP %s provider not configured (%d tools in %s)\n", c.ID, len(defs), c.CatalogMode)
			continue
		}
		selection, err := provider.FirstTool(ctx, PromptRequest{Case: c, Tools: defs})
		if err != nil {
			return summary, fmt.Errorf("case %s provider call: %w", c.ID, err)
		}
		result := CompareResult(c, selection.ToolName, selection.RawMessage)
		if result.Pass {
			summary.Passed++
			fmt.Fprintf(out, "PASS %s expected=%s actual=%s\n", c.ID, result.Expected, result.Actual)
		} else {
			summary.Failed++
			fmt.Fprintf(out, "FAIL %s %s\n", c.ID, result.Detail)
		}
		summary.Results = append(summary.Results, result)
	}
	fmt.Fprintf(out, "summary: total=%d passed=%d failed=%d skipped=%d\n", summary.Total, summary.Passed, summary.Failed, summary.Skipped)
	return summary, nil
}

// LoadToolDefinitions loads registered tool definitions for a catalog mode without executing handlers.
func LoadToolDefinitions(ctx context.Context, toolset safety.Toolset, mode safety.Mode) ([]ToolDefinition, error) {
	client, err := intervals.NewClient(intervals.Options{Config: config.Config{
		APIKey:      strings.Repeat("x", 8),
		AthleteID:   "i12345",
		APIBaseURL:  "https://example.invalid",
		HTTPTimeout: time.Second,
	}, Version: "routing-eval"})
	if err != nil {
		return nil, fmt.Errorf("creating registration-only intervals client: %w", err)
	}
	cfg := config.Config{AthleteID: "i12345", Timezone: "UTC", DeleteMode: mode, Toolset: toolset}
	registry := tools.NewRegistryWithOptions(client, tools.RegistryOptions{Version: "routing-eval", TimezoneFallback: "UTC", Capability: safety.NewCapability(mode), Toolset: toolset})
	catalog, err := icumcp.CollectToolCatalog(ctx, icumcp.CatalogHashOptions{Config: cfg, Registry: registry, Capability: safety.NewCapability(mode), Toolset: toolset})
	if err != nil {
		return nil, err
	}
	defs := make([]ToolDefinition, 0, len(catalog))
	for _, tool := range catalog {
		inputSchema, err := json.Marshal(tool.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("marshalling input schema for %s: %w", tool.Name, err)
		}
		defs = append(defs, ToolDefinition{Name: tool.Name, Description: tool.Description, InputSchema: inputSchema})
	}
	sort.Slice(defs, func(i, j int) bool { return defs[i].Name < defs[j].Name })
	return defs, nil
}

// AnthropicProvider calls Anthropic Messages with tool definitions but never executes tool handlers.
type AnthropicProvider struct {
	APIKey   string
	Model    string
	Endpoint string
	Client   *http.Client
}

func (p AnthropicProvider) FirstTool(ctx context.Context, req PromptRequest) (Selection, error) {
	if strings.TrimSpace(p.APIKey) == "" {
		return Selection{}, errors.New("missing Anthropic API key")
	}
	payload := anthropicRequest{
		Model:       p.Model,
		MaxTokens:   128,
		Temperature: 0,
		System:      "You are evaluating MCP tool routing. Pick the single first icuvisor tool call you would make for the user prompt. If no exposed tool should be called, answer briefly without using a tool. Do not invent unavailable tools.",
		Messages:    []anthropicMessage{{Role: "user", Content: req.Case.Prompt}},
		Tools:       make([]anthropicTool, 0, len(req.Tools)),
	}
	for _, tool := range req.Tools {
		payload.Tools = append(payload.Tools, anthropicTool(tool))
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Selection{}, fmt.Errorf("marshalling Anthropic request: %w", err)
	}
	endpoint := p.Endpoint
	if endpoint == "" {
		endpoint = DefaultAnthropicURL
	}
	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return Selection{}, fmt.Errorf("building Anthropic request: %w", err)
	}
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("x-api-key", p.APIKey)
	resp, err := client.Do(httpReq)
	if err != nil {
		return Selection{}, fmt.Errorf("calling Anthropic: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Selection{}, fmt.Errorf("reading Anthropic response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Selection{}, fmt.Errorf("anthropic returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	var parsed anthropicResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return Selection{}, fmt.Errorf("decoding Anthropic response: %w", err)
	}
	for _, block := range parsed.Content {
		if block.Type == "tool_use" {
			return Selection{ToolName: block.Name, RawMessage: string(respBody)}, nil
		}
	}
	return Selection{RawMessage: string(respBody)}, nil
}

func LoadFixtureFile(path string) (Fixture, error) {
	file, err := os.Open(path)
	if err != nil {
		return Fixture{}, err
	}
	defer file.Close()
	return LoadFixture(file, knownToolsFromCatalog())
}

func knownToolsFromCatalog() map[string]struct{} {
	defs := tools.Catalog()
	out := make(map[string]struct{}, len(defs))
	for _, def := range defs {
		out[def.Name] = struct{}{}
	}
	return out
}

func expectedToolAvailable(c Case, defs []ToolDefinition) bool {
	if c.ExpectedFirstTool == nil {
		return true
	}
	for _, def := range defs {
		if def.Name == *c.ExpectedFirstTool {
			return true
		}
	}
	return false
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature"`
	System      string             `json:"system"`
	Messages    []anthropicMessage `json:"messages"`
	Tools       []anthropicTool    `json:"tools"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Name string `json:"name,omitempty"`
		Text string `json:"text,omitempty"`
	} `json:"content"`
}
