package spos

func ValidateConsensusDataContainer(container *ConsensusDataContainer) error {
	if container == nil {
		return ErrNilConsensusDataContainer
	}

	if container.GetChainHandler() == nil {
		return ErrNilBlockChain
	}

	if container.GetBlockProcessor() == nil {
		return ErrNilBlockProcessor
	}

	if container.GetBootStrapper() == nil {
		return ErrNilBlootstraper
	}

	if container.GetChronology() == nil {
		return ErrNilChronologyHandler
	}

	if container.GetConsensusState() == nil {
		return ErrNilConsensusState
	}

	if container.GetHasher() == nil {
		return ErrNilHasher
	}

	if container.GetMarshalizer() == nil {
		return ErrNilMarshalizer
	}

	if container.GetMultiSigner() == nil {
		return ErrNilMultiSigner
	}

	if container.GetRounder() == nil {
		return ErrNilRounder
	}

	if container.GetShardCoordinator() == nil {
		return ErrNilShardCoordinator
	}

	if container.GetSyncTimer() == nil {
		return ErrNilSyncTimer
	}

	if container.GetValidatorGroupSelector() == nil {
		return ErrNilValidatorGroupSelector
	}

	return nil
}
