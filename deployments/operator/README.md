# Zot Artifact Store Kubernetes Operator

The Zot Artifact Store Operator manages the deployment and lifecycle of ZotArtifactStore instances on Kubernetes and OpenShift.

## Overview

The operator provides:
- Automated deployment and configuration
- Lifecycle management (upgrades, scaling, etc.)
- Health monitoring and self-healing
- Integration with OpenShift features

## Installation

### Prerequisites

- Kubernetes 1.20+ or OpenShift 4.10+
- kubectl or oc CLI
- Cluster admin permissions (for CRD installation)

### Install the CRD

```bash
kubectl apply -f config/crd/zotartifactstore_crd.yaml
```

### Verify CRD Installation

```bash
kubectl get crd zotartifactstores.artifacts.zot.io
```

## Usage

### Create a ZotArtifactStore Instance

#### Minimal Configuration

```bash
kubectl apply -f config/samples/zotartifactstore_minimal.yaml
```

#### Full Configuration

```bash
kubectl apply -f config/samples/zotartifactstore_sample.yaml
```

### Check Status

```bash
kubectl get zotartifactstore
kubectl describe zotartifactstore zot-artifact-store-sample
```

### Access the Service

```bash
# Get the service endpoint
kubectl get zotartifactstore zot-artifact-store-sample -o jsonpath='{.status.serviceEndpoint}'

# Port forward for local access
kubectl port-forward svc/zot-artifact-store-sample 8080:8080
```

## Configuration Options

### Image Configuration

```yaml
spec:
  image:
    repository: quay.io/zot-artifact-store
    tag: "latest"
    pullPolicy: IfNotPresent
    pullSecrets:
      - my-registry-secret
```

### Storage Options

#### Filesystem Storage (Default)

```yaml
spec:
  storage:
    type: filesystem
    size: 100Gi
    storageClassName: standard
```

#### S3 Storage

```yaml
spec:
  storage:
    type: s3
    s3:
      endpoint: s3.amazonaws.com
      region: us-east-1
      bucket: my-artifacts
      accessKeySecret: aws-credentials
      secretKeySecret: aws-credentials
```

#### Azure Blob Storage

```yaml
spec:
  storage:
    type: azure
    azure:
      accountName: myaccount
      container: artifacts
      credentialsSecret: azure-credentials
```

#### Google Cloud Storage

```yaml
spec:
  storage:
    type: gcp
    gcp:
      bucket: my-artifacts
      credentialsSecret: gcp-credentials
```

### Extension Configuration

#### Enable/Disable Extensions

```yaml
spec:
  extensions:
    s3api:
      enabled: true
      maxUploadSize: 10Gi
    rbac:
      enabled: true
      keycloak:
        url: https://keycloak.example.com
        realm: myrealm
        clientId: zot
        clientSecretName: keycloak-secret
    supplychain:
      enabled: true
      signing:
        providers: [cosign, notary]
      sbom:
        formats: [spdx, cyclonedx]
    metrics:
      enabled: true
```

### Resource Limits

```yaml
spec:
  resources:
    limits:
      cpu: 2000m
      memory: 4Gi
    requests:
      cpu: 1000m
      memory: 2Gi
```

## Operator Implementation

The operator will be implemented in Phase 12 using:
- Operator SDK
- Controller Runtime
- Reconciliation loop for desired state management

## Status Fields

The operator updates the status to reflect the current state:

```yaml
status:
  phase: Running
  replicas: 2
  readyReplicas: 2
  serviceEndpoint: http://zot-artifact-store-sample.zot-system.svc:8080
  conditions:
    - type: Ready
      status: "True"
      lastTransitionTime: "2024-01-01T00:00:00Z"
      reason: AllPodsReady
      message: All pods are ready
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -l app=zot-artifact-store
kubectl logs -l app=zot-artifact-store
```

### Check Events

```bash
kubectl get events --field-selector involvedObject.kind=ZotArtifactStore
```

### Common Issues

1. **Pods not starting**: Check resource availability and storage configuration
2. **Service not accessible**: Verify network policies and routes
3. **Extensions failing**: Check configuration and secrets

## Development

The operator will be developed in Phase 12. For now, manual deployment using the CRD is supported.

## Future Enhancements

- Automated backups
- Blue-green deployments
- Auto-scaling based on metrics
- Multi-cluster support
