# EO DataHub Workspace Manager
The Workspace Manager is a serivce that has two distinct roles:

1. Monitors the Workspace CRD. It detects changes to the `status` and produces a message to send directly to the `workspace-status` pulsar topic

2. Listens for Pulsar messages from the `workspace-settings` topic. It then operates on the settings, applying the K8s client to operate on a `Workspace` to reflect the desired status as specified in the message.  The actual reconciliation and management of the workspace resources are handled by the separate `workspace-controller`, which reacts to the creation / modification of these CRDs.w


## Getting Started
### Requisites
- Go 1.22 or higher

### Installation
Clone the repository:
```
git clone git@github.com:EO-DataHub/eodhp-workspace-manager.git
cd eodhp-workspace-manager
```

### Configuration
On deployment, the `workspace-manager` reads a config file. It is templated as follows:

```yaml
pulsar:
  url: ...
  topicProducer: persistent://public/default/workspace-status
  topicConsumer: persistent://public/default/workspace-configuration
  subscription: ...
logLevel: INFO
aws:
  cluster: eodhp-...
  fsId: ...
storage:
  size: 10Gi
  storageClass: file-storage
  pvcName: workspace-pvc
  driver: efs.csi.aws.com
```

### Run Locally

If you wanta local pulsar server running to test against, make sure it is installed and then run `./pulsar standalone`

```
go run cmd/main.go --config {path/to/config.yaml}
```

