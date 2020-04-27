package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/hashing"
	factoryHasher "github.com/ElrondNetwork/elrond-go/hashing/factory"
	"github.com/ElrondNetwork/elrond-go/marshal"
	factoryMarshalizer "github.com/ElrondNetwork/elrond-go/marshal/factory"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// CoreComponentsFactoryArgs holds the arguments needed for creating a core components factory
type CoreComponentsFactoryArgs struct {
	Config      *config.Config
	PathManager storage.PathManagerHandler
	ShardId     string
	ChainID     []byte
}

// CoreComponentsFactory is responsible for creating the core components
type CoreComponentsFactory struct {
	config      *config.Config
	pathManager storage.PathManagerHandler
	shardId     string
	chainID     []byte
	marshalizer marshal.Marshalizer
	hasher      hashing.Hasher
}

// CoreComponents is the DTO used for core components
type CoreComponents struct {
	Hasher                   hashing.Hasher
	InternalMarshalizer      marshal.Marshalizer
	VmMarshalizer            marshal.Marshalizer
	TxSignMarshalizer        marshal.Marshalizer
	TriesContainer           state.TriesHolder
	TrieStorageManagers      map[string]data.StorageManager
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	StatusHandler            core.AppStatusHandler
	ChainID                  []byte
}

// NewCoreComponentsFactory initializes the factory which is responsible to creating core components
func NewCoreComponentsFactory(args CoreComponentsFactoryArgs) (*CoreComponentsFactory, error) {
	if args.Config == nil {
		return nil, ErrNilConfiguration
	}
	if check.IfNil(args.PathManager) {
		return nil, ErrNilPathManager
	}
	return &CoreComponentsFactory{
		config:      args.Config,
		pathManager: args.PathManager,
		shardId:     args.ShardId,
		chainID:     args.ChainID,
	}, nil
}

// Create creates the core components
func (ccf *CoreComponentsFactory) Create() (*CoreComponents, error) {
	hasher, err := factoryHasher.NewHasher(ccf.config.Hasher.Type)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrHasherCreation, err.Error())
	}

	internalMarshalizer, err := factoryMarshalizer.NewMarshalizer(ccf.config.Marshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (internal): %s", ErrMarshalizerCreation, err.Error())
	}

	vmMarshalizer, err := factoryMarshalizer.NewMarshalizer(ccf.config.VmMarshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (vm): %s", ErrMarshalizerCreation, err.Error())
	}

	txSignMarshalizer, err := factoryMarshalizer.NewMarshalizer(ccf.config.TxSignMarshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (tx sign): %s", ErrMarshalizerCreation, err.Error())
	}

	uint64ByteSliceConverter := uint64ByteSlice.NewBigEndianConverter()

	trieStorageManagers, trieContainer, err := ccf.createTries(internalMarshalizer, hasher)

	if err != nil {
		return nil, err
	}

	return &CoreComponents{
		Hasher:                   hasher,
		InternalMarshalizer:      internalMarshalizer,
		VmMarshalizer:            vmMarshalizer,
		TxSignMarshalizer:        txSignMarshalizer,
		TriesContainer:           trieContainer,
		TrieStorageManagers:      trieStorageManagers,
		Uint64ByteSliceConverter: uint64ByteSliceConverter,
		StatusHandler:            statusHandler.NewNilStatusHandler(),
		ChainID:                  ccf.chainID,
	}, nil
}

func (ccf *CoreComponentsFactory) createTries(
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
) (map[string]data.StorageManager, state.TriesHolder, error) {
	trieContainer := state.NewDataTriesHolder()
	trieFactoryArgs := factory.TrieFactoryArgs{
		EvictionWaitingListCfg:   ccf.config.EvictionWaitingList,
		SnapshotDbCfg:            ccf.config.TrieSnapshotDB,
		Marshalizer:              marshalizer,
		Hasher:                   hasher,
		PathManager:              ccf.pathManager,
		ShardId:                  ccf.shardId,
		TrieStorageManagerConfig: ccf.config.TrieStorageManagerConfig,
	}
	trieFactory, err := factory.NewTrieFactory(trieFactoryArgs)
	if err != nil {
		return nil, nil, err
	}

	trieStorageManagers := make(map[string]data.StorageManager)
	userStorageManager, userAccountTrie, err := trieFactory.Create(ccf.config.AccountsTrieStorage, ccf.config.StateTriesConfig.AccountsStatePruningEnabled)
	if err != nil {
		return nil, nil, err
	}
	trieContainer.Put([]byte(factory.UserAccountTrie), userAccountTrie)
	trieStorageManagers[factory.UserAccountTrie] = userStorageManager

	peerStorageManager, peerAccountsTrie, err := trieFactory.Create(ccf.config.PeerAccountsTrieStorage, ccf.config.StateTriesConfig.PeerStatePruningEnabled)
	if err != nil {
		return nil, nil, err
	}
	trieContainer.Put([]byte(factory.PeerAccountTrie), peerAccountsTrie)
	trieStorageManagers[factory.PeerAccountTrie] = peerStorageManager

	return trieStorageManagers, trieContainer, nil
}
