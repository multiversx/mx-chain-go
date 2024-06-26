package presenter

import (
	"math/big"
	"strings"
	"sync"
)

// maxLogLines is used to specify how many lines of logs need to store in slice
var maxLogLines = 100

// PresenterStatusHandler is the AppStatusHandler impl that is able to process and store received data
type PresenterStatusHandler struct {
	presenterMetrics            map[string]interface{}
	mutPresenterMap             sync.RWMutex
	logLines                    []string
	mutLogLineWrite             sync.RWMutex
	oldRound                    uint64
	synchronizationSpeedHistory []uint64
	totalRewardsOld             *big.Float
}

// NewPresenterStatusHandler will return an instance of the struct
func NewPresenterStatusHandler() *PresenterStatusHandler {
	psh := &PresenterStatusHandler{
		presenterMetrics:            make(map[string]interface{}),
		synchronizationSpeedHistory: make([]uint64, 0),
		totalRewardsOld:             big.NewFloat(0),
	}
	return psh
}

// IsInterfaceNil returns true if there is no value under the interface
func (psh *PresenterStatusHandler) IsInterfaceNil() bool {
	return psh == nil
}

// InvalidateCache will clear the entire map
func (psh *PresenterStatusHandler) InvalidateCache() {
	psh.mutPresenterMap.Lock()
	psh.presenterMetrics = make(map[string]interface{})
	psh.mutPresenterMap.Unlock()
}

// SetInt64Value method - will update the value for a key
func (psh *PresenterStatusHandler) SetInt64Value(key string, value int64) {
	psh.mutPresenterMap.Lock()
	psh.presenterMetrics[key] = value
	psh.mutPresenterMap.Unlock()
}

// SetUInt64Value method - will update the value for a key
func (psh *PresenterStatusHandler) SetUInt64Value(key string, value uint64) {
	psh.mutPresenterMap.Lock()
	psh.presenterMetrics[key] = value
	psh.mutPresenterMap.Unlock()
}

// SetStringValue method - will update the value of a key
func (psh *PresenterStatusHandler) SetStringValue(key string, value string) {
	psh.mutPresenterMap.Lock()
	psh.presenterMetrics[key] = value
	psh.mutPresenterMap.Unlock()
}

// Increment - will increment the value of a key
func (psh *PresenterStatusHandler) Increment(key string) {
	psh.mutPresenterMap.Lock()
	defer psh.mutPresenterMap.Unlock()

	keyValueI, ok := psh.presenterMetrics[key]
	if !ok {
		return
	}

	keyValue, ok := keyValueI.(uint64)
	if !ok {
		return
	}

	keyValue++
	psh.presenterMetrics[key] = keyValue
}

// AddUint64 - will increase the value of a key with a value
func (psh *PresenterStatusHandler) AddUint64(key string, value uint64) {
	psh.mutPresenterMap.Lock()
	defer psh.mutPresenterMap.Unlock()

	keyValueI, ok := psh.presenterMetrics[key]
	if !ok {
		return
	}

	keyValue, ok := keyValueI.(uint64)
	if !ok {
		return
	}

	keyValue += value
	psh.presenterMetrics[key] = keyValue
}

// Decrement - will decrement the value of a key
func (psh *PresenterStatusHandler) Decrement(key string) {
	psh.mutPresenterMap.Lock()
	defer psh.mutPresenterMap.Unlock()

	keyValueI, ok := psh.presenterMetrics[key]
	if !ok {
		return
	}

	keyValue, ok := keyValueI.(uint64)
	if !ok {
		return
	}
	if keyValue == 0 {
		return
	}

	keyValue--
	psh.presenterMetrics[key] = keyValue
}

// Close method - won't do anything
func (psh *PresenterStatusHandler) Close() {
}

func (psh *PresenterStatusHandler) Write(p []byte) (n int, err error) {
	go func(p []byte) {
		logLine := string(p)
		stringSlice := strings.Split(logLine, "\n")

		psh.mutLogLineWrite.Lock()
		for _, line := range stringSlice {
			line = strings.Replace(line, "\r", "", len(line))
			if line != "" {
				psh.logLines = append(psh.logLines, line)
			}
		}

		startPos := len(psh.logLines) - maxLogLines
		if startPos < 0 {
			startPos = 0
		}
		psh.logLines = psh.logLines[startPos:len(psh.logLines)]

		psh.mutLogLineWrite.Unlock()
	}(p)

	return len(p), nil
}

// GetLogLines will return log lines that's need to be displayed
func (psh *PresenterStatusHandler) GetLogLines() []string {
	psh.mutLogLineWrite.RLock()
	defer psh.mutLogLineWrite.RUnlock()

	return psh.logLines
}
