package scoring

import (
	"math"
	"testing"

	"github.com/Fresh-Tracks/data-sidecar/util"
)

func TestRecordThreshold(t *testing.T) {
	labels := map[string]string{
		"__name__":  "greatMetric",
		"foo":       "bar",
		"ft_target": "true",
	}
	tr := util.NewRecorder()
	RecordThreshold(util.DataPoint{Val: 100}, labels, "greatModel", tr)
	close((*tr).Chan)
	calls := 0
	for x := range (*tr).Chan {
		mockLabels := x.Desc
		calls++
		if mockLabels["foo"] != "bar" {
			t.Error("foo != bar")
		}
		if mockLabels["__name__"] != "greatModel:greatMetric" {
			t.Error("__name__ != greatModel:greatMetric")
		}
		if _, ok := mockLabels["ft_target"]; ok {
			t.Error("ft_target not filtered")
		}
	}
	if calls != 1 {
		t.Error("callCount != 1:", calls)
	}
}

func TestRecordExit(t *testing.T) {
	labels := map[string]string{
		"__name__":  "greatMetric",
		"foo":       "bar",
		"ft_target": "true",
	}
	tr := util.NewRecorder()
	RecordExit(true, 1, labels, "greatModel", tr)
	close((*tr).Chan)
	calls := 0
	for x := range (*tr).Chan {
		mockLabels := x.Desc
		mockValue := x.Data.Val
		calls++

		if mockLabels["foo"] != "bar" {
			t.Error("foo != bar")
		}

		if mockLabels["__name__"] != "exit" {
			t.Error("__name__ != exit")
		}

		if _, ok := mockLabels["ft_target"]; ok {
			t.Error("ft_target not filtered")
		}

		if math.Abs(1-mockValue) > 1e-8 {
			t.Error("value not 1")
		}
	}

	if calls != 1 {
		t.Log("callCount != 1:", calls)
	}
	tr = util.NewRecorder()
	RecordExit(false, 1, labels, "greatModel", tr)
	close((*tr).Chan)
	calls = 0
	for x := range (*tr).Chan {
		mockValue := x.Data.Val
		if !math.IsNaN(mockValue) {
			t.Error("value not NaN")
		}
	}
}
