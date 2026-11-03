# Chapter 7: Modules, parameters, and reuse

Parameters let you change a spec's constants without editing the file. Modules let you split a large spec into named, reusable pieces.

---

## Parameters — changing the spec without editing it

```spectre
param CAPACITY:  int = 10
param THRESHOLD: int = 7
var items: int

invariant not_full       { items <= CAPACITY  }
invariant below_threshold { items <  THRESHOLD }
```

Three ways to set parameter values:

- **Default** — just run `spectre verify queue.spec`
- **Override** — `spectre verify queue.spec --param CAPACITY=20`
- **Sweep** — `spectre verify queue.spec --param-range CAPACITY=1..20`

### Parameter sweep

```bash
spectre verify queue.spec --param-range CAPACITY=1..10
```

```
CAPACITY  States  Result
--------  ------  ------
       1       3  safe (exhaustive)
       2       6  safe (exhaustive)
       3      10  safe (exhaustive)
       4      15  safe (exhaustive)
       5      21  safe (exhaustive)
      10      66  safe (exhaustive)
```

Sweeping is useful for finding the exact value where a spec breaks or where the state space starts to explode.

---

## Modules — reusable spec components

A module bundles variables, params, init, actions, and invariants under a name.

```spectre
module Counter {
  var count: int
  param MAX: int = 10

  init { count = 0 }

  action increment { require count < MAX  count' = count + 1 }
  action decrement { require count > 0    count' = count - 1 }

  invariant non_negative  { count >= 0      }
  invariant within_bounds { count <= MAX    }
}
```

Import and compose it in another spec:

```spectre
import Counter

var active: bool

init { active = true }

action toggle { active' = not active }

invariant active_non_negative { active → Counter.count >= 0 }
```

```bash
spectre verify composed.spec
# ✓ safe — 22 states explored (exhaustive)
```

---

## Module inheritance — parameterized reuse

```spectre
module Semaphore {
  param SLOTS: int = 1
  var held: int

  init { held = 0 }

  action acquire { require held < SLOTS  held' = held + 1 }
  action release { require held > 0      held' = held - 1 }

  invariant valid { held >= 0 and held <= SLOTS }
}

module Mutex  extends Semaphore { param SLOTS = 1 }
module RWLock extends Semaphore { param SLOTS = 3 }
```

The same `valid` invariant covers both variants — you write it once and get it for free in every subtype.

```bash
spectre verify mutex.spec   # SLOTS=1 → 3 states
spectre verify rwlock.spec  # SLOTS=3 → 5 states
```

---

## Splitting a large spec into modules

**Before** — one flat file, hard to navigate:

```spectre
var user_count:    int
var session_count: int
var db_connections: int
param MAX_USERS:   int = 1000
param MAX_SESSIONS: int = 5000
param MAX_CONNS:   int = 20

action add_user    { ... }
action remove_user { ... }
action open_session { ... }
action close_session { ... }
action open_connection { ... }
action close_connection { ... }
// ... 40 more lines
```

**After** — three focused modules:

```spectre
// users.spec
module UserManager {
  var user_count: int
  param MAX_USERS: int = 1000

  init { user_count = 0 }

  action add_user    { require user_count < MAX_USERS  user_count' = user_count + 1 }
  action remove_user { require user_count > 0          user_count' = user_count - 1 }

  invariant valid_count { user_count >= 0 and user_count <= MAX_USERS }
}
```

```spectre
// sessions.spec
module SessionManager {
  var session_count: int
  param MAX_SESSIONS: int = 5000

  init { session_count = 0 }

  action open_session  { require session_count < MAX_SESSIONS  session_count' = session_count + 1 }
  action close_session { require session_count > 0             session_count' = session_count - 1 }

  invariant valid_sessions { session_count >= 0 and session_count <= MAX_SESSIONS }
}
```

```spectre
// app.spec — the composition
import UserManager
import SessionManager

invariant sessions_per_user {
  SessionManager.session_count <= UserManager.user_count * 10
}
```

Benefits of splitting:
- Each module verifies independently — catch bugs early
- Modules are reusable across projects
- Invariants at the composition layer catch cross-module bugs
- State spaces stay small per module

---

## Workflow: from flat to modular

```
Single flat spec
      │
      ▼
Identify natural groupings
  (resources, actors, subsystems)
      │
      ▼
Extract each group into a module
      │
      ▼
spectre verify each module independently
      │
      ▼
Compose with import
      │
      ▼
Add cross-module invariants
      │
      ▼
spectre verify composition
```

---

## Parameter best practices

- Always set a **sensible default** — specs should verify out of the box
- Use **ALL_CAPS** names for params (`MAX`, `CAPACITY`, `TIMEOUT`) to distinguish from variables
- Keep params **small for development** — verify with `--param N=3` before `--param N=100`
- Sweep params in CI to catch spec regressions across the whole value range
