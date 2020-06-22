package antiflood

import "github.com/ElrondNetwork/elrond-go/process"

func (af *p2pAntiflood) Debugger() process.AntifloodDebugger {
	return af.debugger
}
