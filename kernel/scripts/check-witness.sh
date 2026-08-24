#!/usr/bin/env bash
set -u

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
spool_dir="$repo_root/.vera/spool"
output_file=''
json_tmp=''
version_files=()
FIRST_LINE_VALUE=''
child_pid=''

cleanup() {
  if [[ -n $child_pid ]]; then
    kill "$child_pid" 2>/dev/null || true
    child_pid=''
  fi
  [[ -z $output_file ]] || rm -f "$output_file"
  [[ -z $json_tmp ]] || rm -f "$json_tmp"
  local file
  for file in "${version_files[@]}"; do
    rm -f "$file"
  done
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
trap 'exit 129' HUP

if ! mkdir -p "$spool_dir"; then
  printf 'check-witness: cannot create spool directory\n' >&2
  exit 1
fi

git_env_args=()
while IFS= read -r variable; do
  if [[ $variable == GIT_* ]]; then
    git_env_args+=(-u "$variable")
  fi
done < <(compgen -e)

git_repo() {
  env "${git_env_args[@]}" git -C "$repo_root" "$@"
}

now_ms() {
  local value
  if ! value=$(date +%s%3N 2>/dev/null); then
    return 1
  fi
  if [[ $value =~ ^[0-9]{13}$ ]]; then
    printf '%s' "$value"
  else
    return 1
  fi
}

new_ulid() {
  local alphabet='0123456789ABCDEFGHJKMNPQRSTVWXYZ'
  local value=$1
  local encoded=''
  local index byte
  for ((index = 0; index < 10; index++)); do
    encoded="${alphabet:value%32:1}$encoded"
    value=$((value / 32))
  done
  local bytes
  if ! bytes=$(od -An -N16 -tu1 /dev/urandom); then
    return 1
  fi
  for byte in $bytes; do
    encoded+="${alphabet:byte%32:1}"
  done
  printf '%s' "$encoded"
}

first_line() {
  local output_file line_file line byte
  if ! output_file=$(mktemp "${TMPDIR:-/tmp}/vera-version-output.XXXXXX"); then
    return 1
  fi
  version_files+=("$output_file")
  if ! line_file=$(mktemp "${TMPDIR:-/tmp}/vera-version-line.XXXXXX"); then
    return 1
  fi
  version_files+=("$line_file")
  "$@" >"$output_file" 2>/dev/null &
  child_pid=$!
  if ! wait "$child_pid"; then
    child_pid=''
    rm -f "$output_file" "$line_file"
    FIRST_LINE_VALUE='unavailable'
    return
  fi
  child_pid=''
  if ! sed -n '1p' "$output_file" >"$line_file"; then
    return 1
  fi
  rm -f "$output_file"
  local bytes
  if ! bytes=$(od -An -v -tu1 "$line_file"); then
    rm -f "$line_file"
    return 1
  fi
  for byte in $bytes; do
    if [[ $byte == 0 ]]; then
      rm -f "$line_file"
      return 1
    fi
  done
  if ! command -v iconv >/dev/null 2>&1 || ! iconv -f UTF-8 -t UTF-8 "$line_file" >/dev/null 2>&1; then
    rm -f "$line_file"
    return 1
  fi
  line=$(<"$line_file")
  rm -f "$line_file"
  if [[ -z $line ]]; then
    FIRST_LINE_VALUE='unavailable'
  else
    FIRST_LINE_VALUE=$line
  fi
}

json_escape() {
  local value=$1
  local character code index
  for ((index = 0; index < ${#value}; index++)); do
    character=${value:index:1}
    case $character in
      '"') printf '\\"' ;;
      '\\') printf '\\\\' ;;
      $'\b') printf '\\b' ;;
      $'\f') printf '\\f' ;;
      $'\n') printf '\\n' ;;
      $'\r') printf '\\r' ;;
      $'\t') printf '\\t' ;;
      *)
        printf -v code '%d' "'$character"
        if ((code < 32)); then
          printf '\\u%04x' "$code"
        else
          printf '%s' "$character"
        fi
        ;;
    esac
  done
}

if ! git_sha=$(git_repo rev-parse HEAD 2>/dev/null) || [[ ! $git_sha =~ ^([0-9a-f]{40}|[0-9a-f]{64})$ ]]; then
  printf 'check-witness: cannot observe repository HEAD\n' >&2
  exit 1
fi
if ! git_status=$(git_repo status --porcelain --untracked-files=normal 2>/dev/null); then
  printf 'check-witness: cannot observe repository dirty state\n' >&2
  exit 1
fi
git_dirty=false
if [[ -n $git_status ]]; then
  git_dirty=true
fi
if ! first_line go version; then
  printf 'check-witness: tool version contains inadmissible bytes\n' >&2
  exit 1
fi
go_version=$FIRST_LINE_VALUE
if ! first_line golangci-lint version; then
  printf 'check-witness: tool version contains inadmissible bytes\n' >&2
  exit 1
fi
lint_version=$FIRST_LINE_VALUE
if ! first_line make --version; then
  printf 'check-witness: tool version contains inadmissible bytes\n' >&2
  exit 1
fi
make_version=$FIRST_LINE_VALUE

if ! started_ms=$(now_ms) || ! [[ $started_ms =~ ^[0-9]{13}$ ]]; then
  printf 'check-witness: cannot observe start time\n' >&2
  exit 1
fi
if ! started_at=$(date -u +'%Y-%m-%dT%H:%M:%SZ'); then
  printf 'check-witness: cannot observe start timestamp\n' >&2
  exit 1
fi
if ! run_id=$(new_ulid "$started_ms"); then
  printf 'check-witness: cannot generate run id\n' >&2
  exit 1
fi
if ! output_file=$(mktemp "${TMPDIR:-/tmp}/vera-check-output.XXXXXX"); then
  printf 'check-witness: cannot create output capture\n' >&2
  exit 1
fi
if ! json_tmp=$(mktemp "$spool_dir/.witness.XXXXXX"); then
  printf 'check-witness: cannot create witness temporary\n' >&2
  exit 1
fi

if (cd "$repo_root" && env "${git_env_args[@]}" make check) >"$output_file" 2>&1; then
  exit_code=0
else
  exit_code=$?
fi
if ! cat "$output_file"; then
  printf 'check-witness: cannot publish gate output\n' >&2
  exit 1
fi

if ! finished_ms=$(now_ms) || ! [[ $finished_ms =~ ^[0-9]{13}$ ]]; then
  printf 'check-witness: cannot observe finish time\n' >&2
  exit 1
fi
if ! finished_at=$(date -u +'%Y-%m-%dT%H:%M:%SZ'); then
  printf 'check-witness: cannot observe finish timestamp\n' >&2
  exit 1
fi
duration_ms=$((finished_ms - started_ms))
if command -v sha256sum >/dev/null 2>&1; then
  if ! hash_line=$(sha256sum "$output_file"); then
    printf 'check-witness: cannot hash gate output\n' >&2
    exit 1
  fi
else
  if ! hash_line=$(shasum -a 256 "$output_file"); then
    printf 'check-witness: cannot hash gate output\n' >&2
    exit 1
  fi
fi
if [[ ! $hash_line =~ ^([0-9a-f]{64})[[:space:]] ]]; then
  printf 'check-witness: hash output is malformed\n' >&2
  exit 1
fi
output_sha256=${BASH_REMATCH[1]}

if ! printf '{"schema":"vera.witness.v1","run_id":"%s","command":"make check","exit_code":%d,"started_at":"%s","finished_at":"%s","duration_ms":%d,"output_sha256":"%s","git_sha":"%s","git_dirty":%s,"tool_versions":{"go":"%s","golangci_lint":"%s","make":"%s"}}\n' \
  "$run_id" "$exit_code" "$started_at" "$finished_at" "$duration_ms" "$output_sha256" \
  "$(json_escape "$git_sha")" "$git_dirty" "$(json_escape "$go_version")" \
  "$(json_escape "$lint_version")" "$(json_escape "$make_version")" >"$json_tmp"; then
  printf 'check-witness: cannot serialize witness\n' >&2
  exit 1
fi
if [[ ! -s $json_tmp ]]; then
  printf 'check-witness: serialized witness is empty\n' >&2
  exit 1
fi
if ! mv "$json_tmp" "$spool_dir/$run_id.json"; then
  printf 'check-witness: cannot publish witness\n' >&2
  exit 1
fi
json_tmp=''

exit "$exit_code"
