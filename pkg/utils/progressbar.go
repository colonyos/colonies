package utils

import (
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/muesli/termenv"
)

var (
	env = termenv.EnvColorProfile()
)

func ProgressBar(trackers int) progress.Writer {
	autoStop := true
	hideETA := false
	hideETAOverall := false
	hideOverallTracker := false
	hidePercentage := false
	hideTime := true
	hideValue := false
	showSpeed := false
	showSpeedOverall := false
	showPinned := true

	pw := progress.NewWriter()
	pw.SetAutoStop(autoStop)
	pw.SetTrackerLength(20)
	pw.SetMessageWidth(40)
	pw.SetNumTrackersExpected(trackers)
	pw.SetSortBy(progress.SortByPercentDsc)
	pw.SetStyle(progress.StyleDefault)
	pw.SetTrackerPosition(progress.PositionRight)
	pw.SetUpdateFrequency(time.Millisecond * 10)
	pw.Style().Colors = progress.StyleColorsExample
	pw.Style().Options.PercentFormat = "%4.1f%%"
	pw.Style().Visibility.ETA = !hideETA
	pw.Style().Visibility.ETAOverall = !hideETAOverall
	pw.Style().Visibility.Percentage = !hidePercentage
	pw.Style().Visibility.Speed = showSpeed
	pw.Style().Visibility.SpeedOverall = showSpeedOverall
	pw.Style().Visibility.Time = !hideTime
	pw.Style().Visibility.TrackerOverall = !hideOverallTracker
	pw.Style().Visibility.Value = !hideValue
	pw.Style().Visibility.Pinned = showPinned

	return pw
}
