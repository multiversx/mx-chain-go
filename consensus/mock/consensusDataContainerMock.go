package mock

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ConsensusCoreMock -
type ConsensusCoreMock struct {
	blockChain              data.ChainHandler
	blockProcessor          process.BlockProcessor
	headersSubscriber       consensus.HeadersPoolSubscriber
	bootstrapper            process.Bootstrapper
	broadcastMessenger      consensus.BroadcastMessenger
	chronologyHandler       consensus.ChronologyHandler
	hasher                  hashing.Hasher
	marshalizer             marshal.Marshalizer
	blsPrivateKey           crypto.PrivateKey
	blsSingleSigner         crypto.SingleSigner
	multiSigner             crypto.MultiSigner
	roundHandler            consensus.RoundHandler
	shardCoordinator        sharding.Coordinator
	syncTimer               ntp.SyncTimer
	validatorGroupSelector  sharding.NodesCoordinator
	epochStartNotifier      epochStart.RegistrationHandler
	antifloodHandler        consensus.P2PAntifloodHandler
	peerHonestyHandler      consensus.PeerHonestyHandler
	headerSigVerifier       consensus.HeaderSigVerifier
	fallbackHeaderValidator consensus.FallbackHeaderValidator
	nodeRedundancyHandler   consensus.NodeRedundancyHandler
}

// GetAntiFloodHandler -
func (ccm *ConsensusCoreMock) GetAntiFloodHandler() consensus.P2PAntifloodHandler {
	return ccm.antifloodHandler
}

// Blockchain -
func (ccm *ConsensusCoreMock) Blockchain() data.ChainHandler {
	return ccm.blockChain
}

// BlockProcessor -
func (ccm *ConsensusCoreMock) BlockProcessor() process.BlockProcessor {
	return ccm.blockProcessor
}

// HeadersPoolSubscriber -
func (ccm *ConsensusCoreMock) HeadersPoolSubscriber() consensus.HeadersPoolSubscriber {
	return ccm.headersSubscriber
}

// BootStrapper -
func (ccm *ConsensusCoreMock) BootStrapper() process.Bootstrapper {
	return ccm.bootstrapper
}

// BroadcastMessenger -
func (ccm *ConsensusCoreMock) BroadcastMessenger() consensus.BroadcastMessenger {
	return ccm.broadcastMessenger
}

// Chronology -
func (ccm *ConsensusCoreMock) Chronology() consensus.ChronologyHandler {
	return ccm.chronologyHandler
}

// Hasher -
func (ccm *ConsensusCoreMock) Hasher() hashing.Hasher {
	return ccm.hasher
}

// Marshalizer -
func (ccm *ConsensusCoreMock) Marshalizer() marshal.Marshalizer {
	return ccm.marshalizer
}

// MultiSigner -
func (ccm *ConsensusCoreMock) MultiSigner() crypto.MultiSigner {
	return ccm.multiSigner
}

// RoundHandler -
func (ccm *ConsensusCoreMock) RoundHandler() consensus.RoundHandler {
	return ccm.roundHandler
}

// ShardCoordinator -
func (ccm *ConsensusCoreMock) ShardCoordinator() sharding.Coordinator {
	return ccm.shardCoordinator
}

// SyncTimer -
func (ccm *ConsensusCoreMock) SyncTimer() ntp.SyncTimer {
	return ccm.syncTimer
}

// NodesCoordinator -
func (ccm *ConsensusCoreMock) NodesCoordinator() sharding.NodesCoordinator {
	return ccm.validatorGroupSelector
}

// EpochStartRegistrationHandler -
func (ccm *ConsensusCoreMock) EpochStartRegistrationHandler() epochStart.RegistrationHandler {
	return ccm.epochStartNotifier
}

// SetBlockchain -
func (ccm *ConsensusCoreMock) SetBlockchain(blockChain data.ChainHandler) {
	ccm.blockChain = blockChain
}

// SetSingleSigner -
func (ccm *ConsensusCoreMock) SetSingleSigner(signer crypto.SingleSigner) {
	ccm.blsSingleSigner = signer
}

// SetBlockProcessor -
func (ccm *ConsensusCoreMock) SetBlockProcessor(blockProcessor process.BlockProcessor) {
	ccm.blockProcessor = blockProcessor
}

// SetBootStrapper -
func (ccm *ConsensusCoreMock) SetBootStrapper(bootstrapper process.Bootstrapper) {
	ccm.bootstrapper = bootstrapper
}

// SetBroadcastMessenger -
func (ccm *ConsensusCoreMock) SetBroadcastMessenger(broadcastMessenger consensus.BroadcastMessenger) {
	ccm.broadcastMessenger = broadcastMessenger
}

// SetChronology -
func (ccm *ConsensusCoreMock) SetChronology(chronologyHandler consensus.ChronologyHandler) {
	ccm.chronologyHandler = chronologyHandler
}

// SetHasher -
func (ccm *ConsensusCoreMock) SetHasher(hasher hashing.Hasher) {
	ccm.hasher = hasher
}

// SetMarshalizer -
func (ccm *ConsensusCoreMock) SetMarshalizer(marshalizer marshal.Marshalizer) {
	ccm.marshalizer = marshalizer
}

// SetMultiSigner -
func (ccm *ConsensusCoreMock) SetMultiSigner(multiSigner crypto.MultiSigner) {
	ccm.multiSigner = multiSigner
}

// SetRoundHandler -
func (ccm *ConsensusCoreMock) SetRoundHandler(roundHandler consensus.RoundHandler) {
	ccm.roundHandler = roundHandler
}

// SetShardCoordinator -
func (ccm *ConsensusCoreMock) SetShardCoordinator(shardCoordinator sharding.Coordinator) {
	ccm.shardCoordinator = shardCoordinator
}

// SetSyncTimer -
func (ccm *ConsensusCoreMock) SetSyncTimer(syncTimer ntp.SyncTimer) {
	ccm.syncTimer = syncTimer
}

// SetValidatorGroupSelector -
func (ccm *ConsensusCoreMock) SetValidatorGroupSelector(validatorGroupSelector sharding.NodesCoordinator) {
	ccm.validatorGroupSelector = validatorGroupSelector
}

// PrivateKey -
func (ccm *ConsensusCoreMock) PrivateKey() crypto.PrivateKey {
	return ccm.blsPrivateKey
}

// SingleSigner returns the bls single signer stored in the ConsensusStore
func (ccm *ConsensusCoreMock) SingleSigner() crypto.SingleSigner {
	return ccm.blsSingleSigner
}

// PeerHonestyHandler -
func (ccm *ConsensusCoreMock) PeerHonestyHandler() consensus.PeerHonestyHandler {
	return ccm.peerHonestyHandler
}

// HeaderSigVerifier -
func (ccm *ConsensusCoreMock) HeaderSigVerifier() consensus.HeaderSigVerifier {
	return ccm.headerSigVerifier
}

// SetHeaderSigVerifier -
func (ccm *ConsensusCoreMock) SetHeaderSigVerifier(headerSigVerifier consensus.HeaderSigVerifier) {
	ccm.headerSigVerifier = headerSigVerifier
}

// FallbackHeaderValidator -
func (ccm *ConsensusCoreMock) FallbackHeaderValidator() consensus.FallbackHeaderValidator {
	return ccm.fallbackHeaderValidator
}

// SetFallbackHeaderValidator -
func (ccm *ConsensusCoreMock) SetFallbackHeaderValidator(fallbackHeaderValidator consensus.FallbackHeaderValidator) {
	ccm.fallbackHeaderValidator = fallbackHeaderValidator
}

// NodeRedundancyHandler -
func (ccm *ConsensusCoreMock) NodeRedundancyHandler() consensus.NodeRedundancyHandler {
	return ccm.nodeRedundancyHandler
}

// SetNodeRedundancyHandler -
func (ccm *ConsensusCoreMock) SetNodeRedundancyHandler(nodeRedundancyHandler consensus.NodeRedundancyHandler) {
	ccm.nodeRedundancyHandler = nodeRedundancyHandler
}

// IsInterfaceNil returns true if there is no value under the interface
func (ccm *ConsensusCoreMock) IsInterfaceNil() bool {
	return ccm == nil
}
