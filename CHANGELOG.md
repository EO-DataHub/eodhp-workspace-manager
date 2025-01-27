# Workspace Manager

## v0.1.1 (27-01-2025)

- Added `state` to the `WorkspaceStatus` struct
- Added cluster level root bucket to the config
- Using latest `eodhp-workspace-controller` import version with Workspace CRD updates
- Generate Storage Config access points based on the block store names

## v0.1.0 (13-01-2025)

- Monitors `Workspace` CRD and detects changes and sends status messages to pulsar `workspace-status`
- Listens for Pulsar messages from `workspace-settings` topic and performs operations on the `Workspace` in k8s
