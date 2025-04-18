# Helm Test Suite for Inspect Sandbox Operator

This directory contains Helm test manifests that validate the functionality of the Inspect Sandbox Operator.

## Test Overview

### 1. Operator Deployment Test (`test-operator-deployment.yaml`)

This test verifies that:
- The operator pods are running and ready
- The CRD has been installed correctly

### 2. CRD Functionality Test (`test-crd-functionality.yaml`)

This test verifies basic CRD functionality:
- Creates a simple InspectSandbox resource
- Verifies that StatefulSets and Services are created
- Checks for network policies if Cilium is available
- Validates status updates by the operator
- Cleans up after itself

### 3. Integration Test (`test-integration.yaml`)

This is a comprehensive test that verifies:
- Creation of multiple services within a sandbox
- Network configuration between services
- Volume provisioning
- Labels and network policies
- Status updates for all components

## Running the Tests

The tests can be run with:

```bash
helm test <release-name>
```

## Test Implementation Details

Each test is implemented as a Kubernetes Pod with the `helm.sh/hook: test` annotation. The tests use the built-in kubectl client to validate the operator's behavior.

Tests have a `helm.sh/hook-weight` to control execution order:
- Operator deployment test: default weight (0)
- CRD functionality test: weight 5
- Integration test: weight 10

This ensures tests run in order from basic validation to comprehensive testing.

## Extending the Tests

When adding new features to the operator, consider updating the integration test to validate the new functionality. The test framework follows these principles:

1. **Isolation**: Each test creates its own resources with unique names
2. **Cleanup**: All tests clean up after themselves
3. **Progressive complexity**: Tests are ordered from simple to complex

## CI/CD Integration

These tests are integrated with GitHub Actions in `.github/workflows/test.yml` to provide automated testing on each PR and commit.