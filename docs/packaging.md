<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# Packaging

## Registry: the Go module proxy

pain001-sdk-go is a Go library. Go distributes modules by version
control tag, not by uploading to a registry:

- A release is the signed annotated tag `v<version>` on GitHub.
- The release workflow then asks proxy.golang.org for that version,
  which fetches and caches it; from then on `go get
  github.com/sebastienrousseau/pain001-sdk-go@v<version>` resolves
  through the proxy, and sum.golang.org records its checksum, so every
  later download is verified against the first.
- pkg.go.dev renders the API reference from the proxy's copy.

There is nothing to upload and no registry account or token.

## Release artifacts

Each GitHub release also carries, for audit rather than installation:

- `pain001-sdk-go-<version>.tar.gz`: the source at the tag
  (`git archive`);
- `pain001-sdk-go-<version>.cdx.json`: a CycloneDX SBOM of the module,
  which lists the module and its (empty) dependency set;
- `SHA256SUMS`: the SHA-256 of both;
- a build-provenance attestation for each file, verifiable with
  `gh attestation verify <file> --repo sebastienrousseau/pain001-sdk-go`.

## Licence

Apache-2.0 OR MIT, at your option (`LICENSE-APACHE`, `LICENSE-MIT`).
Every source file carries the SPDX identifier; `REUSE.toml` covers the
files that cannot carry a header, and `reuse lint` runs in CI.

## Verifying a release

```sh
git clone https://github.com/sebastienrousseau/pain001-sdk-go
cd pain001-sdk-go
git config gpg.ssh.allowedSignersFile KEYS.asc
git verify-tag v0.0.88
sha256sum -c SHA256SUMS   # in a directory holding the downloaded assets
```

`KEYS.asc` lists the maintainer's SSH signing keys in OpenSSH
allowed-signers format.

## Distribution packages

Not applicable. Debian, Fedora, Homebrew and Nix package Go programs,
and vendor Go libraries into each program's build from the module
proxy rather than packaging them on their own; a library with no
binary and no cgo has no distribution package to make. The
minimum-toolchain policy (the two Go releases the Go team supports)
and the empty dependency set are what a packager vendoring this module
needs to know.
