package handler

import (
	"net/http"
)

type BasicAuth struct {
	Username string
	Password string
}

func (b BasicAuth) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			writeUnauthorized(w)
			return
		}

		if username != b.Username && password != b.Password {
			writeUnauthorized(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	w.WriteHeader(http.StatusUnauthorized)
}
