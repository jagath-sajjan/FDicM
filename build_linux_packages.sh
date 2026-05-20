#!/usr/bin/env bash

set -e

VERSION="1.0.0"
rm -rf debian_build
mkdir -p debian_build/fdicm_${VERSION}_amd64/DEBIAN
mkdir -p debian_build/fdicm_${VERSION}_amd64/usr/local/bin

cat << 'INNER_EOF' > debian_build/fdicm_${VERSION}_amd64/DEBIAN/control
Package: fdicm
Version: 1.0.0
Section: utils
Priority: optional
Architecture: amd64
Maintainer: jagath-sajjan
Description: Blazing fast responsive split pane terminal dictionary app engine
 High fidelity split pane responsive terminal dictionary client application
 featuring fuzzy matching auto completion and local compressed offline storage fallbacks
INNER_EOF

tar -xzf release/fdicm_v1.0.0_linux_amd64.tar.gz -C debian_build/fdicm_${VERSION}_amd64/usr/local/bin/ fdicm

dpkg-deb --build debian_build/fdicm_${VERSION}_amd64
mv debian_build/fdicm_${VERSION}_amd64.deb release/

rm -rf debian_build
echo "Linux deb package generated cleanly inside release folder"
