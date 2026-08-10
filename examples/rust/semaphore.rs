pub struct Semaphore {
    pub permits: i64,
    pub max_permits: i64,
    pub waiters: i64,
}
impl Semaphore {
    pub fn new(max_permits: i64) -> Self { Self { permits: max_permits, max_permits, waiters: 0 } }
    pub fn acquire(&mut self) { assert!(self.permits > 0); self.permits -= 1; }
    pub fn release(&mut self) { assert!(self.permits < self.max_permits); self.permits += 1; }
    pub fn add_waiter(&mut self) { assert!(self.permits == 0); self.waiters += 1; }
    pub fn notify_waiter(&mut self) { assert!(self.waiters > 0); self.waiters -= 1; self.permits += 1; }
}
