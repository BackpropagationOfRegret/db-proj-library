#!/usr/bin/env bash
# Deploy library stack via Dokploy API.
#
#   export DOKPLOY_URL=http://192.168.3.4:3000
#   export DOKPLOY_API_KEY='your-token'
#   # optional if auto-detect fails:
#   export ENVIRONMENT_ID='...'
#   ./scripts/dokploy-deploy.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

DOKPLOY_URL="${DOKPLOY_URL:-http://127.0.0.1:3000}"
API_KEY="${DOKPLOY_API_KEY:?set DOKPLOY_API_KEY (Dokploy → Settings → API/Tokens)}"
PROJECT_NAME="${PROJECT_NAME:-library}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.dokploy.yml}"

api() {
  local method="$1" path="$2"
  shift 2
  curl -sS -X "$method" "${DOKPLOY_URL}/api${path}" \
    -H "x-api-key: ${API_KEY}" \
    -H "Content-Type: application/json" \
    "$@"
}

echo "Dokploy: ${DOKPLOY_URL}"
echo "Listing projects..."
projects="$(api GET /project.all)"

ENVIRONMENT_ID="${ENVIRONMENT_ID:-}"
if [[ -z "$ENVIRONMENT_ID" ]]; then
  ENVIRONMENT_ID="$(PROJECTS_JSON="$projects" python3 - <<'PY'
import json, os
data = json.loads(os.environ["PROJECTS_JSON"])
items = data if isinstance(data, list) else data.get("projects") or data.get("data") or []
for p in items:
    for env in p.get("environments") or []:
        eid = env.get("environmentId") or env.get("id")
        if eid:
            print(eid)
            raise SystemExit
    eid = p.get("environmentId")
    if eid:
        print(eid)
        raise SystemExit
PY
)" || true
fi

if [[ -z "${ENVIRONMENT_ID}" ]]; then
  echo "$projects" | head -c 2000
  echo
  echo "Could not detect environmentId."
  echo "Create a project in Dokploy UI, then re-run with ENVIRONMENT_ID=..."
  exit 1
fi

echo "Using environmentId=${ENVIRONMENT_ID}"

payload="$(COMPOSE_FILE="$COMPOSE_FILE" ENVIRONMENT_ID="$ENVIRONMENT_ID" PROJECT_NAME="$PROJECT_NAME" python3 - <<'PY'
import json, os
print(json.dumps({
    "name": os.environ["PROJECT_NAME"],
    "environmentId": os.environ["ENVIRONMENT_ID"],
    "composeType": "docker-compose",
    "appName": "library-stack",
    "composeFile": open(os.environ["COMPOSE_FILE"]).read(),
}))
PY
)"

echo "Creating compose app..."
created="$(api POST /compose.create -d "$payload")"
echo "$created"

compose_id="$(CREATED_JSON="$created" python3 - <<'PY'
import json, os
data = json.loads(os.environ["CREATED_JSON"])
print(data.get("composeId") or data.get("id") or "")
PY
)"

if [[ -z "$compose_id" ]]; then
  echo "Failed to get composeId from create response"
  exit 1
fi

echo "Deploying composeId=${compose_id}..."
api POST /compose.deploy -d "{\"composeId\":\"${compose_id}\"}"
echo
echo "Done. Open Dokploy UI to verify the deployment."
