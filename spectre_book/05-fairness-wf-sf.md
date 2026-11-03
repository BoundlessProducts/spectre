# Chapter 5: Fairness — WF and SF

Temporal properties say *what* must eventually happen. Fairness conditions say *which executions count* when checking those properties. Without fairness, the verifier is allowed to assume a scheduler that permanently ignores actions — which makes most liveness properties unprovable.

---

## Why fairness matters

```spectre
var token: bool

init { token = false }

action grant  { require not token  token' = true  }
action revoke { require token      token' = false }

temporal eventually_granted { eventually(token) }
```

```bash
spectre verify token.spec
# ✗ violation — 'eventually_granted' not satisfied
# Counterexample: revoke fires forever, grant never fires
```

The spec *allows* granting — but the verifier found a valid execution where `grant` is always skipped. That execution is "unfair."

```
Without fairness:          With WF(grant):
  ●──revoke──●──revoke──►    ●──grant──► token=true  ✓
  ↑ grant is skipped         (grant must fire if always enabled)
```

---

## Weak fairness: WF(action)

> "If an action stays enabled, it must eventually fire."

```spectre
temporal wf_grant { WF(grant) }
```

Add this to the spec:

```spectre
var token: bool

init { token = false }

action grant  { require not token  token' = true  }
action revoke { require token      token' = false }

temporal eventually_granted { eventually(token)  }
temporal wf_grant           { WF(grant)          }
```

```bash
spectre verify token.spec
# ✓ safe — 2 states explored (exhaustive)
#   Temporal: eventually_granted ✓  wf_grant ✓
```

WF rules out executions where `grant` is permanently skipped while enabled.

### Mutex with progress guarantee

```spectre
enum ProcState { Waiting, Running, Done }
var p1:   ProcState
var p2:   ProcState
var lock: bool

init { p1 = Waiting  p2 = Waiting  lock = false }

action p1_acquire { require p1 = Waiting and not lock  p1' = Running  lock' = true  }
action p1_release { require p1 = Running               p1' = Done     lock' = false }
action p2_acquire { require p2 = Waiting and not lock  p2' = Running  lock' = true  }
action p2_release { require p2 = Running               p2' = Done     lock' = false }

invariant mutex      { not (p1 = Running and p2 = Running) }
temporal p1_progress { WF(p1_acquire) }
temporal p2_progress { WF(p2_acquire) }
temporal p1_finishes { eventually(p1 = Done) }
temporal p2_finishes { eventually(p2 = Done) }
```

```bash
spectre verify mutex.spec
# ✓ safe — 9 states explored (exhaustive)
```

Without `WF(p1_acquire)`, `p1_finishes` would fail — the scheduler could run p2 forever and starve p1.

---

## Strong fairness: SF(action)

> "If an action is enabled infinitely often, it must fire infinitely often."

The difference from WF: SF handles actions that are *intermittently* enabled.

```spectre
var connected: bool
var retries:   int

init { connected = false  retries = 0 }

action disconnect { require connected      connected' = false            }
action connect    { require not connected  connected' = true             }
action retry      { require not connected  retries'   = retries + 1      }

temporal sf_retry { SF(retry) }
temporal eventually_connected { eventually(connected) }
```

`retry` is only enabled when `connected = false` — which isn't always the case. WF(retry) wouldn't help here because `retry` isn't *continuously* enabled. SF(retry) says: whenever `retry` becomes enabled again, it must fire.

### WF vs SF at a glance

| | WF(A) | SF(A) |
|-|-------|-------|
| Fires when | Continuously enabled | Infinitely often enabled |
| Strength | Weaker | Stronger |
| Use for | Lock acquisition, queue dequeue | Retry, intermittent tasks |
| Example | Mutex `acquire` | Network `retry` |

---

## WF and SF in the same spec

You can mix them freely:

```spectre
temporal acquire_wf   { WF(acquire)  }   // always-enabled action
temporal retry_sf     { SF(retry)    }   // intermittently-enabled action
temporal done_for_all { eventually(p1 = Done and p2 = Done) }
```

Spectre checks all temporal properties together.

---

## When the BFS heuristic isn't enough

Spectre's built-in BFS uses a cycle-filtering heuristic for fairness. It works for the majority of specs but can give wrong answers when:

- Multiple fairness conditions interact
- An action is enabled in some states of a cycle but not others
- The spec has deeply nested liveness goals

For these cases, use the TLC backend:

```bash
spectre verify mutex.spec --backend tlc
```

TLC performs proper SCC-based fairness checking — the same algorithm that's been validated in production TLA+ tooling for two decades. It translates `WF(action)` to `WF_vars(Action)` in TLA+ and lets TLC handle the rest.

```
WF/SF in Spectre
       │
       ▼
spectre verify (BFS heuristic)
       │
  confident? ──yes──► done
       │
       no (complex fairness, large spec)
       │
       ▼
spectre verify --backend tlc
(SCC-based, validated)
```

See Chapter 15 for full TLC and Apalache backend details.

---

## Fairness checklist

When writing a spec with liveness properties:

- Add `WF(action)` for every action that represents a "service obligation" (lock acquisition, queue drain, message delivery)
- Add `SF(action)` for actions that are intermittently available (retries, reconnects, elections)
- If `eventually(P)` fails unexpectedly, add fairness before debugging the spec logic
- For critical liveness proofs, validate with `--backend tlc`
