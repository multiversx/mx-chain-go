package spos

import "github.com/multiversx/mx-chain-core-go/core/check"

// ValidateConsensusCore checks for nil all the container objects
func ValidateConsensusCore(container ConsensusCoreHandler) error {
	if check.IfNil(container) {
		return ErrNilConsensusCore
	}
	if check.IfNil(container.Blockchain()) {
		return ErrNilBlockChain
	}
	if check.IfNil(container.BlockProcessor()) {
		return ErrNilBlockProcessor
	}
	if check.IfNil(container.BootStrapper()) {
		return ErrNilBootstrapper
	}
	if check.IfNil(container.BroadcastMessenger()) {
		return ErrNilBroadcastMessenger
	}
	if check.IfNil(container.Chronology()) {
		return ErrNilChronologyHandler
	}
	if check.IfNil(container.Hasher()) {
		return ErrNilHasher
	}
	if check.IfNil(container.Marshalizer()) {
		return ErrNilMarshalizer
	}
	if check.IfNil(container.MultiSignerContainer()) {
		return ErrNilMultiSignerContainer
	}
	multiSigner, _ := container.MultiSignerContainer().GetMultiSigner(0)
	if check.IfNil(multiSigner) {
		return ErrNilMultiSigner
	}
	if check.IfNil(container.RoundHandler()) {
		return ErrNilRoundHandler
	}
	if check.IfNil(container.ShardCoordinator()) {
		return ErrNilShardCoordinator
	}
	if check.IfNil(container.SyncTimer()) {
		return ErrNilSyncTimer
	}
	if check.IfNil(container.NodesCoordinator()) {
		return ErrNilNodesCoordinator
	}
	if check.IfNil(container.GetAntiFloodHandler()) {
		return ErrNilAntifloodHandler
	}
	if check.IfNil(container.PeerHonestyHandler()) {
		return ErrNilPeerHonestyHandler
	}
	if check.IfNil(container.HeaderSigVerifier()) {
		return ErrNilHeaderSigVerifier
	}
	if check.IfNil(container.FallbackHeaderValidator()) {
		return ErrNilFallbackHeaderValidator
	}
	if check.IfNil(container.NodeRedundancyHandler()) {
		return ErrNilNodeRedundancyHandler
	}
	if check.IfNil(container.ScheduledProcessor()) {
		return ErrNilScheduledProcessor
	}
	if check.IfNil(container.MessageSigningHandler()) {
		return ErrNilMessageSigningHandler
	}
	if check.IfNil(container.PeerBlacklistHandler()) {
		return ErrNilPeerBlacklistHandler
	}
	if check.IfNil(container.SigningHandler()) {
		return ErrNilSigningHandler
	}

	return nil
}
