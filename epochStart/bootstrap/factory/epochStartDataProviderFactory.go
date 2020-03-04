package factory

//
//import (
//	"github.com/ElrondNetwork/elrond-go/core/check"
//	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap"
//	"github.com/ElrondNetwork/elrond-go/hashing"
//	"github.com/ElrondNetwork/elrond-go/marshal"
//	"github.com/ElrondNetwork/elrond-go/p2p"
//)
//
//type epochStartDataProviderFactory struct {
//	messenger           p2p.Messenger
//	marshalizer         marshal.Marshalizer
//	hasher              hashing.Hasher
//	nodesConfigProvider bootstrap.NodesConfigProviderHandler
//}
//
//// EpochStartDataProviderFactoryArgs holds the arguments needed for creating aa factory for the epoch start data
//// provider component
//type EpochStartDataProviderFactoryArgs struct {
//	Messenger           p2p.Messenger
//	Marshalizer         marshal.Marshalizer
//	Hasher              hashing.Hasher
//	NodesConfigProvider bootstrap.NodesConfigProviderHandler
//}
//
//// NewEpochStartDataProviderFactory returns a new instance of epochStartDataProviderFactory
//func NewEpochStartDataProviderFactory(args EpochStartDataProviderFactoryArgs) (*epochStartDataProviderFactory, error) {
//	if check.IfNil(args.Messenger) {
//		return nil, bootstrap.ErrNilMessenger
//	}
//	if check.IfNil(args.Marshalizer) {
//		return nil, bootstrap.ErrNilMarshalizer
//	}
//	if check.IfNil(args.Hasher) {
//		return nil, bootstrap.ErrNilHasher
//	}
//	if check.IfNil(args.NodesConfigProvider) {
//		return nil, bootstrap.ErrNilNodesConfigProvider
//	}
//
//	return &epochStartDataProviderFactory{
//		messenger:           args.Messenger,
//		marshalizer:         args.Marshalizer,
//		hasher:              args.Hasher,
//		nodesConfigProvider: args.NodesConfigProvider,
//	}, nil
//}
//
//// Create will init and return an instance of an epoch start data provider
//func (esdpf *epochStartDataProviderFactory) Create() (bootstrap.EpochStartDataProviderHandler, error) {
//	metaBlockInterceptor := bootstrap.NewSimpleMetaBlockInterceptor(esdpf.marshalizer, esdpf.hasher)
//	shardHdrInterceptor := bootstrap.NewSimpleShardHeaderInterceptor(esdpf.marshalizer)
//	metaBlockResolver, err := bootstrap.NewSimpleMetaBlocksResolver(esdpf.messenger, esdpf.marshalizer)
//	if err != nil {
//		return nil, err
//	}
//
//	epochStartDataProvider, err := bootstrap.NewEpochStartDataProvider(
//		esdpf.messenger,
//		esdpf.marshalizer,
//		esdpf.hasher,
//		esdpf.nodesConfigProvider,
//		metaBlockInterceptor,
//		shardHdrInterceptor,
//		metaBlockResolver,
//	)
//
//	return epochStartDataProvider, nil
//}
