package mine

import (
	"path/filepath"
	"strings"
	"testing"
)

// ── EvaluateFile ──────────────────────────────────────────────────────────────

func TestEvaluateFile_Counter_NoBodyOmissions(t *testing.T) {
	r, err := EvaluateFile(testdataPath("counter.rs"))
	if err != nil {
		t.Fatal(err)
	}
	// counter.rs may have SKIPPED_METHOD for new(); body-level (overflow/panic/result) should be zero.
	if r.OmissionsByKind[OmissionOverflowOp]+r.OmissionsByKind[OmissionPanicPattern]+
		r.OmissionsByKind[OmissionResultReturn]+r.OmissionsByKind[OmissionComplexAssign] != 0 {
		t.Errorf("counter.rs: unexpected body-level omissions: %v", r.OmissionsByKind)
	}
}

func TestEvaluateFile_Counter_FieldsExtracted(t *testing.T) {
	r, err := EvaluateFile(testdataPath("counter.rs"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Fields < 2 {
		t.Errorf("counter.rs: want >=2 fields, got %d", r.Fields)
	}
}

func TestEvaluateFile_Counter_MethodsExtracted(t *testing.T) {
	r, err := EvaluateFile(testdataPath("counter.rs"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Methods == 0 {
		t.Errorf("counter.rs: want >0 methods, got 0")
	}
}

func TestEvaluateFile_Counter_FullSuccess(t *testing.T) {
	r, err := EvaluateFile(testdataPath("counter.rs"))
	if err != nil {
		t.Fatal(err)
	}
	if r.SuccessFraction != 1.0 {
		t.Errorf("counter.rs: want 100%% success, got %.2f", r.SuccessFraction)
	}
}

func TestEvaluateFile_BankAccount_HasOmissions(t *testing.T) {
	r, err := EvaluateFile(testdataPath("bank_account.rs"))
	if err != nil {
		t.Fatal(err)
	}
	if r.OmissionCount == 0 {
		t.Error("bank_account.rs: expected omissions (Result/checked_add), got none")
	}
}

func TestEvaluateFile_BankAccount_HasEnum(t *testing.T) {
	r, err := EvaluateFile(testdataPath("bank_account.rs"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Enums == 0 {
		t.Error("bank_account.rs: expected >=1 enum (TxError), got 0")
	}
}

func TestEvaluateFile_BankAccount_SuccessBelow100(t *testing.T) {
	r, err := EvaluateFile(testdataPath("bank_account.rs"))
	if err != nil {
		t.Fatal(err)
	}
	if r.SuccessFraction >= 1.0 {
		t.Errorf("bank_account.rs: expected success < 100%%, got %.2f", r.SuccessFraction)
	}
}

func TestEvaluateFile_StateMachine_TupleVariantOmission(t *testing.T) {
	r, err := EvaluateFile(testdataPath("state_machine.rs"))
	if err != nil {
		t.Fatal(err)
	}
	if r.OmissionsByKind[OmissionTupleVariant] == 0 {
		t.Error("state_machine.rs: expected TUPLE_VARIANT omission, got none")
	}
}

func TestEvaluateFile_StateMachine_OverflowOmission(t *testing.T) {
	r, err := EvaluateFile(testdataPath("state_machine.rs"))
	if err != nil {
		t.Fatal(err)
	}
	if r.OmissionsByKind[OmissionOverflowOp] == 0 {
		t.Error("state_machine.rs: expected OVERFLOW_OP omission, got none")
	}
}

func TestEvaluateFile_NonExistent_ReturnsError(t *testing.T) {
	_, err := EvaluateFile("/nonexistent/path/foo.rs")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestEvaluateFile_PathSetOnReport(t *testing.T) {
	path := testdataPath("counter.rs")
	r, err := EvaluateFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if r.Path != path {
		t.Errorf("Path: want %q, got %q", path, r.Path)
	}
}

// ── EvaluateFiles ─────────────────────────────────────────────────────────────

func TestEvaluateFiles_TotalFiles(t *testing.T) {
	paths := []string{
		testdataPath("counter.rs"),
		testdataPath("bank_account.rs"),
	}
	s, err := EvaluateFiles(paths)
	if err != nil {
		t.Fatal(err)
	}
	if s.TotalFiles != 2 {
		t.Errorf("TotalFiles: want 2, got %d", s.TotalFiles)
	}
}

func TestEvaluateFiles_TotalOmissionsAggregated(t *testing.T) {
	paths := []string{
		testdataPath("counter.rs"),
		testdataPath("bank_account.rs"),
		testdataPath("state_machine.rs"),
	}
	s, err := EvaluateFiles(paths)
	if err != nil {
		t.Fatal(err)
	}
	if s.TotalOmissions == 0 {
		t.Error("TotalOmissions: expected >0 when bank_account.rs and state_machine.rs included")
	}
}

func TestEvaluateFiles_MeanSuccessInRange(t *testing.T) {
	paths := []string{
		testdataPath("counter.rs"),
		testdataPath("bank_account.rs"),
	}
	s, err := EvaluateFiles(paths)
	if err != nil {
		t.Fatal(err)
	}
	if s.MeanSuccess < 0 || s.MeanSuccess > 1 {
		t.Errorf("MeanSuccess out of range [0,1]: %f", s.MeanSuccess)
	}
}

func TestEvaluateFiles_ErrorOnBadPath(t *testing.T) {
	_, err := EvaluateFiles([]string{"/no/such/file.rs"})
	if err == nil {
		t.Error("expected error for bad path, got nil")
	}
}

func TestEvaluateFiles_OmissionsByKindAggregated(t *testing.T) {
	paths := []string{
		testdataPath("bank_account.rs"),
		testdataPath("state_machine.rs"),
	}
	s, err := EvaluateFiles(paths)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.OmissionsByKind) == 0 {
		t.Error("OmissionsByKind: expected non-empty map")
	}
}

// ── FormatReport ──────────────────────────────────────────────────────────────

func TestFormatReport_ContainsHeader(t *testing.T) {
	paths := []string{testdataPath("counter.rs")}
	s, err := EvaluateFiles(paths)
	if err != nil {
		t.Fatal(err)
	}
	out := FormatReport(s)
	if !strings.Contains(out, "Rust Extraction Evaluation Report") {
		t.Error("FormatReport missing header")
	}
}

func TestFormatReport_ContainsFilename(t *testing.T) {
	paths := []string{testdataPath("counter.rs")}
	s, err := EvaluateFiles(paths)
	if err != nil {
		t.Fatal(err)
	}
	out := FormatReport(s)
	if !strings.Contains(out, "counter.rs") {
		t.Error("FormatReport missing filename")
	}
}

func TestFormatReport_ContainsTotals(t *testing.T) {
	paths := []string{testdataPath("counter.rs"), testdataPath("bank_account.rs")}
	s, err := EvaluateFiles(paths)
	if err != nil {
		t.Fatal(err)
	}
	out := FormatReport(s)
	if !strings.Contains(out, "Totals") {
		t.Error("FormatReport missing Totals line")
	}
}

func TestFormatReport_OmissionsSectionWhenPresent(t *testing.T) {
	paths := []string{testdataPath("bank_account.rs")}
	s, err := EvaluateFiles(paths)
	if err != nil {
		t.Fatal(err)
	}
	out := FormatReport(s)
	if !strings.Contains(out, "Omissions by Kind") {
		t.Error("FormatReport missing 'Omissions by Kind' section")
	}
}

func TestFormatReport_ContainsMarkdownTableRows(t *testing.T) {
	paths := []string{testdataPath("counter.rs"), testdataPath("bank_account.rs")}
	s, err := EvaluateFiles(paths)
	if err != nil {
		t.Fatal(err)
	}
	out := FormatReport(s)
	// Markdown table must have at least the header separator row.
	if !strings.Contains(out, "|---") {
		t.Error("FormatReport missing Markdown table separator")
	}
}

// testdataPath resolves a filename relative to testdata/rust_eval/.
func testdataPath(name string) string {
	return filepath.Join("..", "..", "testdata", "rust_eval", name)
}
