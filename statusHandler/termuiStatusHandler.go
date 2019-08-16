package statusHandler

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/statusHandler/termuic"
)

// TermuiStatusHandler will be used when an AppStatusHandler is required, but another one isn't necessary or available
type TermuiStatusHandler struct {
	tui                  *termuic.TermuiConsole
	termuiConsoleMetrics *sync.Map
}

// NewTermuiStatusHandler will return an instance of the struct
func NewTermuiStatusHandler() *TermuiStatusHandler {
	tsh := &TermuiStatusHandler{
		termuiConsoleMetrics: &sync.Map{},
	}
	tsh.tui = termuic.NewTermuiConsole(tsh.termuiConsoleMetrics)

	return tsh
}

// IsInterfaceNil return if there is no value under the interface
func (tsh *TermuiStatusHandler) IsInterfaceNil() bool {
	if tsh == nil {
		return true
	}

	return false
}

// StartTermuiConsole method will start termui console
func (tsh *TermuiStatusHandler) StartTermuiConsole() error {
	err := tsh.tui.Start()

	return err
}

//Termui method - returns address of TermuiConsole structure from TermuiStatusHandler
func (tsh *TermuiStatusHandler) Termui() *termuic.TermuiConsole {
	return tsh.tui
}

// SetInt64Value method - will update the value for a key
func (tsh *TermuiStatusHandler) SetInt64Value(key string, value int64) {
	tsh.termuiConsoleMetrics.Store(key, value)
}

// SetUInt64Value method - will update the value for a key
func (tsh *TermuiStatusHandler) SetUInt64Value(key string, value uint64) {
	tsh.termuiConsoleMetrics.Store(key, value)
}

// SetStringValue method - will update the value of a key
func (tsh *TermuiStatusHandler) SetStringValue(key string, value string) {
	tsh.termuiConsoleMetrics.Store(key, value)
}

// Increment - will increment the value of a key
func (tsh *TermuiStatusHandler) Increment(key string) {
	keyValueI, ok := tsh.termuiConsoleMetrics.Load(key)
	if !ok {
		return
	}

	keyValue, ok := keyValueI.(uint64)
	if !ok {
		return
	}

	keyValue++
	tsh.termuiConsoleMetrics.Store(key, keyValue)
}

// Decrement - will decrement the value of a key
func (tsh *TermuiStatusHandler) Decrement(key string) {
}

// Close method - won't do anything
func (tsh *TermuiStatusHandler) Close() {
}
