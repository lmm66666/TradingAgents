#!/bin/bash
#
# 每 10s 调用 SaveFinancialReportData 保存一只股票的财报数据
# 用法: ./save_financial_report.sh [base_url]
# 示例: ./save_financial_report.sh http://localhost:8080
#

set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"
API_URL="${BASE_URL}/api/stocks/financial-report"
INTERVAL=7

codes=(
  600312
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

  http_code=$(echo "$resp" | tail -1 | tr -d '\r')
  body=$(echo "$resp" | sed '$d')

  if ! [[ "$http_code" =~ ^[0-9]+$ ]]; then
    http_code="000"
  fi

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
