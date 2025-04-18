# Migrating from Helm Chart to InspectSandbox CRD

This guide explains how to migrate from the `agent-env` Helm chart to the new InspectSandbox CRD.

## Helm Chart vs. InspectSandbox CRD

The InspectSandbox CRD offers several advantages over the Helm chart:

- **Declarative Updates**: Change only what you need without redeploying the entire chart
- **Reconciliation**: Automatic drift detection and correction
- **Status**: View status of your resources directly through the Kubernetes API
- **Integration**: Better integration with other Kubernetes tools and operators

## Migration Steps

1. **Extract Helm Values**: Get your current values file from the Helm release
```bash
helm get values <release-name> -n <namespace> > values.yaml
```

2. **Convert to InspectSandbox CR**: Create a new YAML file for your InspectSandbox resource
```yaml
apiVersion: inspect.example.com/v1alpha1
kind: InspectSandbox
metadata:
  name: <name>  # Previously the Helm release name
spec:
  # Copy from your Helm values
  allowDomains: []  # From allowDomains in values.yaml
  networks: {}      # From networks in values.yaml
  services: {}      # From services in values.yaml
  volumes: {}       # From volumes in values.yaml
```

3. **Apply the InspectSandbox CR**
```bash
kubectl apply -f inspect-sandbox.yaml
```

4. **Verify the Resources**
```bash
kubectl get inspectsandbox <name> -o yaml
kubectl get statefulsets -l app.kubernetes.io/instance=<name>
```

5. **Delete the Helm Release** (only after verification)
```bash
helm delete <release-name> -n <namespace>
```

## Example Conversion

### Helm Values
```yaml
allowDomains:
  - "pypi.org"
  - "files.pythonhosted.org"
services:
  default:
    runtimeClassName: gvisor
    image: "python:3.12-bookworm"
    command: ["tail", "-f", "/dev/null"]
    dnsRecord: true
volumes:
  shared-volume:
    size: 5Gi
```

### Equivalent InspectSandbox CR
```yaml
apiVersion: inspect.example.com/v1alpha1
kind: InspectSandbox
metadata:
  name: my-sandbox
spec:
  allowDomains:
    - "pypi.org"
    - "files.pythonhosted.org"
  services:
    default:
      runtimeClassName: gvisor
      image: "python:3.12-bookworm"
      command: ["tail", "-f", "/dev/null"]
      dnsRecord: true
  volumes:
    shared-volume:
      size: 5Gi
```

## Troubleshooting

If you encounter issues during migration:

1. Check the operator logs:
```bash
kubectl logs -n inspect-system deploy/inspect-operator
```

2. Describe the InspectSandbox resource:
```bash
kubectl describe inspectsandbox <name>
```

3. Verify the status of the created resources:
```bash
kubectl get statefulsets,services,pvc -l app.kubernetes.io/instance=<name>
```