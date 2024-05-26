package reporter

import (
	"solparsor/token"
)

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
