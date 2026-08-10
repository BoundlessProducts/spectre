pub enum BreakerState { Closed, Open, HalfOpen }
pub struct CircuitBreaker {
    pub failure_count: i64,
    pub threshold: i64,
    pub open: bool,
    pub half_open: bool,
}
impl CircuitBreaker {
    pub fn new(threshold: i64) -> Self { Self { failure_count: 0, threshold, open: false, half_open: false } }
    pub fn record_failure(&mut self) { assert!(!self.open); self.failure_count += 1; if self.failure_count >= self.threshold { self.open = true; } }
    pub fn record_success(&mut self) { assert!(self.half_open); self.open = false; self.half_open = false; self.failure_count = 0; }
    pub fn attempt_reset(&mut self) { assert!(self.open); assert!(!self.half_open); self.half_open = true; }
    pub fn trip(&mut self) { assert!(!self.open); self.open = true; self.half_open = false; }
}
