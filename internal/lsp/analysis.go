package lsp

// analysis.go — stateless Spectre-specific analysis helpers.
// These functions operate on a parsed AST and produce LSP-typed results.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// ParseResult bundles the parse outcome for a document.
type ParseResult struct {
	File   *ast.File
	Errors []Diagnostic
}

// ParseDocument parses content and returns diagnostics for any parse errors.
// uri is used as the diagnostic source label.
func ParseDocument(uri, content string) *ParseResult {
	l := lexer.New(content)
	p := parser.New(l)
	file := p.ParseFile()

	diags := make([]Diagnostic, 0, len(p.Errors()))
	for _, errMsg := range p.Errors() {
		diags = append(diags, parseErrorToDiagnostic(errMsg))
	}
	return &ParseResult{File: file, Errors: diags}
}

// parseErrorToDiagnostic converts a "line:col: message" string to an LSP Diagnostic.
func parseErrorToDiagnostic(msg string) Diagnostic {
	line, col := 0, 0
	text := msg
	// Try to parse "line:col: message" format emitted by the Spectre parser.
	var l, c int
	if n, _ := fmt.Sscanf(msg, "%d:%d:", &l, &c); n == 2 {
		line = l - 1 // LSP is 0-based
		col = c - 1
		if idx := strings.Index(msg, ": "); idx >= 0 {
			text = msg[idx+2:]
		}
	}
	r := Range{
		Start: Position{Line: line, Character: col},
		End:   Position{Line: line, Character: col + 1},
	}
	return Diagnostic{
		Range:    r,
		Severity: SeverityError,
		Source:   "spectre",
		Message:  text,
	}
}

// CompletionItems returns all named identifiers declared in file as completion items.
func CompletionItems(file *ast.File) []CompletionItem {
	var items []CompletionItem
	seen := map[string]bool{}

	add := func(label string, kind int, detail string) {
		if seen[label] {
			return
		}
		seen[label] = true
		items = append(items, CompletionItem{Label: label, Kind: kind, Detail: detail})
	}

	for _, d := range file.Decls {
		switch v := d.(type) {
		case *ast.VariableDecl:
			add(v.Name, 6 /*Variable*/, fmt.Sprintf("var %s: %s", v.Name, typeString(v.Type)))
		case *ast.ConstantDecl:
			add(v.Name, 21 /*Constant*/, fmt.Sprintf("const %s", v.Name))
		case *ast.FunctionDecl:
			add(v.Name, 3 /*Function*/, fmt.Sprintf("fun %s", v.Name))
		case *ast.ActionDecl:
			add(v.Name, 3 /*Function*/, fmt.Sprintf("action %s", v.Name))
		case *ast.InvariantDecl:
			add(v.Name, 5 /*Field*/, fmt.Sprintf("invariant %s", v.Name))
		case *ast.EnumDecl:
			add(v.Name, 13 /*Enum*/, fmt.Sprintf("enum %s", v.Name))
			for _, val := range v.Values {
				add(v.Name+"."+val, 12 /*EnumMember*/, fmt.Sprintf("%s.%s", v.Name, val))
			}
		case *ast.ParamDecl:
			add(v.Name, 6 /*Variable*/, fmt.Sprintf("param %s: %s", v.Name, typeString(v.Type)))
		case *ast.ModuleDecl:
			add(v.Name, 9 /*Module*/, fmt.Sprintf("module %s", v.Name))
			for _, md := range v.Decls {
				if fn, ok := md.(*ast.FunctionDecl); ok {
					add(v.Name+"."+fn.Name, 3 /*Function*/, fmt.Sprintf("module function %s.%s", v.Name, fn.Name))
				}
			}
		}
	}

	sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })
	return items
}

// HoverInfo returns hover markdown for the identifier closest to pos in content.
// Returns nil if no hover information is available at that position.
func HoverInfo(file *ast.File, content string, pos Position) *Hover {
	word := wordAtPosition(content, pos)
	if word == "" {
		return nil
	}
	for _, d := range file.Decls {
		if h := hoverForDecl(d, word); h != nil {
			return h
		}
	}
	return nil
}

func hoverForDecl(d ast.Decl, word string) *Hover {
	switch v := d.(type) {
	case *ast.VariableDecl:
		if v.Name == word {
			return hoverMarkdown(fmt.Sprintf("**var** `%s`: `%s`", v.Name, typeString(v.Type)))
		}
	case *ast.ConstantDecl:
		if v.Name == word {
			return hoverMarkdown(fmt.Sprintf("**const** `%s`", v.Name))
		}
	case *ast.FunctionDecl:
		if v.Name == word {
			return hoverMarkdown(funSignature(v))
		}
	case *ast.ActionDecl:
		if v.Name == word {
			return hoverMarkdown(actionSignature(v))
		}
	case *ast.InvariantDecl:
		if v.Name == word {
			desc := ""
			if v.Description != "" {
				desc = "\n\n" + v.Description
			}
			return hoverMarkdown(fmt.Sprintf("**invariant** `%s`%s", v.Name, desc))
		}
	case *ast.EnumDecl:
		if v.Name == word {
			return hoverMarkdown(fmt.Sprintf("**enum** `%s` { %s }", v.Name, strings.Join(v.Values, ", ")))
		}
		for _, val := range v.Values {
			if v.Name+"."+val == word || val == word {
				return hoverMarkdown(fmt.Sprintf("`%s.%s` — enum value of `%s`", v.Name, val, v.Name))
			}
		}
	case *ast.ParamDecl:
		if v.Name == word {
			return hoverMarkdown(fmt.Sprintf("**param** `%s`: `%s`", v.Name, typeString(v.Type)))
		}
	case *ast.ModuleDecl:
		if v.Name == word {
			return hoverMarkdown(fmt.Sprintf("**module** `%s`", v.Name))
		}
		for _, md := range v.Decls {
			if fn, ok := md.(*ast.FunctionDecl); ok {
				if v.Name+"."+fn.Name == word || fn.Name == word {
					return hoverMarkdown(fmt.Sprintf("**%s.%s** — %s", v.Name, fn.Name, funSignature(fn)))
				}
			}
		}
	}
	return nil
}

// DefinitionLocation returns the source location of the declaration for word.
// Returns nil if no definition is found.
func DefinitionLocation(file *ast.File, uri, content string, pos Position) *Location {
	word := wordAtPosition(content, pos)
	if word == "" {
		return nil
	}
	for _, d := range file.Decls {
		if loc := locationForDecl(d, word, uri); loc != nil {
			return loc
		}
	}
	return nil
}

func locationForDecl(d ast.Decl, word, uri string) *Location {
	checkPos := func(name string, p ast.Position) *Location {
		if name == word && p.Line > 0 {
			r := Range{
				Start: Position{Line: p.Line - 1, Character: p.Column - 1},
				End:   Position{Line: p.Line - 1, Character: p.Column - 1 + len(name)},
			}
			return &Location{URI: uri, Range: r}
		}
		return nil
	}

	switch v := d.(type) {
	case *ast.VariableDecl:
		return checkPos(v.Name, v.Position)
	case *ast.ConstantDecl:
		return checkPos(v.Name, v.Position)
	case *ast.FunctionDecl:
		return checkPos(v.Name, v.Position)
	case *ast.ActionDecl:
		return checkPos(v.Name, v.Position)
	case *ast.InvariantDecl:
		return checkPos(v.Name, v.Position)
	case *ast.EnumDecl:
		if loc := checkPos(v.Name, v.Position); loc != nil {
			return loc
		}
		for _, val := range v.Values {
			if val == word || v.Name+"."+val == word {
				return &Location{URI: uri, Range: Range{
					Start: Position{Line: v.Position.Line - 1, Character: v.Position.Column - 1},
					End:   Position{Line: v.Position.Line - 1, Character: v.Position.Column - 1 + len(v.Name)},
				}}
			}
		}
	case *ast.ParamDecl:
		return checkPos(v.Name, v.Position)
	case *ast.ModuleDecl:
		if loc := checkPos(v.Name, v.Position); loc != nil {
			return loc
		}
		for _, md := range v.Decls {
			if fn, ok := md.(*ast.FunctionDecl); ok {
				if loc := checkPos(fn.Name, fn.Position); loc != nil {
					return loc
				}
			}
		}
	}
	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func wordAtPosition(content string, pos Position) string {
	lines := strings.Split(content, "\n")
	if pos.Line < 0 || pos.Line >= len(lines) {
		return ""
	}
	line := lines[pos.Line]
	ch := pos.Character
	if ch < 0 || ch > len(line) {
		return ""
	}
	// Walk left and right to find word boundaries.
	start := ch
	for start > 0 && isIdentChar(line[start-1]) {
		start--
	}
	end := ch
	for end < len(line) && isIdentChar(line[end]) {
		end++
	}
	return line[start:end]
}

func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

func hoverMarkdown(s string) *Hover {
	return &Hover{Contents: MarkupContent{Kind: "markdown", Value: s}}
}

func typeString(t ast.Type) string {
	if t == nil {
		return "unknown"
	}
	switch v := t.(type) {
	case *ast.PrimitiveType:
		return v.Name
	case *ast.SetType:
		return fmt.Sprintf("Set[%s]", typeString(v.Element))
	case *ast.ListType:
		return fmt.Sprintf("List[%s]", typeString(v.Element))
	case *ast.MapType:
		return fmt.Sprintf("Map[%s, %s]", typeString(v.Key), typeString(v.Value))
	case *ast.NamedType:
		return v.Name
	case *ast.EnumType:
		return v.Name
	case *ast.OptionType:
		return fmt.Sprintf("Option[%s]", typeString(v.Element))
	case *ast.RecordType:
		fields := make([]string, len(v.Fields))
		for i, f := range v.Fields {
			fields[i] = f.Name + ": " + typeString(f.Type)
		}
		return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
	default:
		return "unknown"
	}
}

func funSignature(f *ast.FunctionDecl) string {
	params := make([]string, len(f.Parameters))
	for i, p := range f.Parameters {
		params[i] = p.Name + ": " + typeString(p.Type)
	}
	ret := ""
	if f.ReturnType != nil {
		ret = ": " + typeString(f.ReturnType)
	}
	return fmt.Sprintf("**fun** `%s`(%s)%s", f.Name, strings.Join(params, ", "), ret)
}

func actionSignature(a *ast.ActionDecl) string {
	params := make([]string, len(a.Parameters))
	for i, p := range a.Parameters {
		params[i] = p.Name + ": " + typeString(p.Type)
	}
	return fmt.Sprintf("**action** `%s`(%s)", a.Name, strings.Join(params, ", "))
}
