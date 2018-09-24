package util

// Recorder is the type of whatever ultimately takes these metrics
type Recorder interface {
	Record(Metric)
	Finish()
}

// ScoringEngine is whatever scores datapoints.
type ScoringEngine interface {
	Add(map[string]string, float64, int64) bool
	Score(map[string]string)
	ScoreData([]DataPoint, map[string]string, bool)
}

// StorageEngine is whatever handles the data work.
type StorageEngine interface {
	Add(map[string]string, float64, int64) bool
	Get(map[string]string) []DataPoint
	UsedKeys() []string
}
