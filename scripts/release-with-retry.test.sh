#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPT="${REPO_ROOT}/scripts/release-with-retry.sh"

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

assert_eq() {
  local got="$1"
  local want="$2"
  local msg="$3"
  if [[ "${got}" != "${want}" ]]; then
    fail "${msg}: got '${got}', want '${want}'"
  fi
}

make_fake_bin_dir() {
  local dir
  dir="$(mktemp -d "${TMPDIR:-/tmp}/release-retry-test.XXXXXX")"
  mkdir -p "${dir}/bin"
  printf '%s\n' "${dir}"
}

write_fake_goreleaser() {
  local dir="$1"
  cat > "${dir}/bin/goreleaser" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

: "${GORELEASER_COUNT_FILE:?missing}"
: "${GORELEASER_FIRST_EXIT:?missing}"
: "${GORELEASER_SECOND_EXIT:?missing}"
: "${GORELEASER_FIRST_OUTPUT:?missing}"
: "${GORELEASER_SECOND_OUTPUT:?missing}"

count=0
if [[ -f "${GORELEASER_COUNT_FILE}" ]]; then
  count="$(cat "${GORELEASER_COUNT_FILE}")"
fi
count=$((count + 1))
printf '%s\n' "${count}" > "${GORELEASER_COUNT_FILE}"

if [[ "${count}" -eq 1 ]]; then
  printf '%s\n' "${GORELEASER_FIRST_OUTPUT}"
  exit "${GORELEASER_FIRST_EXIT}"
fi

printf '%s\n' "${GORELEASER_SECOND_OUTPUT}"
exit "${GORELEASER_SECOND_EXIT}"
EOF
  chmod +x "${dir}/bin/goreleaser"
}

write_fake_sleep() {
  local dir="$1"
  cat > "${dir}/bin/sleep" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

: "${SLEEP_CALLS_FILE:?missing}"
printf '%s\n' "$*" >> "${SLEEP_CALLS_FILE}"
exit 0
EOF
  chmod +x "${dir}/bin/sleep"
}

run_case() {
  local name="$1"
  local expected_status="$2"
  local expected_invocations="$3"
  local expected_sleeps="$4"
  local first_exit="$5"
  local first_output="$6"
  local second_exit="$7"
  local second_output="$8"

  local tempdir fakebin count_file sleep_file status
  tempdir="$(make_fake_bin_dir)"
  fakebin="${tempdir}/bin"
  count_file="${tempdir}/goreleaser.count"
  sleep_file="${tempdir}/sleep.calls"

  write_fake_goreleaser "${tempdir}"
  write_fake_sleep "${tempdir}"

  status=0
  set +e
  PATH="${fakebin}:${PATH}" \
  TMPDIR="${tempdir}" \
  GORELEASER_COUNT_FILE="${count_file}" \
  GORELEASER_FIRST_EXIT="${first_exit}" \
  GORELEASER_SECOND_EXIT="${second_exit}" \
  GORELEASER_FIRST_OUTPUT="${first_output}" \
  GORELEASER_SECOND_OUTPUT="${second_output}" \
  SLEEP_CALLS_FILE="${sleep_file}" \
    bash "${SCRIPT}" >/dev/null 2>&1
  status=$?
  set -e

  local invocations sleeps
  invocations=0
  if [[ -f "${count_file}" ]]; then
    invocations="$(cat "${count_file}")"
  fi

  sleeps=0
  if [[ -f "${sleep_file}" ]]; then
    sleeps="$(wc -l < "${sleep_file}" | tr -d ' ')"
  fi

  assert_eq "${status}" "${expected_status}" "${name} exit status"
  assert_eq "${invocations}" "${expected_invocations}" "${name} goreleaser invocations"
  assert_eq "${sleeps}" "${expected_sleeps}" "${name} sleep calls"

  rm -rf "${tempdir}"
}

main() {
  run_case \
    "first-success" \
    0 1 0 \
    0 \
    "release ok" \
    0 \
    "unused"

  run_case \
    "non-retryable-failure" \
    7 1 0 \
    7 \
    "build failed" \
    0 \
    "unused"

  run_case \
    "apple-then-success" \
    0 2 1 \
    1 \
    $'HTTP 429 Too Many Requests\nRATE_LIMIT\nExceeded hourly limit of requests' \
    0 \
    "release ok"

  run_case \
    "apple-then-apple" \
    9 2 1 \
    9 \
    $'HTTP 429 Too Many Requests\nRATE_LIMIT\nExceeded hourly limit of requests' \
    9 \
    $'HTTP 429 Too Many Requests\nRATE_LIMIT\nExceeded hourly limit of requests'

  run_case \
    "apple-then-different-failure" \
    17 2 1 \
    1 \
    $'HTTP 429 Too Many Requests\nRATE_LIMIT\nExceeded hourly limit of requests' \
    17 \
    "homebrew publish failed"

  echo "All release retry wrapper tests passed."
}

main "$@"
