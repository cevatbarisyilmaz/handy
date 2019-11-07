package handy

import (
	"net/http"
)

type hTTPToHTTPSHandler struct{}

func (h hTTPToHTTPSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Scheme = "https"
	r.URL.Host = r.Host
	http.Redirect(w, r, r.URL.String(), http.StatusPermanentRedirect)
}

// NewHTTPToHTTPSHandler returns a handler that
// redirects requests to https.
func NewHTTPToHTTPSHandler() http.Handler {
	return hTTPToHTTPSHandler{}
}

type hTTPToHTTPSWrapper struct {
	wrapped http.Handler
}

// NewHTTPToHTTPSWrapper wraps the given handler with a handler that
// will redirect request to https if it is in http,
// or call the wrapped handler if it is already in https.
func NewHTTPToHTTPSWrapper(wrapped http.Handler) http.Handler {
	return &hTTPToHTTPSWrapper{
		wrapped: wrapped,
	}
}

func (h *hTTPToHTTPSWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.TLS == nil {
		r.URL.Scheme = "https"
		r.URL.Host = r.Host
		http.Redirect(w, r, r.URL.String(), http.StatusPermanentRedirect)
	} else {
		h.wrapped.ServeHTTP(w, r)
	}
}
