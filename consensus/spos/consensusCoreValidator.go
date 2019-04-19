package spos

func ValidateConsensusCore(container ConsensusCoreHandler) error {
	if container == nil {
		return ErrNilConsensusCore
	}

	if container.Blockchain() == nil {
		return ErrNilBlockChain
	}

	if container.BlockProcessor() == nil {
		return ErrNilBlockProcessor
	}

	if container.BootStrapper() == nil {
		return ErrNilBlootstraper
	}

	if container.Chronology() == nil {
		return ErrNilChronologyHandler
	}

	if container.Hasher() == nil {
		return ErrNilHasher
	}

	if container.Marshalizer() == nil {
		return ErrNilMarshalizer
	}

	if container.MultiSigner() == nil {
		return ErrNilMultiSigner
	}

	if container.Rounder() == nil {
		return ErrNilRounder
	}

	if container.ShardCoordinator() == nil {
		return ErrNilShardCoordinator
	}

	if container.SyncTimer() == nil {
		return ErrNilSyncTimer
	}

	if container.ValidatorGroupSelector() == nil {
		return ErrNilValidatorGroupSelector
	}

	return nil
}
