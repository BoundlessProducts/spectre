# Chapter 4: Temporal properties and liveness

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

Full example — traffic light that must never show `Off`:

```spectre
enum Light { Red, Yellow, Green }
var signal: Light

init { signal = Red }

action to_green  { require signal = Red     signal' = Green  }
action to_yellow { require signal = Green   signal' = Yellow }
action to_red    { require signal = Yellow  signal' = Red    }

invariant always_lit { signal != Off }
temporal  never_off  { always(signal != Off) }
```

```bash
spectre verify traffic.spec
# ✓ safe — 3 states explored (exhaustive)
#   Invariants: always_lit ✓
#   Temporal:   never_off ✓
```

---

## `eventually` — something must happen

```spectre
temporal reaches_max { eventually(value = MAX) }
```

This says: on every execution path, `value` must at some point equal `MAX`.

Counter that counts to MAX and stops:

```spectre
var value: int
param MAX: int = 4

init { value = 0 }

action increment { require value < MAX  value' = value + 1 }

invariant non_negative { value >= 0 }
temporal  reaches_max  { eventually(value = MAX) }
```

```bash
spectre verify counter.spec
# ✓ safe — 5 states explored (exhaustive)
#   Temporal: reaches_max ✓
```

Now add a `reset` action and watch the property fail:

```spectre
action reset { value' = 0 }
```

```bash
spectre verify counter.spec
# ✗ violation — temporal property 'reaches_max' not satisfied
#
# Counterexample (cycle):
#   step 0  init:      value = 0
#   step 1  increment: value = 1
#   step 2  reset:     value = 0   ← loops here forever
#   ...
#   value never reaches 4
```

The BFS found a cycle where `reset` fires repeatedly and `value` never reaches `MAX`.

---

## `until` — hold one condition until another

```spectre
temporal safe_until_done { (errors = 0) until (phase = Done) }
```

This says: `errors = 0` must hold on every step *until* `phase = Done` is reached.

Download manager example:

```spectre
enum Phase { Downloading, Done }
var phase:  Phase
var errors: int

init { phase = Downloading  errors = 0 }

action download { require phase = Downloading  phase' = Done         }
action fail     { require phase = Downloading  errors' = errors + 1  }

temporal no_errors_until_done { (errors = 0) until (phase = Done) }
```

```bash
spectre verify download.spec
# ✗ violation — 'no_errors_until_done' not satisfied
# step 1: fail → errors = 1, phase still Downloading
```

Fix: add `require errors = 0` to `fail`, or remove it if errors are not allowed before completion.

---

## `leads_to` — if P then eventually Q

```spectre
temporal request_served { always(pending → eventually(not pending)) }
```

Every time a request becomes pending, it must eventually be cleared. This is the "leads-to" pattern: `always(P → eventually(Q))`.

Simple request/response server:

```spectre
var pending: bool
var served:  bool

init { pending = false  served = false }

action request { require not pending  pending' = true               }
action serve   { require pending      pending' = false  served' = true }
action reset   { served' = false                                    }

temporal always_served { always(pending → eventually(not pending)) }
```

```bash
spectre verify server.spec
# ✓ safe — 4 states explored (exhaustive)
#   Temporal: always_served ✓
```

Remove the `serve` action and re-verify:

```bash
spectre verify server-broken.spec
# ✗ violation — 'always_served' not satisfied
# pending becomes true and stays true — no serve action ever fires
```

---

## Combining temporal properties

A spec can have multiple temporal properties. Spectre checks all of them.

```spectre
var value:    int
var request:  bool
var response: bool
param MAX: int = 5

init { value = 0  request = false  response = false }

action increment { require value < MAX  value' = value + 1            }
action req       { require not request  request' = true               }
action respond   { require request      request' = false  response' = true }

temporal always_bounded   { always(value <= MAX)                      }
temporal reaches_max      { eventually(value = MAX)                   }
temporal request_answered { always(request → eventually(not request)) }
```

What each property covers:

```
States:  ●───●───●───●───●───●
        v=0 v=1 v=2 v=3 v=4 v=5=MAX

always_bounded:     every node   →  value <= MAX          ✓ per-state
reaches_max:        some path    →  reaches value = MAX   ✓ across paths
request_answered:   each request →  eventually cleared    ✓ per-event pair
```

---

## Temporal verification with large state spaces

If BFS hits `--max-states`, temporal properties are only checked over explored states:

```bash
spectre verify large.spec --max-states 5000
# bounded — hit state limit
# temporal properties checked over 5000 states only
# Use --backend tlc for exhaustive temporal verification
```

Use `--backend tlc` or `--backend apalache` for exhaustive temporal verification on large specs (Chapter 15).

---

## Quick-reference

| Operator | Meaning | Use when |
|----------|---------|----------|
| `always(P)` | P holds in every reachable state | Global safety, expressed temporally |
| `eventually(P)` | Every path eventually reaches P | Progress / reachability goals |
| `P until Q` | P holds on all steps until Q | Precondition that must hold until a goal |
| `always(P → eventually(Q))` | Every P event leads to Q | Request/response, lock/unlock pairs |
| `WF(action)` | Action fires if continuously enabled | Chapter 5 |
| `SF(action)` | Action fires if infinitely often enabled | Chapter 5 |
