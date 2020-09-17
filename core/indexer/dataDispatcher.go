package indexer

import (
	"context"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
)

var log = logger.GetOrCreate("core/indexer")

// Options structure holds the indexer's configuration options
type Options struct {
	IndexerCacheSize  int
	UseKibana         bool
	TxIndexingEnabled bool
}

const (
	backOffTime = time.Second * 10
	maxBackOff  = time.Minute * 5
)

type dataDispatcher struct {
	backOffTime   time.Duration
	chanWorkItems chan workItems.WorkItemHandler
	cancelFunc    func()
}

// NewDataDispatcher creates a new dataDispatcher instance, capable of saving sequentially data in elasticsearch database
func NewDataDispatcher(cacheSize int) (*dataDispatcher, error) {
	if cacheSize < 0 {
		return nil, ErrNegativeCacheSize
	}

	dd := &dataDispatcher{
		chanWorkItems: make(chan workItems.WorkItemHandler, cacheSize),
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
			d.doWork(wi)
		}
	}
}

// Close will close the endless running go routine
func (d *dataDispatcher) Close() error {
	if d.cancelFunc != nil {
		d.cancelFunc()
	}

	return nil
}

// Add will add a new item in queue
func (d *dataDispatcher) Add(item workItems.WorkItemHandler) {
	if check.IfNil(item) {
		log.Warn("dataDispatcher.Add nil item: will do nothing")
		return
	}

	d.chanWorkItems <- item
}

func (d *dataDispatcher) doWork(wi workItems.WorkItemHandler) {
	for {
		err := wi.Save()
		if err == ErrBackOff {
			log.Warn("dataDispatcher.doWork could not index item",
				"received back off:", err.Error())

			d.increaseBackOffTime()
			time.Sleep(d.backOffTime)

			continue
		}
		if err != nil {
			log.Warn("dataDispatcher.doWork", "removing item from queue", err.Error())
		}

		d.backOffTime = 0
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
