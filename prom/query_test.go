package prom

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Fresh-Tracks/data-sidecar/util"
)

var (
	n  = NullScorer{lastTime: make(map[string]int64)}
	pc = NewClient("", 10, 60, &n)
)

type NullScorer struct {
	added    int
	scored   int
	lastTime map[string]int64
}

func (n *NullScorer) Add(labels map[string]string, value float64, ts int64) bool {
	labelString := util.MapSSToS(labels)
	if n.lastTime[labelString] < ts {
		n.added++
		n.lastTime[labelString] = ts
		return true
	}
	return false
}

func (n *NullScorer) Score(map[string]string) {
	n.scored++
}

func (n *NullScorer) ScoreData([]util.DataPoint, map[string]string, bool) {
}

func (n *NullScorer) ScoreCollective() {
}

func (n *NullScorer) Reset() {
	n.added = 0
	n.scored = 0
}

func TestDecode(t *testing.T) {
	inp := []byte(`{"Status":"ok"}`)
	g, err := DecodeRangeQ(inp)
	if (g.Status != "ok") || (err != nil) {
		t.Error(g, err)
	}
	g, err = DecodeRangeQ([]byte("basdf"))
	if err == nil {
		t.Error(err)
	}
}

func TestQueries(t *testing.T) {
	g := pc.SeriesQuery()
	if !strings.Contains(g, "/api/v1/series?match[]={ft_target") {
		t.Error(g)
	}
	g = pc.RangeQuery("abcd")
	if !strings.Contains(g, "/api/v1/query_range?query") {
		t.Error(g)
	}
}

func TestSeriesBits(t *testing.T) {
	x := pc.knownSeries()
	if len(x) > 0 {
		t.Error(x)
	}
	seriesInp := []byte(`{"Status":"ok",
		"Data":[{"__name__":"b"}]}`)
	g := DecodeSeriesMatch(seriesInp)
	num := pc.SeriesInsert(g)
	if num != 1 {
		t.Error(num)
	}
	x = pc.knownSeries()
	if len(x) != 1 {
		t.Error(x)
	}

	rangeInp := []byte(`{"Status":"ok",
		"Data":{"ResultType":"Range",
			"Result":[{"Metric":{"__name__":"b"},
			"Values":[[1,2]]}]}}`)
	h, err := DecodeRangeQ(rangeInp)
	if err != nil {
		t.Error(err)
	}
	pc.RangeInsert(h)
}

func TestFetching(t *testing.T) {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/api/v1/query", func(w http.ResponseWriter, r *http.Request) {
		inp := `{"Status":"ok",
			"Data":{"ResultType":"Range",
				"Result":[{"Metric":{"__name__":"b"},
				"Values":[[1,2],[2,3]]}]}}`
		fmt.Fprint(w, inp)
	})
	serveMux.HandleFunc("/api/v1/query_range", func(w http.ResponseWriter, r *http.Request) {
		inp := `{"Status":"ok",
			"Data":{"ResultType":"Range",
				"Result":[{"Metric":{"__name__":"b"},
				"Values":[[1,2]]}]}}`
		fmt.Fprint(w, inp)
	})
	series := `{"Status":"ok",
		"Data":[{"ResultType":"Range","A":"B"},{"Q":"R"}]}`

	serveMux.HandleFunc("/api/v1/series", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, series)
	})
	server := httptest.NewUnstartedServer(serveMux)
	server.Start()
	//reset internal client
	pc.Res = 0
	pc.Start()
	time.Sleep(1)

	// pull from ""
	pc.client = server.Client()
	pc.client.Timeout = time.Second
	pc.PullData()

	// pull from server.URL
	pc.P8s = server.URL
	n.Reset()
	pc.PullData()
	added := n.added
	scored := n.scored

	// pulling the same data should not add or score
	pc.PullData()
	if n.added != added || n.scored != scored {
		t.Errorf("Expected no more values added or scored. Before: %d, %d, after: %d, %d\n", added, scored, n.added, n.scored)
	}

	// pull from localhost:9090
	temp := pc.P8s
	pc.P8s = "localhost:9090"
	pc.PullData()
	pc.P8s = temp
	pc.Fetch("abcd")
	pc.PullData()
	series = `vb{"Status":"ok","Data":[]}`
	DecodeSeriesMatch([]byte(series))
	pc.PullData()
	series = `{"Status":"ok","Data":[]}`
	DecodeSeriesMatch([]byte(series))
	pc.PullData()
	server.Close()
}
