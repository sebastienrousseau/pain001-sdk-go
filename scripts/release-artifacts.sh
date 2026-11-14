#!/bin/sh
# SPDX-License-Identifier: Apache-2.0 OR MIT
#
# Builds the files attached to a GitHub release into dist/:
#
#   pain001-sdk-go-<version>.tar.gz   the source at HEAD (git archive)
#   pain001-sdk-go-<version>.cdx.json a CycloneDX SBOM of the module
#   SHA256SUMS                        the SHA-256 of both
#
#   scripts/release-artifacts.sh            # version from version.go
#
# The archive is reproducible for a given commit (git archive sets
# file times from the commit). The SBOM tool runs through `go run` at a
# pinned version, like the other analysers.
set -eu

cd "$(dirname "$0")/.."

CYCLONEDX="github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@v1.12.0"

version="$(sed -n 's/^const Version = "\(.*\)"$/\1/p' version.go)"
name="pain001-sdk-go-$version"
rm -rf dist
mkdir -p dist

git archive --format=tar.gz --prefix="$name/" -o "dist/$name.tar.gz" HEAD
go run "$CYCLONEDX" mod -json -licenses -output "dist/$name.cdx.json" .

(
  cd dist
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$name.tar.gz" "$name.cdx.json" >SHA256SUMS
  else
    shasum -a 256 "$name.tar.gz" "$name.cdx.json" >SHA256SUMS
  fi
)
cat dist/SHA256SUMS
