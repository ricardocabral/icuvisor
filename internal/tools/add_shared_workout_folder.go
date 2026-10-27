package tools

import (
	"context"
	"errors"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	addSharedWorkoutFolderName        = "add_shared_workout_folder"
	addSharedWorkoutFolderDescription = "Add one publicly shared folder or training plan to the selected athlete's workout library by folder_id from search_workout_library. This explicit write does not schedule calendar workouts. Success requires confirmation that the requested folder appears in the selected athlete's library. The POST is not automatically retried; if interrupted or unverified, inspect get_workout_library before retrying."
	addSharedWorkoutFolderEndpoint    = "/athlete/{id}/folders/{folderId}/share-with-me"
	addSharedWorkoutFolderMessage     = "could not verify shared folder in the selected athlete's library; check credentials and folder ID, and inspect get_workout_library before retrying"
)

// SharedWorkoutFolderAdderClient adds public folders and verifies membership in the selected athlete's library.
type SharedWorkoutFolderAdderClient interface {
	AddSharedWorkoutFolder(context.Context, string) (intervals.WorkoutFolder, error)
}

type addSharedWorkoutFolderRequest struct {
	FolderID string `json:"folder_id"`
}
type addSharedWorkoutFolderResponse struct {
	Folder sharedWorkoutFolderRow     `json:"folder"`
	Meta   addSharedWorkoutFolderMeta `json:"_meta"`
}
type addSharedWorkoutFolderMeta struct {
	Operation         string `json:"operation"`
	SourceEndpoint    string `json:"source_endpoint"`
	VerifiedInLibrary bool   `json:"verified_in_library"`
}

func newAddSharedWorkoutFolderTool(client SharedWorkoutFolderAdderClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: addSharedWorkoutFolderName, Description: addSharedWorkoutFolderDescription, InputSchema: addSharedWorkoutFolderInputSchema(), OutputSchema: genericOutputSchema("The requested folder summary and _meta.verified_in_library=true after membership verification in the selected athlete's library; no owner details, tokens or child workout payloads."), Requirement: RequirementWrite, Handler: func(ctx context.Context, req Request) (Result, error) {
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
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(addSharedWorkoutFolderMessage, err)
		}
		if folder.ID != id {
			return Result{}, NewUserError(addSharedWorkoutFolderMessage, intervals.ErrSharedWorkoutFolderUnverified)
		}
		payload := addSharedWorkoutFolderResponse{Folder: sharedWorkoutFolderToRow(folder, false), Meta: addSharedWorkoutFolderMeta{Operation: "share_with_me", SourceEndpoint: addSharedWorkoutFolderEndpoint, VerifiedInLibrary: true}}
		return encodeShaped(payload, false, nil, version, debugMetadata, addSharedWorkoutFolderName, response.UnitSystemMetric, shapeCfg)
	}})
}

func addSharedWorkoutFolderInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"folder_id"}, "properties": map[string]any{
		"folder_id": map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "Public folder or plan ID from search_workout_library, supplied as a positive 32-bit integer string. The selected athlete receives the shared folder; the public owner's athlete ID is not a target selector."},
	}, "examples": []map[string]any{{"folder_id": "12345"}}}
}
