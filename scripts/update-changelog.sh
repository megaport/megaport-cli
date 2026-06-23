#!/usr/bin/env bash
# Fold GoReleaser-generated release notes into CHANGELOG.md in Keep a Changelog
# format. Called by the release workflow after a tag is published; the notes are
# the body of the draft GitHub release that GoReleaser created.
#
# Usage: update-changelog.sh <version> <notes-file> [changelog-file]
#   version        the release tag, e.g. v0.13.0
#   notes-file     file holding the generated notes (GoReleaser's "## Changelog"
#                  block with "### Added/### Fixed/..." groups)
#   changelog-file defaults to CHANGELOG.md at the repo root
#
# Idempotent: if the version heading already exists, it is a no-op.
# The release date is today (UTC); override with CHANGELOG_DATE for tests.
set -euo pipefail

version="${1:?usage: update-changelog.sh <version> <notes-file> [changelog-file]}"
notes_file="${2:?missing notes file}"
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
changelog="${3:-$repo_root/CHANGELOG.md}"
repo_url="https://github.com/megaport/megaport-cli"

[ -f "$notes_file" ] || { echo "error: notes file not found: $notes_file" >&2; exit 1; }
[ -f "$changelog" ] || { echo "error: changelog not found: $changelog" >&2; exit 1; }

if grep -qF "## [$version]" "$changelog"; then
  echo "CHANGELOG.md already documents $version; nothing to do."
  exit 0
fi

# The insertion below anchors on these two lines; fail clearly if either is gone.
grep -q '^## \[Unreleased\]' "$changelog" || { echo "error: no '## [Unreleased]' heading in $changelog" >&2; exit 1; }
grep -q '^\[Unreleased\]: ' "$changelog" || { echo "error: no '[Unreleased]:' reference link in $changelog" >&2; exit 1; }

date="${CHANGELOG_DATE:-$(date -u +%F)}"

# Previous version is whatever the [Unreleased] compare link currently points at.
prev="$(sed -n 's#^\[Unreleased\]:.*/compare/\(v[0-9A-Za-z.\-]*\)\.\.\.HEAD.*#\1#p' "$changelog" | head -n1)"

# Transform the generated notes into the version's section body:
#   - drop GoReleaser's "## Changelog" header
#   - turn "*  <msg>" bullets into "- <msg>"
#   - blank line before each group heading after the first
#   - strip the conventional-commit type prefix (the heading already says the
#     category) while keeping the rest of the message intact
section_body="$(
  awk '
    { sub(/\r$/, "") }
    /^## Changelog[[:space:]]*$/ { next }
    /^###[[:space:]]/ {
      if (seen) print ""
      seen = 1
      print
      next
    }
    /^\*[[:space:]]/ {
      sub(/^\*[[:space:]]+/, "- ")
      print
      next
    }
    # Keep only group headings and bullets; drop prose, footers, and stray lines.
    { next }
  ' "$notes_file" \
    | sed -E 's/^- (feat|fix|perf|refactor|build|ci|docs|test|style|chore|revert)(\([^)]*\))?!?: /- /' \
    | sed -E 's/^- [A-Z]+-[0-9]+: /- /'
)"

# An empty section means no qualifying commits (e.g. a docs/tooling-only release),
# not an error: leave the changelog untouched rather than failing the release job.
if [ -z "$section_body" ]; then
  echo "No user-facing changes to record for $version; leaving CHANGELOG.md unchanged."
  exit 0
fi

NEW_BLOCK="$(printf '## [%s] - %s\n\n%s' "$version" "$date" "$section_body")"
export NEW_BLOCK

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

# Insert the new version block right below [Unreleased] (whose body is reset to
# empty, since releases now come from generated notes, not hand-edited items),
# then repoint the [Unreleased] compare link and add the new version's link.
awk -v ver="$version" -v prev="$prev" -v url="$repo_url" '
  /^## \[Unreleased\]/ {
    print
    print ""
    print ENVIRON["NEW_BLOCK"]
    print ""
    inunreleased = 1
    next
  }
  inunreleased && (/^## \[/ || /^\[[^]]+\]:[[:space:]]/) { inunreleased = 0 }
  inunreleased { next }
  /^\[Unreleased\]: / {
    sub(/compare\/[^ ]+\.\.\.HEAD/, "compare/" ver "...HEAD")
    print
    if (prev != "")
      print "[" ver "]: " url "/compare/" prev "..." ver
    else
      print "[" ver "]: " url "/releases/tag/" ver
    next
  }
  { print }
' "$changelog" > "$tmp"

mv "$tmp" "$changelog"
trap - EXIT
echo "CHANGELOG.md updated for $version (previous: ${prev:-none}, date: $date)."
