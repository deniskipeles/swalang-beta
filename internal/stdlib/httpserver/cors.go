// pylearn/internal/stdlib/pyhttpserver/cors.go
package pyhttpserver

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/deniskipeles/pylearn/internal/object"
)

// CORSOptions holds the configuration for the CORS middleware.
type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// CORSWrapper is a Pylearn object that holds the CORS configuration
// and the Pylearn handler it needs to wrap.
type CORSWrapper struct {
	Options     *CORSOptions
	PylnHandler object.Object
}

func (cw *CORSWrapper) Type() object.ObjectType { return "CORS_WRAPPER" }
func (cw *CORSWrapper) Inspect() string {
	origins := strings.Join(cw.Options.AllowedOrigins, ", ")
	return "<CORSWrapper origins=[" + origins + "]>"
}

// This struct is a Pylearn object but does not need to be callable or have attributes.
// It's just a container for the Go middleware logic.
var _ object.Object = (*CORSWrapper)(nil)

// ServeHTTP is the core Go middleware. It gets called by the http.Server
// before the main Pylearn application logic.
func (cw *CORSWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.Handler) {
	origin := r.Header.Get("Origin")
	if origin == "" { // Not a CORS request
		next.ServeHTTP(w, r)
		return
	}

	isAllowedOrigin := false
	if len(cw.Options.AllowedOrigins) == 0 || (len(cw.Options.AllowedOrigins) == 1 && cw.Options.AllowedOrigins[0] == "*") {
		isAllowedOrigin = true
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		for _, allowed := range cw.Options.AllowedOrigins {
			if allowed == origin {
				isAllowedOrigin = true
				w.Header().Set("Access-Control-Allow-Origin", allowed)
				break
			}
		}
	}

	if isAllowedOrigin {
		w.Header().Add("Vary", "Origin")
	}

	// Handle preflight (OPTIONS) requests
	if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
		if !isAllowedOrigin {
			http.NotFound(w, r)
			return
		}
		if cw.Options.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if len(cw.Options.AllowedMethods) > 0 {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(cw.Options.AllowedMethods, ", "))
		}
		if len(cw.Options.AllowedHeaders) > 0 {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(cw.Options.AllowedHeaders, ", "))
		}
		if cw.Options.MaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cw.Options.MaxAge))
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if !isAllowedOrigin {
		next.ServeHTTP(w, r)
		return
	}

	if cw.Options.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	next.ServeHTTP(w, r)
}