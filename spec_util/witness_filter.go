package spec_util

import (
	"strings"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

const (
	XAkitaCLIGitVersion = "x-akita-cli-git-version"
	XAkitaRequestID     = "x-akita-request-id"
	XAkitaDogfood       = "x-akita-dogfood"
)

// Returns true if the witness represents traffic from the CLI to our backend.
// These should get filtered out.
func ContainsCLITraffic(w *pb.Witness) bool {
	isDogfood := false
	containsCLIGitVersion := false

	m := w.GetMethod()
	for _, arg := range m.GetArgs() {
		if containsAkitaDogfoodHeader(arg) {
			isDogfood = true
			break
		} else if containsCLIGitVersionHeader(arg) {
			containsCLIGitVersion = true
			// Don't break because the next header could be x-akita-dogfood.
		}
	}

	if isDogfood {
		// If we're dogfooding our own API, don't filter out this witness, treat
		// it as a regular API call.
		return false
	} else if containsCLIGitVersion {
		return true
	}

	for _, resp := range m.GetResponses() {
		if containsRequestIDHeader(resp) {
			return true
		}
	}
	return false
}

// Heuristic: assume an HTTP request containing x-akita-cli-git-version header
// is sent from our CLI.
func containsCLIGitVersionHeader(arg *pb.Data) bool {
	hdr := HTTPHeaderFromData(arg)
	return strings.ToLower(hdr.GetKey()) == XAkitaCLIGitVersion
}

// Heuristic: assume an HTTP response containing x-akita-request-id header is
// sent from our API.
func containsRequestIDHeader(resp *pb.Data) bool {
	hdr := HTTPHeaderFromData(resp)
	return strings.ToLower(hdr.GetKey()) == XAkitaRequestID
}

func containsAkitaDogfoodHeader(arg *pb.Data) bool {
	hdr := HTTPHeaderFromData(arg)
	return strings.ToLower(hdr.GetKey()) == XAkitaDogfood
}
