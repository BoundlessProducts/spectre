// Simple counter — fully supported.
pub struct Counter {
    pub value: i64,
    pub max: i64,
}

impl Counter {
    pub fn new(max: i64) -> Self {
        Counter { value: 0, max }
    }

    pub fn increment(&mut self) {
        assert!(self.value < self.max);
        self.value += 1;
    }

    pub fn decrement(&mut self) {
        assert!(self.value > 0);
        self.value -= 1;
    }

    pub fn reset(&mut self) {
        self.value = 0;
    }
}
