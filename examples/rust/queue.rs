pub struct Queue {
    pub length: i64,
    pub capacity: i64,
    pub head: i64,
    pub tail: i64,
}
impl Queue {
    pub fn new(capacity: i64) -> Self { Self { length: 0, capacity, head: 0, tail: 0 } }
    pub fn enqueue(&mut self, item: i64) { assert!(self.length < self.capacity); self.length += 1; self.tail = item; }
    pub fn dequeue(&mut self) -> i64 { assert!(self.length > 0); self.length -= 1; let v = self.head; self.head = self.tail; v }
    pub fn clear(&mut self) { self.length = 0; self.head = 0; self.tail = 0; }
}
