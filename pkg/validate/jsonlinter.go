package validate

import (
	"encoding/json"
	"fmt"
	"strings"
)

const prefixStr = "          <-- "

func findErrorLineNumber(jsonStr string, err error) int {
	if syntaxError, ok := err.(*json.SyntaxError); ok {
		newlines := strings.Count(jsonStr[:syntaxError.Offset], "\n")
		return newlines
	}
	return -1
}

func AddStringToLine(originalStr string, lineNumber int, message string) (string, error) {
	lines := strings.Split(originalStr, "\n")

	if lineNumber < 1 || lineNumber > len(lines) {
		return "", fmt.Errorf("line number %d is out of range", lineNumber)
	}

	lines[lineNumber-1] += prefixStr + message

	return strings.Join(lines, "\n"), nil
}

func JSON(err error, jsonStr string, full bool) (string, error) {
	if err == nil {
		return "", nil
	}
	if err != nil {
		lineNumber := findErrorLineNumber(jsonStr, err)
		if lineNumber > 0 {
			if full {
				modifiedStr, err2 := AddStringToLine(jsonStr, lineNumber, err.Error())
				if err2 != nil {
					return "", fmt.Errorf("%v; %v", err, err2)
				}
				return modifiedStr, nil
			} else {
				lines := strings.Split(jsonStr, "\n")
				return lines[lineNumber] + prefixStr + err.Error(), nil
			}
		}
	}

	return "", err
}
