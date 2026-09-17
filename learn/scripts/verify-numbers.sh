#!/usr/bin/env bash
# Re-count every figure in src/data/portage.ts against the Go repository.
#
# Why this exists: two of the ten numbers in GO_USAGE were wrong (assertions
# said 53, the repo has 48; defers said 96, one of them was the word `defer`
# inside a comment). Nobody noticed because nobody re-counted. Now the counting
# is a command, not a memory.
#
# Run from learn/.  Exit code 1 if anything drifted.
set -uo pipefail
# Resolve both paths BEFORE cd, or the relative one stops pointing anywhere.
LEARN="$(cd "$(dirname "$0")/.." && pwd)"
REPO="$(cd "$LEARN/.." && pwd)"
DATA="$LEARN/src/data/portage.ts"
cd "$REPO"

bad=0
check() { # <key> <counted>
  local key=$1 counted=$2
  local declared
  declared=$(grep -oE "^  $key: [0-9]+," "$DATA" | grep -oE '[0-9]+')
  if [ -z "$declared" ]; then
    printf "  %-16s KHÔNG THẤY trong portage.ts\n" "$key"; bad=$((bad+1)); return
  fi
  if [ "$declared" = "$counted" ]; then
    printf "  %-16s %-5s ok\n" "$key" "$declared"
  else
    printf "  %-16s khai báo %-5s đếm lại %-5s  ← LỆCH\n" "$key" "$declared" "$counted"; bad=$((bad+1))
  fi
}

src() { grep -rn "$1" --include='*.go' internal/ cmd/ | grep -v _test | wc -l | tr -d ' '; }
srcE() { grep -rnE "$1" --include='*.go' internal/ cmd/ | grep -v _test | wc -l | tr -d ' '; }

echo "Đếm lại GO_USAGE trên $REPO"
check ctx              "$(src 'context.Context')"
check mutex            "$(srcE 'sync\.(RW)?Mutex')"
check defers           "$(srcE '^\s*defer ')"
check wrap             "$(src '%w')"
check errorsIs         "$(src 'errors.Is')"
check assertions       "$(grep -rn 'var _ .* = (\*' --include='*.go' internal/ cmd/ | wc -l | tr -d ' ')"
check goroutines       "$(srcE '^\s*go [a-zA-Z]')"
check channels         "$(src 'chan ')"
check recovers         "$(grep -rn 'recover()' --include='*.go' internal/ cmd/ | grep -v _test | grep -v '//' | wc -l | tr -d ' ')"
check valueReceivers   "$(grep -rhoE '^func \([a-z]+ [A-Z][A-Za-z]*\)' --include='*.go' internal/domain/ | wc -l | tr -d ' ')"
check pointerReceivers "$(grep -rhoE '^func \([a-z]+ \*[A-Z][A-Za-z]*\)' --include='*.go' internal/domain/ | wc -l | tr -d ' ')"

echo
if [ "$bad" -eq 0 ]; then echo "11 con số, khớp hết."; else echo "$bad con số lệch — sửa src/data/portage.ts rồi render lại."; fi
exit $((bad > 0))
