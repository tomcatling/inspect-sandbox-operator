# Inspect Sandbox Operator

A Helm chart for deploying the Inspect Sandbox Kubernetes Operator.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+

## Installing the Chart

To install the chart with the release name `inspect-operator`:

```bash
helm install inspect-operator ./helm-charts/inspect-sandbox-operator
```

## Uninstalling the Chart

To uninstall/delete the `inspect-operator` deployment:

```bash
helm delete inspect-operator
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| namespace.create | bool | `true` | Create the namespace |
| namespace.name | string | `"inspect-system"` | Namespace name |
| deployment.replicas | int | `1` | Number of replicas |
| deployment.image.repository | string | `"ghcr.io/tomcatling/inspect-sandbox-operator"` | Image repository |
| deployment.image.tag | string | `"latest"` | Image tag |
| deployment.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| deployment.resources.limits.cpu | string | `"100m"` | CPU limit |
| deployment.resources.limits.memory | string | `"256Mi"` | Memory limit |
| deployment.resources.requests.cpu | string | `"100m"` | CPU request |
| deployment.resources.requests.memory | string | `"128Mi"` | Memory request |
| rbac.create | bool | `true` | Create RBAC resources |
| serviceAccount.create | bool | `true` | Create ServiceAccount |
| serviceAccount.name | string | `"inspect-operator"` | ServiceAccount name |
| crds.install | bool | `true` | Install CRDs |