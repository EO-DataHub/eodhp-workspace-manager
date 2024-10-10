package models

// PersistentVolume represents a Kubernetes Persistent Volume.
type PersistentVolume struct {
	PVName          string `json:"pvName"`
	StorageClass    string `json:"storageClass"`
	Size            string `json:"size"`
	Driver          string `json:"driver"`
	AccessPointName string `json:"accessPointName"`
}

// PersistentVolumeClaim represents a Kubernetes Persistent Volume Claim.
type PersistentVolumeClaim struct {
	PVCName      string `json:"pvcName"`
	StorageClass string `json:"storageClass"`
	Size         string `json:"size"`
	PVName       string `json:"pvName"`
}
