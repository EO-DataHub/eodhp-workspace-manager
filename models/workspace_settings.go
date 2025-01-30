package models

import (
	"time"

	"github.com/google/uuid"
)

// WorkspaceSettings represents the configuration of a workspace.
type WorkspaceSettings struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Account     uuid.UUID `json:"account"`
	MemberGroup string    `json:"member_group"`
	Status      string    `json:"status"`
	Stores      *[]Stores `json:"stores"`
	LastUpdated time.Time `json:"last_updated"`
}

// Stores holds lists of object and block stores associated with a workspace.
type Stores struct {
	Object []ObjectStore `json:"object"`
	Block  []BlockStore  `json:"block"`
}

// ObjectStore represents an object storage entry with related metadata.
type ObjectStore struct {
	StoreID        uuid.UUID `json:"store_id"`
	Name           string    `json:"name"`
	Bucket         string    `json:"bucket"`
	Prefix         string    `json:"prefix"`
	Host           string    `json:"host"`
	EnvVar         string    `json:"env_var"`
	AccessPointArn string    `json:"access_point_arn"`
}

// BlockStore represents a block storage entry with related metadata.
type BlockStore struct {
	StoreID       uuid.UUID `json:"store_id"`
	Name          string    `json:"name"`
	AccessPointID string    `json:"access_point_id"`
	MountPoint    string    `json:"mount_point"`
}
