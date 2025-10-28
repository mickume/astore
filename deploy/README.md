# Zot Artifact Store Kubernetes Operator

The Zot Artifact Store Operator manages the lifecycle of Zot Artifact Store deployments on Kubernetes and OpenShift.

## Features

- **Automated Deployment**: Declarative configuration for artifact store instances
- **High Availability**: Built-in support for multi-replica deployments
- **Storage Flexibility**: Support for filesystem, S3, GCS, and Azure Blob storage
- **Database Options**: Embedded BoltDB, PostgreSQL, or MySQL
- **RBAC Integration**: Keycloak integration for authentication and authorization
- **Supply Chain Security**: Built-in support for artifact signing and SBOM management
- **Observability**: Prometheus metrics, health checks, and distributed tracing
- **Auto-scaling**: Horizontal Pod Autoscaler integration
- **High Availability**: Pod Disruption Budgets and anti-affinity rules
- **OpenShift Support**: Native support for OpenShift Routes and SCCs

## Prerequisites

- Kubernetes 1.20+ or OpenShift 4.8+
- kubectl or oc CLI tool
- Cluster admin privileges (for CRD installation)

## Installation

### Quick Install

```bash
# Install CRD
kubectl apply -f deploy/crds/zotartifactstore-crd.yaml

# Install operator
kubectl apply -f deploy/operator/rbac.yaml
kubectl apply -f deploy/operator/deployment.yaml
```

### Verify Installation

```bash
# Check CRD
kubectl get crd zotartifactstores.artifact.zotregistry.io

# Check operator pod
kubectl get pods -n zot-operator-system

# Check operator logs
kubectl logs -n zot-operator-system -l app=zot-artifact-store-operator
```

## Usage

### Minimal Deployment

Create a basic artifact store with default settings:

```bash
kubectl apply -f deploy/examples/minimal.yaml
```

This creates:
- Single replica deployment
- Filesystem storage (5Gi PVC)
- Embedded BoltDB database
- Basic metrics enabled

### Production Deployment

Deploy a production-ready artifact store with all features:

```bash
# Create namespace
kubectl create namespace artifact-store

# Create secrets (customize these!)
kubectl create secret generic s3-credentials \
  --from-literal=access-key=YOUR_ACCESS_KEY \
  --from-literal=secret-key=YOUR_SECRET_KEY \
  -n artifact-store

kubectl create secret generic postgres-credentials \
  --from-literal=username=postgres \
  --from-literal=password=YOUR_PASSWORD \
  -n artifact-store

kubectl create secret generic keycloak-credentials \
  --from-literal=client-secret=YOUR_CLIENT_SECRET \
  -n artifact-store

# Deploy artifact store
kubectl apply -f deploy/examples/production.yaml
```

### OpenShift Deployment

```bash
# Create project
oc new-project artifact-store

# Create secrets
oc create secret generic keycloak-client-secret \
  --from-literal=secret=YOUR_CLIENT_SECRET

# Deploy
oc apply -f deploy/examples/openshift.yaml

# Get route
oc get route -n artifact-store
```

## Configuration Reference

### Storage Configuration

#### Filesystem Storage

```yaml
spec:
  storage:
    type: filesystem
    size: 50Gi
    storageClass: fast-ssd  # Optional
```

#### S3-Compatible Storage

```yaml
spec:
  storage:
    type: s3
    s3:
      endpoint: https://s3.amazonaws.com
      bucket: my-artifacts
      region: us-east-1
      accessKeySecret:
        name: s3-credentials
        accessKeyField: access-key
        secretKeyField: secret-key
```

#### Google Cloud Storage

```yaml
spec:
  storage:
    type: gcs
    gcs:
      bucket: my-artifacts
      credentialsSecret: gcs-service-account
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

### Database Configuration

#### Embedded BoltDB (Default)

```yaml
spec:
  database:
    type: embedded
    embedded:
      storageSize: 5Gi
```

#### PostgreSQL

```yaml
spec:
  database:
    type: postgres
    postgres:
      host: postgres.database.svc.cluster.local
      port: 5432
      database: artifactstore
      credentialsSecret: postgres-credentials
```

#### MySQL

```yaml
spec:
  database:
    type: mysql
    mysql:
      host: mysql.database.svc.cluster.local
      port: 3306
      database: artifactstore
      credentialsSecret: mysql-credentials
```

### RBAC and Authentication

```yaml
spec:
  rbac:
    enabled: true
    keycloak:
      url: https://keycloak.example.com
      realm: artifacts
      clientId: zot-artifact-store
      clientSecretRef:
        name: keycloak-credentials
        key: client-secret
```

### Supply Chain Security

```yaml
spec:
  supplyChain:
    enabled: true
    signing:
      enabled: true
      keySize: 4096
      keySecretRef:
        name: signing-keys
    sbom:
      enabled: true
      defaultFormat: spdx  # or cyclonedx
```

### Networking

#### Service Configuration

```yaml
spec:
  networking:
    service:
      type: LoadBalancer  # ClusterIP, NodePort, or LoadBalancer
      port: 443
      annotations:
        service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
```

#### Ingress Configuration

```yaml
spec:
  networking:
    ingress:
      enabled: true
      className: nginx
      host: artifacts.example.com
      tls:
        enabled: true
        secretName: tls-certificate
      annotations:
        cert-manager.io/cluster-issuer: letsencrypt-prod
```

### Auto-scaling

```yaml
spec:
  autoscaling:
    enabled: true
    minReplicas: 3
    maxReplicas: 10
    targetCPUUtilization: 70
    targetMemoryUtilization: 80
```

### High Availability

```yaml
spec:
  replicas: 3

  podDisruptionBudget:
    enabled: true
    minAvailable: 2

  affinity:
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        - labelSelector:
            matchLabels:
              app: zot-artifact-store
          topologyKey: kubernetes.io/hostname
```

## Monitoring

### Check Status

```bash
# Get artifact store status
kubectl get zotartifactstore -n artifact-store

# Detailed status
kubectl describe zotartifactstore production-artifact-store -n artifact-store
```

### View Metrics

```bash
# Port-forward to metrics endpoint
kubectl port-forward -n artifact-store svc/production-artifact-store-metrics 8081:8081

# Access metrics
curl http://localhost:8081/metrics
```

### View Logs

```bash
# Application logs
kubectl logs -n artifact-store -l app=zot-artifact-store --tail=100 -f

# Operator logs
kubectl logs -n zot-operator-system -l app=zot-artifact-store-operator --tail=100 -f
```

## Upgrading

### Update Image Version

```bash
kubectl patch zotartifactstore production-artifact-store \
  -n artifact-store \
  --type=merge \
  -p '{"spec":{"image":{"tag":"v1.1.0"}}}'
```

### Update Configuration

```bash
# Edit resource
kubectl edit zotartifactstore production-artifact-store -n artifact-store

# Or apply updated YAML
kubectl apply -f deploy/examples/production.yaml
```

### Rolling Update

The operator automatically performs rolling updates when configuration changes.

## Backup and Restore

### Backup

```bash
# Backup CRD instance
kubectl get zotartifactstore production-artifact-store -n artifact-store -o yaml > backup.yaml

# Backup data (filesystem storage)
kubectl exec -n artifact-store production-artifact-store-0 -- tar czf - /var/lib/zot > backup.tar.gz

# Backup data (S3 storage)
aws s3 sync s3://my-artifacts s3://my-artifacts-backup
```

### Restore

```bash
# Restore CRD instance
kubectl apply -f backup.yaml

# Restore data (filesystem storage)
kubectl exec -n artifact-store production-artifact-store-0 -- tar xzf - -C / < backup.tar.gz
```

## Troubleshooting

### Operator Issues

```bash
# Check operator logs
kubectl logs -n zot-operator-system -l app=zot-artifact-store-operator

# Check operator events
kubectl get events -n zot-operator-system --sort-by='.lastTimestamp'

# Restart operator
kubectl rollout restart deployment/zot-artifact-store-operator -n zot-operator-system
```

### Application Issues

```bash
# Check pod status
kubectl get pods -n artifact-store

# Check pod logs
kubectl logs -n artifact-store -l app=zot-artifact-store

# Check pod events
kubectl describe pod -n artifact-store production-artifact-store-0

# Debug pod
kubectl exec -it -n artifact-store production-artifact-store-0 -- /bin/sh
```

### Storage Issues

```bash
# Check PVC status
kubectl get pvc -n artifact-store

# Check PV details
kubectl describe pv $(kubectl get pvc -n artifact-store -o jsonpath='{.items[0].spec.volumeName}')

# Check storage class
kubectl get storageclass
```

### Network Issues

```bash
# Check service
kubectl get svc -n artifact-store

# Check endpoints
kubectl get endpoints -n artifact-store

# Test connectivity
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://production-artifact-store.artifact-store.svc.cluster.local:8080/health
```

## Uninstallation

```bash
# Delete artifact store instances
kubectl delete zotartifactstore --all --all-namespaces

# Delete operator
kubectl delete -f deploy/operator/deployment.yaml
kubectl delete -f deploy/operator/rbac.yaml

# Delete CRD (this removes all instances!)
kubectl delete -f deploy/crds/zotartifactstore-crd.yaml
```

## Advanced Usage

### Custom Resource Scaling

```bash
# Scale up
kubectl scale zotartifactstore production-artifact-store --replicas=5 -n artifact-store

# Scale down
kubectl scale zotartifactstore production-artifact-store --replicas=2 -n artifact-store
```

### Resource Patching

```bash
# Update resources
kubectl patch zotartifactstore production-artifact-store \
  -n artifact-store \
  --type=merge \
  -p '{"spec":{"resources":{"limits":{"memory":"2Gi"}}}}'
```

### Label Management

```bash
# Add labels
kubectl label zotartifactstore production-artifact-store \
  -n artifact-store \
  environment=production

# Remove labels
kubectl label zotartifactstore production-artifact-store \
  -n artifact-store \
  environment-
```

## Best Practices

1. **Use Namespaces**: Deploy artifact stores in dedicated namespaces
2. **Enable RBAC**: Always enable RBAC in production
3. **Use External Storage**: Use S3/GCS/Azure for production deployments
4. **Enable Monitoring**: Configure ServiceMonitor for Prometheus
5. **Set Resource Limits**: Always set appropriate resource requests and limits
6. **Use Auto-scaling**: Enable HPA for variable workloads
7. **Configure PDB**: Prevent disruption during maintenance
8. **Use Anti-affinity**: Spread replicas across nodes
9. **Enable TLS**: Use TLS for all external communication
10. **Regular Backups**: Implement automated backup strategy

## Security Considerations

1. **Run as Non-root**: Operator and workloads run as non-root by default
2. **Security Contexts**: Restrictive security contexts are enforced
3. **Secret Management**: Use Kubernetes secrets for sensitive data
4. **Network Policies**: Implement network policies to restrict traffic
5. **RBAC**: Use least-privilege RBAC roles
6. **Pod Security**: Compatible with Pod Security Standards (restricted)
7. **Image Scanning**: Scan container images for vulnerabilities
8. **Audit Logging**: Enable Kubernetes audit logging

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for development guidelines.

## License

Apache 2.0 - See [LICENSE](../LICENSE) for details.
