#!/bin/sh
set -e

start_model_status_sidecar() {
    if [ "${MODEL_STATUS_ENABLED:-false}" != "true" ]; then
        return 0
    fi

    required_vars="MODEL_STATUS_UPSTREAM_API_BASE_URL MODEL_STATUS_UPSTREAM_API_KEY MODEL_STATUS_ADMIN_BOOTSTRAP_PASSWORD MODEL_STATUS_SESSION_SECRET"
    for var in $required_vars; do
        eval "value=\${$var:-}"
        if [ -z "$value" ]; then
            echo "[model-status] missing required env: $var"
            exit 1
        fi
    done

    MODEL_STATUS_HOST="127.0.0.1"
    MODEL_STATUS_PORT="3001"
    MODEL_STATUS_DATABASE_FILE="/tmp/model-status.db"
    MODEL_STATUS_ADMIN_BOOTSTRAP_USERNAME="${MODEL_STATUS_ADMIN_BOOTSTRAP_USERNAME:-admin}"
    MODEL_STATUS_UPSTREAM_NAME="${MODEL_STATUS_UPSTREAM_NAME:-Default Upstream}"

    if [ -n "${MODEL_STATUS_ACCESS_URL:-}" ]; then
        ACCESS_URL_VALUE="${MODEL_STATUS_ACCESS_URL}"
    elif [ -n "${RENDER_EXTERNAL_URL:-}" ]; then
        ACCESS_URL_VALUE="${RENDER_EXTERNAL_URL%/}/status"
    else
        ACCESS_URL_VALUE="http://127.0.0.1:${SERVER_PORT:-8080}/status"
    fi

    if [ -n "${MODEL_STATUS_UPSTREAM_MODELS_URL:-}" ]; then
        MODEL_STATUS_UPSTREAM_MODELS_URL_VALUE="${MODEL_STATUS_UPSTREAM_MODELS_URL}"
    else
        MODEL_STATUS_UPSTREAM_MODELS_URL_VALUE="${MODEL_STATUS_UPSTREAM_API_BASE_URL%/}/models"
    fi

    WEB_ORIGIN_VALUE="$(printf '%s' "$ACCESS_URL_VALUE" | sed -E 's#(https?://[^/]+).*#\1#')"
    if [ -z "$WEB_ORIGIN_VALUE" ]; then
        WEB_ORIGIN_VALUE="http://127.0.0.1:${MODEL_STATUS_PORT}"
    fi

    echo "[model-status] starting sidecar on ${MODEL_STATUS_HOST}:${MODEL_STATUS_PORT}"
    (
        cd /app/model-status
        HOST="$MODEL_STATUS_HOST" \
        PORT="$MODEL_STATUS_PORT" \
        WEB_ORIGIN="$WEB_ORIGIN_VALUE" \
        ACCESS_URL="$ACCESS_URL_VALUE" \
        DATABASE_FILE="$MODEL_STATUS_DATABASE_FILE" \
        ADMIN_BOOTSTRAP_USERNAME="$MODEL_STATUS_ADMIN_BOOTSTRAP_USERNAME" \
        ADMIN_BOOTSTRAP_PASSWORD="$MODEL_STATUS_ADMIN_BOOTSTRAP_PASSWORD" \
        SESSION_SECRET="$MODEL_STATUS_SESSION_SECRET" \
        MODEL_STATUS_MODELS="${MODEL_STATUS_MODELS:-}" \
        npm run start
    ) >/tmp/model-status.log 2>&1 &
    MODEL_STATUS_PID=$!
    echo "[model-status] pid=${MODEL_STATUS_PID}"

    HEALTH_URL="http://${MODEL_STATUS_HOST}:${MODEL_STATUS_PORT}/api/health"
    for i in $(seq 1 60); do
        if curl -fsS "$HEALTH_URL" >/dev/null 2>&1; then
            break
        fi
        sleep 1
        if [ "$i" -eq 60 ]; then
            echo "[model-status] failed to become healthy, last log lines:"
            tail -n 80 /tmp/model-status.log || true
            exit 1
        fi
    done

    echo "[model-status] bootstrapping runtime settings"
    MODEL_STATUS_BOOTSTRAP_BASE_URL="http://${MODEL_STATUS_HOST}:${MODEL_STATUS_PORT}" \
    MODEL_STATUS_UPSTREAM_MODELS_URL="$MODEL_STATUS_UPSTREAM_MODELS_URL_VALUE" \
    MODEL_STATUS_UPSTREAM_NAME="$MODEL_STATUS_UPSTREAM_NAME" \
    node <<'NODE' || echo "[model-status] bootstrap failed, keep sidecar running (check /tmp/model-status.log)"
const baseUrl = process.env.MODEL_STATUS_BOOTSTRAP_BASE_URL;
const username = process.env.MODEL_STATUS_ADMIN_BOOTSTRAP_USERNAME || "admin";
const password = process.env.MODEL_STATUS_ADMIN_BOOTSTRAP_PASSWORD;
const upstreamName = process.env.MODEL_STATUS_UPSTREAM_NAME || "Default Upstream";
const upstreamApiBaseUrl = process.env.MODEL_STATUS_UPSTREAM_API_BASE_URL;
const upstreamModelsUrl = process.env.MODEL_STATUS_UPSTREAM_MODELS_URL;
const upstreamApiKey = process.env.MODEL_STATUS_UPSTREAM_API_KEY;

function required(name, value) {
  if (!value || String(value).trim().length === 0) {
    throw new Error(`missing required env for bootstrap: ${name}`);
  }
}

async function request(path, options = {}) {
  const resp = await fetch(`${baseUrl}${path}`, options);
  const text = await resp.text();
  let json = null;
  try {
    json = text ? JSON.parse(text) : null;
  } catch {
    json = null;
  }
  if (!resp.ok) {
    const msg = json?.error || json?.message || text || `HTTP ${resp.status}`;
    throw new Error(`${path} failed: ${msg}`);
  }
  return { resp, json };
}

async function main() {
  required("MODEL_STATUS_BOOTSTRAP_BASE_URL", baseUrl);
  required("MODEL_STATUS_ADMIN_BOOTSTRAP_PASSWORD", password);
  required("MODEL_STATUS_UPSTREAM_API_BASE_URL", upstreamApiBaseUrl);
  required("MODEL_STATUS_UPSTREAM_MODELS_URL", upstreamModelsUrl);
  required("MODEL_STATUS_UPSTREAM_API_KEY", upstreamApiKey);

  const login = await request("/api/admin/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  });
  const cookie = login.resp.headers.get("set-cookie");
  if (!cookie) {
    throw new Error("admin login succeeded but no session cookie returned");
  }
  const sessionCookie = cookie.split(";")[0];
  const adminHeaders = {
    "Content-Type": "application/json",
    Cookie: sessionCookie,
  };

  await request("/api/admin/settings", {
    method: "PUT",
    headers: adminHeaders,
    body: JSON.stringify({
      probeIntervalMs: 600000,
      catalogSyncIntervalMs: 1800000,
      probeTimeoutMs: 15000,
      probeConcurrency: 1,
      probeMaxTokens: 1,
      probeTemperature: 0,
      degradedRetryAttempts: 1,
      failedRetryAttempts: 0,
      upstreams: [
        {
          id: "default",
          name: upstreamName,
          apiBaseUrl: upstreamApiBaseUrl,
          modelsUrl: upstreamModelsUrl,
          apiKey: upstreamApiKey,
          isActive: true,
        },
      ],
    }),
  });

  await request("/api/admin/actions/sync-models", {
    method: "POST",
    headers: adminHeaders,
  });

  await request("/api/admin/actions/run-probes", {
    method: "POST",
    headers: adminHeaders,
  });
}

main().catch((err) => {
  console.error(`[model-status bootstrap] ${err instanceof Error ? err.message : String(err)}`);
  process.exit(1);
});
NODE
}

# Fix data directory permissions when running as root.
# Docker named volumes / host bind-mounts may be owned by root,
# preventing the non-root sub2api user from writing files.
if [ "$(id -u)" = "0" ]; then
    mkdir -p /app/data
    # Use || true to avoid failure on read-only mounted files (e.g. config.yaml:ro)
    chown -R sub2api:sub2api /app/data 2>/dev/null || true
    # Re-invoke this script as sub2api so the flag-detection below
    # also runs under the correct user.
    exec su-exec sub2api "$0" "$@"
fi

# Compatibility: if the first arg looks like a flag (e.g. --help),
# prepend the default binary so it behaves the same as the old
# ENTRYPOINT ["/app/sub2api"] style.
if [ "${1#-}" != "$1" ]; then
    set -- /app/sub2api "$@"
fi

start_model_status_sidecar

exec "$@"
