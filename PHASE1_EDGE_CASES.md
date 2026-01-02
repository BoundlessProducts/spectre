# Phase 1 Edge Case Failures

These are the edge cases where unit tests are failing. Note that the **integration test passes** - all real example files tokenize correctly. These failures appear to be related to how identifiers/numbers are read when they appear at the end of input or in specific contexts.

## 1. Identifier Truncation Issues

### Test: `TestNextTokenBasic`
**Input:** `var counter action increment`
**Issue:** Last identifier "increment" is read as "incremen" (missing last character 't')
- Expected: `IDENT "increment"`
- Got: `IDENT "incremen"`

### Test: `TestNextTokenSingleLineComment`
**Input:** 
```
var counter // This is a comment
action increment
```
**Issue:** Last identifier "increment" is read as "incremen"
- Expected: `IDENT "increment"`
- Got: `IDENT "incremen"`

### Test: `TestNextTokenMultiLineComment`
**Input:**
```
var counter /* This is a
multi-line comment */ action increment
```
**Issue:** Last identifier "increment" is read as "incremen"
- Expected: `IDENT "increment"`
- Got: `IDENT "incremen"`

### Test: `TestNextTokenMultipleComments`
**Input:**
```
// First comment
var counter // Second comment
/* Third comment */
action increment
```
**Issue:** Last identifier "increment" is read as "incremen"
- Expected: `IDENT "increment"`
- Got: `IDENT "incremen"`

### Test: `TestNextTokenIdentifiers`
**Input:** `counter myVariable user123 _private`
**Issue:** Last identifier "_private" is read as "_privat" (missing last character 'e')
- Expected: `IDENT "_private"`
- Got: `IDENT "_privat"`

## 2. Number Reading Issues

### Test: `TestNextTokenInteger`
**Input:** `0 123 456789`
**Issue:** Last number "456789" is read as "45678" (missing last digit '9')
- Expected: `INT "456789"`
- Got: `INT "45678"`

### Test: `TestNextTokenFloat`
**Input:** `0.0 123.456 0.5 10.0`
**Issue:** Last float "10.0" is read as "10." (missing trailing zero)
- Expected: `FLOAT "10.0"`
- Got: `FLOAT "10."`

## 3. Keyword Recognition Issues

### Test: `TestNextTokenKeywords`
**Input:** `var const action fun invariant temporal module import extends`
**Issue:** Last keyword "extends" is recognized as `IDENT` instead of `EXTENDS`
- Expected: `EXTENDS "extends"`
- Got: `IDENT "extends"`

### Test: `TestNextTokenWhitespace`
**Input:**
```
var   counter
	
	action
```
**Issue:** "action" is recognized as `IDENT` instead of `ACTION`
- Expected: `ACTION "action"`
- Got: `IDENT "action"`

### Test: `TestNextTokenBoolean`
**Input:** `true false`
**Issue:** "false" is recognized as `IDENT` instead of `BOOL`
- Expected: `BOOL "false"`
- Got: `IDENT "false"`

## Pattern Analysis

All failures follow a pattern:
1. **Last token in input**: All failures occur with the last token before EOF
2. **Truncation**: Identifiers and numbers are missing their last character
3. **Keyword recognition**: Keywords at the end are not recognized as keywords

This suggests the issue is in how we handle EOF - we're likely stopping one character early when reading identifiers/numbers, or not properly checking for EOF before returning.

## Impact Assessment

**Critical Test Status:** ✅ **PASSING**
- `TestLexerWithExampleFiles` passes completely
- All 11 real example files tokenize correctly
- 3,610+ tokens processed with 0 illegal tokens

## Fix Applied

**Status:** ✅ **ALL ISSUES FIXED**

The root cause was in `readIdentifier()` and `readNumber()` functions. When reading the last character before EOF:
1. The functions would check if `readPosition >= len(input)` before calling `readChar()`
2. If EOF was detected, they would set `l.position = l.readPosition` to include the current character
3. They also set `l.ch = 0` to properly signal EOF for the next token

**Changes made:**
- Updated `readIdentifier()` to check for EOF before advancing and include the current character
- Updated `readNumber()` to check for EOF in both integer and float reading loops
- Both functions now properly set `l.ch = 0` when EOF is detected

**Test Results:** ✅ **ALL TESTS PASSING**
- All unit tests pass
- All integration tests pass
- All edge cases resolved

