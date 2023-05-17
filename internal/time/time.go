package time

import "time"

type Time = time.Time
type Duration = time.Duration

type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

var SystemClock Clock = systemClock{}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now()
}

func (systemClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}
