#!/usr/bin/env bash
# Render every cheat-sheet plate to PNG (@2x) and PDF (A3 landscape).
#
# One entry per episode. The output directory matches the episode's outDir in
# its spec.ts, so a plate always lands beside the video it belongs to.
set -euo pipefail
cd "$(dirname "$0")/../cheatsheet"

render() { # <html> <outDir>
  local html=$1 dir=../out/$2
  mkdir -p "$dir"
  google-chrome --headless --disable-gpu --hide-scrollbars \
    --screenshot="$dir/cheatsheet.png" --window-size=1587,1123 --force-device-scale-factor=2 \
    "file://$PWD/$html" >/dev/null 2>&1
  google-chrome --headless --disable-gpu --no-pdf-header-footer \
    --print-to-pdf="$dir/cheatsheet.pdf" "file://$PWD/$html" >/dev/null 2>&1
  printf "%-22s → %-34s %s\n" "$html" "$2" "$(du -h "$dir/cheatsheet.png" | cut -f1)"
}

render go-ep01.html       mua1-go/ep01-vong-doi
render go-ep02.html       mua1-go/ep02-doc-mot-dong
render go-ep03.html       mua1-go/ep03-loi-la-gia-tri
render go-ep04.html       mua1-go/ep04-receiver
render go-ep05.html       mua1-go/ep05-interface-ngam
render go-ep06.html       mua1-go/ep06-zero-value
render go-ep07.html       mua1-go/ep07-embedding
render go-ep08.html       mua1-go/ep08-slice-map
render go-ep09.html       mua1-go/ep09-defer-panic
render go-ep10.html       mua1-go/ep10-dong-thoi

render ddd-ep01.html      mua2-ddd/ep01-thieu-mau
render ddd-ep02.html      mua2-ddd/ep02-value-object
render ddd-ep03.html      mua2-ddd/ep03-aggregate
render ddd-ep04.html      mua2-ddd/ep04-repository
render ddd-ep05.html      mua2-ddd/ep05-bounded-context
render ddd-ep06.html      mua2-ddd/ep06-domain-event

render portage-ep01.html  mua3-portage/ep01-nam-vung
render portage-ep02.html  mua3-portage/ep02-outbox
render portage-ep03.html  mua3-portage/ep03-nhat-quan
render portage-ep04.html  mua3-portage/ep04-port-adapter
render portage-ep05.html  mua3-portage/ep05-khong-quay-dau
render portage-ep06.html  mua3-portage/ep06-bao-gia
render portage-ep07.html  mua3-portage/ep07-read-model
render portage-ep08.html  mua3-portage/ep08-acl
