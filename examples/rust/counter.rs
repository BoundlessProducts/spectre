pub struct Counter {
    pub count: i64,
    pub step: i64,
    pub max: i64,
}
impl Counter {
    pub fn new(step: i64, max: i64) -> Self { Self { count: 0, step, max } }
    pub fn increment(&mut self) { assert!(self.count + self.step <= self.max); self.count += self.step; }
    pub fn decrement(&mut self) { assert!(self.count >= self.step); self.count -= self.step; }
    pub fn reset(&mut self) { self.count = 0; }
}
