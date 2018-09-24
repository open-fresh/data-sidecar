package stat

import (
	"testing"
)

func TestQuantile(t *testing.T) {
	if Quantile(0.5, []float64{1, 2, 3, 4, 5}) > 4 {
		t.Error()
	}

}

func TestMeanStdDev(t *testing.T) {
	if m, st := MeanStdDev([]float64{1, 2, 3, 4, 5}); (m < 2.99) || (m > 3.01) || (st < 1.41) || (st > 1.42) {
		t.Error(m, st)
	}

}

func TestMeanStdDevSkew(t *testing.T) {
	// uniform
	s := NewSuffStat()
	for _, x := range []float64{1, 2, 3, 4, 5} {
		s.Insert(x)
	}
	if _, _, sk := s.MeanStdDevSkew(); sk*sk > 1e-12 {
		t.Error(sk)
	}

	// right-skewed
	s = NewSuffStat()
	for _, x := range []float64{1, 2, 3, 4, 5} {
		s.Insert(x * x)
	}
	if _, _, sk := s.MeanStdDevSkew(); sk*sk < 0.4692*0.4692 {
		t.Error(sk)
	}

	// right-skewed
	s = NewSuffStat()
	for _, x := range []float64{1, 2, 3, 4, 5} {
		s.Insert(-x * x)
	}
	if _, _, sk := s.MeanStdDevSkew(); sk*sk < -0.4692*0.4692 {
		t.Error(sk)
	}

}
