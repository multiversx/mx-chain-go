package spos

import "github.com/ElrondNetwork/elrond-go/core/check"

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
	if check.IfNil(container.MultiSigner()) {
		return ErrNilMultiSigner
	}
	if check.IfNil(container.Rounder()) {
		return ErrNilRounder
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
	if check.IfNil(container.PrivateKey()) {
		return ErrNilBlsPrivateKey
	}
	if check.IfNil(container.SingleSigner()) {
		return ErrNilBlsSingleSigner
	}
	if check.IfNil(container.GetAntiFloodHandler()) {
		return ErrNilAntifloodHandler
	}
	if check.IfNil(container.PeerHonestyHandler()) {
		return ErrNilPeerHonestyHandler
	}
	if check.IfNil(container.FallbackHeaderValidator()) {
		return ErrNilFallbackHeaderValidator
	}

	return nil
}
