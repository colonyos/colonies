package cron

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCronNext(t *testing.T) {
	//cronExpr := "1 * * * * *"
	cronExpr := "* * 8 * * *"
	nextTime, err := Next(cronExpr)
	assert.Nil(t, err)

	diff := nextTime.Sub(time.Now())
	assert.True(t, diff.Milliseconds() >= 0)
}

func TestCronRandom(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	cronExpr := "* * 8 * * *" // random time between now and 08:00
	nextTime, err := Random(cronExpr)
	assert.Nil(t, err)
	diff := nextTime.Sub(time.Now())
	assert.True(t, diff.Milliseconds() >= 0)

	nextTime2, err := Random(cronExpr)
	assert.NotEqual(t, nextTime, nextTime2)
}
