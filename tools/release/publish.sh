#!/usr/bin/env bash
# Publishes a release and its artifacts to the forge this repository
# lives on.
#
# Gitea and GitHub speak almost the same release API - same paths, same
# fields - and differ in exactly two places: where an asset is uploaded
# to, and how it is encoded. Everything else here is shared, which is
# why this is one script rather than two workflows.
#
#   tools/release/publish.sh v1.2.3 dist
#
# Environment (the workflow passes all of these; Actions sets the first
# three itself on both forges):
#   GITHUB_API_URL     https://git.example.org/api/v1  or  https://api.github.com
#   GITHUB_REPOSITORY  owner/name
#   GITHUB_SERVER_URL  https://git.example.org
#   RELEASE_TOKEN      a token allowed to create releases
#   DRY_RUN=1          print what would be sent and stop
set -euo pipefail

tag="${1:?usage: publish.sh <tag> <dist-dir>}"
dist="${2:-dist}"

api="${GITHUB_API_URL:?GITHUB_API_URL is not set}"
repo="${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is not set}"
server="${GITHUB_SERVER_URL:-}"
token="${RELEASE_TOKEN:-${GITHUB_TOKEN:-}}"
dry="${DRY_RUN:-}"

if [ -z "$token" ] && [ -z "$dry" ]; then
  echo "publish: no RELEASE_TOKEN; nothing can be published" >&2
  exit 1
fi

# A tag is what the whole release is named after, so a typo in it is
# worth catching before anything is created.
case "$tag" in
  v[0-9]*.[0-9]*.[0-9]*) ;;
  *) echo "publish: $tag does not look like a version tag (v1.2.3)" >&2; exit 1 ;;
esac

if [ ! -f "$dist/SHA256SUMS" ]; then
  echo "publish: $dist has no SHA256SUMS; run make dist first" >&2
  exit 1
fi

# GitHub uploads assets to a different host and wants the raw bytes;
# Gitea takes them on the API host as a multipart field.
case "$api" in
  *api.github.com*) flavour=github ;;
  *) flavour=gitea ;;
esac
echo "publishing $tag to $repo ($flavour, $api)"

# --- release notes ---------------------------------------------------
# Subjects only: the commit messages in this repository are long by
# design, and a release page is an index, not a reprint.
#
# What to summarise. A shallow checkout has no history to walk and an
# unfetched tag is not a revision, so both degrade to a release with no
# change list rather than a release that fails after the artifacts are
# already built. The workflow checks out with full history to avoid it.
if git rev-parse --verify --quiet "$tag^{commit}" > /dev/null; then
  head="$tag"
else
  echo "publish: $tag is not in this checkout; listing changes up to HEAD" >&2
  head="HEAD"
fi
previous="$(git describe --tags --abbrev=0 "$head^" 2>/dev/null || true)"
if [ -n "$previous" ]; then
  range="$previous..$head"
  heading="Changes since $previous"
else
  range="$head"
  heading="Changes"
fi
changes="$(git log --no-merges --pretty='- %s' "$range" 2>/dev/null || true)"
if [ -z "$changes" ]; then
  changes="- see the commit history"
fi

notes="## $heading

$changes

## Downloads

Every archive holds the \`seid\` binary with the frontend compiled into
it, the example configuration and the README. Verify one with:

\`\`\`
sha256sum -c SHA256SUMS --ignore-missing
\`\`\`
"
if [ -n "$previous" ] && [ -n "$server" ]; then
  notes="$notes
[All commits]($server/$repo/compare/$previous...$tag)"
fi

# --- create the release ----------------------------------------------
payload="$(TAG="$tag" NOTES="$notes" node -e '
  process.stdout.write(JSON.stringify({
    tag_name: process.env.TAG,
    name: process.env.TAG,
    body: process.env.NOTES,
    draft: false,
    prerelease: /-(rc|beta|alpha)/i.test(process.env.TAG),
  }));
')"

if [ -n "$dry" ]; then
  echo "--- would POST $api/repos/$repo/releases"
  echo "$payload" | node -e 'let s="";process.stdin.on("data",d=>s+=d).on("end",()=>{const o=JSON.parse(s);console.log("tag:",o.tag_name,"prerelease:",o.prerelease);console.log(o.body)})'
  echo "--- would upload:"
  ls -1 "$dist"
  exit 0
fi

response="$(curl -sS -X POST "$api/repos/$repo/releases" \
  -H "Authorization: token $token" \
  -H "Content-Type: application/json" \
  -d "$payload")"

id="$(RESPONSE="$response" node -e '
  const r = JSON.parse(process.env.RESPONSE || "{}");
  if (!r.id) { console.error("no release id in the response:", process.env.RESPONSE); process.exit(1); }
  process.stdout.write(String(r.id));
')"
echo "created release $id"

# --- upload every artifact -------------------------------------------
for file in "$dist"/*; do
  name="$(basename "$file")"
  echo "uploading $name"
  if [ "$flavour" = github ]; then
    uploads="${GITHUB_UPLOAD_URL:-https://uploads.github.com}"
    curl -sS -X POST "$uploads/repos/$repo/releases/$id/assets?name=$name" \
      -H "Authorization: token $token" \
      -H "Content-Type: application/octet-stream" \
      --data-binary "@$file" > /dev/null
  else
    curl -sS -X POST "$api/repos/$repo/releases/$id/assets?name=$name" \
      -H "Authorization: token $token" \
      -F "attachment=@$file" > /dev/null
  fi
done

if [ -n "$server" ]; then
  echo "published: $server/$repo/releases/tag/$tag"
else
  echo "published $tag"
fi
