package mcp

import (
	"fmt"
	"log/slog"
	"runtime/debug"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// withPanicRecovery converts SDK protocol-boundary panics into errors so callers can fail safely.
func withPanicRecovery(name string, fn func() error) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			slog.Default().Error("panic recovered", "scope", name, "stack", string(debug.Stack()))
			if recoveredErr, ok := recovered.(error); ok {
				err = fmt.Errorf("%s: %w", name, recoveredErr)
				return
			}
			err = fmt.Errorf("%s: %v", name, recovered)
		}
	}()
	return fn()
}

func newSDKServer(version string, logger *slog.Logger, capabilities *sdkmcp.ServerCapabilities) (server *sdkmcp.Server, err error) {
	err = withPanicRecovery("constructing MCP server", func() error {
		server = sdkmcp.NewServer(&sdkmcp.Implementation{Name: "icuvisor", Version: version}, &sdkmcp.ServerOptions{Logger: logger, Capabilities: capabilities})
		return nil
	})
	return server, err
}
