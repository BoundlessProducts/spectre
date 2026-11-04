# Fairness — WF and SF

Temporal properties say *what* must eventually happen. Fairness conditions say *which executions count* when checking those properties. Without fairness, the verifier is allowed to assume a scheduler that permanently ignores actions — which makes most liveness properties unprovable.

---

## Why fairness matters

```spectre
var token: bool

init { token = false }

action grant  { require !token  token' = true  }
action revoke { require token   token' = false }

temporal eventually_granted { eventually(token) }
```

```bash
spectre verify token.spec
# Exhaustive verification: traversed 2 states (full reachable space)
# Found no violations.
```

BFS finds no violation because from the initial state (`token = false`), only `grant` is enabled — so `token` is forced to become true. BFS checks reachability, and `token = true` is reachable.

The *fairness* issue would appear on a spec where two actions compete. In the mutex below, without `WF(p1_acquire)` the scheduler could run p2 forever and starve p1. Declare `WF(action)` to document the fairness intent and validate it with `--backend tlc`.

```
Without WF (token spec):   token starts false -> only grant fires -> no issue
With WF(grant):            documents that grant must fire if continuously enabled
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

action grant  { require !token  token' = true  }
action revoke { require token   token' = false }

temporal eventually_granted { eventually(token)  }
temporal wf_grant           { WF(grant)          }
```

```bash
spectre verify token.spec
# Exhaustive verification: traversed 2 states (full reachable space)
# Found no violations.
```

WF rules out executions where `grant` is permanently skipped while enabled.

### Mutex with progress guarantee

```spectre
enum ProcState { Waiting, Running, Done }
var p1:   ProcState
var p2:   ProcState
var lock: bool

init { p1 = ProcState.Waiting  p2 = ProcState.Waiting  lock = false }

action p1_acquire { require p1 == ProcState.Waiting && !lock  p1' = ProcState.Running  lock' = true  }
action p1_release { require p1 == ProcState.Running           p1' = ProcState.Done     lock' = false }
action p2_acquire { require p2 == ProcState.Waiting && !lock  p2' = ProcState.Running  lock' = true  }
action p2_release { require p2 == ProcState.Running           p2' = ProcState.Done     lock' = false }

invariant mutex      { !(p1 == ProcState.Running && p2 == ProcState.Running) }
temporal p1_progress { WF(p1_acquire) }
temporal p2_progress { WF(p2_acquire) }
temporal p1_finishes { eventually(p1 == ProcState.Done) }
temporal p2_finishes { eventually(p2 == ProcState.Done) }
```

```bash
spectre verify mutex.spec
# Exhaustive verification: traversed 8 states (full reachable space)
# Found no violations.
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

action disconnect { require connected   connected' = false            }
action connect    { require !connected  connected' = true             }
action retry      { require !connected  retries'   = retries + 1      }

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
temporal done_for_all { eventually(p1 == ProcState.Done && p2 == ProcState.Done) }
```

Spectre checks all temporal properties together.

---

## BFS for learning, TLC for validation

BFS is the learning backend. It explores the reachable trace space and reports cases where an enabled action is never taken — useful for developing intuition about your spec's liveness behaviour. For validated fairness results, use the TLC backend:

```bash
spectre verify mutex.spec --backend tlc
```

TLC performs proper SCC-based (strongly connected component) fairness checking — the same algorithm validated in production TLA+ tooling for two decades. It translates `WF(action)` to `WF_vars(Action)` in TLA+ and lets TLC's cycle-detection engine do the rest. BFS is great for rapid iteration; TLC gives the authoritative answer.

```
WF/SF in Spectre
       │
       v
spectre verify  (BFS — learning mode, trace exploration)
       │
  confident? --yes--> done
       │
       no (need validated fairness, complex interactions, large spec)
       │
       v
spectre verify --backend tlc
(SCC-based, validated liveness)
```

See the TLA+ and Apalache chapter for full TLC and Apalache backend details.

---

## Fairness checklist

When writing a spec with liveness properties:

- Add `WF(action)` for every action that represents a "service obligation" (lock acquisition, queue drain, message delivery)
- Add `SF(action)` for actions that are intermittently available (retries, reconnects, elections)
- If `eventually(P)` fails unexpectedly, add fairness before debugging the spec logic
- For critical liveness proofs, validate with `--backend tlc`
