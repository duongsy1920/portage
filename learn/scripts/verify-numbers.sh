#!/usr/bin/env bash
# Re-count every figure in src/data/portage.ts against the Go repository.
#
# Why this exists: two of the ten numbers in GO_USAGE were wrong (assertions
# said 53, the repo has 48; defers said 96, one of them was the word `defer`
# inside a comment). Nobody noticed because nobody re-counted. Now the counting
# is a command, not a memory.
#
# And a third lesson, 06/10: `channels` counted `chan ` and found 0, which was
# true and still misleading — the code READS from four channels (ctx.Done(),
# ticker.C) without declaring one. A check counts what it is told to count; the
# thing it is told to count can itself be the wrong thing. Hence two numbers.
#
# Run from learn/.  Exit code 1 if anything drifted.
set -uo pipefail
# Resolve both paths BEFORE cd, or the relative one stops pointing anywhere.
LEARN="$(cd "$(dirname "$0")/.." && pwd)"
REPO="$(cd "$LEARN/.." && pwd)"
DATA="$LEARN/src/data/portage.ts"
cd "$REPO"

bad=0; total=0
check() { # <key> <counted>
  local key=$1 counted=$2
  total=$((total+1))
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
# Same, but a line that is only a comment does not count — `chan` and `<-` are
# ordinary words in prose, and a comment is not a channel.
code() { grep -rnE "$1" --include='*.go' internal/ cmd/ | grep -v _test | grep -vE ':\s*//' | wc -l | tr -d ' '; }

echo "Đếm lại GO_USAGE trên $REPO"
check ctx              "$(src 'context.Context')"
check mutex            "$(srcE 'sync\.(RW)?Mutex')"
check defers           "$(srcE '^\s*defer ')"
check wrap             "$(src '%w')"
check errorsIs         "$(src 'errors.Is')"
check assertions       "$(grep -rn 'var _ .* = (\*' --include='*.go' internal/ cmd/ | wc -l | tr -d ' ')"
check goroutines       "$(srcE '^\s*go [a-zA-Z]')"
check channelsDeclared "$(code '\bchan\b')"
check channelsRead     "$(code '<-')"

# Episodes 11–16 (06/10). tst() counts in test files only.
tst() { grep -rnE "$1" --include='*_test.go' internal/ cmd/ | grep -vE ':\s*//' | wc -l | tr -d ' '; }
check inTxCalls        "$(code '\.InTx\(')"
check funcLiterals     "$(grep -rnE 'func\(' --include='*.go' internal/ cmd/ | grep -v _test | grep -vE ':\s*//' | grep -vE ':[0-9]+:\s*func ' | grep -E '\{\s*$|\{ *\}' | wc -l | tr -d ' ')"
check namedReturns     "$(code '\(err error\) \{|, err error\) \{')"
check typeSwitchCases  "$(awk '/switch e := ev.\(type\)/,/^}/' internal/adapter/eventcodec/codec.go | grep -c $'^\tcase ')"
check anyUses          "$(code '\bany\b|interface\{\}')"
check typeAssertOk     "$(code ', ok := .*\.\([A-Za-z*.]+\)')"
check generics         "$(code 'func [a-zA-Z]+\[[A-Za-z]+ (any|comparable)\]')"
check selects          "$(code '^\s*select \{')"
check waitGroupsInTests "$(tst 'sync\.WaitGroup')"
check chansInTests     "$(tst 'make\(chan')"
check jsonTags         "$(code 'json:"')"
check jsonFiles        "$(grep -rl '"encoding/json"' --include='*.go' internal/ cmd/ | grep -v _test | wc -l | tr -d ' ')"
check contractTypes    "$(grep -rhoE 'type [A-Za-z]+V1 struct' internal/contracts/ | wc -l | tr -d ' ')"
check testFiles        "$(find internal cmd -name '*_test.go' | wc -l | tr -d ' ')"
check testFuncs        "$(grep -rhE '^func Test' --include='*_test.go' internal cmd | wc -l | tr -d ' ')"
check subtests         "$(tst 't\.Run\(')"
check tableTests       "$(tst 'for _, (tc|c|tt) := range')"
check benchmarks       "$(grep -rhE '^func Benchmark' --include='*_test.go' internal cmd | wc -l | tr -d ' ')"
check httptestFiles    "$(grep -rl 'httptest' --include='*_test.go' internal cmd | wc -l | tr -d ' ')"
# REPO.testsNoDocker: the suite itself, without a database (~2 s). -count=1: a cached run is not a run.
check testsNoDocker    "$(env -u PORTAGE_TEST_DSN go test ./... -count=1 -v 2>/dev/null | grep -c '^--- PASS')"
check recovers         "$(grep -rn 'recover()' --include='*.go' internal/ cmd/ | grep -v _test | grep -v '//' | wc -l | tr -d ' ')"
check valueReceivers   "$(grep -rhoE '^func \([a-z]+ [A-Z][A-Za-z]*\)' --include='*.go' internal/domain/ | wc -l | tr -d ' ')"
check pointerReceivers "$(grep -rhoE '^func \([a-z]+ \*[A-Z][A-Za-z]*\)' --include='*.go' internal/domain/ | wc -l | tr -d ' ')"

# A number typed by hand goes stale without anyone noticing: 278 lived on in three scenes after the count became 266.
stale=$(grep -rnE '\b278\b' "$(dirname "$0")/../src/videos" | wc -l | tr -d ' ')
if [ "$stale" -ne 0 ]; then bad=$((bad+1)); echo "  ✗ src/videos còn $stale chỗ viết chết '278' — dùng REPO.testsNoDocker"; else echo "  không còn '278' viết chết trong src/videos   ok"; fi

echo
if [ "$bad" -eq 0 ]; then echo "$total con số, khớp hết."; else echo "$bad con số lệch — sửa src/data/portage.ts rồi render lại."; fi
exit $((bad > 0))
