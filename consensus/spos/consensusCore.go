package spos

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/consensus"
)

// ConsensusCore implements ConsensusCoreHandler and provides access to common functionality
//  for the rest of the consensus structures
type ConsensusCore struct {
	blockChain                    data.ChainHandler
	blockProcessor                consensus.BlockProcessor
	bootstrapper                  consensus.Bootstrapper
	broadcastMessenger            consensus.BroadcastMessenger
	chronologyHandler             consensus.ChronologyHandler
	hasher                        hashing.Hasher
	marshalizer                   marshal.Marshalizer
	blsPrivateKey                 crypto.PrivateKey
	blsSingleSigner               crypto.SingleSigner
	multiSignerContainer          consensus.MultiSignerContainer
	roundHandler                  consensus.RoundHandler
	shardCoordinator              consensus.ShardCoordinator
	nodesCoordinator              consensus.NodesCoordinator
	syncTimer                     consensus.SyncTimer
	epochStartRegistrationHandler consensus.RegistrationHandler
	antifloodHandler              consensus.P2PAntifloodHandler
	peerHonestyHandler            consensus.PeerHonestyHandler
	headerSigVerifier             consensus.HeaderSigVerifier
	fallbackHeaderValidator       consensus.FallbackHeaderValidator
	nodeRedundancyHandler         consensus.NodeRedundancyHandler
	scheduledProcessor            consensus.ScheduledProcessor
	signatureHandler              consensus.SignatureHandler
}

// ConsensusCoreArgs store all arguments that are needed to create a ConsensusCore object
type ConsensusCoreArgs struct {
	BlockChain                    data.ChainHandler
	BlockProcessor                consensus.BlockProcessor
	Bootstrapper                  consensus.Bootstrapper
	BroadcastMessenger            consensus.BroadcastMessenger
	ChronologyHandler             consensus.ChronologyHandler
	Hasher                        hashing.Hasher
	Marshalizer                   marshal.Marshalizer
	BlsPrivateKey                 crypto.PrivateKey
	BlsSingleSigner               crypto.SingleSigner
	MultiSignerContainer          consensus.MultiSignerContainer
	RoundHandler                  consensus.RoundHandler
	ShardCoordinator              consensus.ShardCoordinator
	NodesCoordinator              consensus.NodesCoordinator
	SyncTimer                     consensus.SyncTimer
	EpochStartRegistrationHandler consensus.RegistrationHandler
	AntifloodHandler              consensus.P2PAntifloodHandler
	PeerHonestyHandler            consensus.PeerHonestyHandler
	HeaderSigVerifier             consensus.HeaderSigVerifier
	FallbackHeaderValidator       consensus.FallbackHeaderValidator
	NodeRedundancyHandler         consensus.NodeRedundancyHandler
	ScheduledProcessor            consensus.ScheduledProcessor
	SignatureHandler              consensus.SignatureHandler
}

// NewConsensusCore creates a new ConsensusCore instance
func NewConsensusCore(
	args *ConsensusCoreArgs,
) (*ConsensusCore, error) {
	consensusCore := &ConsensusCore{
		blockChain:                    args.BlockChain,
		blockProcessor:                args.BlockProcessor,
		bootstrapper:                  args.Bootstrapper,
		broadcastMessenger:            args.BroadcastMessenger,
		chronologyHandler:             args.ChronologyHandler,
		hasher:                        args.Hasher,
		marshalizer:                   args.Marshalizer,
		blsPrivateKey:                 args.BlsPrivateKey,
		blsSingleSigner:               args.BlsSingleSigner,
		multiSignerContainer:          args.MultiSignerContainer,
		roundHandler:                  args.RoundHandler,
		shardCoordinator:              args.ShardCoordinator,
		nodesCoordinator:              args.NodesCoordinator,
		syncTimer:                     args.SyncTimer,
		epochStartRegistrationHandler: args.EpochStartRegistrationHandler,
		antifloodHandler:              args.AntifloodHandler,
		peerHonestyHandler:            args.PeerHonestyHandler,
		headerSigVerifier:             args.HeaderSigVerifier,
		fallbackHeaderValidator:       args.FallbackHeaderValidator,
		nodeRedundancyHandler:         args.NodeRedundancyHandler,
		scheduledProcessor:            args.ScheduledProcessor,
		signatureHandler:              args.SignatureHandler,
	}

	err := ValidateConsensusCore(consensusCore)
	if err != nil {
		return nil, err
	}

	return consensusCore, nil
}

// Blockchain gets the ChainHandler stored in the ConsensusCore
func (cc *ConsensusCore) Blockchain() data.ChainHandler {
	return cc.blockChain
}

// GetAntiFloodHandler will return the antiflood handler which will be used in subrounds
func (cc *ConsensusCore) GetAntiFloodHandler() consensus.P2PAntifloodHandler {
	return cc.antifloodHandler
}

// BlockProcessor gets the BlockProcessor stored in the ConsensusCore
func (cc *ConsensusCore) BlockProcessor() consensus.BlockProcessor {
	return cc.blockProcessor
}

// BootStrapper gets the Bootstrapper stored in the ConsensusCore
func (cc *ConsensusCore) BootStrapper() consensus.Bootstrapper {
	return cc.bootstrapper
}

// BroadcastMessenger gets the BroadcastMessenger stored in the ConsensusCore
func (cc *ConsensusCore) BroadcastMessenger() consensus.BroadcastMessenger {
	return cc.broadcastMessenger
}

// Chronology gets the ChronologyHandler stored in the ConsensusCore
func (cc *ConsensusCore) Chronology() consensus.ChronologyHandler {
	return cc.chronologyHandler
}

// Hasher gets the Hasher stored in the ConsensusCore
func (cc *ConsensusCore) Hasher() hashing.Hasher {
	return cc.hasher
}

// Marshalizer gets the Marshalizer stored in the ConsensusCore
func (cc *ConsensusCore) Marshalizer() marshal.Marshalizer {
	return cc.marshalizer
}

// MultiSignerContainer gets the MultiSignerContainer stored in the ConsensusCore
func (cc *ConsensusCore) MultiSignerContainer() consensus.MultiSignerContainer {
	return cc.multiSignerContainer
}

//RoundHandler gets the RoundHandler stored in the ConsensusCore
func (cc *ConsensusCore) RoundHandler() consensus.RoundHandler {
	return cc.roundHandler
}

// ShardCoordinator gets the ShardCoordinator stored in the ConsensusCore
func (cc *ConsensusCore) ShardCoordinator() consensus.ShardCoordinator {
	return cc.shardCoordinator
}

//SyncTimer gets the SyncTimer stored in the ConsensusCore
func (cc *ConsensusCore) SyncTimer() consensus.SyncTimer {
	return cc.syncTimer
}

// NodesCoordinator gets the NodesCoordinator stored in the ConsensusCore
func (cc *ConsensusCore) NodesCoordinator() consensus.NodesCoordinator {
	return cc.nodesCoordinator
}

// EpochStartRegistrationHandler returns the epoch start registration handler
func (cc *ConsensusCore) EpochStartRegistrationHandler() consensus.RegistrationHandler {
	return cc.epochStartRegistrationHandler
}

// PrivateKey returns the BLS private key stored in the ConsensusStore
func (cc *ConsensusCore) PrivateKey() crypto.PrivateKey {
	return cc.blsPrivateKey
}

// SingleSigner returns the bls single signer stored in the ConsensusStore
func (cc *ConsensusCore) SingleSigner() crypto.SingleSigner {
	return cc.blsSingleSigner
}

// PeerHonestyHandler will return the peer honesty handler which will be used in subrounds
func (cc *ConsensusCore) PeerHonestyHandler() consensus.PeerHonestyHandler {
	return cc.peerHonestyHandler
}

// HeaderSigVerifier returns the sig verifier handler which will be used in subrounds
func (cc *ConsensusCore) HeaderSigVerifier() consensus.HeaderSigVerifier {
	return cc.headerSigVerifier
}

// FallbackHeaderValidator will return the fallback header validator which will be used in subrounds
func (cc *ConsensusCore) FallbackHeaderValidator() consensus.FallbackHeaderValidator {
	return cc.fallbackHeaderValidator
}

// NodeRedundancyHandler will return the node redundancy handler which will be used in subrounds
func (cc *ConsensusCore) NodeRedundancyHandler() consensus.NodeRedundancyHandler {
	return cc.nodeRedundancyHandler
}

// ScheduledProcessor will return the scheduled processor
func (cc *ConsensusCore) ScheduledProcessor() consensus.ScheduledProcessor {
	return cc.scheduledProcessor
}

// SignatureHandler will return the signature handler component
func (cc *ConsensusCore) SignatureHandler() consensus.SignatureHandler {
	return cc.signatureHandler
}

// IsInterfaceNil returns true if there is no value under the interface
func (cc *ConsensusCore) IsInterfaceNil() bool {
	return cc == nil
}
