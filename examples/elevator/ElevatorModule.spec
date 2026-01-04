// Elevator Module
// Defines the basic elevator state and operations (movement, target management)
//
// This module represents a single elevator with:
// - Current floor position (0-19)
// - Direction of movement (up/down/idle)
// - Set of target floors to visit

module ElevatorModule {
  description "Current floor where the elevator is located (0-19)"
  public var currentFloor: int

  description "Direction the elevator is moving: 1 = up, -1 = down, 0 = idle"
  public var direction: int

  description "Set of floors the elevator needs to visit (target floors)"
  public var targetFloors: Set<int>

  description "Move elevator up one floor"
  description "Can only move up if not at top floor and has targets above"
  public action moveUp {
    require currentFloor < 19
    require targetFloors.exists(f => f > currentFloor)
    currentFloor' = currentFloor + 1
    direction' = 1
  }

  description "Move elevator down one floor"
  description "Can only move down if not at bottom floor and has targets below"
  public action moveDown {
    require currentFloor > 0
    require targetFloors.exists(f => f < currentFloor)
    currentFloor' = currentFloor - 1
    direction' = -1
  }

  description "Add a target floor to the elevator's queue"
  description "The floor must be valid (0-19) and not already a target"
  public action addTarget(floor: int) {
    require floor >= 0 && floor <= 19
    require !targetFloors.contains(floor)
    targetFloors' = targetFloors.union({ floor })
  }

  description "Remove current floor from targets when elevator arrives"
  description "Elevator becomes idle if no more targets"
  public action arriveAtFloor {
    require targetFloors.contains(currentFloor)
    targetFloors' = targetFloors.filter(f => f != currentFloor)
    direction' = if (targetFloors.size() = 0) {
      0
    } else {
      if (targetFloors.exists(f => f > currentFloor)) {
        1
      } else {
        -1
      }
    }
  }

  description "Elevator becomes idle when no targets remain"
  public action becomeIdle {
    require targetFloors.size() = 0
    direction' = 0
  }

  description "Invariant: Elevator floor must be valid (0-19)"
  public invariant validFloor {
    currentFloor >= 0 && currentFloor <= 19
  }

  description "Invariant: Direction must be -1, 0, or 1"
  public invariant validDirection {
    direction >= -1 && direction <= 1
  }

  description "Invariant: Target floors must be valid (0-19)"
  public invariant validTargets {
    targetFloors.forall(f => f >= 0 && f <= 19)
  }

  description "Invariant: If elevator is moving, it must have targets"
  public invariant movingImpliesTargets {
    direction = 0 || targetFloors.size() > 0
  }
}

