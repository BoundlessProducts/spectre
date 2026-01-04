package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/akkeshavan/spectre/internal/errors"
	"github.com/akkeshavan/spectre/internal/exec"
	"github.com/akkeshavan/spectre/internal/explore"
	"github.com/akkeshavan/spectre/internal/lexer"
	"github.com/akkeshavan/spectre/internal/parser"
	"github.com/akkeshavan/spectre/internal/semantic"
	"github.com/akkeshavan/spectre/internal/types"
	"github.com/akkeshavan/spectre/pkg/ast"
)

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

// runVerify verifies a Spectre specification
func runVerify(args []string) error {
	// Check for verbose flag
	verbose := false
	filteredArgs := []string{}
	for _, arg := range args {
		if arg == "--verbose" || arg == "-v" {
			verbose = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
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

	// Explore state space
	explorer := explore.NewExplorer(sm)
	explorer.SetMaxDepth(10)
	explorer.SetMaxStates(50)
	explorer.SetVerbose(verbose)

	result, err := explorer.ExploreBFS()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error exploring state space in %s: %v\n", filename, err)
		return fmt.Errorf("error exploring state space: %w", err)
	}

	// Report results
	hasViolations := len(result.Violations) > 0
	temporalViolations := []*explore.TemporalVerificationResult{}
	
	// Verify temporal properties
	constraintModel := sm.GetConstraintModel()
	temporalProps := constraintModel.GetTemporalProperties()
	if len(temporalProps) > 0 {
		hasher := explore.NewStateHasher()
		temporalVerifier := explore.NewTemporalVerifier(hasher, file)
		temporalVerifier.SetStateMachine(sm) // Set state machine for fairness checking
		
		for _, prop := range temporalProps {
			verificationResult, err := temporalVerifier.VerifyTemporalProperty(prop, result.TransitionGraph, result.InitialStates)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error verifying temporal property %s: %v\n", prop.Name, err)
				continue
			}
			if !verificationResult.Holds {
				hasViolations = true
				temporalViolations = append(temporalViolations, verificationResult)
			}
		}
	}
	
	if hasViolations {
		fmt.Fprintf(os.Stderr, "Verification failed: %d violation(s) found\n", len(result.Violations)+len(temporalViolations))
		
		// Report invariant violations
		for i, violation := range result.Violations {
			fmt.Fprintf(os.Stderr, "\nViolation %d (Invariant):\n", i+1)
			fmt.Fprintf(os.Stderr, "  %s\n", violation.Description)
			if len(violation.Path) > 0 {
				fmt.Fprintf(os.Stderr, "  Path:\n")
				for j, transition := range violation.Path {
					fmt.Fprintf(os.Stderr, "    %d. %s\n", j+1, transition.Action)
				}
			}
		}
		
		// Report temporal property violations
		for i, violation := range temporalViolations {
			fmt.Fprintf(os.Stderr, "\nViolation %d (Temporal Property: %s):\n", len(result.Violations)+i+1, violation.PropertyName)
			fmt.Fprintf(os.Stderr, "  %s\n", violation.Violation.Description)
			if violation.Violation.Trace != nil && violation.Violation.Trace.Length() > 0 {
				fmt.Fprintf(os.Stderr, "  Counterexample trace:\n")
				for j := 0; j < violation.Violation.Trace.Length(); j++ {
					state := violation.Violation.Trace.GetState(j)
					action := violation.Violation.Trace.GetAction(j)
					if action != "" {
						fmt.Fprintf(os.Stderr, "    %d. %s\n", j+1, action)
					}
					if verbose && state != nil {
						for name, value := range state.Variables {
							fmt.Fprintf(os.Stderr, "       %s = %s\n", name, value.String())
						}
					}
				}
			}
			if len(violation.Violation.Cycles) > 0 {
				fmt.Fprintf(os.Stderr, "  Blocking cycles: %d\n", len(violation.Violation.Cycles))
			}
		}
		
		return fmt.Errorf("verification failed with %d violation(s)", len(result.Violations)+len(temporalViolations))
	}

	fmt.Printf("✓ Verification passed for %s\n", filename)
	fmt.Printf("  Explored %d states\n", result.StatesExplored)
	if len(temporalProps) > 0 {
		fmt.Printf("  Verified %d temporal propert%s\n", len(temporalProps), func() string {
			if len(temporalProps) == 1 {
				return "y"
			}
			return "ies"
		}())
	}
	return nil
	})
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

