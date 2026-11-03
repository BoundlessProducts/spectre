// Resource pool — sequential compound assigns, panic patterns, Arc wrapping.
use std::sync::Arc;

pub struct ResourcePool {
    pub total: usize,
    pub available: usize,
    pub acquired: usize,
}

impl ResourcePool {
    pub fn new(total: usize) -> Self {
        ResourcePool { total, available: total, acquired: 0 }
    }

    pub fn acquire(&mut self, n: usize) {
        assert!(n <= self.available, "not enough resources");
        self.available -= n;
        self.acquired += n;
    }

    pub fn release(&mut self, n: usize) {
        assert!(n <= self.acquired, "releasing more than held");
        self.available += n;
        self.acquired -= n;
    }

    pub fn capacity(&self) -> usize {
        self.total
    }
}

pub fn wrap_pool(p: ResourcePool) -> Arc<ResourcePool> {
    Arc::new(p)
}
