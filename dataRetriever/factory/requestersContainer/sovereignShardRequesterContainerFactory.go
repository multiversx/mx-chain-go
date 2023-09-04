package requesterscontainer

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers/requesters"
	"github.com/multiversx/mx-chain-go/process/factory"
)

// TODO: Implement this in MX-14516

type sovereignShardRequestersContainerFactory struct {
	*shardRequestersContainerFactory
}

// NewSovereignShardRequestersContainerFactory creates a new container filled with topic requesters for sovereign shards
func NewSovereignShardRequestersContainerFactory(args FactoryArgs) (*sovereignShardRequestersContainerFactory, error) {
	shardRequester, err := NewShardRequestersContainerFactory(args)
	if err != nil {
		return nil, err
	}

	return &sovereignShardRequestersContainerFactory{
		shardRequestersContainerFactory: shardRequester,
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

	shardID := srcf.shardCoordinator.SelfId()
	identifierHdr := factory.ExtendedHeaderProofTopic + shardC.CommunicationIdentifier(shardID)
	requestSender, err := srcf.createOneRequestSenderWithSpecifiedNumRequests(identifierHdr, EmptyExcludePeersOnTopic, shardID, srcf.numCrossShardPeers, srcf.numIntraShardPeers)
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
