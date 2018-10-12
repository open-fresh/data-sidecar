package storage

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/open-fresh/data-sidecar/util"
)

const (
	max = 22
)

// StoreDetails contains the destination addresses for a data store.
type storeDetails struct {
	Key        string
	Index      int
	Full       bool
	Last       int64
	Meta       map[string]string
	MetaString string
	Data       [max]util.DataPoint
	LastVal    float64
}

// Store contains the individual records.
type Store struct {
	*sync.Mutex
	Data map[string]storeDetails
}

// NewStore returns a Ring Store
func NewStore() *Store {
	var mux sync.Mutex
	s := Store{&mux, nil}
	s.Data = make(map[string]storeDetails)
	return &s
}

// UsedKeys reports the used keys in the store.
func (s *Store) UsedKeys() (out []string) {
	s.Lock()
	defer s.Unlock()
	for xx, yy := range s.Data {
		if (yy.Index > 0) || (yy.Full) {
			out = append(out, xx)
		}
	}
	return
}

// PopulatedKeys reports the used keys in the store.
func (s *Store) PopulatedKeys(howFull int) (out []string) {
	s.Lock()
	defer s.Unlock()
	for xx, yy := range s.Data {
		if (yy.Index > howFull) || (yy.Full) {
			out = append(out, xx)
		}
	}
	return
}

// Add a value to a series for a key
func (s *Store) Add(kvs map[string]string, val float64, dataTime int64) bool {
	if math.IsNaN(val) {
		return false
	}
	key := util.MapSSToS(kvs)
	s.Lock()
	defer s.Unlock()
	_, ok := s.Data[key]
	if !ok {
		store := [max]util.DataPoint{}
		label, _ := json.Marshal(kvs)
		base := storeDetails{key, 0, false, -1, kvs, string(label), store, val}
		s.Data[key] = base
	}
	// Do not add anything unless it is new
	if s.Data[key].Last >= dataTime {
		return false
	}
	temp := s.Data[key]
	temp.Data[temp.Index] = util.DataPoint{Val: val, Time: dataTime}
	temp.Last = dataTime
	if s.Data[key].Index+1 == max {
		temp.Full = true
	}
	temp.Index = (s.Data[key].Index + 1) % max
	temp.LastVal = val
	s.Data[key] = temp
	return true
}

// Get a series for a key
func (s *Store) Get(kvs map[string]string) []util.DataPoint {
	key := util.MapSSToS(kvs)
	s.Lock()
	defer s.Unlock()
	return s.get(key)
}

// get a series for a key
func (s *Store) get(key string) []util.DataPoint {
	count := s.Data[key].Index
	start := 0
	if s.Data[key].Full {
		count = max
		start = s.Data[key].Index
	}
	out := make([]util.DataPoint, count)
	for ii := 0; ii < count; ii++ {
		out[ii] = s.Data[key].Data[(ii+start)%max]
	}
	return out
}

// Delete removes a key from the store
func (s *Store) Delete(key string) bool {
	s.Lock()
	defer s.Unlock()
	delete(s.Data, key)
	return true
}

// Prune kills all entries older than a certain age from the store
func (s *Store) Prune(secs int) map[string]bool {
	cutoff := time.Now().Unix() - int64(secs)
	s.Lock()
	killList := make(map[string]bool)
	for key, val := range s.Data {
		if val.Last < cutoff {
			killList[key] = true
		}
	}
	s.Unlock()
	for key := range killList {
		s.Delete(key)
	}
	return killList
}

// RingSerialize takes a ringStore and turns it to bytes0
func (s *Store) RingSerialize() []byte {
	s.Lock()
	network, _ := json.Marshal(s)
	s.Unlock()
	return network
}

// RingDeserialize does the usual deserialization magic on a Ring.
func RingDeserialize(x []byte) Store {
	var mux sync.Mutex
	s := Store{&mux, nil}
	json.Unmarshal(x, &s)
	return s
}

// DumpMap dumps a map representation of the data in the store.
func (s *Store) DumpMap() map[string][]util.DataPoint {
	known := make(map[string]bool)
	s.Lock()
	defer s.Unlock()
	for key := range s.Data {
		known[key] = true
	}
	series := make(map[string][]util.DataPoint)
	for key := range known {
		series[key] = s.get(key)
	}
	return series
}

// DumpStruct handles data dump formatting.
type DumpStruct struct {
	Key  string
	Data []float64
}

// DataDump drops the whole table into dumpstruct format
func (s *Store) DataDump() map[string]DumpStruct {
	keys := s.UsedKeys()
	proto := make(map[string]DumpStruct)
	s.Lock()
	defer s.Unlock()
	for _, key := range keys {
		temp := s.get(key)
		out := make([]float64, len(temp), len(temp))
		for ii, xx := range temp {
			out[ii] = xx.Val
		}
		proto[key] = DumpStruct{key, out}
	}
	return proto
}

// DumpHandleFunc lets you query for all the names and all the things for each of them.
func (s *Store) DumpHandleFunc(w http.ResponseWriter, r *http.Request) {
	proto := s.DataDump()
	out, _ := json.Marshal(proto)
	fmt.Fprint(w, string(out))
	return
}
