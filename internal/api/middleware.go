package api

import (
	sessionstore "asostechtest/internal/sessionstore"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

// This middleware reads the session id off requests to the
// /sessions/{session_id}/... endpoints and checks that the session exists. If
// it does exist the session is added to the context for easy access down steam
// otherwise we fail fast with a 404.
func (h *Handlers) sessionCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "sessionID")
		session, err := h.sessionStore.GetSession(sessionID)
		h.logger.Info("got here", "session", session)
		if err != nil {
			if err == sessionstore.ErrDatabaseError {
				render.Render(w, r, h.ErrInternalServer(err))
			} else if err == sessionstore.ErrSessionNotFound ||
				err == sessionstore.ErrSessionExpired {
				render.Render(w, r, ErrNotFound())
			}
			return
		}

		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// *****
// Didn't find a middleware for slog and it's the logger I really like so I
// created one for it:

// NewSlogRequestLogger returns a request logger middleware handler which
// utilises slog for logging.
func NewSlogRequestLogger(handler slog.Handler) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&SlogWrapper{Logger: handler})
}

// SLogWrapper implements chi/middleware.LogFormatter interface and allows
// logging middleware to make use of the default slog logger.
type SlogWrapper struct {
	Logger slog.Handler
}

// NewLogEntry creates a new request LogEntry; called by the go-chi request
// logging middleware logic.
func (l *SlogWrapper) NewLogEntry(r *http.Request) middleware.LogEntry {
	var logFields []slog.Attr
	logFields = append(logFields, slog.String("ts", time.Now().UTC().Format(time.RFC3339)))

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields = append(logFields, slog.String("req_id", reqID))
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	handler := l.Logger.WithAttrs(append(logFields,
		slog.String("scheme", scheme),
		slog.String("proto", r.Proto),
		slog.String("method", r.Method),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("user_agent", r.UserAgent()),
		slog.String("uri", fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI))))

	entry := StructuredLoggerEntry{Logger: slog.New(handler)}
	entry.Logger.LogAttrs(r.Context(), slog.LevelInfo, "request made")

	return &entry
}

type StructuredLoggerEntry struct {
	Logger *slog.Logger
}

func (l *StructuredLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	l.Logger.LogAttrs(context.Background(), slog.LevelInfo, "request complete",
		slog.Int("resp_status", status),
		slog.Int("resp_byte_length", bytes),
		slog.Int64("resp_elapsed_ms", elapsed.Milliseconds()),
	)
}

func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.Logger.LogAttrs(context.Background(), slog.LevelInfo, "",
		slog.String("stack", string(stack)),
		slog.String("panic", fmt.Sprintf("%+v", v)),
	)
}
