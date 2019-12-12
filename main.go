// A sidecar for prometheus, it runs data queries and aggregates them together
// in a way that is complicated/slow in other languages, then keeps the data for use in
// other side-sidecars
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/open-fresh/data-sidecar/icarus"
	"github.com/open-fresh/data-sidecar/prom"
	"github.com/open-fresh/data-sidecar/scoring"
	"github.com/open-fresh/data-sidecar/storage"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	//metrics
	attemptCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "sidecar_internal_attempts_count",
		Help: "Number of attempts within sidecar"},
		[]string{"type"})

	requestSummary = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "sidecar_http_request_summary",
		Help: "Number and timing of http requests to sidecar"},
		[]string{"type"})

	// cheats and hacks
	ticker   = Ticker
	logFatal = log.Fatal
	// flags

	port       = flag.Int("port", 8077, "port on which to expose metrics")
	cleanup    = flag.Int("cleanup", 300, "time after which a missing series may be garbage collected (seconds)")
	p8s        = flag.String("prom", "http://querier/api/prom/", "which prometheus to scrape")
	resolution = flag.Int("resolution", 10, "range query resolution (seconds)")
	lookback   = flag.Int("lookback", 60, "empirical lookback window (minutes)")
	prefix     = flag.String("pfx", "ft_", "export prefix for metrics")
	version    = "undefined"
)

func init() {
	prometheus.MustRegister(attemptCounter)
	prometheus.MustRegister(requestSummary)
}

// Monitor passthrough-instruments a handlefunc.
func Monitor(f http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timer := prometheus.NewTimer(requestSummary.WithLabelValues((*r).URL.Path))
		defer timer.ObserveDuration()
		f(w, r)
	})
}

// Ticker is a way to return a ticker of a duration, or to do something completely different for testing.
func Ticker(hh time.Duration) <-chan time.Time {
	hygeineTicker := time.NewTicker(hh)
	return hygeineTicker.C
}

// main starts all the goroutines and web endpoints
func main() {
	// entire ecosystem of channels and tables.
	flag.Parse()
	log.Println("data sidecar version", version)
	mux := http.DefaultServeMux
	log.Println("serving prometheus endpoints on port", *port)

	// serve all the web goodies.
	server := &http.Server{Addr: fmt.Sprintf(":%v", *port), Handler: mux}
	go func() { logFatal(server.ListenAndServe()) }()

	seriesCollection := storage.NewStore()

	mux.HandleFunc("/dump", Monitor(seriesCollection.DumpHandleFunc))
	remote := icarus.NewIcarus(*prefix)
	mux.HandleFunc("/metrics", Monitor(remote.HandleFunc))
	scorer := scoring.NewScorer(seriesCollection, remote)

	mux.HandleFunc("/score", Monitor(scorer.ScoreHandleFunc))

	promClient := prom.NewClient(*p8s, *resolution, *lookback, scorer)
	log.Println(promClient.Status())
	promClient.Start()
	hygeineTicker := ticker(time.Duration(*cleanup)*time.Second + time.Microsecond)
	for range hygeineTicker {
		removed := float64(len(seriesCollection.Prune(*cleanup)))
		attemptCounter.WithLabelValues("deleteSeries").Add(removed)
	}
}
