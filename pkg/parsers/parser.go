package parsers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func ConvertCPUToInt(cpu string) (int64, error) {
	if strings.HasSuffix(cpu, "m") {
		cpu = strings.TrimSuffix(cpu, "m")
	} else {
		return 0, nil
	}

	value, err := strconv.ParseInt(cpu, 10, 64)
	if err != nil {
		return -1, errors.New("Error converting CPU to int")
	}

	return value, nil
}

func ConvertMemoryToInt(mem string) (int64, error) {
	unitMap := map[string]int64{
		"Ki": 1024, "KiB": 1024,
		"Mi": 1024 * 1024, "MiB": 1024 * 1024,
		"Gi": 1024 * 1024 * 1024, "GiB": 1024 * 1024 * 1024,
		"K": 1000, "KB": 1000,
		"M": 1000 * 1000, "MB": 1000 * 1000,
		"G": 1000 * 1000 * 1000, "GB": 1000 * 1000 * 1000,
	}

	// Handling edge cases
	if len(mem) == 0 {
		return 0, nil
	}

	var numStr string
	var unit string
	for u := range unitMap {
		if strings.HasSuffix(strings.ToUpper(mem), strings.ToUpper(u)) {
			numStr = strings.TrimSuffix(mem, u)
			unit = u
			break
		}
	}

	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error converting memory to int: %v", err)
	}

	if unit == "" {
		return 0, fmt.Errorf("no valid unit found in input")
	}

	mb := num * unitMap[unit] / 1000

	return mb, nil
}

func ConvertCPUToString(cpu int64) string {
	return strconv.FormatInt(cpu, 10) + "m"
}

func ConvertMemoryToString(mem int64) string {
	return strconv.FormatInt(mem, 10) + "Mi"
}
