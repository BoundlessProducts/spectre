# Specification Inconsistencies Found

**Status:** ✅ All issues have been fixed!

## Critical Issues

### 1. Weak Fairness Description Contradiction ✅ FIXED
**Location:** SPEC.md lines 753-761

**Issue:** 
- Line 753 states: "Weak fairness states that if an action is **continuously enabled**, it will eventually execute"
- Line 761 states: "This means: if `increment` is **enabled infinitely often**, it will execute infinitely often"

**Problem:** "Continuously enabled" and "enabled infinitely often" are different concepts:
- **Continuously enabled**: Action is enabled at every step from some point onward
- **Enabled infinitely often**: Action becomes enabled infinitely many times (but may be disabled in between)

**Fix:** ✅ Fixed - Line 761 now correctly says: "This means: if `increment` is **continuously enabled** (enabled at every step from some point onward), it will execute infinitely often"

### 2. Missing `super` Keyword Documentation ✅ FIXED
**Location:** SPEC.md - Keywords section

**Issue:** `super.increment()` is used in module examples (line 289) but `super` is not listed in the keywords section.

**Fix:** ✅ Fixed - `super` has been added to keywords list and documented in the Modules section with explanation of how to call parent module's actions.

### 3. Module Instance Syntax Not Fully Explained ✅ FIXED
**Location:** SPEC.md lines 243-251

**Issue:** The syntax `module MyCounter = Counter with { counter = myCounter }` is shown but:
- Not clear what `with` clause does
- Not clear if this creates a new module or an instance
- Not clear how to use the instance

**Fix:** ✅ Fixed - Added detailed explanation of module instantiation, including what the `with` clause does and how to use instances.

### 4. Constants Visibility in Modules Not Documented ✅ FIXED
**Location:** SPEC.md Modules section, examples/modules-example.spec line 62

**Issue:** Example shows `BoundedCounter.MAX_VALUE` being accessed from outside, but it's not documented:
- Can constants be accessed from outside modules?
- Do they need to be `public`?
- What's the access syntax?

**Fix:** ✅ Fixed - Documented constant visibility rules in modules, including public/private modifiers and access syntax (`ModuleName.CONSTANT_NAME`).

## Medium Priority Issues

### 5. Module Extension Behavior Unclear ✅ FIXED
**Location:** SPEC.md lines 230-241

**Issue:** When extending a module:
- Can you override actions? (Example shows it, but not explained)
- What happens to state variables from parent?
- Can you add new state variables?
- How does `super` work?

**Fix:** ✅ Fixed - Added detailed "Extension Rules" section explaining inheritance, overriding, super calls, and adding new members.

### 6. Fairness on Variables Not Fully Explained ✅ FIXED
**Location:** SPEC.md lines 791-805

**Issue:** Fairness on variables (`WF(counter)`) is mentioned but:
- How does it determine which actions modify the variable?
- What if multiple actions modify the variable?
- Is it WF/SF on all actions or any action?

**Fix:** ✅ Fixed - Added "How it works" section explaining that fairness applies to all actions that modify the variable, and how it determines which actions qualify.

### 7. Module Import Scope Not Clear ✅ FIXED
**Location:** SPEC.md lines 221-228

**Issue:** When you `import Counter`:
- What exactly becomes available?
- Do you need to qualify names (e.g., `Counter.increment`)?
- Can you import multiple modules with name conflicts?

**Fix:** ✅ Fixed - Added "Import Rules" section explaining qualified access (`ModuleName.memberName`), name conflicts, public members only, and scope.

### 8. Constants in Pure Functions Not Mentioned ✅ FIXED
**Location:** SPEC.md Constants section, Pure Functions section

**Issue:** Constants section says constants can be used in pure functions, but Pure Functions section doesn't mention constants.

**Fix:** ✅ Fixed - Added constants to the "Pure functions can only use" list in the Pure Functions section.

## Minor Issues

### 9. Inconsistent Example Formatting ✅ FIXED
**Location:** Various examples in SPEC.md

**Issue:** Some examples have descriptions, some don't. Examples in the spec should be consistent.

**Fix:** ✅ Fixed - Added descriptions to all 5 examples in SPEC.md (Simple Counter, User Management, Mutex Lock, Message Queue, Bank Account) for consistency.

### 10. Missing Cross-References ✅ FIXED
**Location:** Throughout SPEC.md

**Issue:** Some sections reference concepts from other sections without links:
- Modules section mentions "types, constants" but no links
- Constants section mentions "parameterized specifications" but doesn't link to modules
- Fairness section mentions "liveness properties" but doesn't link to temporal properties

**Fix:** ✅ Fixed - Added markdown links:
- Modules section now links to types, constants, state variables, actions, invariants, and temporal properties
- Constants section links to Modules section
- Fairness section links to Temporal Properties section

### 11. `oneOf` Syntax Inconsistency ✅ FIXED
**Location:** SPEC.md lines 432-468

**Issue:** Shows three different syntaxes for `oneOf`:
1. Single variable: `init oneOf { counter = 0, counter = 5 }`
2. Multiple variables tuple: `init oneOf { { counter = 0, status = ... }, ... }`
3. Multiple variables block: `init oneOf { { counter = 0 ... }, ... }`

**Problem:** Not clear if all three are valid or which is preferred.

**Fix:** ✅ Fixed - Clarified that all three syntaxes are valid, labeled them clearly (Single Variable Syntax, Multiple Variables - Tuple Syntax, Multiple Variables - Block Syntax), and provided guidance on when to use each.

### 12. Temporal Property Syntax Inconsistency ✅ FIXED
**Location:** SPEC.md Temporal Properties section

**Issue:** Some temporal properties use `→` (arrow), some use `→` in different contexts. Need to clarify:
- Is `→` the same as `leads to`?
- Can it be used outside temporal properties?

**Fix:** ✅ Fixed - Added clarification in "Leads To" section explaining:
- `→` is the "leads to" operator
- It's syntactic sugar for `always (P → eventually Q)`
- It can only be used within temporal properties
- Also updated the Temporal Operators summary to clarify this

## Summary

**Critical:** 4 issues ✅ ALL FIXED
**Medium:** 4 issues ✅ ALL FIXED
**Minor:** 4 issues ✅ ALL FIXED

**Total:** 12 inconsistencies found ✅ ALL FIXED

**Status:** All identified inconsistencies have been resolved. The specification is now consistent and complete.

### What Was Fixed:

1. ✅ Fairness definitions (weak vs strong) - Corrected terminology
2. ✅ Module system documentation (super, instances, visibility) - Added comprehensive documentation
3. ✅ Constants visibility in modules - Documented access rules
4. ✅ Example consistency - Added descriptions to all examples
5. ✅ Cross-references - Added markdown links between related sections
6. ✅ Syntax clarifications - Clarified `oneOf` and `→` operator usage

