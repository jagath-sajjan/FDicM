#!/usr/bin/env bash

set -e

VERSION="v1.0.0"
if command -v git &> /dev/null && git rev-parse --is-inside-work-tree &> /dev/null; then
    VERSION=$(git describe --tags --always 2>/dev/null || echo "v1.0.0")
fi

echo "Compiling artifacts for distributed targets ($VERSION)..."

rm -rf dist release
mkdir -p dist release

cat << EOF > dist/README.md
# FDicM ($VERSION)

The high-fidelity, split-pane responsive terminal dictionary TUI application engine built using Go and Bubble Tea.

## Global Controls
- TAB      : Toggle focus panel viewport frames natively (Vim controls activated on focus).
- p        : Stream live audio pronunciation vocal sample elements via text-to-speech.
- v        : Toggle raw JSON payload interface structure visibility data trees on the fly.
- c        : Snapshot capture full clean structural copy definition out to system clipboard.
- q / ESC : Escape current result window frame or terminate application process cleanly.
EOF

cat << EOF > dist/LICENSE
MIT License

Copyright (c) $(date +%Y) FDicM Authors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files...
EOF

package_target() {
    OS=$1
    ARCH=$2
    SUFFIX=$3
    FORMAT=$4

    BINARY_NAME="fdicm"
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="fdicm.exe"
    fi

    echo "  Building target: OS=$OS ARCH=$ARCH..."

    env CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build \
        -tags cross \
        -ldflags="-s -w -X main.version=${VERSION}" \
        -o "dist/${BINARY_NAME}" main.go

    if [ "$FORMAT" = "tar.gz" ]; then
        tar -czf "release/fdicm_${VERSION}_${OS}_${ARCH}.tar.gz" -C dist "${BINARY_NAME}" README.md LICENSE
    elif [ "$FORMAT" = "zip" ]; then
        (cd dist && zip -q "../release/fdicm_${VERSION}_${OS}_${ARCH}.zip" "${BINARY_NAME}" README.md LICENSE)
    fi

    rm -f "dist/${BINARY_NAME}"
}

package_target "darwin" "arm64" "" "tar.gz"
package_target "darwin" "amd64" "" "tar.gz"
package_target "linux"  "amd64" "" "tar.gz"
package_target "linux"  "arm64" "" "tar.gz"
package_target "windows" "amd64" ".exe" "zip"

rm -f dist/README.md dist/LICENSE
rmdir dist

echo "All production bundles and version-injected packages locked inside /release/"
ls -lh release
