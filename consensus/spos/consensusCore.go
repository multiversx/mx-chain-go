package spos

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

type SPOSConsensusCore struct {
	blockChain             data.ChainHandler
	blockProcessor         process.BlockProcessor
	bootstraper            process.Bootstrapper
	chronologyHandler      consensus.ChronologyHandler
	hasher                 hashing.Hasher
	marshalizer            marshal.Marshalizer
	multiSigner            crypto.MultiSigner
	rounder                consensus.Rounder
	shardCoordinator       sharding.Coordinator
	syncTimer              ntp.SyncTimer
	validatorGroupSelector consensus.ValidatorGroupSelector
}

//NewConsensusCore creates a new SPOSConsensusCore instance
func NewConsensusCore(
	blockChain data.ChainHandler,
	blockProcessor process.BlockProcessor,
	bootstraper process.Bootstrapper,
	chronologyHandler consensus.ChronologyHandler,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	multiSigner crypto.MultiSigner,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	syncTimer ntp.SyncTimer,
	validatorGroupSelector consensus.ValidatorGroupSelector) (*SPOSConsensusCore, error) {

	consensusCore := &SPOSConsensusCore{
		blockChain,
		blockProcessor,
		bootstraper,
		chronologyHandler,
		hasher,
		marshalizer,
		multiSigner,
		rounder,
		shardCoordinator,
		syncTimer,
		validatorGroupSelector,
	}

	err := ValidateConsensusCore(consensusCore)

	if err != nil {
		return nil, err
	}
	return consensusCore, nil
}

func (cdc *SPOSConsensusCore) Blockchain() data.ChainHandler {
	return cdc.blockChain
}
func (cdc *SPOSConsensusCore) BlockProcessor() process.BlockProcessor {
	return cdc.blockProcessor
}
func (cdc *SPOSConsensusCore) BootStrapper() process.Bootstrapper {
	return cdc.bootstraper
}
func (cdc *SPOSConsensusCore) Chronology() consensus.ChronologyHandler {
	return cdc.chronologyHandler
}
func (cdc *SPOSConsensusCore) Hasher() hashing.Hasher {
	return cdc.hasher
}
func (cdc *SPOSConsensusCore) Marshalizer() marshal.Marshalizer {
	return cdc.marshalizer
}
func (cdc *SPOSConsensusCore) MultiSigner() crypto.MultiSigner {
	return cdc.multiSigner
}
func (cdc *SPOSConsensusCore) Rounder() consensus.Rounder {
	return cdc.rounder
}
func (cdc *SPOSConsensusCore) ShardCoordinator() sharding.Coordinator {
	return cdc.shardCoordinator
}
func (cdc *SPOSConsensusCore) SyncTimer() ntp.SyncTimer {
	return cdc.syncTimer
}
func (cdc *SPOSConsensusCore) ValidatorGroupSelector() consensus.ValidatorGroupSelector {
	return cdc.validatorGroupSelector
}
