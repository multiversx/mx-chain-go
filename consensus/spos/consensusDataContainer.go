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

type ConsensusDataContainer struct {
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

func NewConsensusDataContainer(
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
	validatorGroupSelector consensus.ValidatorGroupSelector) (*ConsensusDataContainer, error) {

	consensusContainer := &ConsensusDataContainer{
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

	consensusContainerValidator := ConsensusContainerValidator{}
	err := consensusContainerValidator.ValidateConsensusDataContainer(consensusContainer)

	if err != nil {
		return nil, err
	}
	return consensusContainer, nil
}

func (cdc *ConsensusDataContainer) Blockchain() data.ChainHandler {
	return cdc.blockChain
}
func (cdc *ConsensusDataContainer) BlockProcessor() process.BlockProcessor {
	return cdc.blockProcessor
}
func (cdc *ConsensusDataContainer) BootStrapper() process.Bootstrapper {
	return cdc.bootstraper
}
func (cdc *ConsensusDataContainer) Chronology() consensus.ChronologyHandler {
	return cdc.chronologyHandler
}
func (cdc *ConsensusDataContainer) Hasher() hashing.Hasher {
	return cdc.hasher
}
func (cdc *ConsensusDataContainer) Marshalizer() marshal.Marshalizer {
	return cdc.marshalizer
}
func (cdc *ConsensusDataContainer) MultiSigner() crypto.MultiSigner {
	return cdc.multiSigner
}
func (cdc *ConsensusDataContainer) Rounder() consensus.Rounder {
	return cdc.rounder
}
func (cdc *ConsensusDataContainer) ShardCoordinator() sharding.Coordinator {
	return cdc.shardCoordinator
}
func (cdc *ConsensusDataContainer) SyncTimer() ntp.SyncTimer {
	return cdc.syncTimer
}
func (cdc *ConsensusDataContainer) ValidatorGroupSelector() consensus.ValidatorGroupSelector {
	return cdc.validatorGroupSelector
}
