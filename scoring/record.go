package scoring

import (
	"fmt"
	"math"
	"strings"

	"github.com/Fresh-Tracks/data-sidecar/util"
)

func filterLabels(labels map[string]string) map[string]string {
	filtered := make(map[string]string)
	for key, val := range labels {
		if key == "ft_target" {
			continue
		}

		filtered[key] = val
	}
	return filtered
}

func thresholdLabels(labels map[string]string, model string) map[string]string {
	thresholdLabels := filterLabels(labels)
	thresholdLabels["__name__"] = fmt.Sprintf("%v:%v", strings.Replace(model, ".", "_", -1), labels["__name__"])
	return thresholdLabels
}

// RecordThreshold records a threshold value
func RecordThreshold(val util.DataPoint, labels map[string]string, model string, record util.Recorder) {
	record.Record(util.Metric{Desc: thresholdLabels(labels, model), Data: val})
}

// RecordExit records NaN when a value is within a threshold and 1 when a value is outside of a threshold
func RecordExit(outside bool, t int64, labels map[string]string, model string, record util.Recorder) {
	exitLabels := filterLabels(labels)
	exitLabels["__name__"] = "exit"
	exitLabels["ft_model"] = model
	exitLabels["ft_metric"] = labels["__name__"]

	val := math.NaN()
	if outside {
		val = 1
	}
	record.Record(util.Metric{Desc: exitLabels, Data: util.DataPoint{Val: val, Time: t}})
}
