// User Management System Example with Descriptions
// Demonstrates compound types, sets, and complex state transitions

type User = {
  id: int,
  name: str,
  active: bool,
  role: str
}

description "Collection of all registered users in the system"
var users: Set<User>

description "Next available user ID to assign"
var nextId: int

description "System initializes with no users and first ID set to 1"
init {
  users = {}
  nextId = 1
}

description "Adds a new user to the system with the next available ID"
action addUser(name: str, role: str) {
  users' = users.union({ { id: nextId, name: name, active: true, role: role } })
  nextId' = nextId + 1
}

description "Removes a user from the system by their ID"
action removeUser(id: int) {
  require users.exists(u => u.id = id)
  users' = users.filter(u => u.id != id)
}

description "Deactivates a user account, marking them as inactive"
action deactivateUser(id: int) {
  require users.exists(u => u.id = id && u.active)
  users' = users.map(u => 
    if (u.id = id) { { id: u.id, name: u.name, active: false, role: u.role } } else { u }
  )
}

description "Reactivates a previously deactivated user account"
action activateUser(id: int) {
  require users.exists(u => u.id = id && !u.active)
  users' = users.map(u => 
    if (u.id = id) { { id: u.id, name: u.name, active: true, role: u.role } } else { u }
  )
}

description "Changes the role of an existing user"
action changeRole(id: int, newRole: str) {
  require users.exists(u => u.id = id)
  users' = users.map(u => 
    if (u.id = id) { { id: u.id, name: u.name, active: u.active, role: newRole } } else { u }
  )
}

description "Validates that all user IDs are unique"
invariant uniqueIds {
  users.size() = users.map(u => u.id).toSet().size()
}

description "Ensures that nextId is always positive"
invariant nextIdPositive {
  nextId > 0
}

description "Ensures all user IDs are less than nextId (no ID reuse)"
invariant idConsistency {
  users.forall(u => u.id < nextId)
}

description "Verifies that users will eventually be added to the system"
temporal eventuallyUsersAdded {
  eventually users.size() > 0
}

description "If users exist, they will eventually all be deactivated"
temporal eventuallyDeactivated {
  always (users.exists(u => u.active) → eventually users.forall(u => !u.active))
}

description "If deactivated users exist, they will eventually all be reactivated"
temporal eventuallyAllActive {
  always (users.exists(u => !u.active) → eventually users.forall(u => u.active))
}
