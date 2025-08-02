package reporter

import (
	"bytes"
	"fmt"
	"github.com/ChmielewskiKamil/solbot/token"
	"os"
	"strings"
	"text/template"
)

type Finding struct {
	Title          string
	Severity       string
	Description    string
	Recommendation string
	Locations      []Location
}

type Location struct {
	Position token.Position // Position data of the finding e.g. file, line, column.
	Context  string         // The line with the issue itself or with its surroundings.
}

func (f *Finding) CalculatePositions(file *token.SourceFile) {
	for i := range f.Locations {
		// TODO: This resets reader's state, but it is inefficient.
		reader := strings.NewReader(file.Content())
		token.OffsetToPosition(reader, &f.Locations[i].Position)
		// TODO: Add file name to the position.
		f.Locations[i].Position.Filename = file.Name()
	}
}

const (
	templatePath = "reporter/report_template.md"
)

func GenerateReport(findings []Finding, outputPath string) error {
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("Could not read the report template file: %s", err)
	}

	tmpl, err := template.New("report").Parse(string(templateData))
	if err != nil {
		return fmt.Errorf("Could not parse the report template: %s", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("Could not create the report file: %s", err)
	}
	defer file.Close()

	for _, finding := range findings {
		err := tmpl.Execute(file, finding)
		if err != nil {
			return fmt.Errorf(
				"Could not write the finding in the report file: %s", err)
		}

		_, err = file.WriteString("\n\n")
		if err != nil {
			return fmt.Errorf(
				"Could not add a newline in the report file: %s", err)
		}
	}

	return nil
}

// GenerateCustomDescription allows you to create a custom description template for the finding.
// Each detector can have its own go's text/template to suit its needs.
func GenerateCustomDescription(templPath string, locations []Location) string {
	tmpl, err := template.New("description").Parse(templPath)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]interface{}{
		"Locations": locations,
	})
	if err != nil {
		panic(err)
	}

	return buf.String()
}
