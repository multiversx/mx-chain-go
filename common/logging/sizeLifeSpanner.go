package logging

import (
	"context"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-logger/check"
)

const minMBLifeSpan = 1
const minRefreshInterval = time.Second

type sizeLifeSpanner struct {
	*baseLifeSpanner
	spanInMB        uint32
	refreshInterVal time.Duration
	cancelFunc      context.CancelFunc
	currentFile     string
	fileSizeChecker FileSizeCheckHandler
}

func newSizeLifeSpanner(fileSizeChecker FileSizeCheckHandler, sizeLifeSpanInMB uint32, refreshInterval time.Duration) (*sizeLifeSpanner, error) {
	if check.IfNil(fileSizeChecker) {
		return nil, fmt.Errorf("newSizeLifeSpanner %w, nil file size checker", core.ErrInvalidLogFileMinLifeSpan)
	}

	if sizeLifeSpanInMB < minMBLifeSpan {
		return nil, fmt.Errorf("newSizeLifeSpanner %w, provided size %v, min %v MB", core.ErrInvalidLogFileMinLifeSpan, sizeLifeSpanInMB, minMBLifeSpan)
	}

	if refreshInterval < minRefreshInterval {
		return nil, fmt.Errorf("newSizeLifeSpanner %w, provided refreshInterval %v, min %v", core.ErrInvalidLogFileMinLifeSpan, refreshInterval, minRefreshInterval)
	}

	sls := &sizeLifeSpanner{
		spanInMB:        sizeLifeSpanInMB * 1024 * 1024,
		baseLifeSpanner: newBaseLifeSpanner(),
		fileSizeChecker: fileSizeChecker,
	}

	return sls, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sls *sizeLifeSpanner) IsInterfaceNil() bool {
	return sls == nil
}

// SetCurrentFile sets the file need for monitoring for the size
func (sls *sizeLifeSpanner) SetCurrentFile(path string) {
	if sls.cancelFunc != nil {
		sls.cancelFunc()
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	sls.cancelFunc = cancelFunc

	go sls.startTicker(ctx, path, int64(sls.spanInMB))
}

func (sls *sizeLifeSpanner) startTicker(ctx context.Context, path string, maxSize int64) {
	for {
		select {
		case <-time.After(sls.refreshInterVal):
			size, err := sls.fileSizeChecker.GetSize(path)
			log.LogIfError(err)
			if size > maxSize {
				sls.lifeSpanChannel <- ""
			}
		case <-ctx.Done():
			log.Debug("closing sizeLifeSpanner go routine")
			return
		}
	}
}
