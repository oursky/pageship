package httputil

import (
	"io"
	"net/http"
	"time"
)

type timeoutReader struct {
	r           io.Reader
	ctrl        *http.ResponseController
	readTimeout time.Duration
}

func NewTimeoutReader(r io.Reader, ctrl *http.ResponseController, readTimeout time.Duration) io.Reader {
	return &timeoutReader{r: r, ctrl: ctrl, readTimeout: readTimeout}
}

func (r *timeoutReader) Read(p []byte) (int, error) {
	dl := time.Now().Add(r.readTimeout)
	r.ctrl.SetReadDeadline(dl)
	r.ctrl.SetWriteDeadline(dl)

	n, err := r.r.Read(p)

	r.ctrl.SetReadDeadline(time.Time{})
	r.ctrl.SetWriteDeadline(time.Time{})

	return n, err
}

type timeoutResponseWriter struct {
	w            http.ResponseWriter
	ctrl         *http.ResponseController
	writeTimeout time.Duration
}

func NewTimeoutResponseWriter(w http.ResponseWriter, writeTimeout time.Duration) http.ResponseWriter {
	return &timeoutResponseWriter{w: w, ctrl: http.NewResponseController(w), writeTimeout: writeTimeout}
}

func (w *timeoutResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *timeoutResponseWriter) WriteHeader(statusCode int) {
	w.w.WriteHeader(statusCode)
}

func (w *timeoutResponseWriter) Write(p []byte) (int, error) {
	dl := time.Now().Add(w.writeTimeout)
	w.ctrl.SetWriteDeadline(dl)

	n, err := w.w.Write(p)

	w.ctrl.SetWriteDeadline(time.Time{})

	return n, err
}
