#!/bin/bash

# Stargazer Build and Deploy Script

set -e

echo "🚀 Building Stargazer..."

# Build Docker image
docker build -t stargazer:latest .

echo "✅ Docker image built successfully"

# Tag for pushing (optional)
if [ ! -z "$REGISTRY" ]; then
    docker tag stargazer:latest $REGISTRY/stargazer:latest
    echo "🏷️  Image tagged for registry: $REGISTRY"
fi

# Deploy to Kubernetes if requested
if [ "$1" = "--deploy" ]; then
    echo "📦 Deploying to Kubernetes..."
    kubectl apply -f kustomization.yaml
    echo "✅ Deployed to Kubernetes"
    echo "💡 Check status with: kubectl get pods -l app=stargazer"
fi

echo "🎉 Build complete!"
echo ""
echo "Usage:"
echo "  Local test:     docker run -it --rm stargazer:latest python standalone.py --demo"
echo "  Deploy to K8s:  ./build.sh --deploy"
echo "  Push registry:  REGISTRY=your-registry.com ./build.sh"