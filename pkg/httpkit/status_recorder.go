package httpkit

import (
	"net/http"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	if !r.written {
		r.statusCode = statusCode
		r.written = true
		r.ResponseWriter.WriteHeader(statusCode)
	}
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.written {
		r.statusCode = 200
		r.written = true
	}

	return r.ResponseWriter.Write(b)
}

func (r *statusRecorder) Flush() {
	if !r.written {
		r.written = true
		if r.statusCode == 0 {
			r.statusCode = http.StatusOK
		}
	}

	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}
