package godepsgen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// // // // // // // // // //

type compactJSONItemObj struct {
	Module  string `json:"module"`
	Version string `json:"version"`
	License string `json:"license"`
}

type compactJSONReportObj struct {
	GeneratedAt string               `json:"generated_at"`
	SourceRoot  string               `json:"source_root"`
	ModFile     string               `json:"mod_file"`
	Modules     map[string]string    `json:"modules"`
	Versions    map[string]string    `json:"versions"`
	Licenses    map[string]string    `json:"licenses"`
	Items       []compactJSONItemObj `json:"items"`
}

// //

func decodeCompactJSON(t *testing.T, dataArr []byte) compactJSONReportObj {
	t.Helper()

	var report compactJSONReportObj
	if err := json.Unmarshal(dataArr, &report); err != nil {
		t.Fatal(err)
	}

	return report
}

func resolveCompactItem(
	t *testing.T,
	report compactJSONReportObj,
	index int,
) ItemObj {
	t.Helper()

	if index < 0 || index >= len(report.Items) {
		t.Fatalf("item index out of range: %d", index)
	}

	item := report.Items[index]

	moduleValue, ok := report.Modules[item.Module]
	if !ok {
		t.Fatalf("missing module key: %s", item.Module)
	}

	versionValue, ok := report.Versions[item.Version]
	if !ok {
		t.Fatalf("missing version key: %s", item.Version)
	}

	licenseValue, ok := report.Licenses[item.License]
	if !ok {
		t.Fatalf("missing license key: %s", item.License)
	}

	return ItemObj{
		Module:  moduleValue,
		Version: versionValue,
		License: licenseValue,
	}
}

func TestRunJSONWithLocalReplace(t *testing.T) {
	rootDir := t.TempDir()
	depDir := filepath.Join(rootDir, "localdep")

	if err := os.MkdirAll(depDir, 0755); err != nil {
		t.Fatal(err)
	}

	modText := "module example.com/root\n\ngo 1.19\n\nrequire example.com/dep v1.2.3\n\nreplace example.com/dep => ./localdep\n"
	if err := os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte(modText), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(depDir, "LICENSE"), []byte("test-license"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Run(ConfigObj{
		Source: rootDir,
		Format: "json",
		Stdout: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	report := decodeCompactJSON(t, result.Data)

	if len(report.Items) != 1 {
		t.Fatalf("unexpected items count: %d", len(report.Items))
	}

	item := resolveCompactItem(t, report, 0)

	if item.Module != "example.com/dep" {
		t.Fatalf("unexpected module: %s", item.Module)
	}

	if item.Version != "v1.2.3" {
		t.Fatalf("unexpected version: %s", item.Version)
	}

	if item.License != "test-license" {
		t.Fatalf("unexpected license: %s", item.License)
	}
}

func TestRunJSONWithoutLicenseInNonEmptyModule(t *testing.T) {
	rootDir := t.TempDir()
	depDir := filepath.Join(rootDir, "localdep")

	if err := os.MkdirAll(depDir, 0755); err != nil {
		t.Fatal(err)
	}

	modText := "module example.com/root\n\ngo 1.19\n\nrequire example.com/dep v1.2.3\n\nreplace example.com/dep => ./localdep\n"
	if err := os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte(modText), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(depDir, "README.md"), []byte("no license here"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Run(ConfigObj{
		Source: rootDir,
		Format: "json",
		Stdout: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	report := decodeCompactJSON(t, result.Data)

	if len(report.Items) != 1 {
		t.Fatalf("unexpected items count: %d", len(report.Items))
	}

	item := resolveCompactItem(t, report, 0)

	if item.License != "" {
		t.Fatalf("expected empty license, got: %q", item.License)
	}
}

func TestRunJSONUsesLicensePriority(t *testing.T) {
	rootDir := t.TempDir()
	depDir := filepath.Join(rootDir, "localdep")

	if err := os.MkdirAll(depDir, 0755); err != nil {
		t.Fatal(err)
	}

	modText := "module example.com/root\n\ngo 1.19\n\nrequire example.com/dep v1.2.3\n\nreplace example.com/dep => ./localdep\n"
	if err := os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte(modText), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(depDir, "COPYING"), []byte("copying-license"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(depDir, "LICENSE"), []byte("license-text"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Run(ConfigObj{
		Source: rootDir,
		Format: "json",
		Stdout: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	report := decodeCompactJSON(t, result.Data)

	if len(report.Items) != 1 {
		t.Fatalf("unexpected items count: %d", len(report.Items))
	}

	item := resolveCompactItem(t, report, 0)

	if item.License != "license-text" {
		t.Fatalf("expected LICENSE to win by priority, got: %q", item.License)
	}
}

func TestRunJSONFailsOnEmptyModuleDir(t *testing.T) {
	rootDir := t.TempDir()
	depDir := filepath.Join(rootDir, "localdep")

	if err := os.MkdirAll(depDir, 0755); err != nil {
		t.Fatal(err)
	}

	modText := "module example.com/root\n\ngo 1.19\n\nrequire example.com/dep v1.2.3\n\nreplace example.com/dep => ./localdep\n"
	if err := os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte(modText), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Run(ConfigObj{
		Source: rootDir,
		Format: "json",
		Stdout: true,
	})
	if err == nil {
		t.Fatal("expected error for empty module directory")
	}

	if !strings.Contains(err.Error(), "module cache is incomplete") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveSourcePathsFromDir(t *testing.T) {
	rootDir := t.TempDir()

	sourceRoot, modFile, err := resolveSourcePaths(rootDir)
	if err != nil {
		t.Fatal(err)
	}

	if sourceRoot != rootDir {
		t.Fatalf("unexpected source root: %s", sourceRoot)
	}

	expectedModFile := filepath.Join(rootDir, "go.mod")
	if modFile != expectedModFile {
		t.Fatalf("unexpected mod file: %s", modFile)
	}
}

func TestResolveSourcePathsFromModFile(t *testing.T) {
	rootDir := t.TempDir()
	modFilePath := filepath.Join(rootDir, "go.mod")

	if err := os.WriteFile(modFilePath, []byte("module example.com/root\n"), 0644); err != nil {
		t.Fatal(err)
	}

	sourceRoot, modFile, err := resolveSourcePaths(modFilePath)
	if err != nil {
		t.Fatal(err)
	}

	if sourceRoot != rootDir {
		t.Fatalf("unexpected source root: %s", sourceRoot)
	}

	if modFile != modFilePath {
		t.Fatalf("unexpected mod file: %s", modFile)
	}
}

func TestResolveSourcePathsFromWorkingDir(t *testing.T) {
	rootDir := t.TempDir()

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err = os.Chdir(rootDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	sourceRoot, modFile, err := resolveSourcePaths("")
	if err != nil {
		t.Fatal(err)
	}

	if sourceRoot != rootDir {
		t.Fatalf("unexpected source root: %s", sourceRoot)
	}

	expectedModFile := filepath.Join(rootDir, "go.mod")
	if modFile != expectedModFile {
		t.Fatalf("unexpected mod file: %s", modFile)
	}
}

func TestRunGoSkipLicenses(t *testing.T) {
	rootDir := t.TempDir()

	modText := "module example.com/root\n\ngo 1.19\n\nrequire example.com/dep v1.2.3\n"
	if err := os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte(modText), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Run(ConfigObj{
		Source:       rootDir,
		Format:       "go",
		PackageName:  "sample",
		Stdout:       true,
		SkipLicenses: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	outputText := string(result.Data)
	if !strings.Contains(outputText, "package sample") {
		t.Fatalf("generated file has wrong package: %s", outputText)
	}

	if !strings.Contains(outputText, "cModule0") || !strings.Contains(outputText, "\"example.com/dep\"") {
		t.Fatalf("generated file does not contain module constant: %s", outputText)
	}

	if !strings.Contains(outputText, "cVersion0") || !strings.Contains(outputText, "\"v1.2.3\"") {
		t.Fatalf("generated file does not contain version constant: %s", outputText)
	}

	if !strings.Contains(outputText, "type LicenseObj struct {\n\tcompressedData []byte\n\tisEmpty        bool\n}") {
		t.Fatalf("generated file does not contain full LicenseObj layout: %s", outputText)
	}

	if !strings.Contains(outputText, "license0 = LicenseObj{isEmpty: true}") {
		t.Fatalf("generated file does not contain empty license marker: %s", outputText)
	}

	if !strings.Contains(outputText, "cModule0: cVersion0") {
		t.Fatalf("generated file does not reuse constants in VersionByModule: %s", outputText)
	}

	if strings.Contains(outputText, "\"bytes\"") || strings.Contains(outputText, "\"compress/flate\"") || strings.Contains(outputText, "\"io\"") {
		t.Fatalf("generated file unexpectedly imports compression packages: %s", outputText)
	}

	if !strings.Contains(outputText, "func (item *LicenseObj) String() string {\n\treturn \"\"\n}") {
		t.Fatalf("generated file does not contain simplified String method: %s", outputText)
	}
}

func TestRunForceCreatesOutputDir(t *testing.T) {
	rootDir := t.TempDir()

	modText := "module example.com/root\n\ngo 1.19\n\nrequire example.com/dep v1.2.3\n"
	if err := os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte(modText), 0644); err != nil {
		t.Fatal(err)
	}

	outputFile := filepath.Join(rootDir, "nested", "sub", "gen.go")

	baseConfig := ConfigObj{
		Source:       rootDir,
		Format:       "go",
		PackageName:  "sample",
		OutputFile:   outputFile,
		SkipLicenses: true,
	}

	// Без force отсутствие директории должно приводить к ошибке.
	if _, err := Run(baseConfig); err == nil {
		t.Fatal("expected error when output directory is missing without force")
	}

	// С force дерево директорий создаётся автоматически.
	forceConfig := baseConfig
	forceConfig.Force = true
	if _, err := Run(forceConfig); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(outputFile); err != nil {
		t.Fatalf("expected generated file to exist: %v", err)
	}

	// Если директория уже есть, force для перезаписи не нужен.
	overwriteConfig := baseConfig
	overwriteConfig.PackageName = "sample2"
	if _, err := Run(overwriteConfig); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "package sample2") {
		t.Fatalf("expected overwritten package, got: %s", string(data))
	}
}

func TestRunUsesWorkingDirWhenOutputIsOmitted(t *testing.T) {
	sourceDir := t.TempDir()
	workDir := t.TempDir()

	modText := "module example.com/root\n\ngo 1.19\n\nrequire example.com/dep v1.2.3\n"
	if err := os.WriteFile(filepath.Join(sourceDir, "go.mod"), []byte(modText), 0644); err != nil {
		t.Fatal(err)
	}

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err = os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	result, err := Run(ConfigObj{
		Source:       sourceDir,
		Format:       "go",
		SkipLicenses: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	expectedOutputFile := filepath.Join(workDir, "dependencies_gen.go")
	if result.OutputFile != expectedOutputFile {
		t.Fatalf("unexpected output file: %s", result.OutputFile)
	}

	data, err := os.ReadFile(expectedOutputFile)
	if err != nil {
		t.Fatal(err)
	}

	expectedPackageName := sanitizePackageName(filepath.Base(workDir))
	if !strings.Contains(string(data), "package "+expectedPackageName) {
		t.Fatalf("unexpected generated package: %s", string(data))
	}
}

func TestRunUsesDirectoryOutputPath(t *testing.T) {
	rootDir := t.TempDir()
	outputDir := filepath.Join(rootDir, "nested")

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatal(err)
	}

	modText := "module example.com/root\n\ngo 1.19\n\nrequire example.com/dep v1.2.3\n"
	if err := os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte(modText), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Run(ConfigObj{
		Source:       rootDir,
		Format:       "go",
		OutputFile:   outputDir,
		SkipLicenses: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	expectedOutputFile := filepath.Join(outputDir, "dependencies_gen.go")
	if result.OutputFile != expectedOutputFile {
		t.Fatalf("unexpected output file: %s", result.OutputFile)
	}

	data, err := os.ReadFile(expectedOutputFile)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "package nested") {
		t.Fatalf("unexpected generated package: %s", string(data))
	}
}

func TestRunForceCreatesDirectoryOutputPath(t *testing.T) {
	rootDir := t.TempDir()
	outputDir := filepath.Join(rootDir, "nested", "sub")

	modText := "module example.com/root\n\ngo 1.19\n\nrequire example.com/dep v1.2.3\n"
	if err := os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte(modText), 0644); err != nil {
		t.Fatal(err)
	}

	baseConfig := ConfigObj{
		Source:       rootDir,
		Format:       "go",
		OutputFile:   outputDir,
		SkipLicenses: true,
	}

	if _, err := Run(baseConfig); err == nil {
		t.Fatal("expected error when output directory is missing without force")
	}

	forceConfig := baseConfig
	forceConfig.Force = true
	result, err := Run(forceConfig)
	if err != nil {
		t.Fatal(err)
	}

	expectedOutputFile := filepath.Join(outputDir, "dependencies_gen.go")
	if result.OutputFile != expectedOutputFile {
		t.Fatalf("unexpected output file: %s", result.OutputFile)
	}

	if _, err = os.Stat(expectedOutputFile); err != nil {
		t.Fatalf("expected generated file to exist: %v", err)
	}
}

func TestRunYMLUsesAliases(t *testing.T) {
	rootDir := t.TempDir()
	depDir := filepath.Join(rootDir, "localdep")

	if err := os.MkdirAll(depDir, 0755); err != nil {
		t.Fatal(err)
	}

	modText := "module example.com/root\n\ngo 1.19\n\nrequire example.com/dep v1.2.3\n\nreplace example.com/dep => ./localdep\n"
	if err := os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte(modText), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(depDir, "LICENSE"), []byte("test-license"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Run(ConfigObj{
		Source: rootDir,
		Format: "yml",
		Stdout: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	outputText := string(result.Data)

	if !strings.Contains(outputText, "&module_m0") {
		t.Fatalf("yaml does not contain module anchor: %s", outputText)
	}

	if !strings.Contains(outputText, "*module_m0") {
		t.Fatalf("yaml does not contain module alias: %s", outputText)
	}

	if !strings.Contains(outputText, "&license_l0") {
		t.Fatalf("yaml does not contain license anchor: %s", outputText)
	}

	if !strings.Contains(outputText, "*license_l0") {
		t.Fatalf("yaml does not contain license alias: %s", outputText)
	}
}

func TestRunYAMLUsesStableIndentation(t *testing.T) {
	rootDir := t.TempDir()
	depDir := filepath.Join(rootDir, "localdep")

	if err := os.MkdirAll(depDir, 0755); err != nil {
		t.Fatal(err)
	}

	modText := "module example.com/root\n\ngo 1.19\n\nrequire example.com/dep v1.2.3\n\nreplace example.com/dep => ./localdep\n"
	if err := os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte(modText), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(depDir, "LICENSE"), []byte("line1\nline2"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Run(ConfigObj{
		Source: rootDir,
		Format: "yaml",
		Stdout: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	outputText := string(result.Data)

	if strings.Contains(outputText, "\n    modules:\n") {
		t.Fatalf("yaml has invalid top-level indentation: %s", outputText)
	}

	if !strings.Contains(outputText, "\nmodules:\n  m0: &module_m0 'example.com/dep'\n") {
		t.Fatalf("yaml modules block has wrong indentation: %s", outputText)
	}

	if !strings.Contains(outputText, "\nlicenses:\n  l0: &license_l0 |-\n    line1\n    line2\n") {
		t.Fatalf("yaml license block has wrong indentation: %s", outputText)
	}

	if !strings.Contains(outputText, "\nitems:\n  - module: *module_m0\n    version: *version_v0\n    license: *license_l0") {
		t.Fatalf("yaml items block has wrong indentation: %s", outputText)
	}
}
