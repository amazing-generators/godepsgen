package godepsgen

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"golang.org/x/mod/modfile"
)

// // // // // // // // // //

type replaceObj struct {
	OldPath    string
	OldVersion string
	NewPath    string
	NewVersion string
}

type sourceModuleObj struct {
	Path    string
	Version string
}

type locatedModuleObj struct {
	Item ItemObj
}

// //

func collectNormalized(config normalizedConfigObj) (*ReportObj, error) {
	modData, err := os.ReadFile(config.ModFile)
	if err != nil {
		return nil, fmt.Errorf("read mod file: %w", err)
	}

	parsedFile, err := modfile.Parse(config.ModFile, modData, nil)
	if err != nil {
		return nil, fmt.Errorf("parse mod file: %w", err)
	}

	modulesArr := make([]sourceModuleObj, 0, len(parsedFile.Require))
	for _, item := range parsedFile.Require {
		modulesArr = append(modulesArr, sourceModuleObj{
			Path:    item.Mod.Path,
			Version: item.Mod.Version,
		})
	}

	sort.Slice(modulesArr, func(leftIndex int, rightIndex int) bool {
		return modulesArr[leftIndex].Path < modulesArr[rightIndex].Path
	})

	replaceMap := buildReplaceMap(parsedFile)

	locatedArr, err := locateModules(modulesArr, replaceMap, config)
	if err != nil {
		return nil, err
	}

	itemsArr := make([]ItemObj, 0, len(locatedArr))
	for _, item := range locatedArr {
		itemsArr = append(itemsArr, item.Item)
	}

	return &ReportObj{
		GeneratedAt: config.GeneratedAt.Format(timeLayout()),
		SourceRoot:  config.SourceRoot,
		ModFile:     config.ModFile,
		Items:       itemsArr,
	}, nil
}

func buildReplaceMap(parsedFile *modfile.File) map[string]replaceObj {
	replaceMap := make(map[string]replaceObj, len(parsedFile.Replace)*2)

	// Index each replace under both "path" and "path@version" so version-pinned
	// and bare replace directives can be matched without extra lookups later.
	for _, item := range parsedFile.Replace {
		entry := replaceObj{
			OldPath:    item.Old.Path,
			OldVersion: item.Old.Version,
			NewPath:    item.New.Path,
			NewVersion: item.New.Version,
		}

		replaceMap[item.Old.Path] = entry
		if item.Old.Version != "" {
			replaceMap[item.Old.Path+"@"+item.Old.Version] = entry
		}
	}

	return replaceMap
}

func locateModules(
	modulesArr []sourceModuleObj,
	replaceMap map[string]replaceObj,
	config normalizedConfigObj,
) ([]locatedModuleObj, error) {
	if config.SkipLicenses {
		locatedArr := make([]locatedModuleObj, 0, len(modulesArr))
		for _, item := range modulesArr {
			locatedArr = append(locatedArr, locatedModuleObj{
				Item: ItemObj{
					Module:  item.Path,
					Version: item.Version,
					License: "",
				},
			})
		}
		return locatedArr, nil
	}

	moduleCacheRoot, err := resolveModuleCacheRoot(config.ModuleCacheRoot)
	if err != nil {
		return nil, err
	}

	locatedArr := make([]locatedModuleObj, 0, len(modulesArr))
	missingArr := make([]string, 0)

	for _, item := range modulesArr {
		dir, err := resolveModuleDir(item.Path, item.Version, replaceMap, config.SourceRoot, moduleCacheRoot)
		if err != nil {
			return nil, err
		}
		if dir == "" {
			missingArr = append(missingArr, item.Path+"@"+item.Version)
			continue
		}

		nonEmptyFlag, err := isNonEmptyDir(dir)
		if err != nil {
			return nil, fmt.Errorf("inspect module dir for %s@%s: %w", item.Path, item.Version, err)
		}
		if !nonEmptyFlag {
			missingArr = append(missingArr, item.Path+"@"+item.Version)
			continue
		}

		licenseText, err := readLicenseText(dir, config.LicenseMaxBytes)
		if err != nil {
			return nil, fmt.Errorf("read license for %s@%s: %w", item.Path, item.Version, err)
		}

		locatedArr = append(locatedArr, locatedModuleObj{
			Item: ItemObj{
				Module:  item.Path,
				Version: item.Version,
				License: licenseText,
			},
		})
	}

	// Fail loudly instead of silently emitting empty licenses: the cache must be
	// warmed (go mod download) so the generated output is complete and reproducible.
	if len(missingArr) > 0 {
		return nil, fmt.Errorf(
			"module cache is incomplete, missing %d modules: %s",
			len(missingArr),
			strings.Join(missingArr, ", "),
		)
	}

	return locatedArr, nil
}

func timeLayout() string {
	return "2006-01-02T15:04:05Z07:00"
}
