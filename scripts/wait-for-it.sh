#!/usr/bin/env bash
set -euo pipefail

HOST="${1:?Usage: wait-for-it.sh <host> <port> [-t timeout]}"
PORT="${2:?Usage: wait-for-it.sh <host> <port> [-t timeout]}"
TIMEOUT=30

shift 2
while [[ $# -gt 0 ]]; do
  case "$1" in
    -t) TIMEOUT="$2"; shift 2 ;;
    *) shift ;;
  esac
done

echo "Waiting for $HOST:$PORT (timeout: ${TIMEOUT}s)..."

elapsed=0
while ! (echo > /dev/tcp/"$HOST"/"$PORT") 2>/dev/null; do
  sleep 1
  elapsed=$((elapsed + 1))
  if [ "$elapsed" -ge "$TIMEOUT" ]; then
    echo "Error: $HOST:$PORT not reachable after ${TIMEOUT}s" >&2
    exit 1
  fi
done

echo "$HOST:$PORT is ready."
