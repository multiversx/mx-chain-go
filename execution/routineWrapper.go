package execution

import "sync"

type RoutineStat int

const (
	// go routine is not started
	CLOSED RoutineStat = iota
	// go routine is running
	STARTED
	// go routine will close
	CLOSING
)

type RoutineWrapper struct {
	Parameter interface{}
	stopper   chan bool

	mutStat sync.RWMutex
	stat    RoutineStat

	OnDoSimpleTask func(parameter interface{})
}

func NewRoutineWrapper() *RoutineWrapper {
	rw := RoutineWrapper{stopper: make(chan bool, 1), stat: CLOSED}

	return &rw
}

func (rw *RoutineWrapper) Start() {
	rw.mutStat.Lock()
	if rw.stat != CLOSED {
		rw.mutStat.Unlock()
		return
	}

	rw.stat = STARTED
	rw.mutStat.Unlock()

	go rw.runner()
}

func (rw *RoutineWrapper) Stop() {
	rw.mutStat.Lock()
	defer rw.mutStat.Unlock()

	if rw.stat == STARTED {
		rw.stat = CLOSING

		rw.stopper <- true
	}

}

func (rw *RoutineWrapper) runner() {
	defer func() {
		rw.mutStat.Lock()
		rw.stat = CLOSED
		rw.mutStat.Unlock()
	}()

	for {

		select {
		default:
			if rw.OnDoSimpleTask != nil {
				rw.OnDoSimpleTask(rw.Parameter)
			}
		case <-rw.stopper:
			return
		}
	}
}

func (rw *RoutineWrapper) Stat() RoutineStat {
	rw.mutStat.RLock()
	defer rw.mutStat.RUnlock()

	return rw.stat
}
