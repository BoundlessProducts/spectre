//go:build ignore

package main

import (
	"fmt"
	"github.com/akkeshavan/spectre/internal/mine"
)

const src = `
pub struct BankAccount {
    pub balance: i64,
    owner: &str,
}
impl BankAccount {
    pub fn transfer(&mut self, amount: i64, target: &mut i64) {
        assert!(amount > 0);
        self.balance -= amount;
    }
}
`

func main() {
	ms := mine.MineFromRust(src, "BankAccount", "bank.rs")
	for _, m := range ms.Methods {
		fmt.Printf("method %s params:", m.Name)
		for _, p := range m.Params {
			fmt.Printf(" %s:%q", p.Name, p.SpectreType)
		}
		fmt.Println()
	}
	fmt.Println(ms.Generate(nil))
}
