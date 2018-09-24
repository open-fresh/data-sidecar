package scoring

import (
	"github.com/Fresh-Tracks/data-sidecar/stat"
	"github.com/Fresh-Tracks/data-sidecar/util"
)

// HighwayVal is the kind of value a highway can hold
type HighwayVal struct {
	High float64
	Low  float64
}

// HighwayExits is the kind of value that exits can be/hold
type HighwayExits struct {
	High bool
	Low  bool
}

// Highway adds green highway data based on a histogram
func Highway(curr util.DataPoint, data []util.DataPoint, kvs map[string]string,
	record util.Recorder, storage util.StorageEngine) {

	if len(data) < 20 {
		return
	}

	// put your favorite math here!
	tempSS := stat.NewSuffStat()
	for xx := range data {
		tempSS.Insert(data[xx].Val)
	}
	mean, std := tempSS.MeanStdDev()
	hwy := HighwayVal{High: mean + 3.*std, Low: mean - 3.*std}
	// replace the above calculation with whatever you like
	// to generate upper and lower bounds
	// anything will do, you can even break it out by series
	// names or characteristics or whatever!

	hwy.Record(curr, kvs, record)

	exits := HighwayExits{High: curr.Val > hwy.High,
		Low: curr.Val < hwy.Low}
	exits.Record(curr, kvs, record)

}

// Record records all the relevant exits for a given highway
func (e HighwayExits) Record(curr util.DataPoint, kvs map[string]string, record util.Recorder) {
	RecordExit(e.High, curr.Time, kvs, "high", record)
	RecordExit(e.Low, curr.Time, kvs, "low", record)
	outside := e.Low || e.High
	RecordExit(outside, curr.Time, kvs, "outside", record)
}

// Record gets an entire map of quantile map and the rest and records them all
func (h HighwayVal) Record(curr util.DataPoint, kvs map[string]string, record util.Recorder) {
	RecordThreshold(util.DataPoint{Val: h.High, Time: curr.Time}, kvs, "high", record)
	RecordThreshold(util.DataPoint{Val: h.Low, Time: curr.Time}, kvs, "low", record)
}
