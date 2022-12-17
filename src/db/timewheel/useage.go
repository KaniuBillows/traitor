package timewheel

import (
	"time"
)

var timeWheel = New(time.Second, 3600)

func init() {
	timeWheel.Start()
}

func Delay(duration time.Duration, key string, job func()) {
	timeWheel.AddJob(duration, key, job)
}

func ExeAt(t time.Time, key string, job func()) {
	timeWheel.AddJob(t.Sub(time.Now()), key, job)
}

func Cancel(key string) {
	timeWheel.RemoveJob(key)
}
