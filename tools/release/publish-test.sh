#!/usr/bin/env bash
# Exercises publish.sh against a stand-in forge.
#
# The publishing path is the one part of a release that cannot be tried
# out by running it: the first real attempt creates a public release, and
# a mistake in it is visible to everybody. So it is tested here against a
# small server that speaks both API shapes and records what it was sent.
#
#   tools/release/publish-test.sh
set -euo pipefail

cd "$(dirname "$0")"
work="$(mktemp -d)"
trap 'rm -rf "$work"; [ -n "${server_pid:-}" ] && kill "$server_pid" 2>/dev/null || true' EXIT

# A release directory with the shape make dist produces.
mkdir -p "$work/dist"
echo "binary" > "$work/dist/seid_v9.9.9_linux_amd64.tar.gz"
echo "binary" > "$work/dist/seid_v9.9.9_linux_arm64.tar.gz"
echo "abc  seid_v9.9.9_linux_amd64.tar.gz" > "$work/dist/SHA256SUMS"

cat > "$work/forge.py" <<'PY'
import json, os, sys
from http.server import BaseHTTPRequestHandler, HTTPServer

LOG = os.environ["FORGE_LOG"]

class Forge(BaseHTTPRequestHandler):
    def do_POST(self):
        length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(length)
        with open(LOG, "a") as fh:
            fh.write(json.dumps({
                "path": self.path,
                "auth": self.headers.get("Authorization", ""),
                "type": (self.headers.get("Content-Type") or "").split(";")[0],
                "bytes": len(body),
                "body": body[:8000].decode("utf-8", "replace"),
            }) + "\n")
        self.send_response(201)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(b'{"id": 4711}')

    def log_message(self, *args):
        pass

HTTPServer(("127.0.0.1", int(sys.argv[1])), Forge).serve_forever()
PY

port=8791
log="$work/requests.jsonl"
FORGE_LOG="$log" python3 "$work/forge.py" "$port" &
server_pid=$!
for _ in $(seq 1 40); do
  curl -sS -o /dev/null "http://127.0.0.1:$port/" 2>/dev/null && break || sleep 0.1
done

fail=0
check() {
  if [ "$2" = "$3" ]; then
    echo "  ok: $1"
  else
    echo "  FAILED: $1"
    echo "    got:  $2"
    echo "    want: $3"
    fail=1
  fi
}

field() { RECORD="$1" node -e '
  const r = JSON.parse(process.env.RECORD);
  const parts = process.argv[1].split(".");
  let v = r;
  for (const p of parts) v = v[p];
  process.stdout.write(String(v));
' "$2"; }

# --- Gitea -----------------------------------------------------------
: > "$log"
GITHUB_API_URL="http://127.0.0.1:$port/api/v1" \
GITHUB_REPOSITORY="statusengine/interface" \
GITHUB_SERVER_URL="http://127.0.0.1:$port" \
RELEASE_TOKEN="test-token" \
  ./publish.sh v9.9.9 "$work/dist" > "$work/gitea.out"

echo "Gitea:"
create="$(head -1 "$log")"
check "creates the release on the API host" \
  "$(field "$create" path)" "/api/v1/repos/statusengine/interface/releases"
check "sends the token" "$(field "$create" auth)" "token test-token"
check "sends JSON" "$(field "$create" type)" "application/json"
check "uploads three files" "$(( $(wc -l < "$log") - 1 ))" "3"
upload="$(sed -n '2p' "$log")"
check "uploads to the API host" \
  "$(field "$upload" path)" \
  "/api/v1/repos/statusengine/interface/releases/4711/assets?name=SHA256SUMS"
check "uploads as multipart" "$(field "$upload" type)" "multipart/form-data"

# --- GitHub ----------------------------------------------------------
: > "$log"
GITHUB_API_URL="http://127.0.0.1:$port/api.github.com" \
GITHUB_UPLOAD_URL="http://127.0.0.1:$port/uploads" \
GITHUB_REPOSITORY="statusengine/interface" \
GITHUB_SERVER_URL="https://github.com" \
RELEASE_TOKEN="test-token" \
  ./publish.sh v9.9.9 "$work/dist" > "$work/github.out"

echo "GitHub:"
upload="$(sed -n '2p' "$log")"
check "uploads to the upload host" \
  "$(field "$upload" path)" "/uploads/repos/statusengine/interface/releases/4711/assets?name=SHA256SUMS"
check "uploads raw bytes" "$(field "$upload" type)" "application/octet-stream"

# --- refusals --------------------------------------------------------
echo "refusals:"
if GITHUB_API_URL="http://127.0.0.1:$port/api/v1" GITHUB_REPOSITORY="a/b" \
   RELEASE_TOKEN=t ./publish.sh 1.2.3 "$work/dist" >/dev/null 2>&1; then
  echo "  FAILED: a tag without a v was accepted"; fail=1
else
  echo "  ok: refuses a tag that is not a version"
fi

if GITHUB_API_URL="http://127.0.0.1:$port/api/v1" GITHUB_REPOSITORY="a/b" \
   RELEASE_TOKEN=t ./publish.sh v9.9.9 "$work/empty" >/dev/null 2>&1; then
  echo "  FAILED: a directory with no SHA256SUMS was accepted"; fail=1
else
  echo "  ok: refuses a dist directory that was never built"
fi

if GITHUB_API_URL="http://127.0.0.1:$port/api/v1" GITHUB_REPOSITORY="a/b" \
   ./publish.sh v9.9.9 "$work/dist" >/dev/null 2>&1; then
  echo "  FAILED: published without a token"; fail=1
else
  echo "  ok: refuses to publish without a token"
fi

# A prerelease tag has to be marked as one, or it goes out as the
# version everybody should be running.
: > "$log"
GITHUB_API_URL="http://127.0.0.1:$port/api/v1" GITHUB_REPOSITORY="a/b" \
GITHUB_SERVER_URL="http://127.0.0.1:$port" RELEASE_TOKEN=t \
  ./publish.sh v9.9.9-rc1 "$work/dist" > /dev/null
create="$(head -1 "$log")"
if echo "$(field "$create" body)" | grep -q '"prerelease":true'; then
  echo "  ok: marks a release candidate as a prerelease"
else
  echo "  FAILED: v9.9.9-rc1 was published as a full release"; fail=1
fi

echo
if [ "$fail" -ne 0 ]; then
  echo "publish.sh does not behave as documented"
  exit 1
fi
echo "publish.sh behaves as documented against both forges"
