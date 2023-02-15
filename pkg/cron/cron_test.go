package cron

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCronNext(t *testing.T) {
	cronExpr := "* * 8 * * *"
	nextTime, err := Next(cronExpr)
	assert.Nil(t, err)

	diff := nextTime.Sub(time.Now())
	assert.True(t, diff.Milliseconds() >= 0)
}

func TestCronNextInternvall(t *testing.T) {
	nextTime, err := NextInterval(10)
	assert.Nil(t, err)

	diff := nextTime.Sub(time.Now())
	assert.True(t, diff.Milliseconds() >= 0)
}

func TestCronRandom(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	nextTime, err := Random(60 * 60 * 24 * 7) // random time the coming week
	assert.Nil(t, err)
	diff := nextTime.Sub(time.Now())
	assert.True(t, diff.Milliseconds() >= 0)

	nextTime2, err := Random(60 * 60 * 24 * 7) // random time the coming week
	assert.NotEqual(t, nextTime, nextTime2)
}
