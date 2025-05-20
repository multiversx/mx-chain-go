package antiflood

import "github.com/multiversx/mx-chain-go/process"

func (af *p2pAntiflood) Debugger() process.AntifloodDebugger {
	af.mutDebugger.RLock()
	defer af.mutDebugger.RUnlock()

	return af.debugger
}
