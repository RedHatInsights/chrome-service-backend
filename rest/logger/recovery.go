package logger

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

// Recoverer records request-scoped diagnostics without logging panic contents,
// which may contain untrusted or sensitive data.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		defer func() {
			panicValue := recover()
			if panicValue == nil {
				return
			}
			if panicValue == http.ErrAbortHandler {
				panic(panicValue)
			}

			stack := debug.Stack()
			if len(stack) > 8192 {
				stack = stack[:8192]
			}
			LogFor(r.Context()).WithFields(logrus.Fields{
				"panic_type": fmt.Sprintf("%T", panicValue),
				"stack":      string(stack),
			}).Error("request panic recovered")
			if wrapped.Status() == 0 && wrapped.BytesWritten() == 0 {
				http.Error(wrapped, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(wrapped, r)
	})
}
