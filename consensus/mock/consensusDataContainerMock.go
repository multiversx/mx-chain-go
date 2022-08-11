package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/consensus"
)

// ConsensusCoreMock -
type ConsensusCoreMock struct {
	blockChain              data.ChainHandler
	blockProcessor          consensus.BlockProcessor
	headersSubscriber       consensus.HeadersPoolSubscriber
	bootstrapper            consensus.Bootstrapper
	broadcastMessenger      consensus.BroadcastMessenger
	chronologyHandler       consensus.ChronologyHandler
	hasher                  hashing.Hasher
	marshalizer             marshal.Marshalizer
	blsPrivateKey           crypto.PrivateKey
	blsSingleSigner         crypto.SingleSigner
	multiSignerContainer    consensus.MultiSignerContainer
	roundHandler            consensus.RoundHandler
	shardCoordinator        consensus.ShardCoordinator
	syncTimer               consensus.SyncTimer
	validatorGroupSelector  consensus.NodesCoordinator
	epochStartNotifier      consensus.RegistrationHandler
	antifloodHandler        consensus.P2PAntifloodHandler
	peerHonestyHandler      consensus.PeerHonestyHandler
	headerSigVerifier       consensus.HeaderSigVerifier
	fallbackHeaderValidator consensus.FallbackHeaderValidator
	nodeRedundancyHandler   consensus.NodeRedundancyHandler
	scheduledProcessor      consensus.ScheduledProcessor
	signatureHandler        consensus.SignatureHandler
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
func (ccm *ConsensusCoreMock) BlockProcessor() consensus.BlockProcessor {
	return ccm.blockProcessor
}

// HeadersPoolSubscriber -
func (ccm *ConsensusCoreMock) HeadersPoolSubscriber() consensus.HeadersPoolSubscriber {
	return ccm.headersSubscriber
}

// BootStrapper -
func (ccm *ConsensusCoreMock) BootStrapper() consensus.Bootstrapper {
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

// MultiSignerContainer -
func (ccm *ConsensusCoreMock) MultiSignerContainer() consensus.MultiSignerContainer {
	return ccm.multiSignerContainer
}

// RoundHandler -
func (ccm *ConsensusCoreMock) RoundHandler() consensus.RoundHandler {
	return ccm.roundHandler
}

// ShardCoordinator -
func (ccm *ConsensusCoreMock) ShardCoordinator() consensus.ShardCoordinator {
	return ccm.shardCoordinator
}

// SyncTimer -
func (ccm *ConsensusCoreMock) SyncTimer() consensus.SyncTimer {
	return ccm.syncTimer
}

// NodesCoordinator -
func (ccm *ConsensusCoreMock) NodesCoordinator() consensus.NodesCoordinator {
	return ccm.validatorGroupSelector
}

// EpochStartRegistrationHandler -
func (ccm *ConsensusCoreMock) EpochStartRegistrationHandler() consensus.RegistrationHandler {
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
func (ccm *ConsensusCoreMock) SetBlockProcessor(blockProcessor consensus.BlockProcessor) {
	ccm.blockProcessor = blockProcessor
}

// SetBootStrapper -
func (ccm *ConsensusCoreMock) SetBootStrapper(bootstrapper consensus.Bootstrapper) {
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

// SetMultiSignerContainer -
func (ccm *ConsensusCoreMock) SetMultiSignerContainer(multiSignerContainer consensus.MultiSignerContainer) {
	ccm.multiSignerContainer = multiSignerContainer
}

// SetRoundHandler -
func (ccm *ConsensusCoreMock) SetRoundHandler(roundHandler consensus.RoundHandler) {
	ccm.roundHandler = roundHandler
}

// SetShardCoordinator -
func (ccm *ConsensusCoreMock) SetShardCoordinator(shardCoordinator consensus.ShardCoordinator) {
	ccm.shardCoordinator = shardCoordinator
}

// SetSyncTimer -
func (ccm *ConsensusCoreMock) SetSyncTimer(syncTimer consensus.SyncTimer) {
	ccm.syncTimer = syncTimer
}

// SetValidatorGroupSelector -
func (ccm *ConsensusCoreMock) SetValidatorGroupSelector(validatorGroupSelector consensus.NodesCoordinator) {
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

// ScheduledProcessor -
func (ccm *ConsensusCoreMock) ScheduledProcessor() consensus.ScheduledProcessor {
	return ccm.scheduledProcessor
}

// SetNodeRedundancyHandler -
func (ccm *ConsensusCoreMock) SetNodeRedundancyHandler(nodeRedundancyHandler consensus.NodeRedundancyHandler) {
	ccm.nodeRedundancyHandler = nodeRedundancyHandler
}

// SignatureHandler -
func (ccm *ConsensusCoreMock) SignatureHandler() consensus.SignatureHandler {
	return ccm.signatureHandler
}

// SetSignatureHandler -
func (ccm *ConsensusCoreMock) SetSignatureHandler(signatureHandler consensus.SignatureHandler) {
	ccm.signatureHandler = signatureHandler
}

// IsInterfaceNil returns true if there is no value under the interface
func (ccm *ConsensusCoreMock) IsInterfaceNil() bool {
	return ccm == nil
}
