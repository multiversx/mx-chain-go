package smartContract

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
)

type scQueryServiceDispatcher struct {
	mutList     sync.RWMutex
	list        []process.SCQueryService
	mutIndex    sync.Mutex
	index       int
	maxListSize int
}

// NewScQueryServiceDispatcher returns a smart contract query service dispatcher that for each function call
// will forward the request towards the provided list in a round-robin fashion
func NewScQueryServiceDispatcher(list []process.SCQueryService) (*scQueryServiceDispatcher, error) {
	if len(list) == 0 {
		return nil, fmt.Errorf("%w in NewScQueryServiceDispatcher", process.ErrNilOrEmptyList)
	}
	for i := 0; i < len(list); i++ {
		if check.IfNil(list[i]) {
			return nil, fmt.Errorf("%w at element %d", process.ErrNilScQueryElement, i)
		}
	}

	return &scQueryServiceDispatcher{
		list:        list,
		maxListSize: len(list),
		index:       0,
	}, nil
}

// ExecuteQuery will call this method on one of the element from provided list
func (sqsd *scQueryServiceDispatcher) ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	index := sqsd.getNewIndex()

	sqsd.mutList.RLock()
	defer sqsd.mutList.RUnlock()

	return sqsd.list[index].ExecuteQuery(query)
}

// ComputeScCallGasLimit will call this method on one of the element from provided list
func (sqsd *scQueryServiceDispatcher) ComputeScCallGasLimit(tx *transaction.Transaction) (uint64, error) {
	index := sqsd.getNewIndex()

	sqsd.mutList.RLock()
	defer sqsd.mutList.RUnlock()

	return sqsd.list[index].ComputeScCallGasLimit(tx)
}

func (sqsd *scQueryServiceDispatcher) getNewIndex() int {
	sqsd.mutIndex.Lock()
	updatedValue := sqsd.index

	sqsd.index++
	sqsd.index = sqsd.index % sqsd.maxListSize
	sqsd.mutIndex.Unlock()

	return updatedValue
}

// Close closes all underlying components
//TODO(iulian) add tests
func (sqsd *scQueryServiceDispatcher) Close() error {
	sqsd.mutList.RLock()
	defer sqsd.mutList.RUnlock()

	var errFound error
	for _, scQueryService := range sqsd.list {
		err := scQueryService.Close()
		if err != nil {
			log.Error("error while closing inner SC query service in scQueryServiceDispatcher.Close", "error", err)
			errFound = err
		}
	}

	return errFound
}

// IsInterfaceNil returns true if there is no value under the interface
func (sqsd *scQueryServiceDispatcher) IsInterfaceNil() bool {
	return sqsd == nil
}
