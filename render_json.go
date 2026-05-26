package godepsgen

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"text/template"
)

// // // // // // // // // //

//go:embed render_json.tmpl
var jsonTemplateText string

var jsonTemplate = template.Must(template.New("dependencies-json").Funcs(template.FuncMap{
	"jsonQuote": jsonQuote,
}).Parse(jsonTemplateText))

// //

func jsonQuote(value string) string {
	dataArr, _ := json.Marshal(value)
	return string(dataArr)
}

// //

func renderJSON(report *ReportObj) ([]byte, error) {
	compactReport := buildCompactReport(report)

	var buffer bytes.Buffer
	if err := jsonTemplate.Execute(&buffer, compactReport); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
