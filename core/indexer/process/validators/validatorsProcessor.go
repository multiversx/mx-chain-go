package validators

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
)

var log = logger.GetOrCreate("indexer/process/validators")

type validatorsProcessor struct {
	validatorPubkeyConverter core.PubkeyConverter
}

// NewValidatorsProcessor will create a new instance of validatorsProcessor
func NewValidatorsProcessor(validatorPubkeyConverter core.PubkeyConverter) *validatorsProcessor {
	return &validatorsProcessor{
		validatorPubkeyConverter: validatorPubkeyConverter,
	}
}

// PrepareValidatorsPublicKeys will prepare validators public keys
func (vp *validatorsProcessor) PrepareValidatorsPublicKeys(shardValidatorsPubKeys [][]byte) *types.ValidatorsPublicKeys {
	validatorsPubKeys := &types.ValidatorsPublicKeys{
		PublicKeys: make([]string, 0),
	}

	for _, validatorPk := range shardValidatorsPubKeys {
		strValidatorPk := vp.validatorPubkeyConverter.Encode(validatorPk)

		validatorsPubKeys.PublicKeys = append(validatorsPubKeys.PublicKeys, strValidatorPk)
	}

	return validatorsPubKeys
}
