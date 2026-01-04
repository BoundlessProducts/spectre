// User Module
// Defines user state and request operations
//
// This module represents a user who can request elevator service:
// - User's current floor location
// - Direction request (up/down)
// - Waiting status and assigned elevator ID

module UserModule {
  description "Current floor where the user is located (0-19)"
  public var floor: int

  description "Direction the user wants to go: 1 = up, -1 = down, 0 = no request"
  public var direction: int

  description "Whether the user is waiting for an elevator (true) or riding (false)"
  public var waiting: bool

  description "ID of the elevator assigned to this user, or -1 if none"
  public var assignedElevator: int

  description "User presses the up button from their current floor"
  description "User must not already have a request and must not be at top floor"
  public action pressUp(floorNum: int) {
    require floorNum >= 0 && floorNum < 19
    require direction = 0
    require !waiting
    floor' = floorNum
    direction' = 1
    waiting' = true
    assignedElevator' = -1
  }

  description "User presses the down button from their current floor"
  description "User must not already have a request and must not be at bottom floor"
  public action pressDown(floorNum: int) {
    require floorNum > 0 && floorNum <= 19
    require direction = 0
    require !waiting
    floor' = floorNum
    direction' = -1
    waiting' = true
    assignedElevator' = -1
  }

  description "Assign an elevator to this user"
  description "User must be waiting and elevator ID must be valid (0-3)"
  public action assignElevator(elevatorId: int) {
    require waiting
    require elevatorId >= 0 && elevatorId <= 3
    require assignedElevator = -1
    assignedElevator' = elevatorId
  }

  description "User boards the elevator"
  description "Elevator must have arrived at user's floor"
  public action boardElevator {
    require waiting
    require assignedElevator >= 0
    waiting' = false
  }

  description "User exits the elevator at their destination"
  description "User must be riding (not waiting)"
  public action exitElevator {
    require !waiting
    require assignedElevator >= 0
    direction' = 0
    assignedElevator' = -1
  }

  description "Invariant: User floor must be valid (0-19)"
  public invariant validFloor {
    floor >= 0 && floor <= 19
  }

  description "Invariant: Direction must be -1, 0, or 1"
  public invariant validDirection {
    direction >= -1 && direction <= 1
  }

  description "Invariant: If user is waiting, they must have a direction request"
  public invariant waitingImpliesDirection {
    !waiting || !(direction = 0)
  }

  description "Invariant: Assigned elevator ID must be -1 or 0-3"
  public invariant validElevatorId {
    assignedElevator >= -1 && assignedElevator <= 3
  }
}

