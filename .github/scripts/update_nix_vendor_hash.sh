#!/usr/bin/env bash
# Builds the Nix flake in the given directory. If the build fails only because
# `vendorHash` in flake.nix is stale (go.mod/go.sum changed), rewrites it to the
# hash Nix reports and verifies the build again.
#
# Usage: update_nix_vendor_hash.sh <repo-dir>
# Outputs (to $GITHUB_OUTPUT): changed=true|false, hash=<new hash> when changed.
# Exits non-zero on any failure that is not a vendorHash mismatch.
set -euo pipefail

cd "${1:?usage: $0 <repo-dir>}"

GITHUB_OUTPUT="${GITHUB_OUTPUT:-/dev/null}"
GITHUB_STEP_SUMMARY="${GITHUB_STEP_SUMMARY:-/dev/null}"
log_file="$(mktemp)"

if nix build .#default -L 2>&1 | tee "$log_file"; then
    nix flake check
    echo "changed=false" >>"$GITHUB_OUTPUT"
    exit 0
fi

new_hash="$(grep -oP 'got:\s+\Ksha256-\S+' "$log_file" | head -n 1 || true)"
if [[ -z "$new_hash" ]]; then
    echo "::error::Nix build failed for a reason other than a vendorHash mismatch."
    exit 1
fi

old_hash="$(grep -oP 'vendorHash = "\K[^"]*' flake.nix)"
echo "vendorHash is stale: ${old_hash} -> ${new_hash}"
sed -i -E "s|(vendorHash = \")[^\"]*(\";)|\1${new_hash}\2|" flake.nix

nix build .#default -L
nix flake check

{
    echo "changed=true"
    echo "hash=${new_hash}"
} >>"$GITHUB_OUTPUT"
{
    echo "### Nix vendorHash updated"
    echo ""
    echo "\`go.mod\`/\`go.sum\` changed, so \`vendorHash\` in \`flake.nix\` was stale."
    echo ""
    echo "- old: \`${old_hash}\`"
    echo "- new: \`${new_hash}\`"
} >>"$GITHUB_STEP_SUMMARY"
