// Controller Module
// Implements the elevator assignment logic and optimization algorithms
//
// This module provides utility functions for the elevator controller:
// - Distance calculation between elevator and floor
// - Same-direction detection for optimization
// - Sufficient gap checking (3 floors minimum for same-direction assignment)
// - System configuration constants

module ControllerModule {
  description "Number of elevators in the system"
  public const NUM_ELEVATORS: int = 4

  description "Number of floors in the building"
  public const NUM_FLOORS: int = 20

  description "Maximum number of users in the system"
  public const MAX_USERS: int = 30

  description "Minimum floor gap for same-direction assignment (3 floors)"
  public const MIN_FLOOR_GAP: int = 3

  description "Calculate distance between elevator and floor"
  description "Returns absolute difference in floors"
  public fun distance(elevatorFloor: int, targetFloor: int): int {
    return if (elevatorFloor > targetFloor) {
      elevatorFloor - targetFloor
    } else {
      targetFloor - elevatorFloor
    }
  }

  description "Check if elevator is moving in same direction as user request"
  description "Returns true if elevator direction matches user direction"
  public fun sameDirection(elevatorDir: int, userDir: int): bool {
    return elevatorDir = userDir && !(elevatorDir = 0)
  }

  description "Check if elevator is far enough ahead for same-direction assignment"
  description "For up requests: elevator must be at least MIN_FLOOR_GAP floors below"
  description "For down requests: elevator must be at least MIN_FLOOR_GAP floors above"
  public fun sufficientGap(elevatorFloor: int, userFloor: int, elevatorDir: int, userDir: int): bool {
    return if (userDir = 1) {
      elevatorDir = 1 && elevatorFloor < userFloor && (userFloor - elevatorFloor) >= MIN_FLOOR_GAP
    } else {
      if (userDir = -1) {
        elevatorDir = -1 && elevatorFloor > userFloor && (elevatorFloor - userFloor) >= MIN_FLOOR_GAP
      } else {
        false
      }
    }
  }
}

