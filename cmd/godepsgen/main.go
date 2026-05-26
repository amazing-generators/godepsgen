package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/amazing-generators/godepsgen"
)

// // // // // // // // // //

func main() {
	config := godepsgen.ConfigObj{}

	flag.StringVar(&config.Source, "source", "", "source root or go.mod path; default is current working directory")
	flag.StringVar(&config.OutputFile, "out", "", "output file path; required unless stdout is enabled")
	flag.StringVar(&config.PackageName, "pkg", "", "package name for generated Go file")
	flag.StringVar(&config.Format, "format", "go", "output format: go or json")
	flag.BoolVar(&config.Stdout, "stdout", false, "write output to stdout instead of a file")
	flag.BoolVar(&config.SkipLicenses, "skip-licenses", false, "generate versions only with empty licenses")
	flag.BoolVar(&config.Force, "force", false, "create missing output directories; without it a missing directory is an error")
	flag.Int64Var(&config.LicenseMaxBytes, "license-max-bytes", godepsgen.DefaultLicenseMaxBytes, "maximum allowed license file size")
	flag.StringVar(&config.ModuleCacheRoot, "mod-cache", "", "module cache root; defaults to GOMODCACHE, GOPATH/pkg/mod or ~/go/pkg/mod")
	flag.Parse()

	result, err := godepsgen.Run(config)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	if config.Stdout {
		if _, err = os.Stdout.Write(result.Data); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		return
	}

	_, _ = fmt.Fprintln(os.Stdout, "Generated:", config.OutputFile)
}
