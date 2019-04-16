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
	consensusState         *ConsensusState
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
	consensusState *ConsensusState,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	multiSigner crypto.MultiSigner,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	syncTimer ntp.SyncTimer,
	validatorGroupSelector consensus.ValidatorGroupSelector) *ConsensusDataContainer {

	err := checkNewFactoryParams(
		blockChain,
		blockProcessor,
		bootstraper,
		chronologyHandler,
		consensusState,
		hasher,
		marshalizer,
		multiSigner,
		rounder,
		shardCoordinator,
		syncTimer,
		validatorGroupSelector,
	)

	if err != nil {
		return nil
	}

	consensusDataContainer := ConsensusDataContainer{
		blockChain:             blockChain,
		blockProcessor:         blockProcessor,
		bootstraper:            bootstraper,
		chronologyHandler:      chronologyHandler,
		consensusState:         consensusState,
		hasher:                 hasher,
		marshalizer:            marshalizer,
		multiSigner:            multiSigner,
		rounder:                rounder,
		shardCoordinator:       shardCoordinator,
		syncTimer:              syncTimer,
		validatorGroupSelector: validatorGroupSelector,
	}

	return &consensusDataContainer
}

func checkNewFactoryParams(
	blockChain data.ChainHandler,
	blockProcessor process.BlockProcessor,
	bootstraper process.Bootstrapper,
	chronologyHandler consensus.ChronologyHandler,
	consensusState *ConsensusState,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	multiSigner crypto.MultiSigner,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	syncTimer ntp.SyncTimer,
	validatorGroupSelector consensus.ValidatorGroupSelector,
) error {
	if blockChain == nil {
		return ErrNilBlockChain
	}

	if blockProcessor == nil {
		return ErrNilBlockProcessor
	}

	if bootstraper == nil {
		return ErrNilBlootstraper
	}

	if chronologyHandler == nil {
		return ErrNilChronologyHandler
	}

	if consensusState == nil {
		return ErrNilConsensusState
	}

	if hasher == nil {
		return ErrNilHasher
	}

	if marshalizer == nil {
		return ErrNilMarshalizer
	}

	if multiSigner == nil {
		return ErrNilMultiSigner
	}

	if rounder == nil {
		return ErrNilRounder
	}

	if shardCoordinator == nil {
		return ErrNilShardCoordinator
	}

	if syncTimer == nil {
		return ErrNilSyncTimer
	}

	if validatorGroupSelector == nil {
		return ErrNilValidatorGroupSelector
	}

	return nil
}

func (cdc *ConsensusDataContainer) GetChainHandler() data.ChainHandler {
	return cdc.blockChain
}
func (cdc *ConsensusDataContainer) GetBlockProcessor() process.BlockProcessor {
	return cdc.blockProcessor
}
func (cdc *ConsensusDataContainer) GetBootStrapper() process.Bootstrapper {
	return cdc.bootstraper
}
func (cdc *ConsensusDataContainer) GetChronology() consensus.ChronologyHandler {
	return cdc.chronologyHandler
}
func (cdc *ConsensusDataContainer) GetConsensusState() *ConsensusState {
	return cdc.consensusState
}
func (cdc *ConsensusDataContainer) GetHasher() hashing.Hasher {
	return cdc.hasher
}
func (cdc *ConsensusDataContainer) GetMarshalizer() marshal.Marshalizer {
	return cdc.marshalizer
}
func (cdc *ConsensusDataContainer) GetMultiSigner() crypto.MultiSigner {
	return cdc.multiSigner
}
func (cdc *ConsensusDataContainer) GetRounder() consensus.Rounder {
	return cdc.rounder
}
func (cdc *ConsensusDataContainer) GetShardCoordinator() sharding.Coordinator {
	return cdc.shardCoordinator
}
func (cdc *ConsensusDataContainer) GetSyncTimer() ntp.SyncTimer {
	return cdc.syncTimer
}
func (cdc *ConsensusDataContainer) GetValidatorGroupSelector() consensus.ValidatorGroupSelector {
	return cdc.validatorGroupSelector
}
