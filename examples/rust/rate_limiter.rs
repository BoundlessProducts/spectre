pub struct RateLimiter {
    pub tokens: i64,
    pub capacity: i64,
    pub refill_amount: i64,
}
impl RateLimiter {
    pub fn new(capacity: i64, refill_amount: i64) -> Self { Self { tokens: capacity, capacity, refill_amount } }
    pub fn consume(&mut self, amount: i64) { assert!(amount > 0); assert!(self.tokens >= amount); self.tokens -= amount; }
    pub fn refill(&mut self) { assert!(self.tokens + self.refill_amount <= self.capacity); self.tokens += self.refill_amount; }
    pub fn reset(&mut self) { self.tokens = self.capacity; }
}
