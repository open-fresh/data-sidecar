package scoring

import (
	"math"
	"testing"

	"github.com/open-fresh/data-sidecar/storage"
	"github.com/open-fresh/data-sidecar/util"
)

func TestHighway(t *testing.T) {
	currentValue := util.DataPoint{Val: float64(498)}
	length := 500
	data := make([]util.DataPoint, length)
	vals := make([]float64, length)
	wts := make([]float64, length)
	for ii := range vals {
		data[ii].Val = float64(ii)
		vals[ii] = float64(ii)
		wts[ii] = 1
	}
	tr := util.NewRecorder()
	store := storage.NewStore()

	Highway(currentValue, data, map[string]string{"a": "b", "__name__": "original"}, tr, store)
	close((*tr).Chan)
	Process(t, (*tr).Chan)

	tr = util.NewRecorder()
	names := make([]map[string]string, 5, 5)
	for ii := range names {
		names[ii] = make(map[string]string)
	}
	names[0]["__name__"] = "cpu"
	names[1]["__name__"] = "memory"
	names[2]["__name__"] = "network"
	names[3]["__name__"] = "cortex"
	names[4]["__name__"] = "disk"
	for ii := range names {
		Highway(currentValue, data, names[ii], tr, store)
	}
	close((*tr).Chan)

	tr = util.NewRecorder()
	Highway(currentValue, data, map[string]string{"a": "b", "__name__": "original"}, tr, store)
	close((*tr).Chan)
	Process(t, (*tr).Chan)
}

func Process(t *testing.T, trc chan util.Metric) {
	outputs := map[string]float64{
		"high:original": 4,
		"low:original":  144.4818327679989,
	}
	exits := map[string]float64{
		"low":     1,
		"high":    1,
		"outside": 1,
	}
	for res := range trc {
		x := res.Desc
		y := res.Data.Val
		metricName := x["__name__"]
		if metricName == "exit" {
			val, ok := exits[x["ft_model"]]
			if !ok {
				t.Error("not ok", x["ft_model"], y, exits)
			}
			if math.Abs(y-val) > 1e-8 {
				t.Error("wrong value", x, y, val, exits)
			}
		} else {
			val, ok := outputs[metricName]
			if !ok {
				t.Error(metricName, x, y, outputs)
			}
			if math.IsNaN(y) || math.IsInf(y, 0) {
				t.Error(x, val, y, outputs)
			}
		}
	}
}
