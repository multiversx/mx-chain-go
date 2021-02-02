package indexer

import (
	"context"
	"errors"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer/client"
	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
)

var log = logger.GetOrCreate("core/indexer")

const durationBetweenErrorRetry = time.Second * 3

const (
	closeTimeout = time.Second * 20
	backOffTime  = time.Second * 10
	maxBackOff   = time.Minute * 5
)

type dataDispatcher struct {
	backOffTime         time.Duration
	chanWorkItems       chan workItems.WorkItemHandler
	cancelFunc          func()
	wasClosed           *atomic.Flag
	currentWriteDone    chan struct{}
	closeStartTime      time.Time
	mutexCloseStartTime sync.RWMutex
}

// NewDataDispatcher creates a new dataDispatcher instance, capable of saving sequentially data in elasticsearch database
func NewDataDispatcher(cacheSize int) (*dataDispatcher, error) {
	if cacheSize < 0 {
		return nil, ErrNegativeCacheSize
	}

	dd := &dataDispatcher{
		chanWorkItems:       make(chan workItems.WorkItemHandler, cacheSize),
		wasClosed:           &atomic.Flag{},
		currentWriteDone:    make(chan struct{}),
		mutexCloseStartTime: sync.RWMutex{},
	}

	return dd, nil
}

// StartIndexData will start index data in database
func (d *dataDispatcher) StartIndexData() {
	var ctx context.Context
	ctx, d.cancelFunc = context.WithCancel(context.Background())

	go d.startWorker(ctx)
}

func (d *dataDispatcher) startWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			d.stopWorker()
			return
		case wi := <-d.chanWorkItems:
			timeoutOnClose := d.doWork(wi)
			if timeoutOnClose {
				d.stopWorker()
				return
			}
		}
	}
}

func (d *dataDispatcher) stopWorker() {
	log.Debug("dispatcher's go routine is stopping...")
	d.currentWriteDone <- struct{}{}
}

// Close will close the endless running go routine
func (d *dataDispatcher) Close() error {
	if d.wasClosed.Set() {
		return nil
	}

	d.mutexCloseStartTime.Lock()
	d.closeStartTime = time.Now()
	d.mutexCloseStartTime.Unlock()

	if d.cancelFunc != nil {
		d.cancelFunc()
	}

	<-d.currentWriteDone
	d.consumeRemainingItems()
	return nil
}

func (d *dataDispatcher) consumeRemainingItems() {
	for {
		select {
		case wi := <-d.chanWorkItems:
			isTimeout := d.doWork(wi)
			if isTimeout {
				return
			}
		default:
			return
		}
	}
}

// Add will add a new item in queue
func (d *dataDispatcher) Add(item workItems.WorkItemHandler) {
	if check.IfNil(item) {
		log.Warn("dataDispatcher.Add nil item: will do nothing")
		return
	}
	if d.wasClosed.IsSet() {
		log.Warn("dataDispatcher.Add cannot add item: channel chanWorkItems is closed")
		return
	}

	d.chanWorkItems <- item
}

func (d *dataDispatcher) doWork(wi workItems.WorkItemHandler) bool {
	for {
		if d.exitIfTimeoutOnClose() {
			log.Warn("dataDispatcher.doWork could not index item",
				"error", "timeout")
			return true
		}

		err := wi.Save()
		if errors.Is(err, client.ErrBackOff) {
			log.Warn("dataDispatcher.doWork could not index item",
				"received back off:", err.Error())

			d.increaseBackOffTime()
			time.Sleep(d.backOffTime)

			continue
		}

		d.backOffTime = 0
		if err != nil {
			log.Warn("dataDispatcher.doWork could not index item (will retry)", "error", err.Error())
			time.Sleep(durationBetweenErrorRetry)

			continue
		}

		return false
	}

}

func (d *dataDispatcher) exitIfTimeoutOnClose() bool {
	if !d.wasClosed.IsSet() {
		return false
	}

	d.mutexCloseStartTime.RLock()
	passedTime := time.Since(d.closeStartTime)
	d.mutexCloseStartTime.RUnlock()

	return passedTime > closeTimeout
}

func (d *dataDispatcher) increaseBackOffTime() {
	if d.backOffTime == 0 {
		d.backOffTime = backOffTime
		return
	}
	if d.backOffTime >= maxBackOff {
		return
	}

	d.backOffTime += d.backOffTime / 5
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *dataDispatcher) IsInterfaceNil() bool {
	return d == nil
}
