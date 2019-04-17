package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

type ConsensusDataContainerMock struct {
	blockChain             data.ChainHandler
	blockProcessor         process.BlockProcessor
	bootstraper            process.Bootstrapper
	chronologyHandler      consensus.ChronologyHandler
	consensusState         *spos.ConsensusState
	hasher                 hashing.Hasher
	marshalizer            marshal.Marshalizer
	multiSigner            crypto.MultiSigner
	rounder                consensus.Rounder
	shardCoordinator       sharding.Coordinator
	syncTimer              ntp.SyncTimer
	validatorGroupSelector consensus.ValidatorGroupSelector
}

func (cdc *ConsensusDataContainerMock) Blockchain() data.ChainHandler {
	return cdc.blockChain
}
func (cdc *ConsensusDataContainerMock) BlockProcessor() process.BlockProcessor {
	return cdc.blockProcessor
}
func (cdc *ConsensusDataContainerMock) BootStrapper() process.Bootstrapper {
	return cdc.bootstraper
}
func (cdc *ConsensusDataContainerMock) Chronology() consensus.ChronologyHandler {
	return cdc.chronologyHandler
}
func (cdc *ConsensusDataContainerMock) ConsensusState() *spos.ConsensusState {
	return cdc.consensusState
}
func (cdc *ConsensusDataContainerMock) Hasher() hashing.Hasher {
	return cdc.hasher
}
func (cdc *ConsensusDataContainerMock) Marshalizer() marshal.Marshalizer {
	return cdc.marshalizer
}
func (cdc *ConsensusDataContainerMock) MultiSigner() crypto.MultiSigner {
	return cdc.multiSigner
}
func (cdc *ConsensusDataContainerMock) Rounder() consensus.Rounder {
	return cdc.rounder
}
func (cdc *ConsensusDataContainerMock) ShardCoordinator() sharding.Coordinator {
	return cdc.shardCoordinator
}
func (cdc *ConsensusDataContainerMock) SyncTimer() ntp.SyncTimer {
	return cdc.syncTimer
}
func (cdc *ConsensusDataContainerMock) ValidatorGroupSelector() consensus.ValidatorGroupSelector {
	return cdc.validatorGroupSelector
}

func (cdc *ConsensusDataContainerMock) SetBlockchain(blockChain data.ChainHandler) {
	cdc.blockChain = blockChain
}
func (cdc *ConsensusDataContainerMock) SetBlockProcessor(blockProcessor process.BlockProcessor) {
	cdc.blockProcessor = blockProcessor
}
func (cdc *ConsensusDataContainerMock) SetBootStrapper(bootstraper process.Bootstrapper) {
	cdc.bootstraper = bootstraper
}
func (cdc *ConsensusDataContainerMock) SetChronology(chronologyHandler consensus.ChronologyHandler) {
	cdc.chronologyHandler = chronologyHandler
}
func (cdc *ConsensusDataContainerMock) SetConsensusState(consensusState *spos.ConsensusState) {
	cdc.consensusState = consensusState
}
func (cdc *ConsensusDataContainerMock) SetHasher(hasher hashing.Hasher) {
	cdc.hasher = hasher
}
func (cdc *ConsensusDataContainerMock) SetMarshalizer(marshalizer marshal.Marshalizer) {
	cdc.marshalizer = marshalizer
}
func (cdc *ConsensusDataContainerMock) SetMultiSigner(multiSigner crypto.MultiSigner) {
	cdc.multiSigner = multiSigner
}
func (cdc *ConsensusDataContainerMock) SetRounder(rounder consensus.Rounder) {
	cdc.rounder = rounder
}
func (cdc *ConsensusDataContainerMock) SetShardCoordinator(shardCoordinator sharding.Coordinator) {
	cdc.shardCoordinator = shardCoordinator
}
func (cdc *ConsensusDataContainerMock) SetSyncTimer(syncTimer ntp.SyncTimer) {
	cdc.syncTimer = syncTimer
}
func (cdc *ConsensusDataContainerMock) SetValidatorGroupSelector(validatorGroupSelector consensus.ValidatorGroupSelector) {
	cdc.validatorGroupSelector = validatorGroupSelector
}
