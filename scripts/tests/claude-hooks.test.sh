#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"; tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
mkdir "$tmp/bin"
cat >"$tmp/bin/jq" <<'EOF'
#!/usr/bin/env bash
input=$(cat)
case $* in
  *tool_input.command*) sed -nE 's/.*"command":"([^"]*)".*/\1/p' <<<"$input" ;;
  *tool_input.file_path*) sed -nE 's/.*"file_path":"([^"]*)".*/\1/p' <<<"$input" ;;
  *) exit 0 ;;
esac
EOF
chmod +x "$tmp/bin/jq"
hook_path=$tmp/bin:$PATH
printf '%s\n' '{"tool_input":{"command":"git status"}}' | PATH=$hook_path .claude/hooks/block-secrets.sh
if printf '%s\n' '{"tool_input":{"command":"git push --force origin main"}}' | PATH=$hook_path .claude/hooks/block-secrets.sh >/dev/null 2>&1; then exit 1; fi
if printf '%s\n' '{"tool_input":{"command":"printf x > docs/decisions/INDEX.md"}}' | PATH=$hook_path .claude/hooks/block-secrets.sh >/dev/null 2>&1; then exit 1; fi
printf '@generated fixture\n' >"$tmp/generated.md"
if printf '{"tool_input":{"file_path":"%s"}}\n' "$tmp/generated.md" | PATH=$hook_path .claude/hooks/block-generated-edit.sh >/dev/null 2>&1; then exit 1; fi
printf 'plain\n' >"$tmp/plain.md"; printf '{"tool_input":{"file_path":"%s"}}\n' "$tmp/plain.md" | PATH=$hook_path .claude/hooks/block-generated-edit.sh
empty=$tmp/empty; mkdir "$empty"
if printf '%s\n' '{"tool_input":{"command":"git status"}}' | PATH=$empty /bin/bash .claude/hooks/block-secrets.sh >/dev/null 2>&1; then exit 1; fi
rg -q '"PreToolUse"' .claude/settings.json
rg -q '"PostToolUse"' .claude/settings.json
rg -q '"Stop"' .claude/settings.json
