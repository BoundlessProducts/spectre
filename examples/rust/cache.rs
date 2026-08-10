pub struct Cache {
    pub size: i64,
    pub capacity: i64,
    pub hits: i64,
    pub misses: i64,
}
impl Cache {
    pub fn new(capacity: i64) -> Self { Self { size: 0, capacity, hits: 0, misses: 0 } }
    pub fn lookup_hit(&mut self) { self.hits += 1; }
    pub fn lookup_miss(&mut self) { self.misses += 1; if self.size < self.capacity { self.size += 1; } }
    pub fn evict(&mut self) { assert!(self.size > 0); self.size -= 1; }
    pub fn clear(&mut self) { self.size = 0; }
}
