package trackers

import (
	"net"
	"strings"
)

func IsTrackerDomain(rawHost string) bool {
	// Handle ports in the host.
	host := rawHost
	if strings.Contains(rawHost, ":") {
		h, _, err := net.SplitHostPort(rawHost)
		if err != nil {
			return false
		}
		host = h
	}

	// Strip out subdomains.
	// Example: www.akitasoftware.com -> akitasoftware.com
	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return false
	}
	rootDomain := strings.Join(parts[len(parts)-2:], ".")

	_, ok := domains[strings.ToLower(rootDomain)]
	return ok
}
