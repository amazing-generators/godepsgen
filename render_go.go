package godepsgen

import (
	"bytes"
	"compress/flate"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"go/format"
	"sort"
	"strings"
	"text/template"
)

// // // // // // // // // //

//go:embed render_go.tmpl
var goTemplateText string

// Parsed once at package init and reused for every render call.
var goTemplate = template.Must(template.New("dependencies-go").Funcs(template.FuncMap{
	"quote":    strconvQuote,
	"hexBytes": renderHexBytes,
}).Parse(goTemplateText))

// //

type packedLicenseObj struct {
	Hash  string
	Data  []byte
	Empty bool
}

type goEntryObj struct {
	Module  string
	Version string
	Hash    string
}

type goTemplateDataObj struct {
	PackageName string
	ModFile     string
	GeneratedAt string
	Licenses    []packedLicenseObj
	Entries     []goEntryObj
}

// //

func packLicense(licenseText string) (string, []byte, error) {
	hashArr := sha256.Sum256([]byte(licenseText))
	hashValue := hex.EncodeToString(hashArr[:])[:16]

	if licenseText == "" {
		return hashValue, nil, nil
	}

	var buffer bytes.Buffer
	writer, err := flate.NewWriter(&buffer, flate.BestCompression)
	if err != nil {
		return "", nil, err
	}

	if _, err = writer.Write([]byte(licenseText)); err != nil {
		_ = writer.Close()
		return "", nil, err
	}

	if err = writer.Close(); err != nil {
		return "", nil, err
	}

	return hashValue, buffer.Bytes(), nil
}

// Emits the []byte literal, wrapping every 24 values so big licenses stay readable.
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
	packedMap := make(map[string]packedLicenseObj)
	hashByModuleMap := make(map[string]string, len(report.Items))

	for _, item := range report.Items {
		hashValue, compressedData, err := packLicense(item.License)
		if err != nil {
			return nil, fmt.Errorf("pack license for %s: %w", item.Module, err)
		}

		hashByModuleMap[item.Module] = hashValue
		// Identical licenses (e.g. MIT across many modules) are embedded once, keyed by hash.
		if _, ok := packedMap[hashValue]; ok {
			continue
		}

		packedMap[hashValue] = packedLicenseObj{
			Hash:  hashValue,
			Data:  compressedData,
			Empty: item.License == "",
		}
	}

	// Sort by hash for a stable declaration order in the generated output.
	packedArr := make([]packedLicenseObj, 0, len(packedMap))
	for _, item := range packedMap {
		packedArr = append(packedArr, item)
	}
	sort.Slice(packedArr, func(leftIndex int, rightIndex int) bool {
		return packedArr[leftIndex].Hash < packedArr[rightIndex].Hash
	})

	entriesArr := make([]goEntryObj, 0, len(report.Items))
	for _, item := range report.Items {
		entriesArr = append(entriesArr, goEntryObj{
			Module:  item.Module,
			Version: item.Version,
			Hash:    hashByModuleMap[item.Module],
		})
	}

	data := goTemplateDataObj{
		PackageName: packageName,
		ModFile:     report.ModFile,
		GeneratedAt: report.GeneratedAt,
		Licenses:    packedArr,
		Entries:     entriesArr,
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
