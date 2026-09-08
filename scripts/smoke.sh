#!/usr/bin/env bash
# scripts/smoke.sh — the whole flow against REAL binaries and a REAL Postgres,
# the Linux twin of scripts/smoke.ps1 so CI runs the same thing the laptop does.
#
# A green `go test ./...` proves the pieces work when a test wires them. This
# proves the BINARIES work when systemd does: two processes, a real database,
# real bearer tokens, a real relay between them.
#
#   docker compose up -d                       # or a CI service container
#   PORTAGE_DSN=postgres://… scripts/smoke.sh
set -euo pipefail

DSN="${PORTAGE_DSN:-postgres://portage:portage@localhost:5432/portage?sslmode=disable}"
# PORTAGE_PORT lets the smoke run on a box where something already holds 8080,
# the same reason smoke.ps1 takes -Port. BASE stays overridable on its own for
# the case where the api is already running somewhere else.
PORT="${PORTAGE_PORT:-8080}"
BASE="${PORTAGE_BASE:-http://localhost:$PORT}"
mkdir -p bin

go build -o bin/api ./cmd/api
go build -o bin/worker ./cmd/worker

# The first operator key is cut at start-up by -bootstrap-operator-token — a
# fixed credential in a database that already has users would be a back door.
# The bootstrap only fires while api_tokens is EMPTY, so the smoke starts from
# no keys at all; without this the script works exactly once per database.
wipe_tokens() {
  local sql="DO \$\$ BEGIN IF to_regclass('api_tokens') IS NOT NULL THEN TRUNCATE api_tokens; END IF; END \$\$;"
  if command -v psql >/dev/null; then
    psql "$DSN" -qAt -c "$sql" >/dev/null
  else
    # No psql on the host (a Windows laptop): go through the compose container
    # instead, so this script is runnable in both places rather than only in CI.
    docker compose exec -T postgres psql -q -U portage -d "${PORTAGE_DB:-portage}" -c "$sql" >/dev/null
  fi
}
wipe_tokens

OPTOK="smoke-operator-$RANDOM$RANDOM"
./bin/api -addr ":$PORT" -dsn "$DSN" -bootstrap-operator-token "$OPTOK" >bin/api.log 2>&1 &
API=$!
./bin/worker -dsn "$DSN" -every 300ms >bin/worker.log 2>&1 &
WORKER=$!
trap 'kill "$API" "$WORKER" 2>/dev/null || true; wait 2>/dev/null || true' EXIT

# Wait for the port rather than sleeping a guess: a slow migration on a cold
# database is the difference between a green run and a mystery failure.
for _ in $(seq 1 50); do
  # -S dropped on purpose: the first probes fail while the api is still
  # binding, and printing "Failed to connect" there makes a healthy run look
  # broken. A real failure surfaces when the loop runs out and the next curl
  # (which does use -S) reports it.
  curl -fs -o /dev/null "$BASE/tokens" -H "Authorization: Bearer $OPTOK" && break
  sleep 0.2
done

OP=(-H "Authorization: Bearer $OPTOK" -H 'Content-Type: application/json')
# The operator cuts the customer a key. It comes back ONCE — the store keeps
# only a hash — which is why it is captured here and never re-read.
CUSTOK=$(curl -fsS -X POST "$BASE/tokens" "${OP[@]}" -d '{"kind":"customer","label":"smoke"}' | jq -r .token)
CU=(-H "Authorization: Bearer $CUSTOK" -H 'Content-Type: application/json')

SITE="www.example-$RANDOM.com"
curl -fsS -X POST "$BASE/fx" "${OP[@]}" -d '{"from":"USD","to":"VND","rate":"26000"}' >/dev/null
MERCHANT=$(curl -fsS -X POST "$BASE/merchants" "${OP[@]}" \
  -d "{\"name\":\"Example Sports\",\"site\":\"$SITE\",\"currency\":\"USD\",\"sourcing\":[\"operator\",\"customer\"]}" | jq -r .id)
PRODUCT=$(curl -fsS -X POST "$BASE/products" "${CU[@]}" \
  -d "{\"name\":\"Air Trainer 90\",\"merchant_id\":\"$MERCHANT\",\"category\":\"footwear\",\"source_url\":\"https://$SITE/t/air-trainer-90/abc\",\"price\":\"150.00\",\"currency\":\"USD\"}" | jq -r .id)
echo "merchant $MERCHANT / product $PRODUCT"

# The size is announced here, before publish, and its id is kept: ordering
# only takes a variant id it has heard of through catalog.variant_added, so a
# size invented at order time is a 409, not an order. One variant on purpose —
# the golden flow is exactly 20 events, the same count web/flow.html walks.
VARIANT=$(curl -fsS -X POST "$BASE/products/$PRODUCT/variants" "${OP[@]}" -d '{"size":"US 9","color":"black","merchant_ref":"EX-AT90-9-BLK"}' | jq -r .id)
curl -fsS -X POST "$BASE/products/$PRODUCT/confirm-listing" "${OP[@]}" >/dev/null
curl -fsS -X POST "$BASE/products/$PRODUCT/measure" "${OP[@]}" -d '{"weight_g":1250,"length_mm":340,"width_mm":230,"height_mm":130}' >/dev/null
curl -fsS -X POST "$BASE/products/$PRODUCT/publish" "${OP[@]}" >/dev/null
sleep 1 # the worker relays product_published → pricing's listing

QUOTE=$(curl -fsS -X POST "$BASE/quotes" "${CU[@]}" -d "{\"product_id\":\"$PRODUCT\",\"lane\":\"us_forwarder\"}" | jq -r .id)
VIEW=$(curl -fsS "$BASE/quotes/$QUOTE" "${CU[@]}")
DEPOSIT=$(echo "$VIEW" | jq -r .home.deposit.amount)
echo "quote $QUOTE: $(echo "$VIEW" | jq -r '.status + " " + .class') chargeable $(echo "$VIEW" | jq -r .chargeable_g) g, total $(echo "$VIEW" | jq -r .home.total.amount) VND, deposit $DEPOSIT"
curl -fsS -X POST "$BASE/quotes/$QUOTE/accept" "${CU[@]}" >/dev/null
sleep 1 # quote_accepted → ordering's projection

ORDER=$(curl -fsS -X POST "$BASE/orders" "${CU[@]}" -d "{\"quote_id\":\"$QUOTE\",\"variant_id\":\"$VARIANT\"}" | jq -r .id)
curl -fsS -X POST "$BASE/orders/$ORDER/deposit" "${OP[@]}" -d "{\"amount\":\"$DEPOSIT\",\"currency\":\"VND\"}" >/dev/null
sleep 1 # deposit_paid → procurement opens a task

# The whole row, because the buyer's screen is the point of the subject: a
# name, a size, the shop's own code and the page — copied onto the task when it
# opened, from procurement's projection of catalog.variant_added.
TASKROW=$(curl -fsS "$BASE/purchase-tasks" "${OP[@]}" | jq -r '.[0]')
TASK=$(echo "$TASKROW" | jq -r .id)
LABEL=$(echo "$TASKROW" | jq -r '.variant_label // ""')
echo "order $ORDER: deposited; buyer's list has task $TASK — buy $(echo "$TASKROW" | jq -r '.product_name + " / " + (.variant_label // "-") + " / ref " + (.variant_ref // "-") + " / " + (.source // "-")')"
curl -fsS -X POST "$BASE/purchase-tasks/$TASK/confirm" "${OP[@]}" -d '{"reference":"NK-20260905-001","paid":"163.22","currency":"USD"}' >/dev/null
sleep 1

PARCEL=$(curl -fsS "$BASE/parcels" "${OP[@]}" | jq -r '.[0].id')
echo "order after purchase_confirmed: $(curl -fsS "$BASE/orders/$ORDER" "${OP[@]}" | jq -r .status); warehouse expects parcel $PARCEL"
curl -fsS -X POST "$BASE/parcels/$PARCEL/receive" "${OP[@]}" -d '{"weight_g":1250,"length_mm":340,"width_mm":230,"height_mm":130}' >/dev/null
BATCH=$(curl -fsS -X POST "$BASE/batches" "${OP[@]}" -d '{"lane":"us_forwarder"}' | jq -r .id)
curl -fsS -X POST "$BASE/batches/$BATCH/parcels" "${OP[@]}" -d "{\"parcel_id\":\"$PARCEL\"}" >/dev/null
curl -fsS -X POST "$BASE/batches/$BATCH/close" "${OP[@]}" >/dev/null
ALLOCS=$(curl -fsS -X POST "$BASE/batches/$BATCH/ship" "${OP[@]}" -d '{"freight":"27.50","currency":"USD"}')
echo "batch $BATCH shipped: billed on $(echo "$ALLOCS" | jq -r '.[0].chargeable_g') g → freight $(echo "$ALLOCS" | jq -r '.[0].freight.amount') USD"
sleep 1

curl -fsS -X POST "$BASE/orders/$ORDER/balance" "${OP[@]}" -d "{\"amount\":\"$DEPOSIT\",\"currency\":\"VND\"}" >/dev/null
curl -fsS -X POST "$BASE/orders/$ORDER/deliver" "${OP[@]}" >/dev/null
REC=$(curl -fsS "$BASE/reconciliations/$ORDER" "${OP[@]}")
STATUS=$(curl -fsS "$BASE/orders/$ORDER" "${OP[@]}" | jq -r .status)
VARIANCE=$(echo "$REC" | jq -r .variance.amount)
echo "order at the end: $STATUS; quote vs actual: quoted $(echo "$REC" | jq -r '.quoted.goods.amount + "+" + .quoted.freight.amount'), actual $(echo "$REC" | jq -r '.actual.goods.amount + "+" + .actual.freight.amount') → variance $VARIANCE USD"
sleep 1

# The READ side: one flat row per order, from the events of five contexts.
MINE=$(curl -fsS "$BASE/me/orders" "${CU[@]}")
echo "my orders: $(echo "$MINE" | jq -r 'length') row(s), $(echo "$MINE" | jq -r '.[0].product_name + " - " + .[0].status + ", tracking " + .[0].tracking')"

# The GOLDEN NUMBERS. Printing them is documentation; asserting them is the
# test — a smoke that only prints is a smoke nobody notices going wrong.
fail=0
[ "$STATUS" = "delivered" ] || { echo "FAIL: order status $STATUS"; fail=1; }
[ "$VARIANCE" = "-2.50" ] || { echo "FAIL: variance $VARIANCE, want -2.50"; fail=1; }
[ "$DEPOSIT" = "2696860" ] || { echo "FAIL: deposit $DEPOSIT, want 2696860"; fail=1; }
[ "$LABEL" = "US 9 · black" ] || { echo "FAIL: the buyer's task says '$LABEL', want 'US 9 · black'"; fail=1; }
[ "$(echo "$MINE" | jq -r '.[0].status')" = "delivered" ] || { echo "FAIL: the read model did not catch up"; fail=1; }
[ "$(echo "$MINE" | jq -r '.[0].shop_reference // ""')" = "" ] || { echo "FAIL: the customer's view leaked the shop reference"; fail=1; }

echo "--- worker log (last 20) ---"
tail -20 bin/worker.log | cut -c1-160
exit "$fail"
