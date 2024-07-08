// screamingsnakeconst detects constant variables that are not declared in
// SCREAMING_SNAKE_CASE. Constant variables can be declared either at a File
// level or at a Contract level as state variables.
package screamingsnakeconst

import (
	"regexp"
	"solbot/ast"
	"solbot/reporter"
	"solbot/token"
)

const (
	title          = "Variables declared as `constant` should be in `SCREAMING_SNAKE_CASE`"
	severity       = "Best Practices"
	descTempl      = "Constant variables should be declared with a `SCREAMING_SNAKE_CASE`. The following variables don't follow this practice: {{ range .Locations }}\n- `{{ .Context }}`{{ end }}"
	recommendation = "Consider renaming the variables to make the code more readable and less error-prone."
)

type Detector struct{}

func (*Detector) Detect(node ast.Node) *reporter.Finding {
	finding := reporter.Finding{}
	matches := 0
	switch n := node.(type) {
	case *ast.File:
		for _, decl := range n.Declarations {
			if v, ok := decl.(*ast.StateVariableDeclaration); ok {
				if v == nil {
					continue
				}
				// @TODO: Add immutable variables as well
				if v.Mutability == ast.Constant {
					if !isScreamingSnakeCase(v.Name.Name) {
						finding.Locations = append(
							finding.Locations, reporter.Location{
								Position: token.Position{
									Offset: v.Name.NamePos,
								},
								Context: v.Name.Name, // Save name for report
							})
						matches++
					}
				}
			}
		}

		if matches > 0 {
			// Add the rest of the fields to the finding
			finding.Title = title
			finding.Severity = severity
			finding.Description = reporter.GenerateCustomDescription(descTempl, finding.Locations)
			finding.Recommendation = recommendation
			return &finding
		} else {
			return nil
		}

	default:
		return nil
	}
}

func isScreamingSnakeCase(s string) bool {
	// Regular expression to match SCREAMING_SNAKE_CASE:
	// ^ and $ are anchors to say that the whole string must match the pattern.
	// Without ancors something like "SNAKE_case" would match as a substring.
	// [A-Z0-9_] matches any uppercase letter, number or underscore.
	// The quantifier + means that this pattern needs to appear at least once.
	regex := regexp.MustCompile(`^[A-Z0-9_]+$`)
	return regex.MatchString(s)
}
