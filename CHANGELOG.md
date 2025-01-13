# Workspace Manager

## v0.1.0 (13-01-2025)

- Monitors `Workspace` CRD and detects changes and sends status messages to pulsar `workspace-status`
- Listens for Pulsar messages from `workspace-settings` topic and performs operations on the `Workspace` in k8s
