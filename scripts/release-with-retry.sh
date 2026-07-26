#!/usr/bin/env bash
# Release workflow wrapper around `goreleaser release --clean`.
set -euo pipefail

readonly RETRY_WAIT_SECONDS=3600
readonly FIRST_ATTEMPT_LOG="$(mktemp "${TMPDIR:-/tmp}/goreleaser-release-attempt1.XXXXXX.log")"
readonly SECOND_ATTEMPT_LOG="$(mktemp "${TMPDIR:-/tmp}/goreleaser-release-attempt2.XXXXXX.log")"

cleanup() {
  rm -f "${FIRST_ATTEMPT_LOG}" "${SECOND_ATTEMPT_LOG}"
}
trap cleanup EXIT

run_release_attempt() {
  local log_file="$1"
  goreleaser release --clean 2>&1 | tee "${log_file}"
  local cmd_status=${PIPESTATUS[0]}
  return "${cmd_status}"
}

is_apple_rate_limit_error() {
  local log_file="$1"

  # Apple's hourly-limit message is specific to the notarisation service.
  if grep -Eq 'Exceeded hourly limit' "${log_file}"; then
    return 0
  fi

  # Generic 429/RATE_LIMIT text can also come from GitHub or Homebrew. Require
  # Apple/notarisation context on the same log line before retrying.
  grep -Eiq \
    '(apple|app[[:space:]]+store[[:space:]]+connect|notari[sz](e|ed|ing|ation)?|notary).*(429([[:space:]]+Too[[:space:]]+Many[[:space:]]+Requests)?|RATE_LIMIT)|(429([[:space:]]+Too[[:space:]]+Many[[:space:]]+Requests)?|RATE_LIMIT).*(apple|app[[:space:]]+store[[:space:]]+connect|notari[sz](e|ed|ing|ation)?|notary)' \
    "${log_file}"
}

set +e
run_release_attempt "${FIRST_ATTEMPT_LOG}"
first_status=$?
set -e

if [[ "${first_status}" -eq 0 ]]; then
  exit 0
fi

if ! is_apple_rate_limit_error "${FIRST_ATTEMPT_LOG}"; then
  exit "${first_status}"
fi

echo "::warning::Apple notarization was rate-limited (429/RATE_LIMIT). Cooling down for 1 hour before one retry."
sleep "${RETRY_WAIT_SECONDS}"

set +e
run_release_attempt "${SECOND_ATTEMPT_LOG}"
second_status=$?
set -e

if [[ "${second_status}" -eq 0 ]]; then
  exit 0
fi
exit "${second_status}"
