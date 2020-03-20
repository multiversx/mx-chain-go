package factory

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

type epochStartDataProviderFactory struct {
	pubKey                  crypto.PublicKey
	messenger               p2p.Messenger
	marshalizer             marshal.Marshalizer
	hasher                  hashing.Hasher
	pathManager             storage.PathManagerHandler
	nodesConfigProvider     bootstrap.NodesConfigProviderHandler
	generalConfig           config.Config
	economicsConfig         config.EconomicsConfig
	defaultShardCoordinator sharding.Coordinator
	singleSigner            crypto.SingleSigner
	blockSingleSigner       crypto.SingleSigner
	keyGen                  crypto.KeyGenerator
	blockKeyGen             crypto.KeyGenerator
	shouldSync          bool
	workingDir          string
	defaultDBPath       string
	defaultEpochString  string
}

// EpochStartDataProviderFactoryArgs holds the arguments needed for creating aa factory for the epoch start data
// provider component
type EpochStartDataProviderFactoryArgs struct {
	PubKey                  crypto.PublicKey
	Messenger               p2p.Messenger
	Marshalizer             marshal.Marshalizer
	Hasher                  hashing.Hasher
	NodesConfigProvider     bootstrap.NodesConfigProviderHandler
	PathManager             storage.PathManagerHandler
	DefaultShardCoordinator sharding.Coordinator
	StartTime               time.Time
	OriginalNodesConfig     *sharding.NodesSetup
	GeneralConfig           *config.Config
	EconomicsConfig         *config.EconomicsConfig
	SingleSigner            crypto.SingleSigner
	BlockSingleSigner       crypto.SingleSigner
	KeyGen                  crypto.KeyGenerator
	BlockKeyGen             crypto.KeyGenerator
	IsEpochFoundInStorage   bool
	WorkingDir            string
	DefaultDBPath         string
	DefaultEpochString    string
}

// NewEpochStartDataProviderFactory returns a new instance of epochStartDataProviderFactory
func NewEpochStartDataProviderFactory(args EpochStartDataProviderFactoryArgs) (*epochStartDataProviderFactory, error) {
	if check.IfNil(args.PubKey) {
		return nil, bootstrap.ErrNilPublicKey
	}
	if check.IfNil(args.Messenger) {
		return nil, bootstrap.ErrNilMessenger
	}
	if check.IfNil(args.Marshalizer) {
		return nil, bootstrap.ErrNilMarshalizer
	}
	if check.IfNil(args.PathManager) {
		return nil, bootstrap.ErrNilPathManager
	}
	if check.IfNil(args.Hasher) {
		return nil, bootstrap.ErrNilHasher
	}
	if check.IfNil(args.NodesConfigProvider) {
		return nil, bootstrap.ErrNilNodesConfigProvider
	}
	if check.IfNil(args.DefaultShardCoordinator) {
		return nil, bootstrap.ErrNilDefaultShardCoordinator
	}
	if check.IfNil(args.BlockKeyGen) {
		return nil, bootstrap.ErrNilBlockKeyGen
	}
	if check.IfNil(args.KeyGen) {
		return nil, bootstrap.ErrNilKeyGen
	}
	if check.IfNil(args.SingleSigner) {
		return nil, bootstrap.ErrNilSingleSigner
	}
	if check.IfNil(args.BlockSingleSigner) {
		return nil, bootstrap.ErrNilBlockSingleSigner
	}

	shouldSync := bootstrap.ShouldSyncWithTheNetwork(
		args.StartTime,
		args.IsEpochFoundInStorage,
		args.OriginalNodesConfig,
		args.GeneralConfig,
	)
	shouldSync = true // hardcoded so we can test we can sync

	return &epochStartDataProviderFactory{
		pubKey:                  args.PubKey,
		messenger:               args.Messenger,
		marshalizer:             args.Marshalizer,
		hasher:                  args.Hasher,
		pathManager:             args.PathManager,
		generalConfig:           *args.GeneralConfig,
		economicsConfig:         *args.EconomicsConfig,
		nodesConfigProvider:     args.NodesConfigProvider,
		defaultShardCoordinator: args.DefaultShardCoordinator,
		keyGen:                  args.KeyGen,
		blockKeyGen:             args.BlockKeyGen,
		singleSigner:            args.SingleSigner,
		blockSingleSigner:       args.BlockSingleSigner,
		shouldSync:          shouldSync,
		workingDir:          args.WorkingDir,
		defaultEpochString:  args.DefaultEpochString,
		defaultDBPath:       args.DefaultEpochString,
	}, nil
}

// Create will init and return an instance of an epoch start data provider
func (esdpf *epochStartDataProviderFactory) Create() (bootstrap.EpochStartDataProviderHandler, error) {
	if !esdpf.shouldSync {
		return &disabledEpochStartDataProvider{}, nil
	}

	epochStartMetaBlockInterceptor, err := bootstrap.NewSimpleEpochStartMetaBlockInterceptor(esdpf.marshalizer, esdpf.hasher)
	if err != nil {
		return nil, err
	}
	metaBlockInterceptor, err := bootstrap.NewSimpleMetaBlockInterceptor(esdpf.marshalizer, esdpf.hasher)
	if err != nil {
		return nil, err
	}
	shardHdrInterceptor, err := bootstrap.NewSimpleShardHeaderInterceptor(esdpf.marshalizer, esdpf.hasher)
	if err != nil {
		return nil, err
	}
	miniBlockInterceptor, err := bootstrap.NewSimpleMiniBlockInterceptor(esdpf.marshalizer, esdpf.hasher)
	if err != nil {
		return nil, err
	}

	whiteListCache, err := storageUnit.NewCache(
		storageUnit.CacheType(esdpf.generalConfig.WhiteListPool.Type),
		esdpf.generalConfig.WhiteListPool.Size,
		esdpf.generalConfig.WhiteListPool.Shards,
	)
	if err != nil {
		return nil, err
	}
	whiteListHandler, err := interceptors.NewWhiteListDataVerifier(whiteListCache)
	if err != nil {
		return nil, err
	}

	argsEpochStart := bootstrap.ArgsEpochStartDataProvider{
		PublicKey:                      esdpf.pubKey,
		Messenger:                      esdpf.messenger,
		Marshalizer:                    esdpf.marshalizer,
		Hasher:                         esdpf.hasher,
		NodesConfigProvider:            esdpf.nodesConfigProvider,
		GeneralConfig:                  esdpf.generalConfig,
		EconomicsConfig:                esdpf.economicsConfig,
		PathManager:                    esdpf.pathManager,
		SingleSigner:                   esdpf.singleSigner,
		BlockSingleSigner:              esdpf.blockSingleSigner,
		KeyGen:                         esdpf.keyGen,
		BlockKeyGen:                    esdpf.blockKeyGen,
		DefaultShardCoordinator:        esdpf.defaultShardCoordinator,
		EpochStartMetaBlockInterceptor: epochStartMetaBlockInterceptor,
		MetaBlockInterceptor:           metaBlockInterceptor,
		ShardHeaderInterceptor:         shardHdrInterceptor,
		MiniBlockInterceptor:           miniBlockInterceptor,
		WhiteListHandler:               whiteListHandler,
		WorkingDir:                     esdpf.workingDir,
		DefaultEpochString:             esdpf.defaultEpochString,
		DefaultDBPath:                  esdpf.defaultDBPath,
	}
	return bootstrap.NewEpochStartDataProvider(argsEpochStart)
}
