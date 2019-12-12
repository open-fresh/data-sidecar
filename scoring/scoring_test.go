package scoring

import (
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/open-fresh/data-sidecar/storage"
	"github.com/open-fresh/data-sidecar/util"
)

func withJitter(n float64) float64 {
	return n + 50*rand.Float64() - 5
}

func TestScore(t *testing.T) {
	x := storage.NewStore()

	for i := 0; i < 10000; i++ {
		if rand.Float64() > 0.5 {
			x.Add(map[string]string{"top": "1"}, rand.ExpFloat64(), int64(i))
		} else {
			x.Add(map[string]string{"bottom": "1"}, rand.NormFloat64(), int64(i))
		}
	}
	x.Add(map[string]string{"top": "g"}, 1, 0)
	x.Add(map[string]string{"_ft": "hello"}, 1, 0)
	x.Add(map[string]string{"g": "small"}, 1, 0)
	t.Run("score", func(t *testing.T) {
		rec := util.NewRecorder()
		sc := NewScorer(x, rec)
		tmp := map[string]string{"g": "small"}
		sc.Add(tmp, 1, 1)
		sc.Add(tmp, 1, 2)
		sc.Score(tmp)
		sc.Add(tmp, 1, 3)
		sc.Add(tmp, 1, 4)
		sc.Add(tmp, 1, 5)
		sc.Score(tmp)
		close(rec.Chan)
		somethingCameBack := false
		for g := range rec.Chan {
			t.Log(g)
			somethingCameBack = true
		}
		if somethingCameBack {
			t.Error()
		}
	})
	t.Run("scoreitem", func(t *testing.T) {
		rec := util.NewRecorder()
		store := storage.NewStore()
		store.Add(map[string]string{"a": "b"}, 1., 1)
		store.Add(map[string]string{"a": "b"}, 2., 2)
		store.Add(map[string]string{"a": "b"}, 3., 3)
		ScoreItem(map[string]string{"a": "b"}, rec, store)

		close(rec.Chan)
		somethingCameBack := false
		for _ = range rec.Chan {
			somethingCameBack = true
		}
		if somethingCameBack {
			t.Error()
		}

		rec = util.NewRecorder()

		store = storage.NewStore()
		store.Add(map[string]string{"a": "b"}, 1., 1)
		store.Add(map[string]string{"a": "b"}, 2., 2)
		store.Add(map[string]string{"a": "b"}, 3., 3)
		store.Add(map[string]string{"a": "b"}, 5., 5)
		store.Add(map[string]string{"a": "b"}, 6., 6)
		store.Add(map[string]string{"a": "b"}, 7., 7)
		store.Add(map[string]string{"a": "b"}, 8., 8)
		ScoreItem(map[string]string{"a": "b"}, rec, store)
		close(rec.Chan)
		somethingCameBack = false
		for _ = range rec.Chan {
			somethingCameBack = true
		}
		if somethingCameBack {
			t.Error()
		}
	})
}

func TestHandlers(t *testing.T) {
	x := storage.NewStore()
	rec := util.NewRecorder()
	sc := NewScorer(x, rec)
	t.Run("score", func(t *testing.T) {
		rw := util.NewHTTPResponseWriter()
		r := &http.Request{}
		sc.ScoreHandleFunc(rw, r)
		if g := rw.String(); !strings.Contains(g, "Please query with") {
			t.Error(g)
		}
		rw = util.NewHTTPResponseWriter()
		r = &http.Request{Form: url.Values{"data": []string{"[1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5]"}}}
		sc.ScoreHandleFunc(rw, r)
		if g := rw.String(); !strings.Contains(g, "Data") {
			t.Error(g)
		}
		rw = util.NewHTTPResponseWriter()
		r = &http.Request{Form: url.Values{"data": []string{"[[1,2,3,4,5]]"}}}
		sc.ScoreHandleFunc(rw, r)
		if g := rw.String(); !strings.Contains(g, "[[") {
			t.Error(g)
		}
		rw = util.NewHTTPResponseWriter()
		r = &http.Request{Form: url.Values{"data": []string{"[1,2,3,4,5]"}, "info": []string{`{"__name__":"hello}`}}}
		sc.ScoreHandleFunc(rw, r)
		if g := rw.String(); !strings.Contains(g, "invalid info") {
			t.Error(g)
		}
		rw = util.NewHTTPResponseWriter()
		r = &http.Request{Form: url.Values{"data": []string{"[1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5,1,2,3,4,5]"}, "info": []string{`{"__name__":"hello"}`}}}
		sc.ScoreHandleFunc(rw, r)
		if g := rw.String(); !strings.Contains(g, "Data") || !strings.Contains(g, "hello") {
			t.Error(g)
		}
	})
}
