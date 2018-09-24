// This is a the functionality from gonum.stat that we actually need, optimized for what we
// actually do.
package stat

import (
	"math"
	"time"
)

// SuffStat holds a sufficient statistic
type SuffStat struct {
	Count          float64 // Just add 1 for eachthing that comes in
	Sx             float64 // Adequate to compute mean with the above
	Sx2            float64 // Adequate to compute variance, etc.
	Sx3            float64 // Adequate to compute skewness
	LastInsertTime int64   // This one is for use with table pruning.
}

// DistParams holds calculated parameters for a distribution
type DistParams struct {
	Mean float64
	Std  float64
}

// NewSuffStat generates a new sufficient statistic container
func NewSuffStat() *SuffStat {
	return &SuffStat{}
}

// Copy copies a suffstat into a new suffstat.
func (s *SuffStat) Copy() *SuffStat {
	out := NewSuffStat()
	out.Count = s.Count
	out.Sx = s.Sx
	out.Sx2 = s.Sx2
	out.Sx3 = s.Sx3
	out.LastInsertTime = s.LastInsertTime
	return out
}

// Insert a value into a sufficient statistic
func (s *SuffStat) Insert(val float64) {
	s.Count++
	s.Sx += val
	s.Sx2 += val * val
	s.Sx3 += val * val * val
	s.LastInsertTime = time.Now().Unix()
}

// Remove a value from a sufficient statistic
func (s *SuffStat) Remove(val float64) {
	s.Count--
	s.Sx -= val
	s.Sx2 -= val * val
	s.Sx3 -= val * val * val
	s.LastInsertTime = time.Now().Unix()
}

// Combine two sufficient stats into the base.
func (s *SuffStat) Combine(o *SuffStat, lwt, rwt float64) *SuffStat {
	g := NewSuffStat()
	g.Count = s.Count*lwt + o.Count*rwt
	g.Sx = s.Sx*lwt + o.Sx*rwt
	g.Sx2 = s.Sx2*lwt + o.Sx2*rwt
	g.Sx3 = s.Sx3*lwt + o.Sx3*rwt
	g.LastInsertTime = time.Now().Unix()
	return g
}

// MeanStdDev calculates the mean and standard deviation from sufficient statistics.
func (s *SuffStat) MeanStdDev() (mean, std float64) {
	mean = s.Sx / (s.Count + 1e-12)
	std = math.Sqrt((s.Sx2 - mean*s.Sx) / (s.Count + 1e-12))
	return
}

// MeanStdDevSkew calculates the mean and standard deviation from sufficient statistics.
func (s *SuffStat) MeanStdDevSkew() (mean, std, skew float64) {
	mean, std = s.MeanStdDev()
	skew = s.Sx3/s.Count - 3*mean*std*std - mean*mean*mean
	skew /= (std*std*std + 1e-12)
	return
}

// MeanStdDev is a passthrough to gonum.stat
func MeanStdDev(x []float64) (mean, std float64) {
	g := NewSuffStat()
	for _, xx := range x {
		g.Insert(xx)
	}
	mean, std = g.MeanStdDev()
	return
}

// Quantile computes the quantile
func Quantile(p float64, x []float64) (quantile float64) {
	var cumsum float64
	sumWeights := float64(len(x))
	fidx := p * sumWeights
	for i := range x {
		cumsum++
		if cumsum >= fidx {
			return x[i]
		}
	}
	return x[len(x)-1]
}
