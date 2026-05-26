package godepsgen

import "fmt"

// // // // // // // // // //

type stringValueObj struct {
	Key   string
	Value string
}

type itemRefObj struct {
	ModuleKey  string
	VersionKey string
	LicenseKey string
}

type compactReportObj struct {
	GeneratedAt string
	SourceRoot  string
	ModFile     string
	Modules     []stringValueObj
	Versions    []stringValueObj
	Licenses    []stringValueObj
	Items       []itemRefObj
}

// //

func buildCompactReport(report *ReportObj) *compactReportObj {
	modulesArr := make([]stringValueObj, 0, len(report.Items))
	versionsArr := make([]stringValueObj, 0, len(report.Items))
	licensesArr := make([]stringValueObj, 0, len(report.Items))
	itemsArr := make([]itemRefObj, 0, len(report.Items))

	moduleKeyByValueMap := make(map[string]string, len(report.Items))
	versionKeyByValueMap := make(map[string]string, len(report.Items))
	licenseKeyByValueMap := make(map[string]string, len(report.Items))

	for _, item := range report.Items {
		moduleKey := ensureStringKey(item.Module, "m", moduleKeyByValueMap, &modulesArr)
		versionKey := ensureStringKey(item.Version, "v", versionKeyByValueMap, &versionsArr)
		licenseKey := ensureStringKey(item.License, "l", licenseKeyByValueMap, &licensesArr)

		itemsArr = append(itemsArr, itemRefObj{
			ModuleKey:  moduleKey,
			VersionKey: versionKey,
			LicenseKey: licenseKey,
		})
	}

	return &compactReportObj{
		GeneratedAt: report.GeneratedAt,
		SourceRoot:  report.SourceRoot,
		ModFile:     report.ModFile,
		Modules:     modulesArr,
		Versions:    versionsArr,
		Licenses:    licensesArr,
		Items:       itemsArr,
	}
}

func ensureStringKey(
	value string,
	prefix string,
	keyByValueMap map[string]string,
	itemsArr *[]stringValueObj,
) string {
	if keyValue, ok := keyByValueMap[value]; ok {
		return keyValue
	}

	keyValue := fmt.Sprintf("%s%d", prefix, len(*itemsArr))
	keyByValueMap[value] = keyValue

	*itemsArr = append(*itemsArr, stringValueObj{
		Key:   keyValue,
		Value: value,
	})

	return keyValue
}
