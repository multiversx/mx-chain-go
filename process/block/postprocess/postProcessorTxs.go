package postprocess

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.PostProcessorTxsHandler = (*postProcessorTxs)(nil)

type postProcessorTxs struct {
	mutMapPostProcessorTxs sync.RWMutex
	mapPostProcessorTxs    map[string]struct{}
}

// NewPostProcessorTxs creates a new postProcessorTxs object
func NewPostProcessorTxs() (*postProcessorTxs, error) {
	postProcessorTxs := &postProcessorTxs{
		mapPostProcessorTxs: make(map[string]struct{}),
	}

	return postProcessorTxs, nil
}

// Init resets the post processor txs map
func (ppt *postProcessorTxs) Init() {
	ppt.mutMapPostProcessorTxs.Lock()
	defer ppt.mutMapPostProcessorTxs.Unlock()

	ppt.mapPostProcessorTxs = make(map[string]struct{})
}

// AddPostProcessorTx adds the given tx hash in the post processor txs map
func (ppt *postProcessorTxs) AddPostProcessorTx(txHash []byte) bool {
	ppt.mutMapPostProcessorTxs.Lock()
	defer ppt.mutMapPostProcessorTxs.Unlock()

	if ppt.isPostProcessorTxAdded(txHash) {
		return false
	}

	ppt.mapPostProcessorTxs[string(txHash)] = struct{}{}
	return true
}

// IsPostProcessorTxAdded returns true if the given tx hash has been already added in the post processor txs map, otherwise it returns false
func (ppt *postProcessorTxs) IsPostProcessorTxAdded(txHash []byte) bool {
	ppt.mutMapPostProcessorTxs.RLock()
	defer ppt.mutMapPostProcessorTxs.RUnlock()

	return ppt.isPostProcessorTxAdded(txHash)
}

func (ppt *postProcessorTxs) isPostProcessorTxAdded(txHash []byte) bool {
	_, ok := ppt.mapPostProcessorTxs[string(txHash)]
	return ok
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppt *postProcessorTxs) IsInterfaceNil() bool {
	return ppt == nil
}
