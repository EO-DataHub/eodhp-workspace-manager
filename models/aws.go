package models

import (
	workspacev1alpha1 "github.com/UKEODHP/workspace-controller/api/v1alpha1"
)

// Represents an AWS Elastic File System (EFS) access point.
type AWSEFSAccessPoint struct {
	Name        string `json:"name"`
	FSID        string `json:"fsid"`
	RootDir     string `json:"rootDir"`
	UID         int    `json:"uid"`
	GID         int    `json:"gid"`
	Permissions string `json:"permissions"`
}

// Represents an S3 bucket configuration.
type AWSS3Bucket struct {
	BucketName      string `json:"bucketName"`
	BucketPath      string `json:"bucketPath"`
	AccessPointName string `json:"accessPointName"`
	EnvVar          string `json:"envVar"`
}

// Mapping function to convert request AWSEFSAccessPoints to Workspace CRD format
func MapEFSAccessPoints(accessPoints *[]AWSEFSAccessPoint) []workspacev1alpha1.EFSAccess {
	if accessPoints == nil {
		return nil
	}

	var result []workspacev1alpha1.EFSAccess
	for _, ap := range *accessPoints {
		result = append(result, workspacev1alpha1.EFSAccess{
			Name:          ap.Name,
			FSID:          ap.FSID,
			RootDirectory: ap.RootDir,
			User: workspacev1alpha1.User{
				UID: int64(ap.UID),
				GID: int64(ap.GID),
			},
			Permissions: ap.Permissions,
		})
	}
	return result
}

// Mapping function to convert request AWSS3Buckets to Workspace CRD format
func MapS3Buckets(buckets *[]AWSS3Bucket) []workspacev1alpha1.S3Bucket {
    if buckets == nil {
        return nil
    }

    var result []workspacev1alpha1.S3Bucket
    for _, bucket := range *buckets {
        result = append(result, workspacev1alpha1.S3Bucket{
            Name:            bucket.BucketName,
            Path:            bucket.BucketPath,
            AccessPointName: bucket.AccessPointName,
            EnvVar:          bucket.EnvVar,
        })
    }
    return result
}
