package execution

import (
	"sync"
	"time"
)

type RoutineStat int

const (
	// go routine is not started
	Closed RoutineStat = iota
	// go routine is running
	Started
	// go routine will close
	Closing
)

//RoutineWrapper can manage the execution of a go-routine.
//The event OnDoSimpleTask will be called periodically (if needed with a desired delay that can be set
// in the DurCalls field. Implicitly, this field has 0 value)
//Parameter field is used if one needs to pass some data to the executed event handler.
type RoutineWrapper struct {
	Parameter interface{}
	stopper   chan bool

	mutStat sync.RWMutex
	stat    RoutineStat

	OnDoSimpleTask func(parameter interface{})

	DurCalls time.Duration
}

func NewRoutineWrapper() *RoutineWrapper {
	rw := RoutineWrapper{stopper: make(chan bool, 1), stat: Closed, DurCalls: 0}

	return &rw
}

func (rw *RoutineWrapper) Start() {
	rw.mutStat.Lock()
	if rw.stat != Closed {
		rw.mutStat.Unlock()
		return
	}

	rw.stat = Started
	rw.mutStat.Unlock()

	go rw.runner()
}

func (rw *RoutineWrapper) Stop() {
	rw.mutStat.Lock()
	defer rw.mutStat.Unlock()

	if rw.stat == Started {
		rw.stat = Closing

		rw.stopper <- true
	}

}

func (rw *RoutineWrapper) runner() {
	defer func() {
		rw.mutStat.Lock()
		rw.stat = Closed
		rw.mutStat.Unlock()
	}()

	for {
		//do task
		select {
		default:
			if rw.OnDoSimpleTask != nil {
				rw.OnDoSimpleTask(rw.Parameter)
			}
		case <-rw.stopper:
			return
		}

		//wait if necessary, otherwise continue
		if rw.DurCalls == 0 {
			continue
		}

		time.Sleep(rw.DurCalls)

		//re-test whether there is a stop condition that might have arrived in sleep interval
		//otherwise, continue
		select {
		default:
			continue
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
