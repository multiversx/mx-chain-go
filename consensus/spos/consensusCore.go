package spos

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

// ConsensusCore implements ConsensusCoreHandler and provides access to common functionality
//  for the rest of the consensus structures
type ConsensusCore struct {
	blockChain                    data.ChainHandler
	blockProcessor                process.BlockProcessor
	bootstrapper                  process.Bootstrapper
	broadcastMessenger            consensus.BroadcastMessenger
	chronologyHandler             consensus.ChronologyHandler
	hasher                        hashing.Hasher
	marshalizer                   marshal.Marshalizer
	blsPrivateKey                 crypto.PrivateKey
	blsSingleSigner               crypto.SingleSigner
	multiSigner                   crypto.MultiSigner
	rounder                       consensus.Rounder
	shardCoordinator              sharding.Coordinator
	nodesCoordinator              sharding.NodesCoordinator
	syncTimer                     ntp.SyncTimer
	epochStartRegistrationHandler epochStart.RegistrationHandler
	antifloodHandler              consensus.P2PAntifloodHandler
	peerHonestyHandler            consensus.PeerHonestyHandler
	headerSigVerifier             consensus.HeaderSigVerifier
	fallbackHeaderValidator       consensus.FallbackHeaderValidator
}

// ConsensusCoreArgs store all arguments that are needed to create a ConsensusCore object
type ConsensusCoreArgs struct {
	BlockChain                    data.ChainHandler
	BlockProcessor                process.BlockProcessor
	Bootstrapper                  process.Bootstrapper
	BroadcastMessenger            consensus.BroadcastMessenger
	ChronologyHandler             consensus.ChronologyHandler
	Hasher                        hashing.Hasher
	Marshalizer                   marshal.Marshalizer
	BlsPrivateKey                 crypto.PrivateKey
	BlsSingleSigner               crypto.SingleSigner
	MultiSigner                   crypto.MultiSigner
	Rounder                       consensus.Rounder
	ShardCoordinator              sharding.Coordinator
	NodesCoordinator              sharding.NodesCoordinator
	SyncTimer                     ntp.SyncTimer
	EpochStartRegistrationHandler epochStart.RegistrationHandler
	AntifloodHandler              consensus.P2PAntifloodHandler
	PeerHonestyHandler            consensus.PeerHonestyHandler
	HeaderSigVerifier             consensus.HeaderSigVerifier
	FallbackHeaderValidator       consensus.FallbackHeaderValidator
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
		multiSigner:                   args.MultiSigner,
		rounder:                       args.Rounder,
		shardCoordinator:              args.ShardCoordinator,
		nodesCoordinator:              args.NodesCoordinator,
		syncTimer:                     args.SyncTimer,
		epochStartRegistrationHandler: args.EpochStartRegistrationHandler,
		antifloodHandler:              args.AntifloodHandler,
		peerHonestyHandler:            args.PeerHonestyHandler,
		headerSigVerifier:             args.HeaderSigVerifier,
		fallbackHeaderValidator:       args.FallbackHeaderValidator,
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
func (cc *ConsensusCore) BlockProcessor() process.BlockProcessor {
	return cc.blockProcessor
}

// BootStrapper gets the Bootstrapper stored in the ConsensusCore
func (cc *ConsensusCore) BootStrapper() process.Bootstrapper {
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

// MultiSigner gets the MultiSigner stored in the ConsensusCore
func (cc *ConsensusCore) MultiSigner() crypto.MultiSigner {
	return cc.multiSigner
}

//Rounder gets the Rounder stored in the ConsensusCore
func (cc *ConsensusCore) Rounder() consensus.Rounder {
	return cc.rounder
}

// ShardCoordinator gets the Coordinator stored in the ConsensusCore
func (cc *ConsensusCore) ShardCoordinator() sharding.Coordinator {
	return cc.shardCoordinator
}

//SyncTimer gets the SyncTimer stored in the ConsensusCore
func (cc *ConsensusCore) SyncTimer() ntp.SyncTimer {
	return cc.syncTimer
}

// NodesCoordinator gets the NodesCoordinator stored in the ConsensusCore
func (cc *ConsensusCore) NodesCoordinator() sharding.NodesCoordinator {
	return cc.nodesCoordinator
}

// EpochStartRegistrationHandler returns the epoch start registration handler
func (cc *ConsensusCore) EpochStartRegistrationHandler() epochStart.RegistrationHandler {
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

// IsInterfaceNil returns true if there is no value under the interface
func (cc *ConsensusCore) IsInterfaceNil() bool {
	return cc == nil
}
