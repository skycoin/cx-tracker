package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
)

func httpWriteJson(log logrus.FieldLogger, w http.ResponseWriter, _ *http.Request, code int, v interface{}) {
	// TODO: parse http.Request for flags.

	b, err := json.Marshal(v)
	if err != nil {
		httpWriteError(log, w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)

	if _, err := w.Write(b); err != nil {
		log.WithError(err).Error()
	}
	return
}

func httpWriteError(log logrus.FieldLogger, w http.ResponseWriter, code int, err error) {
	log.WithError(err).Error()
	http.Error(w, err.Error(), code)
}

// HTTP Logger Middleware.

type ctxKeyLogger int

// LoggerKey defines logger HTTP context key.
const LoggerKey ctxKeyLogger = -1

// SetLoggerMiddleware sets logger to context of HTTP requests.
func SetLoggerMiddleware(log logrus.FieldLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if reqID := middleware.GetReqID(ctx); reqID != "" && log != nil {
				ctx = context.WithValue(ctx, LoggerKey, log.WithField("RequestID", reqID))
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

func httpLogger(r *http.Request) logrus.FieldLogger {
	log, ok := r.Context().Value(LoggerKey).(logrus.FieldLogger)
	if !ok || log == nil {
		nl := logrus.New()
		nl.Level = logrus.FatalLevel
		log = nl
	}
	return log
}
