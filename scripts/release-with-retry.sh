#!/usr/bin/env bash
set -euo pipefail

readonly RETRY_WAIT_SECONDS=900
readonly FIRST_ATTEMPT_LOG="$(mktemp "${TMPDIR:-/tmp}/goreleaser-release-attempt1.XXXXXX.log")"
readonly SECOND_ATTEMPT_LOG="$(mktemp "${TMPDIR:-/tmp}/goreleaser-release-attempt2.XXXXXX.log")"

cleanup() {
  rm -f "${FIRST_ATTEMPT_LOG}" "${SECOND_ATTEMPT_LOG}"
}
trap cleanup EXIT

run_release_attempt() {
  local log_file="$1"
  set +e
  goreleaser release --clean 2>&1 | tee "${log_file}"
  local cmd_status=${PIPESTATUS[0]}
  set -e
  return "${cmd_status}"
}

is_apple_rate_limit_error() {
  local log_file="$1"
  grep -Eq '429 Too Many Requests|RATE_LIMIT|Exceeded hourly limit' "${log_file}"
}

if run_release_attempt "${FIRST_ATTEMPT_LOG}"; then
  exit 0
fi
first_status=$?

if ! is_apple_rate_limit_error "${FIRST_ATTEMPT_LOG}"; then
  exit "${first_status}"
fi

echo "::warning::Apple notarization was rate-limited (429/RATE_LIMIT). Cooling down for 15 minutes before one retry."
sleep "${RETRY_WAIT_SECONDS}"

if run_release_attempt "${SECOND_ATTEMPT_LOG}"; then
  exit 0
fi
second_status=$?
exit "${second_status}"
