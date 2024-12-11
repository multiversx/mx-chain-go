package storagerequesterscontainer

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	storagerequesters "github.com/multiversx/mx-chain-go/dataRetriever/storageRequesters"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/factory"
)

type sovereignShardRequestersContainerFactory struct {
	*shardRequestersContainerFactory
}

// NewSovereignShardRequestersContainerFactory creates a sovereign container filled with topic requesters
func NewSovereignShardRequestersContainerFactory(baseContainer *shardRequestersContainerFactory) (*sovereignShardRequestersContainerFactory, error) {
	if check.IfNil(baseContainer) {
		return nil, errorsMx.ErrNilShardRequesterContainerFactory
	}

	return &sovereignShardRequestersContainerFactory{
		shardRequestersContainerFactory: baseContainer,
	}, nil
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

	err = srcf.generateExtendedShardHeaderRequesters()
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

func (srcf *sovereignShardRequestersContainerFactory) generateExtendedShardHeaderRequesters() error {
	shardC := srcf.shardCoordinator
	identifierHdr := factory.ExtendedHeaderProofTopic + shardC.CommunicationIdentifier(shardC.SelfId())

	hdrStorer, err := srcf.store.GetStorer(dataRetriever.ExtendedShardHeadersUnit)
	if err != nil {
		return err
	}

	hdrNonceHashDataUnit := dataRetriever.ExtendedShardHeadersNonceHashDataUnit + dataRetriever.UnitType(shardC.SelfId())
	hdrNonceStore, err := srcf.store.GetStorer(hdrNonceHashDataUnit)
	if err != nil {
		return err
	}

	arg := storagerequesters.ArgHeaderRequester{
		Messenger:                srcf.messenger,
		ResponseTopicName:        identifierHdr,
		NonceConverter:           srcf.uint64ByteSliceConverter,
		HdrStorage:               hdrStorer,
		HeadersNoncesStorage:     hdrNonceStore,
		ManualEpochStartNotifier: srcf.manualEpochStartNotifier,
		ChanGracefullyClose:      srcf.chanGracefullyClose,
		DelayBeforeGracefulClose: defaultBeforeGracefulClose,
	}
	requester, err := storagerequesters.NewHeaderRequester(arg)
	if err != nil {
		return err
	}

	return srcf.container.Add(identifierHdr, requester)
}

// IsInterfaceNil returns true if there is no value under the interface
func (srcf *sovereignShardRequestersContainerFactory) IsInterfaceNil() bool {
	return srcf == nil
}
