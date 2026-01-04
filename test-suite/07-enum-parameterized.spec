// Test 7: Enum in Parameterized Action
// Action takes enum parameter
// Expected: All enum values should be explored

enum Color {
  Red,
  Green,
  Blue
}

var selectedColor: Color

init {
  selectedColor = Color.Red
}

action setColor(c: Color) {
  selectedColor' = c
}

