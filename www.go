package handy

import (
	"github.com/jbenet/go-is-domain"
	"net"
	"net/http"
	"strings"
)

type nakedToWWWHandler struct{}

// NewNakedToWWWHandler returns a handler that
// redirects to www.
func NewNakedToWWWHandler() http.Handler {
	return nakedToWWWHandler{}
}

func (h nakedToWWWHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Host = "www." + r.Host
	http.Redirect(w, r, r.URL.String(), http.StatusPermanentRedirect)
}

type nakedToWWWWrapper struct {
	wrapped http.Handler
}

// NewNakedToWWWWrapper wraps the given handler with a handler that
// will redirect request to www if its domain is naked,
// or call the wrapped handler if it is not a naked domain.
func NewNakedToWWWWrapper(wrapped http.Handler) http.Handler {
	return &nakedToWWWWrapper{wrapped: wrapped}
}

func (h *nakedToWWWWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
	}
	hostParts := strings.Split(host, ".")
	if !isdomain.IsDomain(strings.Join(hostParts[1:], ".")) {
		r.URL.Host = "www." + r.Host
		http.Redirect(w, r, r.URL.String(), http.StatusPermanentRedirect)
	} else {
		h.wrapped.ServeHTTP(w, r)
	}
}
