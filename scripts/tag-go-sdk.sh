#!/usr/bin/env sh
set -eu

VERSION="${1:-}"
if ! printf '%s\n' "$VERSION" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
  echo "usage: $0 vMAJOR.MINOR.PATCH" >&2
  exit 1
fi

ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
TAG="sdk/go/$VERSION"

cd "$ROOT"
if [ -n "$(git status --porcelain)" ]; then
  echo "working tree must be clean before tagging" >&2
  exit 1
fi
if git rev-parse "$TAG" >/dev/null 2>&1; then
  echo "tag already exists: $TAG" >&2
  exit 1
fi

(cd sdk/go && go test ./...)
git tag -a "$TAG" -m "WLTE OpenAPI Go SDK $VERSION"

echo "created $TAG"
echo "push with: git push origin $TAG"
