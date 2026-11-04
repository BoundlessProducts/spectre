# Temporal properties and liveness

Invariants catch bad states. Temporal properties catch bad *trajectories* — things that should eventually happen but don't, or conditions that should hold across a sequence of steps.

---

## Safety vs liveness

| Property kind | Question it answers | Spectre keyword |
|--------------|---------------------|----------------|
| Safety | "Does anything bad ever happen?" | `invariant` |
| Liveness | "Does something good eventually happen?" | `temporal` |

You need both. A spec where nothing ever happens is trivially safe but useless.

---

## `always` — a condition that holds forever

```spectre
temporal always_positive { always(balance >= 0) }
```

This is equivalent to an invariant — but expressed in the temporal language. Use `invariant` for simple per-state checks; use `always` when you're building up compound temporal expressions.

Full example — traffic light that cycles through valid states:

```spectre
enum Light { Red, Yellow, Green }
var signal: Light

init { signal = Light.Red }

action to_green  { require signal == Light.Red     signal' = Light.Green  }
action to_yellow { require signal == Light.Green   signal' = Light.Yellow }
action to_red    { require signal == Light.Yellow  signal' = Light.Red    }

invariant always_lit { signal == Light.Red || signal == Light.Yellow || signal == Light.Green }
temporal  never_off  { always(signal == Light.Red || signal == Light.Yellow || signal == Light.Green) }
```

```bash
spectre verify traffic.spec
# Exhaustive verification: traversed 3 states (full reachable space)
# Found no violations.
```

---

## `eventually` — something must happen

```spectre
temporal reaches_max { eventually(value == MAX) }
```

This says: on every execution path, `value` must at some point equal `MAX`.

Counter that counts to MAX and stops:

```spectre
var value: int
const MAX: int = 4

init { value = 0 }

action increment { require value < MAX  value' = value + 1 }

invariant non_negative { value >= 0 }
temporal  reaches_max  { eventually(value == MAX) }
```

```bash
spectre verify counter.spec
# Exhaustive verification: traversed 5 states (full reachable space)
# Found no violations.
```

Now add a `reset` action:

```spectre
action reset { value' = 0 }
```

```bash
spectre verify counter.spec
# Exhaustive verification: traversed 5 states (full reachable space)
# Found no violations.
```

BFS finds no violation because `value == MAX` is still reachable (increment four times). BFS is the **learning backend** — it checks whether a property *can* become true in the reachable state space, not whether it is *guaranteed* to become true on every execution. Because `reset` can fire repeatedly and deny increment, the property can fail on an unfair execution. To detect that, add `WF(increment)` and validate with `--backend tlc`:

```bash
spectre verify counter.spec --backend tlc
# TLC: VIOLATION — reaches_max not satisfied under a fair scheduler
# (reset can fire indefinitely, starving increment)
```

See the Fairness chapter for WF/SF and the TLA+ and Apalache chapter for the TLC backend.

---

## `until` — hold one condition until another

```spectre
temporal safe_until_done { (errors == 0) until (phase == Phase.Done) }
```

This says: `errors == 0` must hold on every step *until* `phase == Phase.Done` is reached.

Download manager example:

```spectre
enum Phase { Downloading, Done }
var phase: Phase
var error: bool

init { phase = Phase.Downloading  error = false }

action download { require phase == Phase.Downloading  phase' = Phase.Done }
action fail     { require phase == Phase.Downloading  error' = true       }

temporal no_errors_until_done { (!error) until (phase == Phase.Done) }
```

```bash
spectre verify download.spec
# Exhaustive verification: traversed 4 states (full reachable space)
# Violations:
# 1. Temporal Property (no_errors_until_done): P doesn't hold before Q
#    Counterexample: fail
```

`fail` fires before `download`, setting `error=true` while `phase` is still `Downloading`. Fix: add `require !error` to `fail`, or remove the `fail` action if errors are not allowed before completion.

---

## Leads-to — if P then eventually Q

```spectre
temporal request_served { always(pending -> eventually(!pending)) }
```

Every time a request becomes pending, it must eventually be cleared. This is the "leads-to" pattern: `always(P -> eventually(Q))`.

**Spectre uses LTL (Linear Temporal Logic) semantics over execution traces.** `always(P -> eventually(Q))` means: in every suffix of every execution where P holds, Q eventually becomes true. This is equivalent to the leads-to operator P ~> Q. BFS explores traces and checks reachability, which is not sufficient to validate leads-to — use `--backend tlc` or `--backend apalache` for this pattern.

```bash
spectre verify server.spec --backend tlc
# -> passes if serve is reachable from every state where pending=true
```

The correct way to write and verify a leads-to property:

```spectre
var pending: bool
var served:  bool

init { pending = false  served = false }

action request { require !pending  pending' = true               }
action serve   { require pending   pending' = false  served' = true }

temporal request_served { always(pending -> eventually(!pending)) }
temporal wf_serve       { WF(serve) }
```

```bash
spectre verify server.spec --backend tlc
# TLC: verified — pending always leads to !pending (WF ensures serve fires)
```

See the TLA+ and Apalache chapter for the TLC backend details.

---

## Combining temporal properties

A spec can have multiple temporal properties. Spectre checks all of them.

```spectre
var value:    int
var response: bool
const MAX: int = 5

init { value = 0  response = false }

action increment { require value < MAX  value' = value + 1 }
action respond   { response' = true                        }

temporal always_bounded { always(value <= MAX)         }
temporal reaches_max    { eventually(value == MAX)     }
temporal gets_response  { eventually(response)         }
```

What each property covers:

```
States:  *---*---*---*---*---*
        v=0 v=1 v=2 v=3 v=4 v=5=MAX

always_bounded:  every node  ->  value <= MAX       (ok) per-state
reaches_max:     some path   ->  reaches value=MAX  (ok) across paths
gets_response:   every run   ->  response fires     (ok) reachability
```

For `always(P -> eventually(Q))` (leads-to), use `--backend tlc` — BFS trace exploration does not validate it correctly.

---

## Temporal verification with large state spaces

If BFS hits `--max-states`, temporal properties are only checked over explored states:

```bash
spectre verify large.spec --max-states 5000
# bounded -- hit state limit
# temporal properties checked over 5000 states only
# Use --backend tlc for exhaustive temporal verification
```

Use `--backend tlc` or `--backend apalache` for exhaustive temporal verification on large specs (see the TLA+ and Apalache chapter).

---

## Quick-reference

| Operator | Meaning | Use when |
|----------|---------|----------|
| `always(P)` | P holds in every reachable state | Global safety, expressed temporally |
| `eventually(P)` | Every path eventually reaches P | Progress / reachability goals |
| `P until Q` | P holds on all steps until Q | Precondition that must hold until a goal |
| `always(P -> eventually(Q))` | Every P event leads to Q | Request/response, lock/unlock pairs — use `--backend tlc` |
| `WF(action)` | Action fires if continuously enabled | Fairness chapter |
| `SF(action)` | Action fires if infinitely often enabled | Fairness chapter |
