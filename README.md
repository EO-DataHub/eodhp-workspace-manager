# EO DataHub Workspace Manager
The Workspace Manager is a service that listens to Apache Pulsar events and performs CRUD operations on Kubernetes `Workspace` Custom Resources. It transforms incoming requests into Kubernetes CRDs and performs the necessary actions (create, update, patch, delete) in the cluster. The actual reconciliation and management of the workspace resources are handled by the separate `workspace-controller`, which reacts to the creation / modification of these CRDs.

The Workspace Manager only handles the lifecycle of `Workspace` CRDs in Kubernetes. Once the CRDs actions are performed, a separate workspace-controller (which resides in another repository) is responsible for reconciling the workspace resources, creating cloud infrastructure, and provisioning the actual workspace.


## Getting Started
### Requisites
- Go 1.16 or higher

### Installation
Clone the repository:
```
git clone git@github.com:EO-DataHub/eodhp-workspace-manager.git
cd eodhp-workspace-manager
```


### Run Locally

If you wanta local pulsar server running to test against, make sure it is installed and then run `./pulsar standalone`

```
go run main.go
```

### Deployment
TODO

## Payload Structure
The Workspace Manager expects the following JSON payload structures via Apache Pulsar. These payloads determines which CRUD operation will be performed on the Workspace resource in Kubernetes.

### Create Workspace Payload
```
{
  "action": "create",
  "name": "test-workspace",
  "crNamespace": "workspaces",
  "targetNamespace": "ws-test-workspace",
  "serviceAccountName": "default",
  "awsRoleName": "my-aws-role",
  "efsAccessPoint": [
    {
      "fsID": "fs-xxxxxx",
      "name": "efs-access-point",
      "rootDirectory": "/workspaces/test-workspace",
      "permissions": "755",
      "uid": 1000,
      "gid": 1000
    }
  ],
  "s3Buckets": [
    {
      "name": "workspace-bucket",
      "accessPointName": "s3-access-point"
    }
  ],
  "persistentVolumes": [
    {
      "name": "pv-test-workspace",
      "size": "10Gi",
      "storageClass": "standard",
      "volumeSource": {
        "driver": "efs.csi.aws.com",
        "accessPointName": "efs-access-point"
      }
    }
  ],
  "persistentVolumeClaims": [
    {
      "name": "pvc-test-workspace",
      "size": "10Gi",
      "storageClass": "standard",
      "pvName": "pv-test-workspace"
    }
  ]
}
```

### Update Workspace Payload
The update payload is similar to the create payload, but it is used to modify an existing `Workspace`. You provide the full updated workspace specification. For example, you might update the AWS role name or storage settings.

```
{
  "action": "update",
  "name": "test-workspace",
  "crNamespace": "workspaces",
  "targetNamespace": "ws-test-workspace",
  "serviceAccountName": "default",
  "awsRoleName": "updated-aws-role",
  "efsAccessPoint": [
    {
      "fsID": "fs-12345678",
      "name": "efs-access-point",
      "rootDirectory": "/workspaces/test-workspace",
      "permissions": "755",
      "uid": 1000,
      "gid": 1000
    }
  ],
  "s3Buckets": [
    {
      "name": "updated-workspace-bucket",
      "accessPointName": "s3-access-point"
    }
  ],
  "persistentVolumes": [
    {
      "name": "pv-test-workspace",
      "size": "20Gi",  // Increased size
      "storageClass": "standard",
      "volumeSource": {
        "driver": "efs.csi.aws.com",
        "accessPointName": "efs-access-point"
      }
    }
  ],
  "persistentVolumeClaims": [
    {
      "name": "pvc-test-workspace",
      "size": "20Gi",  // Increased size
      "storageClass": "standard",
      "pvName": "pv-test-workspace"
    }
  ]
}
```


### Delete Workspace Payload
The delete payload is used to delete an existing `Workspace`. It only requires the action, name, and crNamespace fields. No other specification is needed, as the entire workspace resource will be deleted.

```
{
  "action": "delete",
  "name": "test-workspace",
  "crNamespace": "workspaces"
}
```

### Patch Workspace Payload
The patch payload is used to partially update an existing `Workspace` resource. You only provide the fields you want to modify using the patchFields key.

In this example, we are patching the `awsRoleName` and increasing the storage size of the persistent volume.

```
{
  "action": "patch",
  "name": "test-workspace",
  "crNamespace": "workspaces",
  "patchFields": {
    "awsRoleName": "patched-aws-role",
    "persistentVolumes": [
      {
        "name": "pv-test-workspace",
        "size": "30Gi"  // Increased size in patch
      }
    ]
  }
}
```