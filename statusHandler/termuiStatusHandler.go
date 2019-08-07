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
func (psh *TermuiStatusHandler) Termui() *termuic.TermuiConsole {
	return psh.tui
}

// Increment method - won't do anything
func (psh *TermuiStatusHandler) Increment(key string) {
	return
}

// Decrement method - won't do anything
func (psh *TermuiStatusHandler) Decrement(key string) {
	return
}

// SetInt64Value method - won't do anything
func (psh *TermuiStatusHandler) SetInt64Value(key string, value int64) {
	psh.tui.SetInt64Value(key, value)
}

// SetUInt64Value method - won't do anything
func (psh *TermuiStatusHandler) SetUInt64Value(key string, value uint64) {
	psh.tui.SetUInt64Value(key, value)
}

// Close method - won't do anything
func (psh *TermuiStatusHandler) Close() {
}
