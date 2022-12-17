package schedule

import (
	"container/list"
	"time"
	"traitor/logger"
)

const (
	interval = 1 // default 1s.
	slotNums = 3600
)

type location struct {
	slotIndex int
	elem      *list.Element
}

type timeWheel struct {
	ticker         *time.Ticker
	interval       time.Duration
	slotNum        int
	currentPos     int
	slots          []*list.List
	locationMap    map[string]*location
	addTaskChan    chan task
	removeTaskChan chan string
	stopChannel    chan bool
	running        bool
}
type task struct {
	delay         time.Duration
	circle        int
	initialCircle int
	job           func()
	key           string
}

func makeTimeWheel() *timeWheel {
	var timeWheel = timeWheel{
		interval:       time.Second * interval,
		slotNum:        slotNums,
		currentPos:     0,
		slots:          make([]*list.List, slotNums),
		locationMap:    make(map[string]*location),
		addTaskChan:    make(chan task),
		removeTaskChan: make(chan string),
		stopChannel:    make(chan bool),
		running:        false,
	}
	for i := 0; i < slotNums; i++ {
		timeWheel.slots[i] = list.New()
	}
	return &timeWheel
}

func (t *timeWheel) start() {
	if t.running != false {
		return
	}
	t.ticker = time.NewTicker(t.interval)
	go t.handleEvent()
	t.running = true
}
func (t *timeWheel) stop() {
	t.stopChannel <- false
}
func (t *timeWheel) removeTask(key string) {
	t.removeTaskChan <- key
}

func (t *timeWheel) handleRemove(key string) {
	if loc, ok := t.locationMap[key]; ok {
		ls := t.slots[loc.slotIndex]
		ls.Remove(loc.elem)
		delete(t.locationMap, key)
	}
}

// AddJob into the timeWheel.
func (t *timeWheel) AddJob(delay time.Duration, key string, job func()) {
	if delay < 0 {
		return
	}
	t.addTaskChan <- task{delay: delay, key: key, job: job}
}
func (t *timeWheel) getPositionAndCircle(d time.Duration) (pos int, circle int) {
	delaySeconds := int(d.Seconds())
	intervalSeconds := int(t.interval.Seconds())
	circle = delaySeconds / intervalSeconds / t.slotNum
	pos = (t.currentPos + delaySeconds/intervalSeconds) % t.slotNum

	return
}
func (t *timeWheel) handleAddTask(task task) {
	pos, circle := t.getPositionAndCircle(task.delay)
	task.circle = circle

	e := t.slots[pos].PushBack(task)
	loc := &location{
		slotIndex: pos,
		elem:      e,
	}
	if task.key != "" {
		_, ok := t.locationMap[task.key] //if the same key has already exists.
		if ok {
			t.removeTask(task.key) // cover the same key.
		}
	}
	t.locationMap[task.key] = loc
}

func (t *timeWheel) handleEvent() {
	for {
		select {
		case <-t.ticker.C:
			{
				t.tickHandler()
			}
		case task := <-t.addTaskChan:
			{
				t.handleAddTask(task)
			}
		case key := <-t.removeTaskChan:
			{
				t.handleRemove(key)
			}
		case stop := <-t.stopChannel:
			{
				if stop == true {
					t.ticker.Stop()
					return
				}
			}
		}
	}
}
func (t *timeWheel) tickHandler() {
	// find current slots
	l := t.slots[t.currentPos]
	// update currentPos to next and wait for next tick tok.
	if t.currentPos == t.slotNum-1 {
		t.currentPos = 0 //circular queue
	} else {
		t.currentPos++
	}
	go t.scanAndRunTask(l) // new coroutine handle tasks in the slotIndex.
}

func (t *timeWheel) scanAndRunTask(l *list.List) {
	for elem := l.Front(); elem != nil; {
		task := elem.Value.(*task)
		if task.circle > 0 {
			task.circle--
			elem = elem.Next()
			continue
		}
		// execute job async.
		go func() {
			defer func() {
				if err := recover(); err != nil {
					logger.Error(err)
				}
			}()
			job := task.job
			job()
		}()
		next := elem.Next()
		l.Remove(elem)
		delete(t.locationMap, task.key)

		elem = next
	}
}
