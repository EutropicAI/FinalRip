# FinalRip 微服务启动方式对比：Docker Compose vs Kubernetes

## 当前微服务架构分析

### 现有 Docker Compose 部署方式

#### 1. 快速部署（单机模式）
```bash
# 使用 lite 版本的 docker-compose
docker-compose -f deploy/docker-compose/lite/docker-compose.yml up -d
```

**特点：**
- 所有服务运行在单一主机上
- 简单易用，适合开发和测试
- 资源共享，性能受限于单机

#### 2. 分布式部署（多阶段）
```bash
# 第一步：基础服务
docker-compose -f deploy/docker-compose/docker-compose-base.yml up -d

# 第二步：核心服务
docker-compose -f deploy/docker-compose/docker-compose-server.yml up -d

# 第三步：编码工作节点（可多主机）
docker-compose -f deploy/docker-compose/docker-compose-encode.yml up -d
```

**特点：**
- 支持多主机分布式部署
- 可以根据需要在不同主机部署编码节点
- 手动管理节点，扩展性有限

### 服务组件分析

| 组件 | 功能 | 端口 | 依赖 |
|------|------|------|------|
| **finalrip-server** | API 服务器，任务调度 | 8848 | MongoDB, Redis, Consul |
| **finalrip-dashboard** | Web 前端界面 | 8989(80) | Server |
| **worker-cut** | 视频切割工作节点 | - | MongoDB, Redis, MinIO |
| **worker-merge** | 视频合并工作节点 | - | MongoDB, Redis, MinIO |
| **worker-encode** | 视频编码工作节点（GPU加速）| - | MongoDB, Redis, MinIO |
| **mongodb** | 数据库 | 27017 | - |
| **redis** | 队列系统 | 6379 | - |
| **minio** | 对象存储 | 9000, 9001 | - |
| **consul** | 配置管理 | 8500 | - |
| **asynqmon** | 队列监控 | 8080 | Redis |

## Kubernetes 原生部署方案

### 优势对比

| 特性 | Docker Compose | Kubernetes |
|------|----------------|------------|
| **扩展性** | 手动扩展，有限 | 自动/手动扩展，无限制 |
| **高可用** | 单点故障 | 多副本，故障自恢复 |
| **负载均衡** | 简单轮询 | 智能负载均衡 |
| **服务发现** | 容器名称 | DNS + Service |
| **配置管理** | 环境变量/文件挂载 | ConfigMap + Secret |
| **存储管理** | 本地卷/NFS | 持久卷 + 存储类 |
| **网络隔离** | Docker 网络 | Network Policy |
| **监控** | 外部工具 | 原生监控支持 |
| **CI/CD** | 手动或简单脚本 | 成熟的 GitOps |

### 1. 基础部署（Kustomize）

```bash
# 开发环境
kubectl apply -k deploy/k8s/overlays/development

# 生产环境  
kubectl apply -k deploy/k8s/overlays/production
```

**特点：**
- 声明式配置
- 环境差异化管理
- 版本控制友好

### 2. Helm 部署

```bash
# 默认配置安装
helm install finalrip deploy/k8s/helm/finalrip

# 自定义配置安装
helm install finalrip deploy/k8s/helm/finalrip -f custom-values.yaml

# 升级
helm upgrade finalrip deploy/k8s/helm/finalrip
```

**特点：**
- 参数化配置
- 版本管理
- 依赖管理

### 3. 快速部署脚本

```bash
# 一键部署
./deploy/k8s/deploy.sh development

# 一键清理
./deploy/k8s/undeploy.sh
```

## 核心改进点

### 1. 配置管理改进

**Docker Compose 方式：**
```yaml
environment:
  - FINALRIP_REMOTE_CONFIG_HOST=consul:8500
  - FINALRIP_DB_HOST=mongodb
  - FINALRIP_REDIS_HOST=redis
```

**Kubernetes 方式：**
```yaml
# ConfigMap 集中管理
apiVersion: v1
kind: ConfigMap
metadata:
  name: finalrip-config
data:
  finalrip.yml: |
    server:
      port: 8848
    db:
      host: mongodb
      
# Secret 安全管理
apiVersion: v1
kind: Secret
metadata:
  name: finalrip-secrets
data:
  mongodb-password: MTIzNDU2
```

### 2. 服务发现改进

**Docker Compose：** 依赖容器名称
**Kubernetes：** 使用 Service + DNS

```yaml
apiVersion: v1
kind: Service
metadata:
  name: mongodb
spec:
  selector:
    app.kubernetes.io/name: mongodb
  ports:
  - port: 27017
```

### 3. 存储管理改进

**Docker Compose：** 本地卷挂载
```yaml
volumes:
  - ./data/mongodb:/data/db
```

**Kubernetes：** 持久卷声明
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mongodb-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

### 4. 扩展性改进

**Docker Compose：** 手动扩展
```bash
docker-compose scale worker-encode=3
```

**Kubernetes：** 自动扩展
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: worker-encode-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: finalrip-worker-encode
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

## 迁移路径

### 1. 渐进式迁移

```bash
# 阶段1：部署基础服务到 K8s
kubectl apply -f deploy/k8s/base/mongodb.yaml
kubectl apply -f deploy/k8s/base/redis.yaml
kubectl apply -f deploy/k8s/base/minio.yaml

# 阶段2：迁移应用服务
kubectl apply -f deploy/k8s/base/server.yaml
kubectl apply -f deploy/k8s/base/dashboard.yaml

# 阶段3：迁移工作节点
kubectl apply -f deploy/k8s/base/workers.yaml
```

### 2. 数据迁移

```bash
# 导出 Docker 卷数据
docker run --rm -v finalrip_mongodb:/source -v $(pwd):/backup alpine tar czf /backup/mongodb.tar.gz -C /source .

# 导入到 K8s PV
kubectl cp mongodb.tar.gz finalrip/mongodb-pod:/tmp/
kubectl exec -it mongodb-pod -n finalrip -- tar xzf /tmp/mongodb.tar.gz -C /data/db
```

### 3. 配置迁移

```bash
# 从 Consul 导出配置
docker exec finalrip-consul consul kv export > config-export.json

# 创建 K8s ConfigMap
kubectl create configmap finalrip-config --from-file=finalrip.yml=./conf/finalrip.yml -n finalrip
```

## 运维优势

### 1. 监控和日志

**K8s 原生支持：**
```bash
# 查看所有 Pod 状态
kubectl get pods -n finalrip

# 查看服务日志
kubectl logs -l app.kubernetes.io/name=finalrip-server -n finalrip

# 实时监控资源
kubectl top pods -n finalrip
```

### 2. 滚动更新

```bash
# 零停机更新
kubectl set image deployment/finalrip-server server=lychee0/finalrip:server-v2 -n finalrip

# 回滚
kubectl rollout undo deployment/finalrip-server -n finalrip
```

### 3. 健康检查

```yaml
livenessProbe:
  httpGet:
    path: /ping
    port: 8848
  initialDelaySeconds: 30
  periodSeconds: 10
readinessProbe:
  httpGet:
    path: /ping
    port: 8848
  initialDelaySeconds: 10
  periodSeconds: 5
```

## 成本效益

### 资源利用率
- **Docker Compose：** 静态资源分配
- **Kubernetes：** 动态资源调度，提高利用率

### 维护成本
- **Docker Compose：** 手动运维，容易出错
- **Kubernetes：** 自动化运维，降低人工成本

### 扩展成本
- **Docker Compose：** 需要手动添加节点
- **Kubernetes：** 弹性扩展，按需付费

## 推荐部署策略

### 开发环境
```bash
# 使用 Kustomize 开发覆盖层
kubectl apply -k deploy/k8s/overlays/development
```

### 测试环境
```bash
# 使用 Helm 进行参数化部署
helm install finalrip-test deploy/k8s/helm/finalrip -f test-values.yaml
```

### 生产环境
```bash
# 使用 Helm + GitOps
helm install finalrip-prod deploy/k8s/helm/finalrip -f production-values.yaml
```

## 总结

从 Docker Compose 迁移到 Kubernetes 原生部署，FinalRip 将获得：

1. **更强的扩展性**：支持大规模分布式部署
2. **更高的可用性**：故障自恢复和负载均衡
3. **更好的运维体验**：声明式配置和自动化管理
4. **更强的安全性**：网络隔离和访问控制
5. **更好的可观测性**：原生监控和日志管理

这种转换不仅提供了更现代的部署方式，还为后续的云原生化发展奠定了基础。