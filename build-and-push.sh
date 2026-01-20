#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="${SCRIPT_DIR}/../talos-deploy"
DEPLOYMENT_FILE="${DEPLOY_DIR}/manifests/stargazer/deployment.yaml"

# Generate beta version with timestamp
TIMESTAMP=$(date +%Y%m%d%H%M%S)
BETA_VERSION="v0.1.0-beta.${TIMESTAMP}"
LATEST_BETA="v0.1.0-beta-latest"
IMAGE_NAME="maplecitymadman/stargazer:${BETA_VERSION}"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Stargazer Build and Deploy${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Step 1: Build Go binary
echo -e "${BLUE}Step 1: Building Go binary for linux/amd64...${NC}"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/stargazer-linux-amd64 cmd/stargazer/main.go
echo -e "${GREEN}✅ Binary built${NC}"
echo ""

# Step 2: Build and push Docker image
echo -e "${BLUE}Step 2: Building and pushing Docker image...${NC}"
echo "Tags:"
echo "  - ${IMAGE_NAME}"
echo "  - maplecitymadman/stargazer:${LATEST_BETA}"
echo ""

docker buildx build --platform linux/amd64 \
  --provenance=false \
  --sbom=false \
  -t ${IMAGE_NAME} \
  -t maplecitymadman/stargazer:${LATEST_BETA} \
  --push .

echo ""
echo -e "${GREEN}✅ Image pushed successfully${NC}"
echo ""

# Step 3: Update deployment YAML file
echo -e "${BLUE}Step 3: Updating deployment YAML file...${NC}"
if [ ! -f "${DEPLOYMENT_FILE}" ]; then
    echo -e "${YELLOW}⚠️  Warning: Deployment file not found at ${DEPLOYMENT_FILE}${NC}"
    echo "   Falling back to kubectl set image..."
    kubectl set image deployment/stargazer stargazer=${IMAGE_NAME} -n stargazer
else
    # Update the image in the deployment YAML
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s|image: maplecitymadman/stargazer:.*|image: ${IMAGE_NAME}|" "${DEPLOYMENT_FILE}"
    else
        # Linux
        sed -i "s|image: maplecitymadman/stargazer:.*|image: ${IMAGE_NAME}|" "${DEPLOYMENT_FILE}"
    fi
    echo -e "${GREEN}✅ Deployment YAML updated to: ${IMAGE_NAME}${NC}"
    
    # Apply the updated deployment
    echo "Applying updated deployment..."
    kubectl apply -f "${DEPLOYMENT_FILE}"
    echo -e "${GREEN}✅ Deployment applied${NC}"
fi
echo ""

# Step 4: Wait for pod to be healthy
echo -e "${BLUE}Step 4: Waiting for pod to be healthy...${NC}"
echo "This may take a minute..."
kubectl rollout status deployment/stargazer -n stargazer --timeout=5m

# Check if rollout was successful
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Pod is healthy!${NC}"
else
    echo -e "${YELLOW}⚠️  Rollout status check timed out or failed${NC}"
    echo "Checking pod status manually..."
    kubectl get pods -n stargazer -l app=stargazer
    echo ""
    echo "If the pod is not ready, check logs with:"
    echo "  kubectl logs -n stargazer -l app=stargazer"
fi
echo ""

# Step 5: Port forward the service
echo -e "${BLUE}Step 5: Setting up port forward...${NC}"
echo -e "${GREEN}✅ Port forwarding stargazer service on port 8000${NC}"
echo ""
echo "Access stargazer at: http://localhost:8000"
echo ""
echo "Press Ctrl+C to stop port forwarding"
echo ""

# Port forward in the background and capture PID
kubectl port-forward -n stargazer svc/stargazer 8000:8000 &
PF_PID=$!

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "Stopping port forward..."
    kill $PF_PID 2>/dev/null || true
    exit 0
}

trap cleanup SIGINT SIGTERM

# Wait for port forward to be ready
sleep 2

# Check if port forward is still running
if ps -p $PF_PID > /dev/null; then
    echo -e "${GREEN}✅ Port forward active (PID: $PF_PID)${NC}"
    echo ""
    wait $PF_PID
else
    echo -e "${YELLOW}⚠️  Port forward failed to start${NC}"
    exit 1
fi
