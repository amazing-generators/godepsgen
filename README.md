# dependencies

Self-contained Go dependencies generator.

Examples:

```bash
godepsgen -source /path/to/project -out /path/to/project/internal/gen/dependencies_gen.go -pkg gen
godepsgen -source /path/to/project/go.mod -format json -stdout
godepsgen -source /path/to/project -out /tmp/dependencies.json -format json
godepsgen -source /path/to/project -out /path/to/project/internal/gen/dependencies_gen.go -skip-licenses
```

Notes:

- `-source` is optional: when omitted, the current working directory is used.
- `-source` may point either to a module directory or directly to a `go.mod` file.
- The tool parses `go.mod` in pure Go and does not call `go list`.
- Remote modules are resolved through the module cache.
- If a required module directory is missing or empty, the command fails and expects a warmed cache.
- Missing license files are allowed and generate empty license values when the module directory exists and is non-empty.
- By default a missing output directory is an error; pass `-force` to create the directory tree. Existing files are
  always overwritten.
