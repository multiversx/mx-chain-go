package sharding

import "github.com/multiversx/mx-chain-core-go/core"

// NodesSetupArgs defines arguments needed to create a genesis nodes setup handler
type NodesSetupArgs struct {
	NodesFilePath            string
	AddressPubKeyConverter   core.PubkeyConverter
	ValidatorPubKeyConverter core.PubkeyConverter
	GenesisMaxNumShards      uint32
}
