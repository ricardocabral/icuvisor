package toolcatalog

import "testing"

func TestValidateACLPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "star", input: " * ", want: "*"},
		{name: "exact", input: "get_athlete_profile", want: "get_athlete_profile"},
		{name: "prefix", input: "get_*", want: "get_*"},
		{name: "unknown exact", input: "get_athlete_profiel", wantErr: true},
		{name: "coach control tool outside ACL", input: "select_athlete", wantErr: true},
		{name: "unknown prefix", input: "bogus_*", wantErr: true},
		{name: "interior wildcard", input: "get_*_profile", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ValidateACLPattern(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatal("ValidateACLPattern() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateACLPattern() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("ValidateACLPattern() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestToolSets(t *testing.T) {
	t.Parallel()

	if !IsKnownTool(SelectAthlete) || IsAthleteScopedTool(SelectAthlete) {
		t.Fatal("select_athlete should be known but outside per-athlete ACLs")
	}
	if !IsKnownTool(ICUvisorListAdvancedCapabilities) || IsAthleteScopedTool(ICUvisorListAdvancedCapabilities) {
		t.Fatal("advanced capabilities should be known but outside per-athlete ACLs")
	}
	if !IsAthleteScopedTool(GetAthleteProfile) {
		t.Fatal("get_athlete_profile should be athlete-scoped")
	}
	if len(AllToolNames()) <= len(AthleteScopedToolNames()) {
		t.Fatal("all tool names should include non-athlete-scoped control tools")
	}
}
