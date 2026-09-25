#!/usr/bin/env bash
# verify-agent.sh — mechanical audit of agent/ (design: docs/CODING-AGENT.md §6.7).
# Every check prints OK/FAIL with the count it measured. Exit 1 if any check fails.
# The script knows nothing about any project's language: it only checks shape,
# closed vocabularies, and that shared files stay free of project-specific names.
set -u
cd "$(dirname "$0")/.." || exit 2          # agent/
FAIL=0; N=0
ok()   { N=$((N+1)); printf '[%02d] %-58s OK   %s\n'   "$N" "$1" "${2:-}"; }
fail() { N=$((N+1)); printf '[%02d] %-58s FAIL %s\n' "$N" "$1" "${2:-}"; FAIL=$((FAIL+1)); }

# Files that must stay project-neutral (rules), as opposed to data files
# (knowledge/, learning/, logs/, projects/) that may name a project.
RULE_FILES=(CONTRACT.md ROLE.md CAPABILITIES.md README.md routines/*/SKILL.md decisions/*.md)
ROUTINES=(routines/*/)

# 01 — every routine folder has SKILL.md whose frontmatter name equals the folder
bad=0; n=0
for d in "${ROUTINES[@]}"; do
  r=$(basename "$d"); n=$((n+1))
  [ -f "$d/SKILL.md" ] || { bad=$((bad+1)); echo "     missing $d/SKILL.md"; continue; }
  nm=$(awk '/^name:/{print $2; exit}' "$d/SKILL.md")
  [ "$nm" = "$r" ] || { bad=$((bad+1)); echo "     $d name: '$nm' != '$r'"; }
done
[ $bad -eq 0 ] && ok "routines: SKILL.md có, name khớp thư mục" "($n/$n)" || fail "routines: SKILL.md có, name khớp thư mục" "($bad lệch)"

# 02 — routine set == {intake} ∪ CONTRACT step vocabulary; nothing extra, nothing missing
want=$(grep -m1 -E '^step:' CONTRACT.md | sed -E 's/^step:[[:space:]]*//; s/#.*//' | tr '|' '\n' | tr -d ' ' | grep . ; echo intake)
have=$(for d in "${ROUTINES[@]}"; do basename "$d"; done)
missing=$(comm -23 <(echo "$want" | sort -u) <(echo "$have" | sort -u))
extra=$(comm -13 <(echo "$want" | sort -u) <(echo "$have" | sort -u))
if [ -z "$missing$extra" ]; then ok "routines: đúng tập CONTRACT khai (intake + step)" "($(echo "$want" | wc -l))"
else fail "routines: đúng tập CONTRACT khai" "thiếu[$(echo $missing)] dư[$(echo $extra)]"; fi

# 03 — Step 0 lines 0.0..0.4 present in every routine
bad=0
for d in "${ROUTINES[@]}"; do for k in 0.0 0.1 0.2 0.3 0.4; do
  grep -qE "^$k  " "$d/SKILL.md" || { bad=$((bad+1)); echo "     $d thiếu dòng $k"; }
done; done
[ $bad -eq 0 ] && ok "routines: Step 0 đủ năm dòng 0.0–0.4" "($(( ${#ROUTINES[@]} * 5 )) dòng)" || fail "routines: Step 0 đủ năm dòng" "($bad thiếu)"

# 04 — every routine has the fixed tail sections and mentions R1–R4
bad=0
for d in "${ROUTINES[@]}"; do
  for h in '^## Cuối run' '^## Không bao giờ ghi ra' '^## Corrections'; do
    grep -qE "$h" "$d/SKILL.md" || { bad=$((bad+1)); echo "     $d thiếu $h"; }
  done
  grep -qE 'R1–R4|R1-R4' "$d/SKILL.md" || { bad=$((bad+1)); echo "     $d không nhắc R1–R4"; }
done
[ $bad -eq 0 ] && ok "routines: có Cuối run · Không bao giờ ghi ra · Corrections · R1–R4" || fail "routines: đuôi cố định" "($bad thiếu)"

# 05 — leader mindset: every routine carries at least one recommendation (CONTRACT §5.0)
bad=0
for d in "${ROUTINES[@]}"; do grep -qi 'đề nghị' "$d/SKILL.md" || { bad=$((bad+1)); echo "     $d không có 'đề nghị'"; }; done
[ $bad -eq 0 ] && ok "routines: mỗi routine có 'đề nghị' (§5.0)" || fail "routines: mỗi routine có 'đề nghị'" "($bad thiếu)"

# 06 — rule files name no language, tool, framework, or project
pat='\bGo\b|\bPHP\b|golang|symfony|phpunit|gofmt|go test|go vet|\bnpm\b|composer|portage|/home/'
hits=$(grep -nE "$pat" "${RULE_FILES[@]}" 2>/dev/null | grep -vE 'name-of-a-|<tên' || true)
[ -z "$hits" ] && ok "file luật không nhắc ngôn ngữ/công cụ/project" "($(printf '%s\n' "${RULE_FILES[@]}" | wc -l) file)" || { fail "file luật không nhắc ngôn ngữ/công cụ/project"; echo "$hits" | sed 's/^/     /'; }

# 07 — absolute paths only inside PROJECT.md
hits=$(grep -rnE '(^|[^A-Za-z0-9_./-])/(home|Users|opt|var|usr|mnt|srv|tmp|etc)/' --include=*.md . | grep -v '/PROJECT.md:' || true)
[ -z "$hits" ] && ok "đường dẫn tuyệt đối chỉ ở PROJECT.md" || { fail "đường dẫn tuyệt đối chỉ ở PROJECT.md"; echo "$hits" | sed 's/^/     /'; }

# 08 — PROJECT.md schema (CONTRACT §2): name==dir · path exists · rules exists · commands.test · branches.default · releases
bad=0; n=0
for pf in projects/*/PROJECT.md; do
  n=$((n+1)); dir=$(basename "$(dirname "$pf")")
  v() { awk -v k="$1" '$0 ~ "^"k":" {sub("^"k":[[:space:]]*",""); sub("[[:space:]]*#.*",""); print; exit}' "$pf"; }
  nm=$(v name); path=$(v path); rules=$(v rules)
  [ "$nm" = "$dir" ] || { bad=$((bad+1)); echo "     $pf name '$nm' != dir '$dir'"; }
  [ -d "$path" ] || { bad=$((bad+1)); echo "     $pf path không tồn tại: $path"; }
  [ -f "$path/$rules" ] || { bad=$((bad+1)); echo "     $pf rules không tồn tại: $path/$rules"; }
  grep -qE '^  test:[[:space:]]*[^[:space:]]' "$pf" || { bad=$((bad+1)); echo "     $pf thiếu commands.test"; }
  grep -qE '^  default:[[:space:]]*[^[:space:]]' "$pf" || { bad=$((bad+1)); echo "     $pf thiếu branches.default"; }
  grep -qE '^  prefix:[[:space:]]*[^[:space:]]' "$pf" || { bad=$((bad+1)); echo "     $pf thiếu branches.prefix"; }
  grep -qE '^releases:' "$pf" || { bad=$((bad+1)); echo "     $pf thiếu releases"; }
done
[ $bad -eq 0 ] && ok "PROJECT.md đúng schema §2" "($n project)" || fail "PROJECT.md đúng schema §2" "($bad lỗi)"

# 09 — every capability a routine or CONTRACT calls exists in CAPABILITIES.md; unused ones are warned
declared=$(grep -oE '^\| *`[a-z]+\.[a-z]+`' CAPABILITIES.md | tr -d '|` ' | sort -u)
used=$(grep -ohE '`(file|code|shell|test|git|secret)\.[a-z]+`' routines/*/SKILL.md CONTRACT.md | tr -d '`' | sort -u)
undeclared=$(comm -13 <(echo "$declared") <(echo "$used"))
unused=$(comm -23 <(echo "$declared") <(echo "$used"))
[ -z "$undeclared" ] && ok "capability được gọi đều có khai" "($(echo "$used" | wc -l) dùng / $(echo "$declared" | wc -l) khai)" || fail "capability được gọi đều có khai" "chưa khai: $(echo $undeclared)"
[ -n "$unused" ] && echo "     (cảnh báo: khai mà chưa routine nào gọi: $(echo $unused))"

# 10 — closed vocabularies wherever they appear in rule files
bad=0
while IFS= read -r w; do
  case "$w" in plan-approval|commit-approval|push-approval|question) ;; *) bad=$((bad+1)); echo "     waiting_for lạ: $w";; esac
done < <(grep -ohE 'WAITING_HUMAN */ *[a-z-]+' "${RULE_FILES[@]}" | sed -E 's#WAITING_HUMAN */ *##' | sort -u)
while IFS= read -r w; do
  case "$w" in TODO|IN_PROGRESS|WAITING_HUMAN|DONE|BLOCKED|REJECTED|FAILED) ;; *) bad=$((bad+1)); echo "     status lạ: $w";; esac
done < <(grep -ohE 'status: *[A-Z_]+' "${RULE_FILES[@]}" | sed -E 's/status: *//' | sort -u)
[ $bad -eq 0 ] && ok "từ vựng đóng: status (7) · waiting_for (4)" || fail "từ vựng đóng" "($bad giá trị lạ)"

# 11 — every TASK.md: frontmatter consistent with its folder and the closed vocabularies; plan_digest matches plan.md (T7)
bad=0; n=0
for tf in projects/*/tasks/*/TASK.md; do
  [ -f "$tf" ] || continue; n=$((n+1))
  tdir=$(basename "$(dirname "$tf")"); proj=$(basename "$(dirname "$(dirname "$(dirname "$tf")")")")
  fm=$(awk 'NR==1&&/^---$/{f=1;next} f&&/^---$/{exit} f' "$tf")
  g() { echo "$fm" | awk -v k="$1" '$0 ~ "^"k":" {sub("^"k":[[:space:]]*",""); sub("[[:space:]]*#.*",""); print; exit}'; }
  id=$(g id); slug=$(g slug); st=$(g status); step=$(g step); wf=$(g waiting_for); vf=$(g verify); pj=$(g project); br=$(g branch)
  [ "$tdir" = "$id-$slug" ] || { bad=$((bad+1)); echo "     $tf: thư mục '$tdir' != '$id-$slug'"; }
  [ "$pj" = "$proj" ] || { bad=$((bad+1)); echo "     $tf: project '$pj' != '$proj'"; }
  prefix=$(awk '/^  prefix:/{print $2; exit}' "projects/$proj/PROJECT.md")
  [ "$br" = "$prefix$id-$slug" ] || { bad=$((bad+1)); echo "     $tf: branch '$br' != '$prefix$id-$slug' (T6)"; }
  case "$st" in TODO|IN_PROGRESS|WAITING_HUMAN|DONE|BLOCKED|REJECTED|FAILED) ;; *) bad=$((bad+1)); echo "     $tf: status lạ '$st'";; esac
  case "$vf" in not-run|passed|gate-failed) ;; *) bad=$((bad+1)); echo "     $tf: verify lạ '$vf'";; esac
  case "$st" in DONE|REJECTED|FAILED) [ -z "$step" ] || { bad=$((bad+1)); echo "     $tf: $st không có step"; };;
    *) case "$step" in analyze|plan|implement|verify|review) ;; *) bad=$((bad+1)); echo "     $tf: $st cần step hợp lệ, có '$step'";; esac;; esac
  if [ "$st" = WAITING_HUMAN ]; then case "$wf" in plan-approval|commit-approval|push-approval|question) ;; *) bad=$((bad+1)); echo "     $tf: WAITING_HUMAN cần waiting_for hợp lệ, có '$wf'";; esac
  else [ -z "$wf" ] || { bad=$((bad+1)); echo "     $tf: waiting_for chỉ khi WAITING_HUMAN"; }; fi
  if [ "$st" = DONE ]; then
    [ -f "$(dirname "$tf")/review.md" ] && [ "$vf" = passed ] && echo "$fm" | grep -q 'gate: *commit' || { bad=$((bad+1)); echo "     $tf: DONE thiếu review.md / verify passed / approvals.commit (T2)"; }
  fi
  pd=$(g plan_digest)
  if [ -n "$pd" ] && [ -f "$(dirname "$tf")/plan.md" ]; then
    real=$(sha256sum "$(dirname "$tf")/plan.md" | cut -d' ' -f1)
    [ "$pd" = "$real" ] || { bad=$((bad+1)); echo "     $tf: plan_digest lệch plan.md (T7)"; }
  fi
done
[ $bad -eq 0 ] && ok "TASK.md: thư mục · project · branch · từ vựng · T2 · T7" "($n task)" || fail "TASK.md nhất quán" "($bad lỗi)"

# 12 — run records: name pattern, run==filename, closed status, invariants line, no diff/trace/secret-ish content
bad=0; n=0
for rf in logs/runs/*.md; do
  [ -f "$rf" ] || continue; n=$((n+1)); b=$(basename "$rf" .md)
  echo "$b" | grep -qE '^[0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9]{4}-T-[0-9]{3}-[a-z]+$' || { bad=$((bad+1)); echo "     $rf: tên không đúng mẫu"; }
  grep -qE "^run:[[:space:]]+$b\$" "$rf" || { bad=$((bad+1)); echo "     $rf: run: != tên file"; }
  grep -qE '^status:[[:space:]]+(ok|partial|failed|skipped-paused|blocked)$' "$rf" || { bad=$((bad+1)); echo "     $rf: status ngoài từ vựng"; }
  grep -qE '^invariants:' "$rf" || { bad=$((bad+1)); echo "     $rf: thiếu invariants"; }
  grep -qE '^(\+\+\+ |--- a/|@@ |diff --git|goroutine [0-9]+ \[|panic: |Traceback)' "$rf" && { bad=$((bad+1)); echo "     $rf: chứa diff/trace"; }
done
[ $bad -eq 0 ] && ok "run record: tên · run · status · invariants · không diff/trace" "($n record)" || fail "run record" "($bad lỗi)"

# 13 — nothing that looks like a credential anywhere under agent/
hits=$(grep -rnE 'AKIA[0-9A-Z]{16}|-----BEGIN [A-Z ]*PRIVATE KEY|ghp_[A-Za-z0-9]{36}|sk-[A-Za-z0-9]{20,}|[a-z]+://[^/[:space:]]+:[^@[:space:]]+@|(password|passwd|secret|token)[[:space:]]*[:=][[:space:]]*["'\''][^"'\'']{4,}' . 2>/dev/null | grep -v 'verify-agent.sh' || true)
[ -z "$hits" ] && ok "không chuỗi giống credential trong agent/" || { fail "không chuỗi giống credential"; echo "$hits" | sed 's/^/     /'; }

# 14 — agent/ holds only Markdown, this script, and PAUSED: no source files of any language
hits=$(find . -type f ! -name '*.md' ! -path './scripts/*.sh' ! -name PAUSED ! -name '.gitkeep' ! -path './projects/*/tasks/*/shots/*.png' | sort)
[ -z "$hits" ] && ok "agent/ chỉ có .md, scripts/*.sh, PAUSED, tasks/*/shots/*.png" "($(find . -type f | wc -l) file)" || { fail "agent/ chỉ có .md, scripts/*.sh, PAUSED, shots/*.png"; echo "$hits" | sed 's/^/     /'; }

# 15 — every ADR has a "Xem lại khi" section; top-level rule files end with ## Corrections
bad=0
for a in decisions/ADR-*.md; do grep -q '^## Xem lại khi' "$a" || { bad=$((bad+1)); echo "     $a thiếu 'Xem lại khi'"; }; done
for f in CONTRACT.md ROLE.md CAPABILITIES.md; do grep -q '^## Corrections' "$f" || { bad=$((bad+1)); echo "     $f thiếu ## Corrections"; }; done
[ $bad -eq 0 ] && ok "ADR có 'Xem lại khi' · file luật có ## Corrections" "($(ls decisions/ADR-*.md | wc -l) ADR)" || fail "ADR / Corrections" "($bad thiếu)"

# 16 — the harness entry point (one skill) names every routine, and nothing else claims to be one
skill=../.claude/skills/agent/SKILL.md
if [ -f "$skill" ]; then
  bad=0; for d in "${ROUTINES[@]}"; do r=$(basename "$d"); grep -qw "$r" "$skill" || { bad=$((bad+1)); echo "     skill không nhắc routine $r"; }; done
  [ $bad -eq 0 ] && ok "skill /agent nhắc đủ mọi routine" || fail "skill /agent nhắc đủ mọi routine" "($bad thiếu)"
else fail "skill /agent tồn tại" "không thấy $skill"; fi

echo
if [ $FAIL -eq 0 ]; then echo "verify-agent: $N/$N OK"; else echo "verify-agent: $FAIL/$N FAIL"; exit 1; fi
