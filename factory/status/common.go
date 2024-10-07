package status

import (
	nodeData "github.com/multiversx/mx-chain-core-go/data"
	outportCore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

// CreateSaveValidatorsPubKeysEventHandler creates an epoch start action handler able to save validators pub keys
// in outport handler
func CreateSaveValidatorsPubKeysEventHandler(
	nodesCoordinator nodesCoordinator.NodesCoordinator,
	outportHandler outport.OutportHandler,
) epochStart.ActionHandler {
	subscribeHandler := notifier.NewHandlerForEpochStart(func(hdr nodeData.HeaderHandler) {
		currentEpoch := hdr.GetEpoch()
		validatorsPubKeys, err := nodesCoordinator.GetAllEligibleValidatorsPublicKeys(currentEpoch)
		if err != nil {
			log.Warn("pc.nodesCoordinator.GetAllEligibleValidatorPublicKeys for current epoch failed",
				"epoch", currentEpoch,
				"error", err.Error())
		}

		outportHandler.SaveValidatorsPubKeys(&outportCore.ValidatorsPubKeys{
			ShardID:                hdr.GetShardID(),
			ShardValidatorsPubKeys: outportCore.ConvertPubKeys(validatorsPubKeys),
			Epoch:                  currentEpoch,
		})

	}, func(_ nodeData.HeaderHandler) {}, common.IndexerOrder)

	return subscribeHandler
}
