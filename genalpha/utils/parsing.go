package utils

import "strconv"

func ParseNumber(s string) float64 {
	result, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic("Error parsing number " + s)
	}

	return result
}

func ParseBool(s string) bool {
	result, err := strconv.ParseBool(s)
	if err != nil {
		panic("Error parsing bool " + s)
	}

	return result
}
