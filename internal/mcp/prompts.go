package mcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ricardocabral/icuvisor/internal/prompts"
)

const genericPromptErrorMessage = "prompt failed; try again or check icuvisor logs"

type safePromptRegistrar struct {
	server          *sdkmcp.Server
	logger          *slog.Logger
	names           map[string]struct{}
	registeredCount int
}

func (r *safePromptRegistrar) AddPrompt(prompt prompts.Prompt) (err error) {
	if err := r.validatePrompt(prompt); err != nil {
		return err
	}
	r.names[prompt.Name] = struct{}{}
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("registering prompt %q: %v", prompt.Name, recovered)
		}
	}()

	r.server.AddPrompt(&sdkmcp.Prompt{
		Name:        prompt.Name,
		Title:       prompt.Title,
		Description: prompt.Description,
		Arguments:   convertPromptArguments(prompt.Arguments),
	}, func(ctx context.Context, req *sdkmcp.GetPromptRequest) (*sdkmcp.GetPromptResult, error) {
		result, err := prompt.Handler(ctx, prompts.Request{
			Name:      req.Params.Name,
			Arguments: req.Params.Arguments,
		})
		if err != nil {
			r.logger.Error("prompt handler failed", "prompt", prompt.Name, "error", err)
			return nil, errors.New(publicPromptErrorMessage(err))
		}
		return convertPromptResult(result), nil
	})
	r.registeredCount++
	return nil
}

func (r *safePromptRegistrar) validatePrompt(prompt prompts.Prompt) error {
	if !snakeCaseToolName.MatchString(prompt.Name) {
		return fmt.Errorf("invalid prompt name %q; use snake_case", prompt.Name)
	}
	if _, exists := r.names[prompt.Name]; exists {
		return fmt.Errorf("duplicate prompt name %q", prompt.Name)
	}
	if prompt.Title == "" {
		return fmt.Errorf("prompt %q is missing a title", prompt.Name)
	}
	if prompt.Description == "" {
		return fmt.Errorf("prompt %q is missing a description", prompt.Name)
	}
	for _, arg := range prompt.Arguments {
		if !snakeCaseToolName.MatchString(arg.Name) {
			return fmt.Errorf("prompt %q has invalid argument name %q; use snake_case", prompt.Name, arg.Name)
		}
		if arg.Description == "" {
			return fmt.Errorf("prompt %q argument %q is missing a description", prompt.Name, arg.Name)
		}
	}
	if prompt.Handler == nil {
		return fmt.Errorf("prompt %q is missing a handler", prompt.Name)
	}
	return nil
}

func convertPromptArguments(args []prompts.Argument) []*sdkmcp.PromptArgument {
	if len(args) == 0 {
		return nil
	}
	converted := make([]*sdkmcp.PromptArgument, 0, len(args))
	for _, arg := range args {
		converted = append(converted, &sdkmcp.PromptArgument{
			Name:        arg.Name,
			Title:       arg.Title,
			Description: arg.Description,
			Required:    arg.Required,
		})
	}
	return converted
}

func convertPromptResult(result prompts.Result) *sdkmcp.GetPromptResult {
	messages := make([]*sdkmcp.PromptMessage, 0, len(result.Messages))
	for _, message := range result.Messages {
		messages = append(messages, &sdkmcp.PromptMessage{
			Role:    sdkmcp.Role(message.Role),
			Content: &sdkmcp.TextContent{Text: message.Text},
		})
	}
	return &sdkmcp.GetPromptResult{
		Description: result.Description,
		Messages:    messages,
	}
}

func publicPromptErrorMessage(err error) string {
	if message, ok := prompts.PublicErrorMessage(err); ok {
		return message
	}
	return genericPromptErrorMessage
}
