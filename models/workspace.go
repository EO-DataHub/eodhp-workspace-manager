package models

// Represents the payload of a Pulsar message
type WorkspacePayload struct {

	// Defines the action (create, update, patch, delete)
	Action string `json:"action"`

	Name string `json:"name"`

	// Namespace for the Workspace CR (where the Workspace is submitted in K8s - default: workspaces)
	CRNamespace string `json:"crNamespace"`

	// Namespace for the Workspace resources (e.g., PVCs, PVs, etc. - prefixed with 'ws-' by default)
	TargetNamespace string `json:"targetNamespace"`

	// Workspace resource fields - only populated for create and update actions
	ServiceAccountName     *string                  `json:"serviceAccountName,omitempty"`
	AWSRoleName            *string                  `json:"awsRoleName,omitempty"`
	EFSAccessPoint         *[]AWSEFSAccessPoint     `json:"efsAccessPoint,omitempty"`
	S3Buckets              *[]AWSS3Bucket           `json:"s3Buckets,omitempty"`
	PersistentVolumes      *[]PersistentVolume      `json:"persistentVolume,omitempty"`
	PersistentVolumeClaims *[]PersistentVolumeClaim `json:"persistentVolumeClaim,omitempty"`

	// Specific fields for patch operations (e.g., partial updates) - dynamic
	PatchFields map[string]interface{} `json:"patchFields,omitempty"`
}
