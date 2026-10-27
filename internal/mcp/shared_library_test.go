package mcp

import (
	"net/http"
	"sync/atomic"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/toolcatalog"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

func TestProtocolSharedLibraryCoachAuthorization(t *testing.T) {
	var posts atomic.Int32
	client, closeUpstream := newProtocolIntervalsClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/athlete/i222/folders/10/share-with-me" && r.URL.Path != "/athlete/i222/folders" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(403)
			return
		}
		if r.Method == http.MethodPost {
			posts.Add(1)
		}
		_, _ = w.Write([]byte(`[{"id":10,"athlete_id":"i999"}]`))
	}))
	defer closeUpstream()
	cfg := config.Config{AthleteID: "i111", CoachMode: coach.ModeOn, Coach: coach.Config{DefaultAthleteID: "i111", Athletes: []coach.Athlete{{ID: "i111", AllowedTools: []string{toolcatalog.GetAthleteProfile}}, {ID: "i222", AllowedTools: []string{toolcatalog.AddSharedWorkoutFolder}}}}}
	cap := safety.NewCapability(safety.ModeSafe)
	registry := tools.NewRegistryWithOptions(client, tools.RegistryOptions{Version: "test", Capability: cap, Toolset: safety.ToolsetFull, CoachModeEnabled: true, CoachConfig: cfg.Coach})
	ctx, session, cleanup := connectTestClientWithOptions(t, Options{Config: cfg, Registry: registry, Capability: cap, Toolset: safety.ToolsetFull})
	defer cleanup()
	listed, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if hasTool(listed.Tools, toolcatalog.AddSharedWorkoutFolder) {
		t.Fatal("denied default athlete sees sharing")
	}
	if !hasTool(listed.Tools, toolcatalog.SearchWorkoutLibrary) {
		t.Fatal("global public search hidden by per-athlete ACL")
	}
	for _, target := range []string{"i111", "i333", "i222"} {
		got, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: toolcatalog.AddSharedWorkoutFolder, Arguments: map[string]any{"folder_id": "10", "athlete_id": target}})
		if err != nil {
			t.Fatal(err)
		}
		if got.IsError != (target != "i222") {
			t.Fatalf("target %s IsError=%t", target, got.IsError)
		}
		if target != "i222" && posts.Load() != 0 {
			t.Fatal("unauthorized target reached POST")
		}
	}
	if posts.Load() != 1 {
		t.Fatalf("posts=%d", posts.Load())
	}
}
