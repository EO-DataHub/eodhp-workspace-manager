# Workspace Manager

## v0.1.5 (31-03-2025)

- PV/PVC names based on the `<workspace-name>` template - bugfix


## v0.1.4 (12-02-2025)

- PV/PVC names based on the `<workspace-name>` template
- EFS root directory is `/workspaces/<workspace-name>`

## v0.1.3 (04-02-2025)

- PV/PVC names based on the `<workspace-name>-<block-store-name>` template
- EFS Root directory now set to `/workspaces/<workspace-name>/<block-store-name>`
- Revised user response details - i.e. added MountPoint, removed FSID (not needed) etc..

## v0.1.2 (30-01-2025)

- Added `Host`, `Prefix` and `Bucket` to the workspace settings struct for object stores

## v0.1.1 (27-01-2025)

- Added `state` to the `WorkspaceStatus` struct
- Added cluster level root bucket to the config
- Using latest `eodhp-workspace-controller` import version with Workspace CRD updates
- Generate Storage Config access points based on the block store names

## v0.1.0 (13-01-2025)

- Monitors `Workspace` CRD and detects changes and sends status messages to pulsar `workspace-status`
- Listens for Pulsar messages from `workspace-settings` topic and performs operations on the `Workspace` in k8s
