package handy

import (
	"net"
	"net/http"
	"strings"
)

type hostAuthWrapper struct {
	wrapped    http.Handler
	hostParts  []string
	subdomains map[string]bool
}

// NewHostAuthWrapper wraps the given handler with a handler that
// will turn down the requests directed to a host that server is
// not authorized for.
//
// If subdomains are nil or empty, it will accept all subdomains
// as well as naked domain.

// If not, it will only accept the given subdomains.
// Add empty string to accept the naked domain as well.
func NewHostAuthWrapper(wrapped http.Handler, host string, subdomains []string) http.Handler {
	var subdomainMap map[string]bool
	if subdomains == nil || len(subdomains) == 0 {
		subdomainMap = nil
	} else {
		subdomainMap = map[string]bool{}
		for _, subdomain := range subdomains {
			subdomainMap[subdomain] = true
		}
	}
	return &hostAuthWrapper{
		wrapped:    wrapped,
		hostParts:  strings.Split(host, "."),
		subdomains: subdomainMap,
	}
}

func (h *hostAuthWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var host string
	var err error
	host, _, err = net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
	}
	hostParts := strings.Split(host, ".")
	diff := len(hostParts) - len(h.hostParts)
	if diff < 0 {
		http.Error(w, http.StatusText(http.StatusMisdirectedRequest), http.StatusMisdirectedRequest)
		return
	}
	for i, part := range h.hostParts {
		if part != hostParts[i+diff] {
			http.Error(w, http.StatusText(http.StatusMisdirectedRequest), http.StatusMisdirectedRequest)
			return
		}
	}
	if h.subdomains != nil {
		if diff > 0 {
			if !h.subdomains[strings.Join(hostParts[:diff], ".")] {
				http.Error(w, http.StatusText(http.StatusMisdirectedRequest), http.StatusMisdirectedRequest)
				return
			}
		} else if !h.subdomains[""] {
			http.Error(w, http.StatusText(http.StatusMisdirectedRequest), http.StatusMisdirectedRequest)
			return
		}
	}
	h.wrapped.ServeHTTP(w, r)
}
