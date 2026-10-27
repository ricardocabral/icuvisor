package tools

import (
	"context"
	"errors"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	addSharedWorkoutFolderName        = "add_shared_workout_folder"
	addSharedWorkoutFolderDescription = "Add one publicly shared folder or training plan to the selected athlete's workout library by folder_id from search_workout_library. This explicit write does not schedule calendar workouts. Terse summaries omit folder prose; include_full adds only description and blurb while retaining owner, share-token and child-workout exclusions. Success requires confirmation that the requested folder appears in the selected athlete's library. The POST is not automatically retried; if interrupted or unverified, inspect get_workout_library before retrying."
	addSharedWorkoutFolderEndpoint    = "/athlete/{id}/folders/{folderId}/share-with-me"
	addSharedWorkoutFolderMessage     = "could not verify shared folder in the selected athlete's library; check credentials and folder ID, and inspect get_workout_library before retrying"
)

// SharedWorkoutFolderAdderClient adds public folders and verifies membership in the selected athlete's library.
type SharedWorkoutFolderAdderClient interface {
	AddSharedWorkoutFolder(context.Context, string) (intervals.WorkoutFolder, error)
}

type addSharedWorkoutFolderRequest struct {
	FolderID    string `json:"folder_id"`
	IncludeFull bool   `json:"include_full,omitempty"`
}
type addSharedWorkoutFolderResponse struct {
	Folder sharedWorkoutFolderRow     `json:"folder"`
	Meta   addSharedWorkoutFolderMeta `json:"_meta"`
}
type addSharedWorkoutFolderMeta struct {
	Operation         string `json:"operation"`
	SourceEndpoint    string `json:"source_endpoint"`
	VerifiedInLibrary bool   `json:"verified_in_library"`
	IncludeFull       bool   `json:"include_full"`
}

func newAddSharedWorkoutFolderTool(client SharedWorkoutFolderAdderClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: addSharedWorkoutFolderName, Description: addSharedWorkoutFolderDescription, InputSchema: addSharedWorkoutFolderInputSchema(), OutputSchema: genericOutputSchema("The requested folder summary and _meta.verified_in_library=true after membership verification in the selected athlete's library; include_full adds folder descriptions/blurb only, never owner details, tokens or child workout payloads."), Requirement: RequirementWrite, Handler: func(ctx context.Context, req Request) (Result, error) {
		args, err := DecodeStrict[addSharedWorkoutFolderRequest](req.Arguments)
		if err != nil {
			return Result{}, NewUserError("invalid add_shared_workout_folder arguments; provide a positive numeric folder_id from public search", err)
		}
		id, err := intervals.NormalizeWorkoutFolderID(args.FolderID)
		if err != nil {
			return Result{}, NewUserError("invalid add_shared_workout_folder arguments; provide a positive numeric folder_id from public search", err)
		}
		if client == nil {
			return Result{}, NewUserError(addSharedWorkoutFolderMessage, errors.New("missing shared folder client"))
		}
		folder, err := client.AddSharedWorkoutFolder(ctx, id)
		if err != nil {
			return Result{}, NewUserError(addSharedWorkoutFolderMessage, err)
		}
		if folder.ID != id {
			return Result{}, NewUserError(addSharedWorkoutFolderMessage, intervals.ErrSharedWorkoutFolderUnverified)
		}
		payload := addSharedWorkoutFolderResponse{Folder: sharedWorkoutFolderToRow(folder, args.IncludeFull), Meta: addSharedWorkoutFolderMeta{Operation: "share_with_me", SourceEndpoint: addSharedWorkoutFolderEndpoint, VerifiedInLibrary: true, IncludeFull: args.IncludeFull}}
		return encodeShaped(payload, args.IncludeFull, nil, version, debugMetadata, addSharedWorkoutFolderName, response.UnitSystemMetric, shapeCfg)
	}})
}

func addSharedWorkoutFolderInputSchema() map[string]any {
	examples := []map[string]any{{"folder_id": "12345"}, {"folder_id": "12345", "include_full": true}}
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"folder_id"}, "properties": map[string]any{
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "Include the verified folder description and blurb. Owner details, share tokens and child workout payloads are always omitted; this affects output detail only."},
		"folder_id":    map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "Public folder or plan ID from search_workout_library, supplied as a positive 32-bit integer string. The selected athlete receives the shared folder; the public owner's athlete ID is not a target selector."},
	}, "examples": examples, "input_examples": examples}
}
