package icarus

import (
	"testing"

	"github.com/Fresh-Tracks/data-sidecar/util"
)

func SuiteTestStore(t *testing.T, x *IcarusStore, cap int) {
	metric := util.Metric{Desc: map[string]string{"a": "B", "__name__": "hello"}, Data: util.DataPoint{Val: 1.}}
	t.Run("in-out loop", func(t *testing.T) {
		x.Insert(metric)
		if metric.Data.Val != x.Dump()[0].Data.Val {
			t.Error()
		}
	})
	t.Run("aging keeps for a bit", func(t *testing.T) {
		x.Roll()
		if metric.Data.Val != x.Dump()[0].Data.Val {
			t.Error()
		}
	})
	t.Run("not forever", func(t *testing.T) {
		for i := 1; i < cap; i++ {
			x.Roll()
		}
		if 0 != len(x.Dump()) {
			t.Error()
		}
	})

}

func TestRollingStore(t *testing.T) {
	g := NewRollingStore(2)
	SuiteTestStore(t, g, 2)
}
