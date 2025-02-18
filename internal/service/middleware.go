package service

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	log "github.com/obalunenko/logger"
)

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := log.FromContext(r.Context())

		ctx := log.ContextWithLogger(r.Context(), l)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func logRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		ctx := r.Context()

		rw := newResponseWriter(w)

		next.ServeHTTP(rw, r)

		l := log.WithFields(ctx, log.Fields{
			"method":  r.Method,
			"url":     r.URL.String(),
			"latency": time.Since(now).String(),
			"status":  rw.status,
		})

		switch rw.status {
		case http.StatusInternalServerError:
			l.Error("Request failed")
		case http.StatusNotFound, http.StatusUnauthorized, http.StatusBadRequest, http.StatusForbidden:
			l.Warn("Request processed")
		default:
			l.Info("Request processed")
		}
	})
}

type requestIDKey struct{}

func requestIDMiddleware(next http.Handler) http.Handler {
	key := http.CanonicalHeaderKey("X-Request-ID")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get(key)

		if rid == "" {
			// New random request ID.
			rid = newRequestID()

			r.Header.Set(key, rid)
		}

		ctx := r.Context()

		ctx = context.WithValue(ctx, requestIDKey{}, rid)

		l := log.FromContext(r.Context())
		l = l.WithField("request_id", rid)

		ctx = log.ContextWithLogger(ctx, l)

		w.Header().Set(key, rid)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func newRequestID() string {
	u := uuid.New()

	return u.String()
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status

	rw.ResponseWriter.WriteHeader(status)
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.WithField(r.Context(), "error", err).Error("Panic recovered")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
