# scripts/smoke.ps1 — the whole flow against REAL binaries and a REAL Postgres:
#   api (-dsn) + worker (-dsn) in the background, curl through catalog → pricing → ordering → procurement → logistics → ordering → pricing (Quote vs Actual).
# Needs: docker compose up -d (docs/SETUP.md §5d). Run from the repo root:
#   powershell -ExecutionPolicy Bypass -File scripts\smoke.ps1
# -Port lets the smoke run beside a console API that already holds 8080
# (cmd/api -web). Nothing else in the script cares which port it is.
param([int]$Port = 8080)
$ErrorActionPreference = "Stop"
$dsn = "postgres://portage:portage@localhost:5432/portage?sslmode=disable"
$base = "http://localhost:$Port"

go build -o bin\api.exe ./cmd/api
go build -o bin\worker.exe ./cmd/worker
# The first operator key is cut at start-up by -bootstrap-operator-token —
# a fixed credential in a database that already has users would be a back
# door. Every later key comes from POST /tokens, which is what the customer
# below uses...
# ...and the bootstrap only fires while api_tokens is EMPTY, so the smoke
# starts from no keys at all. Without this wipe the script works exactly once
# per database: the second run bootstraps nothing and every request is 401.
$wipe = @'
DO $$ BEGIN IF to_regclass('api_tokens') IS NOT NULL THEN TRUNCATE api_tokens; END IF; END $$;
'@
docker compose exec -T postgres psql -q -U portage -d portage -c $wipe | Out-Null
$optok = "smoke-operator-" + (Get-Random)
$api = Start-Process -FilePath bin\api.exe -ArgumentList "-addr", ":$Port", "-dsn", $dsn, "-bootstrap-operator-token", $optok -RedirectStandardError bin\api.log -PassThru -NoNewWindow
$worker = Start-Process -FilePath bin\worker.exe -ArgumentList "-dsn", $dsn, "-every", "300ms" -RedirectStandardError bin\worker.log -PassThru -NoNewWindow
try {
    Start-Sleep -Seconds 2
    $site = "www.example-" + (Get-Random) + ".com"
    $h = @{ "Content-Type" = "application/json"; "Authorization" = "Bearer $optok" }
    # The operator cuts the customer a key. It comes back ONCE — the store
    # keeps only a hash — which is why it is captured here and never re-read.
    $custok = (Invoke-RestMethod -Method Post -Uri "$base/tokens" -Headers $h -Body '{"kind":"customer","label":"smoke"}').token
    $ch = @{ "Content-Type" = "application/json"; "Authorization" = "Bearer $custok" }
    # POST /fx: today's rate, published through the same handler the seed uses.
    # Same 26000 on purpose — the golden numbers below must not move.
    Invoke-RestMethod -Method Post -Uri "$base/fx" -Headers $h -Body '{"from":"USD","to":"VND","rate":"26000"}' | Out-Null
    $merchant = (Invoke-RestMethod -Method Post -Uri "$base/merchants" -Headers $h -Body "{`"name`":`"Example Sports`",`"site`":`"$site`",`"currency`":`"USD`",`"sourcing`":[`"operator`",`"customer`"]}").id
    $product = (Invoke-RestMethod -Method Post -Uri "$base/products" -Headers $ch -Body "{`"name`":`"Air Trainer 90`",`"merchant_id`":`"$merchant`",`"category`":`"footwear`",`"source_url`":`"https://$site/t/air-trainer-90/abc`",`"price`":`"150.00`",`"currency`":`"USD`"}").id
    Write-Host "merchant $merchant / product $product"
    $oph = $h   # the operator is the token now; no separate identity header
    # The size is announced HERE, before publish, and its id is kept: ordering
    # only takes a variant id it has heard of through catalog.variant_added, so
    # a size invented at order time is a 409, not an order. One variant on
    # purpose — the golden flow is exactly 20 events, the same count
    # web/flow.html walks through.
    $variant = (Invoke-RestMethod -Method Post -Uri "$base/products/$product/variants" -Headers $h -Body '{"size":"US 9","color":"black","merchant_ref":"EX-AT90-9-BLK"}').id
    Invoke-RestMethod -Method Post -Uri "$base/products/$product/confirm-listing" -Headers $oph | Out-Null
    Invoke-RestMethod -Method Post -Uri "$base/products/$product/measure" -Headers $oph -Body '{"weight_g":1250,"length_mm":340,"width_mm":230,"height_mm":130}' | Out-Null
    Invoke-RestMethod -Method Post -Uri "$base/products/$product/publish" -Headers $h | Out-Null
    Write-Host "published; waiting for the worker to relay product_published..."
    Start-Sleep -Seconds 1
    $quote = (Invoke-RestMethod -Method Post -Uri "$base/quotes" -Headers $h -Body "{`"product_id`":`"$product`",`"lane`":`"us_forwarder`"}").id
    $view = Invoke-RestMethod -Method Get -Uri "$base/quotes/$quote" -Headers $h
    Write-Host ("quote {0}: {1} {2} chargeable {3} g, total {4} {5}, deposit {6}" -f $quote, $view.status, $view.class, $view.chargeable_g, $view.home.total.amount, $view.home.total.currency, $view.home.deposit.amount)
    Invoke-RestMethod -Method Post -Uri "$base/quotes/$quote/accept" -Headers $ch | Out-Null
    Start-Sleep -Seconds 1   # quote_accepted → ordering's projection

    # ordering: place, pay the deposit exactly
    $order = (Invoke-RestMethod -Method Post -Uri "$base/orders" -Headers $ch -Body "{`"quote_id`":`"$quote`",`"variant_id`":`"$variant`"}").id
    Invoke-RestMethod -Method Post -Uri "$base/orders/$order/deposit" -Headers $h -Body "{`"amount`":`"$($view.home.deposit.amount)`",`"currency`":`"VND`"}" | Out-Null
    Start-Sleep -Seconds 1   # deposit_paid → procurement opens a task
    $tasks = Invoke-RestMethod -Method Get -Uri "$base/purchase-tasks" -Headers $h
    Write-Host ("order {0}: deposited; buyer's list: {1} open task(s), first for order {2} in {3}" -f $order, @($tasks).Count, $tasks[0].order_id, $tasks[0].currency)
    # The buyer's screen in words, not uuids: procurement copied these from its
    # own projection of catalog.variant_added when the task opened.
    Write-Host ("buy: {0} / {1} / ref {2} / {3}" -f $tasks[0].product_name, $tasks[0].variant_label, $tasks[0].variant_ref, $tasks[0].source)

    # procurement: the buyer bought it → ordering passes the point of no return; logistics expects a box
    Invoke-RestMethod -Method Post -Uri "$base/purchase-tasks/$($tasks[0].id)/confirm" -Headers $oph -Body '{"reference":"NK-20260905-001","paid":"163.22","currency":"USD"}' | Out-Null
    Start-Sleep -Seconds 1
    $ov = Invoke-RestMethod -Method Get -Uri "$base/orders/$order" -Headers $h
    $parcels = Invoke-RestMethod -Method Get -Uri "$base/parcels" -Headers $h
    Write-Host ("order after purchase_confirmed: {0}; warehouse expects {1} parcel(s), ref {2}" -f $ov.status, @($parcels).Count, $parcels[0].reference)

    # logistics: weigh it, box it, ship the box with the carrier's invoice
    Invoke-RestMethod -Method Post -Uri "$base/parcels/$($parcels[0].id)/receive" -Headers $oph -Body '{"weight_g":1250,"length_mm":340,"width_mm":230,"height_mm":130}' | Out-Null
    $batch = (Invoke-RestMethod -Method Post -Uri "$base/batches" -Headers $h -Body '{"lane":"us_forwarder"}').id
    Invoke-RestMethod -Method Post -Uri "$base/batches/$batch/parcels" -Headers $h -Body "{`"parcel_id`":`"$($parcels[0].id)`"}" | Out-Null
    Invoke-RestMethod -Method Post -Uri "$base/batches/$batch/close" -Headers $h | Out-Null
    $allocs = Invoke-RestMethod -Method Post -Uri "$base/batches/$batch/ship" -Headers $h -Body '{"freight":"27.50","currency":"USD"}'
    Write-Host ("batch {0} shipped: order {1} billed on {2} g → freight {3} {4}" -f $batch, $allocs[0].order_id, $allocs[0].chargeable_g, $allocs[0].freight.amount, $allocs[0].freight.currency)
    Start-Sleep -Seconds 1   # batch_shipped → ordering (in transit) + pricing (actual freight)

    # ordering: balance, delivery; pricing: Quote vs Actual
    Invoke-RestMethod -Method Post -Uri "$base/orders/$order/balance" -Headers $h -Body "{`"amount`":`"$($view.home.deposit.amount)`",`"currency`":`"VND`"}" | Out-Null
    Invoke-RestMethod -Method Post -Uri "$base/orders/$order/deliver" -Headers $h | Out-Null
    $ov = Invoke-RestMethod -Method Get -Uri "$base/orders/$order" -Headers $h
    $rec = Invoke-RestMethod -Method Get -Uri "$base/reconciliations/$order" -Headers $h
    Write-Host ("order at the end: {0}; quote vs actual: quoted {1}+{2}, actual {3}+{4} → variance {5} {6}" -f $ov.status, $rec.quoted.goods.amount, $rec.quoted.freight.amount, $rec.actual.goods.amount, $rec.actual.freight.amount, $rec.variance.amount, $rec.variance.currency)
    # The READ side: one flat row per order, built from the events of five
    # contexts. The customer's screen and the operator's queue read the same
    # table, and the customer's does not carry the shop's own order number.
    Start-Sleep -Seconds 1
    $mine = Invoke-RestMethod -Method Get -Uri "$base/me/orders" -Headers $ch
    $queue = Invoke-RestMethod -Method Get -Uri "$base/orders?status=delivered" -Headers $h
    Write-Host ("my orders: {0} row(s), {1} - {2}, tracking {3}, shop ref on customer view: '{4}'; operator queue(delivered): {5} row(s), ref {6}" -f @($mine).Count, $mine[0].product_name, $mine[0].status, $mine[0].tracking, $mine[0].shop_reference, @($queue).Count, $queue[0].shop_reference)
    Start-Sleep -Seconds 1
} finally {
    # Stop the binaries BEFORE reading their log. On Windows a tail of a file
    # another process still holds open can block — which turned a 20-second
    # smoke into a hung script once already.
    Stop-Process -Id $api.Id, $worker.Id -Force -ErrorAction SilentlyContinue
    Start-Sleep -Milliseconds 500
    Write-Host "--- worker log (last 20) ---"
    Get-Content bin\worker.log -Tail 20 | ForEach-Object { if ($_.Length -gt 160) { $_.Substring(0, 160) + " …" } else { $_ } }
}
