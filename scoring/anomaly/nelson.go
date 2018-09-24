package anomaly

import (
	"github.com/Fresh-Tracks/data-sidecar/stat"
)

func anomalyLabels(labels map[string]string, model string) map[string]string {
	anomalyLabels := make(map[string]string)
	for xx, yy := range labels {
		if xx == "ft_target" {
			continue
		}
		anomalyLabels[xx] = yy
	}
	anomalyLabels["__name__"] = "anomaly"
	anomalyLabels["ft_model"] = model
	anomalyLabels["ft_metric"] = labels["__name__"]
	return anomalyLabels
}

func anomalyHelper(aName string, fire bool, name map[string]string, record *[]map[string]string) {
	// this once did more, and could again...
	if fire {
		*record = append(*record, anomalyLabels(name, aName))
	}
}

// Nelson computes the nelson rules on a slice of data.
func Nelson(data []float64, name map[string]string) []map[string]string {
	// need enough information to do nelson rules on.
	// calculate quantiles instead of using the mean+std approach.
	var (
		mean, std = stat.MeanStdDev(data)
		mm3sd     = mean - 3*std
		mm2sd     = mean - 2*std
		mm1sd     = mean - 1*std
		mp1sd     = mean + 1*std
		mp2sd     = mean + 2*std
		mp3sd     = mean + 3*std
	)
	record := make([]map[string]string, 0)
	// of the nelson rules, we found that only these three really hold up in general as useful
	// indicators of anything.
	anomalyHelper("nelson_large_ooc", NelsonLargeOoC(data, mm3sd, mp3sd), name, &record)
	anomalyHelper("nelson_medium_ooc", NelsonMediumOoC(data, mm2sd, mp2sd), name, &record)
	anomalyHelper("nelson_small_ooc", NelsonSmallOoC(data, mm1sd, mp1sd), name, &record)
	return record
}

// NelsonLargeOoC reports if the most recent point is outside of the
// equivalent of 3sds of the mean
func NelsonLargeOoC(data []float64, low, high float64) bool {
	if low == high {
		return false
	}
	if data[len(data)-1] > high {
		return true
	}
	if data[len(data)-1] < low {
		return true
	}

	return false
}

// NelsonMediumOoC reports if 2 of the most recent 3 points are outside of the
// equivalent of 2sds of the mean
func NelsonMediumOoC(data []float64, low, high float64) bool {
	if len(data) < 3 {
		return false
	}
	if low == high {
		return false
	}
	lows := 0
	highs := 0
	for ii := 0; ii < 3; ii++ {
		if data[len(data)-1-ii] < low {
			lows++
		}
		if data[len(data)-1-ii] > high {
			highs++
		}
	}
	if (lows > 1) || (highs > 1) {
		return true
	}
	return false
}

// NelsonSmallOoC reports 4 of the 5 most recent point is outside of the
// equivalent of 1sds of the mean
func NelsonSmallOoC(data []float64, low, high float64) bool {
	if len(data) < 5 {
		return false
	}
	if low == high {
		return false
	}
	lows := 0
	highs := 0
	for ii := 0; ii < 5; ii++ {
		if data[len(data)-1-ii] < low {
			lows++
		}
		if data[len(data)-1-ii] > high {
			highs++
		}
	}
	if (lows > 3) || (highs > 3) {
		return true
	}
	return false
}
