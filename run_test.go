package godepsgen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// // // // // // // // // //

// //

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

	var report ReportObj
	if err = json.Unmarshal(result.Data, &report); err != nil {
		t.Fatal(err)
	}

	if len(report.Items) != 1 {
		t.Fatalf("unexpected items count: %d", len(report.Items))
	}

	if report.Items[0].Module != "example.com/dep" {
		t.Fatalf("unexpected module: %s", report.Items[0].Module)
	}

	if report.Items[0].Version != "v1.2.3" {
		t.Fatalf("unexpected version: %s", report.Items[0].Version)
	}

	if report.Items[0].License != "test-license" {
		t.Fatalf("unexpected license: %s", report.Items[0].License)
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

	var report ReportObj
	if err = json.Unmarshal(result.Data, &report); err != nil {
		t.Fatal(err)
	}

	if len(report.Items) != 1 {
		t.Fatalf("unexpected items count: %d", len(report.Items))
	}

	if report.Items[0].License != "" {
		t.Fatalf("expected empty license, got: %q", report.Items[0].License)
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

	var report ReportObj
	if err = json.Unmarshal(result.Data, &report); err != nil {
		t.Fatal(err)
	}

	if len(report.Items) != 1 {
		t.Fatalf("unexpected items count: %d", len(report.Items))
	}

	if report.Items[0].License != "license-text" {
		t.Fatalf("expected LICENSE to win by priority, got: %q", report.Items[0].License)
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

	if !strings.Contains(outputText, "example.com/dep") {
		t.Fatalf("generated file does not contain module path: %s", outputText)
	}

	if !strings.Contains(outputText, "isEmpty: true") {
		t.Fatalf("generated file does not contain empty license marker: %s", outputText)
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

	// Without force a missing output directory must fail.
	if _, err := Run(baseConfig); err == nil {
		t.Fatal("expected error when output directory is missing without force")
	}

	// With force the directory tree is created and the file is written.
	forceConfig := baseConfig
	forceConfig.Force = true
	if _, err := Run(forceConfig); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(outputFile); err != nil {
		t.Fatalf("expected generated file to exist: %v", err)
	}

	// Directory already exists: overwrite is allowed even without force.
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
