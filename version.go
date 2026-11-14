// SPDX-License-Identifier: Apache-2.0 OR MIT

package pain001

// Version is this module's release, in lockstep with the pain001 suite.
// scripts/verify-release-versions checks every file that restates it.
const Version = "0.0.88"

// userAgent identifies the client on every request, so a server's access
// log can tell which SDK and release sent a call.
const userAgent = "pain001-sdk-go/" + Version
