package icarus

import (
	"sync"

	"github.com/Fresh-Tracks/data-sidecar/util"
)

// IcarusStore holds sets of metrics and retires them as necessary.
type IcarusStore struct {
	*sync.Mutex
	Keep    int
	Index   int
	Metrics []map[string]util.Metric
}

// Get back a new implementation of the rolling store
func NewRollingStore(lookback int) *IcarusStore {
	var mux sync.Mutex
	out := IcarusStore{&mux, lookback,
		0, make([]map[string]util.Metric, lookback, lookback)}
	for ii := range out.Metrics {
		out.Metrics[ii] = make(map[string]util.Metric)
	}
	return &out
}

// Roll the rolling store
func (r *IcarusStore) Roll() {
	r.Lock()
	defer r.Unlock()
	r.Index = (r.Index + 1) % r.Keep
	r.Metrics[r.Index] = make(map[string]util.Metric)
}

// Insert something into the current store in the rolling store
func (r *IcarusStore) Insert(met util.Metric) {
	r.Lock()
	defer r.Unlock()
	label := util.MapSSToS(met.Desc)
	r.Metrics[r.Index][label] = met
}

// Dump all the []Metrics in the rolling store.
func (r *IcarusStore) Dump() []util.Metric {
	r.Lock()
	defer r.Unlock()
	temp := make(map[string]util.Metric)
	for ii := 1; ii <= r.Keep; ii++ {
		loc := (r.Index + ii) % r.Keep
		for key, val := range r.Metrics[loc] {
			temp[key] = val
		}
	}
	out := make([]util.Metric, len(temp), len(temp))
	index := 0
	for _, val := range temp {
		out[index] = val
		index++
	}
	return out
}
