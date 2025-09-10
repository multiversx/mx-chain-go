package coordinator

import (
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/process"
)

type preProcData struct {
	mutPreProcessor sync.RWMutex
	txPreProcessors map[block.Type]process.PreProcessor
	keysTxPreProcs  []block.Type
}

func NewPreProcData(container process.PreProcessorsContainer) (*preProcData, error) {
	ppd := &preProcData{}
	ppd.txPreProcessors = make(map[block.Type]process.PreProcessor)
	ppd.keysTxPreProcs = container.Keys()
	sort.Slice(ppd.keysTxPreProcs, func(i, j int) bool {
		return ppd.keysTxPreProcs[i] < ppd.keysTxPreProcs[j]
	})
	for _, value := range ppd.keysTxPreProcs {
		preProc, errGet := container.Get(value)
		if errGet != nil {
			return nil, errGet
		}
		ppd.txPreProcessors[value] = preProc
	}

	return ppd, nil
}

func (ppd *preProcData) getPreProcessor(blockType block.Type) process.PreProcessor {
	ppd.mutPreProcessor.RLock()
	preprocessor, exists := ppd.txPreProcessors[blockType]
	ppd.mutPreProcessor.RUnlock()

	if !exists {
		return nil
	}

	return preprocessor
}
