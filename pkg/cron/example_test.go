package cron

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestCronExample(t *testing.T) {
	// https://www.freeformatter.com/cron-expression-generator-quartz.html
	expression := "59 3 15 * * MON"

	parser := NewParser(Second | Minute | Hour | Dom | Month | Dow | Descriptor)

	s, err := parser.Parse(expression)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	now := time.Now()
	fmt.Println("Now:", now)
	fmt.Println("Next", s.Next(now))

}
