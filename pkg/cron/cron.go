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

func NextIntervall(intervall int) (time.Time, error) {
	now := time.Now()
	next := now.Add(time.Duration(intervall * int(time.Second)))

	return next, nil
}

func Random(intervall int) (time.Time, error) {
	now := time.Now()
	min := 00
	max := intervall
	randomDiff := rand.Intn(max-min) + min
	next := now.Add(time.Duration(randomDiff * int(time.Second)))

	return next, nil
}
