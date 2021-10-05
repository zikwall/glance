package metric

import (
	"math"
	"strconv"
)

func StringToInt(s string) int {
	n, err := strconv.Atoi(s)

	if err != nil {
		return 0
	}

	return n
}

func StringToFloat64(s string) float64 {
	value, err := strconv.ParseFloat(s, 64)

	if err != nil {
		return 0
	}

	if math.IsInf(value, -1) || math.IsInf(value, 1) || math.IsInf(value, 0) {
		return 0
	}

	return value
}
