package sharding

import "github.com/multiversx/mx-chain-core-go/core"

type NodesSetupArgs struct {
	NodesFilePath            string
	AddressPubKeyConverter   core.PubkeyConverter
	ValidatorPubKeyConverter core.PubkeyConverter
	GenesisMaxNumShards      uint32
}
