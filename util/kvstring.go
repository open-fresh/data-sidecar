package util

import (
	"sort"
	"strings"
)

// MapSSToS takes a map, orders the keys, and converts the whole thing to a string.
// maps have no order, but this will.
func MapSSToS(mapIn map[string]string) string {
	temp := make([]string, len(mapIn), len(mapIn))
	ii := 0
	for xx := range mapIn {
		temp[ii] = xx
		ii++
	}
	sort.Strings(temp)
	output := make([]string, len(temp), len(temp))
	for ii, xx := range temp {
		if (xx == "") && (mapIn[xx] == "") {
			continue
		}
		output[ii] = "\"" + xx + "\":\"" + mapIn[xx] + "\""
	}
	return "{" + strings.Join(output, ", ") + "}"
}
