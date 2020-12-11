package process

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/outport/drivers/websockets/types"
)

type validatorsPubKeysProcessor struct {
	validatorPubkeyConverter core.PubkeyConverter
}

func newValidatorsPubKeyProcessor(validatorPubkeyConverter core.PubkeyConverter) (*validatorsPubKeysProcessor, error) {
	return &validatorsPubKeysProcessor{
		validatorPubkeyConverter: validatorPubkeyConverter,
	}, nil
}

func (vpp *validatorsPubKeysProcessor) prepareValidatorPubKey(
	validatorsPubKeys map[uint32][][]byte,
	epoch uint32,
) []types.ShardValidatorsKeys {
	shardsValidatorsKeys := make([]types.ShardValidatorsKeys, 0)
	for shardID, shardValidatorsKeys := range validatorsPubKeys {
		shardValidators := types.ShardValidatorsKeys{
			Epoch:      epoch,
			ShardID:    shardID,
			PublicKeys: make([]string, 0),
		}

		for _, key := range shardValidatorsKeys {
			strValidatorPk := vpp.validatorPubkeyConverter.Encode(key)
			shardValidators.PublicKeys = append(shardValidators.PublicKeys, strValidatorPk)
		}

		shardsValidatorsKeys = append(shardsValidatorsKeys, shardValidators)
	}

	return shardsValidatorsKeys
}
