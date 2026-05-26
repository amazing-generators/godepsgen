package godepsgen

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"
)

// // // // // // // // // //

//go:embed render_yaml.tmpl
var yamlTemplateText string

var yamlTemplate = template.Must(template.New("dependencies-yaml").Funcs(template.FuncMap{
	"yamlInline": yamlInline,
	"yamlValue":  yamlValue,
}).Parse(yamlTemplateText))

// //

func yamlInline(value string) string {
	if value == "" {
		return `""`
	}

	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func yamlValue(value string, indent int) string {
	if value == "" {
		return `""`
	}

	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")

	if !strings.Contains(value, "\n") {
		return yamlInline(value)
	}

	prefix := strings.Repeat(" ", indent)
	linesArr := strings.Split(value, "\n")

	var builder strings.Builder
	builder.WriteString("|-\n")

	for index, line := range linesArr {
		if index > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(prefix)
		builder.WriteString(line)
	}

	return builder.String()
}

// //

func renderYAML(report *ReportObj) ([]byte, error) {
	compactReport := buildCompactReport(report)

	var buffer bytes.Buffer
	if err := yamlTemplate.Execute(&buffer, compactReport); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
