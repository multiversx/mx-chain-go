package postprocess

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.PostProcessorTxsHandler = (*postProcessorTxs)(nil)

type postProcessorTxs struct {
	txCoordinator          process.TransactionCoordinator
	mutMapPostProcessorTxs sync.RWMutex
	mapPostProcessorTxs    map[string]data.TransactionHandler
}

// NewPostProcessorTxs creates a new postProcessorTxs object
func NewPostProcessorTxs(
	txCoordinator process.TransactionCoordinator,
) (*postProcessorTxs, error) {
	if check.IfNil(txCoordinator) {
		return nil, process.ErrNilTransactionCoordinator
	}

	postProcessorTxs := &postProcessorTxs{
		txCoordinator:       txCoordinator,
		mapPostProcessorTxs: make(map[string]data.TransactionHandler),
	}

	return postProcessorTxs, nil
}

// Init resets the post processor txs map
func (ppt *postProcessorTxs) Init() {
	ppt.mutMapPostProcessorTxs.Lock()
	defer ppt.mutMapPostProcessorTxs.Unlock()

	ppt.mapPostProcessorTxs = make(map[string]data.TransactionHandler)
}

// AddPostProcessorTx adds the given tx in the post processor txs map
func (ppt *postProcessorTxs) AddPostProcessorTx(txHash []byte, txHandler data.TransactionHandler) bool {
	ppt.mutMapPostProcessorTxs.Lock()
	defer ppt.mutMapPostProcessorTxs.Unlock()

	if ppt.isPostProcessorTxAdded(txHash) {
		return false
	}

	ppt.mapPostProcessorTxs[string(txHash)] = txHandler
	return true
}

// GetPostProcessorTx gets the tx from the post processor txs map with the given tx hash
func (ppt *postProcessorTxs) GetPostProcessorTx(txHash []byte) data.TransactionHandler {
	ppt.mutMapPostProcessorTxs.RLock()
	defer ppt.mutMapPostProcessorTxs.RUnlock()

	return ppt.mapPostProcessorTxs[string(txHash)]
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

// SetTransactionCoordinator sets the transaction coordinator needed by scheduled txs execution component
func (ppt *postProcessorTxs) SetTransactionCoordinator(txCoordinator process.TransactionCoordinator) {
	ppt.txCoordinator = txCoordinator
}

// GetProcessedResults gets all the intermediate transactions, since the last init, separated by block type
func (ppt *postProcessorTxs) GetProcessedResults() map[block.Type]map[uint32][]*process.TxInfo {
	return ppt.txCoordinator.GetProcessedResults()
}

// InitProcessedResults initializes the processed results
func (ppt *postProcessorTxs) InitProcessedResults() {
	ppt.txCoordinator.InitProcessedResults()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppt *postProcessorTxs) IsInterfaceNil() bool {
	return ppt == nil
}
