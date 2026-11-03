// Bank account — uses overflow-checked arithmetic and Result returns.
pub struct BankAccount {
    pub balance: i64,
    pub owner: String,
    pub frozen: bool,
}

pub enum TxError {
    InsufficientFunds,
    AccountFrozen,
    Overflow,
}

impl BankAccount {
    pub fn new(owner: String) -> Self {
        BankAccount { balance: 0, owner, frozen: false }
    }

    pub fn deposit(&mut self, amount: i64) -> Result<(), TxError> {
        if self.frozen {
            return Err(TxError::AccountFrozen);
        }
        self.balance = self.balance.checked_add(amount)
            .ok_or(TxError::Overflow)?;
        Ok(())
    }

    pub fn withdraw(&mut self, amount: i64) -> Result<(), TxError> {
        if self.frozen {
            return Err(TxError::AccountFrozen);
        }
        if self.balance < amount {
            return Err(TxError::InsufficientFunds);
        }
        self.balance -= amount;
        Ok(())
    }

    pub fn freeze(&mut self) {
        self.frozen = true;
    }

    pub fn unfreeze(&mut self) {
        self.frozen = false;
    }
}
