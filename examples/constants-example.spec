// Constants Example
// Demonstrates using constants for parameterized specifications

description "Maximum number of users allowed in the system"
const MAX_USERS: int = 100

description "Maximum number of retry attempts"
const MAX_RETRIES: int = 3

description "Default timeout in seconds"
const DEFAULT_TIMEOUT: int = 30

description "Server configuration"
const SERVER_HOST: str = "api.example.com"
const SERVER_PORT: int = 8080

description "Computed constants"
const MAX_TIMEOUT: int = DEFAULT_TIMEOUT * 2
const RETRY_DELAY: int = DEFAULT_TIMEOUT / 2

type User = {
  id: int,
  name: str,
  active: bool
}

description "Collection of users"
var users: Set<User>

description "Number of retry attempts made"
var retryCount: int

description "System initializes with no users and zero retries"
init {
  users = Set.empty()
  retryCount = 0
}

description "Adds a new user, enforcing maximum user limit"
action addUser(name: str) {
  require users.size() < MAX_USERS
  users' = users.union(Set.of({ id: users.size() + 1, name: name, active: true }))
}

description "Increments retry count up to maximum allowed"
action retry {
  require retryCount < MAX_RETRIES
  retryCount' = retryCount + 1
}

description "Resets retry count"
action resetRetries {
  retryCount' = 0
}

description "Ensures user count never exceeds maximum"
invariant userLimit {
  users.size() <= MAX_USERS
}

description "Ensures retry count never exceeds maximum"
invariant retryLimit {
  retryCount <= MAX_RETRIES
}

description "Verifies that users will eventually be added"
temporal eventuallyUsers {
  eventually users.size() > 0
}

