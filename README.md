# EO DataHub Workspace Manager
The Workspace Manager is a service that has two distinct roles:

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
  topicConsumer: persistent://public/default/workspace-settings
  subscription: ...
  tokenFile: "{{ .PULSAR_TOKEN_FILE }}"
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

- `pulsar.topicConsumer` can be a comma-separated list, e.g. `persistent://public/default/workspace-settings,persistent://public/workspaces/workspace-settings`, to read old and new topics with the same subscription while topics move.
- `pulsar.tokenFile` is the path to a Pulsar JWT. When set, the client authenticates with the token and re-reads the file on every connection, so a rotated token is picked up without a restart. The manager fails to start if the file does not exist. When empty or omitted the client connects anonymously. The config is rendered with `text/template`, so an unset `PULSAR_TOKEN_FILE` renders as `<no value>` and fails; only reference it where the variable is set.

### Run Locally

If you wanta local pulsar server running to test against, make sure it is installed and then run `./pulsar standalone`

```
go run cmd/main.go --config {path/to/config.yaml}
```

