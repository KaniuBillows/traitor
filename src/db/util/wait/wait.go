package wait

import (
	"sync"
	"time"
)

type Wait struct {
	waitGroup sync.WaitGroup
}

func (w *Wait) Add(d int) {
	w.waitGroup.Add(d)
}

func (w *Wait) Wait() {
	w.waitGroup.Wait()
}
func (w *Wait) Done() {
	w.waitGroup.Done()
}

// WaitTimeOut
// wait if counter is zero or timeout.
// if timout return true.
func (w *Wait) WaitTimeOut(timeout time.Duration) bool {
	c := make(chan struct{}, 1)
	go func() {
		defer close(c)
		w.Wait()
		c <- struct{}{}
	}()

	select {
	case _ = <-c:
		{
			return false
		}
	case _ = <-time.After(timeout):
		{
			return true
		}
	}

}
