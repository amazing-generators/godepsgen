# godepsgen

`godepsgen` generates a dependency report from a Go module's `go.mod`.

It reads the module's `require` directives, keeps the output stable, resolves `replace` directives, optionally reads
license files from local replacements or the Go module cache, and emits one of three formats:

- generated Go source
- compact JSON
- compact YAML

The tool does not run `go list`, `go mod download`, or network requests. If license collection is enabled, the required
modules must already exist in the local module cache unless they are replaced by local paths.

Examples of generated output are available in [examples](/mnt/w541-data/shared/GolandProjects/dependencies/examples):

- [examples/with_licenses/dependencies_gen.go](/mnt/w541-data/shared/GolandProjects/dependencies/examples/with_licenses/dependencies_gen.go)
- [examples/with_licenses/dependencies.json](/mnt/w541-data/shared/GolandProjects/dependencies/examples/with_licenses/dependencies.json)
- [examples/with_licenses/dependencies.yml](/mnt/w541-data/shared/GolandProjects/dependencies/examples/with_licenses/dependencies.yml)
- [examples/without_licenses/dependencies_gen.go](/mnt/w541-data/shared/GolandProjects/dependencies/examples/without_licenses/dependencies_gen.go)
- [examples/without_licenses/dependencies.json](/mnt/w541-data/shared/GolandProjects/dependencies/examples/without_licenses/dependencies.json)
- [examples/without_licenses/dependencies.yml](/mnt/w541-data/shared/GolandProjects/dependencies/examples/without_licenses/dependencies.yml)

## What It Reads

`godepsgen` parses only the target `go.mod` file.

It includes:

- modules listed in `require`
- `replace` directives, both bare and version-specific
- local path replacements relative to the source module root
- cache-based replacements such as `replace x/y => x/y v1.2.3`

It does not:

- resolve transitive dependencies beyond what is explicitly present in `require`
- download missing modules
- inspect source code imports
- infer SPDX identifiers or normalize license metadata

## Output Formats

### Go

`-format go` generates a self-contained Go file that exposes:

- `List []EntryObj`
- `VersionByModule map[string]string`
- `LicenseByModule map[string]*LicenseObj`
- `(*LicenseObj).String() string`

License texts are compressed in the generated file and decompressed on demand by `(*LicenseObj).String()`.

If `-pkg` is omitted, the package name is derived from the output directory name. If that still cannot produce a usable
package name, `dependenciesgen` is used.

### JSON

`-format json` produces a compact deduplicated report with these top-level fields:

- `generated_at`
- `source_root`
- `mod_file`
- `modules`
- `versions`
- `licenses`
- `items`

Each `items[]` entry stores keys into the shared dictionaries:

- `module`
- `version`
- `license`

### YAML

`-format yaml` and `-format yml` produce the same logical data as JSON, but reuse repeated values through YAML anchors
and aliases.

## License Resolution

When license collection is enabled, `godepsgen` searches only the module root directory and picks the first matching
file from this priority list:

- `LICENSE`
- `LICENSE.txt`
- `LICENSE.md`
- `LICENSE.rst`
- `COPYING`
- `COPYING.txt`
- `COPYING.md`
- `COPYING.rst`
- `NOTICE`
- `NOTICE.txt`
- `NOTICE.md`
- `NOTICE.rst`

If no known license file is found, the license field is emitted as an empty string.

If a module directory is missing or empty, generation fails with an error instead of silently producing incomplete data.

## Module Cache Behavior

For non-replaced modules, the tool reads module contents from the local Go module cache.

Cache root resolution order:

1. `-mod-cache`
2. `GOMODCACHE`
3. first entry from `GOPATH`, with `/pkg/mod`
4. `~/go/pkg/mod`

If you want license collection to succeed reliably, warm the cache first:

```bash
go mod download
```

If you only need module names and versions, use `-skip-licenses`. In that mode the tool does not touch the module cache.

## Install

```bash
go install github.com/amazing-generators/godepsgen/cmd/godepsgen@latest
```

## Build Locally

```bash
go build -o ./bin/godepsgen ./cmd/godepsgen
```

## Quick Start

Generate Go output into the current working directory:

```bash
godepsgen
```

Generate Go output for another module:

```bash
godepsgen -source /path/to/project -out /path/to/project/internal/gen/dependencies_gen.go -pkg gen
```

Generate JSON to stdout:

```bash
godepsgen -source /path/to/project -format json -stdout
```

Generate versions only without reading licenses:

```bash
godepsgen -source /path/to/project -skip-licenses -format yaml -out ./artifacts
```

## Run Without Installing

```bash
go run github.com/amazing-generators/godepsgen/cmd/godepsgen@latest \
  -source . \
  -out ./internal/gen/dependencies_gen.go \
  -pkg gen
```

For reproducible usage, pin a version:

```bash
go run github.com/amazing-generators/godepsgen/cmd/godepsgen@v0.1.0 \
  -source . \
  -format json \
  -stdout
```

## CLI

```text
godepsgen [flags]
```

Flags:

- `-source`
  Module directory or direct path to `go.mod`. If omitted, the current working directory is used.
- `-out`
  Output file path or output directory path. If omitted, the default file is created in the current working directory.
- `-pkg`
  Package name for generated Go output. Ignored for JSON and YAML.
- `-format`
  Output format: `go`, `json`, `yaml`, or `yml`. Default: `go`.
- `-stdout`
  Write output to stdout instead of a file.
- `-skip-licenses`
  Skip module directory inspection and emit empty licenses.
- `-force`
  Create missing output directories.
- `-license-max-bytes`
  Maximum allowed size for a single license file. Default: `5242880` bytes (`5 MiB`).
- `-mod-cache`
  Override the module cache root.

Default output files:

- Go: `dependencies_gen.go`
- JSON: `dependencies.json`
- YAML: `dependencies.yml`

If `-out` points to a directory, the default file name for the selected format is created inside that directory.

## Output Path Rules

- If `-out` is omitted and `-stdout` is not set, the file is written into the current working directory.
- If `-out` points to an existing directory, the default file name for the selected format is created inside it.
- If `-out` ends with a path separator or has no file extension, it is treated as a directory path hint.
- Without `-force`, writing into a missing parent directory fails.
- Existing files are replaced atomically.
- If the generated content is unchanged, the existing file is left untouched.

## Common Examples

Read `./go.mod` and write `./dependencies_gen.go`:

```bash
godepsgen
```

Write JSON into `./dependencies.json`:

```bash
godepsgen -format json
```

Write into a directory and let the tool choose the file name:

```bash
godepsgen -out ./internal/gen
```

Read another module but still write into the current directory:

```bash
godepsgen -source /path/to/project
```

Read an explicit `go.mod`:

```bash
godepsgen -source /path/to/project/go.mod -format json -stdout
```

Create nested output directories automatically:

```bash
godepsgen -source /path/to/project -out ./internal/gen/meta/dependencies_gen.go -pkg meta -force
```

Use a custom module cache:

```bash
godepsgen -source /path/to/project -mod-cache /custom/pkg/mod -stdout
```

## `go:generate`

`go generate` runs commands from the package directory that contains the directive. In that setup, `-source .` is often
redundant.

Examples:

```go
//go:generate godepsgen -out ./dependencies_gen.go
```

```go
//go:generate godepsgen -out ./dependencies_gen.go -pkg currentpkg
```

```go
//go:generate go run github.com/amazing-generators/godepsgen/cmd/godepsgen@latest -out ./dependencies_gen.go -pkg currentpkg
```

```go
//go:generate go run github.com/amazing-generators/godepsgen/cmd/godepsgen@v0.1.0 -out ./dependencies_gen.go -pkg currentpkg
```
