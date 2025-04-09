package process

import (
	"fmt"
	"sync"
)

const maxHistory = 10

type ReceiverData struct {
	mu          sync.RWMutex
	estimations [maxHistory]float64
	index       int
	count       int
	sum         float64
}

var receiverMgr *ReceiverManager

func GetReceiverManager() *ReceiverManager {
	if receiverMgr == nil {
		receiverMgr = &ReceiverManager{
			receivers: make(map[string]*ReceiverData),
		}
	}
	return receiverMgr
}

// ReceiverManager manages multiple receivers and their estimation history.
type ReceiverManager struct {
	receivers map[string]*ReceiverData
	mu        sync.RWMutex
}

func NewReceiverManager() *ReceiverManager {
	return &ReceiverManager{
		receivers: make(map[string]*ReceiverData),
	}
}

// UpdateEstimation adds a new estimation to the receiver’s history.
func (rm *ReceiverManager) UpdateEstimation(name string, estimation float64) {
	rm.mu.RLock()
	rcv, exists := rm.receivers[name]
	rm.mu.RUnlock()

	if !exists {
		rm.mu.Lock()
		rcv = &ReceiverData{}
		rm.receivers[name] = rcv
		rm.mu.Unlock()
	}

	rcv.mu.Lock()
	old := rcv.estimations[rcv.index]

	if rcv.count < maxHistory {
		rcv.sum += estimation
		rcv.count++
	} else {
		rcv.sum += estimation - old
	}

	rcv.estimations[rcv.index] = estimation
	rcv.index = (rcv.index + 1) % maxHistory
	rcv.mu.Unlock()
}

// PredictEstimation returns the moving average prediction.
func (rm *ReceiverManager) PredictEstimation(name string) float64 {
	rm.mu.RLock()
	rcv, exists := rm.receivers[name]
	rm.mu.RUnlock()

	if !exists || rcv.count == 0 {
		return 1.0
	}

	rcv.mu.RLock()
	defer rcv.mu.RUnlock()

	computedIndex := (rcv.index - 1 + maxHistory) % maxHistory
	// last value is 2x more important
	return (rcv.sum + rcv.estimations[computedIndex]) / (float64(rcv.count) + 1)
}

// SimulateAsyncUpdates simulates estimation updates coming in asynchronously.
func SimulateAsyncUpdates(rm *ReceiverManager, updates chan [2]interface{}) {
	for update := range updates {
		name := update[0].(string)
		estimation := update[1].(float64)
		rm.UpdateEstimation(name, estimation)
		fmt.Printf("Updated %s with estimation: %.2f | Predicted: %.2f\n", name, estimation, rm.PredictEstimation(name))
	}
}

//func main() {
//	manager := NewReceiverManager()
//	updates := make(chan [2]interface{})
//
//	go SimulateAsyncUpdates(manager, updates)
//
//	receivers := []string{"Compute", "Storage", "Networking"}
//	for i := 1; i <= 150; i++ {
//		for _, name := range receivers {
//			updates <- [2]interface{}{name, float64(i) + float64(len(name))}
//		}
//	}
//
//	close(updates)
//}
