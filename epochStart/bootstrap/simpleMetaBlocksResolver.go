package bootstrap

import (
	"errors"
	"fmt"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer"
	"github.com/ElrondNetwork/elrond-go/marshal"
	mock2 "github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	mock3 "github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// simpleMetaBlocksResolver initializes a HeaderResolver and sends requests from it
type simpleMetaBlocksResolver struct {
	messenger   p2p.Messenger
	marshalizer marshal.Marshalizer
	mbResolver  dataRetriever.HeaderResolver
}

// NewSimpleMetaBlocksResolver returns a new instance of simpleMetaBlocksResolver
func NewSimpleMetaBlocksResolver(
	messenger p2p.Messenger,
	marshalizer marshal.Marshalizer,
) (*simpleMetaBlocksResolver, error) {
	smbr := &simpleMetaBlocksResolver{
		messenger:   messenger,
		marshalizer: marshalizer,
	}
	err := smbr.init()
	if err != nil {
		return nil, err
	}

	return smbr, nil
}

func (smbr *simpleMetaBlocksResolver) init() error {
	storageService := &mock2.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock2.StorerMock{}
		},
	}
	cacher := mock3.NewPoolsHolderMock()
	dataPacker, err := partitioning.NewSimpleDataPacker(smbr.marshalizer)
	if err != nil {
		return err
	}
	triesHolder := state.NewDataTriesHolder()

	resolversContainerArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:         mock2.NewOneShardCoordinatorMock(),
		Messenger:                smbr.messenger,
		Store:                    storageService,
		Marshalizer:              smbr.marshalizer,
		DataPools:                cacher,
		Uint64ByteSliceConverter: uint64ByteSlice.NewBigEndianConverter(),
		DataPacker:               dataPacker,
		TriesContainer:           triesHolder,
		SizeCheckDelta:           0,
	}
	metaChainResolverContainer, err := resolverscontainer.NewMetaResolversContainerFactory(resolversContainerArgs)
	if err != nil {
		return err
	}

	numPeersToQuery := int(0.4 * float64(len(smbr.messenger.Peers())))
	if numPeersToQuery == 0 {
		numPeersToQuery = 2
	}
	resolver, err := metaChainResolverContainer.CreateMetaChainHeaderResolver(factory.MetachainBlocksTopic, numPeersToQuery, 0)
	if err != nil {
		return err
	}

	castedResolver, ok := resolver.(dataRetriever.HeaderResolver)
	if !ok {
		return errors.New("invalid resolver type")
	}
	smbr.mbResolver = castedResolver

	return nil
}

// RequestEpochStartMetaBlock will request the metablock to the peers
func (smbr *simpleMetaBlocksResolver) RequestEpochStartMetaBlock(epoch uint32) error {
	return smbr.mbResolver.RequestDataFromEpoch([]byte(fmt.Sprintf("epochStartBlock_%d", epoch)))
}
