package safety

import (
	"log/slog"
	"strings"
)

const EnvDeleteMode = "ICUVISOR_DELETE_MODE"

// Mode controls which write and delete tools may be registered.
type Mode string

const (
	// ModeSafe allows write tools but skips delete tools.
	ModeSafe Mode = "safe"
	// ModeFull allows write and delete tools.
	ModeFull Mode = "full"
	// ModeNone skips all write and delete tools.
	ModeNone Mode = "none"
)

// ParseMode resolves raw ICUVISOR_DELETE_MODE values. Empty or unknown values
// intentionally fall back to safe so misconfiguration never unlocks deletes.
func ParseMode(value string) Mode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(ModeFull):
		return ModeFull
	case string(ModeNone):
		return ModeNone
	case string(ModeSafe), "":
		return ModeSafe
	default:
		return ModeSafe
	}
}

// Capability reports the process-global write/delete permissions used when registering tools.
type Capability interface {
	CanDelete() bool
	CanWrite() bool
	Mode() string
}

type staticCapability struct {
	mode Mode
}

// NewCapability returns an immutable, concurrent-read-safe capability value for mode.
func NewCapability(mode Mode) Capability {
	return staticCapability{mode: ParseMode(mode.String())}
}

// NewCapabilityFromEnv resolves ICUVISOR_DELETE_MODE once and returns the resulting capability.
func NewCapabilityFromEnv(getenv func(string) string) Capability {
	if getenv == nil {
		return NewCapability(ModeSafe)
	}
	return NewCapability(ParseMode(getenv(EnvDeleteMode)))
}

func (c staticCapability) CanDelete() bool {
	return c.mode == ModeFull
}

func (c staticCapability) CanWrite() bool {
	return c.mode == ModeSafe || c.mode == ModeFull
}

func (c staticCapability) Mode() string {
	return c.mode.String()
}

// LogResolvedMode emits the single startup log entry for the process delete mode.
func LogResolvedMode(logger *slog.Logger, mode Mode) {
	if logger == nil {
		logger = slog.Default()
	}
	logger.Info("resolved delete mode", "mode", mode.String())
}

func (m Mode) String() string {
	switch m {
	case ModeFull, ModeNone, ModeSafe:
		return string(m)
	default:
		return string(ModeSafe)
	}
}
