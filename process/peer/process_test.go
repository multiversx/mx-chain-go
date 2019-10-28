package peer_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/peer"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNewValidatorStatisticsProcessor_NilPeerAdaptersShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: &mock.AddressConverterMock{},
		PeerAdapter: nil,
	}
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilPeerAccountsAdapter, err)
}

func TestNewValidatorStatisticsProcessor_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: nil,
		PeerAdapter: getAccountsMock(),
	}
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewValidatorStatisticsProcessor_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: nil,
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: &mock.AddressConverterMock{},
		PeerAdapter: getAccountsMock(),
	}
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilNodesCoordinator, err)
}

func TestNewValidatorStatisticsProcessor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: nil,
		AdrConv: &mock.AddressConverterMock{},
		PeerAdapter: getAccountsMock(),
	}
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewValidatorStatisticsProcessor_NilShardHeaderStorageShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: nil,
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: &mock.AddressConverterMock{},
		PeerAdapter: getAccountsMock(),
	}
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilShardHeaderStorage, err)
}

func TestNewValidatorStatisticsProcessor_NilMetaHeaderStorageShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: nil,
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: &mock.AddressConverterMock{},
		PeerAdapter: getAccountsMock(),
	}
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilMetaHeaderStorage, err)
}

func TestNewValidatorStatisticsProcessor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: nil,
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: &mock.AddressConverterMock{},
		PeerAdapter: getAccountsMock(),
	}
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, validatorStatistics)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewValidatorStatisticsProcessor(t *testing.T) {
	t.Parallel()

	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: &mock.AddressConverterMock{},
		PeerAdapter: getAccountsMock(),
	}
	validatorStatistics, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.NotNil(t, validatorStatistics)
	assert.Nil(t, err)
}

func TestValidatorStatisticsProcessor_SaveInitialStateErrOnInvalidNode(t *testing.T) {
	t.Parallel()

	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: &mock.AddressConverterMock{},
		PeerAdapter: getAccountsMock(),
	}
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	initialNodes := []*sharding.InitialNode{{PubKey:"", Address: ""}}
	err := validatorStatistics.SaveInitialState(initialNodes)

	assert.Equal(t, process.ErrInvalidInitialNodesState, err)
}

func TestValidatorStatisticsProcessor_SaveInitialStateErrOnWrongAddressConverter(t *testing.T) {
	t.Parallel()

	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	addressErr := errors.New("hex address error")
	addressConverter := &mock.AddressConverterStub{
		CreateAddressFromHexCalled: func(hexAddress string) (container state.AddressContainer, e error) {
			return nil, addressErr
		},
	}

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: addressConverter,
		PeerAdapter: getAccountsMock(),
	}
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	initialNodes := []*sharding.InitialNode{{PubKey:"aaaa", Address: "aaaa"}}
	err := validatorStatistics.SaveInitialState(initialNodes)

	assert.Equal(t, addressErr, err)
}

func TestValidatorStatisticsProcessor_SaveInitialStateErrOnGetAccountFail(t *testing.T) {
	t.Parallel()

	adapterError := errors.New("account error")
	peerAdapters := &mock.AccountsStub{
		GetAccountWithJournalCalled: func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
			return nil, adapterError
		},
	}
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	addressConverter := &mock.AddressConverterStub{
		CreateAddressFromHexCalled: func(hexAddress string) (container state.AddressContainer, e error) {
			return &mock.AddressMock{}, nil
		},
	}
	initialNodes := []*sharding.InitialNode{{PubKey:"aaaa", Address: "aaaa"}}

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: initialNodes,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: addressConverter,
		PeerAdapter: peerAdapters,
	}
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
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	addressConverter := &mock.AddressConverterStub{
		CreateAddressFromHexCalled: func(hexAddress string) (container state.AddressContainer, e error) {
			return &mock.AddressMock{}, nil
		},
	}

	initialNodes := []*sharding.InitialNode{{PubKey:"aaaa", Address: "aaaa"}}

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: initialNodes,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: addressConverter,
		PeerAdapter: peerAdapter,
	}
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
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	addressConverter := &mock.AddressConverterStub{
		CreateAddressFromHexCalled: func(hexAddress string) (container state.AddressContainer, e error) {
			return &mock.AddressMock{}, nil
		},
	}

	initialNodes := []*sharding.InitialNode{{PubKey:"aaaa", Address: "aaaa"}}

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: initialNodes,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: addressConverter,
		PeerAdapter: peerAdapter,
	}
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
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	addressConverter := &mock.AddressConverterStub{
		CreateAddressFromHexCalled: func(hexAddress string) (container state.AddressContainer, e error) {
			return &mock.AddressMock{}, nil
		},
	}

	initialNodes := []*sharding.InitialNode{{PubKey:"aaaa", Address: "aaaa"}}

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: initialNodes,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: addressConverter,
		PeerAdapter: peerAdapter,
	}
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
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	addressConverter := &mock.AddressConverterStub{
		CreateAddressFromHexCalled: func(hexAddress string) (container state.AddressContainer, e error) {
			return &mock.AddressMock{}, nil
		},
	}

	initialNodes := []*sharding.InitialNode{{PubKey:"aaaa", Address: "aaaa"}}

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: initialNodes,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: addressConverter,
		PeerAdapter: peerAdapter,
	}
	_, err := peer.NewValidatorStatisticsProcessor(arguments)

	assert.Nil(t, err)
}

func TestValidatorStatisticsProcessor_IsNodeValidEmptyAddressShoudErr(t *testing.T) {
	t.Parallel()

	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: &mock.AddressConverterMock{},
		PeerAdapter: getAccountsMock(),
	}
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	isValid := validatorStatistics.IsNodeValid(&sharding.InitialNode{Address: "", PubKey: "aaaaa"})
	assert.False(t, isValid)
}

func TestValidatorStatisticsProcessor_IsNodeValidEmptyPubKeyShoudErr(t *testing.T) {
	t.Parallel()

	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: &mock.AddressConverterMock{},
		PeerAdapter: getAccountsMock(),
	}
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	isValid := validatorStatistics.IsNodeValid(&sharding.InitialNode{Address: "aaaaa", PubKey: ""})
	assert.False(t, isValid)
}

func TestValidatorStatisticsProcessor_IsNodeValid(t *testing.T) {
	t.Parallel()

	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: &mock.AddressConverterMock{},
		PeerAdapter: getAccountsMock(),
	}
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	isValid := validatorStatistics.IsNodeValid(&sharding.InitialNode{Address: "aaaaa", PubKey: "aaaaaa"})
	assert.True(t, isValid)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateComputeValidatorErrShouldError(t *testing.T) {
	t.Parallel()

	computeValidatorsErr := errors.New("compute validators error")
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{
			ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
				return nil, computeValidatorsErr
			},
		},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: &mock.AddressConverterMock{},
		PeerAdapter: getAccountsMock(),
	}
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	_, err := validatorStatistics.UpdatePeerState(header)

	assert.Equal(t, computeValidatorsErr, err)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateCreateAddressFromPublicKeyBytesErr(t *testing.T) {
	t.Parallel()

	createAddressErr := errors.New("create address error")
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{
			ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
				return []sharding.Validator{&mock.ValidatorMock{}}, nil
			},
		},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: &mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
				return nil, createAddressErr
			},
		},
		PeerAdapter: getAccountsMock(),
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
	adapter.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return nil, existingAccountErr
	}
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{
			ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
				return []sharding.Validator{&mock.ValidatorMock{}}, nil
			},
		},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: &mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
				return &mock.AddressMock{}, nil
			},
		},
		PeerAdapter: adapter,
	}
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	_, err := validatorStatistics.UpdatePeerState(header)

	assert.Equal(t, existingAccountErr, err)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateGetExistingAccountInvalidType(t *testing.T) {
	t.Parallel()

	adapter := getAccountsMock()
	adapter.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &mock.AccountWrapMock{}, nil
	}
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: &mock.MarshalizerMock{},
		ShardHeaderStorage: &mock.StorerStub{},
		MetaHeaderStorage: &mock.StorerStub{},
		NodesCoordinator: &mock.NodesCoordinatorMock{
			ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
				return []sharding.Validator{&mock.ValidatorMock{}}, nil
			},
		},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: &mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
				return &mock.AddressMock{}, nil
			},
		},
		PeerAdapter: adapter,
	}
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
	shardHeaderStorage := &mock.StorerStub{
		GetCalled: func(key []byte) (bytes []byte, e error) {
			return nil, getHeaderError
		},
	}
	metaHeaderStorage := &mock.StorerStub{}
	adapter.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &mock.PeerAccountHandlerMock{}, nil
	}
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: marshalizer,
		ShardHeaderStorage: shardHeaderStorage,
		MetaHeaderStorage: metaHeaderStorage,
		NodesCoordinator: &mock.NodesCoordinatorMock{
			ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
				return []sharding.Validator{&mock.ValidatorMock{}, &mock.ValidatorMock{}}, nil
			},
		},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: &mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
				return &mock.AddressMock{}, nil
			},
		},
		PeerAdapter: adapter,
	}
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	_, err := validatorStatistics.UpdatePeerState(header)

	assert.Equal(t, getHeaderError, err)
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
	shardHeaderStorage := &mock.StorerStub{
		GetCalled: func(key []byte) (bytes []byte, e error) {
			return nil, nil
		},
	}
	metaHeaderStorage := &mock.StorerStub{}
	adapter.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &mock.PeerAccountHandlerMock{}, nil
	}
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: marshalizer,
		ShardHeaderStorage: shardHeaderStorage,
		MetaHeaderStorage: metaHeaderStorage,
		NodesCoordinator: &mock.NodesCoordinatorMock{
			ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
				return []sharding.Validator{&mock.ValidatorMock{}, &mock.ValidatorMock{}}, nil
			},
		},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: &mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
				return &mock.AddressMock{}, nil
			},
		},
		PeerAdapter: adapter,
	}
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	_, err := validatorStatistics.UpdatePeerState(header)

	assert.Equal(t, getHeaderUnmarshalError, err)
}

func TestValidatorStatisticsProcessor_UpdatePeerStateCallsIncrease(t *testing.T) {
	t.Parallel()

	adapter := getAccountsMock()
	increaseLeaderCalled := false
	increaseValidatorCalled := false
	marshalizer := &mock.MarshalizerStub{}
	shardHeaderStorage := &mock.StorerStub{}
	metaHeaderStorage := &mock.StorerStub{}
	adapter.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
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

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes: nil,
		Marshalizer: marshalizer,
		ShardHeaderStorage: shardHeaderStorage,
		MetaHeaderStorage: metaHeaderStorage,
		NodesCoordinator: &mock.NodesCoordinatorMock{
			ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
				return []sharding.Validator{&mock.ValidatorMock{}, &mock.ValidatorMock{}}, nil
			},
		},
		ShardCoordinator: shardCoordinatorMock,
		AdrConv: &mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
				return &mock.AddressMock{}, nil
			},
		},
		PeerAdapter: adapter,
	}
	validatorStatistics, _ := peer.NewValidatorStatisticsProcessor(arguments)

	header := getMetaHeaderHandler([]byte("header"))
	shardHeaderStorage.GetCalled = func(key []byte) (bytes []byte, e error) {
		return nil, nil
	}
	marshalizer.UnmarshalCalled = func(obj interface{}, buff []byte) error {
		switch v := obj.(type) {
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

func getMetaHeaderHandler(randSeed []byte) *block.MetaBlock {
	return &block.MetaBlock{
		Nonce: 1,
		PrevRandSeed: randSeed,
		PrevHash: randSeed,
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
