#!/usr/bin/env bash
# codesign-darwin.sh — GoReleaser builds.hooks.post handler for darwin signing.
#
# Signs a darwin binary with the Developer ID Application certificate that was
# previously imported into the temporary keychain by the release workflow.
# Silently skips for non-darwin targets or when codesign is absent (e.g. Linux
# snapshot builds on ubuntu-latest).
#
# Required environment variables (injected by GoReleaser hook env):
#   ARTIFACT_PATH   — absolute or relative path to the built binary
#   ARTIFACT_TARGET — GoReleaser build target string (e.g. darwin_amd64_v1,
#                     linux_arm64). Sign only when this starts with "darwin_".
#   SIGNING_IDENTITY — Developer ID Application identity name, extracted after
#                      certificate import and stored in GITHUB_ENV before GoReleaser
#                      runs. Not a secret value.
#
# Called by .goreleaser.yaml builds[*].hooks.post

set -euo pipefail

ARTIFACT_PATH="${ARTIFACT_PATH:?ARTIFACT_PATH must be set by the GoReleaser hook env}"
ARTIFACT_TARGET="${ARTIFACT_TARGET:?ARTIFACT_TARGET must be set by the GoReleaser hook env}"

# Only sign darwin binaries; skip all other platforms without error.
if [[ "$ARTIFACT_TARGET" != darwin_* ]]; then
  exit 0
fi

# On non-macOS runners (e.g. ubuntu-latest snapshot builds), codesign is absent.
# Skip gracefully so snapshot dry-runs work without Apple credentials.
if ! command -v codesign > /dev/null 2>&1; then
  echo "codesign not available; skipping signing for: $ARTIFACT_PATH"
  exit 0
fi

# SIGNING_IDENTITY is set in GITHUB_ENV by the certificate import step before
# GoReleaser runs. If it is absent, signing cannot proceed.
if [ -z "${SIGNING_IDENTITY:-}" ]; then
  echo "codesign: SIGNING_IDENTITY not set; skipping signing for: $ARTIFACT_PATH"
  echo "(Expected on local builds without a Developer ID certificate.)"
  exit 0
fi

echo "Signing: $ARTIFACT_PATH"
codesign \
  --force --verify --verbose \
  --sign "$SIGNING_IDENTITY" \
  --options runtime \
  --timestamp \
  "$ARTIFACT_PATH"

# Verify the signature immediately after signing.
codesign --verify --deep --strict --verbose=2 "$ARTIFACT_PATH"
echo "Signed and verified: $ARTIFACT_PATH"
