package consensus

import "github.com/multiversx/mx-chain-go/process"

func (cc *consensusComponents) BootStrapper() process.Bootstrapper {
	return cc.bootstrapper
}
