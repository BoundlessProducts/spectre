pub struct Stack {
    pub size: i64,
    pub capacity: i64,
    pub top: i64,
}
impl Stack {
    pub fn new(capacity: i64) -> Self { Self { size: 0, capacity, top: -1 } }
    pub fn push(&mut self, value: i64) { assert!(self.size < self.capacity); self.size += 1; self.top = value; }
    pub fn pop(&mut self) -> i64 { assert!(self.size > 0); self.size -= 1; let v = self.top; self.top = -1; v }
    pub fn clear(&mut self) { self.size = 0; self.top = -1; }
}
