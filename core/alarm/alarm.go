package alarm

import (
	"context"
	"sync"
	"time"
)

type eventType int

const (
	add eventType = iota
	cancel
)

const timeoutNoAlarm = time.Second * 100000
const toleranceExpiry = time.Millisecond * 10

type alarmEvent struct {
	alarmID string
	event   eventType
	alarm   *alarmItem
}

type alarmItem struct {
	initialDuration   time.Duration
	remainingDuration time.Duration
	callback          func(string)
}

type alarmScheduler struct {
	cancelFunc         context.CancelFunc
	scheduledAlarms    map[string]*alarmItem
	event              chan alarmEvent
	mutScheduledAlarms sync.RWMutex
}

// NewAlarmScheduler creates a new alarm scheduler instance and starts it's process loop
func NewAlarmScheduler() *alarmScheduler {
	as := &alarmScheduler{
		cancelFunc:      nil,
		scheduledAlarms: make(map[string]*alarmItem),
		event:           make(chan alarmEvent),
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	as.cancelFunc = cancelFunc

	go as.startProcessLoop(ctx)

	return as
}

// Add adds a new alarm to the alarm scheduler
func (as *alarmScheduler) Add(callback func(alarmID string), duration time.Duration, alarmID string) {
	alarm := &alarmItem{
		initialDuration:   duration,
		remainingDuration: duration,
		callback:          callback,
	}

	evt := alarmEvent{
		alarmID: alarmID,
		alarm:   alarm,
		event:   add,
	}

	as.event <- evt
}

// Cancel cancels a scheduled alarm
func (as *alarmScheduler) Cancel(alarmID string) {
	as.mutScheduledAlarms.RLock()
	_, ok := as.scheduledAlarms[alarmID]
	as.mutScheduledAlarms.RUnlock()

	if !ok {
		return
	}

	evt := alarmEvent{
		alarmID: alarmID,
		alarm:   nil,
		event:   cancel,
	}

	as.event <- evt
}

func (as *alarmScheduler) startProcessLoop(ctx context.Context) {
	waitTime := timeoutNoAlarm
	var startTime time.Time

	for {
		startTime = time.Now()

		select {
		case <-ctx.Done():
			return
		case evt := <-as.event:
			elapsedTime := time.Since(startTime)
			waitTime = as.handleEvent(evt, elapsedTime)

		case <-time.After(waitTime):
			waitTime = as.updateAlarms(waitTime)
		}
	}
}

func (as *alarmScheduler) handleEvent(evt alarmEvent, elapsedSinceLastUpdate time.Duration) time.Duration {
	var waitTime time.Duration
	switch evt.event {
	case add:
		waitTime = as.handleAdd(elapsedSinceLastUpdate, evt.alarm, evt.alarmID)
	case cancel:
		waitTime = as.handleCancel(elapsedSinceLastUpdate, evt.alarmID)
	default:
		waitTime = as.updateAlarms(elapsedSinceLastUpdate)
	}

	return waitTime
}

func (as *alarmScheduler) handleAdd(
	elapsedSinceLastUpdate time.Duration,
	alarm *alarmItem,
	alarmID string,
) time.Duration {
	waitTime := as.updateAlarms(elapsedSinceLastUpdate)

	as.mutScheduledAlarms.Lock()
	as.scheduledAlarms[alarmID] = alarm
	as.mutScheduledAlarms.Unlock()

	if waitTime > alarm.remainingDuration {
		waitTime = alarm.remainingDuration
	}

	return waitTime
}

func (as *alarmScheduler) handleCancel(elapsedSinceLastUpdate time.Duration, alarmID string) time.Duration {
	as.mutScheduledAlarms.Lock()
	delete(as.scheduledAlarms, alarmID)
	as.mutScheduledAlarms.Unlock()

	return as.updateAlarms(elapsedSinceLastUpdate)
}

// updateAlarms updates the remaining duration for all alarms and returns the remaining minimum duration
func (as *alarmScheduler) updateAlarms(elapsed time.Duration) time.Duration {
	minDuration := timeoutNoAlarm

	as.mutScheduledAlarms.Lock()
	defer as.mutScheduledAlarms.Unlock()

	for alarmID, alarm := range as.scheduledAlarms {
		if alarm.remainingDuration <= elapsed+toleranceExpiry {
			go alarm.callback(alarmID)
			delete(as.scheduledAlarms, alarmID)
		} else {
			alarm.remainingDuration -= elapsed
			if minDuration > alarm.remainingDuration {
				minDuration = alarm.remainingDuration
			}
		}
	}

	return minDuration
}

// Close closes the alarm scheduler stopping the process loop
func (as *alarmScheduler) Close() {
	as.cancelFunc()
}

// Reset resets the alarm with the given id
func (as *alarmScheduler) Reset(alarmID string) {
	as.mutScheduledAlarms.RLock()
	alarm, ok := as.scheduledAlarms[alarmID]
	if !ok {
		as.mutScheduledAlarms.RUnlock()
		return
	}

	callback := alarm.callback
	duration := alarm.initialDuration
	as.mutScheduledAlarms.RUnlock()

	evt := alarmEvent{
		alarmID: alarmID,
		alarm:   nil,
		event:   cancel,
	}

	as.event <- evt

	as.Add(callback, duration, alarmID)
}

// IsInterfaceNil returns true if interface is nil
func (as *alarmScheduler) IsInterfaceNil() bool {
	return as == nil
}
