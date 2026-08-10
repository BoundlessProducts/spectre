package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/akkeshavan/spectre/internal/codegen"
	"github.com/akkeshavan/spectre/internal/diagnose"
	"github.com/akkeshavan/spectre/internal/mine"
	"github.com/akkeshavan/spectre/internal/errors"
	"github.com/akkeshavan/spectre/internal/exec"
	"github.com/akkeshavan/spectre/internal/explore"
	"github.com/akkeshavan/spectre/internal/itf"
	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/internal/parser"
	"github.com/akkeshavan/spectre/internal/semantic"
	"github.com/akkeshavan/spectre/internal/state"
	"github.com/akkeshavan/spectre/internal/types"
	"github.com/akkeshavan/spectre/pkg/ast"
)

// parseInt parses a string to int
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Run         func(args []string) error
}

// Commands maps command names to their implementations
var Commands = map[string]*Command{
	"parse": {
		Name:        "parse",
		Description: "Parse a Spectre specification file and report syntax errors",
		Run:         runParse,
	},
	"typecheck": {
		Name:        "typecheck",
		Description: "Type-check a Spectre specification file and report type errors",
		Run:         runTypecheck,
	},
	"verify": {
		Name:        "verify",
		Description: "Verify a Spectre specification (check invariants and temporal properties)",
		Run:         runVerify,
	},
	"generate-driver": {
		Name:        "generate-driver",
		Description: "Generate a Rust driver skeleton for model-based testing via spectre-connect",
		Run:         runGenerateDriver,
	},
	"generate-monitor": {
		Name:        "generate-monitor",
		Description: "Generate an embedded Rust runtime monitor that checks spec invariants in production",
		Run:         runGenerateMonitor,
	},
	"mine": {
		Name:        "mine",
		Description: "Mine a Spectre spec skeleton from Rust source code",
		Run:         runMine,
	},
	"check-refinement": {
		Name:        "check-refinement",
		Description: "Check that ITF traces conform to the spec via the driver.json refinement mapping",
		Run:         runCheckRefinement,
	},
}

// runParse parses a Spectre file and reports syntax errors
func runParse(args []string) error {
	// Filter out flags
	filteredArgs := []string{}
	for _, arg := range args {
		if arg != "--verbose" && arg != "-v" {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	
	if len(filteredArgs) == 0 {
		return fmt.Errorf("no file specified")
	}

	processor := NewFileProcessor()
	return processor.ProcessFiles(filteredArgs, func(filename string) error {
		// First, do a quick parse to check if the file has imports
		content, err := os.ReadFile(filename)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Error: File not found - %s\n", filename)
				return fmt.Errorf("File not found - %s", filename)
			}
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filename, err)
			return fmt.Errorf("error reading file %s: %w", filename, err)
		}

		l := lexer.New(string(content))
		p := parser.New(l)
		file := p.ParseFile()

		if len(p.Errors()) > 0 {
			fmt.Fprintf(os.Stderr, "Parse errors in %s:\n", filename)
			for _, errMsg := range p.Errors() {
				fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
			}
			return fmt.Errorf("parse failed with %d error(s)", len(p.Errors()))
		}

		// Check if file has any imports
		hasImports := false
		for _, decl := range file.Decls {
			if _, ok := decl.(*ast.ImportDecl); ok {
				hasImports = true
				break
			}
		}

		// If file has imports, use module loader (which requires modules)
		// If file has no imports, use the old parser (which doesn't require modules)
		if hasImports {
			fileDir := filepath.Dir(filename)
			
			// Use module loader to handle file-based imports
			loader := semantic.NewModuleLoader(fileDir)
			_, loadErrors := loader.LoadModule(filename)

			if len(loadErrors) > 0 {
				fmt.Fprintf(os.Stderr, "Parse/module errors in %s:\n", filename)
				for _, errMsg := range loadErrors {
					fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
				}
				
				// Check for circular dependencies
				circularErrors := loader.CheckCircularDependencies()
				if len(circularErrors) > 0 {
					for _, errMsg := range circularErrors {
						fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
					}
				}
				
				return fmt.Errorf("parse/module loading failed with %d error(s)", len(loadErrors)+len(circularErrors))
			}

			// Check for circular dependencies
			circularErrors := loader.CheckCircularDependencies()
			if len(circularErrors) > 0 {
				fmt.Fprintf(os.Stderr, "Circular dependency errors:\n")
				for _, errMsg := range circularErrors {
					fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
				}
				return fmt.Errorf("circular dependencies detected")
			}
		}

		fmt.Printf("✓ Successfully parsed %s\n", filename)
		return nil
	})
}

// runTypecheck type-checks a Spectre file and reports type errors
func runTypecheck(args []string) error {
	// Filter out flags
	filteredArgs := []string{}
	for _, arg := range args {
		if arg != "--verbose" && arg != "-v" {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	
	if len(filteredArgs) == 0 {
		return fmt.Errorf("no file specified")
	}

	processor := NewFileProcessor()
	return processor.ProcessFiles(filteredArgs, func(filename string) error {
	// First, do a quick parse to check if the file has imports
	content, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: File not found - %s\n", filename)
			return fmt.Errorf("File not found - %s", filename)
		}
		return fmt.Errorf("error reading file %s: %w", filename, err)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "Parse errors in %s:\n", filename)
		for _, errMsg := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
		}
		return fmt.Errorf("parse failed with %d error(s)", len(p.Errors()))
	}

	// Check if file has any imports
	hasImports := false
	for _, decl := range file.Decls {
		if _, ok := decl.(*ast.ImportDecl); ok {
			hasImports = true
			break
		}
	}

	// Declare loader outside the if block so it's accessible later
	var loader *semantic.ModuleLoader

	// If file has imports, use module loader (which requires modules)
	// If file has no imports, use the parsed file directly
	if hasImports {
		fileDir := filepath.Dir(filename)
		
		// Use module loader to handle file-based imports
		loader = semantic.NewModuleLoader(fileDir)
		moduleInfo, loadErrors := loader.LoadModule(filename)

		if len(loadErrors) > 0 {
			fmt.Fprintf(os.Stderr, "Parse/module errors in %s:\n", filename)
			for _, errMsg := range loadErrors {
				fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
			}
			
			// Check for circular dependencies
			circularErrors := loader.CheckCircularDependencies()
			if len(circularErrors) > 0 {
				for _, errMsg := range circularErrors {
					fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
				}
			}
			
			return fmt.Errorf("parse/module loading failed with %d error(s)", len(loadErrors)+len(circularErrors))
		}

		// Check for circular dependencies
		circularErrors := loader.CheckCircularDependencies()
		if len(circularErrors) > 0 {
			fmt.Fprintf(os.Stderr, "Circular dependency errors:\n")
			for _, errMsg := range circularErrors {
				fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
			}
			return fmt.Errorf("circular dependencies detected")
		}

		// Use the loaded file from module loader
		file = moduleInfo.File
	}

	// Build symbol table
	builder := semantic.NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Symbol table build errors in %s:\n", filename)
		for _, errMsg := range buildErrors {
			fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
		}
		return fmt.Errorf("symbol table build failed with %d error(s)", len(buildErrors))
	}

	// Only use old resolver if no imports at all (for backward compatibility with same-file modules)
	// If imports are present, module loader has already resolved them
	if !hasImports {
		resolver := semantic.NewModuleResolver(symbolTable)
		resolutionErrors := resolver.ResolveModules(file)
		if len(resolutionErrors) > 0 {
			fmt.Fprintf(os.Stderr, "Module resolution errors in %s:\n", filename)
			for _, errMsg := range resolutionErrors {
				fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
			}
			return fmt.Errorf("module resolution failed with %d error(s)", len(resolutionErrors))
		}
	}

	// Build type environment from declarations (same logic as typecheck)
	// First pass: collect type aliases (including those inside modules)
	typeEnv := types.NewEnvironment()
	typeAliases := make(map[string]ast.Type) // Store AST types for later resolution
	
	var collectTypeAliases func(ast.Decl)
	collectTypeAliases = func(d ast.Decl) {
		switch decl := d.(type) {
		case *ast.TypeAliasDecl:
			// Store the AST type for later resolution (we need all types defined first)
			typeAliases[decl.Name] = decl.Type
		case *ast.ModuleDecl:
			// Also collect type aliases from inside modules
			for _, moduleDecl := range decl.Decls {
				collectTypeAliases(moduleDecl)
			}
		}
	}
	
	for _, decl := range file.Decls {
		collectTypeAliases(decl)
	}
	
	// Second pass: resolve type aliases and build environment
	// Create a resolver function that can resolve named types
	var resolveType func(ast.Type) (types.Type, error)
	resolveType = func(astType ast.Type) (types.Type, error) {
		return types.FromASTWithResolver(astType, func(name string) (types.Type, bool) {
			// First check if it's a type alias we've defined
			if aliasAST, ok := typeAliases[name]; ok {
				// Recursively resolve the alias type
				resolved, err := resolveType(aliasAST)
				if err == nil {
					return resolved, true
				}
			}
			// Then check the environment (for already-resolved type aliases)
			return typeEnv.LookupType(name)
		})
	}
	
	// Now resolve all type aliases and store them
	for name, aliasAST := range typeAliases {
		resolvedType, err := resolveType(aliasAST)
		if err == nil {
			typeEnv.DeclareType(name, resolvedType)
		}
	}
	
	// Create a temporary checker for resolving named types (after type aliases are stored)
	tempCheckerForResolve := types.NewChecker(typeEnv)
	
	// Create a new resolveType function that fully resolves types (for use after type aliases are stored)
	// This uses the environment where all type aliases have been resolved and stored
	resolveTypeFully := func(astType ast.Type) (types.Type, error) {
		// First resolve using FromASTWithResolver which should resolve NamedType to the actual type
		typ, err := types.FromASTWithResolver(astType, func(name string) (types.Type, bool) {
			// Look up in environment (all type aliases should be stored by now)
			if resolved, found := typeEnv.LookupType(name); found {
				// Fully resolve the looked-up type to handle any nested Named types
				// Multiple passes to ensure deep resolution
				resolvedType := resolved
				for i := 0; i < 5; i++ {
					prevResolved := resolvedType
					resolvedType = tempCheckerForResolve.ResolveNamedTypesInType(resolvedType)
					if resolvedType == prevResolved {
						break
					}
				}
				return resolvedType, true
			}
			return nil, false
		})
		if err != nil {
			return nil, err
		}
		// Fully resolve any named types in the result (multiple passes to ensure everything is resolved)
		// This handles cases where FromASTWithResolver might have created Named types
		resolved := typ
		for i := 0; i < 5; i++ { // More passes to ensure deep resolution
			prevResolved := resolved
			resolved = tempCheckerForResolve.ResolveNamedTypesInType(resolved)
			if resolved == prevResolved {
				break
			}
		}
		return resolved, nil
	}
	
	// Third pass: add enums, variables, constants, and functions (using the resolver)
	var processDecl func(ast.Decl)
	processDecl = func(d ast.Decl) {
		switch decl := d.(type) {
		case *ast.EnumDecl:
			// Declare enum as a type
			enumType := &types.Enum{
				Name:   decl.Name,
				Values: decl.Values,
			}
			typeEnv.DeclareType(decl.Name, enumType)
		case *ast.TypeAliasDecl:
			// Already handled in first pass
		case *ast.VariableDecl:
			typ, err := resolveTypeFully(decl.Type)
			if err != nil {
				// If type resolution fails, try to resolve it as a named type
				// This handles cases where the type alias might not be fully resolved yet
				if namedType, ok := decl.Type.(*ast.NamedType); ok {
					if resolvedType, found := typeEnv.LookupType(namedType.Name); found {
						typeEnv.DeclareVariable(decl.Name, resolvedType)
					}
				}
			} else {
				typeEnv.DeclareVariable(decl.Name, typ)
			}
		case *ast.ConstantDecl:
			typ, err := resolveType(decl.Type)
			if err == nil {
				typeEnv.DeclareConstant(decl.Name, typ)
			}
		case *ast.FunctionDecl:
			params := make([]types.Type, len(decl.Parameters))
			for i, param := range decl.Parameters {
				paramType, err := resolveType(param.Type)
				if err == nil {
					params[i] = paramType
				}
			}
			returnType, err := resolveType(decl.ReturnType)
			if err == nil {
				sig := &types.FunctionSignature{
					Parameters: params,
					Return:     returnType,
				}
				typeEnv.DeclareFunction(decl.Name, sig)
			}
		case *ast.ModuleDecl:
			// Process declarations inside modules
			for _, moduleDecl := range decl.Decls {
				processDecl(moduleDecl)
			}
		}
	}
	
	for _, decl := range file.Decls {
		processDecl(decl)
	}

	// Type check
	checker := types.NewChecker(typeEnv)
	
	// Populate modules map for module-qualified name resolution
	// This includes both imported modules (from loader) and same-file modules
	modulesMap := make(map[string]*ast.ModuleDecl)
	
	// Add modules from loader if we have imports
	if hasImports && loader != nil {
		allModules := loader.GetAllModules()
		for moduleName, moduleInfo := range allModules {
			if moduleInfo.Module != nil {
				modulesMap[moduleName] = moduleInfo.Module
			}
		}
	}
	
	// Also add modules from the current file (same-file modules)
	for _, decl := range file.Decls {
		if moduleDecl, ok := decl.(*ast.ModuleDecl); ok {
			modulesMap[moduleDecl.Name] = moduleDecl
		}
	}
	
	checker.SetModules(modulesMap)
	
	var typeCheckDecl func(ast.Decl)
	typeCheckDecl = func(d ast.Decl) {
		switch decl := d.(type) {
		case *ast.ActionDecl:
			// Create a new environment for this action with its parameters
			actionEnv := types.NewChildEnvironment(typeEnv)
			for _, param := range decl.Parameters {
				paramType, err := types.FromASTWithResolver(param.Type, func(name string) (types.Type, bool) {
					return typeEnv.LookupType(name)
				})
				if err == nil {
					actionEnv.DeclareVariable(param.Name, paramType)
				}
			}
			actionChecker := types.NewChecker(actionEnv)
			// Copy modules map to action checker so it can resolve module-qualified names
			// Build the modules map (same logic as for main checker)
			actionModulesMap := make(map[string]*ast.ModuleDecl)
			if hasImports && loader != nil {
				allModules := loader.GetAllModules()
				for moduleName, moduleInfo := range allModules {
					if moduleInfo.Module != nil {
						actionModulesMap[moduleName] = moduleInfo.Module
					}
				}
			}
			// Also add modules from the current file (same-file modules)
			for _, decl := range file.Decls {
				if moduleDecl, ok := decl.(*ast.ModuleDecl); ok {
					actionModulesMap[moduleDecl.Name] = moduleDecl
				}
			}
			actionChecker.SetModules(actionModulesMap)
			
			if decl.Body != nil {
				for _, stmt := range decl.Body.Statements {
					switch s := stmt.(type) {
					case *ast.AssignStmt:
						actionChecker.CheckAssignment(s)
					case *ast.RequireStmt:
						actionChecker.CheckExpression(s.Condition)
					case *ast.EnsureStmt:
						actionChecker.CheckExpression(s.Condition)
					}
				}
			}
			// Merge errors from action checker
			checker.MergeErrors(actionChecker)
		case *ast.InvariantDecl:
			checker.CheckExpression(decl.Condition)
		case *ast.TemporalDecl:
			// Type-check temporal expressions - they contain expressions that reference state variables
			checker.CheckExpression(decl.Expression)
		case *ast.ModuleDecl:
			// Type-check declarations inside modules
			for _, moduleDecl := range decl.Decls {
				typeCheckDecl(moduleDecl)
			}
		}
	}
	
	for _, decl := range file.Decls {
		typeCheckDecl(decl)
	}

	typeErrors := checker.Errors()
	if len(typeErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Type errors in %s:\n", filename)
		formatter := errors.NewErrorFormatter()
		for _, typeErr := range typeErrors {
			formatted := formatter.FormatTypeError(typeErr.Message, typeErr.Position)
			fmt.Fprintf(os.Stderr, "  %s\n", formatted)
		}
		return fmt.Errorf("type checking failed with %d error(s)", len(typeErrors))
	}

		fmt.Printf("✓ Successfully type-checked %s\n", filename)
		return nil
	})
}

// printStutterAnalysis prints repair suggestions for detected stuttering steps.
func printStutterAnalysis(stuttering []*explore.Stuttering, file *ast.File, sm *exec.StateMachine) {
	repairs := diagnose.AnalyzeStuttering(stuttering, file, sm)
	if len(repairs) == 0 {
		return
	}
	// Cap output to avoid flooding for specs with many stutter conditions.
	if len(repairs) > 3 {
		repairs = repairs[:3]
	}
	fmt.Printf("\n  Stutter Analysis:\n")
	for i, r := range repairs {
		fmt.Printf("  [%d] %s\n", i+1, r.NoOpSummary)
		for j, opt := range r.Options {
			fmt.Printf("      Option %d — %s\n", j+1, opt.Title)
			fmt.Printf("        %s\n", opt.Explanation)
			for _, line := range strings.Split(opt.SpectreCode, "\n") {
				fmt.Printf("        %s\n", line)
			}
		}
	}
}

// verifyCEGISFix temporarily injects a require guard into an action, re-runs BFS, and
// returns true if the named invariant is no longer violated.  Always restores the action.
func verifyCEGISFix(
	file *ast.File,
	allFiles []*ast.File,
	actionName string,
	wpExpr ast.Expr,
	invariantName string,
	maxStates, maxDepth int,
) bool {
	// Find the action declaration (top-level or inside a module).
	var actionDecl *ast.ActionDecl
outer:
	for _, d := range file.Decls {
		if a, ok := d.(*ast.ActionDecl); ok && a.Name == actionName {
			actionDecl = a
			break
		}
		if m, ok := d.(*ast.ModuleDecl); ok {
			for _, md := range m.Decls {
				if a, ok := md.(*ast.ActionDecl); ok && a.Name == actionName {
					actionDecl = a
					break outer
				}
			}
		}
	}
	if actionDecl == nil || actionDecl.Body == nil {
		return false
	}

	// Temporarily prepend the synthesized guard.
	origStmts := actionDecl.Body.Statements
	actionDecl.Body.Statements = append(
		[]ast.Stmt{&ast.RequireStmt{Condition: wpExpr}},
		origStmts...,
	)
	defer func() { actionDecl.Body.Statements = origStmts }()

	// Build a fresh state machine with the fix in place.
	newSM, err := exec.NewStateMachine(file, allFiles...)
	if err != nil {
		return false
	}

	// Re-run BFS silently — no progress output, no interactive callbacks.
	exp := explore.NewExplorer(newSM)
	exp.SetSilent(true)
	exp.SetMaxStates(maxStates)
	exp.SetMaxDepth(maxDepth)
	fixedResult, err := exp.ExploreBFS()
	if err != nil {
		return false
	}

	for _, v := range fixedResult.Violations {
		if v.Invariant == invariantName {
			return false
		}
	}
	return true
}

// runVerify verifies a Spectre specification
func runVerify(args []string) error {
	// Check for flags
	verbose := false
	maxStates := 5000  // Default: increased for large specs like elevator controller
	maxDepth := 100    // Default: increased for deep state spaces
	emitTraces := ""   // Output file for ITF trace (empty = disabled)
	emitRustTests := "" // Output file for Rust regression tests (empty = disabled)
	coverageMode := "action" // action | transition-pair | boundary | rare-action | property
	useCache := false        // --use-cache: skip BFS if spec hash matches cached graph
	incremental := false     // --incremental: re-verify only the changed action (requires --changed-action)
	changedAction := ""      // --changed-action NAME: action whose semantics changed
	checkRepairMinimal := false // --check-repair-minimal: verify guards don't block valid traces
	params := map[string]int64{} // --param Name=Value spec parameter bindings
	filteredArgs := []string{}

	i := 0
	for i < len(args) {
		arg := args[i]
		switch arg {
		case "--verbose", "-v":
			verbose = true
		case "--emit-traces":
			if i+1 < len(args) {
				emitTraces = args[i+1]
				i++
			} else {
				return fmt.Errorf("--emit-traces requires a file path")
			}
		case "--emit-rust-tests":
			if i+1 < len(args) {
				emitRustTests = args[i+1]
				i++
			} else {
				return fmt.Errorf("--emit-rust-tests requires a file path")
			}
		case "--coverage-mode":
			if i+1 < len(args) {
				coverageMode = args[i+1]
				i++
			} else {
				return fmt.Errorf("--coverage-mode requires a value (action|transition-pair|boundary|rare-action|property)")
			}
		case "--incremental":
			incremental = true
		case "--changed-action":
			if i+1 < len(args) {
				changedAction = args[i+1]
				i++
			} else {
				return fmt.Errorf("--changed-action requires an action name")
			}
		case "--max-states":
			if i+1 < len(args) {
				value := args[i+1]
				// Check for special values: infinity, unlimited, or -1
				if value == "infinity" || value == "unlimited" || value == "-1" {
					maxStates = -1 // -1 means unlimited
				} else if n, err := parseInt(value); err == nil && n > 0 {
					maxStates = n
				} else if n, err := parseInt(value); err == nil && n == -1 {
					maxStates = -1
				} else {
					return fmt.Errorf("invalid value for --max-states: %s (must be a positive integer, 'infinity', 'unlimited', or -1)", value)
				}
				i++ // Skip next argument as it's the value
			} else {
				return fmt.Errorf("--max-states requires a value")
			}
		case "--max-depth":
			if i+1 < len(args) {
				value := args[i+1]
				// Check for special values: infinity, unlimited, or -1
				if value == "infinity" || value == "unlimited" || value == "-1" {
					maxDepth = -1 // -1 means unlimited
				} else if n, err := parseInt(value); err == nil && n > 0 {
					maxDepth = n
				} else if n, err := parseInt(value); err == nil && n == -1 {
					maxDepth = -1
				} else {
					return fmt.Errorf("invalid value for --max-depth: %s (must be a positive integer, 'infinity', 'unlimited', or -1)", value)
				}
				i++ // Skip next argument as it's the value
			} else {
				return fmt.Errorf("--max-depth requires a value")
			}
		case "--use-cache":
			useCache = true
		case "--check-repair-minimal":
			checkRepairMinimal = true
		case "--param":
			if i+1 < len(args) {
				kv := args[i+1]
				i++
				eqIdx := strings.Index(kv, "=")
				if eqIdx < 0 {
					return fmt.Errorf("--param requires Name=Value format, got %q", kv)
				}
				pName := kv[:eqIdx]
				pVal, parseErr := strconv.ParseInt(kv[eqIdx+1:], 10, 64)
				if parseErr != nil {
					return fmt.Errorf("--param value must be an integer: %v", parseErr)
				}
				params[pName] = pVal
			} else {
				return fmt.Errorf("--param requires a Name=Value argument")
			}
		default:
			filteredArgs = append(filteredArgs, arg)
		}
		i++
	}
	args = filteredArgs
	
	if len(args) == 0 {
		return fmt.Errorf("no file specified")
	}

	processor := NewFileProcessor()
	return processor.ProcessFiles(args, func(filename string) error {
	// First, do a quick parse to check if the file has imports
	content, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: File not found - %s\n", filename)
			return fmt.Errorf("File not found - %s", filename)
		}
		return fmt.Errorf("error reading file %s: %w", filename, err)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		fmt.Fprintf(os.Stderr, "Parse errors in %s:\n", filename)
		for _, errMsg := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
		}
		return fmt.Errorf("parse failed with %d error(s)", len(p.Errors()))
	}

	// Check if file has any imports
	hasImports := false
	for _, decl := range file.Decls {
		if _, ok := decl.(*ast.ImportDecl); ok {
			hasImports = true
			break
		}
	}

	// Declare loader outside the if block so it's accessible later
	var loader *semantic.ModuleLoader

	// If file has imports, use module loader (which requires modules)
	// If file has no imports, use the parsed file directly
	if hasImports {
		fileDir := filepath.Dir(filename)
		
		// Use module loader to handle file-based imports
		loader = semantic.NewModuleLoader(fileDir)
		moduleInfo, loadErrors := loader.LoadModule(filename)

		if len(loadErrors) > 0 {
			fmt.Fprintf(os.Stderr, "Parse/module errors in %s:\n", filename)
			for _, errMsg := range loadErrors {
				fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
			}
			
			// Check for circular dependencies
			circularErrors := loader.CheckCircularDependencies()
			if len(circularErrors) > 0 {
				for _, errMsg := range circularErrors {
					fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
				}
			}
			
			return fmt.Errorf("parse/module loading failed with %d error(s)", len(loadErrors)+len(circularErrors))
		}

		// Check for circular dependencies
		circularErrors := loader.CheckCircularDependencies()
		if len(circularErrors) > 0 {
			fmt.Fprintf(os.Stderr, "Circular dependency errors:\n")
			for _, errMsg := range circularErrors {
				fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
			}
			return fmt.Errorf("circular dependencies detected")
		}

		// Use the loaded file from module loader
		file = moduleInfo.File
	}

	// Build symbol table
	builder := semantic.NewBuilder()
	symbolTable, buildErrors := builder.BuildSymbolTable(file)
	if len(buildErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Symbol table build errors in %s:\n", filename)
		for _, errMsg := range buildErrors {
			fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
		}
		return fmt.Errorf("symbol table build failed with %d error(s)", len(buildErrors))
	}

	// Only use old resolver if no imports at all (for backward compatibility with same-file modules)
	// If imports are present, module loader has already resolved them
	if !hasImports {
		resolver := semantic.NewModuleResolver(symbolTable)
		resolutionErrors := resolver.ResolveModules(file)
		if len(resolutionErrors) > 0 {
			fmt.Fprintf(os.Stderr, "Module resolution errors in %s:\n", filename)
			for _, errMsg := range resolutionErrors {
				fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
			}
			return fmt.Errorf("module resolution failed with %d error(s)", len(resolutionErrors))
		}
	}

	// Build type environment from declarations (same logic as typecheck)
	// First pass: collect type aliases (including those inside modules)
	typeEnv := types.NewEnvironment()
	typeAliases := make(map[string]ast.Type) // Store AST types for later resolution
	
	var collectTypeAliases func(ast.Decl)
	collectTypeAliases = func(d ast.Decl) {
		switch decl := d.(type) {
		case *ast.TypeAliasDecl:
			// Store the AST type for later resolution (we need all types defined first)
			typeAliases[decl.Name] = decl.Type
		case *ast.ModuleDecl:
			// Also collect type aliases from inside modules
			for _, moduleDecl := range decl.Decls {
				collectTypeAliases(moduleDecl)
			}
		}
	}
	
	for _, decl := range file.Decls {
		collectTypeAliases(decl)
	}
	
	// Second pass: resolve type aliases and build environment
	// Create a resolver function that can resolve named types
	var resolveType func(ast.Type) (types.Type, error)
	resolveType = func(astType ast.Type) (types.Type, error) {
		return types.FromASTWithResolver(astType, func(name string) (types.Type, bool) {
			// First check if it's a type alias we've defined
			if aliasAST, ok := typeAliases[name]; ok {
				// Recursively resolve the alias type
				resolved, err := resolveType(aliasAST)
				if err == nil {
					return resolved, true
				}
			}
			// Then check the environment (for already-resolved type aliases)
			return typeEnv.LookupType(name)
		})
	}
	
	// Now resolve all type aliases and store them
	for name, aliasAST := range typeAliases {
		resolvedType, err := resolveType(aliasAST)
		if err == nil {
			typeEnv.DeclareType(name, resolvedType)
		}
	}
	
	// Create a temporary checker for resolving named types (after type aliases are stored)
	tempCheckerForResolve := types.NewChecker(typeEnv)
	
	// Create a new resolveType function that fully resolves types (for use after type aliases are stored)
	// This uses the environment where all type aliases have been resolved and stored
	resolveTypeFully := func(astType ast.Type) (types.Type, error) {
		// First resolve using FromASTWithResolver which should resolve NamedType to the actual type
		typ, err := types.FromASTWithResolver(astType, func(name string) (types.Type, bool) {
			// Look up in environment (all type aliases should be stored by now)
			if resolved, found := typeEnv.LookupType(name); found {
				// Fully resolve the looked-up type to handle any nested Named types
				// Multiple passes to ensure deep resolution
				resolvedType := resolved
				for i := 0; i < 5; i++ {
					prevResolved := resolvedType
					resolvedType = tempCheckerForResolve.ResolveNamedTypesInType(resolvedType)
					if resolvedType == prevResolved {
						break
					}
				}
				return resolvedType, true
			}
			return nil, false
		})
		if err != nil {
			return nil, err
		}
		// Fully resolve any named types in the result (multiple passes to ensure everything is resolved)
		// This handles cases where FromASTWithResolver might have created Named types
		resolved := typ
		for i := 0; i < 5; i++ { // More passes to ensure deep resolution
			prevResolved := resolved
			resolved = tempCheckerForResolve.ResolveNamedTypesInType(resolved)
			if resolved == prevResolved {
				break
			}
		}
		return resolved, nil
	}
	
	// Third pass: add enums, variables, constants, and functions (using the resolver)
	var processDecl func(ast.Decl)
	processDecl = func(d ast.Decl) {
		switch decl := d.(type) {
		case *ast.EnumDecl:
			// Declare enum as a type
			enumType := &types.Enum{
				Name:   decl.Name,
				Values: decl.Values,
			}
			typeEnv.DeclareType(decl.Name, enumType)
		case *ast.TypeAliasDecl:
			// Already handled in first pass
		case *ast.VariableDecl:
			typ, err := resolveTypeFully(decl.Type)
			if err != nil {
				// If type resolution fails, try to resolve it as a named type
				// This handles cases where the type alias might not be fully resolved yet
				if namedType, ok := decl.Type.(*ast.NamedType); ok {
					if resolvedType, found := typeEnv.LookupType(namedType.Name); found {
						typeEnv.DeclareVariable(decl.Name, resolvedType)
					}
				}
			} else {
				typeEnv.DeclareVariable(decl.Name, typ)
			}
		case *ast.ConstantDecl:
			typ, err := resolveType(decl.Type)
			if err == nil {
				typeEnv.DeclareConstant(decl.Name, typ)
			}
		case *ast.FunctionDecl:
			params := make([]types.Type, len(decl.Parameters))
			for i, param := range decl.Parameters {
				paramType, err := resolveType(param.Type)
				if err == nil {
					params[i] = paramType
				}
			}
			returnType, err := resolveType(decl.ReturnType)
			if err == nil {
				sig := &types.FunctionSignature{
					Parameters: params,
					Return:     returnType,
				}
				typeEnv.DeclareFunction(decl.Name, sig)
			}
		case *ast.ModuleDecl:
			// Process declarations inside modules
			for _, moduleDecl := range decl.Decls {
				processDecl(moduleDecl)
			}
		}
	}
	
	for _, decl := range file.Decls {
		processDecl(decl)
	}

	// Type check
	checker := types.NewChecker(typeEnv)
	
	// Populate modules map for module-qualified name resolution
	// This includes both imported modules (from loader) and same-file modules
	modulesMap := make(map[string]*ast.ModuleDecl)
	
	// Add modules from loader if we have imports
	if hasImports && loader != nil {
		allModules := loader.GetAllModules()
		for moduleName, moduleInfo := range allModules {
			if moduleInfo.Module != nil {
				modulesMap[moduleName] = moduleInfo.Module
			}
		}
	}
	
	// Also add modules from the current file (same-file modules)
	for _, decl := range file.Decls {
		if moduleDecl, ok := decl.(*ast.ModuleDecl); ok {
			modulesMap[moduleDecl.Name] = moduleDecl
		}
	}
	
	checker.SetModules(modulesMap)
	
	var typeCheckDecl func(ast.Decl)
	typeCheckDecl = func(d ast.Decl) {
		switch decl := d.(type) {
		case *ast.ActionDecl:
			// Create a new environment for this action with its parameters
			actionEnv := types.NewChildEnvironment(typeEnv)
			for _, param := range decl.Parameters {
				paramType, err := types.FromASTWithResolver(param.Type, func(name string) (types.Type, bool) {
					return typeEnv.LookupType(name)
				})
				if err == nil {
					actionEnv.DeclareVariable(param.Name, paramType)
				}
			}
			actionChecker := types.NewChecker(actionEnv)
			// Copy modules map to action checker so it can resolve module-qualified names
			// Build the modules map (same logic as for main checker)
			actionModulesMap := make(map[string]*ast.ModuleDecl)
			if hasImports && loader != nil {
				allModules := loader.GetAllModules()
				for moduleName, moduleInfo := range allModules {
					if moduleInfo.Module != nil {
						actionModulesMap[moduleName] = moduleInfo.Module
					}
				}
			}
			// Also add modules from the current file (same-file modules)
			for _, decl := range file.Decls {
				if moduleDecl, ok := decl.(*ast.ModuleDecl); ok {
					actionModulesMap[moduleDecl.Name] = moduleDecl
				}
			}
			actionChecker.SetModules(actionModulesMap)
			
			if decl.Body != nil {
				for _, stmt := range decl.Body.Statements {
					switch s := stmt.(type) {
					case *ast.AssignStmt:
						actionChecker.CheckAssignment(s)
					case *ast.RequireStmt:
						actionChecker.CheckExpression(s.Condition)
					case *ast.EnsureStmt:
						actionChecker.CheckExpression(s.Condition)
					}
				}
			}
			// Merge errors from action checker
			checker.MergeErrors(actionChecker)
		case *ast.InvariantDecl:
			checker.CheckExpression(decl.Condition)
		case *ast.TemporalDecl:
			// Type-check temporal expressions - they contain expressions that reference state variables
			checker.CheckExpression(decl.Expression)
		case *ast.ModuleDecl:
			// Type-check declarations inside modules
			for _, moduleDecl := range decl.Decls {
				typeCheckDecl(moduleDecl)
			}
		}
	}
	
	for _, decl := range file.Decls {
		typeCheckDecl(decl)
	}

	typeErrors := checker.Errors()
	if len(typeErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Type errors in %s:\n", filename)
		formatter := errors.NewErrorFormatter()
		for _, typeErr := range typeErrors {
			formatted := formatter.FormatTypeError(typeErr.Message, typeErr.Position)
			fmt.Fprintf(os.Stderr, "  %s\n", formatted)
		}
		return fmt.Errorf("type checking failed with %d error(s)", len(typeErrors))
	}

	// Create state machine
	// For files with imports, we need to pass all module files to register constants/enums from imports
	var allFiles []*ast.File
	if hasImports && loader != nil {
		allFiles = []*ast.File{file}
		allModules := loader.GetAllModules()
		for _, moduleInfo := range allModules {
			if moduleInfo.File != nil && moduleInfo.File != file {
				allFiles = append(allFiles, moduleInfo.File)
			}
		}
	} else {
		allFiles = []*ast.File{file}
	}
	
	sm, err := exec.NewStateMachine(file, allFiles...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating state machine in %s: %v\n", filename, err)
		return fmt.Errorf("error creating state machine: %w", err)
	}

	// Bind spec params supplied via --param flags.
	for pName, pVal := range params {
		sm.SetParam(pName, pVal)
	}

	// Explore state space
	explorer := explore.NewExplorer(sm)
	explorer.SetMaxDepth(maxDepth)
	explorer.SetMaxStates(maxStates)
	explorer.SetVerbose(verbose)
	
	// Track violations for immediate reporting
	invariantViolationsSet := make(map[string]bool) // Track which invariants have been reported
	
	// Set callback to report violations immediately when detected during exploration
	explorer.SetInvariantViolationCallback(func(violation *explore.Violation) bool {
		// Create a unique key for this violation
		violationKey := fmt.Sprintf("%s:%s", violation.Invariant, violation.Description)
		
		// Skip if already reported
		if invariantViolationsSet[violationKey] {
			return true // Continue exploration
		}
		invariantViolationsSet[violationKey] = true
		
		// Report immediately
		fmt.Fprintf(os.Stderr, "\n[VIOLATION DETECTED] Invariant: %s\n", violation.Invariant)
		fmt.Fprintf(os.Stderr, "  %s\n", violation.Description)
		if violation.Path != nil && len(violation.Path) > 0 {
			fmt.Fprintf(os.Stderr, "  Counterexample trace:\n")
			for j, trans := range violation.Path {
				if trans.Action != "" {
					fmt.Fprintf(os.Stderr, "    %d. %s\n", j+1, trans.Action)
				}
			}
		}
		os.Stderr.Sync()
		return true // always continue exploration
	})
	
	// Set up temporal verification callback for incremental checking
	constraintModel := sm.GetConstraintModel()
	temporalProps := constraintModel.GetTemporalProperties()
	if len(temporalProps) > 0 {
		hasher := explore.NewStateHasher()
		temporalVerifier := explore.NewTemporalVerifier(hasher, file)
		temporalVerifier.SetStateMachine(sm)
		
		// Track temporal violations for immediate reporting
		temporalViolationsSet := make(map[string]bool)
		shouldContinueTemporal := true
		
		// Set callback for immediate temporal violation reporting
		temporalVerifier.SetViolationCallback(func(result *explore.TemporalVerificationResult) {
			if temporalViolationsSet[result.PropertyName] {
				return
			}
			temporalViolationsSet[result.PropertyName] = true
			
			// Report immediately with flushed output
			fmt.Fprintf(os.Stderr, "\n[VIOLATION DETECTED] Temporal Property: %s\n", result.PropertyName)
			fmt.Fprintf(os.Stderr, "  %s\n", result.Violation.Description)
			if result.Violation.Trace != nil && result.Violation.Trace.Length() > 0 {
				fmt.Fprintf(os.Stderr, "  Counterexample trace:\n")
				for j := 0; j < result.Violation.Trace.Length(); j++ {
					action := result.Violation.Trace.GetAction(j)
					if action != "" {
						fmt.Fprintf(os.Stderr, "    %d. %s\n", j+1, action)
					}
				}
			}
			os.Stderr.Sync()
			shouldContinueTemporal = true // always continue exploration
		})
		
		// Set incremental temporal verification callback (check every 5 states)
		// Capture shouldContinueTemporal in closure to track user response
		shouldContinueTemporalPtr := &shouldContinueTemporal
		explorer.SetTemporalVerificationCallback(func(graph *explore.TransitionGraph, initialStates []*state.State) bool {
			// Check each temporal property incrementally
			for _, prop := range temporalProps {
				// Skip if already found violation for this property
				if temporalViolationsSet[prop.Name] {
					continue
				}
				
				verificationResult, err := temporalVerifier.VerifyTemporalProperty(prop, graph, initialStates)
				if err != nil {
					// Skip on error, continue checking
					continue
				}
				
				if !verificationResult.Holds {
					// Violation detected - callback already handled reporting and user prompt
					// Return the user's response (whether to continue)
					return *shouldContinueTemporalPtr
				}
			}
			return true // Continue exploration
		}, 5) // Check every 5 states
	}
	
	// Warn if unlimited exploration is enabled
	if maxStates == -1 || maxDepth == -1 {
		fmt.Fprintf(os.Stderr, "Warning: Unlimited exploration enabled (--max-states: %v, --max-depth: %v)\n", 
			func() string {
				if maxStates == -1 {
					return "unlimited"
				}
				return fmt.Sprintf("%d", maxStates)
			}(),
			func() string {
				if maxDepth == -1 {
					return "unlimited"
				}
				return fmt.Sprintf("%d", maxDepth)
			}())
		fmt.Fprintf(os.Stderr, "This may run until the state space is fully explored or memory is exhausted.\n\n")
	}

	// Build the cache configuration key from params + exploration limits.
	cacheCfg := explore.CacheConfig{
		Params:    params,
		MaxStates: maxStates,
		MaxDepth:  maxDepth,
	}

	// Try to restore from cache when --use-cache is set and spec+config unchanged.
	var result *explore.ExplorationResult
	if useCache {
		if cache, cacheErr := explore.LoadCache(filename, cacheCfg); cache != nil {
			fmt.Printf("Restored state graph from cache (%d states).\n", len(cache.States))
			result = explore.RestoreResult(cache)
			if incremental && changedAction != "" {
				// Re-verify only the changed action, leaving all other transitions intact.
				hasher := explore.NewStateHasher()
				result, err = explore.IncrementalReverify(result, changedAction, sm, hasher, verbose)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: incremental re-verify failed, falling back to full BFS: %v\n", err)
					result = nil
				}
			}
		} else if cacheErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: cache read error: %v\n", cacheErr)
		}
	}
	if result == nil {
		result, err = explorer.ExploreBFS()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error exploring state space in %s: %v\n", filename, err)
			return fmt.Errorf("error exploring state space: %w", err)
		}
		if useCache {
			if saveErr := explore.SaveCache(filename, result, content, cacheCfg); saveErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not save cache: %v\n", saveErr)
			}
		}
	}

	// Report results
	hasViolations := len(result.Violations) > 0
	temporalViolations := []*explore.TemporalVerificationResult{}
	
	// Final verification of temporal properties (only if not already checked incrementally)
	// Incremental verification during exploration already checks and reports violations,
	// but we do a final check to catch any violations that might have been missed
	// No user prompts in final verification - just collect violations for final report
	if len(temporalProps) > 0 {
		hasherFinal := explore.NewStateHasher()
		temporalVerifierFinal := explore.NewTemporalVerifier(hasherFinal, file)
		temporalVerifierFinal.SetStateMachine(sm)
		
		// Track violations for final reporting (only report if not already reported incrementally)
		temporalViolationsSetFinal := make(map[string]bool)
		
		// No callback for final verification - we just collect violations for the final report
		// (No user prompts at the end)
		
		for _, prop := range temporalProps {
			verificationResult, err := temporalVerifierFinal.VerifyTemporalProperty(prop, result.TransitionGraph, result.InitialStates)
			if err != nil {
				// Skip on error, continue checking
				continue
			}
			if !verificationResult.Holds {
				// If not already reported incrementally, add it now for final report
				if !temporalViolationsSetFinal[verificationResult.PropertyName] {
					temporalViolationsSetFinal[verificationResult.PropertyName] = true
					hasViolations = true
					temporalViolations = append(temporalViolations, verificationResult)
				}
			}
		}
	}
	
	// Emit ITF trace if requested (always, regardless of violations).
	if emitTraces != "" {
		var trace *itf.Trace
		if len(result.Violations) > 0 {
			trace = itf.FromViolation(filename, result.Violations[0])
		} else {
			hasher := explore.NewStateHasher()
			rng := rand.New(rand.NewSource(42))
			switch coverageMode {
			case "transition-pair":
				trace = itf.TransitionPairWalk(filename, result.TransitionGraph, result.InitialStates, 20, rng, hasher)
			case "boundary":
				trace = itf.BoundaryCoverageWalk(filename, result.TransitionGraph, result.InitialStates, 20, rng, hasher)
			case "rare-action":
				trace = itf.RareActionWalk(filename, result.TransitionGraph, result.InitialStates, 20, rng, hasher)
			case "property":
				trace = itf.PropertyDirectedWalk(filename, result.TransitionGraph, result.InitialStates, 20, rng, hasher)
			default: // "action" or unrecognised
				trace = itf.RandomWalk(filename, result.TransitionGraph, result.InitialStates, 20, rng, hasher)
			}
		}
		if trace != nil {
			data, jsonErr := json.MarshalIndent(trace, "", "  ")
			if jsonErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not marshal ITF trace: %v\n", jsonErr)
			} else if writeErr := os.WriteFile(emitTraces, data, 0644); writeErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not write ITF trace to %s: %v\n", emitTraces, writeErr)
			} else {
				fmt.Printf("ITF trace written to %s\n", emitTraces)
			}
		}
	}

	// Emit Rust regression tests if requested (only when violations exist).
	if emitRustTests != "" && len(result.Violations) > 0 {
		specBase := strings.TrimSuffix(filepath.Base(filename), ".spec")
		metaPath := codegen.MetaPath(filename)
		meta, _ := codegen.LoadDriverMeta(metaPath)
		var testCases []codegen.RustTestCase
		for _, v := range result.Violations {
			tc := codegen.RustTestCase{
				InvariantName: v.Invariant,
			}
			var initVars map[string]interface{}
			if len(v.Path) > 0 && v.Path[0].FromState != nil {
				initVars = make(map[string]interface{})
				for k, val := range v.Path[0].FromState.Variables {
					initVars[k] = val.String()
				}
			} else if v.State != nil {
				initVars = make(map[string]interface{})
				for k, val := range v.State.Variables {
					initVars[k] = val.String()
				}
			}
			tc.InitVars = initVars
			for _, trans := range v.Path {
				tc.ActionPath = append(tc.ActionPath, trans.Action)
				step := codegen.RustTestStep{Action: trans.Action}
				if trans.ToState != nil {
					step.Expected = make(map[string]interface{})
					for k, val := range trans.ToState.Variables {
						step.Expected[k] = val.String()
					}
				}
				tc.Steps = append(tc.Steps, step)
			}
			testCases = append(testCases, tc)
		}
		rustCode := codegen.EmitRustTests(specBase, meta, testCases)
		if writeErr := os.WriteFile(emitRustTests, []byte(rustCode), 0644); writeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not write Rust tests to %s: %v\n", emitRustTests, writeErr)
		} else {
			fmt.Printf("Rust regression tests written to %s (%d test(s))\n", emitRustTests, len(testCases))
		}
	}

	// Final report
	fmt.Printf("Traversed %d states\n", result.StatesExplored)

	if hasViolations {
		fmt.Printf("Violations:\n")
		
		// Report invariant violations
		violationNum := 1
		for _, violation := range result.Violations {
			fmt.Printf("%d. Invariant: %s\n", violationNum, violation.Description)
			if len(violation.Path) > 0 {
				fmt.Printf("   Path: ")
				actions := make([]string, 0, len(violation.Path))
				for _, transition := range violation.Path {
					if transition.Action != "" {
						actions = append(actions, transition.Action)
					}
				}
				fmt.Printf("%s\n", strings.Join(actions, " → "))
			}
			violationNum++
		}
		
		// Report temporal property violations
		for _, violation := range temporalViolations {
			fmt.Printf("%d. Temporal Property (%s): %s\n", violationNum, violation.PropertyName, violation.Violation.Description)
			if violation.Violation.Trace != nil && violation.Violation.Trace.Length() > 0 {
				fmt.Printf("   Counterexample: ")
				actions := make([]string, 0, violation.Violation.Trace.Length())
				for j := 0; j < violation.Violation.Trace.Length(); j++ {
					action := violation.Violation.Trace.GetAction(j)
					if action != "" {
						actions = append(actions, action)
					}
				}
				fmt.Printf("%s\n", strings.Join(actions, " → "))
			}
			violationNum++
		}
		
		// CEGIS repair suggestions for invariant violations
		if len(result.Violations) > 0 {
			cegisRepairs := diagnose.SynthesizeCEGIS(result.Violations, file)
			if len(cegisRepairs) > 0 {
				fmt.Printf("\nCEGIS Repair Suggestions:\n")
				for _, repair := range cegisRepairs {
					fmt.Printf("\n  Invariant `%s` violated via action `%s`", repair.InvariantName, repair.ActionName)
					if len(repair.ActionArgs) > 0 {
						argStrs := make([]string, len(repair.ActionArgs))
						for i, a := range repair.ActionArgs {
							argStrs[i] = a.String()
						}
						fmt.Printf("(%s)", strings.Join(argStrs, ", "))
					}
					fmt.Printf(":\n")
					if repair.PreState != nil {
						pairs := make([]string, 0, len(repair.PreState.Variables))
						for k, v := range repair.PreState.Variables {
							pairs = append(pairs, k+" = "+v.String())
						}
						sort.Strings(pairs)
						fmt.Printf("    Pre-state: {%s}\n", strings.Join(pairs, ", "))
					}
					if len(repair.CounterexPath) > 0 {
						fmt.Printf("    Counterexample: %s\n", strings.Join(repair.CounterexPath, " → "))
					}
					// Structured explanation.
					{
						var repairPtr *diagnose.CEGISRepair
						repairPtr = &repair
						// find the matching violation for this repair
						for _, viol := range result.Violations {
							if viol.Invariant == repair.InvariantName {
								expl := diagnose.Explain(viol, repairPtr, file)
								fmt.Printf("%s", diagnose.PrintExplanation(expl))
								break
							}
						}
					}
					if len(repair.Guards) == 0 {
						fmt.Printf("    (Could not synthesize an automatic repair for this violation)\n")
						continue
					}
					for i, guard := range repair.Guards {
						fmt.Printf("    Option %d — weakest precondition for `%s' = %s`:\n", i+1, guard.VarName, guard.AssignRHS)
						fmt.Printf("      %s\n", diagnose.RootCauseDesc(guard, repair.PreState, repair.ActionArgs))
						for _, line := range strings.Split(guard.SpectreCode, "\n") {
							fmt.Printf("      %s\n", line)
						}
						verified := verifyCEGISFix(file, allFiles, repair.ActionName, guard.WPExpr, repair.InvariantName, maxStates, maxDepth)
						if verified {
							fmt.Printf("      ✓ Verified: re-explored with this guard applied — invariant holds\n")
						} else {
							fmt.Printf("      ! Apply and re-run `spectre verify` to check for remaining violations\n")
						}
						if checkRepairMinimal {
							rng := rand.New(rand.NewSource(42))
							hasher := explore.NewStateHasher()
							posWalk := itf.RandomWalk(filename, result.TransitionGraph, result.InitialStates, 50, rng, hasher)
							if posWalk != nil {
								rawTrace := posWalk.RawStates()
								blocked := diagnose.CheckGuardMinimality(guard, repair.ActionName, rawTrace, file)
								if len(blocked) > 0 {
									fmt.Printf("      ! Guard may be over-restrictive: blocks %d valid %q step(s) in a positive trace\n",
										len(blocked), repair.ActionName)
								} else {
									fmt.Printf("      ✓ Minimality: guard does not block any valid %q steps\n", repair.ActionName)
								}
							}
						}
					}
				}
			}
		}

		// Report stuttering warnings even when there are violations
		if len(result.Stuttering) > 0 {
			fmt.Printf("\nWarnings (Stuttering):\n")
			fmt.Printf("Found %d stuttering step(s) where a state transitions back to itself.\n", len(result.Stuttering))
			fmt.Printf("Stuttering can indicate missing fairness constraints or incomplete specifications.\n")
			if verbose {
				for i, stutter := range result.Stuttering {
					fmt.Printf("  %d. %s\n", i+1, stutter.Description)
					if len(stutter.Args) > 0 {
						argsStr := make([]string, len(stutter.Args))
						for j, arg := range stutter.Args {
							argsStr[j] = arg.String()
						}
						fmt.Printf("     Action: %s(%s)\n", stutter.Action, strings.Join(argsStr, ", "))
					} else {
						fmt.Printf("     Action: %s\n", stutter.Action)
					}
				}
			}
			printStutterAnalysis(result.Stuttering, file, sm)
		}

		return fmt.Errorf("verification failed with %d violation(s)", len(result.Violations)+len(temporalViolations))
	}

	// Report stuttering warnings when there are no violations
	if len(result.Stuttering) > 0 {
		fmt.Printf("\nWarnings (Stuttering):\n")
		fmt.Printf("Found %d stuttering step(s) where a state transitions back to itself.\n", len(result.Stuttering))
		fmt.Printf("Stuttering can indicate missing fairness constraints or incomplete specifications.\n")
		if verbose {
			for i, stutter := range result.Stuttering {
				fmt.Printf("  %d. %s\n", i+1, stutter.Description)
				if len(stutter.Args) > 0 {
					argsStr := make([]string, len(stutter.Args))
					for j, arg := range stutter.Args {
						argsStr[j] = arg.String()
					}
					fmt.Printf("     Action: %s(%s)\n", stutter.Action, strings.Join(argsStr, ", "))
				} else {
					fmt.Printf("     Action: %s\n", stutter.Action)
				}
			}
		}
		printStutterAnalysis(result.Stuttering, file, sm)
	}

	fmt.Printf("Found no violations.\n")
	return nil
	})
}

// runGenerateDriver generates a Rust driver skeleton for a Spectre spec.
func runGenerateDriver(args []string) error {
	lang := "rust"
	output := ""
	filteredArgs := []string{}

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--lang":
			if i+1 < len(args) {
				lang = args[i+1]
				i++
			}
		case "--output", "-o":
			if i+1 < len(args) {
				output = args[i+1]
				i++
			}
		default:
			filteredArgs = append(filteredArgs, args[i])
		}
		i++
	}

	if len(filteredArgs) == 0 {
		return fmt.Errorf("no spec file specified")
	}
	if lang != "rust" {
		return fmt.Errorf("unsupported language %q (only 'rust' is supported)", lang)
	}

	filename := filteredArgs[0]
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", filename, err)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		return fmt.Errorf("parse errors in %s: %s", filename, strings.Join(p.Errors(), "; "))
	}

	specName := filepath.Base(filename)
	driver := codegen.ExtractFromAST(file, specName)

	// Load mining sidecar for zero-driver generation; fall back to stub skeleton.
	metaPath := codegen.MetaPath(filename)
	meta, metaErr := codegen.LoadDriverMeta(metaPath)
	if metaErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not read driver sidecar %s: %v\n", metaPath, metaErr)
	}
	var code string
	if meta != nil {
		code = driver.GenerateWithMeta(meta)
		fmt.Fprintf(os.Stderr, "Using mining sidecar %s (struct: %s) — apply_action and current_state fully generated.\n",
			metaPath, meta.StructName)
	} else {
		code = driver.Generate()
	}

	if output == "" {
		fmt.Print(code)
		return nil
	}
	if err := os.WriteFile(output, []byte(code), 0644); err != nil {
		return fmt.Errorf("error writing driver to %s: %w", output, err)
	}
	fmt.Printf("Rust driver written to %s\n", output)
	return nil
}

// runGenerateMonitor generates an embedded Rust runtime monitor for a Spectre spec.
func runGenerateMonitor(args []string) error {
	lang := "rust"
	output := ""
	filteredArgs := []string{}

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--lang":
			if i+1 < len(args) {
				lang = args[i+1]
				i++
			}
		case "--output", "-o":
			if i+1 < len(args) {
				output = args[i+1]
				i++
			}
		default:
			filteredArgs = append(filteredArgs, args[i])
		}
		i++
	}

	if len(filteredArgs) == 0 {
		return fmt.Errorf("no spec file specified")
	}
	if lang != "rust" {
		return fmt.Errorf("unsupported language %q (only 'rust' is supported)", lang)
	}

	filename := filteredArgs[0]
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", filename, err)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		return fmt.Errorf("parse errors in %s: %s", filename, strings.Join(p.Errors(), "; "))
	}

	specName := strings.TrimSuffix(filepath.Base(filename), ".spec")
	mon := codegen.ExtractMonitorFromAST(file, specName)
	code := mon.Generate()

	if output == "" {
		fmt.Print(code)
		return nil
	}
	if err := os.WriteFile(output, []byte(code), 0644); err != nil {
		return fmt.Errorf("error writing monitor to %s: %w", output, err)
	}
	fmt.Printf("Rust runtime monitor written to %s\n", output)
	fmt.Printf("  Drop monitor.rs into your project and call monitor.step(action, new_state) after every transition.\n")
	if len(mon.Liveness) > 0 {
		fmt.Printf("  Call monitor.unmet_liveness_properties() at shutdown to check liveness obligations.\n")
	}
	return nil
}

// runMine mines a Spectre spec skeleton from Rust source code.
func runMine(args []string) error {
	lang := "rust"
	output := ""
	specName := ""
	useAI := false
	filteredArgs := []string{}

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--lang":
			if i+1 < len(args) {
				lang = args[i+1]
				i++
			}
		case "--output", "-o":
			if i+1 < len(args) {
				output = args[i+1]
				i++
			}
		case "--spec-name", "--name":
			if i+1 < len(args) {
				specName = args[i+1]
				i++
			}
		case "--ai":
			useAI = true
		default:
			filteredArgs = append(filteredArgs, args[i])
		}
		i++
	}

	if len(filteredArgs) == 0 {
		return fmt.Errorf("no Rust source file specified")
	}
	if lang != "rust" {
		return fmt.Errorf("unsupported language %q (only 'rust' is supported)", lang)
	}

	filename := filteredArgs[0]

	if specName == "" {
		specName = strings.TrimSuffix(filepath.Base(filename), ".rs")
	}

	mined, err := mine.MineFromRustFile(filename, specName)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", filename, err)
	}

	var suggestions *mine.LLMSuggestions
	if useAI {
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "Warning: --ai requires ANTHROPIC_API_KEY to be set; running without LLM enhancement.\n")
		} else {
			fmt.Fprintf(os.Stderr, "Enhancing with LLM (model: claude-haiku)...\n")
			suggestions, err = mine.EnhanceWithLLM(mined, apiKey)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: LLM enhancement failed: %v; continuing with static analysis.\n", err)
				suggestions = nil
			}
		}
	}

	specText := mined.Generate(suggestions)

	if output == "" {
		fmt.Print(specText)
		return nil
	}
	if err := os.WriteFile(output, []byte(specText), 0644); err != nil {
		return fmt.Errorf("error writing spec to %s: %w", output, err)
	}

	// Write driver sidecar alongside the spec so generate-driver can emit
	// a fully-connected adapter without any manual edits.
	sidecarPath := codegen.MetaPath(output)
	meta := buildDriverMeta(mined, filename)
	if metaData, err := json.Marshal(meta); err == nil {
		if writeErr := os.WriteFile(sidecarPath, metaData, 0644); writeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not write driver sidecar %s: %v\n", sidecarPath, writeErr)
		}
	}

	fmt.Printf("Mined spec written to %s\n", output)
	fmt.Printf("  Structs found:    %d fields from %s\n", len(mined.Fields), mined.StructName)
	fmt.Printf("  Enums found:      %d\n", len(mined.Enums))
	fmt.Printf("  Actions mined:    %d\n", len(mined.Methods))
	if useAI && suggestions != nil {
		fmt.Printf("  LLM invariants:   %d\n", len(suggestions.ExtraInvariants))
		fmt.Printf("  LLM temporals:    %d\n", len(suggestions.ExtraTemporal))
	}
	fmt.Printf("  Driver sidecar:   %s\n", sidecarPath)
	fmt.Printf("  Edit %s and run: spectre verify %s\n", output, output)
	fmt.Printf("  Zero-driver:      spectre generate-driver %s\n", output)
	return nil
}

// buildDriverMeta converts a MinedSpec into the sidecar JSON that
// generate-driver uses to emit a fully-connected Rust adapter.
func buildDriverMeta(mined *mine.MinedSpec, sourceFile string) *codegen.DriverMeta {
	meta := &codegen.DriverMeta{
		StructName: mined.StructName,
		SourceFile: sourceFile,
	}
	for _, f := range mined.Fields {
		meta.Fields = append(meta.Fields, codegen.MetaField{
			SpecVar:     f.Name,
			RustField:   f.Name,
			RustType:    f.RustType,
			SpectreType: f.SpectreType,
		})
	}
	for _, m := range mined.Methods {
		if !m.MutSelf {
			continue
		}
		action := codegen.MetaAction{
			SpecAction: m.Name,
			RustMethod: m.Name,
		}
		for _, p := range m.Params {
			rt := spectreParamToRust(p.SpectreType)
			action.Params = append(action.Params, codegen.MetaParam{
				Name:        p.Name,
				SpectreType: p.SpectreType,
				RustType:    rt,
			})
		}
		meta.Actions = append(meta.Actions, action)
	}
	return meta
}

// spectreParamToRust maps a Spectre type string to a Rust primitive type.
func spectreParamToRust(t string) string {
	switch t {
	case "int":
		return "i64"
	case "bool":
		return "bool"
	case "str":
		return "String"
	case "float":
		return "f64"
	default:
		return t
	}
}

// runCheckRefinement checks that ITF trace files conform to the spec via the
// driver.json mapping.  Usage: spectre check-refinement [--traces f.itf.json,...] spec.spec
func runCheckRefinement(args []string) error {
	var traceFiles []string
	filteredArgs := []string{}

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--traces":
			if i+1 < len(args) {
				for _, t := range strings.Split(args[i+1], ",") {
					traceFiles = append(traceFiles, strings.TrimSpace(t))
				}
				i++
			} else {
				return fmt.Errorf("--traces requires a comma-separated list of ITF JSON files")
			}
		default:
			filteredArgs = append(filteredArgs, args[i])
		}
		i++
	}

	if len(filteredArgs) == 0 {
		return fmt.Errorf("no spec file specified")
	}

	specFile := filteredArgs[0]
	content, err := os.ReadFile(specFile)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", specFile, err)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	file := p.ParseFile()
	if len(p.Errors()) > 0 {
		return fmt.Errorf("parse errors in %s: %s", specFile, strings.Join(p.Errors(), "; "))
	}

	metaPath := codegen.MetaPath(specFile)
	meta, metaErr := codegen.LoadDriverMeta(metaPath)
	if metaErr != nil || meta == nil {
		return fmt.Errorf("could not load driver sidecar %s: %v\nRun `spectre mine --output` first to generate it", metaPath, metaErr)
	}

	// Load trace files.
	var traces [][]map[string]interface{}
	for _, tf := range traceFiles {
		data, readErr := os.ReadFile(tf)
		if readErr != nil {
			return fmt.Errorf("error reading trace file %s: %w", tf, readErr)
		}
		var root map[string]interface{}
		if jsonErr := json.Unmarshal(data, &root); jsonErr != nil {
			return fmt.Errorf("invalid JSON in trace file %s: %w", tf, jsonErr)
		}
		statesRaw, ok := root["states"].([]interface{})
		if !ok {
			return fmt.Errorf("trace file %s has no 'states' array", tf)
		}
		var states []map[string]interface{}
		for _, s := range statesRaw {
			if m, ok2 := s.(map[string]interface{}); ok2 {
				states = append(states, m)
			}
		}
		traces = append(traces, states)
	}

	if len(traces) == 0 {
		return fmt.Errorf("no trace files loaded; use --traces f1.itf.json,f2.itf.json")
	}

	report := codegen.CheckRefinement(file, meta, traces)
	fmt.Printf("Refinement check: %s (struct: %s)\n", specFile, report.StructName)
	fmt.Printf("  Transitions checked: %d\n", report.CheckedTransitions)
	fmt.Printf("  Legal steps:         %d\n", report.LegalSteps)
	fmt.Printf("  Stutter steps:       %d\n", report.StutterSteps)
	fmt.Printf("  Violations:          %d\n", len(report.Violations))

	if len(report.Violations) > 0 {
		fmt.Printf("\nViolations:\n")
		for i, v := range report.Violations {
			fmt.Printf("  %d. trace[%d] step %d action=%q\n", i+1, v.TraceIndex, v.StepIndex, v.Action)
			fmt.Printf("     %s\n", v.Reason)
		}
		return fmt.Errorf("refinement check failed: %d violation(s)", len(report.Violations))
	}

	fmt.Printf("\n✓ All transitions are legal or stutter steps.\n")
	return nil
}

// printUsage prints usage information
func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: spectre <command> [arguments]\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	for _, cmd := range Commands {
		fmt.Fprintf(os.Stderr, "  %-15s %s\n", cmd.Name, cmd.Description)
	}
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  spectre parse examples/counter.spec\n")
	fmt.Fprintf(os.Stderr, "  spectre typecheck examples/counter.spec\n")
	fmt.Fprintf(os.Stderr, "  spectre verify examples/counter.spec\n")
	fmt.Fprintf(os.Stderr, "  spectre verify examples/counter.spec --max-states 10000 --max-depth 200\n")
	fmt.Fprintf(os.Stderr, "\nFlags for verify command:\n")
	fmt.Fprintf(os.Stderr, "  --verbose, -v           Enable verbose output\n")
	fmt.Fprintf(os.Stderr, "  --max-states <number>   Maximum number of states to explore (default: 5000)\n")
	fmt.Fprintf(os.Stderr, "                         Use 'infinity', 'unlimited', or -1 for unlimited\n")
	fmt.Fprintf(os.Stderr, "  --max-depth <number>    Maximum exploration depth (default: 100)\n")
	fmt.Fprintf(os.Stderr, "                         Use 'infinity', 'unlimited', or -1 for unlimited\n")
	fmt.Fprintf(os.Stderr, "  --emit-traces <file>    Write an ITF execution trace to <file> for use\n")
	fmt.Fprintf(os.Stderr, "                         with spectre-connect Rust model-based testing\n")
	fmt.Fprintf(os.Stderr, "\nFlags for generate-driver command:\n")
	fmt.Fprintf(os.Stderr, "  --lang rust             Target language (only 'rust' is supported)\n")
	fmt.Fprintf(os.Stderr, "  --output, -o <file>     Write driver to <file> instead of stdout\n")
	fmt.Fprintf(os.Stderr, "\nFlags for generate-monitor command:\n")
	fmt.Fprintf(os.Stderr, "  --lang rust             Target language (only 'rust' is supported)\n")
	fmt.Fprintf(os.Stderr, "  --output, -o <file>     Write monitor to <file> instead of stdout\n")
	fmt.Fprintf(os.Stderr, "\nModel-based testing workflow:\n")
	fmt.Fprintf(os.Stderr, "  1. spectre verify myspec.spec --emit-traces trace.itf.json\n")
	fmt.Fprintf(os.Stderr, "  2. spectre generate-driver --lang rust myspec.spec -o driver.rs\n")
	fmt.Fprintf(os.Stderr, "  3. Fill in driver.rs, then: cargo run -- trace.itf.json\n")
	fmt.Fprintf(os.Stderr, "\nRuntime monitoring workflow:\n")
	fmt.Fprintf(os.Stderr, "  1. spectre generate-monitor myspec.spec -o src/monitor.rs\n")
	fmt.Fprintf(os.Stderr, "  2. Call monitor.step(action, new_state) after every state transition\n")
	fmt.Fprintf(os.Stderr, "  3. Call monitor.unmet_liveness_properties() at shutdown\n")
	fmt.Fprintf(os.Stderr, "\nFlags for mine command:\n")
	fmt.Fprintf(os.Stderr, "  --lang rust             Source language (only 'rust' is supported)\n")
	fmt.Fprintf(os.Stderr, "  --output, -o <file>     Write mined spec to <file> instead of stdout\n")
	fmt.Fprintf(os.Stderr, "  --spec-name <name>      Override spec name (default: filename without .rs)\n")
	fmt.Fprintf(os.Stderr, "  --ai                    Enhance with LLM (requires ANTHROPIC_API_KEY env var)\n")
	fmt.Fprintf(os.Stderr, "\nSpec mining workflow:\n")
	fmt.Fprintf(os.Stderr, "  1. spectre mine --lang rust src/account.rs -o account.spec\n")
	fmt.Fprintf(os.Stderr, "  2. Review and refine account.spec\n")
	fmt.Fprintf(os.Stderr, "  3. spectre verify account.spec\n")
}

// findSpecFiles finds all .spec files in a directory
func findSpecFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".spec") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

