#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

usage() {
  echo "Usage: $0 <environment> <version>"
  echo "  environment: dev | staging | prod"
  echo "  version:     image tag (e.g. sha, semver)"
  exit 1
}

[ $# -eq 2 ] || usage

ENVIRONMENT="$1"
VERSION="$2"
NAMESPACE="quantify"
REGISTRY="${REGISTRY:-ghcr.io}"
IMAGE_PREFIX="${IMAGE_PREFIX:-quantify}"

case "$ENVIRONMENT" in
  dev)     CONTEXT="dev" ;;
  staging) CONTEXT="staging" ;;
  prod)    CONTEXT="prod" ;;
  *)
    echo "Error: invalid environment '$ENVIRONMENT' (expected dev|staging|prod)" >&2
    exit 1
    ;;
esac

echo "=== Deploying to ${ENVIRONMENT} (version: ${VERSION}) ==="

echo "Setting kubectl context..."
kubectl config use-context "$CONTEXT"

echo "Running migrations..."
kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: quantify-migrate-${VERSION//[^a-z0-9-]/-}
  namespace: ${NAMESPACE}
  labels:
    app.kubernetes.io/part-of: quantify
    app.kubernetes.io/component: migration
spec:
  backoffLimit: 3
  ttlSecondsAfterFinished: 600
  template:
    spec:
      restartPolicy: Never
      containers:
        - name: migrate
          image: ${REGISTRY}/${IMAGE_PREFIX}-backend:${VERSION}
          command: ["/app/inventory-backend", "-migrate"]
          envFrom:
            - configMapRef:
                name: quantify-backend
            - secretRef:
                name: quantify-backend
EOF

echo "Waiting for migration job..."
kubectl wait --for=condition=complete \
  "job/quantify-migrate-${VERSION//[^a-z0-9-]/-}" \
  -n "$NAMESPACE" --timeout=300s

echo "Deploying Helm charts..."
helm upgrade --install quantify-backend "$ROOT_DIR/deploy/helm/quantify-backend" \
  --namespace "$NAMESPACE" \
  --values "$ROOT_DIR/deploy/helm/quantify-backend/values-${ENVIRONMENT}.yaml" \
  --set image.repository="${REGISTRY}/${IMAGE_PREFIX}-backend" \
  --set image.tag="$VERSION" \
  --wait --timeout 10m

helm upgrade --install quantify-ai-service "$ROOT_DIR/deploy/helm/quantify-ai-service" \
  --namespace "$NAMESPACE" \
  --values "$ROOT_DIR/deploy/helm/quantify-ai-service/values-${ENVIRONMENT}.yaml" \
  --set image.repository="${REGISTRY}/${IMAGE_PREFIX}-ai-service" \
  --set image.tag="$VERSION" \
  --wait --timeout 10m

echo "Waiting for rollout..."
kubectl rollout status deployment/quantify-backend -n "$NAMESPACE" --timeout=300s
kubectl rollout status deployment/quantify-ai-service -n "$NAMESPACE" --timeout=300s

echo "Running smoke tests..."
BASE_URL="${SMOKE_TEST_URL:-http://localhost:8080}"
"$SCRIPT_DIR/smoke-test.sh" "$BASE_URL"

echo "=== Deploy to ${ENVIRONMENT} complete ==="
