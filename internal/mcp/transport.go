package mcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// StreamableHTTPPath is the local HTTP path serving the MCP Streamable HTTP endpoint.
const StreamableHTTPPath = "/mcp"

const streamableHTTPSessionTimeout = 30 * time.Minute
const streamableHTTPShutdownTimeout = 5 * time.Second

type streamableRequestServerKey struct{}

// StreamableHTTPHandlerOptions configures a reusable SDK Streamable HTTP handler.
type StreamableHTTPHandlerOptions struct {
	Logger              *slog.Logger
	Stateless           bool
	JSONResponse        bool
	FactoryErrorMessage string
}

// StreamableHTTPServerFactory builds or resolves an MCP server for one HTTP request.
type StreamableHTTPServerFactory func(*http.Request) (*Server, error)

// NewStreamableHTTPHandler adapts icuvisor Server construction to the SDK Streamable HTTP handler.
func NewStreamableHTTPHandler(factory StreamableHTTPServerFactory, opts StreamableHTTPHandlerOptions) http.Handler {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	sdkHandler := sdkmcp.NewStreamableHTTPHandler(func(req *http.Request) *sdkmcp.Server {
		if server, ok := req.Context().Value(streamableRequestServerKey{}).(*Server); ok && server != nil {
			return server.server
		}
		server, err := factory(req)
		if err != nil || server == nil {
			return nil
		}
		return server.server
	}, &sdkmcp.StreamableHTTPOptions{
		Stateless:                  opts.Stateless,
		JSONResponse:               opts.JSONResponse,
		Logger:                     logger,
		SessionTimeout:             streamableHTTPSessionTimeout,
		DisableLocalhostProtection: false,
		CrossOriginProtection:      nil,
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodGet {
			server, err := factory(r)
			if err != nil {
				message := err.Error()
				if opts.FactoryErrorMessage != "" {
					message = opts.FactoryErrorMessage
				}
				http.Error(w, message, http.StatusUnauthorized)
				return
			}
			if server == nil {
				http.Error(w, "no server available", http.StatusBadRequest)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), streamableRequestServerKey{}, server))
		}
		sdkHandler.ServeHTTP(w, r)
	})
}

// Run serves one MCP session until the client disconnects or ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	if s == nil || s.server == nil || s.transport == nil {
		return errors.New("mcp server is not initialized")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	logger := s.logger
	if logger == nil {
		logger = slog.Default()
	}
	version := s.version
	if version == "" {
		version = "dev"
	}
	serverSession, err := s.server.Connect(ctx, s.transport, nil)
	if err != nil {
		logger.Error("server startup failed", "version", version, "transport", transportName(s.transport), "error", err)
		return err
	}
	logger.Info("server started listening", "version", version, "transport", transportName(s.transport))

	closed := make(chan error, 1)
	go func() {
		defer close(closed)
		closed <- serverSession.Wait()
	}()

	err, cancelled := waitForSessionClose(ctx, closed, func() { serverSession.Close() })
	if cancelled {
		logger.Error("server run cancelled", "version", version, "transport", transportName(s.transport), "error", err)
		return err
	}
	if err != nil {
		logger.Error("server session ended with error", "version", version, "transport", transportName(s.transport), "error", err)
	} else {
		logger.Info("server session ended", "version", version, "transport", transportName(s.transport))
	}
	return err
}

func waitForSessionClose(ctx context.Context, closed <-chan error, closeSession func()) (error, bool) {
	select {
	case <-ctx.Done():
		closeSession()
		<-closed
		return ctx.Err(), true
	case err := <-closed:
		return err, false
	}
}

// RunStreamableHTTP serves the shared MCP server over Streamable HTTP.
func (s *Server) RunStreamableHTTP(ctx context.Context, address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listening for streamable HTTP on %s: %w", address, err)
	}
	return s.ServeStreamableHTTP(ctx, listener)
}

// ServeStreamableHTTP serves Streamable HTTP on listener until ctx is cancelled.
func (s *Server) ServeStreamableHTTP(ctx context.Context, listener net.Listener) error {
	if s == nil || s.server == nil {
		return errors.New("mcp server is not initialized")
	}
	if listener == nil {
		return errors.New("streamable HTTP listener is nil")
	}
	if err := ctx.Err(); err != nil {
		_ = listener.Close()
		return err
	}
	logger := s.logger
	if logger == nil {
		logger = slog.Default()
	}
	version := s.version
	if version == "" {
		version = "dev"
	}

	streamableHandler := NewStreamableHTTPHandler(func(*http.Request) (*Server, error) {
		return s, nil
	}, StreamableHTTPHandlerOptions{Logger: logger})
	mux := http.NewServeMux()
	mux.Handle(StreamableHTTPPath, streamableHandler)
	httpServer := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	serveDone := make(chan error, 1)
	go func() {
		serveDone <- httpServer.Serve(listener)
	}()
	logger.Info("server started listening", "version", version, "transport", "streamable_http", "address", listener.Addr().String(), "path", StreamableHTTPPath)

	select {
	case err := <-serveDone:
		return normalizeHTTPServerError(err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), streamableHTTPShutdownTimeout)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("streamable HTTP shutdown failed; closing listener", "version", version, "error", err)
			_ = httpServer.Close()
		}
		if err := normalizeHTTPServerError(<-serveDone); err != nil {
			return err
		}
		logger.Info("server session ended", "version", version, "transport", "streamable_http")
		return ctx.Err()
	}
}

func normalizeHTTPServerError(err error) error {
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func transportName(transport sdkmcp.Transport) string {
	switch transport.(type) {
	case *sdkmcp.StdioTransport:
		return "stdio"
	case *sdkmcp.IOTransport:
		return "io"
	case *sdkmcp.InMemoryTransport:
		return "in_memory"
	case *sdkmcp.StreamableServerTransport:
		return "streamable_http"
	default:
		return fmt.Sprintf("%T", transport)
	}
}
