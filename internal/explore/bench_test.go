package explore_test

import (
	"testing"

	"github.com/BoundlessProducts/spectre/internal/exec"
	"github.com/BoundlessProducts/spectre/internal/explore"
	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
)

func buildBenchSM(b *testing.B, src string) *exec.StateMachine {
	b.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		b.Fatalf("parse errors: %v", p.Errors())
	}
	sm, err := exec.NewStateMachine(file)
	if err != nil {
		b.Fatalf("NewStateMachine: %v", err)
	}
	return sm
}

func bfsUntilDone(b *testing.B, sm *exec.StateMachine, maxStates int) {
	b.Helper()
	e := explore.NewExplorer(sm)
	e.SetMaxStates(maxStates)
	e.SetSilent(true)
	if _, err := e.ExploreBFS(); err != nil {
		b.Fatalf("BFS error: %v", err)
	}
}

// BenchmarkBFS_Counter measures BFS on a simple 11-state counter.
func BenchmarkBFS_Counter(b *testing.B) {
	const src = `
var n: int
init { n = 0 }
action incr { require n < 10  n' = n + 1 }
action decr { require n > 0   n' = n - 1 }
invariant bounded { n >= 0 && n <= 10 }
`
	sm := buildBenchSM(b, src)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bfsUntilDone(b, sm, 200)
	}
}

// BenchmarkBFS_BankAccount measures BFS on a bank account with 3 variables.
func BenchmarkBFS_BankAccount(b *testing.B) {
	const src = `
var balance: int
var deposits: int
var withdrawals: int
init { balance = 0  deposits = 0  withdrawals = 0 }
action deposit  { require balance < 50  balance' = balance + 10  deposits' = deposits + 10  withdrawals' = withdrawals }
action withdraw { require balance >= 10  balance' = balance - 10  deposits' = deposits  withdrawals' = withdrawals + 10 }
invariant non_negative { balance >= 0 }
invariant accounting   { balance == deposits - withdrawals }
`
	sm := buildBenchSM(b, src)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bfsUntilDone(b, sm, 2000)
	}
}

// BenchmarkBFS_Mutex measures BFS on a two-process mutex (~9 states).
func BenchmarkBFS_Mutex(b *testing.B) {
	const src = `
enum PS { Idle, Waiting, Critical }
var p1: PS
var p2: PS
var locked: bool
init { p1 = PS.Idle  p2 = PS.Idle  locked = false }
action p1_req { require p1 == PS.Idle    p1' = PS.Waiting  p2' = p2  locked' = locked }
action p1_acq { require p1 == PS.Waiting && !locked  p1' = PS.Critical  p2' = p2  locked' = true }
action p1_rel { require p1 == PS.Critical  p1' = PS.Idle  p2' = p2  locked' = false }
action p2_req { require p2 == PS.Idle    p1' = p1  p2' = PS.Waiting  locked' = locked }
action p2_acq { require p2 == PS.Waiting && !locked  p1' = p1  p2' = PS.Critical  locked' = true }
action p2_rel { require p2 == PS.Critical  p1' = p1  p2' = PS.Idle  locked' = false }
invariant mutex { !(p1 == PS.Critical && p2 == PS.Critical) }
`
	sm := buildBenchSM(b, src)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bfsUntilDone(b, sm, 500)
	}
}

// BenchmarkBFS_ResourceAlloc measures BFS on a 3-client resource allocator.
func BenchmarkBFS_ResourceAlloc(b *testing.B) {
	const src = `
var free: int
var ha: int
var hb: int
var hc: int
init { free = 4  ha = 0  hb = 0  hc = 0 }
action aa { require free > 0  free' = free-1  ha' = ha+1  hb' = hb  hc' = hc }
action ab { require free > 0  free' = free-1  ha' = ha    hb' = hb+1  hc' = hc }
action ac { require free > 0  free' = free-1  ha' = ha    hb' = hb    hc' = hc+1 }
action fa { require ha > 0    free' = free+1  ha' = ha-1  hb' = hb    hc' = hc }
action fb { require hb > 0    free' = free+1  ha' = ha    hb' = hb-1  hc' = hc }
action fc { require hc > 0    free' = free+1  ha' = ha    hb' = hb    hc' = hc-1 }
invariant conservation { free + ha + hb + hc == 4 }
invariant non_neg { free >= 0 && ha >= 0 && hb >= 0 && hc >= 0 }
`
	sm := buildBenchSM(b, src)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bfsUntilDone(b, sm, 5000)
	}
}
