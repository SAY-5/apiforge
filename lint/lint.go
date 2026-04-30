// Package lint walks a parsed spec and emits Findings against a
// configurable rule set. Each rule is a function (Spec, Operation) →
// []Finding so users can add custom rules without touching the
// engine.
package lint

import (
	"regexp"
	"strings"

	"github.com/SAY-5/apiforge/spec"
)

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

type Finding struct {
	Rule     string
	Severity Severity
	Path     string
	Method   string
	Message  string
}

type Rule func(s *spec.Spec) []Finding

// Default rule set covers the violations most teams agree on.
// Production config adds custom rules + per-rule severity overrides.
func DefaultRules() []Rule {
	return []Rule{
		RulePathLowercaseKebab(),
		RuleHTTPSuccessResponseRequired(),
		RuleOperationIDPresent(),
		RulePathParamsDeclared(),
		RuleHTTPMethodWordingMatchesPath(),
	}
}

func Run(s *spec.Spec, rules []Rule) []Finding {
	out := []Finding{}
	for _, r := range rules {
		out = append(out, r(s)...)
	}
	return out
}

// RulePathLowercaseKebab: paths must be lowercase-with-dashes (no
// snake_case, no PascalCase, no camelCase). Many style guides
// require this; the linter enforces it consistently.
func RulePathLowercaseKebab() Rule {
	bad := regexp.MustCompile(`[A-Z_]`)
	return func(s *spec.Spec) []Finding {
		out := []Finding{}
		for path := range s.Paths {
			// Strip path-parameter braces before checking.
			cleaned := regexp.MustCompile(`\{[^}]+\}`).ReplaceAllString(path, "")
			if bad.MatchString(cleaned) {
				out = append(out, Finding{
					Rule:     "path-lowercase-kebab",
					Severity: SeverityError,
					Path:     path,
					Message:  "path must be lowercase-with-dashes",
				})
			}
		}
		return out
	}
}

// RuleHTTPSuccessResponseRequired: every operation must declare at
// least one 2xx response. Catches operations that only document
// failures.
func RuleHTTPSuccessResponseRequired() Rule {
	return func(s *spec.Spec) []Finding {
		out := []Finding{}
		for _, ref := range s.Operations() {
			has := false
			for code := range ref.Op.Responses {
				if strings.HasPrefix(code, "2") {
					has = true
					break
				}
			}
			if !has {
				out = append(out, Finding{
					Rule:     "success-response-required",
					Severity: SeverityError,
					Path:     ref.Path,
					Method:   ref.Method,
					Message:  "operation declares no 2xx response",
				})
			}
		}
		return out
	}
}

// RuleOperationIDPresent: machine-readable operationId is required
// for code generators (mock server, SDK).
func RuleOperationIDPresent() Rule {
	return func(s *spec.Spec) []Finding {
		out := []Finding{}
		for _, ref := range s.Operations() {
			if ref.Op.OperationID == "" {
				out = append(out, Finding{
					Rule:     "operation-id-present",
					Severity: SeverityWarning,
					Path:     ref.Path,
					Method:   ref.Method,
					Message:  "operationId missing",
				})
			}
		}
		return out
	}
}

// RulePathParamsDeclared: every {param} in a path must have a
// corresponding `in: path` parameter declared on the operation.
func RulePathParamsDeclared() Rule {
	param := regexp.MustCompile(`\{([^}]+)\}`)
	return func(s *spec.Spec) []Finding {
		out := []Finding{}
		for _, ref := range s.Operations() {
			declared := map[string]bool{}
			for _, p := range ref.Op.Parameters {
				if p.In == "path" {
					declared[p.Name] = true
				}
			}
			for _, m := range param.FindAllStringSubmatch(ref.Path, -1) {
				if !declared[m[1]] {
					out = append(out, Finding{
						Rule:     "path-params-declared",
						Severity: SeverityError,
						Path:     ref.Path,
						Method:   ref.Method,
						Message:  "path parameter '" + m[1] + "' not declared",
					})
				}
			}
		}
		return out
	}
}

// RuleHTTPMethodWordingMatchesPath: GET on /things should not have
// 'create' / 'delete' / 'update' in the operationId (lexical
// mismatch is a strong smell).
func RuleHTTPMethodWordingMatchesPath() Rule {
	bad := map[string][]string{
		"GET":    {"create", "delete", "update"},
		"POST":   {"delete"},
		"DELETE": {"create"},
	}
	return func(s *spec.Spec) []Finding {
		out := []Finding{}
		for _, ref := range s.Operations() {
			id := strings.ToLower(ref.Op.OperationID)
			for _, word := range bad[ref.Method] {
				if strings.Contains(id, word) {
					out = append(out, Finding{
						Rule:     "method-wording-matches-path",
						Severity: SeverityWarning,
						Path:     ref.Path,
						Method:   ref.Method,
						Message:  "operationId contains '" + word + "' but method is " + ref.Method,
					})
				}
			}
		}
		return out
	}
}
