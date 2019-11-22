package mock

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type ConsensusCoreMock struct {
	blockChain             data.ChainHandler
	blockProcessor         process.BlockProcessor
	bootstrapper           process.Bootstrapper
	broadcastMessenger     consensus.BroadcastMessenger
	chronologyHandler      consensus.ChronologyHandler
	hasher                 hashing.Hasher
	marshalizer            marshal.Marshalizer
	blsPrivateKey          crypto.PrivateKey
	blsSingleSigner        crypto.SingleSigner
	multiSigner            crypto.MultiSigner
	rounder                consensus.Rounder
	shardCoordinator       sharding.Coordinator
	syncTimer              ntp.SyncTimer
	validatorGroupSelector sharding.NodesCoordinator
}

func (ccm *ConsensusCoreMock) Blockchain() data.ChainHandler {
	return ccm.blockChain
}

func (ccm *ConsensusCoreMock) BlockProcessor() process.BlockProcessor {
	return ccm.blockProcessor
}

func (ccm *ConsensusCoreMock) BootStrapper() process.Bootstrapper {
	return ccm.bootstrapper
}

func (ccm *ConsensusCoreMock) BroadcastMessenger() consensus.BroadcastMessenger {
	return ccm.broadcastMessenger
}

func (ccm *ConsensusCoreMock) Chronology() consensus.ChronologyHandler {
	return ccm.chronologyHandler
}

func (ccm *ConsensusCoreMock) Hasher() hashing.Hasher {
	return ccm.hasher
}

func (ccm *ConsensusCoreMock) Marshalizer() marshal.Marshalizer {
	return ccm.marshalizer
}

func (ccm *ConsensusCoreMock) MultiSigner() crypto.MultiSigner {
	return ccm.multiSigner
}

func (ccm *ConsensusCoreMock) Rounder() consensus.Rounder {
	return ccm.rounder
}

func (ccm *ConsensusCoreMock) ShardCoordinator() sharding.Coordinator {
	return ccm.shardCoordinator
}

func (ccm *ConsensusCoreMock) SyncTimer() ntp.SyncTimer {
	return ccm.syncTimer
}

func (ccm *ConsensusCoreMock) NodesCoordinator() sharding.NodesCoordinator {
	return ccm.validatorGroupSelector
}

func (ccm *ConsensusCoreMock) SetBlockchain(blockChain data.ChainHandler) {
	ccm.blockChain = blockChain
}

func (ccm *ConsensusCoreMock) SetSingleSigner(signer crypto.SingleSigner) {
	ccm.blsSingleSigner = signer
}

func (ccm *ConsensusCoreMock) SetBlockProcessor(blockProcessor process.BlockProcessor) {
	ccm.blockProcessor = blockProcessor
}

func (ccm *ConsensusCoreMock) SetBootStrapper(bootstrapper process.Bootstrapper) {
	ccm.bootstrapper = bootstrapper
}

func (ccm *ConsensusCoreMock) SetBroadcastMessenger(broadcastMessenger consensus.BroadcastMessenger) {
	ccm.broadcastMessenger = broadcastMessenger
}

func (ccm *ConsensusCoreMock) SetChronology(chronologyHandler consensus.ChronologyHandler) {
	ccm.chronologyHandler = chronologyHandler
}

func (ccm *ConsensusCoreMock) SetHasher(hasher hashing.Hasher) {
	ccm.hasher = hasher
}

func (ccm *ConsensusCoreMock) SetMarshalizer(marshalizer marshal.Marshalizer) {
	ccm.marshalizer = marshalizer
}

func (ccm *ConsensusCoreMock) SetMultiSigner(multiSigner crypto.MultiSigner) {
	ccm.multiSigner = multiSigner
}

func (ccm *ConsensusCoreMock) SetRounder(rounder consensus.Rounder) {
	ccm.rounder = rounder
}
func (ccm *ConsensusCoreMock) SetShardCoordinator(shardCoordinator sharding.Coordinator) {
	ccm.shardCoordinator = shardCoordinator
}

func (ccm *ConsensusCoreMock) SetSyncTimer(syncTimer ntp.SyncTimer) {
	ccm.syncTimer = syncTimer
}

func (ccm *ConsensusCoreMock) SetValidatorGroupSelector(validatorGroupSelector sharding.NodesCoordinator) {
	ccm.validatorGroupSelector = validatorGroupSelector
}

func (ccm *ConsensusCoreMock) PrivateKey() crypto.PrivateKey {
	return ccm.blsPrivateKey
}

// SingleSigner returns the bls single signer stored in the ConsensusStore
func (ccm *ConsensusCoreMock) SingleSigner() crypto.SingleSigner {
	return ccm.blsSingleSigner
}

// IsInterfaceNil returns true if there is no value under the interface
func (ccm *ConsensusCoreMock) IsInterfaceNil() bool {
	return ccm == nil
}
