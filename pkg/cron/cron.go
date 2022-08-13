package cron

import (
	"math/rand"
	"time"
)

func Next(cronExpr string) (time.Time, error) {
	parser := NewParser(Second | Minute | Hour | Dom | Month | Dow | Descriptor)

	p, err := parser.Parse(cronExpr)
	if err != nil {
		return time.Time{}, err
	}

	now := time.Now()
	return p.Next(now), nil
}

func Random(cronExpr string) (time.Time, error) {
	parser := NewParser(Second | Minute | Hour | Dom | Month | Dow | Descriptor)

	p, err := parser.Parse(cronExpr)
	if err != nil {
		return time.Time{}, err
	}

	now := time.Now()
	nextTime := p.Next(now)
	diff := nextTime.Sub(time.Now()).Milliseconds()

	min := 00
	max := int(diff)
	randomDiff := rand.Intn(max-min) + min
	now = now.Add(time.Duration(randomDiff * int(time.Millisecond)))

	return now, nil
}
