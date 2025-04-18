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
```

### Creating an InspectSandbox

```bash
kubectl apply -f config/samples/inspect_v1alpha1_inspectsandbox.yaml
```

## Development

### Building the operator

```bash
# Clone the repository
git clone https://github.com/example/inspect-operator.git
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