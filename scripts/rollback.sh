#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: $0 <environment>"
  echo "  environment: dev | staging | prod"
  exit 1
}

[ $# -eq 1 ] || usage

ENVIRONMENT="$1"
NAMESPACE="quantify"

case "$ENVIRONMENT" in
  dev)     CONTEXT="dev" ;;
  staging) CONTEXT="staging" ;;
  prod)    CONTEXT="prod" ;;
  *)
    echo "Error: invalid environment '$ENVIRONMENT' (expected dev|staging|prod)" >&2
    exit 1
    ;;
esac

echo "=== Rolling back ${ENVIRONMENT} ==="

kubectl config use-context "$CONTEXT"

PREV_REVISION=$(helm history quantify-backend -n "$NAMESPACE" -o json \
  | jq -r '[.[] | select(.status == "deployed")] | sort_by(.revision) | .[-2].revision // empty')

if [ -z "$PREV_REVISION" ]; then
  echo "Error: no previous deployed revision found" >&2
  exit 1
fi

echo "Rolling back quantify-backend to revision ${PREV_REVISION}..."
helm rollback quantify-backend "$PREV_REVISION" -n "$NAMESPACE" --wait --timeout 5m

PREV_AI_REVISION=$(helm history quantify-ai-service -n "$NAMESPACE" -o json \
  | jq -r '[.[] | select(.status == "deployed")] | sort_by(.revision) | .[-2].revision // empty')

if [ -n "$PREV_AI_REVISION" ]; then
  echo "Rolling back quantify-ai-service to revision ${PREV_AI_REVISION}..."
  helm rollback quantify-ai-service "$PREV_AI_REVISION" -n "$NAMESPACE" --wait --timeout 5m
fi

echo "Waiting for rollout..."
kubectl rollout status deployment/quantify-backend -n "$NAMESPACE" --timeout=300s
kubectl rollout status deployment/quantify-ai-service -n "$NAMESPACE" --timeout=300s

echo "Verifying health..."
for i in $(seq 1 10); do
  if kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/name=quantify-backend \
    -o jsonpath='{.items[*].status.conditions[?(@.type=="Ready")].status}' | grep -q "True"; then
    echo "Backend pods are ready"
    break
  fi
  echo "Waiting for pods... ($i/10)"
  sleep 10
done

echo "=== Rollback of ${ENVIRONMENT} complete ==="
