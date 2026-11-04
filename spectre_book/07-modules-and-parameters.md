# Modules, parameters, and reuse

Constants let you change a spec's values without editing the logic. Modules let you split a large spec into named, reusable pieces.

---

## Constants — changing the spec without editing it

```spectre
const CAPACITY:  int = 10
const THRESHOLD: int = 7
var items: int

invariant not_full       { items <= CAPACITY  }
invariant below_threshold { items <  THRESHOLD }
```

Three ways to set constant values:

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

A module bundles variables, constants, init, actions, and invariants under a name.

```spectre
module Counter {
  var count: int
  const MAX: int = 10

  init { count = 0 }

  action increment { require count < MAX  count' = count + 1 }
  action decrement { require count > 0    count' = count - 1 }

  invariant non_negative  { count >= 0   }
  invariant within_bounds { count <= MAX }
}
```

Import and compose it in another spec:

```spectre
import Counter

var active: bool

init { active = true }

action toggle { active' = !active }

invariant active_non_negative { !active || Counter.count >= 0 }
```

```bash
spectre verify composed.spec
# Exhaustive verification: traversed 22 states (full reachable space)
```

---

## Module inheritance — parameterized reuse

```spectre
module Semaphore {
  const SLOTS: int = 1
  var held: int

  init { held = 0 }

  action acquire { require held < SLOTS  held' = held + 1 }
  action release { require held > 0      held' = held - 1 }

  invariant valid { held >= 0 && held <= SLOTS }
}

module Mutex  extends Semaphore { const SLOTS = 1 }
module RWLock extends Semaphore { const SLOTS = 3 }
```

> **How `extends` works:** Spectre modules use **parameterized macro substitution**, not OOP inheritance. `extends` copies all declarations from the parent module into the child scope and substitutes overridden constants. There is no dynamic dispatch, method override, or multiple independent instances of the same module. It is equivalent to textual include with constant replacement — a convenience for sharing invariants and structure, not a class hierarchy.

The same `valid` invariant covers both variants — you write it once and get it for free in every subtype.

```bash
spectre verify mutex.spec   # SLOTS=1 -> 3 states
spectre verify rwlock.spec  # SLOTS=3 -> 5 states
```

---

## Splitting a large spec into modules

**Before** — one flat file, hard to navigate:

```spectre
var user_count:    int
var session_count: int
var db_connections: int
const MAX_USERS:    int = 1000
const MAX_SESSIONS: int = 5000
const MAX_CONNS:    int = 20

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
  const MAX_USERS: int = 1000

  init { user_count = 0 }

  action add_user    { require user_count < MAX_USERS  user_count' = user_count + 1 }
  action remove_user { require user_count > 0          user_count' = user_count - 1 }

  invariant valid_count { user_count >= 0 && user_count <= MAX_USERS }
}
```

```spectre
// sessions.spec
module SessionManager {
  var session_count: int
  const MAX_SESSIONS: int = 5000

  init { session_count = 0 }

  action open_session  { require session_count < MAX_SESSIONS  session_count' = session_count + 1 }
  action close_session { require session_count > 0             session_count' = session_count - 1 }

  invariant valid_sessions { session_count >= 0 && session_count <= MAX_SESSIONS }
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
      v
Identify natural groupings
  (resources, actors, subsystems)
      │
      v
Extract each group into a module
      │
      v
spectre verify each module independently
      │
      v
Compose with import
      │
      v
Add cross-module invariants
      │
      v
spectre verify composition
```

---

## Constant best practices

- Always set a **sensible default** — specs should verify out of the box
- Use **ALL_CAPS** names for constants (`MAX`, `CAPACITY`, `TIMEOUT`) to distinguish from variables
- Keep constants **small for development** — verify with `--param N=3` before `--param N=100`
- Sweep constant values in CI to catch spec regressions across the whole value range
