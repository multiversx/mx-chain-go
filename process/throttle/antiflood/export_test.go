package antiflood

import "github.com/multiversx/mx-chain-go/process"

func (af *p2pAntiflood) Debugger() process.AntifloodDebugger {
	return af.debugger
}
