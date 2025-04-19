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

## Testing

This chart includes Helm tests that verify the installation and functionality of the operator. To run the tests:

```bash
helm test inspect-operator
```

The test suite includes:

1. **Operator Deployment Test**: Verifies that the operator is deployed and running
2. **CRD Functionality Test**: Tests the basic functionality of the InspectSandbox CRD
3. **Integration Test**: A comprehensive test that verifies multiple services, networks, and volumes

These tests can be used for:
- Verifying a successful installation
- Running integration tests in CI/CD pipelines
- Validating operator functionality after upgrades

For CI/CD integration, see the GitHub workflow in `.github/workflows/test.yml` for an example of how to run these tests in a pipeline.

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