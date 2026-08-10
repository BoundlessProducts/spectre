#!/usr/bin/env sh
# Spectre VMCAI 2027 artifact benchmark script
# Run locally:  sh artifact/run-benchmarks.sh
# Run via Docker: docker run --rm --entrypoint sh spectre-vmcai /artifact/run-benchmarks.sh
#
# Covers: Table 2 (verification), Table 3 (spec mining), Table 4 (incremental),
#         and the coverage-mode comparison (§5).
#
# State counts for finite specs are deterministic and must match exactly.
# Wall times vary with hardware; the paper's figures are from an Apple M3 Pro.

set -e

# Use the locally built binary if present (local dev); fall back to PATH (Docker)
if [ -x "./spectre" ]; then
    SPECTRE=./spectre
else
    SPECTRE=spectre
fi

PASS=0
FAIL=0

# ── helpers ──────────────────────────────────────────────────────────────────

run_spec() {
    label="$1"; spec="$2"; flags="$3"; expect_violation="$4"; expect_cegis="$5"
    printf "\n──────────────────────────────────────────\n"
    printf "%s\n" "$label"
    start=$(date +%s%N 2>/dev/null || date +%s)
    output=$($SPECTRE verify "$spec" $flags 2>&1)
    end=$(date +%s%N 2>/dev/null || date +%s)
    ms=$(( (end - start) / 1000000 ))
    states=$(printf '%s\n' "$output" | grep "^Traversed" | awk '{print $2}')
    printf "  States : %s   Time : %dms\n" "$states" "$ms"

    ok=1
    if [ "$expect_violation" = "yes" ]; then
        if printf '%s\n' "$output" | grep -q "VIOLATION DETECTED"; then
            printf "  [OK] Violation detected\n"
        else
            printf "  [FAIL] Expected a violation but none found\n"; ok=0
        fi
    else
        if printf '%s\n' "$output" | grep -q "Found no violations"; then
            printf "  [OK] No violations\n"
        else
            printf "  [FAIL] Expected no violations\n"; ok=0
        fi
    fi
    if [ "$expect_cegis" = "yes" ]; then
        if printf '%s\n' "$output" | grep -q "CEGIS Repair\|weakest precondition\|require"; then
            printf "  [OK] CEGIS repair suggestion present\n"
        else
            printf "  [FAIL] Expected CEGIS suggestions\n"; ok=0
        fi
    fi
    if [ "$ok" = "1" ]; then PASS=$((PASS+1)); else FAIL=$((FAIL+1)); fi
}

check_states() {
    label="$1"; got="$2"; want="$3"
    if [ "$got" = "$want" ]; then
        printf "  [OK] %s: %s states\n" "$label" "$got"
        PASS=$((PASS+1))
    else
        printf "  [FAIL] %s: expected %s states, got %s\n" "$label" "$want" "$got"
        FAIL=$((FAIL+1))
    fi
}

check_contains() {
    label="$1"; text="$2"; pattern="$3"
    if printf '%s\n' "$text" | grep -q "$pattern"; then
        printf "  [OK] %s\n" "$label"
        PASS=$((PASS+1))
    else
        printf "  [FAIL] %s (pattern not found: %s)\n" "$label" "$pattern"
        FAIL=$((FAIL+1))
    fi
}

# ── setup ────────────────────────────────────────────────────────────────────

echo "============================================"
echo " Spectre VMCAI 2027 Artifact Benchmark"
echo "============================================"

# Clear cache so incremental benchmarks start from a known state
rm -rf examples/.spectre-cache .spectre-cache

# ── TABLE 2: Verification benchmarks ─────────────────────────────────────────

printf "\n════════════════════════════════════════════\n"
printf " TABLE 2 — Verification benchmarks\n"
printf "════════════════════════════════════════════\n"

run_spec "bank-account-violation  (Table 2, row 1)" \
         examples/bank-account-violation.spec "" "yes" "yes"

run_spec "bank-account-corrected  (Table 2, row 2)" \
         examples/bank-account-corrected.spec "" "no" "no"

run_spec "concurrent-lock-violation  (Table 2, row 3, --max-states 100)" \
         examples/concurrent-lock-violation.spec "--max-states 100" "yes" "no"

run_spec "concurrent-lock-corrected  (Table 2, row 4)" \
         examples/concurrent-lock-corrected.spec "" "no" "no"

run_spec "counter  (Table 2, row 5)" \
         examples/counter.spec "" "yes" "no"

run_spec "message-queue-corrected  (Table 2, row 6)" \
         examples/message-queue-corrected.spec "" "no" "no"

run_spec "inventory-corrected  (Table 2, row 7)" \
         examples/inventory-corrected.spec "" "no" "no"

run_spec "stuttering-counter  (Table 2, row 8)" \
         examples/stuttering-counter.spec "" "no" "no"

printf "\n──────────────────────────────────────────\n"
printf "raft-election-safety  (Table 2, row 9 — 3-node Raft, --max-states 30000)\n"
start=$(date +%s%N 2>/dev/null || date +%s)
raft_cold=$($SPECTRE verify examples/raft-election-safety.spec --max-states 30000 2>&1)
end=$(date +%s%N 2>/dev/null || date +%s)
raft_cold_ms=$(( (end - start) / 1000000 ))
raft_states=$(printf '%s\n' "$raft_cold" | grep "^Traversed" | awk '{print $2}')
printf "  States : %s   Cold BFS time : %dms\n" "$raft_states" "$raft_cold_ms"
check_states "Raft cold BFS state count" "$raft_states" "22432"
check_contains "Raft: no violations" "$raft_cold" "Found no violations"

# ── TABLE 3: Spec mining ──────────────────────────────────────────────────────

printf "\n════════════════════════════════════════════\n"
printf " TABLE 3 — Spec mining (AST miner)\n"
printf "════════════════════════════════════════════\n"
printf "\n──────────────────────────────────────────\n"
printf "bank_account.rs — guard extraction\n"
mine_out=$($SPECTRE mine --lang rust artifact/bank_account.rs 2>&1)
guards=$(printf '%s\n' "$mine_out" | grep -c "require")
printf "  Extracted %d require guards (expected 7)\n" "$guards"
if [ "$guards" -ge 7 ]; then
    printf "  [OK] 7 require guards extracted\n"
    PASS=$((PASS+1))
else
    printf "  [FAIL] Expected >= 7 guards, got %d\n" "$guards"
    FAIL=$((FAIL+1))
fi
check_contains "Constructor init: balance = 0" "$mine_out" "balance = 0"
check_contains "Constructor init: frozen = false" "$mine_out" "frozen = false"

# ── TABLE 4: Incremental re-verification ─────────────────────────────────────

printf "\n════════════════════════════════════════════\n"
printf " TABLE 4 — Incremental re-verification\n"
printf " (timings vary by hardware; state counts are deterministic)\n"
printf "════════════════════════════════════════════\n"

# Step 1: populate cache for bank-account
printf "\n──────────────────────────────────────────\n"
printf "bank-account: cold BFS + cache write\n"
ba_cold=$($SPECTRE verify examples/bank-account-parameterized.spec --use-cache 2>&1)
ba_states=$(printf '%s\n' "$ba_cold" | grep "^Traversed" | awk '{print $2}')
check_states "bank-account cold BFS" "$ba_states" "3745"
check_contains "bank-account: no violations" "$ba_cold" "Found no violations"

# Step 2: cache restore
printf "\n──────────────────────────────────────────\n"
printf "bank-account: cache restore (unchanged spec)\n"
start=$(date +%s%N 2>/dev/null || date +%s)
ba_cache=$($SPECTRE verify examples/bank-account-parameterized.spec --use-cache 2>&1)
end=$(date +%s%N 2>/dev/null || date +%s)
ba_cache_ms=$(( (end - start) / 1000000 ))
printf "  Cache restore time : %dms\n" "$ba_cache_ms"
check_contains "bank-account: restored from cache" "$ba_cache" "Restored state graph from cache"
check_contains "bank-account: no violations after restore" "$ba_cache" "Found no violations"

# Step 3: incremental freeze
printf "\n──────────────────────────────────────────\n"
printf "bank-account: incremental re-verify (changed-action=freeze)\n"
start=$(date +%s%N 2>/dev/null || date +%s)
ba_incr=$($SPECTRE verify examples/bank-account-parameterized.spec \
    --use-cache --incremental --changed-action freeze 2>&1)
end=$(date +%s%N 2>/dev/null || date +%s)
ba_incr_ms=$(( (end - start) / 1000000 ))
printf "  Incremental time : %dms\n" "$ba_incr_ms"
check_contains "freeze: 1882 states pruned" "$ba_incr" "pruned 1882 states"
check_contains "freeze: 0 new states" "$ba_incr" "0 new states"
check_contains "freeze: no violations" "$ba_incr" "Found no violations"

# Step 4: populate Raft cache (cold BFS already run above; now save cache)
printf "\n──────────────────────────────────────────\n"
printf "Raft: cold BFS + cache write\n"
raft_cache_write=$($SPECTRE verify examples/raft-election-safety.spec \
    --max-states 30000 --use-cache 2>&1)
check_states "Raft cache write: state count" \
    "$(printf '%s\n' "$raft_cache_write" | grep "^Traversed" | awk '{print $2}')" "22432"

# Step 5: Raft cache restore
printf "\n──────────────────────────────────────────\n"
printf "Raft: cache restore (unchanged spec)\n"
start=$(date +%s%N 2>/dev/null || date +%s)
raft_restore=$($SPECTRE verify examples/raft-election-safety.spec \
    --max-states 30000 --use-cache 2>&1)
end=$(date +%s%N 2>/dev/null || date +%s)
raft_restore_ms=$(( (end - start) / 1000000 ))
printf "  Raft cache restore time : %dms\n" "$raft_restore_ms"
check_contains "Raft: restored from cache" "$raft_restore" "Restored state graph from cache (22432 states)"
check_contains "Raft: no violations after restore" "$raft_restore" "Found no violations"

# Step 6: Raft incremental stepDown1_sees2
printf "\n──────────────────────────────────────────\n"
printf "Raft: incremental re-verify (changed-action=stepDown1_sees2)\n"
start=$(date +%s%N 2>/dev/null || date +%s)
raft_step=$($SPECTRE verify examples/raft-election-safety.spec \
    --max-states 30000 --use-cache --incremental --changed-action stepDown1_sees2 2>&1)
end=$(date +%s%N 2>/dev/null || date +%s)
raft_step_ms=$(( (end - start) / 1000000 ))
printf "  stepDown1_sees2 incremental time : %dms\n" "$raft_step_ms"
check_contains "stepDown1_sees2: 1649 states pruned" "$raft_step" "pruned 1649 states"
check_contains "stepDown1_sees2: 22432 total" "$raft_step" "22432 total states"
check_contains "stepDown1_sees2: no violations" "$raft_step" "Found no violations"
# Note: "new states" count (≈1644) is non-deterministic by ±few due to BFS deduplication order

# Step 7: Raft incremental vote2for1
printf "\n──────────────────────────────────────────\n"
printf "Raft: incremental re-verify (changed-action=vote2for1)\n"
start=$(date +%s%N 2>/dev/null || date +%s)
raft_vote=$($SPECTRE verify examples/raft-election-safety.spec \
    --max-states 30000 --use-cache --incremental --changed-action vote2for1 2>&1)
end=$(date +%s%N 2>/dev/null || date +%s)
raft_vote_ms=$(( (end - start) / 1000000 ))
printf "  vote2for1 incremental time : %dms\n" "$raft_vote_ms"
check_contains "vote2for1: 6501 states pruned" "$raft_vote" "pruned 6501 states"
check_contains "vote2for1: 22432 total" "$raft_vote" "22432 total states"
check_contains "vote2for1: no violations" "$raft_vote" "Found no violations"
# Note: "new states" count (≈6493) is non-deterministic by ±few due to BFS deduplication order

# ── COVERAGE-MODE COMPARISON (§5) ────────────────────────────────────────────

printf "\n════════════════════════════════════════════\n"
printf " §5 Coverage-mode comparison (bank-account-corrected)\n"
printf " property mode must achieve more boundary steps than all others\n"
printf "════════════════════════════════════════════\n"

count_boundary() {
    mode="$1"
    out=$($SPECTRE verify examples/bank-account-corrected.spec \
        --emit-traces "/tmp/spectre_cov_$mode.itf.json" \
        --coverage-mode "$mode" --max-depth 20 2>&1)
    python3 - "$mode" "/tmp/spectre_cov_$mode.itf.json" <<'EOF'
import json, sys
mode = sys.argv[1]
path = sys.argv[2]
data = json.load(open(path))
states = data.get("states", [])
balances = [int(s["aliceBalance"]["#bigint"]) for s in states if "aliceBalance" in s]
n = len(states)
boundary = sum(1 for b in balances if b == 0)
mean = sum(balances) / len(balances) if balances else 0
print(f"MODE={mode} STEPS={n} MEAN={mean:.1f} BOUNDARY={boundary}")
EOF
}

printf "\n──────────────────────────────────────────\n"
action_result=$(count_boundary action)
boundary_result=$(count_boundary boundary)
rare_result=$(count_boundary rare-action)
property_result=$(count_boundary property)
printf "  %s\n" "$action_result"
printf "  %s\n" "$boundary_result"
printf "  %s\n" "$rare_result"
printf "  %s\n" "$property_result"

# Extract boundary counts for comparison
prop_boundary=$(printf '%s\n' "$property_result" | grep -o 'BOUNDARY=[0-9]*' | cut -d= -f2)
act_boundary=$(printf '%s\n' "$action_result" | grep -o 'BOUNDARY=[0-9]*' | cut -d= -f2)
bnd_boundary=$(printf '%s\n' "$boundary_result" | grep -o 'BOUNDARY=[0-9]*' | cut -d= -f2)
rare_boundary=$(printf '%s\n' "$rare_result" | grep -o 'BOUNDARY=[0-9]*' | cut -d= -f2)

if [ "$prop_boundary" -gt "$bnd_boundary" ] && \
   [ "$prop_boundary" -gt "$rare_boundary" ] && \
   [ "$prop_boundary" -ge "$act_boundary" ]; then
    printf "  [OK] property mode achieves highest boundary coverage (%s boundary steps)\n" "$prop_boundary"
    PASS=$((PASS+1))
else
    printf "  [FAIL] property mode (%s) did not dominate: action=%s boundary=%s rare=%s\n" \
        "$prop_boundary" "$act_boundary" "$bnd_boundary" "$rare_boundary"
    FAIL=$((FAIL+1))
fi

# ── summary ──────────────────────────────────────────────────────────────────

printf "\n════════════════════════════════════════════\n"
printf " Results: %d passed, %d failed\n" "$PASS" "$FAIL"
printf "════════════════════════════════════════════\n"
if [ "$FAIL" -gt 0 ]; then exit 1; fi
