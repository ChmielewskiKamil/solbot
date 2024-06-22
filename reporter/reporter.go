package reporter

import (
	"fmt"
	"os"
	"solbot/token"
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

func (f *Finding) CalculatePositions(src string, fileName string) {
	for i := range f.Locations {
		// @TODO: This resets reader's state, but it is inefficient.
		reader := strings.NewReader(src)
		token.OffsetToPosition(reader, &f.Locations[i].Position)
		// @TODO: Add file name to the position.
		f.Locations[i].Position.Filename = fileName
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
