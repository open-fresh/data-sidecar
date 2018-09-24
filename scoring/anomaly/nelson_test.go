package anomaly

import (
	"testing"
)

func reverse(inp []float64) (out []float64) {
	out = make([]float64, len(inp), len(inp))
	for ii := range inp {
		out[ii] = inp[len(inp)-1-ii]
	}
	return
}

func TestNelson(t *testing.T) {
	length := 500
	vals := make([]float64, length)
	wts := make([]float64, length)
	distant := make([]float64, length)
	center := make([]float64, length)
	alternate := make([]float64, length)
	for ii := range vals {
		vals[ii] = float64(ii)
		wts[ii] = float64(ii) * float64(length-ii)
		distant[ii] = 1
		center[ii] = float64(length) / 2
		alternate[ii] = float64(ii % 2)
	}
	distant[length-1] = -10000

	t.Run("nelson large occ", func(t *testing.T) {
		didNelson1 := false
		name := map[string]string{"a": "n1", "ft_target": "fake"}
		if NelsonLargeOoC([]float64{0, 0, 0, 0, 0}, 0, 0) {
			t.Error()
		}
		output := Nelson(distant, name)
		for _, x := range output {
			highway := x["ft_model"]
			if highway == "nelson_large_ooc" {
				didNelson1 = true
			}

		}
		if !didNelson1 {
			t.Error(didNelson1)
		}
	})
	t.Run("nelson medium ooc", func(t *testing.T) {
		didNelson5 := false
		name := map[string]string{"a": "n5"}
		testpts := []float64{-2, -2, -2, -2, -2, -2, -2, -2, -2, -2, -2, -2, -2, -2, -2, 10, 10, 10}

		if NelsonMediumOoC([]float64{}, 0, 0) {
			t.Error()
		}

		if NelsonMediumOoC([]float64{0, 0, 0, 0}, 0, 0) {
			t.Error()
		}
		output := Nelson(testpts, name)

		for _, x := range output {

			highway := x["ft_model"]
			if highway == "nelson_medium_ooc" {
				didNelson5 = true
			}
		}
		if !didNelson5 {
			t.Fail()
		}
	})
	t.Run("nelson small ooc", func(t *testing.T) {
		didNelson6 := false
		name := map[string]string{"a": "n6"}
		testpts := reverse([]float64{1.1, 1.1, 1.1, 1.1, 0.5, 0.5, 0.5, 0.5, 1, 1, 1, 1, 0.5, 0.5, 0.5, 0.5})
		if NelsonSmallOoC([]float64{}, 0, 0) {
			t.Error()
		}
		if NelsonSmallOoC([]float64{00, 0, 0, 0, 0, 0}, 0, 0) {
			t.Error()
		}

		output := Nelson(testpts, name)

		for _, x := range output {

			highway := x["ft_model"]
			if highway == "nelson_small_ooc" {
				didNelson6 = true
			}
		}
		if !didNelson6 {
			t.Fail()
		}
	})
}
