package godepsgen

import "encoding/json"

// // // // // // // // // //

func renderJSON(report *ReportObj) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}
