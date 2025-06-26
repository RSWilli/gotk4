#!/usr/bin/env bash

# This script downloads the gtk-rs/gir-files repository and copies the GIR files into the
# girs directory

set -euo pipefail

repository="https://github.com/gtk-rs/gir-files.git"

destination="girs"

if [ ! -d "$destination" ]; then
    echo "Creating directory $destination"
    mkdir "$destination"
fi

# clone into a temporary directory
temp_dir=$(mktemp -d)
trap 'rm -rf "$temp_dir"' EXIT


git clone "$repository" "$temp_dir"


cp -r "$temp_dir"/*.gir "$destination"
