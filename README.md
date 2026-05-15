![workflow status](https://github.com/jinxf0120/rancher-exporter/actions/workflows/test-build-publish.yml/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/jinxf0120/rancher-exporter)](https://goreportcard.com/report/github.com/jinxf0120/rancher-exporter)

# Unofficial Prometheus Exporter for Rancher

**Note** : This project is not officially supported by Rancher/Suse.

Repository: https://github.com/jinxf0120/rancher-exporter

# Quickstart

## Option 1: Helm (Recommended)

```bash
helm upgrade --install rancher-exporter chart/rancher-exporter \
  --namespace cattle-system-exporter --create-namespace
```

## Option 2: Static Manifests

```bash
kubectl apply -f ./manifests/exporter.yaml
kubectl apply -f ./manifests/servicemonitor-backup-metrics.yaml
```

## Configuration

You can set timers by environment variable:

- `INFORMER_RESYNC_PERIOD=300` - informer resync interval in seconds, default 300. For Rancher CRDs that do not support watch, this is the polling interval.
- `TIMER_GET_LATEST_RANCHER_VERSION=1` - how often get latest rancher version from github, in minutes

When using Helm, override these in `values.yaml`:

```yaml
env:
  INFORMER_RESYNC_PERIOD: "300"
  TIMER_GET_LATEST_RANCHER_VERSION: "1"
```

# How it works

The exporter uses the Kubernetes Informer mechanism to collect Rancher metrics:

- **Informer + Lister**: Watches or lists Rancher custom resources via dynamic client, caches them locally, and provides a lister interface for querying.
- **Watch fallback**: Rancher CRDs (e.g. `management.cattle.io`, `provisioning.cattle.io`) typically do not support the `watch` operation. The informer automatically falls back to list-only mode, relying on the resync period for periodic updates. Resources that do support watch will receive real-time event-driven updates.
- **Event handlers**: When informer cache changes (add/update/delete), registered handlers update the corresponding Prometheus metrics.
- **Offline support**: In air-gapped environments where GitHub API is unreachable, `latest_rancher_version` will report `"unavailable"`.

# Exported Metrics

## Rancher Version

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `installed_rancher_version` | GaugeVec | `version` | Version of the installed Rancher instance |
| `latest_rancher_version` | GaugeVec | `version` | Version of the most recent Rancher release. `"unavailable"` in air-gapped environments |

## Cluster

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `rancher_managed_clusters` | Gauge | - | Number of clusters this Rancher instance is currently managing (excludes `local`) |
| `rancher_managed_rke_clusters` | Gauge | - | Number of managed RKE clusters |
| `rancher_managed_rke2_clusters` | Gauge | - | Number of managed RKE2 clusters |
| `rancher_managed_k3s_clusters` | Gauge | - | Number of managed K3s clusters |
| `rancher_managed_eks_clusters` | Gauge | - | Number of managed EKS clusters |
| `rancher_managed_aks_clusters` | Gauge | - | Number of managed AKS clusters |
| `rancher_managed_gke_clusters` | Gauge | - | Number of managed GKE clusters |
| `cluster_connected` | GaugeVec | `Name` | Whether a downstream cluster is connected to Rancher (1=connected) |
| `cluster_not_connected` | GaugeVec | `Name` | Whether a downstream cluster is NOT connected to Rancher (1=not connected) |
| `cluster_k8s_version` | GaugeVec | `Name`, `Version` | Kubernetes version running in the downstream cluster |
| `rancher_cluster_labels` | GaugeVec | `cluster_id`, `cluster_display_name`, `cluster_name`, `cluster_label_key`, `cluster_label_value` | Labels associated with Rancher clusters |

## Node

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `rancher_managed_nodes` | Gauge | - | Number of managed nodes |
| `rancher_managed_nodes_info` | GaugeVec | `name`, `parent_cluster`, `is_control_plane`, `is_etcd`, `is_worker`, `architecture`, `container_runtime_version`, `kernel_version`, `os`, `os_image` | Additional metadata about downstream cluster nodes |

## User & Token

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `rancher_users` | Gauge | - | Number of users in this Rancher instance |
| `rancher_tokens` | Gauge | - | Number of tokens issued by Rancher |

## Project

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `rancher_projects` | Gauge | - | Number of Projects globally |
| `rancher_project_labels` | GaugeVec | `cluster_name`, `project_id`, `project_display_name`, `project_label_key`, `project_label_value` | Labels associated with Rancher Projects |
| `rancher_project_annotations` | GaugeVec | `cluster_name`, `project_id`, `project_display_name`, `project_annotation_key`, `project_annotation_value` | Annotations associated with Rancher Projects |
| `rancher_project_resourcequota` | GaugeVec | `cluster_name`, `project_id`, `project_display_name`, `project_resource_key`, `project_resource_type` | Default namespace resource quota set for the project (`project_resource_type` = `hard` or `used`) |

## Custom Resources

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `rancher_custom_resource_count` | GaugeVec | `resource_name` | Raw count of Rancher custom resources (`*.cattle.io`) by CRD name |

## Backup & Restore (requires Rancher Backup Operator)

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `rancher_backups_count` | Gauge | - | Number of Rancher backups |
| `rancher_restore_count` | Gauge | - | Number of Rancher restores |
| `rancher_backup` | GaugeVec | `name`, `resourceSetName`, `retentionCount`, `backupType`, `status`, `filename`, `storageLocation`, `nextSnapshot`, `lastSnapshot` | Details regarding a specific backup operation |
| `rancher_restore` | GaugeVec | `name`, `fileName`, `prune`, `storageLocation`, `status`, `restoreTime` | Details regarding a specific restore operation |

# Grafana Dashboards

`./manifests/grafana-dashboard.json` includes a basic dashboard in JSON format that can be imported into Grafana.
`./manifests/grafana-dashboard-projects.json` includes a Rancher project-focused dashboard in JSON format that can be imported into Grafana.
`./manifests/grafana-dashboard-all-cr.json` includes a dynamic dashboard showing counts for all Rancher custom resources (*.cattle.io).
`./manifests/grafana-dashboard-nodes.json` includes a dynamic dashboard showing global node metadata across all managed clusters (ie Kernel/OS versions, roles, etc)
`./manifests/grafana-dashboard-backups.json` includes a dynamic dashboard showing details about Rancher backup operator jobs.

# Building & Running

## Docker

```bash
# Build
make build

# Push
make push

# Run locally
make run
```

## Offline / Air-gapped Build

In offline environments where direct internet access is unavailable, configure proxy and Go module proxy:

```bash
# Using HTTP proxy
make build HTTP_PROXY=http://proxy:port HTTPS_PROXY=http://proxy:port NO_PROXY=localhost,127.0.0.1

# Using GOPROXY (e.g. Athens, goproxy.cn, or internal proxy)
make build GOPROXY=https://goproxy.cn,direct

# Combined
make build HTTP_PROXY=http://proxy:port HTTPS_PROXY=http://proxy:port GOPROXY=https://goproxy.cn,direct
```

For deploying in air-gapped Kubernetes clusters, push the image to your private registry first:

```bash
# Build and push to private registry
make build REG=your-registry.com ORG=your-org
make push REG=your-registry.com ORG=your-org

# Then update values.yaml when installing with Helm
helm upgrade --install rancher-exporter chart/rancher-exporter \
  --namespace cattle-system-exporter --create-namespace \
  --set image.repository=your-registry.com/your-org/rancher-exporter
```

# Developing

By default, the exporter will use in-cluster authentication via an associated service account.

## External cluster config

To test using external authentication via the local `kubeconfig`, set the following environment variable:

```bash
export RANCHER_EXPORTER_EXTERNAL_AUTH=TRUE
```

* `go run main.go`
* Navigate to http://localhost:8080/metrics
