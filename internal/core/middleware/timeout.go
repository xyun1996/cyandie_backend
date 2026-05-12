package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()

			done := make(chan struct{})
			tw := &timeoutWriter{ResponseWriter: w}

			go func() {
				next.ServeHTTP(tw, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				if !tw.written {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusServiceUnavailable)
					json.NewEncoder(w).Encode(map[string]any{
						"ok": false,
						"error": map[string]any{
							"code":      "TIMEOUT",
							"message":   "request timeout",
							"requestId": r.Header.Get("X-Request-ID"),
						},
					})
				}
			}
		})
	}
}

type timeoutWriter struct {
	http.ResponseWriter
	written bool
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.written = true
	tw.ResponseWriter.WriteHeader(code)
}

func (tw *timeoutWriter) Write(p []byte) (int, error) {
	tw.written = true
	return tw.ResponseWriter.Write(p)
}
