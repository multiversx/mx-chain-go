package statusHandler

import "github.com/ElrondNetwork/elrond-go/termuic"

// TermuiStatusHandler will be used when an AppStatusHandler is required, but another one isn't necessary or available
type TermuiStatusHandler struct {
	tui *termuic.TermuiConsole
}

// NewTermuiStatusHandler will return an instance of the struct
func NewTermuiStatusHandler() *TermuiStatusHandler {

	tsh := new(TermuiStatusHandler)
	tsh.tui = termuic.NewTermuiConsole()
	tsh.tui.Start()
	return tsh
}

//Termui method - returns address of TermuiConsole structure from TermuiStatusHandler
func (tsh *TermuiStatusHandler) Termui() (*termuic.TermuiConsole, error) {
	if tsh.tui == nil {
		return nil, ErrNilTermuiConsole
	}
	return tsh.tui, nil
}

// Increment method - will increment the value for a key
func (tsh *TermuiStatusHandler) Increment(key string) {
	tsh.tui.Increment(key)
}

// Decrement method - will decrement the value for a key
func (tsh *TermuiStatusHandler) Decrement(key string) {
	return
}

// SetInt64Value method - will update the value for a key
func (tsh *TermuiStatusHandler) SetInt64Value(key string, value int64) {
	tsh.tui.SetInt64Value(key, value)
}

// SetUInt64Value method - will update the value for a key
func (tsh *TermuiStatusHandler) SetUInt64Value(key string, value uint64) {
	tsh.tui.SetUInt64Value(key, value)
}

// SetStringValue - will update the value for a key
func (tsh *TermuiStatusHandler) SetStringValue(key string, value string) {
	tsh.tui.SetStringValue(key, value)
}

// Close method - won't do anything
func (tsh *TermuiStatusHandler) Close() {
}
