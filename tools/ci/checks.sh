#!/usr/bin/env bash
# Everything that has to be true before a change lands, in one place.
#
# The workflow calls this rather than listing steps in YAML, so what CI
# runs is the same thing a person can run - and so a failure can be
# reproduced without pushing a commit to find out.
#
#   tools/ci/checks.sh          everything
#   tools/ci/checks.sh go       just the Go side
#   tools/ci/checks.sh frontend just the frontend
set -euo pipefail

cd "$(dirname "$0")/../.."
what="${1:-all}"
failed=()

run() {
  local name="$1"; shift
  echo "::group::$name"
  if "$@"; then
    echo "  ok: $name"
  else
    echo "  FAILED: $name"
    failed+=("$name")
  fi
  echo "::endgroup::"
}

gofmt_check() {
  local unformatted
  unformatted="$(gofmt -l cmd internal)"
  if [ -n "$unformatted" ]; then
    echo "not gofmt'd:"; echo "$unformatted"
    return 1
  fi
}

if [ "$what" = all ] || [ "$what" = go ]; then
  run "go vet" go vet ./...
  run "gofmt" gofmt_check
  run "go test" go test ./...
fi

if [ "$what" = all ] || [ "$what" = frontend ]; then
  run "prettier" npm --prefix frontend run lint:format
  run "frontend tests" npm --prefix frontend run test -- --watch=false
  run "translations" node tools/i18n/check.mjs
fi

build_everything() {
  # make is not on every runner image, and a check suite that cannot run
  # because of that tells you nothing about the code. The fallback does
  # the same two steps the Makefile does.
  if command -v make > /dev/null; then
    make build
  else
    echo "no make on this machine; building directly"
    npm --prefix frontend run build \
      && go build -trimpath -o seid ./cmd/seid
  fi
}

if [ "$what" = all ]; then
  # Last, because it is the slowest and the least likely to be the
  # reason somebody is reading this output.
  run "build" build_everything
fi

if [ ${#failed[@]} -gt 0 ]; then
  echo
  echo "failed: ${failed[*]}"
  exit 1
fi
echo
echo "all checks passed"
