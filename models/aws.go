package models

// AWSEFSAccessPoint represents an AWS Elastic File System (EFS) access point.
type AWSEFSAccessPoint struct {
	Name        string `json:"name"`
	FSID        string `json:"fsid"`
	RootDir     string `json:"rootDir"`
	UID         int    `json:"uid"`
	GID         int    `json:"gid"`
	Permissions string `json:"permissions"`
}

// AWSS3Bucket represents an S3 bucket configuration.
type AWSS3Bucket struct {
	BucketName      string `json:"bucketName"`
	BucketPath      string `json:"bucketPath"`
	AccessPointName string `json:"accessPointName"`
	EnvVar          string `json:"envVar"`
}