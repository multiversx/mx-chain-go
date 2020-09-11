package workItems

type itemValidators struct {
	indexer           saveValidatorsIndexer
	epoch             uint32
	validatorsPubKeys map[uint32][][]byte
}

// NewItemValidators will create a new instance of itemValidators
func NewItemValidators(
	indexer saveValidatorsIndexer,
	epoch uint32,
	validatorsPubKeys map[uint32][][]byte,
) WorkItemHandler {
	return &itemValidators{
		indexer:           indexer,
		epoch:             epoch,
		validatorsPubKeys: validatorsPubKeys,
	}
}

// Save will save information about validators
func (wiv *itemValidators) Save() error {
	for shardID, shardPubKeys := range wiv.validatorsPubKeys {
		err := wiv.indexer.SaveShardValidatorsPubKeys(shardID, wiv.epoch, shardPubKeys)
		if err != nil {
			log.Warn("itemValidators.Save",
				"could not index validators public keys, ",
				"for shard", shardID,
				"error", err.Error())
			return err
		}
	}

	return nil
}
