package models

import (
	workspacev1alpha1 "github.com/UKEODHP/workspace-controller/api/v1alpha1"
)

// WorkspaceRequest represents the request data for a workspace operation.
type WorkspaceRequest struct {
	Name                   string                  `json:"name"`
	Namespace              string                  `json:"namespace"`
	ServiceAccountName     string                  `json:"serviceAccountName"`
	AWSRoleName            string                  `json:"awsRoleName"`
	EFSAccessPoint         []AWSEFSAccessPoint     `json:"efsAccessPoint"`
	S3Buckets              []AWSS3Bucket           `json:"s3Buckets"`
	PersistentVolumes      []PersistentVolume      `json:"persistentVolume"`
	PersistentVolumeClaims []PersistentVolumeClaim `json:"persistentVolumeClaim"`
}

type WorkspaceMessage struct {
	Action string           `json:"action"`
	Spec   WorkspaceRequest `json:"spec"` // Using models from workspace-manager
}

func WorkspaceRequestToSpec(req *WorkspaceRequest) *workspacev1alpha1.WorkspaceSpec {
	return &workspacev1alpha1.WorkspaceSpec{
		Namespace: req.Namespace, // Map Namespace directly
		AWS: workspacev1alpha1.AWSSpec{
			RoleName: req.AWSRoleName,                        // Map Role Name
			EFS:      mapEFSAccessPoints(req.EFSAccessPoint), // Map EFS Access Points
			S3:       mapS3Buckets(req.S3Buckets),            // Map S3 Buckets
		},
		ServiceAccount: workspacev1alpha1.ServiceAccountSpec{
			Name: req.ServiceAccountName, // Map Service Account Name
		},
		Storage: workspacev1alpha1.StorageSpec{
			PersistentVolumes:      mapPersistentVolumes(req.PersistentVolumes),           // Map Persistent Volumes
			PersistentVolumeClaims: mapPersistentVolumeClaims(req.PersistentVolumeClaims), // Map Persistent Volume Claims
		},
	}
}

// Helper functions to map EFS, S3, PersistentVolumes, etc.
func mapEFSAccessPoints(efsAccessPoints []AWSEFSAccessPoint) workspacev1alpha1.EFSSpec {
	mappedEFS := workspacev1alpha1.EFSSpec{
		AccessPoints: []workspacev1alpha1.EFSAccess{}, // Initialize an empty slice of EFSAccess
	}
	for _, efs := range efsAccessPoints {
		mappedEFS.AccessPoints = append(mappedEFS.AccessPoints, workspacev1alpha1.EFSAccess{
			Name:          efs.Name,
			FSID:          efs.FSID,
			RootDirectory: efs.RootDir,
			User: workspacev1alpha1.User{
				UID: int64(efs.UID),
				GID: int64(efs.GID),
			},
			Permissions: efs.Permissions,
		})
	}
	return mappedEFS
}

func mapS3Buckets(s3Buckets []AWSS3Bucket) workspacev1alpha1.S3Spec {
	mappedS3 := workspacev1alpha1.S3Spec{
		Buckets: []workspacev1alpha1.S3Bucket{}, // Initialize an empty slice of S3Bucket
	}
	for _, bucket := range s3Buckets {
		mappedS3.Buckets = append(mappedS3.Buckets, workspacev1alpha1.S3Bucket{
			Name:            bucket.BucketName,
			Path:            bucket.BucketPath,
			AccessPointName: bucket.AccessPointName,
			EnvVar:          bucket.EnvVar,
		})
	}
	return mappedS3
}

func mapPersistentVolumes(pvs []PersistentVolume) []workspacev1alpha1.PVSpec {
	mappedPVs := []workspacev1alpha1.PVSpec{}
	for _, pv := range pvs {
		mappedPVs = append(mappedPVs, workspacev1alpha1.PVSpec{
			Name:         pv.PVName,
			StorageClass: pv.StorageClass,
			Size:         pv.Size,
			VolumeSource: &workspacev1alpha1.VolumeSource{
				Driver:          pv.Driver,
				AccessPointName: pv.AccessPointName,
			},
		})
	}
	return mappedPVs
}

func mapPersistentVolumeClaims(pvcs []PersistentVolumeClaim) []workspacev1alpha1.PVCSpec {
	var mappedPVCs []workspacev1alpha1.PVCSpec
	for _, pvc := range pvcs {
		mappedPVCs = append(mappedPVCs, workspacev1alpha1.PVCSpec{
			PVSpec: workspacev1alpha1.PVSpec{
				Name:         pvc.PVCName,
				StorageClass: pvc.StorageClass,
				Size:         pvc.Size,
			},
			PVName: pvc.PVName,
		})
	}
	return mappedPVCs
}
