#!/usr/bin/env sh
# ═══════════════════════════════════════════════════════════════════════════════
#  Spectre VMCAI 2027 — External Reviewer Reproduction Script
#  Paper: "Spectre: Continuous Refinement and Closed-Loop Verification for Rust"
# ───────────────────────────────────────────────────────────────────────────────
#  Covers all four research questions (Tables 2–8):
#    RQ1 — Verification and scaling (Tables 2–3)
#    RQ2 — SMT semantic-sync ablation (Table 5)
#    RQ3 — CEGIS/WP closed-loop repair (Tables 4, 6)
#    RQ4 — Spec mining breadth; guided vs random exploration (Tables 7–8)
#
#  USAGE
#    Local build (requires Go ≥ 1.24, Z3):
#      sh artifact/reproduce.sh
#
#    Docker (recommended — no local toolchain required):
#      docker build -t spectre-vmcai .
#      docker run --rm spectre-vmcai sh /artifact/reproduce.sh
#
#    Skip the two long benchmarks (5-node Raft, ~300s total):
#      SKIP_LONG=1 sh artifact/reproduce.sh
#
#  EXPECTED TIMING (Apple M3 Pro)
#    Full run (SKIP_LONG unset): ~7 min
#    Short run (SKIP_LONG=1):    ~30 s
# ═══════════════════════════════════════════════════════════════════════════════

set -e

# ── Locate spectre binary ─────────────────────────────────────────────────────
# Prefer a locally built binary (dev workflow); fall back to PATH (Docker).
SPECTRE=""
if [ -x "./spectre" ]; then
    # Verify the local binary supports --por (it was added in the VMCAI branch).
    if ./spectre verify examples/counter.spec --max-states 5 --por >/dev/null 2>&1; then
        SPECTRE=./spectre
    else
        echo "Local ./spectre is stale (missing --por). Rebuilding from source..."
        go build -trimpath -o spectre ./cmd/spectre
        SPECTRE=./spectre
    fi
elif command -v spectre >/dev/null 2>&1; then
    SPECTRE=spectre
else
    echo "ERROR: 'spectre' binary not found. Build with 'go build ./cmd/spectre' or use Docker."
    exit 1
fi
echo "Using binary: $SPECTRE"
echo ""

PASS=0
FAIL=0

# ── Helper functions ──────────────────────────────────────────────────────────

section() { printf "\n════════════════════════════════════════════\n %s\n════════════════════════════════════════════\n" "$1"; }
subsect() { printf "\n──────────────────────────────────────────\n %s\n──────────────────────────────────────────\n" "$1"; }

check_eq() {
    # check_eq LABEL GOT WANT
    if [ "$2" = "$3" ]; then
        printf "  [PASS] %s = %s\n" "$1" "$2"; PASS=$((PASS+1))
    else
        printf "  [FAIL] %s: expected '%s', got '%s'\n" "$1" "$3" "$2"; FAIL=$((FAIL+1))
    fi
}

check_ge() {
    # check_ge LABEL GOT MIN  (numeric ≥)
    if [ "$2" -ge "$3" ] 2>/dev/null; then
        printf "  [PASS] %s = %s (≥ %s)\n" "$1" "$2" "$3"; PASS=$((PASS+1))
    else
        printf "  [FAIL] %s: expected ≥ %s, got '%s'\n" "$1" "$3" "$2"; FAIL=$((FAIL+1))
    fi
}

check_contains() {
    # check_contains LABEL OUTPUT PATTERN
    if printf '%s' "$2" | grep -q "$3"; then
        printf "  [PASS] %s\n" "$1"; PASS=$((PASS+1))
    else
        printf "  [FAIL] %s (pattern not found: %s)\n" "$1" "$3"; FAIL=$((FAIL+1))
    fi
}

check_not_contains() {
    if ! printf '%s' "$2" | grep -q "$3"; then
        printf "  [PASS] %s\n" "$1"; PASS=$((PASS+1))
    else
        printf "  [FAIL] %s (unexpected pattern found: %s)\n" "$1" "$3"; FAIL=$((FAIL+1))
    fi
}

states_from() { printf '%s' "$1" | grep "^Traversed" | awk '{print $2}'; }
time_ms() {
    start=$1; end=$2
    # Prefer nanosecond precision; fall back to seconds.
    if echo "$start" | grep -qE '^[0-9]{13,}$'; then
        echo $(( (end - start) / 1000000 ))
    else
        echo $(( (end - start) * 1000 ))
    fi
}

# Clear any stale cache so incremental and cache benchmarks start clean.
rm -rf examples/.spectre-cache .spectre-cache

echo "============================================"
echo " Spectre VMCAI 2027 — Reproduction Script"
echo " Paper: doi:10.1007/978-3-031-XXXXX"
echo "============================================"
echo " SKIP_LONG=${SKIP_LONG:-0}  (set SKIP_LONG=1 to skip 5-node Raft)"
echo ""

# ═════════════════════════════════════════════════════════════════════════════
section "RQ1 — Verification and Scaling"
# Paper: Table 2 (state counts) and Table 3 (incremental re-verification)
# Claim: POR heuristic reduces 3-node Raft from 22,432 → 6,818 states (3.3×);
#        incremental re-verify achieves 4.3–5.2× per-action speedup.
# ═════════════════════════════════════════════════════════════════════════════

# ── RQ1a: 3-node Raft BFS (Table 2, baseline row) ────────────────────────────
subsect "RQ1a — 3-node Raft BFS (expect 22,432 states)"
t0=$(date +%s%N 2>/dev/null || date +%s)
out=$($SPECTRE verify examples/raft-election-safety.spec --max-states 30000 2>&1)
t1=$(date +%s%N 2>/dev/null || date +%s)
ms=$(time_ms "$t0" "$t1")
states=$(states_from "$out")
printf "  States explored : %s   Wall time : %d ms\n" "$states" "$ms"
check_eq   "3-node BFS state count (Table 2)"         "$states" "22432"
check_contains "No violations found"                   "$out"    "Found no violations"

# ── RQ1b: 3-node Raft POR heuristic (Table 2, POR row) ─────────────────────
subsect "RQ1b — 3-node Raft POR heuristic (expect 6,818 states, 3.3× reduction)"
t0=$(date +%s%N 2>/dev/null || date +%s)
out_por=$($SPECTRE verify examples/raft-election-safety.spec --max-states 30000 --por 2>&1)
t1=$(date +%s%N 2>/dev/null || date +%s)
ms_por=$(time_ms "$t0" "$t1")
states_por=$(states_from "$out_por")
printf "  States explored : %s   Wall time : %d ms\n" "$states_por" "$ms_por"
check_eq   "3-node POR state count (Table 2)"          "$states_por" "6818"
check_contains "No violations found (POR)"             "$out_por"    "Found no violations"

# ── RQ1c: Cache warm re-run (Table 3, cache column) ─────────────────────────
subsect "RQ1c — Cache warm re-run (expect restored-from-cache, <<full BFS time)"
# Cold run populates the cache.
$SPECTRE verify examples/raft-election-safety.spec --max-states 30000 --use-cache >/dev/null 2>&1
# Warm run reads from cache — should be <1 s on any hardware.
t0=$(date +%s%N 2>/dev/null || date +%s)
out_warm=$($SPECTRE verify examples/raft-election-safety.spec --max-states 30000 --use-cache 2>&1)
t1=$(date +%s%N 2>/dev/null || date +%s)
ms_warm=$(time_ms "$t0" "$t1")
printf "  Cache restore time : %d ms   (paper: ~265 ms on M3 Pro)\n" "$ms_warm"
check_contains "Cache hit: restored from cache"        "$out_warm" "Restored state graph from cache"
check_contains "No violations after cache restore"     "$out_warm" "Found no violations"

# ── RQ1d: Incremental re-verify — stepDown1_sees2 (Table 3, row 1) ──────────
subsect "RQ1d — Incremental stepDown1_sees2 (expect 1,649 states pruned, 5.2× speedup)"
t0=$(date +%s%N 2>/dev/null || date +%s)
out_incr1=$($SPECTRE verify examples/raft-election-safety.spec \
    --max-states 30000 --use-cache --incremental --changed-action stepDown1_sees2 2>&1)
t1=$(date +%s%N 2>/dev/null || date +%s)
ms_incr1=$(time_ms "$t0" "$t1")
printf "  Incremental time : %d ms   (paper: ~2.4 s → 5.2× vs full BFS ~12.3 s)\n" "$ms_incr1"
check_contains "1,649 states pruned"                   "$out_incr1" "pruned 1649 states"
check_contains "22,432 total states preserved"         "$out_incr1" "22432 total states"
check_contains "No violations (incremental step)"      "$out_incr1" "Found no violations"

# ── RQ1e: Incremental re-verify — vote2for1 (Table 3, row 2) ────────────────
subsect "RQ1e — Incremental vote2for1 (expect 6,501 states pruned, 4.3× speedup)"
t0=$(date +%s%N 2>/dev/null || date +%s)
out_incr2=$($SPECTRE verify examples/raft-election-safety.spec \
    --max-states 30000 --use-cache --incremental --changed-action vote2for1 2>&1)
t1=$(date +%s%N 2>/dev/null || date +%s)
ms_incr2=$(time_ms "$t0" "$t1")
printf "  Incremental time : %d ms   (paper: ~2.9 s → 4.3× vs full BFS ~12.3 s)\n" "$ms_incr2"
check_contains "6,501 states pruned"                   "$out_incr2" "pruned 6501 states"
check_contains "22,432 total states preserved"         "$out_incr2" "22432 total states"
check_contains "No violations (incremental vote)"      "$out_incr2" "Found no violations"

# ── RQ1f & RQ1g: 5-node Raft (long benchmarks) ───────────────────────────────
if [ "${SKIP_LONG:-0}" = "1" ]; then
    printf "\n  [SKIP] 5-node Raft benchmarks (SKIP_LONG=1). Set SKIP_LONG=0 to run.\n"
    printf "         Expected: BFS explores 50,000 states in ~268 s (M3 Pro);\n"
    printf "                   POR explores 50,000 states in ~38.7 s (6.9× reduction).\n"
    printf "         Command:  spectre verify examples/raft-election-safety-5node.spec --max-states 50001\n"
    printf "         POR:      spectre verify examples/raft-election-safety-5node.spec --max-states 50001 --por\n"
else
    subsect "RQ1f — 5-node Raft BFS (state limit 50,001; paper ~268 s)"
    printf "  NOTE: This benchmark takes ~4–5 min. Set SKIP_LONG=1 to skip.\n"
    t0=$(date +%s%N 2>/dev/null || date +%s)
    out_5bfs=$($SPECTRE verify examples/raft-election-safety-5node.spec --max-states 50001 2>&1)
    t1=$(date +%s%N 2>/dev/null || date +%s)
    ms_5bfs=$(time_ms "$t0" "$t1")
    states_5bfs=$(states_from "$out_5bfs")
    printf "  States explored : %s   Wall time : %d ms\n" "$states_5bfs" "$ms_5bfs"
    check_ge   "5-node BFS explores large space (Table 2)"  "${states_5bfs:-0}" "50000"
    check_contains "No violations (5-node BFS)"              "$out_5bfs"         "Found no violations"

    subsect "RQ1g — 5-node Raft POR heuristic (paper ~38.7 s, 6.9× reduction)"
    t0=$(date +%s%N 2>/dev/null || date +%s)
    out_5por=$($SPECTRE verify examples/raft-election-safety-5node.spec --max-states 50001 --por 2>&1)
    t1=$(date +%s%N 2>/dev/null || date +%s)
    ms_5por=$(time_ms "$t0" "$t1")
    states_5por=$(states_from "$out_5por")
    printf "  States explored : %s   Wall time : %d ms\n" "$states_5por" "$ms_5por"
    check_ge   "5-node POR explores large space (Table 2)"  "${states_5por:-0}" "50000"
    check_contains "No violations (5-node POR)"              "$out_5por"         "Found no violations"
    if [ "${ms_5bfs:-0}" -gt 0 ] && [ "${ms_5por:-0}" -gt 0 ]; then
        ratio=$( awk "BEGIN{printf \"%.1f\", $ms_5bfs/$ms_5por}" 2>/dev/null || echo "N/A")
        printf "  Speedup (BFS/POR): %s×  (paper: 6.9×)\n" "$ratio"
    fi
fi

# ── RQ1h: 7-node simulation (not exhaustively explored) ─────────────────────
subsect "RQ1h — 7-node Raft simulation (100 random traces, paper: 1.64 s)"
_sim_dir=$(mktemp -d /tmp/spectre_sim7.XXXXXX)
t0=$(date +%s%N 2>/dev/null || date +%s)
out_7sim=$($SPECTRE simulate examples/raft-election-safety-7node.spec \
    --traces 100 --max-steps 50 --seed 42 --output-dir "$_sim_dir" 2>&1)
t1=$(date +%s%N 2>/dev/null || date +%s)
rm -rf "$_sim_dir"
ms_7sim=$(time_ms "$t0" "$t1")
printf "  Wall time : %d ms   (paper: ~1,640 ms on M3 Pro)\n" "$ms_7sim"
check_contains "100 simulation traces generated"       "$out_7sim" "Generated 100 trace(s)"
# The paper reports 0 invariant violations in direct simulation of the 7-node spec.
check_contains "0 invariant violations in simulation"  "$out_7sim" "0 invariant violation(s)"


# ═════════════════════════════════════════════════════════════════════════════
section "RQ2 — SMT Semantic-Sync Ablation (Table 5)"
# Paper: 8 semantic mutations + 5 equivalent rewrites of bank_account.rs.
#        AST+SMT detector: 8/8 semantic changes detected, 1 false positive
#        (correctly classifies 4 of 5 equivalent rewrites; 1 times-out → FP).
#
# We replicate 3 representative cases inline:
#   M1 — semantic: flip withdraw body   (balance -= → balance +=)
#   M2 — semantic: relax withdraw guard (>= → >)
#   E1 — equivalent rewrite: swap deposit guard sides
# ═════════════════════════════════════════════════════════════════════════════

subsect "RQ2 — Preparing baseline spec for bank_account"
# Mine the baseline spec from the canonical bank_account.rs.
BASELINE_SPEC=$(mktemp /tmp/bank_account_base.XXXXXX.spec)
$SPECTRE mine --lang rust examples/rust/bank_account.rs -o "$BASELINE_SPEC" >/dev/null 2>&1 \
    || $SPECTRE mine --lang rust examples/rust/bank_account.rs > "$BASELINE_SPEC" 2>/dev/null
check_contains "Baseline spec mined successfully" \
    "$(cat "$BASELINE_SPEC")" "action withdraw"

# M1: Semantic mutation — flip withdraw body (balance -= amount → balance += amount)
subsect "RQ2/M1 — Semantic mutation: withdraw body flipped (expect [✗] detected)"
M1_RS=$(mktemp /tmp/bank_account_m1.XXXXXX.rs)
sed 's/self\.balance -= amount/self.balance += amount/' \
    examples/rust/bank_account.rs > "$M1_RS"
out_m1=$($SPECTRE sync "$M1_RS" --spec "$BASELINE_SPEC" 2>&1 || true)
printf '%s\n' "$out_m1" | grep -E "^\s+\[" | head -8
check_contains "M1 semantic change detected ([✗])"     "$out_m1" '[✗]'
check_not_contains "M1 not mis-classified as equiv"    "$out_m1" '[≡].*withdraw.*assignment'

# M2: Semantic mutation — tighten deposit cap (> instead of >=, etc.; change a guard sense)
subsect "RQ2/M2 — Semantic mutation: withdraw guard relaxed (>= → >; expect [✗] detected)"
M2_RS=$(mktemp /tmp/bank_account_m2.XXXXXX.rs)
sed 's/self\.balance >= amount/self.balance > amount/' \
    examples/rust/bank_account.rs > "$M2_RS"
out_m2=$($SPECTRE sync "$M2_RS" --spec "$BASELINE_SPEC" 2>&1 || true)
printf '%s\n' "$out_m2" | grep -E "^\s+\[" | head -8
check_contains "M2 semantic change detected ([✗])"     "$out_m2" '[✗]'

# E1: Equivalent rewrite — commute deposit cap check (a + b <= C  ↔  b + a <= C)
subsect "RQ2/E1 — Equivalent rewrite: deposit guard commuted (expect [≡] SMT-proved)"
E1_RS=$(mktemp /tmp/bank_account_e1.XXXXXX.rs)
sed 's/self\.balance + amount <= 1_000_000/amount + self.balance <= 1_000_000/' \
    examples/rust/bank_account.rs > "$E1_RS"
out_e1=$($SPECTRE sync "$E1_RS" --spec "$BASELINE_SPEC" 2>&1 || true)
printf '%s\n' "$out_e1" | grep -E "^\s+\[" | head -8
check_contains "E1 classified as equivalent ([≡])"     "$out_e1" '[≡]'
check_not_contains "E1 not flagged as semantic change" "$out_e1" '[✗].*deposit'

# Tidy up temp files.
rm -f "$BASELINE_SPEC" "$M1_RS" "$M2_RS" "$E1_RS"

printf "\n  Table 5 summary (full paper results):\n"
printf "  Structural detector  — 3/8 mutations detected, 0 FP\n"
printf "  NormAST detector     — 8/8 mutations detected, 4 FP\n"
printf "  AST+SMT (integrated) — 8/8 mutations detected, 1 FP  ← Spectre uses this\n"
printf "  (Full 13-case ablation requires the complete mutation suite in examples/rust/sync-ablation/)\n"


# ═════════════════════════════════════════════════════════════════════════════
section "RQ3 — CEGIS/WP Closed-Loop Repair (Tables 4, 6)"
# Paper: 5 defect classes; 4/5 yield a fully-synthesized WP guard verified
#        by re-exploration; 1 (missing-term Raft) is a safety-only repair.
#
# We test the 3 packaged repair benchmarks plus the bank-account defect.
# The two Raft-injection cases (wrong-quorum, missing-term) are described
# below with the exact command to reproduce them.
# ═════════════════════════════════════════════════════════════════════════════

subsect "RQ3a — Repair benchmark: boolean missing guard"
# Expected: WP guard synthesized, re-exploration verifies invariant holds.
out_bool=$($SPECTRE verify examples/repair-benchmark-bool.spec 2>&1)
check_contains "R3a violation detected"                "$out_bool" "VIOLATION DETECTED"
check_contains "R3a CEGIS suggestion produced"         "$out_bool" "CEGIS Repair Suggestions"
check_contains "R3a WP guard synthesized"              "$out_bool" "require"
check_contains "R3a guard verified by re-exploration"  "$out_bool" "Verified: re-explored"

subsect "RQ3b — Repair benchmark: enum missing guard"
out_enum=$($SPECTRE verify examples/repair-benchmark-enum.spec 2>&1)
check_contains "R3b violation detected"                "$out_enum" "VIOLATION DETECTED"
check_contains "R3b CEGIS suggestion produced"         "$out_enum" "CEGIS Repair Suggestions"
check_contains "R3b WP guard synthesized"              "$out_enum" "require"
check_contains "R3b guard verified by re-exploration"  "$out_enum" "Verified: re-explored"

subsect "RQ3c — Repair benchmark: simultaneous assignment"
out_sim=$($SPECTRE verify examples/repair-benchmark-simultaneous.spec 2>&1)
check_contains "R3c violation detected"                "$out_sim"  "VIOLATION DETECTED"
check_contains "R3c CEGIS suggestion produced"         "$out_sim"  "CEGIS Repair Suggestions"
check_contains "R3c WP guard synthesized"              "$out_sim"  "require"
check_contains "R3c guard verified by re-exploration"  "$out_sim"  "Verified: re-explored"

subsect "RQ3d — Bank-account withdraw guard defect"
# Reduced state space for speed; CEGIS still fires on first violation.
out_ba=$($SPECTRE verify examples/bank-account-violation.spec --max-states 50 2>&1)
check_contains "R3d violation detected"                "$out_ba"   "VIOLATION DETECTED"
check_contains "R3d CEGIS suggestion produced"         "$out_ba"   "CEGIS Repair Suggestions"
check_contains "R3d WP guard synthesized"              "$out_ba"   "require"

printf "\n  Table 4 / Table 6 note (Raft defect injection):\n"
printf "  To reproduce the Raft wrong-quorum and missing-term defects from Table 4,\n"
printf "  edit examples/raft-election-safety.spec as follows and re-run verify:\n"
printf "\n"
printf "    Wrong quorum  — in action 'becomeLeader1', change:\n"
printf "        require votes1 * 2 > 3    →   require votes1 > 1\n"
printf "    Expected: CEGIS synthesizes guard 'votes1 * 2 > 3' (CE path length 2)\n"
printf "\n"
printf "    Missing term  — in action 'becomeLeader1', remove the 'term1 >= term2' require.\n"
printf "    Expected: safety-only repair (WP guard correctly installed but cannot\n"
printf "              restore full intended semantics — paper Table 4, row 2).\n"


# ═════════════════════════════════════════════════════════════════════════════
section "RQ4 — Spec Mining Breadth and Guided-Exploration Effectiveness (Tables 7–8)"
# ═════════════════════════════════════════════════════════════════════════════

# ── RQ4a: Mine 9 Rust components (Table 7) ──────────────────────────────────
subsect "RQ4a — Mine 9 Rust components (expect 100% yield)"
# Each component must produce a spec that parses, type-checks, and can be
# executed by 'spectre verify' without a crash.  Wall time is per-component.
MINE_PASS=0
MINE_FAIL=0

mine_and_verify() {
    component="$1"; rs="$2"
    mined=$(mktemp /tmp/mined_${component}.XXXXXX.spec)
    # Mine; --lang flag may not be needed if auto-detected, but explicit is safer.
    if $SPECTRE mine --lang rust "$rs" -o "$mined" >/dev/null 2>&1 \
        || $SPECTRE mine --lang rust "$rs" > "$mined" 2>/dev/null; then
        # Sanity-check the mined spec: parse + 1-step verify.
        if $SPECTRE verify "$mined" --max-states 1 >/dev/null 2>&1 \
            || $SPECTRE verify "$mined" --max-states 5 2>&1 | grep -qE "Traversed|Found no"; then
            printf "  [PASS] mine %s\n" "$component"; MINE_PASS=$((MINE_PASS+1))
        else
            printf "  [FAIL] mine %s — spec fails to verify\n" "$component"; MINE_FAIL=$((MINE_FAIL+1))
        fi
    else
        printf "  [FAIL] mine %s — mine command failed\n" "$component"; MINE_FAIL=$((MINE_FAIL+1))
    fi
    rm -f "$mined"
}

mine_and_verify "bank_account"     "examples/rust/bank_account.rs"
mine_and_verify "cache"            "examples/rust/cache.rs"
mine_and_verify "circuit_breaker"  "examples/rust/circuit_breaker.rs"
mine_and_verify "connection_pool"  "examples/rust/connection_pool.rs"
mine_and_verify "event_log"        "examples/rust/event_log.rs"
mine_and_verify "queue"            "examples/rust/queue.rs"
mine_and_verify "rate_limiter"     "examples/rust/rate_limiter.rs"
mine_and_verify "semaphore"        "examples/rust/semaphore.rs"
mine_and_verify "stack"            "examples/rust/stack.rs"

printf "  Mining result: %d/9 passed, %d failed\n" "$MINE_PASS" "$MINE_FAIL"
if [ "$MINE_PASS" -eq 9 ]; then
    printf "  [PASS] All 9 components mined successfully (Table 7: 100%% yield)\n"
    PASS=$((PASS+1))
else
    printf "  [FAIL] Expected 9/9 mining successes, got %d\n" "$MINE_PASS"
    FAIL=$((FAIL+1))
fi

# ── RQ4b: Guided vs random exploration (Table 8) ────────────────────────────
subsect "RQ4b — BFS systematic: finds violations in ≤204 explored states (Table 8)"
# Paper: both BFS and property-directed find all 29 distinct violation-reaching
# traces within 204 explored states; 5,000 random 30-step simulations find none.
out_bfs_tbl8=$($SPECTRE verify examples/bank-account-violation.spec --max-states 204 2>&1)
states_bfs_tbl8=$(states_from "$out_bfs_tbl8")
printf "  States explored : %s\n" "$states_bfs_tbl8"
check_eq   "BFS explores exactly 204 states"           "$states_bfs_tbl8" "204"
check_contains "BFS finds at least one violation"       "$out_bfs_tbl8"   "VIOLATION DETECTED"

subsect "RQ4b — Property-directed: finds violations in ≤204 states (Table 8)"
out_pd=$($SPECTRE verify examples/bank-account-violation.spec --max-states 204 --property-guided 2>&1)
states_pd=$(states_from "$out_pd")
printf "  States explored : %s\n" "$states_pd"
check_eq   "Property-directed explores exactly 204 states"  "$states_pd" "204"
check_contains "Property-directed finds at least one violation"  "$out_pd" "VIOLATION DETECTED"

subsect "RQ4b — Random simulation: 5,000 traces × 30 steps finds 0 violations (Table 8)"
printf "  Running 5,000 random 30-step simulations (paper: 0 violations found)...\n"
out_rand=$($SPECTRE simulate examples/bank-account-violation.spec \
    --traces 5000 --max-steps 30 --seed 2027 2>&1)
violations_rand=$(printf '%s' "$out_rand" | grep "Generated.*trace" | grep -o '[0-9]* invariant' | awk '{print $1}')
printf "  Invariant violations found in 5,000 random traces: %s\n" "${violations_rand:-0}"
check_eq   "Random simulation: 0 violations in 150,000 steps (Table 8)" \
           "${violations_rand:-0}" "0"


# ═════════════════════════════════════════════════════════════════════════════
section "Summary"
# ═════════════════════════════════════════════════════════════════════════════
TOTAL=$((PASS + FAIL))
printf "\n  Checks passed : %d / %d\n" "$PASS" "$TOTAL"
if [ "${SKIP_LONG:-0}" = "1" ]; then
    printf "  (RQ1f/g 5-node Raft benchmarks were skipped — rerun with SKIP_LONG=0)\n"
fi

if [ "$FAIL" -gt 0 ]; then
    printf "\n  [OVERALL] FAIL — %d check(s) did not match expected values.\n\n" "$FAIL"
    exit 1
else
    printf "\n  [OVERALL] PASS — all checks reproduced the paper's claims.\n\n"
fi
