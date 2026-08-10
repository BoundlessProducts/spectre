#!/usr/bin/env sh
# Quick smoke test (~10 seconds) — no slow Raft BFS.
# Checks three deterministic results before running the full benchmark suite.
#
# Usage: sh artifact/smoke-test.sh

set -e

if [ -x "./spectre" ]; then SPECTRE=./spectre; else SPECTRE=spectre; fi

PASS=0; FAIL=0

check() {
    label="$1"; pattern="$2"; output="$3"
    if printf '%s\n' "$output" | grep -q "$pattern"; then
        printf "[OK]   %s\n" "$label"; PASS=$((PASS+1))
    else
        printf "[FAIL] %s\n" "$label"; FAIL=$((FAIL+1))
    fi
}

printf "=== Spectre smoke test ===\n\n"

# 1. Deterministic finite state count
out=$($SPECTRE verify examples/concurrent-lock-corrected.spec 2>&1)
states=$(printf '%s\n' "$out" | grep "^Traversed" | awk '{print $2}')
if [ "$states" = "1928" ]; then
    printf "[OK]   concurrent-lock-corrected: 1928 states\n"; PASS=$((PASS+1))
else
    printf "[FAIL] concurrent-lock-corrected: expected 1928, got %s\n" "$states"; FAIL=$((FAIL+1))
fi

# 2. Invariant violation detection
out=$($SPECTRE verify examples/bank-account-violation.spec 2>&1)
check "bank-account-violation: CEGIS output present" "CEGIS Repair\|weakest precondition" "$out"
check "bank-account-violation: violation detected" "VIOLATION DETECTED" "$out"

# 3. Spec mining: guard extraction
out=$($SPECTRE mine --lang rust artifact/bank_account.rs 2>&1)
guards=$(printf '%s\n' "$out" | grep -c "require")
if [ "$guards" -ge 7 ]; then
    printf "[OK]   spec mining: %d require guards extracted\n" "$guards"; PASS=$((PASS+1))
else
    printf "[FAIL] spec mining: expected >= 7 guards, got %d\n" "$guards"; FAIL=$((FAIL+1))
fi

printf "\n=== %d passed, %d failed ===\n" "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
