#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
awk '/^check:/,/^[^[:space:]]/ {if ($0 ~ /-short/) exit 1}' Makefile
awk '/^short:/ {seen=1; if ($0 !~ /hooks-test/) exit 1} /go test \.\/\.\.\. -short/ {go=1} END {exit !(seen && go)}' Makefile
for target in backup meta-tax wrap-verify laws-lock state; do
  rg -q "^${target}:" Makefile || { echo "missing Make target: $target" >&2; exit 1; }
done
