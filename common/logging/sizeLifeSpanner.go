package logging

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

const minMBLifeSpan = 1

type sizeLifeSpanner struct {
	*baseLifeSpanner
	spanInMB    uint32
	cancelFunc  context.CancelFunc
	currentFile string
}

func newSizeLifeSpanner(sizeLifeSpan uint32) (*sizeLifeSpanner, error) {
	log.Info("newSizeLifeSpanner entered", "timespan", sizeLifeSpan)
	if sizeLifeSpan < minMBLifeSpan {
		return nil, fmt.Errorf("newSizeLifeSpanner %w, provided %v", core.ErrInvalidLogFileMinLifeSpan, sizeLifeSpan)
	}

	sls := &sizeLifeSpanner{
		spanInMB:        sizeLifeSpan * 1024 * 1024,
		baseLifeSpanner: newBaseLifeSpanner(),
	}

	log.Info("the newSizeLifeSpanner", "timespan", sizeLifeSpan)

	return sls, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sls *sizeLifeSpanner) IsInterfaceNil() bool {
	return sls == nil
}

// SetCurrentFile sets the file need for monitoring for the size
func (sls *sizeLifeSpanner) SetCurrentFile(path string) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	sls.cancelFunc = cancelFunc

	go sls.startTicker(ctx, path, int64(sls.spanInMB))
}

func (sls *sizeLifeSpanner) startTicker(ctx context.Context, path string, maxSize int64) {
	ct := 0
	for {
		ct++
		select {
		case <-time.After(time.Minute):
			log.Info("Ticked once", "path", path, "ct", ct)
			fi, err := os.Stat(path)
			if err != nil {
				log.LogIfError(err)
				continue
			}
			size := fi.Size()
			log.Info("Ticked inside", "maxSize", maxSize, "size", size)
			if size > maxSize {
				sls.tickChannel <- ""
			}
		case <-ctx.Done():
			log.Debug("closing secondsLifeSpanner go routine")
			return
		}
	}
}
