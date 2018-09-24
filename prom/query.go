package prom

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/Fresh-Tracks/data-sidecar/util"
	"github.com/prometheus/client_golang/prometheus"
)

const scopeOrgIDHeader = "X-Scope-OrgID"

var (
	//errors
	errProm      = errors.New("Prometheus returned non-success output")
	errorCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "sidecar_internal_errors_count",
		Help: "Number of errors within sidecar"},
		[]string{"type"})
	internalDataSummary = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "sidecar_internal_data_size_summary",
		Help: "The size of data returns held by the sidecar.",
	},
		[]string{"type"})
	queryDurationsSummary = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "sidecar_query_duration_summary",
		Help: "How long are range queries taking?",
	},
		[]string{"type"})
)

// Client queries prometheus.
type Client struct {
	*sync.Mutex
	Store    util.ScoringEngine
	P8s      string
	Res      int
	Lookback int
	start    int
	end      int
	client   *http.Client
	series   map[string]bool
	Stopped  bool
}

// RangeQ represents a range query
type RangeQ struct {
	Status string
	Data   struct {
		ResultType string
		Result     []struct {
			Metric map[string]string
			Values [][]json.Number
		}
	}
}

// SeriesMatch holds the results of the /api/v1/series?match[]= endpoint
type SeriesMatch struct {
	Status string
	Data   []map[string]string
}

func init() {
	prometheus.MustRegister(errorCounter)
	prometheus.MustRegister(internalDataSummary)
	prometheus.MustRegister(queryDurationsSummary)
}

// NewClient builds a prometheus client.
func NewClient(p8s string, res, lbk int, store util.ScoringEngine) *Client {
	var mux sync.Mutex
	client, _ := httpClient()
	start := int(time.Now().Unix()) - lbk*60
	end := int(time.Now().Unix())
	return &Client{&mux, store, p8s, res, lbk, start, end, client, make(map[string]bool), false}
}

// HTTPClient generates an http client from the configuration
// provided in the yaml input.
func httpClient() (*http.Client, error) {
	client := &http.Client{
		Transport: util.SingleConnNoKeepAliveTransporter(),
		Timeout:   15 * time.Second,
	}
	return client, nil
}

// Fetch queries prometheus over http at a given endpoint and returns the body
func (c *Client) Fetch(endpt string) ([]byte, error) {
	pfx := queryExtract(endpt)
	timer := prometheus.NewTimer(queryDurationsSummary.WithLabelValues(pfx))
	req, _ := http.NewRequest("GET", endpt, nil)
	resp, err := c.client.Do(req)
	if err != nil {
		errorCounter.WithLabelValues("reaching_p8s").Inc()
		return []byte{}, err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	timer.ObserveDuration()

	return body, nil
}

// DecodeRangeQ takes a response from the p8s query_range endpoint and decodes it.
func DecodeRangeQ(response []byte) (RangeQ, error) {
	var target RangeQ
	jsonErr := json.Unmarshal(response, &target)
	if jsonErr != nil {
		errorCounter.WithLabelValues("json_decode").Inc()
		return target, jsonErr
	}
	return target, nil
}

// DecodeSeriesMatch takes a response from the p8s series?match[] endpoint and decodes it.
func DecodeSeriesMatch(response []byte) SeriesMatch {
	var target SeriesMatch
	jsonErr := json.Unmarshal(response, &target)
	if jsonErr != nil {
		errorCounter.WithLabelValues("json_decode").Inc()
	}
	if len(target.Data) == 0 {
		errorCounter.WithLabelValues("noSeries").Inc()
	}

	return target
}

// SeriesQuery generates a series querying string to be fetchdecoded.
func (c *Client) SeriesQuery() string {
	return fmt.Sprintf("%s/api/v1/series?match[]={ft_target=\"true\"}&start=%d&end=%d", c.P8s, c.start, time.Now().Unix())
}

// knownSeries returns a list of known series names for p8s.
func (c *Client) knownSeries() []string {
	c.Lock()
	defer c.Unlock()
	temp := make([]string, len(c.series))
	index := 0
	for xx := range c.series {
		temp[index] = xx
		index++
	}
	return temp
}

//SeriesInsert puts series into a series store
func (c *Client) SeriesInsert(input SeriesMatch) (number int) {
	c.Lock()
	defer c.Unlock()
	for _, xx := range input.Data {
		metricName := xx["__name__"]
		c.series[metricName] = true
	}
	for _ = range c.series {
		number++
	}
	internalDataSummary.WithLabelValues("series").Observe(float64(number))
	return
}

// RangeQuery describes a prometheus range query needs timing and step information
func (c *Client) RangeQuery(series string) string {
	return fmt.Sprintf("%s/api/v1/query_range?query=%s{ft_target=\"true\"}&start=%v&end=%v&step=%vs",
		c.P8s, series, c.start, c.end, c.Res)
}

// RangeInsert turns RangeQ and puts them into internal storage.
func (c *Client) RangeInsert(result RangeQ) {
	internalDataSummary.WithLabelValues("range").Observe(float64(len(result.Data.Result)))
	for _, xx := range result.Data.Result {
		mydata := make([]util.DataPoint, 0, len(xx.Values))
		for _, yy := range xx.Values {
			val, _ := yy[1].Float64()
			time, _ := yy[0].Int64()
			if !math.IsNaN(val) && !math.IsInf(val, 0) {
				mydata = append(mydata, util.DataPoint{Val: val, Time: time})
			}
		}
		if len(mydata) > 0 {
			c.Store.ScoreData(mydata, xx.Metric, true)
		}
	}
}

// RangeBatch does a range query for all the things that we know about.
func (c *Client) RangeBatch() {
	for _, xx := range c.knownSeries() {
		query := c.RangeQuery(xx)
		resp, err := c.Fetch(query)
		if err != nil {
			errorCounter.WithLabelValues("range query error").Inc()
			return
		}

		series, _ := DecodeRangeQ(resp)
		c.RangeInsert(series)
	}
}

// queryExtract pulls the query endpoint out of the query string. With short=false, it includes the name.
func queryExtract(query string) string {
	symbol := '?'
	for xx, yy := range query {
		if yy == symbol {
			return query[0:xx]
		}
	}
	return query
}

// SeriesBatch does an iteration of series work
func (c *Client) SeriesBatch() int {
	query := c.SeriesQuery()
	resp, err := c.Fetch(query)
	if err != nil {
		errorCounter.WithLabelValues("series query error").Inc()
		return 0
	}
	series := DecodeSeriesMatch(resp)
	return c.SeriesInsert(series)
}

// PullData does all the prom stuff.
func (c *Client) PullData() (out int) {
	out = c.SeriesBatch()
	c.RangeBatch()
	return
}

// Status is a human-readable output of what prom is trying to do.
func (c *Client) Status() string {
	return fmt.Sprintf("Looking for prometheus at %s", c.P8s)
}

// Stop a prometheus client
func (c *Client) Stop() {
	c.Stopped = true
}

// Restart a stopped prom client
func (c *Client) Restart() {
	c.Stopped = false
	c.client, _ = httpClient()
	c.start = int(time.Now().Unix()) - c.Lookback*60
	c.end = int(time.Now().Unix())
}

func (c *Client) cycle() {
	tck := time.NewTicker(time.Duration(c.Res)*time.Second + time.Microsecond)
	for _ = range tck.C {
		if c.Stopped {
			continue
		}
		numSeries := c.PullData()
		if numSeries == 0 {
			errorCounter.WithLabelValues("failed to get any series from p8s").Inc()
			c.client, _ = httpClient()
		} else {
			c.start = c.end
		}
		c.end = int(time.Now().Unix())
	}
}

// Start the runtime cycle.
func (c *Client) Start() {
	go c.cycle()
}
