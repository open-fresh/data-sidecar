package storage

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Fresh-Tracks/data-sidecar/util"
)

func TestAddGet(t *testing.T) {
	t.Run("Add-Get1", func(t *testing.T) {
		x := NewStore()
		x.Add(map[string]string{"1": "1"}, 5.0, 1)
		y := x.Get(map[string]string{"1": "1"})
		t.Log(y)
		if y[0].Val != 5.0 {
			t.Error(len(y))
		}
		x.Add(map[string]string{"1": "1"}, math.NaN(), 2)
	})

	//make sure the elements go in as expected.
	t.Run("Add-Get2", func(t *testing.T) {
		x := NewStore()
		x.Add(map[string]string{"1": "1"}, 3.0, 3)
		y := x.Get(map[string]string{"1": "1"})
		t.Log(y)
		if y[0].Val != 3.0 {
			t.Error(y[0])
		}
		x.Add(map[string]string{"1": "1"}, 8.0, 4)
		y = x.Get(map[string]string{"1": "1"})
		t.Log(y)
		if y[1].Val != 8.0 {
			t.Error(y[1])
		}
	})

	//make sure the elements go in as expected.
	t.Run("Add-Get3", func(t *testing.T) {
		x := NewStore()
		for ii := 0; ii < 44; ii++ {
			x.Add(map[string]string{"1": "2"}, float64(ii), int64(ii)+5)
		}
		y := x.Get(map[string]string{"1": "2"})
		for ii := 0; ii < 22; ii++ {
			if y[ii].Val != float64(ii)+22 {
				t.Error("mismatch")
			}
		}
	})

	// Make sure duplicates are not added
	t.Run("Add-Get-Duplicates", func(t *testing.T) {
		// Use a new store to make sure data is clean
		x := NewStore()
		kvs1 := map[string]string{"duplicates": "1"}
		kvs2 := map[string]string{"duplicates": "0"}
		x.Add(kvs1, 5.0, 1)
		x.Add(kvs1, 3.0, 1)
		x.Add(kvs2, 3.0, 1)

		y := x.Get(kvs1)
		if (len(y) != 1) || (y[0].Val != 5.0) {
			t.Errorf("Duplicates added: y = %v", y)
		}
		z := x.Get(kvs2)
		if len(z) != 1 || z[0].Val != 3.0 {
			t.Fail()
		}
		x.Add(kvs1, math.NaN(), 2)
		x.Add(kvs2, math.NaN(), 2)
	})

}
func TestAddGetWrap(t *testing.T) {
	// tests the ring behavior, namely that it should wrap eventually
	t.Run("Add-Get3", func(t *testing.T) {
		x := NewStore()
		for xx := 0; xx < 1000; xx++ {
			x.Add(map[string]string{"1": "1"}, 1.0, int64(xx)+1)
		}
		y := x.Get(map[string]string{"1": "1"})
		t.Log(y)
		if (len(y) != max) || (y[0].Val != 1.0) {
			t.Fail()
		}
		prevlen := len(y)
		x.Add(map[string]string{"1": "1"}, 6.0, 2)
		y = x.Get(map[string]string{"1": "1"})
		if len(y) != prevlen {
			t.Fail()
		}

	})

	// tests the concurrency behavior, namely that it should wrap eventually
	t.Run("Add-Get4", func(t *testing.T) {
		x := NewStore()
		go func() {
			for xx := 0; xx < 1000; xx++ {
				x.Add(map[string]string{"1": "1"}, 1.0, int64(xx)+1)
			}
			y := x.Get(map[string]string{"1": "1"})
			if (len(y) != max) || (y[0].Val != 1.0) {
				t.Fail()
			}
			prevlen := len(y)
			x.Add(map[string]string{"1": "1"}, 6.0, 1001)
			y = x.Get(map[string]string{"1": "1"})
			if len(y) != prevlen {
				t.Fail()
			}
		}()
		for xx := 0; xx < 10000; xx++ {
			x.Add(map[string]string{"1": "1"}, 1.0, int64(xx)+1)
		}
		y := x.Get(map[string]string{"1": "1"})
		t.Log(y)
		if (len(y) != max) || (y[0].Val != 1.0) {
			t.Fail()
		}
		prevlen := len(y)
		x.Add(map[string]string{"1": "1"}, 6.0, 1000000)
		y = x.Get(map[string]string{"1": "1"})
		if len(y) != prevlen {
			t.Fail()
		}
	})
	t.Run("DumpMap", func(t *testing.T) {
		x := NewStore()
		x.Add(map[string]string{"1": "1"}, 3.0, 3)
		x.Add(map[string]string{"1": "1"}, 1.0, 5)

		if val, ok := x.DumpMap()[util.MapSSToS(map[string]string{"1": "1"})]; ok {
			if val[0].Val != 3 {
				t.Error(val)
			}
			if val[len(val)-1].Val != 1 {
				t.Error(val)
			}

		} else {
			t.Error(val)
		}
	})
}

func TestKnown(t *testing.T) {
	x := NewStore()
	x.Add(map[string]string{"1": "1"}, 3.0, 3)
	x.Add(map[string]string{"1": "2"}, 1.0, 5)

	g := x.UsedKeys()
	if len(g) == 0 {
		t.Error()
	} else {
		x.Delete(g[0])
		x.Add(map[string]string{"blast": "oldone"}, 0.0, 0)
		x.Delete(g[0])
		h := x.UsedKeys()
		// delete same key twice should be fine, but we added one, so should
		// be the same length.
		if len(g) != len(h) {
			t.Error()
		}
	}
}

func TestSerDe(t *testing.T) {
	x := NewStore()
	x.Add(map[string]string{"1": "1"}, 3.0, 3)
	x.Add(map[string]string{"1": "2"}, 1.0, 5)

	t.Run("serialize", func(t *testing.T) {
		y := RingDeserialize(x.RingSerialize())
		z := RingDeserialize(y.RingSerialize())
		test := []map[string]string{map[string]string{"1": "1"}, map[string]string{"1": "2"}, map[string]string{"2": "3"}, map[string]string{"2": "4"}}
		for _, yy := range test {
			vx := z.Get(yy)
			vy := y.Get(yy)
			t.Log(yy, vx, vy, z.Data, y.Data)
			if len(vx) != len(vy) {
				t.Fail()
			} else {
				for ii, xx := range vx {
					if vy[ii] != xx {
						t.Fail()
					}
				}

			}
		}

	})

}

func TestHandlers(t *testing.T) {
	x := NewStore()
	x.Add(map[string]string{"1": "1"}, 3.0, 3)
	x.Add(map[string]string{"1": "2"}, 1.0, 5)

	rw := util.NewHTTPResponseWriter()
	r := &http.Request{Form: url.Values{}}
	x.DumpHandleFunc(rw, r)
	if g := rw.String(); !strings.Contains(g, ",") {
		t.Error(g)
	}
}

func TestPrune(t *testing.T) {
	x := NewStore()
	for xx := 0; xx < 1000; xx++ {
		x.Add(map[string]string{"1": "1"}, 1.0, time.Now().Unix())
	}
	g := x.Prune(10)
	if len(g) > 0 {
		t.Error(g)
	}

	for xx := 0; xx < 1000; xx++ {
		x.Add(map[string]string{"1": "1"}, 1.0, time.Now().Unix())
	}
	g = x.Prune(-10)
	if len(g) == 0 {
		t.Error("none deleted")
	}
}

func BenchmarkAdd(b *testing.B) {
	x := NewStore()

	for i := 0; i < b.N; i++ {
		x.Add(map[string]string{"1": fmt.Sprintf("%d", (i % 2))}, float64(i), int64(i))
	}
}
func BenchmarkGetNull(b *testing.B) {
	x := NewStore()

	for i := 0; i < b.N; i++ {
		a := x.Get(map[string]string{"1": "4"})
		_ = a
	}
}
func BenchmarkGetReal(b *testing.B) {
	x := NewStore()

	for i := 0; i < b.N; i++ {
		a := x.Get(map[string]string{"1": "1"})
		_ = a
	}
}
