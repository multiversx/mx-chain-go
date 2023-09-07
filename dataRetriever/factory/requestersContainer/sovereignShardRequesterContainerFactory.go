package requesterscontainer

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers/requesters"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/factory"
)

type sovereignShardRequestersContainerFactory struct {
	*shardRequestersContainerFactory
}

// NewSovereignShardRequestersContainerFactory creates a new container filled with topic requesters for sovereign shards
func NewSovereignShardRequestersContainerFactory(shardReqContainerFactory *shardRequestersContainerFactory) (*sovereignShardRequestersContainerFactory, error) {
	if check.IfNil(shardReqContainerFactory) {
		return nil, errors.ErrNilShardRequesterContainerFactory
	}

	return &sovereignShardRequestersContainerFactory{
		shardRequestersContainerFactory: shardReqContainerFactory,
	}, nil
}

// Create returns a requesters container that will hold all sovereign requesters
func (srcf *sovereignShardRequestersContainerFactory) Create() (dataRetriever.RequestersContainer, error) {
	_, err := srcf.shardRequestersContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	err = srcf.generateExtendedShardHeaderRequesters()
	if err != nil {
		return nil, err
	}

	return srcf.container, nil
}

func (srcf *sovereignShardRequestersContainerFactory) generateExtendedShardHeaderRequesters() error {
	shardC := srcf.shardCoordinator
	shardID := shardC.SelfId()

	identifierHdr := factory.ExtendedHeaderProofTopic + shardC.CommunicationIdentifier(shardID)
	requestSender, err := srcf.createOneRequestSenderWithSpecifiedNumRequests(identifierHdr, EmptyExcludePeersOnTopic, shardID, 0, srcf.numIntraShardPeers)
	if err != nil {
		return err
	}

	arg := requesters.ArgHeaderRequester{
		ArgBaseRequester: requesters.ArgBaseRequester{
			RequestSender: requestSender,
			Marshaller:    srcf.marshaller,
		},
		NonceConverter: srcf.uint64ByteSliceConverter,
	}
	requester, err := requesters.NewHeaderRequester(arg)
	if err != nil {
		return err
	}

	return srcf.container.Add(identifierHdr, requester)
}

// IsInterfaceNil returns true if there is no value under the interface
func (srcf *sovereignShardRequestersContainerFactory) IsInterfaceNil() bool {
	return srcf == nil
}
