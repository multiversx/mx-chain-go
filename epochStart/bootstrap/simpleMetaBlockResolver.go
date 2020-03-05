package bootstrap

import (
	"errors"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const percentageOfPeersToSendRequests = 0.4
const defaultNumOfPeersToSendRequests = 2

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
	if check.IfNil(messenger) {
		return nil, ErrNilMessenger
	}
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}

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
	storageService := &disabled.ChainStorer{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return disabled.NewDisabledStorer()
		},
	}
	cacher := disabled.NewDisabledPoolsHolder()
	dataPacker, err := partitioning.NewSimpleDataPacker(smbr.marshalizer)
	if err != nil {
		return err
	}
	triesHolder := state.NewDataTriesHolder()
	shardCoordinator, err := sharding.NewMultiShardCoordinator(3, core.MetachainShardId)
	if err != nil {
		return err
	}

	resolversContainerArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:         shardCoordinator,
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

	numPeersToQuery := int(percentageOfPeersToSendRequests * float64(len(smbr.messenger.Peers())))
	if numPeersToQuery == 0 {
		numPeersToQuery = defaultNumOfPeersToSendRequests
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

// IsInterfaceNil returns true if there is no value under the interface
func (smbr *simpleMetaBlocksResolver) IsInterfaceNil() bool {
	return smbr == nil
}
