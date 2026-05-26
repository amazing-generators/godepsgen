package godepsgen

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// // // // // // // // // //

// ensureOutputDir prepares the parent directory of the output file.
// In force mode the whole tree is created; otherwise a missing directory is an
// error, while writing into an existing one (overwriting a present file) is allowed.
func ensureOutputDir(outputDir string, force bool) error {
	if force {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}
		return nil
	}

	info, err := os.Stat(outputDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("output directory does not exist: %s (use -force to create it)", outputDir)
		}
		return fmt.Errorf("stat output directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("output parent is not a directory: %s", outputDir)
	}

	return nil
}

// //

func writeFileAtomically(outputFile string, dataArr []byte, force bool) error {
	if outputFile == "" {
		return fmt.Errorf("output path is empty")
	}

	// Skip the rewrite when content is unchanged to avoid touching build caches.
	if existingData, err := os.ReadFile(outputFile); err == nil {
		if bytes.Equal(existingData, dataArr) {
			return nil
		}
	}

	outputDir := filepath.Dir(outputFile)
	if err := ensureOutputDir(outputDir, force); err != nil {
		return err
	}

	tempFile, err := os.CreateTemp(outputDir, ".dependencies-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	tempPath := tempFile.Name()
	removeTempFlag := true

	defer func() {
		_ = tempFile.Close()
		if removeTempFlag {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err = tempFile.Write(dataArr); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err = tempFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	// Atomic replace: rename over any existing file in the same directory.
	if err = os.Rename(tempPath, outputFile); err != nil {
		return fmt.Errorf("replace output file: %w", err)
	}

	removeTempFlag = false
	return nil
}
