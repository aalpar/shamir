#!/usr/bin/env bash
set -euo pipefail

COMPONENT="${1:-}"
VERSION_FILE="VERSION"

if [[ ! -f "$VERSION_FILE" ]]; then
  echo "Error: $VERSION_FILE not found" >&2
  exit 1
fi

CURRENT=$(tr -d '[:space:]' < "$VERSION_FILE")
CURRENT="${CURRENT#v}"

IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"

case "$COMPONENT" in
  major)
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
    ;;
  minor)
    MINOR=$((MINOR + 1))
    PATCH=0
    ;;
  patch)
    PATCH=$((PATCH + 1))
    ;;
  *)
    echo "Usage: $0 {major|minor|patch}" >&2
    exit 1
    ;;
esac

NEW="v${MAJOR}.${MINOR}.${PATCH}"
echo "$NEW" > "$VERSION_FILE"
echo "Bumped version: v${CURRENT} → ${NEW}"
