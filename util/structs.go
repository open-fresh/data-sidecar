package util

//Metric is a metric packet
type Metric struct {
	Desc map[string]string
	Data DataPoint
}

// DataPoint holds a time-value pair.
type DataPoint struct {
	Val  float64
	Time int64
}
