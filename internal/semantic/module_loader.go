package semantic

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BoundlessProducts/spectre/internal/lexer"
	"github.com/BoundlessProducts/spectre/internal/parser"
	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// ModuleInfo contains information about a loaded module
type ModuleInfo struct {
	File     *ast.File
	FilePath string
	Module   *ast.ModuleDecl
	Loaded   bool
}

// ModuleLoader handles loading modules from files and resolving imports
type ModuleLoader struct {
	modules        map[string]*ModuleInfo // Module name -> ModuleInfo
	loadedFiles    map[string]bool        // Track loaded files to prevent reloading
	loadingFiles   map[string]bool        // Track files being loaded (for circular dependency detection)
	baseDir        string                  // Base directory for resolving relative paths
	errors         []string
}

// NewModuleLoader creates a new module loader
func NewModuleLoader(baseDir string) *ModuleLoader {
	if baseDir == "" {
		baseDir, _ = os.Getwd()
	}
	absBaseDir, _ := filepath.Abs(baseDir)
	return &ModuleLoader{
		modules:      make(map[string]*ModuleInfo),
		loadedFiles:  make(map[string]bool),
		loadingFiles: make(map[string]bool),
		baseDir:      absBaseDir,
		errors:       []string{},
	}
}

// LoadModule loads a module from a file path
func (ml *ModuleLoader) LoadModule(filePath string) (*ModuleInfo, []string) {
	ml.errors = []string{}
	
	// Convert to absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		ml.addError(0, 0, "invalid file path: %s: %v", filePath, err)
		return nil, ml.errors
	}

	// Check if already loaded
	if info, exists := ml.findModuleByPath(absPath); exists {
		return info, ml.errors
	}

	// Check for circular dependency
	if ml.loadingFiles[absPath] {
		ml.addError(0, 0, "circular dependency detected: file %s is already being loaded", absPath)
		return nil, ml.errors
	}

	// Mark as loading
	ml.loadingFiles[absPath] = true
	defer delete(ml.loadingFiles, absPath)

	// Read and parse file
	content, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			ml.addError(0, 0, "File not found - %s", absPath)
		} else {
			ml.addError(0, 0, "error reading file %s: %v", absPath, err)
		}
		return nil, ml.errors
	}

	// Parse file
	l := lexer.New(string(content))
	p := parser.New(l)
	file := p.ParseFile()

	if len(p.Errors()) > 0 {
		for _, errMsg := range p.Errors() {
			ml.addError(0, 0, "parse error in %s: %s", absPath, errMsg)
		}
		return nil, ml.errors
	}

	// Check that file has exactly one module
	modules := []*ast.ModuleDecl{}
	for _, decl := range file.Decls {
		if moduleDecl, ok := decl.(*ast.ModuleDecl); ok {
			modules = append(modules, moduleDecl)
		}
	}

	if len(modules) == 0 {
		ml.addError(0, 0, "file %s must contain exactly one module declaration, found 0", absPath)
		return nil, ml.errors
	}

	if len(modules) > 1 {
		ml.addError(0, 0, "file %s must contain exactly one module declaration, found %d", absPath, len(modules))
		return nil, ml.errors
	}

	module := modules[0]

	// Check that module name matches file name (without .spec extension)
	expectedName := strings.TrimSuffix(filepath.Base(absPath), ".spec")
	if module.Name != expectedName {
		ml.addError(module.Pos().Line, module.Pos().Column, "module name '%s' must match file name '%s'", module.Name, expectedName)
		return nil, ml.errors
	}

	// Mark file as loaded
	ml.loadedFiles[absPath] = true

	// Create module info
	info := &ModuleInfo{
		File:     file,
		FilePath: absPath,
		Module:   module,
		Loaded:   true,
	}

	// Store module by name
	ml.modules[module.Name] = info

	// Resolve imports
	importErrors := ml.resolveImports(file, absPath)
	ml.errors = append(ml.errors, importErrors...)

	return info, ml.errors
}

// resolveImports resolves all imports in a file and loads required modules
func (ml *ModuleLoader) resolveImports(file *ast.File, currentFilePath string) []string {
	var errors []string
	currentDir := filepath.Dir(currentFilePath)

	for _, decl := range file.Decls {
		importDecl, ok := decl.(*ast.ImportDecl)
		if !ok {
			continue
		}

		var targetPath string

		if importDecl.Path != "" {
			// Path-based import: import "path/to/module"
			if filepath.IsAbs(importDecl.Path) {
				targetPath = importDecl.Path
			} else {
				// Relative to current file's directory
				targetPath = filepath.Join(currentDir, importDecl.Path)
			}
			// Add .spec extension if not present
			if !strings.HasSuffix(targetPath, ".spec") {
				targetPath += ".spec"
			}
		} else if importDecl.Module != "" {
			// Module name import: import ModuleName (from same directory)
			targetPath = filepath.Join(currentDir, importDecl.Module+".spec")
		} else {
			errors = append(errors, fmt.Sprintf("%d:%d: import declaration missing both module name and path", 
				importDecl.Pos().Line, importDecl.Pos().Column))
			continue
		}

		// Load the module
		_, loadErrors := ml.LoadModule(targetPath)
		if len(loadErrors) > 0 {
			errors = append(errors, loadErrors...)
			continue
		}
	}

	return errors
}

// findModuleByPath finds a module by its file path
func (ml *ModuleLoader) findModuleByPath(filePath string) (*ModuleInfo, bool) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, false
	}

	for _, info := range ml.modules {
		if info.FilePath == absPath {
			return info, true
		}
	}

	return nil, false
}

// GetModule returns a module by name
func (ml *ModuleLoader) GetModule(name string) (*ModuleInfo, bool) {
	info, exists := ml.modules[name]
	return info, exists
}

// GetAllModules returns all loaded modules
func (ml *ModuleLoader) GetAllModules() map[string]*ModuleInfo {
	return ml.modules
}

// CheckCircularDependencies checks for circular dependencies in the loaded modules
func (ml *ModuleLoader) CheckCircularDependencies() []string {
	var errors []string
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for moduleName := range ml.modules {
		if !visited[moduleName] {
			cycle := ml.detectCycle(moduleName, visited, recursionStack, []string{})
			if len(cycle) > 0 {
				cyclePath := strings.Join(cycle, " -> ")
				errors = append(errors, fmt.Sprintf("circular dependency detected: %s", cyclePath))
			}
		}
	}

	return errors
}

// detectCycle detects cycles in module dependencies using DFS
func (ml *ModuleLoader) detectCycle(moduleName string, visited, recursionStack map[string]bool, path []string) []string {
	visited[moduleName] = true
	recursionStack[moduleName] = true
	path = append(path, moduleName)

	info, exists := ml.modules[moduleName]
	if !exists {
		return nil
	}

	// Check imports
	currentDir := filepath.Dir(info.FilePath)
	for _, decl := range info.File.Decls {
		importDecl, ok := decl.(*ast.ImportDecl)
		if !ok {
			continue
		}

		var importedModuleName string
		if importDecl.Path != "" {
			var targetPath string
			if filepath.IsAbs(importDecl.Path) {
				targetPath = importDecl.Path
			} else {
				targetPath = filepath.Join(currentDir, importDecl.Path)
			}
			if !strings.HasSuffix(targetPath, ".spec") {
				targetPath += ".spec"
			}
			absPath, _ := filepath.Abs(targetPath)
			// Find module by path
			for name, info := range ml.modules {
				if info.FilePath == absPath {
					importedModuleName = name
					break
				}
			}
		} else if importDecl.Module != "" {
			importedModuleName = importDecl.Module
		}

		if importedModuleName == "" {
			continue
		}

		// Check if this creates a cycle
		if recursionStack[importedModuleName] {
			// Found a cycle - return the cycle path
			cycleStart := -1
			for i, name := range path {
				if name == importedModuleName {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := append(path[cycleStart:], importedModuleName)
				return cycle
			}
		}

		if !visited[importedModuleName] {
			cycle := ml.detectCycle(importedModuleName, visited, recursionStack, path)
			if len(cycle) > 0 {
				return cycle
			}
		}
	}

	recursionStack[moduleName] = false
	return nil
}

// addError adds an error message
func (ml *ModuleLoader) addError(line, column int, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if line > 0 {
		ml.errors = append(ml.errors, fmt.Sprintf("%d:%d: %s", line, column, msg))
	} else {
		ml.errors = append(ml.errors, msg)
	}
}

