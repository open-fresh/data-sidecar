package scoring

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	"github.com/Fresh-Tracks/data-sidecar/scoring/anomaly"
	"github.com/Fresh-Tracks/data-sidecar/storage"
	"github.com/Fresh-Tracks/data-sidecar/util"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	modelDurationSummary = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "sidecar_model_duration_summary",
		Help: "Summary of model execution duration"},
		[]string{"type"})
	modelStorageSummary = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "sidecar_model_storage_summary",
		Help: "Summary of model storage sizes"},
		[]string{"type"})
)

func init() {
	prometheus.MustRegister(modelDurationSummary)
	prometheus.MustRegister(modelStorageSummary)
}

// Scorer holds the things the scorer needs for its work. A place to store data, a place to send it, and a way to learn about it.
type Scorer struct {
	storage util.StorageEngine
	record  util.Recorder
}

// NewScorer returns a pointer to a scorer.
func NewScorer(store util.StorageEngine, record util.Recorder) *Scorer {
	return &Scorer{store, record}

}

// Add is a passthrough
func (s *Scorer) Add(kvs map[string]string, val float64, time int64) bool {
	return s.storage.Add(kvs, val, time)
}

// Score tells the scorer that you're done adding points right now and to score the item.
func (s *Scorer) Score(kvs map[string]string) {
	ScoreItem(kvs, s.record, s.storage)
}

type sortInfo struct {
	Sort uint64
	Name uint64
	Work uint64
	Key  uint64
}

type infoCollect struct {
	Array []sortInfo
}

// Len is sorting infrastructure
func (a infoCollect) Len() int { return len(a.Array) }

// Swap is sorting infrastucture
func (a infoCollect) Swap(i, j int) { a.Array[i], a.Array[j] = a.Array[j], a.Array[i] }
func (a infoCollect) Less(i, j int) bool {
	return (a.Array[i].Sort < a.Array[j].Sort) && (a.Array[i].Name < a.Array[j].Name) && (a.Array[i].Work < a.Array[j].Work)
}

// ModelTimer times model evaluations.
func ModelTimer(name string, model func()) {
	timer := prometheus.NewTimer(modelDurationSummary.WithLabelValues(name))
	defer timer.ObserveDuration()
	model()
}

// ScoreItem scores individual time series
func ScoreItem(labels map[string]string, destination util.Recorder, store util.StorageEngine) {
	data := store.Get(labels)

	if (data == nil) || (len(data) <= 1) {
		return
	}

	currentValue := data[len(data)-1]
	ModelTimer("highway", func() {
		Highway(currentValue, data, labels, destination, store)
	})
	lookbackPoints := 30
	if len(data) <= lookbackPoints {
		lookbackPoints = len(data) - 1
	}
	vals := make([]float64, lookbackPoints, lookbackPoints)
	for ii := range vals {
		vals[ii] = data[ii].Val
	}
	ModelTimer("nelsonRules", func() {
		anoms := anomaly.Nelson(vals, labels)
		for _, x := range anoms {
			destination.Record(util.Metric{Desc: x, Data: util.DataPoint{Val: 1.0, Time: currentValue.Time}})
		}
	})
}

// ScoreOutput will help marshal scoring output.
type ScoreOutput struct {
	Key  map[string]string
	Data [][]float64
}

// ScoreHandleFunc lets you query for all the names and all the things for each of them.
func (s *Scorer) ScoreHandleFunc(w http.ResponseWriter, r *http.Request) {
	preData := (*r).FormValue("data")
	preInfo := (*r).FormValue("info")
	if preData == "" {
		fmt.Fprint(w, "Please query with data=<[]float64 data> info=<map[string]string>")
		return
	}
	var data []float64
	err := json.Unmarshal([]byte(preData), &data)
	if err != nil {
		fmt.Fprint(w, fmt.Sprintf("invalid data %s, %s", preData, err))
		return
	}
	var info map[string]string
	if len(preInfo) > 0 {
		err = json.Unmarshal([]byte(preInfo), &info)
		if err != nil {
			fmt.Fprint(w, fmt.Sprintf("invalid info %s, %s", preInfo, err))
			return
		}
	}
	useOut := ScoreOverTime(data, info)
	output, _ := json.Marshal(useOut)
	fmt.Fprint(w, string(output))
	return
}

func (s *Scorer) ScoreData(data []util.DataPoint, kvs map[string]string, lastOnly bool) {
	ScoreRange(data, kvs, s.record, s.storage, lastOnly)
}

// ScoreRange is the main scoring loop for ranges.
func ScoreRange(data []util.DataPoint, kvs map[string]string, recorder util.Recorder, store util.StorageEngine, lastOnly bool) {
	null := util.NewNullRecorder()
	for time := range data {
		mydata := make([]util.DataPoint, time, time)
		for ii := range mydata {
			mydata[ii] = data[ii]
		}
		store.Add(kvs, data[time].Val, data[time].Time)
		if lastOnly && (time != len(data)-1) {
			ScoreItem(kvs, null, store)
		} else {
			ScoreItem(kvs, recorder, store)
		}
	}
	recorder.Finish()
}

// ScoreOverTime scores an individual series
func ScoreOverTime(data []float64, kvs map[string]string) []ScoreOutput {
	store := storage.NewStore()
	output := make([]ScoreOutput, 0)
	temp := make(map[string]ScoreOutput)
	recorder := util.NewRecorder()
	mydata := make([]util.DataPoint, 0, len(data))
	for ii := range data {
		temp := util.DataPoint{}
		temp.Val = data[ii]
		if math.IsNaN(data[ii]) || math.IsInf(data[ii], 0) {
			continue
		}
		temp.Time = int64(ii)
		mydata = append(mydata, temp)
	}
	go ScoreRange(mydata, kvs, recorder, store, false)
	time := 0
	for x := range recorder.Chan {
		if math.IsNaN(x.Data.Val) {
			continue
		}
		loc := util.MapSSToS(x.Desc)
		_, ok := temp[loc]
		if !ok {
			temp[loc] = ScoreOutput{x.Desc, make([][]float64, 0)}
		}
		val := temp[loc]
		val.Data = append(val.Data, []float64{float64(x.Data.Time), x.Data.Val})
		temp[loc] = val
		time++
	}
	for _, val := range temp {
		output = append(output, val)
	}
	return output
}
