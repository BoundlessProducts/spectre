package mine

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// FileReport summarises extraction quality for one Rust source file.
type FileReport struct {
	Path          string
	StructName    string
	Fields        int
	Enums         int
	Methods       int
	OmissionCount int
	OmissionsByKind map[OmissionKind]int
	SuccessFraction float64 // fraction of methods with no body omissions
}

// EvalSummary aggregates stats across all evaluated files.
type EvalSummary struct {
	Files         []*FileReport
	TotalFiles    int
	TotalFields   int
	TotalEnums    int
	TotalMethods  int
	TotalOmissions int
	OmissionsByKind map[OmissionKind]int
	MeanSuccess   float64
}

// EvaluateFile reads, mines, and evaluates a single Rust source file.
func EvaluateFile(path string) (*FileReport, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	spec := MineFromRust(string(src), "", path)
	return reportFromSpec(path, spec), nil
}

// EvaluateFiles evaluates multiple Rust source files and returns an aggregate.
func EvaluateFiles(paths []string) (*EvalSummary, error) {
	summary := &EvalSummary{
		OmissionsByKind: make(map[OmissionKind]int),
	}
	for _, p := range paths {
		r, err := EvaluateFile(p)
		if err != nil {
			return nil, err
		}
		summary.Files = append(summary.Files, r)
		summary.TotalFiles++
		summary.TotalFields += r.Fields
		summary.TotalEnums += r.Enums
		summary.TotalMethods += r.Methods
		summary.TotalOmissions += r.OmissionCount
		for k, n := range r.OmissionsByKind {
			summary.OmissionsByKind[k] += n
		}
	}
	if summary.TotalFiles > 0 {
		var total float64
		for _, r := range summary.Files {
			total += r.SuccessFraction
		}
		summary.MeanSuccess = total / float64(summary.TotalFiles)
	}
	return summary, nil
}

func reportFromSpec(path string, spec *MinedSpec) *FileReport {
	r := &FileReport{
		Path:            path,
		StructName:      spec.StructName,
		Fields:          len(spec.Fields),
		Enums:           len(spec.Enums),
		Methods:         len(spec.Methods),
		OmissionCount:   len(spec.Omissions),
		OmissionsByKind: make(map[OmissionKind]int),
	}
	for _, o := range spec.Omissions {
		r.OmissionsByKind[o.Kind]++
	}

	// Success fraction: methods that have no body-level omission.
	// Body-level omissions are those tied to overflow, panic, result, or complex assigns.
	bodyOmissionMethods := methodsWithBodyOmission(spec)
	if r.Methods == 0 {
		r.SuccessFraction = 1.0
	} else {
		r.SuccessFraction = float64(r.Methods-bodyOmissionMethods) / float64(r.Methods)
	}
	return r
}

// methodsWithBodyOmission counts distinct methods that have at least one
// body-level omission (overflow, panic, result, complex assign).
func methodsWithBodyOmission(spec *MinedSpec) int {
	seen := map[string]bool{}
	for _, o := range spec.Omissions {
		switch o.Kind {
		case OmissionOverflowOp, OmissionPanicPattern, OmissionResultReturn, OmissionComplexAssign:
			seen[o.Context] = true
		}
	}
	return len(seen)
}

// FormatReport renders an EvalSummary as a Markdown table.
func FormatReport(summary *EvalSummary) string {
	var sb strings.Builder

	sb.WriteString("# Rust Extraction Evaluation Report\n\n")
	sb.WriteString("## Per-File Summary\n\n")
	sb.WriteString("| File | Struct | Fields | Enums | Methods | Omissions | SuccessFraction |\n")
	sb.WriteString("|------|--------|--------|-------|---------|-----------|----------------|\n")
	for _, r := range summary.Files {
		name := r.Path
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %d | %d | %d | %.0f%% |\n",
			name, r.StructName, r.Fields, r.Enums, r.Methods, r.OmissionCount,
			r.SuccessFraction*100))
	}

	sb.WriteString(fmt.Sprintf("\n**Totals** — files: %d | fields: %d | enums: %d | methods: %d | omissions: %d | mean success: %.0f%%\n\n",
		summary.TotalFiles, summary.TotalFields, summary.TotalEnums,
		summary.TotalMethods, summary.TotalOmissions, summary.MeanSuccess*100))

	if len(summary.OmissionsByKind) > 0 {
		sb.WriteString("## Omissions by Kind\n\n")
		sb.WriteString("| Kind | Count |\n")
		sb.WriteString("|------|-------|\n")

		// Sort by kind value for stable output.
		kinds := make([]int, 0, len(summary.OmissionsByKind))
		for k := range summary.OmissionsByKind {
			kinds = append(kinds, int(k))
		}
		sort.Ints(kinds)
		for _, ki := range kinds {
			k := OmissionKind(ki)
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", k.String(), summary.OmissionsByKind[k]))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
