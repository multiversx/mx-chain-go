package disabled

import "github.com/multiversx/mx-chain-core-go/core"

// AntifloodDebugger is a disabled instance of the antoiflood debugger
type AntifloodDebugger struct {
}

// AddData does nothing
func (ad *AntifloodDebugger) AddData(_ core.PeerID, _ string, _ uint32, _ uint64, _ []byte, _ bool) {
}

// Close returns nil
func (ad *AntifloodDebugger) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ad *AntifloodDebugger) IsInterfaceNil() bool {
	return ad == nil
}
