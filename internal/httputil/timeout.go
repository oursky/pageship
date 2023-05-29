package httputil

import (
	"io"
	"net/http"
	"time"
)

type timeoutReader struct {
	r           io.Reader
	ctrl        *http.ResponseController
	idleTimeout time.Duration
}

func NewTimeoutReader(r io.Reader, ctrl *http.ResponseController, idleTimeout time.Duration) io.Reader {
	return &timeoutReader{r: r, ctrl: ctrl, idleTimeout: idleTimeout}
}

func (r *timeoutReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)

	dl := time.Now().Add(r.idleTimeout)
	r.ctrl.SetReadDeadline(dl)
	r.ctrl.SetWriteDeadline(dl)
	return n, err
}

type timeoutResponseWriter struct {
	w           http.ResponseWriter
	ctrl        *http.ResponseController
	idleTimeout time.Duration
}

func NewTimeoutResponseWriter(w http.ResponseWriter, idleTimeout time.Duration) http.ResponseWriter {
	return &timeoutResponseWriter{w: w, ctrl: http.NewResponseController(w), idleTimeout: idleTimeout}
}

func (w *timeoutResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *timeoutResponseWriter) WriteHeader(statusCode int) {
	w.w.WriteHeader(statusCode)
}

func (w *timeoutResponseWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)

	dl := time.Now().Add(w.idleTimeout)
	w.ctrl.SetWriteDeadline(dl)
	return n, err
}
