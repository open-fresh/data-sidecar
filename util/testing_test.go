package util

import (
	"testing"
)

func TestRecord(t *testing.T) {
	x := NewRecorder()
	x.Record(Metric{Desc: map[string]string{"a": "b"}, Data: DataPoint{Val: 0.0}})
	close(x.Chan)
	for y := range x.Chan {
		if y.Data.Val != 0.0 {
			t.Error("value")
		}
		if val, ok := y.Desc["a"]; (val != "b") || (!ok) {
			t.Error("description")
		}
	}
}

func TestHTTP(t *testing.T) {
	x := NewHTTPResponseWriter()
	y := x.Header()
	if len(y) != 0 {
		t.Error()
	}
	x.Write([]byte("hello:everybody"))
	z := x.String()
	if z != "hello:everybody" {
		t.Error(z)
	}

}
