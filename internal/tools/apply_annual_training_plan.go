package tools

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

const (
	applyAnnualTrainingPlanName                    = "apply_annual_training_plan"
	applyAnnualTrainingPlanDescription             = "Safely preview and apply phase notes from propose_annual_training_plan. Defaults to dry_run:true, requires a returned preview_token before dry_run:false writes, uses deterministic external_id values for retry-safe Icuvisor-owned ATP phase notes, and reports protected races, notes, workouts, existing ATP rows, and unknown categories instead of deleting or overwriting them."
	invalidApplyAnnualTrainingPlanArgumentsMessage = "invalid apply_annual_training_plan arguments; provide a valid propose_annual_training_plan proposal, dry_run preview first, and a preview_token for commit"
	applyAnnualTrainingPlanMessage                 = "could not apply annual training plan; check proposal JSON, calendar conflicts, credentials, athlete ID, and delete-mode configuration"
	applyAnnualTrainingPlanConflictSkip            = "skip_existing"
	applyAnnualTrainingPlanConflictReplaceOwned    = "replace_icuvisor_notes"
	applyAnnualTrainingPlanExternalIDPrefix        = "icuvisor-season-plan-v1-"
	applyAnnualTrainingPlanPreviewTokenPrefix      = "season-plan-preview-v1-"
	applyAnnualTrainingPlanDigestLength            = 24
	applyAnnualTrainingPlanMaxWeeks                = seasonPlanMaxWeeks
	applyAnnualTrainingPlanNoteMarker              = "<!-- icuvisor:season-plan:v1 -->"
)

// ApplyAnnualTrainingPlanClient reads calendar conflicts and writes season-plan phase notes.
type ApplyAnnualTrainingPlanClient interface {
	EventsClient
	EventWriterClient
}

type applyAnnualTrainingPlanRequest struct {
	Proposal       seasonPlanProposalResponse `json:"proposal"`
	DryRun         *bool                      `json:"dry_run,omitempty"`
	ConflictPolicy string                     `json:"conflict_policy,omitempty"`
	PreviewToken   string                     `json:"preview_token,omitempty"`
	IncludeFull    bool                       `json:"include_full,omitempty"`
}

type applyAnnualTrainingPlanWireRequest struct {
	Proposal       json.RawMessage `json:"proposal"`
	DryRun         *bool           `json:"dry_run,omitempty"`
	ConflictPolicy string          `json:"conflict_policy,omitempty"`
	PreviewToken   string          `json:"preview_token,omitempty"`
	IncludeFull    bool            `json:"include_full,omitempty"`
}

type applyAnnualTrainingPlanResponse struct {
	ProposedNotes   []applyAnnualTrainingPlanProposedNote `json:"proposed_notes"`
	AppliedNotes    []getEventsRow                        `json:"applied_notes,omitempty"`
	ProtectedEvents []applyTrainingPlanConflict           `json:"protected_events"`
	Meta            applyAnnualTrainingPlanMeta           `json:"_meta"`
}

type applyAnnualTrainingPlanProposedNote struct {
	PhaseID     string                      `json:"phase_id"`
	Date        string                      `json:"date"`
	EndDate     string                      `json:"end_date"`
	ExternalID  string                      `json:"external_id"`
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Operation   string                      `json:"operation"`
	Conflicts   []applyTrainingPlanConflict `json:"conflicts"`
}

type applyAnnualTrainingPlanMeta struct {
	DryRun             bool          `json:"dry_run"`
	WritesPerformed    bool          `json:"writes_performed"`
	PreviewToken       string        `json:"preview_token"`
	ConflictPolicy     string        `json:"conflict_policy"`
	DeleteMode         string        `json:"delete_mode"`
	Timezone           string        `json:"timezone"`
	DateRange          dateRangeMeta `json:"date_range"`
	ProposedCount      int           `json:"proposed_count"`
	CreateCount        int           `json:"create_count"`
	UpdateCount        int           `json:"update_count"`
	IdempotentCount    int           `json:"idempotent_count"`
	ProtectedCount     int           `json:"protected_count"`
	AppliedExternalIDs []string      `json:"applied_external_ids,omitempty"`
	FailedExternalID   string        `json:"failed_external_id,omitempty"`
	RetrySafe          bool          `json:"retry_safe"`
}

type applyAnnualTrainingPlanPreparedNote struct {
	phase       seasonPlanProposalPhase
	params      intervals.WriteEventParams
	description string
}

func newApplyAnnualTrainingPlanTool(client ApplyAnnualTrainingPlanClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, capability safety.Capability, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: applyAnnualTrainingPlanName, Description: applyAnnualTrainingPlanDescription, InputSchema: applyAnnualTrainingPlanInputSchema(capabilityOrSafe(capability)), OutputSchema: applyAnnualTrainingPlanOutputSchema(), Requirement: RequirementWrite, Handler: applyAnnualTrainingPlanHandler(client, profileClient, version, timezoneFallback, debugMetadata, capabilityOrSafe(capability), shapeCfg)})
}

func applyAnnualTrainingPlanHandler(client ApplyAnnualTrainingPlanClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, capability safety.Capability, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeApplyAnnualTrainingPlanRequest(req.Arguments, capabilityOrSafe(capability))
		if err != nil {
			return Result{}, NewUserError(invalidApplyAnnualTrainingPlanArgumentsMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(applyAnnualTrainingPlanMessage, errors.New("missing apply annual training plan client"))
		}
		unitSystem, timezoneName, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(applyAnnualTrainingPlanMessage, err)
		}
		payload, err := applyAnnualTrainingPlan(ctx, client, args, timezoneName, capabilityOrSafe(capability))
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			var userErr *UserError
			if errors.As(err, &userErr) {
				return Result{}, err
			}
			return Result{}, NewUserError(applyAnnualTrainingPlanMessage, err)
		}
		return encodeShaped(payload, args.IncludeFull, []string{"proposed_notes", "applied_notes", "protected_events"}, version, debugMetadata, applyAnnualTrainingPlanName, unitSystem, shapeCfg)
	}
}

func decodeApplyAnnualTrainingPlanRequest(raw json.RawMessage, capability safety.Capability) (applyAnnualTrainingPlanRequest, error) {
	var args applyAnnualTrainingPlanRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[applyAnnualTrainingPlanWireRequest](raw)
	if err != nil {
		return args, err
	}
	if len(decoded.Proposal) == 0 || strings.TrimSpace(string(decoded.Proposal)) == "" {
		return args, errors.New("proposal is required")
	}
	var proposal seasonPlanProposalResponse
	if err := json.Unmarshal(decoded.Proposal, &proposal); err != nil {
		return args, fmt.Errorf("decoding proposal: %w", err)
	}
	args = applyAnnualTrainingPlanRequest{Proposal: proposal, DryRun: decoded.DryRun, ConflictPolicy: decoded.ConflictPolicy, PreviewToken: decoded.PreviewToken, IncludeFull: decoded.IncludeFull}
	args.ConflictPolicy = strings.TrimSpace(args.ConflictPolicy)
	args.PreviewToken = strings.TrimSpace(args.PreviewToken)
	if args.ConflictPolicy == "" {
		args.ConflictPolicy = applyAnnualTrainingPlanConflictSkip
	}
	if args.ConflictPolicy != applyAnnualTrainingPlanConflictSkip && args.ConflictPolicy != applyAnnualTrainingPlanConflictReplaceOwned {
		return args, errors.New("conflict_policy must be skip_existing or replace_icuvisor_notes")
	}
	if args.ConflictPolicy == applyAnnualTrainingPlanConflictReplaceOwned && !capabilityOrSafe(capability).CanDelete() {
		return args, errors.New("replace_icuvisor_notes requires ICUVISOR_DELETE_MODE=full")
	}
	if err := validateSeasonPlanProposalForApply(args.Proposal); err != nil {
		return args, err
	}
	return args, nil
}

func applyAnnualTrainingPlan(ctx context.Context, client ApplyAnnualTrainingPlanClient, args applyAnnualTrainingPlanRequest, timezoneName string, capability safety.Capability) (applyAnnualTrainingPlanResponse, error) {
	dryRun := true
	if args.DryRun != nil {
		dryRun = *args.DryRun
	}
	prepared := prepareAnnualTrainingPlanNotes(args.Proposal)
	eventsByDate, protectedEvents, err := fetchApplyAnnualTrainingPlanEvents(ctx, client, args.Proposal.Summary.StartDate, args.Proposal.Summary.EndDate, args.ConflictPolicy)
	if err != nil {
		return applyAnnualTrainingPlanResponse{}, err
	}
	payload := applyAnnualTrainingPlanResponse{ProposedNotes: make([]applyAnnualTrainingPlanProposedNote, 0, len(prepared)), ProtectedEvents: protectedEvents, Meta: applyAnnualTrainingPlanMeta{DryRun: dryRun, ConflictPolicy: args.ConflictPolicy, DeleteMode: capabilityOrSafe(capability).Mode(), Timezone: timezoneName, DateRange: dateRangeMeta{Oldest: args.Proposal.Summary.StartDate, Newest: args.Proposal.Summary.EndDate}, ProposedCount: len(prepared), RetrySafe: true}}
	for _, note := range prepared {
		conflicts := applyAnnualTrainingPlanConflictsForNote(note, eventsByDate[note.params.Date], args.ConflictPolicy)
		operation := applyAnnualTrainingPlanOperation(conflicts)
		row := applyAnnualTrainingPlanProposedNote{PhaseID: note.phase.PhaseID, Date: note.params.Date, EndDate: note.phase.EndDate, ExternalID: note.params.ExternalID, Name: note.params.Name, Description: note.description, Operation: operation, Conflicts: conflicts}
		payload.ProposedNotes = append(payload.ProposedNotes, row)
		switch operation {
		case "idempotent_existing":
			payload.Meta.IdempotentCount++
		case "update_owned":
			payload.Meta.UpdateCount++
		case "create":
			payload.Meta.CreateCount++
		}
		for _, conflict := range conflicts {
			if conflict.Protected {
				payload.Meta.ProtectedCount++
			}
		}
	}
	payload.Meta.PreviewToken = applyAnnualTrainingPlanPreviewToken(payload.ProposedNotes, payload.ProtectedEvents, args.ConflictPolicy, payload.Meta.DateRange)
	if dryRun {
		return payload, nil
	}
	if args.PreviewToken == "" || args.PreviewToken != payload.Meta.PreviewToken {
		return applyAnnualTrainingPlanResponse{}, NewUserError(invalidApplyAnnualTrainingPlanArgumentsMessage, errors.New("dry_run:false requires the preview_token returned by a matching dry-run preview"))
	}
	return payload, nil
}

func validateSeasonPlanProposalForApply(proposal seasonPlanProposalResponse) error {
	if proposal.Meta.SchemaVersion != seasonPlanProposalSchemaVersion || !proposal.Meta.ReadOnly || proposal.Meta.WritesPerformed {
		return errors.New("proposal _meta must be read-only season_plan_proposal.v1 with writes_performed=false")
	}
	start, err := time.Parse(time.DateOnly, proposal.Summary.StartDate)
	if err != nil {
		return errors.New("proposal summary.start_date must be YYYY-MM-DD")
	}
	end, err := time.Parse(time.DateOnly, proposal.Summary.EndDate)
	if err != nil {
		return errors.New("proposal summary.end_date must be YYYY-MM-DD")
	}
	goal, err := time.Parse(time.DateOnly, proposal.Summary.GoalDate)
	if err != nil {
		return errors.New("proposal summary.goal_date must be YYYY-MM-DD")
	}
	if end.Before(start) || goal.Before(start) || goal.After(end) {
		return errors.New("proposal summary dates are inconsistent")
	}
	totalWeeks := int(end.Sub(start).Hours()/24)/7 + 1
	if totalWeeks < 1 || totalWeeks > applyAnnualTrainingPlanMaxWeeks || proposal.Summary.TotalWeeks != totalWeeks {
		return fmt.Errorf("proposal total_weeks must be 1-%d and match the summary date range", applyAnnualTrainingPlanMaxWeeks)
	}
	if proposal.Summary.PhaseCount != len(proposal.Phases) || proposal.Summary.RecoveryWeekCount != len(proposal.RecoveryWeeks) || proposal.Summary.RaceAnchorCount != len(proposal.RaceAnchors) {
		return errors.New("proposal summary counts must match phase/recovery/race arrays")
	}
	phaseIDs := map[string]struct{}{}
	phaseRanges := make([]seasonPlanProposalPhase, len(proposal.Phases))
	copy(phaseRanges, proposal.Phases)
	sort.SliceStable(phaseRanges, func(i, j int) bool { return phaseRanges[i].StartDate < phaseRanges[j].StartDate })
	var previousEnd time.Time
	for idx, phase := range phaseRanges {
		if strings.TrimSpace(phase.PhaseID) == "" {
			return errors.New("proposal phases must have phase_id")
		}
		if _, exists := phaseIDs[phase.PhaseID]; exists {
			return errors.New("proposal phase_id values must be unique")
		}
		phaseIDs[phase.PhaseID] = struct{}{}
		phaseStart, err := time.Parse(time.DateOnly, phase.StartDate)
		if err != nil {
			return fmt.Errorf("proposal phase %s start_date must be YYYY-MM-DD", phase.PhaseID)
		}
		phaseEnd, err := time.Parse(time.DateOnly, phase.EndDate)
		if err != nil {
			return fmt.Errorf("proposal phase %s end_date must be YYYY-MM-DD", phase.PhaseID)
		}
		if phaseEnd.Before(phaseStart) || phaseStart.Before(start) || phaseEnd.After(end) {
			return fmt.Errorf("proposal phase %s range is outside summary range", phase.PhaseID)
		}
		if idx > 0 && !phaseStart.After(previousEnd) {
			return errors.New("proposal phase ranges must not overlap")
		}
		previousEnd = phaseEnd
	}
	for _, week := range proposal.WeeklyTargets {
		if _, ok := phaseIDs[week.PhaseID]; !ok {
			return fmt.Errorf("proposal weekly target references unknown phase_id %s", week.PhaseID)
		}
		if !validDate(week.WeekStartDate) || !validDate(week.WeekEndDate) {
			return errors.New("proposal weekly target dates must be YYYY-MM-DD")
		}
	}
	for _, recovery := range proposal.RecoveryWeeks {
		if _, ok := phaseIDs[recovery.PhaseID]; !ok {
			return fmt.Errorf("proposal recovery week references unknown phase_id %s", recovery.PhaseID)
		}
		if !validDate(recovery.WeekStartDate) {
			return errors.New("proposal recovery week dates must be YYYY-MM-DD")
		}
	}
	for _, anchor := range proposal.RaceAnchors {
		anchorDate, err := time.Parse(time.DateOnly, anchor.Date)
		if err != nil || anchorDate.Before(start) || anchorDate.After(end) {
			return errors.New("proposal race anchors must have dates inside the summary range")
		}
	}
	return nil
}

func prepareAnnualTrainingPlanNotes(proposal seasonPlanProposalResponse) []applyAnnualTrainingPlanPreparedNote {
	phases := append([]seasonPlanProposalPhase(nil), proposal.Phases...)
	sort.SliceStable(phases, func(i, j int) bool {
		if phases[i].StartDate != phases[j].StartDate {
			return phases[i].StartDate < phases[j].StartDate
		}
		return phases[i].PhaseID < phases[j].PhaseID
	})
	notes := make([]applyAnnualTrainingPlanPreparedNote, 0, len(phases))
	for _, phase := range phases {
		description := applyAnnualTrainingPlanNoteDescription(proposal, phase)
		params := intervals.WriteEventParams{ExternalID: applyAnnualTrainingPlanExternalID(proposal, phase), Date: phase.StartDate, Category: "NOTE", Name: applyAnnualTrainingPlanNoteName(phase), Description: &description}
		notes = append(notes, applyAnnualTrainingPlanPreparedNote{phase: phase, params: params, description: description})
	}
	return notes
}

func applyAnnualTrainingPlanNoteName(phase seasonPlanProposalPhase) string {
	name := strings.TrimSpace(phase.Name)
	if name == "" {
		name = strings.TrimSpace(phase.PhaseType)
	}
	if name == "" {
		name = phase.PhaseID
	}
	return "Icuvisor season plan: " + name
}

func applyAnnualTrainingPlanNoteDescription(proposal seasonPlanProposalResponse, phase seasonPlanProposalPhase) string {
	lines := []string{
		applyAnnualTrainingPlanNoteMarker,
		"Generated by icuvisor from propose_annual_training_plan.",
		"Phase ID: " + phase.PhaseID,
		"Phase type: " + phase.PhaseType,
		"Date range: " + phase.StartDate + " to " + phase.EndDate,
		fmt.Sprintf("Week range: %d-%d of %d", phase.StartWeekIndex, phase.EndWeekIndex, proposal.Summary.TotalWeeks),
		"Goal date: " + proposal.Summary.GoalDate,
	}
	return strings.Join(lines, "\n")
}

func fetchApplyAnnualTrainingPlanEvents(ctx context.Context, client EventsClient, oldest string, newest string, conflictPolicy string) (map[string][]intervals.Event, []applyTrainingPlanConflict, error) {
	events, err := client.ListEvents(ctx, intervals.ListEventsParams{Oldest: oldest, Newest: newest, Limit: maxEventsLimit})
	if err != nil {
		return nil, nil, fmt.Errorf("fetching annual-plan calendar conflicts: %w", err)
	}
	eventsByDate := map[string][]intervals.Event{}
	protected := []applyTrainingPlanConflict{}
	for _, event := range events {
		date := eventDateOnly(event)
		if date == "" {
			continue
		}
		eventsByDate[date] = append(eventsByDate[date], event)
		conflict := applyAnnualTrainingPlanConflictFromEvent(event, "existing_event_in_range", conflictPolicy)
		if conflict.Protected {
			protected = append(protected, conflict)
		}
	}
	sortApplyAnnualTrainingPlanConflicts(protected)
	return eventsByDate, protected, nil
}

func applyAnnualTrainingPlanConflictsForNote(note applyAnnualTrainingPlanPreparedNote, events []intervals.Event, conflictPolicy string) []applyTrainingPlanConflict {
	conflicts := []applyTrainingPlanConflict{}
	for _, event := range events {
		if eventDateOnly(event) != note.params.Date {
			continue
		}
		conflict := applyAnnualTrainingPlanConflictFromEvent(event, "existing_event_on_phase_start", conflictPolicy)
		ownedPhaseNote := isIcuvisorAnnualPlanPhaseNote(event)
		if eventMatchesExternalID(event, note.params.ExternalID) {
			if applyAnnualTrainingPlanEventBodyMatches(event, note.params) {
				conflict.Reason = "idempotent_existing"
				conflict.Protected = true
			} else if ownedPhaseNote && conflictPolicy == applyAnnualTrainingPlanConflictReplaceOwned {
				conflict.Reason = "icuvisor_owned_phase_note"
				conflict.Protected = false
			} else {
				conflict.Reason = "external_id_body_mismatch"
				conflict.Protected = true
			}
		} else if ownedPhaseNote {
			if conflictPolicy == applyAnnualTrainingPlanConflictReplaceOwned {
				conflict.Reason = "icuvisor_owned_phase_note"
				conflict.Protected = false
			} else {
				conflict.Reason = "existing_icuvisor_phase_note"
				conflict.Protected = true
			}
		}
		conflicts = append(conflicts, conflict)
	}
	sortApplyAnnualTrainingPlanConflicts(conflicts)
	return conflicts
}

func applyAnnualTrainingPlanConflictFromEvent(event intervals.Event, reason string, conflictPolicy string) applyTrainingPlanConflict {
	conflict := applyTrainingPlanConflictFromEvent(event, reason)
	conflict.Protected = applyAnnualTrainingPlanConflictProtected(conflict, event, conflictPolicy)
	return conflict
}

func applyAnnualTrainingPlanConflictProtected(conflict applyTrainingPlanConflict, event intervals.Event, conflictPolicy string) bool {
	if conflict.Reason == "idempotent_existing" || conflict.Reason == "matching_external_id" {
		return true
	}
	if isIcuvisorAnnualPlanPhaseNote(event) && conflictPolicy == applyAnnualTrainingPlanConflictReplaceOwned {
		return false
	}
	return true
}

func applyAnnualTrainingPlanOperation(conflicts []applyTrainingPlanConflict) string {
	operation := "create"
	for _, conflict := range conflicts {
		switch conflict.Reason {
		case "idempotent_existing":
			return "idempotent_existing"
		case "matching_external_id", "icuvisor_owned_phase_note":
			operation = "update_owned"
		}
		if conflict.Protected && conflict.Reason != "idempotent_existing" {
			return "blocked"
		}
	}
	return operation
}

func applyAnnualTrainingPlanExternalID(proposal seasonPlanProposalResponse, phase seasonPlanProposalPhase) string {
	parts := []string{
		"schema_version=" + proposal.Meta.SchemaVersion,
		"summary_start=" + proposal.Summary.StartDate,
		"summary_end=" + proposal.Summary.EndDate,
		"goal_date=" + proposal.Summary.GoalDate,
		"phase_id=" + strings.TrimSpace(phase.PhaseID),
		"phase_type=" + strings.TrimSpace(phase.PhaseType),
		"phase_name=" + strings.TrimSpace(phase.Name),
		"phase_start=" + strings.TrimSpace(phase.StartDate),
		"phase_end=" + strings.TrimSpace(phase.EndDate),
	}
	digest := sha256.Sum256([]byte(strings.Join(parts, "\n")))
	return applyAnnualTrainingPlanExternalIDPrefix + hex.EncodeToString(digest[:])[:applyAnnualTrainingPlanDigestLength]
}

func applyAnnualTrainingPlanPreviewToken(notes []applyAnnualTrainingPlanProposedNote, protected []applyTrainingPlanConflict, conflictPolicy string, dateRange dateRangeMeta) string {
	material := struct {
		Notes          []applyAnnualTrainingPlanProposedNote `json:"notes"`
		Protected      []applyTrainingPlanConflict           `json:"protected"`
		ConflictPolicy string                                `json:"conflict_policy"`
		DateRange      dateRangeMeta                         `json:"date_range"`
	}{Notes: notes, Protected: protected, ConflictPolicy: conflictPolicy, DateRange: dateRange}
	data, _ := json.Marshal(material)
	digest := sha256.Sum256(data)
	return applyAnnualTrainingPlanPreviewTokenPrefix + hex.EncodeToString(digest[:])[:applyAnnualTrainingPlanDigestLength]
}

func applyAnnualTrainingPlanEventBodyMatches(event intervals.Event, params intervals.WriteEventParams) bool {
	return sameText(firstNonEmpty(stringValue(event.Category), anyString(event.Raw["category"])), params.Category) && strings.TrimSpace(stringValue(event.Name)) == strings.TrimSpace(params.Name) && stringValue(event.Description) == stringValue(params.Description)
}

func isIcuvisorAnnualPlanPhaseNote(event intervals.Event) bool {
	category := firstNonEmpty(stringValue(event.Category), anyString(event.Raw["category"]))
	description := firstNonEmpty(stringValue(event.Description), anyString(event.Raw["description"]))
	externalID := firstNonEmpty(stringValue(event.ExternalID), anyString(event.Raw["external_id"]))
	return strings.EqualFold(strings.TrimSpace(category), "NOTE") && strings.Contains(description, applyAnnualTrainingPlanNoteMarker) && strings.HasPrefix(strings.TrimSpace(externalID), applyAnnualTrainingPlanExternalIDPrefix)
}

func sortApplyAnnualTrainingPlanConflicts(conflicts []applyTrainingPlanConflict) {
	sort.SliceStable(conflicts, func(i, j int) bool {
		if conflicts[i].Date != conflicts[j].Date {
			return conflicts[i].Date < conflicts[j].Date
		}
		if conflicts[i].EventID != conflicts[j].EventID {
			return conflicts[i].EventID < conflicts[j].EventID
		}
		return conflicts[i].Reason < conflicts[j].Reason
	})
}

func applyAnnualTrainingPlanInputSchema(capability safety.Capability) map[string]any {
	conflictEnum := []string{applyAnnualTrainingPlanConflictSkip}
	if capabilityOrSafe(capability).CanDelete() {
		conflictEnum = append(conflictEnum, applyAnnualTrainingPlanConflictReplaceOwned)
	}
	examples := applyAnnualTrainingPlanInputExamples()
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"proposal"}, "examples": examples, "input_examples": examples, "properties": map[string]any{
		"proposal":        map[string]any{"type": "object", "description": "Required unmodified structured response from propose_annual_training_plan. The proposal must have _meta.schema_version season_plan_proposal.v1, read_only:true, writes_performed:false, valid bounded dates, consistent counts, unique phases, and weekly/recovery rows referencing known phases."},
		"dry_run":         map[string]any{"type": "boolean", "default": true, "description": "Safety default is true. dry_run:false is rejected unless preview_token exactly matches the token returned by a fresh dry-run preview for the same proposal/conflicts/policy."},
		"preview_token":   map[string]any{"type": "string", "description": "Required for dry_run:false. Copy exactly from _meta.preview_token returned by the dry-run preview."},
		"conflict_policy": map[string]any{"type": "string", "default": applyAnnualTrainingPlanConflictSkip, "enum": conflictEnum, "description": "skip_existing reports conflicts and protects existing rows. replace_icuvisor_notes is accepted only when ICUVISOR_DELETE_MODE=full and may replace only Icuvisor-owned phase notes with deterministic external_id values; races, notes, workouts, existing ATP rows without Icuvisor external IDs, and unknown categories remain protected."},
		"include_full":    map[string]any{"type": "boolean", "default": false, "description": "When true, include full applied event rows when writes are performed; dry-run remains terse."},
	}}
}

func applyAnnualTrainingPlanInputExamples() []map[string]any {
	return []map[string]any{
		{"proposal": map[string]any{"summary": map[string]any{"start_date": "2026-07-13", "end_date": "2026-08-09", "goal_date": "2026-08-03", "total_weeks": 4, "phase_count": 1, "recovery_week_count": 0, "race_anchor_count": 1}, "phases": []any{map[string]any{"phase_id": "phase_01_base", "phase_type": "base", "name": "Base", "start_date": "2026-07-13", "end_date": "2026-08-09", "week_count": 4, "start_week_index": 1, "end_week_index": 4}}, "weekly_targets": []any{}, "recovery_weeks": []any{}, "race_anchors": []any{map[string]any{"date": "2026-08-03", "type": "race", "source": "input", "week_start_date": "2026-08-03"}}, "assumptions": []any{}, "warnings": []any{}, "_meta": map[string]any{"schema_version": seasonPlanProposalSchemaVersion, "read_only": true, "writes_performed": false}}, "dry_run": true},
		{"proposal": map[string]any{"summary": map[string]any{"start_date": "2026-07-13", "end_date": "2026-08-09", "goal_date": "2026-08-03", "total_weeks": 4, "phase_count": 1, "recovery_week_count": 0, "race_anchor_count": 1}, "phases": []any{map[string]any{"phase_id": "phase_01_base", "phase_type": "base", "name": "Base", "start_date": "2026-07-13", "end_date": "2026-08-09", "week_count": 4, "start_week_index": 1, "end_week_index": 4}}, "weekly_targets": []any{}, "recovery_weeks": []any{}, "race_anchors": []any{map[string]any{"date": "2026-08-03", "type": "race", "source": "input", "week_start_date": "2026-08-03"}}, "assumptions": []any{}, "warnings": []any{}, "_meta": map[string]any{"schema_version": seasonPlanProposalSchemaVersion, "read_only": true, "writes_performed": false}}, "dry_run": false, "preview_token": "season-plan-preview-v1-copy-from-dry-run"},
	}
}

func applyAnnualTrainingPlanOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Season-plan apply preview or write result. Dry-run returns proposed phase NOTE rows with deterministic external_id values, operation hints, protected conflicts, and _meta.preview_token. Commits require that token and report created, updated, idempotent, protected, and retry-safe recovery metadata without deleting races, notes, workouts, unknown categories, or non-Icuvisor ATP rows."}
}
