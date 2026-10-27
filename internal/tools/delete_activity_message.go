package tools

const deleteActivityMessageName = "delete_activity_message"

func newDeleteActivityMessageTool(client ActivityMessageEditorClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	return fullTool(Tool{Name: deleteActivityMessageName, Description: "Remove your own existing comment from one activity and return its previous text for recovery. This destructive tool is registered only when ICUVISOR_DELETE_MODE=full; already-deleted own comments return already_deleted without another write. Upstream permissions still apply.", InputSchema: changeActivityMessageInputSchema(true), OutputSchema: changeActivityMessageOutputSchema(), Requirement: RequirementDelete, Handler: changeActivityMessageHandler(client, true, version, timezoneFallback, debugMetadata, responseShapingOrDefault(shaping))})
}
