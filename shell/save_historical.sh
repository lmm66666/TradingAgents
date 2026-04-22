#!/bin/bash
#
# 每 10s 调用 SaveStockHistoricalData 保存一只股票的历史数据
# 用法: ./save_historical.sh [base_url]
# 示例: ./save_historical.sh http://localhost:8080
#

set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"
API_URL="${BASE_URL}/api/stocks/historical"
INTERVAL=10

codes=(
  603833 603858 603899 603986 603993 605169
)

total=${#codes[@]}

echo "Found ${total} stocks, interval=${INTERVAL}s, target=${API_URL}"
echo "---"

success=0
fail=0

for i in "${!codes[@]}"; do
  code="${codes[$i]}"
  idx=$((i + 1))

  resp=$(curl -s -w "\n%{http_code}" -X POST "$API_URL" \
    -H "Content-Type: application/json" \
    -d "{\"code\": \"${code}\"}" \
    --max-time 30)

  http_code=$(echo "$resp" | tail -1)
  body=$(echo "$resp" | sed '$d')

  if [ "$http_code" -eq 200 ]; then
    echo "[${idx}/${total}] ${code} OK"
    success=$((success + 1))
  else
    echo "[${idx}/${total}] ${code} FAIL (HTTP ${http_code}) ${body}"
    fail=$((fail + 1))
  fi

  if [ "$idx" -lt "$total" ]; then
    sleep "$INTERVAL"
  fi
done

echo "---"
echo "Done: ${success} success, ${fail} failed, ${total} total"
