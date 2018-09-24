package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Fresh-Tracks/data-sidecar/util"
)

func init() {
	ticker = overrideTicker
}

func overrideTicker(time.Duration) <-chan time.Time {
	x := make(chan time.Time, 2)
	x <- time.Now()
	close(x)
	return x
}

func SimpleHandleFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello")
	return
}

func TestMonitor(t *testing.T) {
	rw := util.NewHTTPResponseWriter()
	r := &http.Request{URL: &url.URL{Path: "abcd"}}

	x := Monitor(SimpleHandleFunc)
	x(rw, r)
	if g := rw.String(); !strings.Contains(g, "hello") {
		t.Error(g)
	}

}

func TestTicker(t *testing.T) {
	g := Ticker(time.Microsecond)
	found := false
	for ii := 0; ii < 10; ii++ {
		select {
		case _ = <-g:
			found = true
		}
	}
	if !found {
		t.Error("not found")
	}

}

func TestMain(t *testing.T) {
	main()
}
