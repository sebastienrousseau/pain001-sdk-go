<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

## What this changes

<!-- What behaviour changes, and why. -->

## How it was verified

<!-- What you ran, and what you observed. Name the test and the result
     it produced; say whether you ran `make check` with
     PAIN001_BASE_URL set against a live server. -->

## Checklist

- [ ] Every commit is signed off (`git commit -s`), see [DCO.txt](../DCO.txt)
- [ ] `make check` passes, with and without `PAIN001_BASE_URL`
- [ ] Tests cover the change
- [ ] `conformance/` is untouched, or replaced whole from core with `SHA256SUMS` regenerated
- [ ] Documentation updated if behaviour or interfaces changed
