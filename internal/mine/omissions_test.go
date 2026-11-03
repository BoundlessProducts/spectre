package mine

import (
	"strings"
	"testing"
)

// ── SupportedSubset ───────────────────────────────────────────────────────────

func TestSupportedSubset_NotEmpty(t *testing.T) {
	cats := SupportedSubset()
	if len(cats) == 0 {
		t.Fatal("SupportedSubset returned no categories")
	}
}

func TestSupportedSubset_HasFieldCategory(t *testing.T) {
	for _, c := range SupportedSubset() {
		if c.Category == "Struct fields" {
			return
		}
	}
	t.Error("expected 'Struct fields' category")
}

func TestSupportedSubset_HasEnumsCategory(t *testing.T) {
	for _, c := range SupportedSubset() {
		if c.Category == "Enums" {
			return
		}
	}
	t.Error("expected 'Enums' category")
}

func TestSupportedSubset_HasMethodsCategory(t *testing.T) {
	for _, c := range SupportedSubset() {
		if c.Category == "Methods / actions" {
			return
		}
	}
	t.Error("expected 'Methods / actions' category")
}

// ── OmissionKind.String() ─────────────────────────────────────────────────────

func TestOmissionKindString(t *testing.T) {
	cases := []struct {
		kind OmissionKind
		want string
	}{
		{OmissionUnsupportedType, "UNSUPPORTED_TYPE"},
		{OmissionSkippedMethod, "SKIPPED_METHOD"},
		{OmissionTupleVariant, "TUPLE_VARIANT"},
		{OmissionOverflowOp, "OVERFLOW_OP"},
		{OmissionPanicPattern, "PANIC_PATTERN"},
		{OmissionResultReturn, "RESULT_RETURN"},
		{OmissionComplexAssign, "COMPLEX_ASSIGN"},
		{OmissionUnsupportedPattern, "UNSUPPORTED_PATTERN"},
	}
	for _, tc := range cases {
		if got := tc.kind.String(); got != tc.want {
			t.Errorf("OmissionKind(%d).String() = %q, want %q", tc.kind, got, tc.want)
		}
	}
}

func TestOmissionString_WithContext(t *testing.T) {
	o := ExtractionOmission{Kind: OmissionPanicPattern, Context: "transfer", Detail: "unwrap dropped"}
	s := o.String()
	if !strings.Contains(s, "PANIC_PATTERN") || !strings.Contains(s, "transfer") || !strings.Contains(s, "unwrap dropped") {
		t.Errorf("unexpected String(): %q", s)
	}
}

func TestOmissionString_NoContext(t *testing.T) {
	o := ExtractionOmission{Kind: OmissionSkippedMethod, Detail: "new() skipped"}
	s := o.String()
	if strings.Contains(s, " in ") {
		t.Errorf("expected no 'in' clause when Context is empty: %q", s)
	}
}

// ── ScanOmissions — unsupported types ────────────────────────────────────────

const rustWithUnsupportedField = `
pub struct State {
    pub count: i32,
    pub handle: SomeComplexType,
}
`

func TestScanOmissions_UnsupportedType(t *testing.T) {
	spec := MineFromRust(rustWithUnsupportedField, "State", "state.rs")
	found := false
	for _, o := range spec.Omissions {
		if o.Kind == OmissionUnsupportedType && strings.Contains(o.Context, "handle") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected OmissionUnsupportedType for 'handle' field; omissions: %v", spec.Omissions)
	}
}

func TestScanOmissions_KnownIntType_NoOmission(t *testing.T) {
	src := `
pub struct Counter {
    pub value: i64,
    pub total: u32,
}
`
	spec := MineFromRust(src, "Counter", "counter.rs")
	for _, o := range spec.Omissions {
		if o.Kind == OmissionUnsupportedType {
			t.Errorf("unexpected OmissionUnsupportedType for primitive fields: %v", o)
		}
	}
}

// ── ScanOmissions — overflow ops (Item 20) ────────────────────────────────────

const rustWithOverflow = `
pub struct Account {
    pub balance: i64,
}
impl Account {
    pub fn deposit(&mut self, amount: i64) {
        self.balance = self.balance.checked_add(amount).unwrap_or(i64::MAX);
    }
    pub fn withdraw(&mut self, amount: i64) {
        self.balance = self.balance.saturating_sub(amount);
    }
}
`

func TestScanOmissions_CheckedAdd(t *testing.T) {
	spec := MineFromRust(rustWithOverflow, "Account", "account.rs")
	found := false
	for _, o := range spec.Omissions {
		if o.Kind == OmissionOverflowOp && o.Context == "deposit" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected OmissionOverflowOp for 'deposit'; omissions: %v", spec.Omissions)
	}
}

func TestScanOmissions_SaturatingSub(t *testing.T) {
	spec := MineFromRust(rustWithOverflow, "Account", "account.rs")
	found := false
	for _, o := range spec.Omissions {
		if o.Kind == OmissionOverflowOp && o.Context == "withdraw" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected OmissionOverflowOp for 'withdraw'; omissions: %v", spec.Omissions)
	}
}

func TestScanOmissions_WrappingAdd(t *testing.T) {
	src := `
pub struct Ring { pub pos: u8 }
impl Ring {
    pub fn advance(&mut self) {
        self.pos = self.pos.wrapping_add(1);
    }
}
`
	spec := MineFromRust(src, "Ring", "ring.rs")
	found := false
	for _, o := range spec.Omissions {
		if o.Kind == OmissionOverflowOp {
			found = true
		}
	}
	if !found {
		t.Errorf("expected OmissionOverflowOp for wrapping_add; omissions: %v", spec.Omissions)
	}
}

// ── ScanOmissions — panic patterns (Item 20) ─────────────────────────────────

const rustWithPanic = `
pub struct Vault {
    pub locked: bool,
}
impl Vault {
    pub fn unlock(&mut self, key: u32) {
        if key != 42 {
            panic!("wrong key");
        }
        self.locked = false;
    }
    pub fn must_be_open(&mut self) {
        let v = some_call().unwrap();
        self.locked = v;
    }
    pub fn open_or_die(&mut self, msg: &str) {
        let v = other_call().expect(msg);
        self.locked = v;
    }
}
`

func TestScanOmissions_PanicBang(t *testing.T) {
	spec := MineFromRust(rustWithPanic, "Vault", "vault.rs")
	found := false
	for _, o := range spec.Omissions {
		if o.Kind == OmissionPanicPattern && o.Context == "unlock" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected OmissionPanicPattern for 'unlock'; omissions: %v", spec.Omissions)
	}
}

func TestScanOmissions_Unwrap(t *testing.T) {
	spec := MineFromRust(rustWithPanic, "Vault", "vault.rs")
	found := false
	for _, o := range spec.Omissions {
		if o.Kind == OmissionPanicPattern && o.Context == "must_be_open" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected OmissionPanicPattern for 'must_be_open'; omissions: %v", spec.Omissions)
	}
}

func TestScanOmissions_Expect(t *testing.T) {
	spec := MineFromRust(rustWithPanic, "Vault", "vault.rs")
	found := false
	for _, o := range spec.Omissions {
		if o.Kind == OmissionPanicPattern && o.Context == "open_or_die" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected OmissionPanicPattern for 'open_or_die'; omissions: %v", spec.Omissions)
	}
}

// ── ScanOmissions — Result / ? operator (Item 20) ────────────────────────────

const rustWithResult = `
pub struct Ledger {
    pub balance: i64,
}
impl Ledger {
    pub fn debit(&mut self, amount: i64) -> Result<(), String> {
        if amount > self.balance {
            return Err("insufficient funds".to_string());
        }
        self.balance -= amount;
        Ok(())
    }
    pub fn credit(&mut self, amount: i64) -> Result<(), String> {
        self.balance = self.balance.checked_add(amount)?;
        Ok(())
    }
}
`

func TestScanOmissions_ResultReturn(t *testing.T) {
	spec := MineFromRust(rustWithResult, "Ledger", "ledger.rs")
	found := false
	for _, o := range spec.Omissions {
		if o.Kind == OmissionResultReturn && o.Context == "debit" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected OmissionResultReturn for 'debit'; omissions: %v", spec.Omissions)
	}
}

func TestScanOmissions_QuestionOperator(t *testing.T) {
	spec := MineFromRust(rustWithResult, "Ledger", "ledger.rs")
	found := false
	for _, o := range spec.Omissions {
		if o.Kind == OmissionResultReturn && o.Context == "credit" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected OmissionResultReturn for 'credit'; omissions: %v", spec.Omissions)
	}
}

// ── ScanOmissions — tuple enum variants ──────────────────────────────────────

const rustWithTupleVariant = `
pub enum Event {
    Tick,
    Deposit(i64),
    Withdraw(i64),
}
pub struct Machine { pub state: Event }
`

func TestScanOmissions_TupleVariant(t *testing.T) {
	spec := MineFromRust(rustWithTupleVariant, "Machine", "machine.rs")
	count := 0
	for _, o := range spec.Omissions {
		if o.Kind == OmissionTupleVariant {
			count++
		}
	}
	if count < 2 {
		t.Errorf("expected at least 2 OmissionTupleVariant (Deposit, Withdraw); omissions: %v", spec.Omissions)
	}
}

func TestScanOmissions_UnitVariant_NoOmission(t *testing.T) {
	src := `
pub enum Color { Red, Green, Blue }
pub struct Canvas { pub color: Color }
`
	spec := MineFromRust(src, "Canvas", "canvas.rs")
	for _, o := range spec.Omissions {
		if o.Kind == OmissionTupleVariant {
			t.Errorf("unexpected OmissionTupleVariant for unit-only enum: %v", o)
		}
	}
}

// ── ScanOmissions — skipped methods ──────────────────────────────────────────

const rustWithConstructors = `
pub struct Counter { pub n: i32 }
impl Counter {
    pub fn new() -> Self { Counter { n: 0 } }
    pub fn default() -> Self { Counter { n: 0 } }
    pub fn clone(&self) -> Self { Counter { n: self.n } }
    pub fn increment(&mut self) { self.n += 1; }
}
`

func TestScanOmissions_SkipsNew(t *testing.T) {
	spec := MineFromRust(rustWithConstructors, "Counter", "counter.rs")
	found := false
	for _, o := range spec.Omissions {
		if o.Kind == OmissionSkippedMethod && o.Context == "new" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected OmissionSkippedMethod for 'new'; omissions: %v", spec.Omissions)
	}
}

func TestScanOmissions_SkipsClone(t *testing.T) {
	spec := MineFromRust(rustWithConstructors, "Counter", "counter.rs")
	found := false
	for _, o := range spec.Omissions {
		if o.Kind == OmissionSkippedMethod && o.Context == "clone" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected OmissionSkippedMethod for 'clone'; omissions: %v", spec.Omissions)
	}
}

func TestScanOmissions_RegularMethodNotSkipped(t *testing.T) {
	spec := MineFromRust(rustWithConstructors, "Counter", "counter.rs")
	for _, o := range spec.Omissions {
		if o.Kind == OmissionSkippedMethod && o.Context == "increment" {
			t.Errorf("increment should not be reported as skipped: %v", o)
		}
	}
}

// ── Item 20: compound assign extraction ──────────────────────────────────────

const rustWithCompoundAssign = `
pub struct Score {
    pub points: i32,
    pub lives: i32,
}
impl Score {
    pub fn add_points(&mut self, n: i32) {
        self.points += n;
    }
    pub fn lose_life(&mut self) {
        self.lives -= 1;
    }
    pub fn double_score(&mut self) {
        self.points *= 2;
    }
}
`

func TestExtractAssign_PlusEquals(t *testing.T) {
	spec := MineFromRust(rustWithCompoundAssign, "Score", "score.rs")
	found := false
	for _, m := range spec.Methods {
		if m.Name == "add_points" {
			for _, a := range m.Assigns {
				if a.Field == "points" && a.Op == "+=" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("expected += assignment for 'points' in add_points")
	}
}

func TestExtractAssign_MinusEquals(t *testing.T) {
	spec := MineFromRust(rustWithCompoundAssign, "Score", "score.rs")
	found := false
	for _, m := range spec.Methods {
		if m.Name == "lose_life" {
			for _, a := range m.Assigns {
				if a.Field == "lives" && a.Op == "-=" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("expected -= assignment for 'lives' in lose_life")
	}
}

func TestExtractAssign_StarEquals(t *testing.T) {
	spec := MineFromRust(rustWithCompoundAssign, "Score", "score.rs")
	found := false
	for _, m := range spec.Methods {
		if m.Name == "double_score" {
			for _, a := range m.Assigns {
				if a.Field == "points" && a.Op == "*=" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("expected *= assignment for 'points' in double_score")
	}
}

func TestRustAssignToSpectre_PlusEquals(t *testing.T) {
	result := rustAssignToSpectre("count", "+=", "n")
	if result != "count + n" {
		t.Errorf("expected 'count + n', got %q", result)
	}
}

func TestRustAssignToSpectre_MinusEquals(t *testing.T) {
	result := rustAssignToSpectre("lives", "-=", "1")
	if result != "lives - 1" {
		t.Errorf("expected 'lives - 1', got %q", result)
	}
}

func TestRustAssignToSpectre_StarEquals(t *testing.T) {
	result := rustAssignToSpectre("score", "*=", "2")
	if result != "score * 2" {
		t.Errorf("expected 'score * 2', got %q", result)
	}
}
