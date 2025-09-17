#!/bin/bash

# Uninstall script for FinalRip Kubernetes deployment
# Usage: ./undeploy.sh

set -e

NAMESPACE="finalrip"

echo "🗑️ Removing FinalRip from Kubernetes..."

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed or not in PATH"
    exit 1
fi

# Check if namespace exists
if ! kubectl get namespace $NAMESPACE &> /dev/null; then
    echo "ℹ️ Namespace $NAMESPACE does not exist, nothing to remove"
    exit 0
fi

echo "🔧 Removing application services..."
kubectl delete -f deploy/k8s/base/ingress.yaml --ignore-not-found=true
kubectl delete -f deploy/k8s/base/worker-encode.yaml --ignore-not-found=true
kubectl delete -f deploy/k8s/base/workers.yaml --ignore-not-found=true
kubectl delete -f deploy/k8s/base/dashboard.yaml --ignore-not-found=true
kubectl delete -f deploy/k8s/base/server.yaml --ignore-not-found=true
kubectl delete -f deploy/k8s/base/asynqmon.yaml --ignore-not-found=true

echo "🗄️ Removing data services..."
kubectl delete -f deploy/k8s/base/consul.yaml --ignore-not-found=true
kubectl delete -f deploy/k8s/base/minio.yaml --ignore-not-found=true
kubectl delete -f deploy/k8s/base/redis.yaml --ignore-not-found=true
kubectl delete -f deploy/k8s/base/mongodb.yaml --ignore-not-found=true

echo "📦 Removing configuration..."
kubectl delete -f deploy/k8s/base/secrets.yaml --ignore-not-found=true
kubectl delete -f deploy/k8s/base/configmap.yaml --ignore-not-found=true

echo "⚠️ Persistent data will be preserved. To remove data volumes:"
echo "kubectl delete pvc --all -n $NAMESPACE"
echo ""

read -p "Do you want to remove persistent data as well? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🗄️ Removing persistent volumes..."
    kubectl delete pvc --all -n $NAMESPACE --ignore-not-found=true
fi

echo "🗑️ Removing namespace..."
kubectl delete namespace $NAMESPACE --ignore-not-found=true

echo "✅ FinalRip has been removed from Kubernetes"