package app

import (
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
)

func parseDefaultArgs(args []string) (config.Options, error) {
	var opts config.Options
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--config":
			value, next, err := requireFlagValue(args, i, "--config", "/path/to/icuvisor.json")
			if err != nil {
				return config.Options{}, err
			}
			opts.Path = value
			i = next
		case strings.HasPrefix(arg, "--config="):
			value, err := requireInlineFlagValue(arg, "--config", "/path/to/icuvisor.json")
			if err != nil {
				return config.Options{}, err
			}
			opts.Path = value
		case arg == "--transport":
			value, next, err := requireFlagValue(args, i, "--transport", "stdio|http")
			if err != nil {
				return config.Options{}, err
			}
			opts.Transport = value
			i = next
		case strings.HasPrefix(arg, "--transport="):
			value, err := requireInlineFlagValue(arg, "--transport", "stdio|http")
			if err != nil {
				return config.Options{}, err
			}
			opts.Transport = value
		case arg == "--http-bind":
			value, next, err := requireFlagValue(args, i, "--http-bind", "127.0.0.1:8765")
			if err != nil {
				return config.Options{}, err
			}
			opts.HTTPBindAddress = value
			i = next
		case strings.HasPrefix(arg, "--http-bind="):
			value, err := requireInlineFlagValue(arg, "--http-bind", "127.0.0.1:8765")
			if err != nil {
				return config.Options{}, err
			}
			opts.HTTPBindAddress = value
		case arg == "--http-stateless":
			value, next, err := requireFlagValue(args, i, "--http-stateless", "true|false")
			if err != nil {
				return config.Options{}, err
			}
			opts.HTTPStateless = value
			i = next
		case strings.HasPrefix(arg, "--http-stateless="):
			value, err := requireInlineFlagValue(arg, "--http-stateless", "true|false")
			if err != nil {
				return config.Options{}, err
			}
			opts.HTTPStateless = value
		case arg == "--http-json-response":
			value, next, err := requireFlagValue(args, i, "--http-json-response", "true|false")
			if err != nil {
				return config.Options{}, err
			}
			opts.HTTPJSONResponse = value
			i = next
		case strings.HasPrefix(arg, "--http-json-response="):
			value, err := requireInlineFlagValue(arg, "--http-json-response", "true|false")
			if err != nil {
				return config.Options{}, err
			}
			opts.HTTPJSONResponse = value
		case arg == "--env-file":
			value, next, err := requireFlagValue(args, i, "--env-file", "/path/to/icuvisor.env")
			if err != nil {
				return config.Options{}, err
			}
			opts.DotEnvPath = value
			opts.DotEnvExplicit = true
			i = next
		case strings.HasPrefix(arg, "--env-file="):
			value, err := requireInlineFlagValue(arg, "--env-file", "/path/to/icuvisor.env")
			if err != nil {
				return config.Options{}, err
			}
			opts.DotEnvPath = value
			opts.DotEnvExplicit = true
		default:
			return config.Options{}, newUsageError("unknown command or flag %q (try: icuvisor version)", arg)
		}
	}
	return opts, nil
}

func requireFlagValue(args []string, index int, name string, example string) (string, int, error) {
	next := index + 1
	if next >= len(args) || strings.TrimSpace(args[next]) == "" || strings.HasPrefix(args[next], "--") {
		return "", index, newUsageError("missing value for %s; use %s %s", name, name, example)
	}
	return args[next], next, nil
}

func requireInlineFlagValue(arg string, name string, example string) (string, error) {
	value, _ := strings.CutPrefix(arg, name+"=")
	if strings.TrimSpace(value) == "" {
		return "", newUsageError("missing value for %s; use %s %s", name, name, example)
	}
	return value, nil
}
