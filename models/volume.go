package models

import (
	workspacev1alpha1 "github.com/UKEODHP/workspace-controller/api/v1alpha1"
)

// Represents a Kubernetes Persistent Volume.
type PersistentVolume struct {
	PVName          string `json:"pvName"`
	StorageClass    string `json:"storageClass"`
	Size            string `json:"size"`
	Driver          string `json:"driver"`
	AccessPointName string `json:"accessPointName"`
}

// Represents a Kubernetes Persistent Volume Claim.
type PersistentVolumeClaim struct {
	PVCName      string `json:"pvcName"`
	StorageClass string `json:"storageClass"`
	Size         string `json:"size"`
	PVName       string `json:"pvName"`
}

// Mapping function to convert request PersistentVolumes to Workspace PVSpec format
func MapPersistentVolumes(pvs *[]PersistentVolume) []workspacev1alpha1.PVSpec {
	if pvs == nil {
		return nil
	}

	var result []workspacev1alpha1.PVSpec
	for _, pv := range *pvs {
		result = append(result, workspacev1alpha1.PVSpec{
			Name:         pv.PVName,
			StorageClass: pv.StorageClass,
			Size:         pv.Size,
			VolumeSource: &workspacev1alpha1.VolumeSource{
				Driver:          pv.Driver,
				AccessPointName: pv.AccessPointName,
			},
		})
	}
	return result
}

// Mapping function to convert request PersistentVolumeClaims to Workspace PVCSpec format
func MapPersistentVolumeClaims(pvcs *[]PersistentVolumeClaim) []workspacev1alpha1.PVCSpec {
	if pvcs == nil {
		return nil
	}

	var result []workspacev1alpha1.PVCSpec
	for _, pvc := range *pvcs {
		result = append(result, workspacev1alpha1.PVCSpec{
			PVSpec: workspacev1alpha1.PVSpec{
				Name:         pvc.PVCName,
				StorageClass: pvc.StorageClass,
				Size:         pvc.Size,
			},
			PVName: pvc.PVName,
		})
	}
	return result
}
