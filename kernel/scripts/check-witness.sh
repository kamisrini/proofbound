#!/usr/bin/env bash
set -u

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
spool_dir="$repo_root/.vera/spool"
mkdir -p "$spool_dir"

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
  value=$(date +%s%3N 2>/dev/null || true)
  if [[ $value =~ ^[0-9]{13}$ ]]; then
    printf '%s' "$value"
  else
    printf '%s000' "$(date +%s)"
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
  for byte in $(od -An -N16 -tu1 /dev/urandom); do
    encoded+="${alphabet:byte%32:1}"
  done
  printf '%s' "$encoded"
}

first_line() {
  local output line
  if ! output=$("$@" 2>/dev/null); then
    printf 'unavailable'
    return
  fi
  IFS= read -r line <<<"$output"
  if [[ -z $line ]]; then
    printf 'unavailable'
  else
    printf '%s' "$line"
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

started_ms=$(now_ms)
started_at=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
run_id=$(new_ulid "$started_ms")
output_file=$(mktemp "${TMPDIR:-/tmp}/vera-check-output.XXXXXX")
json_tmp=$(mktemp "$spool_dir/.witness.XXXXXX")
trap 'rm -f "$output_file" "$json_tmp"' EXIT

if (cd "$repo_root" && make check) >"$output_file" 2>&1; then
  exit_code=0
else
  exit_code=$?
fi
cat "$output_file"

finished_ms=$(now_ms)
finished_at=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
duration_ms=$((finished_ms - started_ms))
if command -v sha256sum >/dev/null 2>&1; then
  read -r output_sha256 _ < <(sha256sum "$output_file")
else
  read -r output_sha256 _ < <(shasum -a 256 "$output_file")
fi

go_version=$(first_line go version)
lint_version=$(first_line golangci-lint version)
make_version=$(first_line make --version)

printf '{"schema":"vera.witness.v1","run_id":"%s","command":"make check","exit_code":%d,"started_at":"%s","finished_at":"%s","duration_ms":%d,"output_sha256":"%s","git_sha":"%s","git_dirty":%s,"tool_versions":{"go":"%s","golangci_lint":"%s","make":"%s"}}\n' \
  "$run_id" "$exit_code" "$started_at" "$finished_at" "$duration_ms" "$output_sha256" \
  "$(json_escape "$git_sha")" "$git_dirty" "$(json_escape "$go_version")" \
  "$(json_escape "$lint_version")" "$(json_escape "$make_version")" >"$json_tmp"
mv "$json_tmp" "$spool_dir/$run_id.json"

exit "$exit_code"
