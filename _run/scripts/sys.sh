#!/usr/bin/env bash

set -Eeuo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
values_dir="$(dirname "$script_dir")/values"

name_file="${values_dir}/name.txt"
version_file="${values_dir}/ver.txt"

increment_flag=false
target_part=""

# Короткая справка по внутреннему version store.
usage() {
    cat <<'EOF'
Usage:
  sys.sh --name
  sys.sh --ver
  sys.sh --major
  sys.sh --minor
  sys.sh --patch
  sys.sh --increment --major
  sys.sh --increment --minor
  sys.sh --increment --patch
EOF
}

ensure_files() {
    test -f "$name_file" || { echo "Error: missing file $name_file" >&2; exit 1; }
    test -f "$version_file" || { echo "Error: missing file $version_file" >&2; exit 1; }
}

parse_version() {
    local version_value

    version_value="$(tr -d '\r\n[:space:]' < "$version_file")"
    if [[ ! "$version_value" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
        echo "Error: invalid version format in $version_file" >&2
        exit 1
    fi

    ver_major="${BASH_REMATCH[1]}"
    ver_minor="${BASH_REMATCH[2]}"
    ver_patch="${BASH_REMATCH[3]}"
}

print_version() {
    echo "${ver_major}.${ver_minor}.${ver_patch}"
}

increment_version() {
    case "$target_part" in
        major)
            ((ver_major += 1))
            ver_minor=0
            ver_patch=0
            ;;
        minor)
            ((ver_minor += 1))
            ver_patch=0
            ;;
        patch)
            ((ver_patch += 1))
            ;;
        *)
            echo "Error: increment target is required" >&2
            exit 1
            ;;
    esac

    print_version > "$version_file"
    print_version
}

ensure_files
parse_version

if [[ $# -eq 0 ]]; then
    usage
    exit 1
fi

while [[ $# -gt 0 ]]; do
    case "$1" in
        --increment|-i)
            increment_flag=true
            ;;
        --name|-n)
            tr -d '\r\n' < "$name_file"
            exit 0
            ;;
        --ver|-v)
            print_version
            exit 0
            ;;
        --major|-ma)
            target_part="major"
            ;;
        --minor|-mi)
            target_part="minor"
            ;;
        --patch|-pa)
            target_part="patch"
            ;;
        --help|-h)
            usage
            exit 0
            ;;
        *)
            echo "Error: unknown parameter $1" >&2
            exit 1
            ;;
    esac
    shift
done

if [[ "$increment_flag" == "true" ]]; then
    increment_version
    exit 0
fi

case "$target_part" in
    major)
        echo "$ver_major"
        ;;
    minor)
        echo "$ver_minor"
        ;;
    patch)
        echo "$ver_patch"
        ;;
    *)
        echo "Error: increment target is required for numeric output" >&2
        exit 1
        ;;
esac
