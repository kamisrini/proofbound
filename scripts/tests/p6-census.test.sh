#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

script=$PWD/scripts/p6-census.sh
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

new_fixture() {
  local dir=$1
  mkdir -p "$dir/docs/plans" "$dir/docs/decisions" "$dir/scripts"
  cp "$script" "$dir/scripts/p6-census.sh"
  printf '# fixture\n' >"$dir/README.md"
  printf 'all:\n\t@true\n' >"$dir/Makefile"
  git -C "$dir" init -q
  git -C "$dir" config user.email fixture@example.invalid
  git -C "$dir" config user.name fixture
}

write_valid_registry() {
  local dir=$1
  {
    printf '# universe-profile\tfixture\n'
    printf 'row_id\tcategory\tsubject\tobservation\tprobe\tdisposition\tdecision\tevidence\n'
    for n in {1..8}; do
      printf 'C%d-001\tC%d\tfixture-%d\tfixture row\tpath:README.md\tclose-in-P6\t—\tREADME.md\n' "$n" "$n" "$n"
    done
  } >"$dir/docs/plans/p6-census-rows.tsv"
}

commit_fixture() {
  local dir=$1
  git -C "$dir" add .
  git -C "$dir" commit -qm fixture
}

base=$tmp/base
new_fixture "$base"
write_valid_registry "$base"
commit_fixture "$base"
bash "$base/scripts/p6-census.sh" --write --root "$base"
git -C "$base" add docs/plans/p6-census.md
git -C "$base" commit -qm census
bash "$base/scripts/p6-census.sh" --check --root "$base"
echo 'ok generated-freshness'

printf '\n' >>"$base/docs/plans/p6-census.md"
if bash "$base/scripts/p6-census.sh" --check --root "$base" >/dev/null 2>&1; then
  echo 'stale generated census was accepted' >&2
  exit 1
fi
echo 'ok generated-freshness-drift'

p5result=$tmp/p5result
cp -a "$base" "$p5result"
git -C "$p5result" reset -q --hard HEAD
mkdir -p "$p5result/docs/verification"
printf '#!/usr/bin/env bash\nexit 0\n' >"$p5result/scripts/p6-task6-results.sh"
chmod +x "$p5result/scripts/p6-task6-results.sh"
printf '| F-1 | `fixture-7` | `false` | fixture observation | fixture:artifact | no action |\n' \
  >"$p5result/docs/verification/p6-measurements-falsifiers.md"
sed -i '/^C7-001\t/ s#path:README.md\tclose-in-P6\t—\tREADME.md#p5-result:measurement:fixture-7\tclose-in-P6\t—\tdocs/verification/p6-measurements-falsifiers.md#' \
  "$p5result/docs/plans/p6-census-rows.tsv"
git -C "$p5result" add docs/plans/p6-census-rows.tsv docs/verification/p6-measurements-falsifiers.md scripts/p6-task6-results.sh
p5_output=$(bash "$p5result/scripts/p6-census.sh" --render --root "$p5result")
if ! grep -Fq '| C7-001 | C7 | fixture-7 | fixture row | p5-result:measurement:fixture-7 | closed |' <<<"$p5_output"; then
  echo 'p5-result probe did not close a valid result row' >&2
  exit 1
fi
echo 'ok p5-result-probe'

schema=$tmp/schema
cp -a "$base" "$schema"
git -C "$schema" reset -q --hard HEAD
sed -i 's/C2-001/C1-001/' "$schema/docs/plans/p6-census-rows.tsv"
if bash "$schema/scripts/p6-census.sh" --render --root "$schema" >/dev/null 2>&1; then
  echo 'duplicate row ID was accepted' >&2
  exit 1
fi
echo 'ok registry-schema'

missing=$tmp/missing
cp -a "$base" "$missing"
git -C "$missing" reset -q --hard HEAD
sed -i '/C8-001/d' "$missing/docs/plans/p6-census-rows.tsv"
if bash "$missing/scripts/p6-census.sh" --render --root "$missing" >/dev/null 2>&1; then
  echo 'missing category was accepted' >&2
  exit 1
fi
echo 'ok unknown-and-missing-universe'

evidence=$tmp/evidence
cp -a "$base" "$evidence"
git -C "$evidence" reset -q --hard HEAD
sed -i 's/C3-001\tC3\tfixture-3\tfixture row\tpath:README.md\tclose-in-P6\t—\tREADME.md/C3-001\tC3\tfixture-3\tfixture row\tpath:missing.md\tclose-in-P6\t—\tREADME.md/' "$evidence/docs/plans/p6-census-rows.tsv"
if bash "$evidence/scripts/p6-census.sh" --render --root "$evidence" >/dev/null 2>&1; then
  echo 'open row with claimed evidence was accepted' >&2
  exit 1
fi
echo 'ok closed-evidence'

decision=$tmp/decision
cp -a "$base" "$decision"
git -C "$decision" reset -q --hard HEAD
sed -i 's/C4-001\tC4\tfixture-4\tfixture row\tpath:README.md\tclose-in-P6\t—\tREADME.md/C4-001\tC4\tfixture-4\tfixture row\tpath:missing.md\tdefer-to-P7\tdocs\/decisions\/VD-missing.md\t—/' "$decision/docs/plans/p6-census-rows.tsv"
if bash "$decision/scripts/p6-census.sh" --render --root "$decision" >/dev/null 2>&1; then
  echo 'unaccepted deferral decision was accepted' >&2
  exit 1
fi
echo 'ok decision-closure'

unknown=$tmp/unknown
cp -a "$base" "$unknown"
git -C "$unknown" reset -q --hard HEAD
printf '# use `make absent`\n' >"$unknown/README.md"
git -C "$unknown" add README.md
if bash "$unknown/scripts/p6-census.sh" --render --root "$unknown" >/dev/null 2>&1; then
  echo 'unknown mechanism subject was accepted' >&2
  exit 1
fi
echo 'ok unknown-scan-subject'

echo 'ok task-zero-dod'
