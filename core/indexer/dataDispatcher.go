package indexer

import (
	"context"
	"errors"
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
	closeTimeout = time.Second * 5
	backOffTime  = time.Second * 10
	maxBackOff   = time.Minute * 5
)

type dataDispatcher struct {
	backOffTime   time.Duration
	chanWorkItems chan workItems.WorkItemHandler
	cancelFunc    func()
	wasClosed     atomic.Flag
}

// NewDataDispatcher creates a new dataDispatcher instance, capable of saving sequentially data in elasticsearch database
func NewDataDispatcher(cacheSize int) (*dataDispatcher, error) {
	if cacheSize < 0 {
		return nil, ErrNegativeCacheSize
	}

	dd := &dataDispatcher{
		chanWorkItems: make(chan workItems.WorkItemHandler, cacheSize),
		wasClosed:     atomic.Flag{},
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
			log.Debug("dispatcher's go routine is stopping...")
			return
		case wi := <-d.chanWorkItems:
			d.doWork(wi, 0)
		}
	}
}

// Close will close the endless running go routine
func (d *dataDispatcher) Close() error {
	start := time.Now()
	if d.cancelFunc != nil {
		d.cancelFunc()
	}

	d.wasClosed.Set()
	close(d.chanWorkItems)
	for wi := range d.chanWorkItems {
		if time.Since(start) > closeTimeout {
			log.Warn("cannot write all items from the queue",
				"error", "timeout",
			)
			return nil
		}

		remainingTime := closeTimeout - time.Since(start)
		d.doWork(wi, remainingTime)
	}

	return nil
}

// Add will add a new item in queue
func (d *dataDispatcher) Add(item workItems.WorkItemHandler) {
	if check.IfNil(item) {
		log.Warn("dataDispatcher.Add nil item: will do nothing")
		return
	}
	if d.wasClosed.IsSet() {
		return
	}

	d.chanWorkItems <- item
}

func (d *dataDispatcher) doWork(wi workItems.WorkItemHandler, timeout time.Duration) {
	start := time.Now()
	for {
		isTimeout := time.Since(start) > timeout
		// if timeout is 0 should be ignored
		if isTimeout && timeout != 0 {
			log.Warn("dataDispatcher.doWork could not index item",
				"error", "timeout")
			return
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

		return
	}

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
