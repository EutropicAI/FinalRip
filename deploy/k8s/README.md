# FinalRip Kubernetes Native Deployment

This directory contains Kubernetes-native deployment configurations for FinalRip, a distributed video processing system.

## Architecture Overview

FinalRip consists of the following components:

### Core Services
- **Server**: API server that handles requests and orchestrates tasks
- **Dashboard**: Web frontend for user interaction
- **Worker-Cut**: Cuts videos into clips for parallel processing
- **Worker-Merge**: Merges processed clips back into final video
- **Worker-Encode**: Processes video clips (GPU-accelerated)

### Infrastructure Services
- **MongoDB**: Database for task and metadata storage
- **Redis**: Queue system for task distribution (via Asynq)
- **MinIO**: Object storage for video files
- **Consul**: Configuration management and service discovery
- **AsynqMon**: Queue monitoring interface

## Deployment Options

### 1. Quick Deployment with Kustomize

For development environments:
```bash
# Apply base configuration
kubectl apply -k deploy/k8s/overlays/development

# Or for production
kubectl apply -k deploy/k8s/overlays/production
```

### 2. Helm Chart Deployment

```bash
# Install with default values
helm install finalrip deploy/k8s/helm/finalrip

# Install with custom values
helm install finalrip deploy/k8s/helm/finalrip -f custom-values.yaml

# Upgrade existing deployment
helm upgrade finalrip deploy/k8s/helm/finalrip
```

### 3. Manual Deployment

```bash
# Apply each component individually
kubectl apply -f deploy/k8s/base/namespace.yaml
kubectl apply -f deploy/k8s/base/configmap.yaml
kubectl apply -f deploy/k8s/base/secrets.yaml
kubectl apply -f deploy/k8s/base/mongodb.yaml
kubectl apply -f deploy/k8s/base/redis.yaml
kubectl apply -f deploy/k8s/base/minio.yaml
kubectl apply -f deploy/k8s/base/consul.yaml
kubectl apply -f deploy/k8s/base/asynqmon.yaml
kubectl apply -f deploy/k8s/base/server.yaml
kubectl apply -f deploy/k8s/base/dashboard.yaml
kubectl apply -f deploy/k8s/base/workers.yaml
kubectl apply -f deploy/k8s/base/worker-encode.yaml
kubectl apply -f deploy/k8s/base/ingress.yaml
```

## Prerequisites

### Required Kubernetes Features
- **Persistent Volumes**: For data storage
- **LoadBalancer/Ingress**: For external access
- **GPU Support** (optional): For encode workers
  - NVIDIA Device Plugin
  - GPU nodes with appropriate labels

### Required Resources
- **Minimum**: 4 CPU cores, 8GB RAM, 100GB storage
- **Recommended**: 8 CPU cores, 16GB RAM, 500GB storage
- **GPU**: NVIDIA GPU for video encoding acceleration

## Configuration

### Environment-Specific Configurations

#### Development
- Single replicas for all services
- Reduced resource requirements
- Debug logging enabled
- Local storage

#### Production
- Multiple replicas for high availability
- Increased resource limits
- Production logging levels
- Persistent storage with proper backup
- SSL/TLS termination
- Resource monitoring

### Custom Configuration

1. **Via ConfigMap** (for application config):
   ```yaml
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: finalrip-config
   data:
     finalrip.yml: |
       # Your custom configuration
   ```

2. **Via Environment Variables**:
   ```yaml
   env:
   - name: FINALRIP_DB_HOST
     value: "your-mongodb-host"
   - name: FINALRIP_REDIS_HOST
     value: "your-redis-host"
   ```

3. **Via Helm Values**:
   ```yaml
   # custom-values.yaml
   server:
     replicaCount: 3
   workers:
     encode:
       replicaCount: 5
   ```

## Storage Considerations

### Persistent Volumes
- **MongoDB**: Requires persistent storage for database
- **Redis**: Optional persistence for queue durability
- **MinIO**: Primary storage for video files
- **Consul**: Configuration data persistence

### Storage Classes
Configure appropriate storage classes based on your infrastructure:
- **Fast SSD**: For databases and active processing
- **Object Storage**: For video file storage
- **Network Storage**: For shared access across nodes

## Networking

### Service Discovery
All services use Kubernetes DNS for internal communication:
- `mongodb.finalrip.svc.cluster.local`
- `redis.finalrip.svc.cluster.local`
- `minio.finalrip.svc.cluster.local`
- `consul.finalrip.svc.cluster.local`

### External Access
Configure ingress based on your environment:
- **Development**: NodePort or port-forward
- **Production**: Ingress with SSL/TLS

### Required Ports
- **8848**: API Server
- **80**: Dashboard
- **8080**: Asynq Monitor
- **9000/9001**: MinIO API/Console
- **8500**: Consul UI

## GPU Support

For GPU-accelerated encoding:

1. **Install NVIDIA Device Plugin**:
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/NVIDIA/k8s-device-plugin/main/nvidia-device-plugin.yml
   ```

2. **Label GPU Nodes**:
   ```bash
   kubectl label nodes <gpu-node> accelerator=nvidia-tesla-gpu
   ```

3. **Configure Worker-Encode**:
   ```yaml
   workers:
     encode:
       resources:
         requests:
           nvidia.com/gpu: 1
         limits:
           nvidia.com/gpu: 1
   ```

## Scaling

### Manual Scaling
```bash
# Scale server replicas
kubectl scale deployment finalrip-server --replicas=3 -n finalrip

# Scale encode workers
kubectl scale deployment finalrip-worker-encode --replicas=5 -n finalrip
```

### Horizontal Pod Autoscaler
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: finalrip-server-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: finalrip-server
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Monitoring and Observability

### Built-in Monitoring
- **AsynqMon**: Queue monitoring at `/asynq`
- **Consul UI**: Configuration management at `/consul`
- **MinIO Console**: Storage management at `/minio`

### Additional Monitoring
Consider integrating with:
- **Prometheus**: Metrics collection
- **Grafana**: Visualization
- **ELK Stack**: Log aggregation
- **Jaeger**: Distributed tracing

## Troubleshooting

### Common Issues

1. **Pod Startup Failures**:
   ```bash
   kubectl logs <pod-name> -n finalrip
   kubectl describe pod <pod-name> -n finalrip
   ```

2. **Storage Issues**:
   ```bash
   kubectl get pvc -n finalrip
   kubectl describe pvc <pvc-name> -n finalrip
   ```

3. **Service Connectivity**:
   ```bash
   kubectl get svc -n finalrip
   kubectl exec -it <pod-name> -n finalrip -- nslookup mongodb
   ```

4. **GPU Issues**:
   ```bash
   kubectl get nodes -o yaml | grep nvidia
   kubectl describe node <gpu-node>
   ```

### Health Checks
All services include health checks:
- **Liveness Probes**: Restart unhealthy containers
- **Readiness Probes**: Remove unhealthy pods from service

## Security Considerations

### Secrets Management
- Database passwords
- Object storage credentials
- API tokens

### Network Policies
Implement network policies to restrict inter-pod communication:
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: finalrip-network-policy
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/part-of: finalrip
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app.kubernetes.io/part-of: finalrip
```

### RBAC
Configure Role-Based Access Control for service accounts.

## Migration from Docker Compose

### Key Differences

1. **Service Discovery**: DNS-based instead of container names
2. **Configuration**: ConfigMaps/Secrets instead of environment files
3. **Storage**: Persistent Volumes instead of bind mounts
4. **Networking**: Services and Ingress instead of Docker networks
5. **Scaling**: Native Kubernetes scaling instead of Docker Compose scale

### Migration Steps

1. **Export existing data** from Docker Compose volumes
2. **Create PVCs** and import data
3. **Update configuration** for Kubernetes DNS
4. **Deploy services** using provided manifests
5. **Configure ingress** for external access
6. **Test functionality** and performance

## Performance Tuning

### Resource Allocation
- **CPU**: Adjust based on workload
- **Memory**: Monitor usage and adjust limits
- **GPU**: Optimize for video processing workloads

### Storage Performance
- Use appropriate storage classes
- Consider local storage for temporary processing
- Implement storage tiering for different data types

### Network Optimization
- Use appropriate service types
- Configure ingress for optimal routing
- Consider service mesh for advanced traffic management

## Backup and Recovery

### Database Backup
```bash
# MongoDB backup
kubectl exec -it mongodb-pod -n finalrip -- mongodump --out /backup

# Copy backup from pod
kubectl cp finalrip/mongodb-pod:/backup ./mongodb-backup
```

### Configuration Backup
```bash
# Export current configuration
kubectl get configmap finalrip-config -n finalrip -o yaml > config-backup.yaml
kubectl get secret finalrip-secrets -n finalrip -o yaml > secrets-backup.yaml
```

### Disaster Recovery
1. **Regular backups** of persistent data
2. **Configuration versioning** in Git
3. **Infrastructure as Code** for cluster recreation
4. **Multi-region deployment** for high availability