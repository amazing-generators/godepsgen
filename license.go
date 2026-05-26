package godepsgen

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/module"
)

// // // // // // // // // //

var licenseCandidatesArr = []string{
	"LICENSE",
	"LICENSE.txt",
	"LICENSE.md",
	"LICENSE.rst",
	"COPYING",
	"COPYING.txt",
	"COPYING.md",
	"COPYING.rst",
	"NOTICE",
	"NOTICE.txt",
	"NOTICE.md",
	"NOTICE.rst",
}

// //

// Resolves the module cache root following Go's own precedence:
// explicit override, then GOMODCACHE, then GOPATH/pkg/mod, then ~/go/pkg/mod.
func resolveModuleCacheRoot(moduleCacheRoot string) (string, error) {
	if moduleCacheRoot != "" {
		return moduleCacheRoot, nil
	}

	if value := strings.TrimSpace(os.Getenv("GOMODCACHE")); value != "" {
		return filepath.Abs(value)
	}

	gopathValue := strings.TrimSpace(os.Getenv("GOPATH"))
	if gopathValue != "" {
		for _, item := range filepath.SplitList(gopathValue) {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			return filepath.Abs(filepath.Join(item, "pkg", "mod"))
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve module cache root: %w", err)
	}

	return filepath.Join(homeDir, "go", "pkg", "mod"), nil
}

func resolveModuleDir(
	modulePath string,
	moduleVersion string,
	replaceMap map[string]replaceObj,
	sourceRoot string,
	moduleCacheRoot string,
) (string, error) {
	if replaceEntry, ok := replaceMap[modulePath+"@"+moduleVersion]; ok {
		return resolveReplaceDir(replaceEntry, sourceRoot, moduleCacheRoot)
	}

	if replaceEntry, ok := replaceMap[modulePath]; ok {
		return resolveReplaceDir(replaceEntry, sourceRoot, moduleCacheRoot)
	}

	return resolveCacheDir(modulePath, moduleVersion, moduleCacheRoot)
}

func resolveReplaceDir(
	replaceEntry replaceObj,
	sourceRoot string,
	moduleCacheRoot string,
) (string, error) {
	// No version means a local-path replace (=> ./dir), resolved against the source
	// root; otherwise it points at another module version inside the cache.
	if replaceEntry.NewVersion == "" {
		dir := replaceEntry.NewPath
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(sourceRoot, dir)
		}

		info, err := os.Stat(dir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return "", nil
			}
			return "", err
		}
		if !info.IsDir() {
			return "", fmt.Errorf("replace target is not a directory: %s", dir)
		}
		return filepath.Abs(dir)
	}

	return resolveCacheDir(replaceEntry.NewPath, replaceEntry.NewVersion, moduleCacheRoot)
}

// Escapes uppercase letters to "!lower" form, matching the on-disk cache layout
// (e.g. github.com/Azure -> github.com/!azure).
func resolveCacheDir(modulePath string, moduleVersion string, moduleCacheRoot string) (string, error) {
	escapedPath, err := module.EscapePath(modulePath)
	if err != nil {
		return "", fmt.Errorf("escape module path %s: %w", modulePath, err)
	}

	escapedVersion, err := module.EscapeVersion(moduleVersion)
	if err != nil {
		return "", fmt.Errorf("escape module version %s: %w", moduleVersion, err)
	}

	dir := filepath.Join(moduleCacheRoot, escapedPath+"@"+escapedVersion)
	info, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	if !info.IsDir() {
		return "", nil
	}

	return dir, nil
}

func readLicenseText(dir string, licenseMaxBytes int64) (string, error) {
	entriesArr, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	entriesByNameMap := make(map[string]os.DirEntry, len(entriesArr))
	for _, entry := range entriesArr {
		if entry.IsDir() {
			continue
		}
		entriesByNameMap[strings.ToLower(entry.Name())] = entry
	}

	for _, candidate := range licenseCandidatesArr {
		entry, ok := entriesByNameMap[strings.ToLower(candidate)]
		if !ok {
			continue
		}

		entryName := entry.Name()
		info, err := entry.Info()
		if err != nil {
			return "", err
		}
		if info.Size() > licenseMaxBytes {
			return "", fmt.Errorf(
				"license file exceeds limit: %s (%d > %d bytes)",
				filepath.Join(dir, entryName),
				info.Size(),
				licenseMaxBytes,
			)
		}

		data, err := os.ReadFile(filepath.Join(dir, entryName))
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(string(data)), nil
	}

	return "", nil
}

func isNonEmptyDir(dir string) (bool, error) {
	entriesArr, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	return len(entriesArr) > 0, nil
}
