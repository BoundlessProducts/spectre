pub struct ConnectionPool {
    pub available: i64,
    pub in_use: i64,
    pub max_size: i64,
}
impl ConnectionPool {
    pub fn new(max_size: i64) -> Self { Self { available: max_size, in_use: 0, max_size } }
    pub fn acquire(&mut self) { assert!(self.available > 0); self.available -= 1; self.in_use += 1; }
    pub fn release(&mut self) { assert!(self.in_use > 0); self.in_use -= 1; self.available += 1; }
    pub fn drain(&mut self) { assert!(self.in_use == 0); self.available = 0; }
    pub fn replenish(&mut self) { assert!(self.available == 0); self.available = self.max_size; }
}
