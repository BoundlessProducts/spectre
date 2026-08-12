package errors

import (
	"fmt"
	"strings"

	"github.com/BoundlessProducts/spectre/pkg/ast"
)

// ErrorFormatter formats errors with context and descriptions
type ErrorFormatter struct {
	includeDescriptions bool
	includeStackTraces   bool
}

// NewErrorFormatter creates a new error formatter
func NewErrorFormatter() *ErrorFormatter {
	return &ErrorFormatter{
		includeDescriptions: true,
		includeStackTraces:   false,
	}
}

// SetIncludeDescriptions sets whether to include descriptions in error messages
func (ef *ErrorFormatter) SetIncludeDescriptions(include bool) {
	ef.includeDescriptions = include
}

// SetIncludeStackTraces sets whether to include stack traces in error messages
func (ef *ErrorFormatter) SetIncludeStackTraces(include bool) {
	ef.includeStackTraces = include
}

// FormatError formats an error with context
func (ef *ErrorFormatter) FormatError(message string, context *ErrorContext) string {
	var parts []string

	// Add position
	if context.Position.Line > 0 {
		parts = append(parts, fmt.Sprintf("%s:", context.FormatPosition()))
	}

	// Add element information
	if context.ElementName != "" || context.ElementType != "" {
		elementInfo := context.FormatElement()
		parts = append(parts, elementInfo)
	}

	// Add description if available and enabled
	if ef.includeDescriptions && context.HasDescription() {
		parts = append(parts, fmt.Sprintf("(%s)", context.Description))
	}

	// Add error message
	parts = append(parts, message)

	return strings.Join(parts, " ")
}

// FormatValidationError formats a validation error with context
func (ef *ErrorFormatter) FormatValidationError(errorType string, name string, message string, context *ErrorContext) string {
	var parts []string

	// Add position
	if context.Position.Line > 0 {
		parts = append(parts, fmt.Sprintf("%s:", context.FormatPosition()))
	}

	// Add error type and element name
	if name != "" {
		parts = append(parts, fmt.Sprintf("%s '%s' violated:", errorType, name))
	} else {
		parts = append(parts, fmt.Sprintf("%s violated:", errorType))
	}

	// Add description if available
	if ef.includeDescriptions && context.HasDescription() {
		parts = append(parts, fmt.Sprintf("(%s)", context.Description))
	}

	// Add error message
	parts = append(parts, message)

	return strings.Join(parts, " ")
}

// FormatInvariantViolation formats an invariant violation error
func (ef *ErrorFormatter) FormatInvariantViolation(name string, message string, context *ErrorContext) string {
	return ef.FormatValidationError("Invariant", name, message, context)
}

// FormatPostconditionViolation formats a postcondition violation error
func (ef *ErrorFormatter) FormatPostconditionViolation(actionName string, name string, message string, context *ErrorContext) string {
	var parts []string

	// Add position
	if context.Position.Line > 0 {
		parts = append(parts, fmt.Sprintf("%s:", context.FormatPosition()))
	}

	// Add action and postcondition info
	if actionName != "" && name != "" {
		parts = append(parts, fmt.Sprintf("Postcondition '%s' of action '%s' violated:", name, actionName))
	} else if name != "" {
		parts = append(parts, fmt.Sprintf("Postcondition '%s' violated:", name))
	} else {
		parts = append(parts, "Postcondition violated:")
	}

	// Add description if available
	if ef.includeDescriptions && context.HasDescription() {
		parts = append(parts, fmt.Sprintf("(%s)", context.Description))
	}

	// Add error message
	parts = append(parts, message)

	return strings.Join(parts, " ")
}

// FormatPreconditionViolation formats a precondition violation error
func (ef *ErrorFormatter) FormatPreconditionViolation(actionName string, name string, message string, context *ErrorContext) string {
	var parts []string

	// Add position
	if context.Position.Line > 0 {
		parts = append(parts, fmt.Sprintf("%s:", context.FormatPosition()))
	}

	// Add action and precondition info
	if actionName != "" && name != "" {
		parts = append(parts, fmt.Sprintf("Precondition '%s' of action '%s' violated:", name, actionName))
	} else if name != "" {
		parts = append(parts, fmt.Sprintf("Precondition '%s' violated:", name))
	} else {
		parts = append(parts, "Precondition violated:")
	}

	// Add description if available
	if ef.includeDescriptions && context.HasDescription() {
		parts = append(parts, fmt.Sprintf("(%s)", context.Description))
	}

	// Add error message
	parts = append(parts, message)

	return strings.Join(parts, " ")
}

// FormatTemporalViolation formats a temporal property violation error
func (ef *ErrorFormatter) FormatTemporalViolation(name string, message string, context *ErrorContext) string {
	return ef.FormatValidationError("Temporal property", name, message, context)
}

// FormatTypeError formats a type error with context
func (ef *ErrorFormatter) FormatTypeError(message string, position ast.Position) string {
	var parts []string

	if position.Line > 0 {
		parts = append(parts, fmt.Sprintf("%d:%d:", position.Line, position.Column))
	}

	parts = append(parts, "Type error:", message)

	return strings.Join(parts, " ")
}

// FormatParseError formats a parse error
func (ef *ErrorFormatter) FormatParseError(message string, position ast.Position) string {
	var parts []string

	if position.Line > 0 {
		parts = append(parts, fmt.Sprintf("%d:%d:", position.Line, position.Column))
	}

	parts = append(parts, "Parse error:", message)

	return strings.Join(parts, " ")
}

