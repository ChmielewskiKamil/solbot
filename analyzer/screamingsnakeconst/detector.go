// screamingsnakeconst detects constant variables that are not declared in
// SCREAMING_SNAKE_CASE. Constant variables can be declared either at a File
// level or at a Contract level as state variables.
package screamingsnakeconst

import (
	"regexp"
	"solparsor/ast"
	"solparsor/token"
)

const (
	title          = "Variables declared as `constant` should be in SCREAMING_SNAKE_CASE"
	severity       = "Best Practices"
	description    = "Description of the finding goes here."
	recommendation = "Recommendation goes here."
)

func Detect(node ast.Node) *Finding {
	finding := Finding{}
	matches := 0
	switch n := node.(type) {
	case *ast.File:
		for _, decl := range n.Declarations {
			if v, ok := decl.(*ast.VariableDeclaration); ok {
				if v.Constant {
					if !isScreamingSnakeCase(v.Name.Name) {
						finding.Locations = append(finding.Locations, Location{
							File:    "filename",
							Line:    decl.Start(),
							Context: "context",
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
			finding.Description = description
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
	println(s)
	return regex.MatchString(s)
}

type Finding struct {
	Title          string
	Severity       string
	Description    string
	Recommendation string
	Locations      []Location
}

type Location struct {
	File    string         // The filename where the issue was found.
	Line    token.Position // The line number where the issue was found.
	Context string         // The line with the issue itself or with its surroundings.
}
