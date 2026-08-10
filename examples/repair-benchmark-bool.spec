// Repair benchmark: Boolean assignment — activate should only fire when inactive.

var active: bool
var count:  int

init {
  active = false
  count  = 0
}

description "Activate: set active flag (missing guard — should require !active)"
action activate {
  active' = true
}

description "Process: increment count while active"
action process {
  require active == true
  require count < 3
  count' = count + 1
}

description "Deactivate: clear flag"
action deactivate {
  require active == true
  active' = false
}

description "Active flag is set at most once before being cleared"
invariant noDoubleActivation {
  !(active == true && count == 0)
}
