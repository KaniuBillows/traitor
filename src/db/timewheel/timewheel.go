package timewheel

import (
	"container/list"
	"time"
	"traitor/logger"
)

type location struct {
	slotIndex int           //the index of the slot where the job is located.
	element   *list.Element // the element in the list that the job is.
}

type TimeWheel struct {
	interval time.Duration
	ticker   *time.Ticker
	slots    []*list.List

	locationMap       map[string]*location // a map store the key's location.
	currentPos        int
	slotNum           int
	addTaskChannel    chan task
	removeTaskChannel chan string
	stopChannel       chan bool
}
type task struct {
	delay  time.Duration
	circle int
	key    string
	job    func()
}

func New(interval time.Duration, slotNum int) *TimeWheel {
	if interval <= 0 || slotNum <= 0 {
		panic("invalid argument")
	}
	tw := &TimeWheel{
		interval:          interval,
		slots:             make([]*list.List, slotNum),
		locationMap:       make(map[string]*location),
		currentPos:        0,
		slotNum:           slotNum,
		addTaskChannel:    make(chan task),
		removeTaskChannel: make(chan string),
		stopChannel:       make(chan bool),
	}
	tw.initSlots()
	return tw
}
func (tw *TimeWheel) initSlots() {
	for i := 0; i < tw.slotNum; i++ {
		tw.slots[i] = list.New()
	}
}
func (tw *TimeWheel) Start() {
	tw.ticker = time.NewTicker(tw.interval)
	go tw.start()
}
func (tw *TimeWheel) start() {
	for {
		select {
		case <-tw.ticker.C:
			{
				tw.tickHandler()
			}
		case task := <-tw.addTaskChannel:
			{
				tw.addTask(&task)
			}
		case key := <-tw.removeTaskChannel:
			{
				tw.removeTask(key)
			}
		case <-tw.stopChannel:
			{
				tw.ticker.Stop()
				return
			}
		}
	}
}

func (tw *TimeWheel) tickHandler() {
	// find current slots
	l := tw.slots[tw.currentPos]
	// update currentPos to next and wait for next tick tok.
	if tw.currentPos == tw.slotNum-1 {
		tw.currentPos = 0 //circular queue
	} else {
		tw.currentPos++
	}
	go tw.scanAndRunTask(l) // new coroutine handle tasks in the slotIndex.
}

func (tw *TimeWheel) scanAndRunTask(l *list.List) {
	for e := l.Front(); e != nil; {
		task := e.Value.(*task)
		//task with long delays.
		if task.circle > 0 {
			task.circle--
			e = e.Next()
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
		next := e.Next()
		l.Remove(e)
		if task.key != "" {
			delete(tw.locationMap, task.key)
		}
		e = next
	}
}

func (tw *TimeWheel) addTask(task *task) {
	pos, circle := tw.getPositionAndCircle(task.delay)
	task.circle = circle

	e := tw.slots[pos].PushBack(task)
	loc := &location{
		slotIndex: pos,
		element:   e,
	}
	if task.key != "" {
		_, ok := tw.locationMap[task.key] //if the same key has already exists.
		if ok {
			tw.removeTask(task.key) // cover the same key.
		}
	}
	tw.locationMap[task.key] = loc
}

/*
calculate the circle count.
*/
func (tw *TimeWheel) getPositionAndCircle(d time.Duration) (pos int, circle int) {
	delaySeconds := int(d.Seconds())
	intervalSeconds := int(tw.interval.Seconds())
	circle = int(delaySeconds / intervalSeconds / tw.slotNum)
	pos = int(tw.currentPos+delaySeconds/intervalSeconds) % tw.slotNum

	return
}

func (tw *TimeWheel) removeTask(key string) {
	if loc, ok := tw.locationMap[key]; ok { // key  exists.
		ls := tw.slots[loc.slotIndex]
		ls.Remove(loc.element)
		delete(tw.locationMap, key)
	}

}

// AddJob into the timeWheel.
func (tw *TimeWheel) AddJob(delay time.Duration, key string, job func()) {
	if delay < 0 {
		return
	}
	tw.addTaskChannel <- task{delay: delay, key: key, job: job}
}
func (tw *TimeWheel) RemoveJob(key string) {
	if key == "" {
		return
	}
	tw.removeTaskChannel <- key
}
