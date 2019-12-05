package antiflood

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const minOperations = 1

// countersMap represents a cache of counters used in antiflooding mechanism
type countersMap struct {
	mutOperation  sync.Mutex
	cacher        storage.Cacher
	maxOperations int
}

// NewCountersMap creates a new countersMap instance
func NewCountersMap(cacher storage.Cacher, maxOperations int) (*countersMap, error) {
	if cacher == nil {
		return nil, process.ErrNilCacher
	}
	if maxOperations < minOperations {
		return nil,
			fmt.Errorf("%w raised in NewCountersMap, provided %d, minimum %d",
				process.ErrInvalidValue,
				maxOperations,
				minOperations,
			)
	}

	return &countersMap{
		cacher:        cacher,
		maxOperations: maxOperations,
	}, nil
}

// TryIncrement tries to increment the counter value held at "identifier" position
// It returns true if it had succeeded (existing counter value is lower or equal with provided maxOperations)
func (cm *countersMap) TryIncrement(identifier string) bool {
	//we need the mutOperation here as the get and put should be done atomically.
	// Otherwise we might yield a slightly higher number of false valid increments
	cm.mutOperation.Lock()
	defer cm.mutOperation.Unlock()

	value, ok := cm.cacher.Get([]byte(identifier))
	if !ok {
		cm.cacher.Put([]byte(identifier), 1)
		return true
	}

	intVal, isInt := value.(int)
	if !isInt {
		cm.cacher.Put([]byte(identifier), 1)
		return true
	}

	if intVal < cm.maxOperations {
		cm.cacher.Put([]byte(identifier), intVal+1)
		return true
	}

	return false
}

// Reset clears all map values
func (cm *countersMap) Reset() {
	cm.mutOperation.Lock()
	defer cm.mutOperation.Unlock()

	//TODO change this if cacher.Clear() is time consuming
	cm.cacher.Clear()
}

// IsInterfaceNil returns true if there is no value under the interface
func (cm *countersMap) IsInterfaceNil() bool {
	return cm == nil
}
