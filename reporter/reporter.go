package reporter

import (
	"fmt"
	"os"
	"solparsor/token"
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
	File    string         // The filename where the issue was found.
	Line    token.Position // The line number where the issue was found.
	Context string         // The line with the issue itself or with its surroundings.
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

		_, err = file.WriteString("\n")
		if err != nil {
			return fmt.Errorf(
				"Could not add a newline in the report file: %s", err)
		}
	}

	return nil
}
