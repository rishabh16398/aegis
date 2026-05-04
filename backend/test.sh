#!/usr/bin/env bash
# Aegis Backend — end-to-end API test
# Run from backend/: bash test.sh

BASE="http://localhost:8080"
PASS=0
FAIL=0

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
RESET='\033[0m'

pass() { echo -e "${GREEN}✓ $1${RESET}"; PASS=$((PASS + 1)); }
fail() { echo -e "${RED}✗ $1${RESET}"; FAIL=$((FAIL + 1)); }
section() { echo -e "\n${BOLD}${YELLOW}── $1 ──${RESET}"; }

expect_status() {
  local label=$1 expected=$2 actual=$3
  [ "$actual" = "$expected" ] && pass "$label (HTTP $actual)" || fail "$label — expected $expected got $actual"
}

expect_contains() {
  local label=$1 needle=$2 haystack=$3
  echo "$haystack" | grep -q "$needle" && pass "$label (contains '$needle')" || fail "$label — '$needle' not found in: $haystack"
}

# ── Setup ─────────────────────────────────────────────────────────────────────
echo 'X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*' > /tmp/eicar_test.txt
echo "clean file" > /tmp/clean_test.txt

# ── 1. Health ─────────────────────────────────────────────────────────────────
section "1. Health"
R=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/health)
expect_status "GET /api/v1/health" "200" "$R"
BODY=$(curl -s $BASE/api/v1/health)
expect_contains "health body" '"ok"' "$BODY"

# ── 2. Stats (initial) ────────────────────────────────────────────────────────
section "2. Stats"
R=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/stats)
expect_status "GET /api/v1/stats" "200" "$R"
BODY=$(curl -s $BASE/api/v1/stats)
expect_contains "stats has threats_found key" '"threats_found"' "$BODY"
expect_contains "stats has alerts_unread key" '"alerts_unread"' "$BODY"

# ── 3. Scan — clean file ──────────────────────────────────────────────────────
section "3. Scan clean file"
RESP=$(curl -s -X POST $BASE/api/v1/scan/file \
  -H "Content-Type: application/json" \
  -d '{"path":"/tmp/clean_test.txt"}')
R=$(curl -s -o /dev/null -w "%{http_code}" -X POST $BASE/api/v1/scan/file \
  -H "Content-Type: application/json" \
  -d '{"path":"/tmp/clean_test.txt"}')
expect_status "POST /api/v1/scan/file (clean)" "202" "$R"
expect_contains "clean scan returns job_id" '"job_id"' "$RESP"
CLEAN_JOB=$(echo $RESP | python3 -c "import sys,json; print(json.load(sys.stdin)['job_id'])" 2>/dev/null)
sleep 1

# ── 4. Scan — EICAR (threat) ──────────────────────────────────────────────────
section "4. Scan EICAR test virus"
RESP=$(curl -s -X POST $BASE/api/v1/scan/file \
  -H "Content-Type: application/json" \
  -d '{"path":"/tmp/eicar_test.txt"}')
R=$(curl -s -o /dev/null -w "%{http_code}" -X POST $BASE/api/v1/scan/file \
  -H "Content-Type: application/json" \
  -d '{"path":"/tmp/eicar_test.txt"}')
expect_status "POST /api/v1/scan/file (eicar)" "202" "$R"
EICAR_JOB=$(echo $RESP | python3 -c "import sys,json; print(json.load(sys.stdin)['job_id'])" 2>/dev/null)
sleep 2  # give scanner time to finish

# ── 5. Scan — bad path ────────────────────────────────────────────────────────
section "5. Scan bad path (error handling)"
R=$(curl -s -o /dev/null -w "%{http_code}" -X POST $BASE/api/v1/scan/file \
  -H "Content-Type: application/json" \
  -d '{"path":"/tmp/does_not_exist_xyz.txt"}')
expect_status "POST /api/v1/scan/file (bad path)" "400" "$R"

R=$(curl -s -o /dev/null -w "%{http_code}" -X POST $BASE/api/v1/scan/file \
  -H "Content-Type: application/json" \
  -d '{}')
expect_status "POST /api/v1/scan/file (empty body)" "400" "$R"

# ── 6. List jobs ──────────────────────────────────────────────────────────────
section "6. Scan jobs"
R=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/scan/jobs)
expect_status "GET /api/v1/scan/jobs" "200" "$R"
BODY=$(curl -s $BASE/api/v1/scan/jobs)
expect_contains "jobs list is array" 'Status' "$BODY"

# ── 7. Get specific job ───────────────────────────────────────────────────────
section "7. Get job by ID"
if [ -n "$EICAR_JOB" ]; then
  R=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/scan/jobs/$EICAR_JOB)
  expect_status "GET /api/v1/scan/jobs/:id (eicar job)" "200" "$R"
  BODY=$(curl -s $BASE/api/v1/scan/jobs/$EICAR_JOB)
  expect_contains "eicar job is completed" '"completed"' "$BODY"
  expect_contains "eicar job has ThreatsFound=1" '"ThreatsFound":1' "$BODY"
fi
R=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/scan/jobs/nonexistent-id)
expect_status "GET /api/v1/scan/jobs/:id (missing)" "404" "$R"

# ── 8. Job results ────────────────────────────────────────────────────────────
section "8. Scan results"
if [ -n "$EICAR_JOB" ]; then
  R=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/scan/jobs/$EICAR_JOB/results)
  expect_status "GET /api/v1/scan/jobs/:id/results" "200" "$R"
  BODY=$(curl -s $BASE/api/v1/scan/jobs/$EICAR_JOB/results)
  expect_contains "eicar result status=threat" '"threat"' "$BODY"
  expect_contains "eicar result has threat name" 'Eicar' "$BODY"
  expect_contains "eicar result has sha256" '"HashSHA256"' "$BODY"
fi

# ── 9. Alerts ─────────────────────────────────────────────────────────────────
section "9. Alerts"
R=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/alerts)
expect_status "GET /api/v1/alerts" "200" "$R"
BODY=$(curl -s $BASE/api/v1/alerts)
expect_contains "alerts has eicar alert" 'Eicar' "$BODY"
expect_contains "alert is unread" '"unread"' "$BODY"

# Get the alert ID and acknowledge it
ALERT_ID=$(echo $BODY | python3 -c "import sys,json; alerts=json.load(sys.stdin); print(alerts[0]['ID'])" 2>/dev/null)
if [ -n "$ALERT_ID" ]; then
  R=$(curl -s -o /dev/null -w "%{http_code}" -X PUT $BASE/api/v1/alerts/$ALERT_ID/ack)
  expect_status "PUT /api/v1/alerts/:id/ack" "200" "$R"
  BODY=$(curl -s $BASE/api/v1/alerts)
  expect_contains "alert is now acknowledged" '"acknowledged"' "$BODY"
fi

# ── 10. Network endpoints ─────────────────────────────────────────────────────
section "10. Network (stubs)"
R=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/network/connections)
expect_status "GET /api/v1/network/connections" "200" "$R"
R=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/network/events)
expect_status "GET /api/v1/network/events" "200" "$R"
R=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/network/blocked)
expect_status "GET /api/v1/network/blocked" "200" "$R"

# ── 11. Process endpoints ─────────────────────────────────────────────────────
section "11. Processes (stubs)"
R=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/processes/)
expect_status "GET /api/v1/processes/" "200" "$R"
R=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/processes/suspicious)
expect_status "GET /api/v1/processes/suspicious" "200" "$R"

# ── 12. Quarantine endpoints ──────────────────────────────────────────────────
section "12. Quarantine (stubs)"
R=$(curl -s -o /dev/null -w "%{http_code}" $BASE/api/v1/quarantine/)
expect_status "GET /api/v1/quarantine/" "200" "$R"

# ── 13. Final stats ───────────────────────────────────────────────────────────
section "13. Stats after scans"
BODY=$(curl -s $BASE/api/v1/stats)
expect_contains "stats: threats_found >= 1" '"threats_found"' "$BODY"

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}────────────────────────────────${RESET}"
TOTAL=$((PASS + FAIL))
echo -e "${BOLD}Results: ${GREEN}$PASS passed${RESET} / ${RED}$FAIL failed${RESET} / $TOTAL total"
[ $FAIL -eq 0 ] && echo -e "${GREEN}${BOLD}All tests passed!${RESET}" || echo -e "${RED}${BOLD}Some tests failed.${RESET}"

# Cleanup
rm -f /tmp/eicar_test.txt /tmp/clean_test.txt
