package models

import (
	"time"

	workspacev1alpha1 "github.com/EO-DataHub/eodhp-workspace-controller/api/v1alpha1"
)

// WorkspaceStatus represents the status of a Workspace
type WorkspaceStatus struct {
	Name        string                      `json:"name"`
	Namespace   string                      `json:"namespace"`
	AWS         workspacev1alpha1.AWSStatus `json:"status"`
	LastUpdated time.Time                   `json:"last_updated"`
	State       string                      `json:"state"`
}
