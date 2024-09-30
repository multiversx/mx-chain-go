package storagerequesterscontainer

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process/factory"
)

type sovereignShardRequestersContainerFactory struct {
	*shardRequestersContainerFactory
}

// Create returns a requester container that will hold all requesters in the system
func (srcf *sovereignShardRequestersContainerFactory) Create() (dataRetriever.RequestersContainer, error) {
	err := srcf.generateTxRequesters(
		factory.TransactionTopic,
		dataRetriever.TransactionUnit,
	)
	if err != nil {
		return nil, err
	}

	err = srcf.generateTxRequesters(
		factory.UnsignedTransactionTopic,
		dataRetriever.UnsignedTransactionUnit,
	)
	if err != nil {
		return nil, err
	}

	err = srcf.generateMiniBlocksRequesters()
	if err != nil {
		return nil, err
	}

	err = srcf.generatePeerAuthenticationRequester()
	if err != nil {
		return nil, err
	}

	validatorInfoTopicID := common.ValidatorInfoTopic + srcf.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	err = srcf.generateValidatorInfoRequester(validatorInfoTopicID)
	if err != nil {
		return nil, err
	}

	err = srcf.generateHeaderRequesters(core.SovereignChainShardId)
	if err != nil {
		return nil, err
	}

	return srcf.container, nil
}

func (srcf *sovereignShardRequestersContainerFactory) generateTxRequesters(
	topic string,
	unit dataRetriever.UnitType,
) error {
	keys := make([]string, 1)
	requestersSlice := make([]dataRetriever.Requester, 1)

	identifierTx := topic + srcf.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	requester, err := srcf.createTxRequester(identifierTx, unit)
	if err != nil {
		return err
	}

	requestersSlice[0] = requester
	keys[0] = identifierTx

	return srcf.container.AddMultiple(keys, requestersSlice)
}

func (srcf *sovereignShardRequestersContainerFactory) generateMiniBlocksRequesters() error {
	keys := make([]string, 1)
	requestersSlice := make([]dataRetriever.Requester, 1)

	identifierMiniBlocks := factory.MiniBlocksTopic + srcf.shardCoordinator.CommunicationIdentifier(core.SovereignChainShardId)
	requester, err := srcf.createMiniBlocksRequester(identifierMiniBlocks)
	if err != nil {
		return err
	}

	requestersSlice[0] = requester
	keys[0] = identifierMiniBlocks

	return srcf.container.AddMultiple(keys, requestersSlice)
}
