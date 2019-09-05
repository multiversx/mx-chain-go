package spos

// ValidateConsensusCore checks for nil all the container objects
func ValidateConsensusCore(container ConsensusCoreHandler) error {
	if container == nil || container.IsInterfaceNil() {
		return ErrNilConsensusCore
	}
	if container.Blockchain() == nil || container.Blockchain().IsInterfaceNil() {
		return ErrNilBlockChain
	}
	if container.BlockProcessor() == nil || container.BlockProcessor().IsInterfaceNil() {
		return ErrNilBlockProcessor
	}
	if container.BlocksTracker() == nil || container.BlocksTracker().IsInterfaceNil() {
		return ErrNilBlocksTracker
	}
	if container.PeerProcessor() == nil || container.PeerProcessor().IsInterfaceNil() {
		return ErrNilPeerProcessor
	}
	if container.BootStrapper() == nil || container.BootStrapper().IsInterfaceNil() {
		return ErrNilBootstrapper
	}
	if container.BroadcastMessenger() == nil || container.BroadcastMessenger().IsInterfaceNil() {
		return ErrNilBroadcastMessenger
	}
	if container.Chronology() == nil || container.Chronology().IsInterfaceNil() {
		return ErrNilChronologyHandler
	}
	if container.Hasher() == nil || container.Hasher().IsInterfaceNil() {
		return ErrNilHasher
	}
	if container.Marshalizer() == nil || container.Marshalizer().IsInterfaceNil() {
		return ErrNilMarshalizer
	}
	if container.MultiSigner() == nil || container.MultiSigner().IsInterfaceNil() {
		return ErrNilMultiSigner
	}
	if container.Rounder() == nil || container.Rounder().IsInterfaceNil() {
		return ErrNilRounder
	}
	if container.ShardCoordinator() == nil || container.ShardCoordinator().IsInterfaceNil() {
		return ErrNilShardCoordinator
	}
	if container.SyncTimer() == nil || container.SyncTimer().IsInterfaceNil() {
		return ErrNilSyncTimer
	}
	if container.NodesCoordinator() == nil || container.NodesCoordinator().IsInterfaceNil() {
		return ErrNilValidatorGroupSelector
	}
	if container.RandomnessPrivateKey() == nil || container.RandomnessPrivateKey().IsInterfaceNil() {
		return ErrNilBlsPrivateKey
	}
	if container.RandomnessSingleSigner() == nil || container.RandomnessSingleSigner().IsInterfaceNil() {
		return ErrNilBlsSingleSigner
	}

	return nil
}
