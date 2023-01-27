package health

import (
	"container/list"
	"runtime"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
)

var _ record = (*dummyRecord)(nil)
var _ diagnosable = (*dummyDiagnosable)(nil)
var _ memory = (*dummyMemory)(nil)
var _ clock = (*dummyClock)(nil)

var dummySignal struct{}

type dummyRecord struct {
	importance int
	saved      bool
	deleted    bool
}

func newDummyRecord(importance int) *dummyRecord {
	return &dummyRecord{importance: importance}
}

func (dummy *dummyRecord) save() error {
	dummy.saved = true
	return nil
}

func (dummy *dummyRecord) delete() error {
	dummy.deleted = true
	return nil
}

func (dummy *dummyRecord) isMoreImportantThan(otherRecord record) bool {
	return dummy.importance > otherRecord.(*dummyRecord).importance
}

type dummyDiagnosable struct {
	numShallowDiagnoses atomic.Counter
	numDeepDiagnoses    atomic.Counter
}

// Diagnose -
func (dummy *dummyDiagnosable) Diagnose(deep bool) {
	if deep {
		dummy.numDeepDiagnoses.Increment()
	} else {
		dummy.numShallowDiagnoses.Increment()
	}
}

// IsInterfaceNil -
func (dummy *dummyDiagnosable) IsInterfaceNil() bool {
	return dummy == nil
}

type dummyNotDiagnosable struct {
}

type dummyMemory struct {
	inUse             int
	numGetStatsCalled atomic.Counter
}

func newDummyMemory(inUse int) *dummyMemory {
	return &dummyMemory{
		inUse: inUse,
	}
}

func (dummy *dummyMemory) getStats() runtime.MemStats {
	dummy.numGetStatsCalled.Increment()
	return runtime.MemStats{HeapInuse: uint64(dummy.inUse)}
}

// dummyEvent objects are managed by the dummyClock
type dummyEvent struct {
	channel chan time.Time
	time    time.Time
}

func (dummy *dummyEvent) makeItHappen() {
	dummy.channel <- dummy.time
}

// dummyClock simulates a real clock
type dummyClock struct {
	mutex          sync.RWMutex
	ticks          int
	eventsSchedule *list.List
}

func newDummyClock() *dummyClock {
	return &dummyClock{
		eventsSchedule: list.New(),
	}
}

func (dummy *dummyClock) now() time.Time {
	dummy.mutex.RLock()
	defer dummy.mutex.RUnlock()
	return dummy.nowNoLock()
}

func (dummy *dummyClock) nowNoLock() time.Time {
	return time.Time{}.Add(time.Duration(dummy.ticks) * time.Second)
}

func (dummy *dummyClock) after(d time.Duration) <-chan time.Time {
	dummy.mutex.Lock()
	defer dummy.mutex.Unlock()

	eventTime := dummy.nowNoLock().Add(d)
	eventChannel := make(chan time.Time)

	dummy.eventsSchedule.PushBack(&dummyEvent{
		time:    eventTime,
		channel: eventChannel,
	})

	return eventChannel
}

// tick makes time flow (in one-second steps)
func (dummy *dummyClock) tick() {
	dummy.mutex.Lock()
	dummy.ticks++

	eventToHappen, eventToHappenElement := dummy.getEventToHappenNoLock()
	if eventToHappen != nil {
		dummy.eventsSchedule.Remove(eventToHappenElement)
	}

	dummy.mutex.Unlock()

	if eventToHappen != nil {
		eventToHappen.makeItHappen()
	}
}

func (dummy *dummyClock) getEventToHappenNoLock() (eventToHappen *dummyEvent, eventToHappenElement *list.Element) {
	now := dummy.nowNoLock()

	for element := dummy.eventsSchedule.Front(); element != nil; element = element.Next() {
		event := element.Value.(*dummyEvent)
		if now.After(event.time) || now == event.time {
			eventToHappen = event
			eventToHappenElement = element
			break
		}
	}

	return
}
