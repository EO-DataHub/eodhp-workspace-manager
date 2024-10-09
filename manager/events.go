package manager

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"

	pulsar "github.com/apache/pulsar-client-go/pulsar"
	"github.com/google/uuid"
)

type WorkspaceMessage struct {
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Action      string    `json:"action"`
}

func HandleMessage(msg pulsar.Message) error {
	var workspaceMsg WorkspaceMessage

	// Parse the message payload
	err := json.Unmarshal(msg.Payload(), &workspaceMsg)
	if err != nil {
		log.Printf("Failed to parse message payload: %v", err)
		return err
	}

	log.Info().Str("action", workspaceMsg.Action).Str("workspace_id", workspaceMsg.WorkspaceID.String()).Msg("Handling incoming message")

	switch workspaceMsg.Action {
	case "create":
		err = handleCreateWorkspace(workspaceMsg)
	case "update":
		err = handleUpdateWorkspace(workspaceMsg)
	case "patch":
		err = handlePatchWorkspace(workspaceMsg)
	case "delete":
		err = handleDeleteWorkspace(workspaceMsg)
	default:
		err = fmt.Errorf("unknown action: %s", workspaceMsg.Action)
	}

	if err != nil {
		log.Error().Err(err).Msg("Error handling workspace operation")
		return err
	}

	log.Info().Str("action", workspaceMsg.Action).Msg("Successfully handled workspace action")
	return nil
}

func handleCreateWorkspace(msg WorkspaceMessage) error {
	log.Info().Str("workspace_id", msg.WorkspaceID.String()).Msg("Creating workspace")
	// Simulate success
	return nil
}

func handleUpdateWorkspace(msg WorkspaceMessage) error {
	log.Info().Str("workspace_id", msg.WorkspaceID.String()).Msg("Updating workspace")
	// Simulate success
	return nil
}

func handlePatchWorkspace(msg WorkspaceMessage) error {
	log.Info().Str("workspace_id", msg.WorkspaceID.String()).Msg("Patching workspace")
	// Simulate success
	return nil
}

func handleDeleteWorkspace(msg WorkspaceMessage) error {
	log.Info().Str("workspace_id", msg.WorkspaceID.String()).Msg("Deleting workspace")
	// Simulate success
	return nil
}
