package config

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/safety"
)

func readDotEnv(ctx context.Context, path string) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read .env: %w", err)
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if !recognizedEnvKey(key) {
			continue
		}
		values[key] = cleanEnvValue(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read .env: %w", err)
	}
	return values, nil
}

func cleanEnvValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		quote := value[0]
		if (quote == '\'' || quote == '"') && value[len(value)-1] == quote {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func recognizedEnvKey(key string) bool {
	switch key {
	case EnvAPIKey, EnvAthleteID, EnvConfigPath, EnvTimezone, EnvAPIBaseURL, EnvHTTPTimeout, EnvTransport, EnvHTTPBind, EnvHTTPStateless, EnvHTTPJSONResponse, EnvCoachMode, safety.EnvDeleteMode, safety.EnvToolset:
		return true
	default:
		return false
	}
}
