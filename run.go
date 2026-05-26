package godepsgen

import "fmt"

// // // // // // // // // //

func Run(config ConfigObj) (*ResultObj, error) {
	normalizedConfig, err := normalizeConfig(config)
	if err != nil {
		return nil, err
	}

	report, err := collectNormalized(*normalizedConfig)
	if err != nil {
		return nil, err
	}

	var dataArr []byte

	switch normalizedConfig.Format {
	case "go":
		dataArr, err = renderGo(report, normalizedConfig.PackageName)
	case "json":
		dataArr, err = renderJSON(report)
	default:
		err = fmt.Errorf("unsupported format: %s", normalizedConfig.Format)
	}
	if err != nil {
		return nil, err
	}

	if !normalizedConfig.Stdout {
		if err = writeFileAtomically(normalizedConfig.OutputFile, dataArr, normalizedConfig.Force); err != nil {
			return nil, err
		}
	}

	return &ResultObj{
		Report: report,
		Data:   dataArr,
	}, nil
}
