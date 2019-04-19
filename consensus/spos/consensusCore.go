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

type ConsensusCore struct {
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

//NewConsensusCore creates a new ConsensusCore instance
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
	validatorGroupSelector consensus.ValidatorGroupSelector) (*ConsensusCore, error) {

	consensusCore := &ConsensusCore{
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

//Blockchain gets the ChainHandler stored in the ConsensusCore
func (cdc *ConsensusCore) Blockchain() data.ChainHandler {
	return cdc.blockChain
}

//BlockProcessor gets the BlockProcessor stored in the ConsensusCore
func (cdc *ConsensusCore) BlockProcessor() process.BlockProcessor {
	return cdc.blockProcessor
}

//BootStrapper gets the Bootstrapper stored in the ConsensusCore
func (cdc *ConsensusCore) BootStrapper() process.Bootstrapper {
	return cdc.bootstraper
}

//Chronology gets the ChronologyHandler stored in the ConsensusCore
func (cdc *ConsensusCore) Chronology() consensus.ChronologyHandler {
	return cdc.chronologyHandler
}

//Hasher gets the Hasher stored in the ConsensusCore
func (cdc *ConsensusCore) Hasher() hashing.Hasher {
	return cdc.hasher
}

//Marshalizer gets the Marshalizer stored in the ConsensusCore
func (cdc *ConsensusCore) Marshalizer() marshal.Marshalizer {
	return cdc.marshalizer
}

//MultiSigner gets the MultiSigner stored in the ConsensusCore
func (cdc *ConsensusCore) MultiSigner() crypto.MultiSigner {
	return cdc.multiSigner
}

//Rounder gets the Rounder stored in the ConsensusCore
func (cdc *ConsensusCore) Rounder() consensus.Rounder {
	return cdc.rounder
}

//ShardCoordinator gets the Coordinator stored in the ConsensusCore
func (cdc *ConsensusCore) ShardCoordinator() sharding.Coordinator {
	return cdc.shardCoordinator
}

//SyncTimer gets the SyncTimer stored in the ConsensusCore
func (cdc *ConsensusCore) SyncTimer() ntp.SyncTimer {
	return cdc.syncTimer
}

//ValidatorGroupSelector gets the ValidatorGroupSelector stored in the ConsensusCore
func (cdc *ConsensusCore) ValidatorGroupSelector() consensus.ValidatorGroupSelector {
	return cdc.validatorGroupSelector
}
