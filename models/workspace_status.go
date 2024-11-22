package models

import (
	workspacev1alpha1 "github.com/EO-DataHub/eodhp-workspace-controller/api/v1alpha1"
)

// WorkspaceStatus represents the status of a Workspace
type WorkspaceStatus struct {
	Name      string                      `json:"name"`
	Namespace string                      `json:"namespace"`
	AWS       workspacev1alpha1.AWSStatus `json:"status"`
}
