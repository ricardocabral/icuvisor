package app

import (
	"fmt"
	"io"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

func writeTopLevelHelp(w io.Writer) error {
	_, err := fmt.Fprintf(w, `icuvisor connects intervals.icu training data to MCP-compatible AI clients.

Usage:
  icuvisor [flags]
  icuvisor <command> [flags]

Commands:
  (no command)  Run the MCP server (stdio transport by default).
  version       Print the icuvisor version and exit.
  help          Print this help and exit.

Flags:
  --config <path>        JSON config file path. Can also be set with %[1]s.
  --env-file <path>      Env-file path to load before process env. Can also be set with %[2]s. Default: .env when present.
  --transport <name>     MCP transport: stdio or http. Default: %[3]s.
  --http-bind <addr>     HTTP bind address for --transport http. Default: %[4]s.
  -h, --help             Print this help and exit.

Environment variables:
  %[5]s      intervals.icu API key. Required unless provided by config/keychain.
  %[6]s   Athlete ID, with or without leading i. Required unless provided by config.
  %[1]s            JSON config file path used when --config is omitted.
  %[2]s          Env-file path used when --env-file is omitted.
  %[7]s          Athlete timezone. Default: %[8]s.
  %[9]s      intervals.icu API base URL. Default: %[10]s.
  %[11]s      HTTP client timeout. Default: %[12]s.
  %[13]s         MCP transport: stdio or http. Default: %[3]s.
  %[14]s         HTTP bind address for Streamable HTTP. Default: %[4]s.
  %[15]s       Write/delete registration mode: safe, full, or none. Default: %[16]s.
  %[17]s           Tool catalog tier: core or full. Default: %[18]s.
  %[19]s    Include debug metadata in MCP responses when set to true. Default: false.

Examples:
  icuvisor
  ICUVISOR_TRANSPORT=http icuvisor
  icuvisor --transport http --http-bind 127.0.0.1:8765
  icuvisor --config /path/to/icuvisor.json

Exit codes:
  0  Success, including help and version output.
  2  Usage error, such as an unknown flag or missing flag value.
  1  Runtime error while loading config or running the server.

For deeper documentation, see README.md and docs/prd/PRD-icuvisor.md.
`, config.EnvConfigPath, config.EnvDotEnvPath, config.TransportStdio, config.DefaultHTTPBindAddress,
		config.EnvAPIKey, config.EnvAthleteID, config.EnvTimezone, config.DefaultTimezone,
		config.EnvAPIBaseURL, config.DefaultAPIBaseURL, config.EnvHTTPTimeout, config.DefaultHTTPTimeout,
		config.EnvTransport, config.EnvHTTPBind, safety.EnvDeleteMode, safety.ModeSafe,
		safety.EnvToolset, safety.ToolsetCore, config.EnvDebugMetadata)
	if err != nil {
		return fmt.Errorf("writing help: %w", err)
	}
	return nil
}

func writeVersionHelp(w io.Writer) error {
	_, err := fmt.Fprint(w, `Print the icuvisor version and exit.

Usage:
  icuvisor version [--help]

Exit codes:
  0  Success, including help and version output.
`)
	if err != nil {
		return fmt.Errorf("writing version help: %w", err)
	}
	return nil
}
