// Sample Rust file used to demonstrate spectre mine --lang rust
// Corresponds to Listing 2 in the VMCAI 2027 paper

pub struct BankAccount {
    pub balance: i64,
    pub owner:   String,
    pub frozen:  bool,
}

impl BankAccount {
    pub fn new(owner: String) -> Self {
        Self { balance: 0, owner, frozen: false }
    }

    pub fn deposit(&mut self, amount: i64) {
        assert!(amount > 0);
        assert!(self.balance + amount <= 1_000_000);
        self.balance += amount;
    }

    pub fn withdraw(&mut self, amount: i64) {
        assert!(amount > 0);
        assert!(self.balance >= amount);
        assert!(!self.frozen);
        self.balance -= amount;
    }

    pub fn freeze(&mut self)   { assert!(!self.frozen); self.frozen = true;  }
    pub fn unfreeze(&mut self) { assert!(self.frozen);  self.frozen = false; }
}
