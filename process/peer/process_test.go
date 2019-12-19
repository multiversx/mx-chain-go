package peer_test

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/peer"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

func CreateMockArguments() peer.ArgValidatorStatisticsProcessor {
	economicsData, _ := economics.NewEconomicsData(
		&config.ConfigEconomics{
			EconomicsAddresses: config.EconomicsAddresses{
				CommunityAddress: "addr1",
				BurnAddress:      "addr2",
			},
			RewardsSettings: config.RewardsSettings{
				RewardsValue:        "1000",
				CommunityPercentage: 0.10,
				LeaderPercentage:    0.50,
				BurnPercentage:      0.40,
			},
			FeeSettings: config.FeeSettings{
				MaxGasLimitPerBlock: "10000000",
				MinGasPrice:         "10",
				MinGasLimit:         "10",
			},
			ValidatorSettings: config.ValidatorSettings{
				StakeValue:    "500",
				UnBoundPeriod: "5",
			},
		},
	)

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes:     nil,
		Marshalizer:      &mock.MarshalizerMock{},
		DataPool:         &mock.MetaPoolsHolderFake{},
		StorageService:   &mock.ChainStorerMock{},
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		AdrConv:          &mock.AddressConverterMock{},
		PeerAdapter:      getAccountsMock(),
		StakeValue:       economicsData.StakeValue(),
		Rater:            createMockRater(),
	}
	return arguments
}

func createMockRater() *mock.RaterMock {
	rater := mock.GetNewMockRater()
	return rater
}

func TestNewValidatorStatisticsProcessor_NilPeerAdaptersShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	arguments.PeerAdapter = nil
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilPeerAccountsAdapter, err)
}

func TestNewValidatorStatisticsProcessor_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	arguments.AdrConv = nil
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewValidatorStatisticsProcessor_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	arguments.NodesCoordinator = nil
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilNodesCoordinator, err)
}

func TestNewValidatorStatisticsProcessor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	arguments.ShardCoordinator = nil
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewValidatorStatisticsProcessor_NilStorageShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	arguments.StorageService = nil
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewValidatorStatisticsProcessor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	arguments.Marshalizer = nil
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewValidatorStatisticsProcessor_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	arguments.DataPool = nil
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestNewValidatorStatisticsProcessor(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.NotNil(t, validatorStatistics)
	assert.Nil(t, err)
}

func TestValidatorStatisticsProcessor_SaveInitialStateErrOnInvalidNode(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	initialNodes := []*sharding.InitialNode{{PubKey: "", Address: ""}}
	err := validatorStatistics.SaveInitialState(initialNodes, big.NewInt(100), uint32(5))

	assert.Equal(t, process.ErrInvalidInitialNodesState, err)
}

func TestValidatorStatisticsProcessor_SaveInitialStateErrOnWrongAddressConverter(t *testing.T) {
	t.Parallel()

	addressErr := errors.New("hex address error")
	addressConverter := &mock.AddressConverterStub{
		CreateAddressFromHexCalled: func(hexAddress string) (container state.AddressContainer, e error) {
			return nil, addressErr
		},
	}

	arguments := CreateMockArguments()
	arguments.InitialNodes = []*sharding.InitialNode{{PubKey: "aaaa", Address: "aaaa"}}
	arguments.AdrConv = addressConverter
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Equal(t, addressErr, err)
	assert.Nil(t, validatorStatistics)
}

func TestValidatorStatisticsProcessor_SaveInitialStateErrOnGetAccountFail(t *testing.T) {
	t.Parallel()

	adapterError := errors.New("account error")
	peerAdapters := &mock.AccountsStub{
		GetAccountWithJournalCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, adapterError
		},
	}

	addressConverter := &mock.AddressConverterStub{
		CreateAddressFromHexCalled: func(hexAddress string) (container state.AddressContainer, e error) {
			return &mock.AddressMock{}, nil
		},
	}

	arguments := CreateMockArguments()
	arguments.InitialNodes = []*sharding.InitialNode{{PubKey: "aaaa", Address: "aaaa"}}
	arguments.PeerAdapter = peerAdapters
	arguments.AdrConv = addressConverter
	_, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Equal(t, adapterError, err)
}

func TestValidatorStatisticsProcessor_SaveInitialStateGetAccountReturnsInvalid(t *testing.T) {
	t.Parallel()

	peerAdapter := &mock.AccountsStub{
		GetAccountWithJournalCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return &mock.AccountWrapMock{}, nil
		},
	}

	addressConverter := &mock.AddressConverterStub{
		CreateAddressFromHexCalled: func(hexAddress string) (container state.AddressContainer, e error) {
			return &mock.AddressMock{}, nil
		},
	}

	arguments := CreateMockArguments()
	arguments.InitialNodes = []*sharding.InitialNode{{PubKey: "aaaa", Address: "aaaa"}}
	arguments.PeerAdapter = peerAdapter
	arguments.AdrConv = addressConverter
	_, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Equal(t, process.ErrInvalidPeerAccount, err)
}

func TestValidatorStatisticsProcessor_SaveInitialStateSetAddressErrors(t *testing.T) {
	t.Parallel()

	saveAccountError := errors.New("save account error")
	peerAccount, _ := state.NewPeerAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {

		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return saveAccountError
		},
	})
	peerAdapter := &mock.AccountsStub{
		GetAccountWithJournalCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return peerAccount, nil
		},
	}

	addressConverter := &mock.AddressConverterStub{
		CreateAddressFromHexCalled: func(hexAddress string) (container state.AddressContainer, e error) {
			return &mock.AddressMock{}, nil
		},
	}

	arguments := CreateMockArguments()
	arguments.InitialNodes = []*sharding.InitialNode{{PubKey: "aaaa", Address: "aaaa"}}
	arguments.PeerAdapter = peerAdapter
	arguments.AdrConv = addressConverter
	_, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Equal(t, saveAccountError, err)
}

func TestValidatorStatisticsProcessor_SaveInitialStateCommitErrors(t *testing.T) {
	t.Parallel()

	commitError := errors.New("commit error")
	peerAccount, _ := state.NewPeerAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {

		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return nil
		},
	})
	peerAdapter := &mock.AccountsStub{
		GetAccountWithJournalCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return peerAccount, nil
		},
		CommitCalled: func() (bytes []byte, e error) {
			return nil, commitError
		},
	}

	addressConverter := &mock.AddressConverterStub{
		CreateAddressFromHexCalled: func(hexAddress string) (container state.AddressContainer, e error) {
			return &mock.AddressMock{}, nil
		},
	}

	arguments := CreateMockArguments()
	arguments.PeerAdapter = peerAdapter
	arguments.AdrConv = addressConverter
	_, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Equal(t, commitError, err)
}

func TestValidatorStatisticsProcessor_SaveInitialStateCommit(t *testing.T) {
	t.Parallel()

	peerAccount, _ := state.NewPeerAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {

		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return nil
		},
	})
	peerAdapter := &mock.AccountsStub{
		GetAccountWithJournalCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return peerAccount, nil
		},
		CommitCalled: func() (bytes []byte, e error) {
			return nil, nil
		},
	}

	addressConverter := &mock.AddressConverterStub{
		CreateAddressFromHexCalled: func(hexAddress string) (container state.AddressContainer, e error) {
			return &mock.AddressMock{}, nil
		},
	}

	arguments := CreateMockArguments()
	arguments.PeerAdapter = peerAdapter
	arguments.AdrConv = addressConverter
	_, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, err)
}

func TestValidatorStatisticsProcessor_IsNodeValidEmptyAddressShoudErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	isValid := validatorStatistics.IsNodeValid(&sharding.InitialNode{Address: "", PubKey: "aaaaa"})
	assert.False(t, isValid)
}

func TestValidatorStatisticsProcessor_IsNodeValidEmptyPubKeyShoudErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	isValid := validatorStatistics.IsNodeValid(&sharding.InitialNode{Address: "aaaaa", PubKey: ""})
	assert.False(t, isValid)
}

func TestValidatorStatisticsProcessor_IsNodeValid(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	isValid := validatorStatistics.IsNodeValid(&sharding.InitialNode{Address: "aaaaa", PubKey: "aaaaaa"})
	assert.True(t, isValid)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateReturnsRootHashForGenesis(t *testing.T) {
	t.Parallel()

	expectedRootHash := []byte("root hash")
	peerAdapter := getAccountsMock()
	peerAdapter.RootHashCalled = func() (bytes []byte, e error) {
		return expectedRootHash, nil
	}

	arguments := CreateMockArguments()
	arguments.PeerAdapter = peerAdapter
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	header.Nonce = 0
	rootHash, err := validatorStatistics.UpdatePeerState(header)

	assert.Nil(t, err)
	assert.Equal(t, expectedRootHash, rootHash)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateReturnsErrForRootHashErr(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	peerAdapter := getAccountsMock()
	peerAdapter.RootHashCalled = func() (bytes []byte, e error) {
		return nil, expectedError
	}

	arguments := CreateMockArguments()
	arguments.PeerAdapter = peerAdapter
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	header.Nonce = 0
	_, err := validatorStatistics.UpdatePeerState(header)

	assert.Equal(t, expectedError, err)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateComputeValidatorErrShouldError(t *testing.T) {
	t.Parallel()

	computeValidatorsErr := errors.New("compute validators error")

	arguments := CreateMockArguments()
	arguments.InitialNodes = nil
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
			return nil, computeValidatorsErr
		},
	}
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	_, err := validatorStatistics.UpdatePeerState(header)

	assert.Equal(t, computeValidatorsErr, err)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateCreateAddressFromPublicKeyBytesErr(t *testing.T) {
	t.Parallel()

	createAddressErr := errors.New("create address error")

	arguments := CreateMockArguments()
	arguments.InitialNodes = nil
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
			return []sharding.Validator{&mock.ValidatorMock{}}, nil
		},
	}
	arguments.AdrConv = &mock.AddressConverterStub{
		CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
			return nil, createAddressErr
		},
	}
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	_, err := validatorStatistics.UpdatePeerState(header)

	assert.Equal(t, createAddressErr, err)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateGetExistingAccountErr(t *testing.T) {
	t.Parallel()

	existingAccountErr := errors.New("existing account err")
	adapter := getAccountsMock()
	adapter.GetAccountWithJournalCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return nil, existingAccountErr
	}

	arguments := CreateMockArguments()
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
			return []sharding.Validator{&mock.ValidatorMock{}}, nil
		},
	}
	arguments.AdrConv = &mock.AddressConverterStub{
		CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
			return &mock.AddressMock{}, nil
		},
	}
	arguments.PeerAdapter = adapter
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	_, err := validatorStatistics.UpdatePeerState(header)

	assert.Equal(t, existingAccountErr, err)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateGetExistingAccountInvalidType(t *testing.T) {
	t.Parallel()

	adapter := getAccountsMock()
	adapter.GetAccountWithJournalCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &mock.AccountWrapMock{}, nil
	}

	arguments := CreateMockArguments()
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
			return []sharding.Validator{&mock.ValidatorMock{}}, nil
		},
	}
	arguments.AdrConv = &mock.AddressConverterStub{
		CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
			return &mock.AddressMock{}, nil
		},
	}
	arguments.PeerAdapter = adapter
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	_, err := validatorStatistics.UpdatePeerState(header)

	assert.Equal(t, process.ErrInvalidPeerAccount, err)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateGetHeaderError(t *testing.T) {
	t.Parallel()

	getHeaderError := errors.New("get header error")
	adapter := getAccountsMock()
	marshalizer := &mock.MarshalizerStub{}

	adapter.GetAccountWithJournalCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &mock.PeerAccountHandlerMock{}, nil
	}
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := CreateMockArguments()
	arguments.Marshalizer = marshalizer
	arguments.DataPool = &mock.MetaPoolsHolderFake{
		MetaBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, false
				},
			}
		},
	}
	arguments.StorageService = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, e error) {
					return nil, getHeaderError
				},
			}
		},
	}
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
			return []sharding.Validator{&mock.ValidatorMock{}, &mock.ValidatorMock{}}, nil
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.AdrConv = &mock.AddressConverterStub{
		CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
			return &mock.AddressMock{}, nil
		},
	}
	arguments.PeerAdapter = adapter
	arguments.Rater = mock.GetNewMockRater()
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	header.Nonce = 2
	_, err := validatorStatistics.UpdatePeerState(header)

	assert.Equal(t, process.ErrMissingHeader, err)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateGetHeaderUnmarshalError(t *testing.T) {
	t.Parallel()

	getHeaderUnmarshalError := errors.New("get header unmarshal error")
	adapter := getAccountsMock()
	marshalizer := &mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return getHeaderUnmarshalError
		},
	}

	adapter.GetAccountWithJournalCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &mock.PeerAccountHandlerMock{}, nil
	}
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := CreateMockArguments()
	arguments.Marshalizer = marshalizer
	arguments.DataPool = &mock.MetaPoolsHolderFake{
		MetaBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, false
				},
			}
		},
	}
	arguments.StorageService = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, e error) {
					return nil, nil
				},
			}
		},
	}
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
			return []sharding.Validator{&mock.ValidatorMock{}, &mock.ValidatorMock{}}, nil
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.AdrConv = &mock.AddressConverterStub{
		CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
			return &mock.AddressMock{}, nil
		},
	}
	arguments.PeerAdapter = adapter
	arguments.Rater = mock.GetNewMockRater()

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	header.Nonce = 2
	_, err := validatorStatistics.UpdatePeerState(header)

	assert.Equal(t, process.ErrUnmarshalWithoutSuccess, err)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateCallsIncrease(t *testing.T) {
	t.Parallel()

	adapter := getAccountsMock()
	increaseLeaderCalled := false
	increaseValidatorCalled := false
	marshalizer := &mock.MarshalizerStub{}

	adapter.GetAccountWithJournalCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &mock.PeerAccountHandlerMock{
			IncreaseLeaderSuccessRateWithJournalCalled: func() error {
				increaseLeaderCalled = true
				return nil
			},
			IncreaseValidatorSuccessRateWithJournalCalled: func() error {
				increaseValidatorCalled = true
				return nil
			},
		}, nil
	}
	adapter.RootHashCalled = func() (bytes []byte, e error) {
		return nil, nil
	}
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := CreateMockArguments()
	arguments.InitialNodes = nil
	arguments.Marshalizer = marshalizer
	arguments.DataPool = &mock.MetaPoolsHolderFake{
		MetaBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, false
				},
			}
		},
	}
	arguments.StorageService = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, e error) {
					return nil, nil
				},
			}
		},
	}
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
			return []sharding.Validator{&mock.ValidatorMock{}, &mock.ValidatorMock{}}, nil
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.AdrConv = &mock.AddressConverterStub{
		CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
			return &mock.AddressMock{}, nil
		},
	}
	arguments.PeerAdapter = adapter
	arguments.Rater = mock.GetNewMockRater()
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))

	marshalizer.UnmarshalCalled = func(obj interface{}, buff []byte) error {
		switch v := obj.(type) {
		case *block.MetaBlock:
			*v = block.MetaBlock{}
		case *block.Header:
			*v = block.Header{}
		default:
			fmt.Println(v)
		}

		return nil
	}

	_, err := validatorStatistics.UpdatePeerState(header)

	assert.Nil(t, err)
	assert.True(t, increaseLeaderCalled)
	assert.True(t, increaseValidatorCalled)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateCheckForMissedBlocksErr(t *testing.T) {
	t.Parallel()

	adapter := getAccountsMock()
	missedBlocksErr := errors.New("missed blocks error")
	marshalizer := &mock.MarshalizerStub{}

	adapter.GetAccountWithJournalCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &mock.PeerAccountHandlerMock{
			DecreaseLeaderSuccessRateWithJournalCalled: func() error {
				return missedBlocksErr
			},
		}, nil
	}
	adapter.RootHashCalled = func() (bytes []byte, e error) {
		return nil, nil
	}
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := CreateMockArguments()
	arguments.DataPool = &mock.MetaPoolsHolderFake{
		MetaBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, false
				},
			}
		},
	}
	arguments.StorageService = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				GetCalled: func(key []byte) (bytes []byte, e error) {
					return nil, nil
				},
			}
		},
	}
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
			return []sharding.Validator{&mock.ValidatorMock{}, &mock.ValidatorMock{}}, nil
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.AdrConv = &mock.AddressConverterStub{
		CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
			return &mock.AddressMock{}, nil
		},
	}
	arguments.PeerAdapter = adapter
	arguments.Marshalizer = marshalizer
	arguments.Rater = mock.GetNewMockRater()

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	header.Nonce = 2
	header.Round = 2

	marshalizer.UnmarshalCalled = func(obj interface{}, buff []byte) error {
		switch v := obj.(type) {
		case *block.MetaBlock:
			*v = block.MetaBlock{
				Nonce: 0,
			}
		case *block.Header:
			*v = block.Header{}
		default:
			fmt.Println(v)
		}

		return nil
	}

	_, err := validatorStatistics.UpdatePeerState(header)

	assert.Equal(t, missedBlocksErr, err)
}

func TestValidatorStatisticsProcessor_CheckForMissedBlocksNoMissedBlocks(t *testing.T) {
	t.Parallel()

	computeValidatorGroupCalled := false
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := CreateMockArguments()
	arguments.Marshalizer = &mock.MarshalizerMock{}
	arguments.DataPool = &mock.MetaPoolsHolderFake{}
	arguments.StorageService = &mock.ChainStorerMock{}
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
			computeValidatorGroupCalled = true
			return nil, nil
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.AdrConv = &mock.AddressConverterMock{}
	arguments.PeerAdapter = getAccountsMock()

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	err := validatorStatistics.CheckForMissedBlocks(1, 0, []byte("prev"), 0)
	assert.Nil(t, err)
	assert.False(t, computeValidatorGroupCalled)

	err = validatorStatistics.CheckForMissedBlocks(1, 1, []byte("prev"), 0)
	assert.Nil(t, err)
	assert.False(t, computeValidatorGroupCalled)

	err = validatorStatistics.CheckForMissedBlocks(2, 1, []byte("prev"), 0)
	assert.Nil(t, err)
	assert.False(t, computeValidatorGroupCalled)
}

func TestValidatorStatisticsProcessor_CheckForMissedBlocksErrOnComputeValidatorList(t *testing.T) {
	t.Parallel()

	computeErr := errors.New("compute err")
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := CreateMockArguments()
	arguments.Marshalizer = &mock.MarshalizerMock{}
	arguments.DataPool = &mock.MetaPoolsHolderFake{}
	arguments.StorageService = &mock.ChainStorerMock{}
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
			return nil, computeErr
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.AdrConv = &mock.AddressConverterMock{}
	arguments.PeerAdapter = getAccountsMock()
	arguments.Rater = mock.GetNewMockRater()
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	err := validatorStatistics.CheckForMissedBlocks(2, 0, []byte("prev"), 0)
	assert.Equal(t, computeErr, err)
}

func TestValidatorStatisticsProcessor_CheckForMissedBlocksErrOnGetPeerAcc(t *testing.T) {
	t.Parallel()

	peerAccErr := errors.New("peer acc err")
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := CreateMockArguments()
	arguments.Marshalizer = &mock.MarshalizerMock{}
	arguments.DataPool = &mock.MetaPoolsHolderFake{}
	arguments.StorageService = &mock.ChainStorerMock{}
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
			return []sharding.Validator{
				&mock.ValidatorMock{},
			}, nil
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.AdrConv = &mock.AddressConverterStub{
		CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (addressContainer state.AddressContainer, e error) {
			return nil, peerAccErr
		},
	}
	arguments.PeerAdapter = getAccountsMock()
	arguments.Rater = mock.GetNewMockRater()
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	err := validatorStatistics.CheckForMissedBlocks(2, 0, []byte("prev"), 0)
	assert.Equal(t, peerAccErr, err)
}

func TestValidatorStatisticsProcessor_CheckForMissedBlocksErrOnDecrease(t *testing.T) {
	t.Parallel()

	decreaseErr := errors.New("peer acc err")
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerAdapter := getAccountsMock()
	peerAdapter.GetAccountWithJournalCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &mock.PeerAccountHandlerMock{
			DecreaseLeaderSuccessRateWithJournalCalled: func() error {
				return decreaseErr
			},
		}, nil
	}

	arguments := CreateMockArguments()
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
			return []sharding.Validator{
				&mock.ValidatorMock{},
			}, nil
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.AdrConv = &mock.AddressConverterStub{
		CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (addressContainer state.AddressContainer, e error) {
			return nil, nil
		},
	}
	arguments.PeerAdapter = peerAdapter
	arguments.Rater = mock.GetNewMockRater()
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	err := validatorStatistics.CheckForMissedBlocks(2, 0, []byte("prev"), 0)
	assert.Equal(t, decreaseErr, err)
}

func TestValidatorStatisticsProcessor_CheckForMissedBlocksCallsDecrease(t *testing.T) {
	t.Parallel()

	currentHeaderRound := 10
	previousHeaderRound := 4
	decreaseCount := 0

	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerAdapter := getAccountsMock()
	peerAdapter.GetAccountWithJournalCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &mock.PeerAccountHandlerMock{
			DecreaseLeaderSuccessRateWithJournalCalled: func() error {
				decreaseCount++
				return nil
			},
		}, nil
	}

	arguments := CreateMockArguments()
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
			return []sharding.Validator{
				&mock.ValidatorMock{},
			}, nil
		},
	}
	arguments.ShardCoordinator = shardCoordinatorMock
	arguments.AdrConv = &mock.AddressConverterStub{
		CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (addressContainer state.AddressContainer, e error) {
			return nil, nil
		},
	}
	arguments.PeerAdapter = peerAdapter
	arguments.Rater = mock.GetNewMockRater()
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	_ = validatorStatistics.CheckForMissedBlocks(uint64(currentHeaderRound), uint64(previousHeaderRound), []byte("prev"), 0)
	assert.Equal(t, currentHeaderRound-previousHeaderRound-1, decreaseCount)
}

func TestValidatorStatisticsProcessor_GetMatchingPrevShardDataEmptySDReturnsNil(t *testing.T) {
	arguments := CreateMockArguments()

	currentShardData := block.ShardData{}
	shardInfo := make([]block.ShardData, 0)

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	sd := validatorStatistics.GetMatchingPrevShardData(currentShardData, shardInfo)

	assert.Nil(t, sd)
}

func TestValidatorStatisticsProcessor_GetMatchingPrevShardDataNoMatch(t *testing.T) {
	arguments := CreateMockArguments()

	currentShardData := block.ShardData{ShardID: 1, Nonce: 10}
	shardInfo := []block.ShardData{{ShardID: 1, Nonce: 8}, {ShardID: 2, Nonce: 9}}

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	sd := validatorStatistics.GetMatchingPrevShardData(currentShardData, shardInfo)

	assert.Nil(t, sd)
}

func TestValidatorStatisticsProcessor_GetMatchingPrevShardDataFindsMatch(t *testing.T) {
	arguments := CreateMockArguments()

	currentShardData := block.ShardData{ShardID: 1, Nonce: 10}
	shardInfo := []block.ShardData{{ShardID: 1, Nonce: 9}, {ShardID: 2, Nonce: 9}}

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	sd := validatorStatistics.GetMatchingPrevShardData(currentShardData, shardInfo)

	assert.NotNil(t, sd)
}

func TestValidatorStatisticsProcessor_LoadPreviousShardHeadersSkipsFirst(t *testing.T) {
	arguments := CreateMockArguments()

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	currentHeader := &block.MetaBlock{Nonce: 1, ShardInfo: []block.ShardData{{Nonce: 1}}}
	prevHeader := &block.MetaBlock{}

	err := validatorStatistics.LoadPreviousShardHeaders(currentHeader, prevHeader)
	assert.Nil(t, err)
	prevShardHeaders := validatorStatistics.PrevShardInfo()
	assert.Empty(t, prevShardHeaders)
}

func TestValidatorStatisticsProcessor_LoadPreviousShardHeadersMetaSkipsFirst(t *testing.T) {
	arguments := CreateMockArguments()

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	currentHeader := &block.MetaBlock{Nonce: 2, ShardInfo: []block.ShardData{{Nonce: 1}}}
	prevHeader := &block.MetaBlock{}

	err := validatorStatistics.LoadPreviousShardHeadersMeta(currentHeader, prevHeader)
	assert.Nil(t, err)
	prevShardHeaders := validatorStatistics.PrevShardInfo()
	assert.Empty(t, prevShardHeaders)
}

func TestValidatorStatisticsProcessor_LoadPreviousShardHeadersFindsMatchInCurrentAndFindsPrev(t *testing.T) {
	arguments := CreateMockArguments()

	sd1 := block.ShardData{ShardID: 1, Nonce: 1}
	sd2 := block.ShardData{ShardID: 1, Nonce: 2}
	sd3 := block.ShardData{ShardID: 1, Nonce: 3}

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	currentHeader := &block.MetaBlock{Nonce: 3, ShardInfo: []block.ShardData{sd3, sd2}}
	prevHeader := &block.MetaBlock{Nonce: 2, ShardInfo: []block.ShardData{sd1}}

	err := validatorStatistics.LoadPreviousShardHeaders(currentHeader, prevHeader)
	assert.Nil(t, err)

	prevShardHeaders := validatorStatistics.PrevShardInfo()
	sdKey2 := validatorStatistics.BuildShardDataKey(sd2)
	sdKey3 := validatorStatistics.BuildShardDataKey(sd3)
	assert.Equal(t, map[string]block.ShardData{sdKey3: sd2, sdKey2: sd1}, prevShardHeaders)
}

func TestValidatorStatisticsProcessor_LoadPreviousShardHeadersMeta(t *testing.T) {
	sd1 := block.ShardData{ShardID: 1, Nonce: 1}
	sd2 := block.ShardData{ShardID: 1, Nonce: 2}
	sd3 := block.ShardData{ShardID: 1, Nonce: 3}

	currentHeader := &block.MetaBlock{Nonce: 3, ShardInfo: []block.ShardData{sd3, sd2}}
	prevHeader := &block.Header{Nonce: sd1.Nonce, ShardId: sd1.ShardID}

	arguments := CreateMockArguments()
	arguments.DataPool = &mock.MetaPoolsHolderFake{
		ShardHeadersCalled: func() storage.Cacher {
			return &mock.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					return prevHeader, true
				},
			}
		},
	}

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	err := validatorStatistics.LoadPreviousShardHeadersMeta(currentHeader, &block.MetaBlock{})
	assert.Nil(t, err)

	prevShardHeaders := validatorStatistics.PrevShardInfo()
	sdKey2 := validatorStatistics.BuildShardDataKey(sd2)
	sdKey3 := validatorStatistics.BuildShardDataKey(sd3)

	// For every key sd1 since our mocked storer returns prevHeader every time
	assert.Equal(t, map[string]block.ShardData{sdKey3: sd1, sdKey2: sd1}, prevShardHeaders)
}

func TestValidatorStatisticsProcessor_LoadPreviousShardHeadersLoadsMissingFromStorage(t *testing.T) {
	arguments := CreateMockArguments()
	sd1 := block.ShardData{ShardID: 1, Nonce: 1}
	sd2 := block.ShardData{ShardID: 1, Nonce: 2}
	sd3 := block.ShardData{ShardID: 1, Nonce: 3}
	currentHeader := &block.MetaBlock{Nonce: 4, ShardInfo: []block.ShardData{sd3, sd2}}
	prevHeader := &block.MetaBlock{Nonce: 3, ShardInfo: []block.ShardData{}}
	storageHeader := &block.MetaBlock{Nonce: 2, ShardInfo: []block.ShardData{sd1}}

	arguments.DataPool = &mock.MetaPoolsHolderFake{
		MetaBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					return storageHeader, true
				},
			}
		},
	}

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	err := validatorStatistics.LoadPreviousShardHeaders(currentHeader, prevHeader)
	assert.Nil(t, err)

	prevShardHeaders := validatorStatistics.PrevShardInfo()
	sdKey2 := validatorStatistics.BuildShardDataKey(sd2)
	sdKey3 := validatorStatistics.BuildShardDataKey(sd3)
	assert.Equal(t, map[string]block.ShardData{sdKey3: sd2, sdKey2: sd1}, prevShardHeaders)
}

func TestValidatorStatisticsProcessor_LoadPreviousShardHeadersErrForStorage(t *testing.T) {
	arguments := CreateMockArguments()

	sd2 := block.ShardData{ShardID: 1, Nonce: 2}
	sd3 := block.ShardData{ShardID: 1, Nonce: 3}
	currentHeader := &block.MetaBlock{Nonce: 4, ShardInfo: []block.ShardData{sd3, sd2}}
	prevHeader := &block.MetaBlock{Nonce: 3, ShardInfo: []block.ShardData{}}

	arguments.DataPool = &mock.MetaPoolsHolderFake{
		MetaBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, false
				},
			}
		},
	}

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	err := validatorStatistics.LoadPreviousShardHeaders(currentHeader, prevHeader)
	assert.Equal(t, dataRetriever.ErrNilHeadersStorage, err)
}

func TestValidatorStatisticsProcessor_LoadPreviousShardHeadersMetaErrForStorage(t *testing.T) {
	arguments := CreateMockArguments()

	sd2 := block.ShardData{ShardID: 1, Nonce: 2}
	sd3 := block.ShardData{ShardID: 1, Nonce: 3}
	currentHeader := &block.MetaBlock{Nonce: 4, ShardInfo: []block.ShardData{sd3, sd2}}
	prevHeader := &block.MetaBlock{Nonce: 3, ShardInfo: []block.ShardData{}}

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	err := validatorStatistics.LoadPreviousShardHeadersMeta(currentHeader, prevHeader)
	assert.Equal(t, dataRetriever.ErrNilCacher, err)
}

func TestValidatorStatisticsProcessor_LoadPreviousShardHeadersErrIfStillMissing(t *testing.T) {
	arguments := CreateMockArguments()
	sd2 := block.ShardData{ShardID: 1, Nonce: 2}
	sd3 := block.ShardData{ShardID: 1, Nonce: 3}
	currentHeader := &block.MetaBlock{Nonce: 4, ShardInfo: []block.ShardData{sd3, sd2}}
	prevHeader := &block.MetaBlock{Nonce: 3, ShardInfo: []block.ShardData{}}
	storageHeader := &block.MetaBlock{Nonce: 1, ShardInfo: []block.ShardData{}}

	arguments.DataPool = &mock.MetaPoolsHolderFake{
		MetaBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					return storageHeader, true
				},
			}
		},
	}

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	err := validatorStatistics.LoadPreviousShardHeaders(currentHeader, prevHeader)
	assert.Equal(t, process.ErrMissingShardDataInStorage, err)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateCallsPubKeyForValidator(t *testing.T) {
	pubKeyCalled := false
	addressCalled := false
	arguments := CreateMockArguments()
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
			return []sharding.Validator{&mock.ValidatorMock{
				PubKeyCalled: func() []byte {
					pubKeyCalled = true
					return make([]byte, 0)
				},
				AddressCalled: func() []byte {
					addressCalled = true
					return make([]byte, 0)
				},
			}, &mock.ValidatorMock{}}, nil
		},
	}

	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)
	header := getMetaHeaderHandler([]byte("header"))

	_, _ = validatorStatistics.UpdatePeerState(header)

	assert.True(t, pubKeyCalled)
	assert.False(t, addressCalled)
}

func getMetaHeaderHandler(randSeed []byte) *block.MetaBlock {
	return &block.MetaBlock{
		Nonce:         1,
		PrevRandSeed:  randSeed,
		PrevHash:      randSeed,
		PubKeysBitmap: randSeed,
	}
}

func getAccountsMock() *mock.AccountsStub {
	return &mock.AccountsStub{
		CommitCalled: func() (bytes []byte, e error) {
			return make([]byte, 0), nil
		},
		GetAccountWithJournalCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return &mock.AccountWrapMock{}, nil
		},
	}
}
