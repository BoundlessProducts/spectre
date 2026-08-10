pub struct EventLog {
    pub count: i64,
    pub capacity: i64,
    pub oldest_index: i64,
    pub locked: bool,
}
impl EventLog {
    pub fn new(capacity: i64) -> Self { Self { count: 0, capacity, oldest_index: 0, locked: false } }
    pub fn append(&mut self) { assert!(!self.locked); assert!(self.count < self.capacity); self.count += 1; }
    pub fn evict_oldest(&mut self) { assert!(self.count > 0); self.count -= 1; self.oldest_index += 1; }
    pub fn lock(&mut self) { assert!(!self.locked); self.locked = true; }
    pub fn unlock(&mut self) { assert!(self.locked); self.locked = false; }
    pub fn clear(&mut self) { assert!(!self.locked); self.count = 0; self.oldest_index = 0; }
}
