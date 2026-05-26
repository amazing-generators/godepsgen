package godepsgen

import (
	"bytes"
	"compress/flate"
	_ "embed"
	"fmt"
	"go/format"
	"strings"
	"text/template"
)

// // // // // // // // // //

//go:embed render_go.tmpl
var goTemplateText string

var goTemplate = template.Must(template.New("dependencies-go").Funcs(template.FuncMap{
	"quote":    strconvQuote,
	"hexBytes": renderHexBytes,
}).Parse(goTemplateText))

// //

type goStringConstObj struct {
	Name  string
	Value string
}

type goEntryObj struct {
	ModuleConst  string
	VersionConst string
	LicenseVar   string
}

type goLicenseObj struct {
	Name  string
	Data  []byte
	Empty bool
}

type goTemplateDataObj struct {
	PackageName         string
	ModFile             string
	GeneratedAt         string
	Modules             []goStringConstObj
	Versions            []goStringConstObj
	Licenses            []goLicenseObj
	Entries             []goEntryObj
	HasEmbeddedLicenses bool
}

// //

func compressLicense(licenseText string) ([]byte, error) {
	if licenseText == "" {
		return nil, nil
	}

	var buffer bytes.Buffer
	writer, err := flate.NewWriter(&buffer, flate.BestCompression)
	if err != nil {
		return nil, err
	}

	if _, err = writer.Write([]byte(licenseText)); err != nil {
		_ = writer.Close()
		return nil, err
	}

	if err = writer.Close(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func renderHexBytes(dataArr []byte) string {
	if len(dataArr) == 0 {
		return ""
	}

	var builder strings.Builder

	for index, item := range dataArr {
		if index > 0 {
			builder.WriteString(", ")
		}
		if index > 0 && index%24 == 0 {
			builder.WriteString("\n\t\t\t")
		}
		builder.WriteString(fmt.Sprintf("0x%02x", item))
	}

	return builder.String()
}

func strconvQuote(value string) string {
	return fmt.Sprintf("%q", value)
}

// //

func renderGo(report *ReportObj, packageName string) ([]byte, error) {
	compactReport := buildCompactReport(report)

	modulesArr := make([]goStringConstObj, 0, len(compactReport.Modules))
	versionsArr := make([]goStringConstObj, 0, len(compactReport.Versions))
	licensesArr := make([]goLicenseObj, 0, len(compactReport.Licenses))
	entriesArr := make([]goEntryObj, 0, len(compactReport.Items))

	moduleConstByKeyMap := make(map[string]string, len(compactReport.Modules))
	versionConstByKeyMap := make(map[string]string, len(compactReport.Versions))
	licenseVarByKeyMap := make(map[string]string, len(compactReport.Licenses))
	hasEmbeddedLicensesFlag := false

	for index, item := range compactReport.Modules {
		constName := fmt.Sprintf("cModule%d", index)
		moduleConstByKeyMap[item.Key] = constName

		modulesArr = append(modulesArr, goStringConstObj{
			Name:  constName,
			Value: item.Value,
		})
	}

	for index, item := range compactReport.Versions {
		constName := fmt.Sprintf("cVersion%d", index)
		versionConstByKeyMap[item.Key] = constName

		versionsArr = append(versionsArr, goStringConstObj{
			Name:  constName,
			Value: item.Value,
		})
	}

	for index, item := range compactReport.Licenses {
		compressedData, err := compressLicense(item.Value)
		if err != nil {
			return nil, fmt.Errorf("pack license %s: %w", item.Key, err)
		}
		if item.Value != "" {
			hasEmbeddedLicensesFlag = true
		}

		varName := fmt.Sprintf("license%d", index)
		licenseVarByKeyMap[item.Key] = varName

		licensesArr = append(licensesArr, goLicenseObj{
			Name:  varName,
			Data:  compressedData,
			Empty: item.Value == "",
		})
	}

	for _, item := range compactReport.Items {
		entriesArr = append(entriesArr, goEntryObj{
			ModuleConst:  moduleConstByKeyMap[item.ModuleKey],
			VersionConst: versionConstByKeyMap[item.VersionKey],
			LicenseVar:   licenseVarByKeyMap[item.LicenseKey],
		})
	}

	data := goTemplateDataObj{
		PackageName:         packageName,
		ModFile:             report.ModFile,
		GeneratedAt:         report.GeneratedAt,
		Modules:             modulesArr,
		Versions:            versionsArr,
		Licenses:            licensesArr,
		Entries:             entriesArr,
		HasEmbeddedLicenses: hasEmbeddedLicensesFlag,
	}

	var buffer bytes.Buffer
	if err := goTemplate.Execute(&buffer, data); err != nil {
		return nil, fmt.Errorf("execute go template: %w", err)
	}

	formattedData, err := format.Source(buffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated go: %w", err)
	}

	return formattedData, nil
}
