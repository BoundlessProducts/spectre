// Pure Functions Example with Descriptions
// Demonstrates pure functions for computations without state changes

type User = {
  id: int,
  name: str,
  age: int,
  active: bool
}

description "Collection of all users in the system"
var users: Set<User>

description "Counter tracking number of operations"
var counter: int

description "System starts with no users and counter at zero"
init {
  users = {}
  counter = 0
}

description "Adds two integers together"
fun add(x: int, y: int): int {
  return x + y
}

description "Multiplies two integers"
fun multiply(x: int, y: int): int {
  return x * y
}

description "Returns the maximum of two integers"
fun max(a: int, b: int): int {
  if (a > b) {
    return a
  } else {
    return b
  }
}

description "Returns the minimum of two integers"
fun min(a: int, b: int): int {
  if (a < b) {
    return a
  } else {
    return b
  }
}

description "Counts the number of active users in a set"
fun countActiveUsers(userSet: Set<User>): int {
  return userSet.filter(u => u.active).size()
}

description "Extracts all user names from a user set into a list"
fun getUserNames(userSet: Set<User>): List<str> {
  return userSet.map(u => u.name).toList()
}

description "Finds a user by their ID, returning an optional result"
fun findUserById(userSet: Set<User>, id: int): Option<User> {
  let matches = userSet.filter(u => u.id = id)
  if (matches.size() = 0) {
    return Option.none()
  } else {
    return Option.some(matches.head())
  }
}

description "Calculates the average age of users in a set"
fun averageAge(userSet: Set<User>): float {
  return if (userSet.size() = 0) {
    0.0
  } else {
    userSet.map(u => u.age).reduce(0, (acc, age) => acc + age) / userSet.size()
  }
}

description "Checks if a user is eligible (active and 18 or older)"
fun isEligible(user: User): bool {
  return user.active && user.age >= 18
}

description "Formats a user's display name with their ID"
fun getUserDisplayName(user: User): str {
  return user.name + " (ID: " + user.id.toString() + ")"
}

description "Validates that a user ID is within acceptable range"
fun isValidUserId(id: int): bool {
  return id > 0 && id < 1000000
}

description "Validates that a user name is not empty and within length limits"
fun isValidUserName(name: str): bool {
  return name.length() > 0 && name.length() <= 100
}

description "Calculates a score for a user based on age and active status"
fun calculateUserScore(user: User): int {
  let baseScore = 100
  let ageBonus = if (user.age >= 18) { 50 } else { 0 }
  let activeBonus = if (user.active) { 25 } else { 0 }
  return baseScore + ageBonus + activeBonus
}

description "Calculates the factorial of a number recursively"
fun factorial(n: int): int {
  if (n <= 1) {
    return 1
  } else {
    return n * factorial(n - 1)
  }
}

description "Sums all integers in a range recursively"
fun sumRange(start: int, end: int): int {
  if (start > end) {
    return 0
  } else {
    if (start = end) {
      return start
    } else {
      return start + sumRange(start + 1, end)
    }
  }
}

description "Adds a new user to the system, validating inputs using pure functions"
action addUser(id: int, name: str, age: int) {
  require isValidUserId(id)
  require isValidUserName(name)
  require !users.exists(u => u.id = id)
  
  let newUser = { id: id, name: name, age: age, active: true }
  users' = users.union({ { id: id, name: name, age: age, active: true } })
  counter' = add(counter, 1)
}

description "Removes a user from the system by ID"
action removeUser(id: int) {
  require users.exists(u => u.id = id)
  users' = users.filter(u => u.id != id)
  counter' = counter - 1
}

description "Activates a user account"
action activateUser(id: int) {
  require users.exists(u => u.id = id)
  users' = users.map(u => 
    if (u.id = id) { { id: u.id, name: u.name, age: u.age, active: true } } else { u }
  )
}

description "Ensures active user count is always non-negative"
invariant activeUserCount {
  countActiveUsers(users) >= 0
}

description "Ensures counter matches the number of users"
invariant counterMatchesUsers {
  counter = users.size()
}

description "Validates that all users have valid IDs and names"
invariant allUsersValid {
  users.forall(u => isValidUserId(u.id) && isValidUserName(u.name))
}

description "Ensures eligible users are always active"
invariant eligibleUsersActive {
  users.forall(u => !isEligible(u) || u.active)
}

description "Verifies that active users will eventually exist"
temporal eventuallyActiveUsers {
  eventually countActiveUsers(users) > 0
}

description "Verifies that average age will eventually reach 30 or higher"
temporal eventuallyHighAverageAge {
  eventually averageAge(users) >= 30.0
}

