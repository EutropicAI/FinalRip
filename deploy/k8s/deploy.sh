#!/bin/bash

# Quick deployment script for FinalRip on Kubernetes
# Usage: ./deploy.sh [development|production]

set -e

ENVIRONMENT=${1:-development}
NAMESPACE="finalrip"

echo "🚀 Deploying FinalRip to Kubernetes in $ENVIRONMENT mode..."

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed or not in PATH"
    exit 1
fi

# Check if cluster is accessible
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Cannot connect to Kubernetes cluster"
    exit 1
fi

echo "✅ Kubernetes cluster is accessible"

# Create namespace if it doesn't exist
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

echo "📦 Deploying infrastructure services..."

# Deploy infrastructure first
kubectl apply -f deploy/k8s/base/namespace.yaml
kubectl apply -f deploy/k8s/base/configmap.yaml
kubectl apply -f deploy/k8s/base/secrets.yaml

echo "🗄️ Deploying data services..."
kubectl apply -f deploy/k8s/base/mongodb.yaml
kubectl apply -f deploy/k8s/base/redis.yaml
kubectl apply -f deploy/k8s/base/minio.yaml
kubectl apply -f deploy/k8s/base/consul.yaml

echo "⏳ Waiting for data services to be ready..."
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=mongodb -n $NAMESPACE --timeout=300s
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=redis -n $NAMESPACE --timeout=300s
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=minio -n $NAMESPACE --timeout=300s
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=consul -n $NAMESPACE --timeout=300s

echo "🔧 Deploying application services..."
kubectl apply -f deploy/k8s/base/asynqmon.yaml
kubectl apply -f deploy/k8s/base/server.yaml
kubectl apply -f deploy/k8s/base/dashboard.yaml
kubectl apply -f deploy/k8s/base/workers.yaml

if [ "$ENVIRONMENT" = "production" ]; then
    echo "🎬 Deploying GPU-enabled encode workers..."
    kubectl apply -f deploy/k8s/base/worker-encode.yaml
fi

echo "🌐 Configuring network access..."
if [ "$ENVIRONMENT" = "development" ]; then
    kubectl apply -f deploy/k8s/base/ingress.yaml
    
    # Port forward for local access
    echo "🔗 Setting up port forwarding for local access..."
    echo "Dashboard will be available at: http://localhost:8989"
    echo "API will be available at: http://localhost:8848"
    echo "Asynq Monitor will be available at: http://localhost:8080"
    echo ""
    echo "To access services, run:"
    echo "kubectl port-forward svc/finalrip-dashboard 8989:80 -n $NAMESPACE &"
    echo "kubectl port-forward svc/finalrip-server 8848:8848 -n $NAMESPACE &"
    echo "kubectl port-forward svc/asynqmon 8080:8080 -n $NAMESPACE &"
fi

echo "⏳ Waiting for application services to be ready..."
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=finalrip-server -n $NAMESPACE --timeout=300s
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=finalrip-dashboard -n $NAMESPACE --timeout=300s

echo "✅ FinalRip deployment complete!"
echo ""
echo "📊 Deployment status:"
kubectl get pods -n $NAMESPACE

echo ""
echo "🔧 To check service status:"
echo "kubectl get svc -n $NAMESPACE"
echo ""
echo "📝 To view logs:"
echo "kubectl logs -l app.kubernetes.io/name=finalrip-server -n $NAMESPACE"
echo ""
echo "🗑️ To uninstall:"
echo "./undeploy.sh"