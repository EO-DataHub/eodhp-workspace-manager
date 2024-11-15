package k8s

import (
	"context"

	"github.com/EO-DataHub/eodhp-workspace-manager/models"
)

func (c *Client) CreateWorkspace(ctx context.Context, payload models.WorkspacePayload) error {
	// Implement create logic...
	return nil
}

func (c *Client) UpdateWorkspace(ctx context.Context, payload models.WorkspacePayload) error {
	// Implement update logic...
	return nil
}

func (c *Client) DeleteWorkspace(ctx context.Context, payload models.WorkspacePayload) error {
	// Implement delete logic...
	return nil
}

func (c *Client) PatchWorkspace(ctx context.Context, payload models.WorkspacePayload) error {
	// Implement patch logic...
	return nil
}
