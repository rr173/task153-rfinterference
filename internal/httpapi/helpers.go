package httpapi

import (
	"net/http"
	"time"
)

func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", r.Header.Get("X-Request-ID"))
		next.ServeHTTP(w, r)
	})
}
func Server(addr string, handler http.Handler) *http.Server {
	return &http.Server{Addr: addr, Handler: WithRequestID(handler), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second}
}
