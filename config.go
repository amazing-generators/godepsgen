package godepsgen

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// // // // // // // // // //

const (
	cDefaultFormat      = "go"
	cDefaultPackageName = "dependenciesgen"
	cDefaultGoOutput    = "dependencies_gen.go"
	cDefaultJSONOutput  = "dependencies.json"
	cDefaultYAMLOutput  = "dependencies.yml"

	DefaultLicenseMaxBytes = int64(5 << 20)
)

// //

type ConfigObj struct {
	Source          string
	OutputFile      string
	PackageName     string
	Format          string
	Stdout          bool
	SkipLicenses    bool
	Force           bool
	LicenseMaxBytes int64
	ModuleCacheRoot string
	GeneratedAt     time.Time
}

type ItemObj struct {
	Module  string `json:"module"`
	Version string `json:"version"`
	License string `json:"license"`
}

type ReportObj struct {
	GeneratedAt string    `json:"generated_at"`
	SourceRoot  string    `json:"source_root"`
	ModFile     string    `json:"mod_file"`
	Items       []ItemObj `json:"items"`
}

type ResultObj struct {
	OutputFile string
	Report     *ReportObj
	Data       []byte
}

type normalizedConfigObj struct {
	SourceRoot      string
	ModFile         string
	OutputFile      string
	PackageName     string
	Format          string
	Stdout          bool
	SkipLicenses    bool
	Force           bool
	LicenseMaxBytes int64
	ModuleCacheRoot string
	GeneratedAt     time.Time
}

// //

func normalizeConfig(config ConfigObj) (*normalizedConfigObj, error) {
	format := strings.TrimSpace(config.Format)
	if format == "" {
		format = cDefaultFormat
	}

	if format == "yml" {
		format = "yaml"
	}

	switch format {
	case "go", "json", "yaml":
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	licenseMaxBytes := config.LicenseMaxBytes
	if licenseMaxBytes <= 0 {
		licenseMaxBytes = DefaultLicenseMaxBytes
	}

	generatedAt := config.GeneratedAt
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	}

	sourceRoot, modFile, err := resolveSourcePaths(config.Source)
	if err != nil {
		return nil, err
	}

	outputFile, err := resolveOutputFile(config.OutputFile, format, config.Stdout)
	if err != nil {
		return nil, err
	}

	packageName := strings.TrimSpace(config.PackageName)
	if format == "go" {
		if packageName == "" && outputFile != "" {
			packageName = sanitizePackageName(filepath.Base(filepath.Dir(outputFile)))
		}
		if packageName == "" {
			packageName = cDefaultPackageName
		}
	}

	moduleCacheRoot := strings.TrimSpace(config.ModuleCacheRoot)
	if moduleCacheRoot != "" {
		moduleCacheRoot, err = filepath.Abs(moduleCacheRoot)
		if err != nil {
			return nil, fmt.Errorf("resolve module cache root: %w", err)
		}
	}

	if format != "go" && packageName != "" {
		packageName = ""
	}

	return &normalizedConfigObj{
		SourceRoot:      sourceRoot,
		ModFile:         modFile,
		OutputFile:      outputFile,
		PackageName:     packageName,
		Format:          format,
		Stdout:          config.Stdout,
		SkipLicenses:    config.SkipLicenses,
		Force:           config.Force,
		LicenseMaxBytes: licenseMaxBytes,
		ModuleCacheRoot: moduleCacheRoot,
		GeneratedAt:     generatedAt,
	}, nil
}

// Пустой путь означает запись в текущую директорию с именем по умолчанию.
// Существующая директория трактуется как каталог назначения, а не как файл.
func resolveOutputFile(outputValue string, format string, stdout bool) (string, error) {
	outputValue = strings.TrimSpace(outputValue)
	if outputValue == "" && stdout {
		return "", nil
	}

	defaultFileName := defaultOutputFileName(format)
	if outputValue == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("read working directory: %w", err)
		}
		return filepath.Join(cwd, defaultFileName), nil
	}

	isDirHint := strings.HasSuffix(outputValue, string(os.PathSeparator)) || filepath.Ext(outputValue) == ""

	outputPath, err := filepath.Abs(outputValue)
	if err != nil {
		return "", fmt.Errorf("resolve output path: %w", err)
	}

	info, err := os.Stat(outputPath)
	if err == nil {
		if info.IsDir() {
			return filepath.Join(outputPath, defaultFileName), nil
		}
		return outputPath, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat output path: %w", err)
	}

	if isDirHint {
		return filepath.Join(outputPath, defaultFileName), nil
	}

	return outputPath, nil
}

func defaultOutputFileName(format string) string {
	switch format {
	case "json":
		return cDefaultJSONOutput
	case "yaml":
		return cDefaultYAMLOutput
	default:
		return cDefaultGoOutput
	}
}

func resolveSourcePaths(source string) (string, string, error) {
	if strings.TrimSpace(source) != "" {
		return resolveSourceValue(source)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("read working directory: %w", err)
	}

	return cwd, filepath.Join(cwd, "go.mod"), nil
}

func resolveSourceValue(source string) (string, string, error) {
	sourcePath, err := filepath.Abs(source)
	if err != nil {
		return "", "", fmt.Errorf("resolve source path: %w", err)
	}

	info, err := os.Stat(sourcePath)
	if err != nil {
		return "", "", fmt.Errorf("stat source path: %w", err)
	}

	if info.IsDir() {
		return sourcePath, filepath.Join(sourcePath, "go.mod"), nil
	}

	if filepath.Base(sourcePath) != "go.mod" {
		return "", "", errors.New("source file must be go.mod")
	}

	return filepath.Dir(sourcePath), sourcePath, nil
}

func sanitizePackageName(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return cDefaultPackageName
	}

	var out strings.Builder

	for index, symbol := range raw {
		switch {
		case symbol >= 'a' && symbol <= 'z':
			out.WriteRune(symbol)
		case symbol >= 'A' && symbol <= 'Z':
			out.WriteRune(symbol + ('a' - 'A'))
		case symbol >= '0' && symbol <= '9':
			if index == 0 {
				out.WriteByte('p')
			}
			out.WriteRune(symbol)
		default:
			out.WriteByte('_')
		}
	}

	packageName := out.String()
	if packageName == "" {
		return cDefaultPackageName
	}

	return packageName
}
