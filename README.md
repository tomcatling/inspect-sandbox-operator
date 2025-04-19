# InspectSandbox Operator

A Kubernetes operator for managing InspectSandbox environments, replacing the Helm chart-based approach with a custom resource and controller.

## Architecture

The operator follows the Kubernetes operator pattern to manage InspectSandbox resources:

1. **Custom Resource Definition (CRD)**: Defines the InspectSandbox resource type
2. **Controller**: Reconciles the desired state with the actual state
3. **Resources Created**: StatefulSets, Services, PersistentVolumeClaims

## Getting Started

### Prerequisites

- A Kubernetes cluster
- kubectl configured to communicate with your cluster
- go 1.20+ (for development)

### Installation

#### Using Helm

The recommended way to install the operator is using Helm:

```bash
# Add the Helm repository
helm repo add inspect-operator https://tomcatling.github.io/inspect-operator/
helm repo update

# Install the operator
helm install inspect-operator inspect-operator/inspect-sandbox-operator
```

For more information about the Helm chart, see the [Helm chart documentation](./helm-charts/inspect-sandbox-operator/README.md).

#### Manual Installation

1. Install the CRD:

```bash
kubectl apply -f config/crd/bases/inspect.example.com_inspectsandboxes.yaml
```

2. Deploy the operator:

```bash
# Build and deploy the operator (for development)
make deploy

# Or using a pre-built image
kubectl apply -f config/deploy/operator.yaml

# Or using the GitHub Container Registry image
kubectl apply -f config/deploy/operator-ghcr.yaml
```

### Creating an InspectSandbox

```bash
kubectl apply -f examples/inspect_v1alpha1_inspectsandbox.yaml
```

## Development

### Building the operator

```bash
# Clone the repository
git clone https://github.com/tomcatling/inspect-operator.git
cd inspect-operator

# Install dependencies
go mod tidy

# Build
go build -o bin/manager main.go

# Run locally
make run
```

### Running tests

```bash
go test ./... -v
```

### Container Image

The operator container image is automatically built and published to GitHub Container Registry 
on pushes to the main branch and when tags are created.

You can pull the image using:

```bash
docker pull ghcr.io/tomcatling/inspect-operator:latest
```

## Converting from Helm chart

This operator replaces the Helm chart defined in the `agent-env` chart. The key mapping is:

| Helm Chart Value | InspectSandbox CR Field |
|------------------|-------------------------|
| `services` | `spec.services` |
| `allowDomains` | `spec.allowDomains` |
| `networks` | `spec.networks` |
| `volumes` | `spec.volumes` |

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0.