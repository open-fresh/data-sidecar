package util

import (
	"bytes"
	"net/http"
)

// SimpleRecorder is a recorder for testing purposes.
type SimpleRecorder struct {
	Chan chan Metric
}

// NewRecorder makes a new SimpleRecorder for testing purposes
func NewRecorder() *SimpleRecorder {
	tr := SimpleRecorder{make(chan Metric, 10000)}
	return &tr
}

// Record records to the channel
func (t *SimpleRecorder) Record(x Metric) {
	(*t).Chan <- x
}

func (t *SimpleRecorder) Finish() {
	close((*t).Chan)
}

// NullRecorder throws stuff away.
type NullRecorder struct{}

// NewNullRecorder does what it says.
func NewNullRecorder() *NullRecorder {
	return &NullRecorder{}
}

// Record for a null recorder,well, doesn't.
func (n *NullRecorder) Record(x Metric) {}

// Finish also does nothing.
func (n *NullRecorder) Finish() {}

// HTTPRespWrite is a usable class to manipulate http functionality.
type HTTPRespWrite struct {
	Store *bytes.Buffer
}

//NewHTTPResponseWriter generates a mock http response writer for testing handlefuncs
func NewHTTPResponseWriter() *HTTPRespWrite {
	return &HTTPRespWrite{bytes.NewBuffer([]byte{})}
}

// Header does nothing
func (h *HTTPRespWrite) Header() http.Header {
	return make(http.Header)
}

// Write puts bytes in the buffer
func (h *HTTPRespWrite) Write(bts []byte) (int, error) {
	return (*h).Store.Write(bts)
}

// WriteHeader also does nothing.
func (h *HTTPRespWrite) WriteHeader(int) {}

// String dumps a string out of the buffer
func (h *HTTPRespWrite) String() string {
	return h.Store.String()
}
