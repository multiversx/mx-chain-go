package peer_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/peer"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNewPeerProcessor_NilPeerAdaptersShouldErr(t *testing.T) {
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, err := peer.NewValidatorStatisticsProcessor(
		nil,
		nil,
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
	)

	assert.Nil(t, peerProcessor)
	assert.Equal(t, process.ErrNilPeerAccountsAdapter, err)
}

func TestNewPeerProcessor_NilAddressConverterShouldErr(t *testing.T) {
	peerAdapters := make(peer.ShardedPeerAdapters, 0)
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, err := peer.NewValidatorStatisticsProcessor(
		nil,
		peerAdapters,
		nil,
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
	)

	assert.Nil(t, peerProcessor)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewPeerProcessor_NilNodesCoordinatorShouldErr(t *testing.T) {
	peerAdapters := make(peer.ShardedPeerAdapters, 0)
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, err := peer.NewValidatorStatisticsProcessor(
		nil,
		peerAdapters,
		&mock.AddressConverterMock{},
		nil,
		shardCoordinatorMock,
		&mock.StorerStub{},
	)

	assert.Nil(t, peerProcessor)
	assert.Equal(t, process.ErrNilNodesCoordinator, err)
}

func TestNewPeerProcessor_ShardCoordinatorShouldErr(t *testing.T) {
	peerAdapters := make(peer.ShardedPeerAdapters, 0)
	peerProcessor, err := peer.NewValidatorStatisticsProcessor(
		nil,
		peerAdapters,
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{},
		nil,
		&mock.StorerStub{},
	)

	assert.Nil(t, peerProcessor)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewPeerProcessor_ShardHeaderStorageShouldErr(t *testing.T) {
	peerAdapters := make(peer.ShardedPeerAdapters, 0)
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, err := peer.NewValidatorStatisticsProcessor(
		nil,
		peerAdapters,
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		nil,
	)

	assert.Nil(t, peerProcessor)
	assert.Equal(t, process.ErrNilShardHeaderStorage, err)
}

func TestNewPeerProcessor(t *testing.T) {
	peerAdapters := make(peer.ShardedPeerAdapters, 0)
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, err := peer.NewValidatorStatisticsProcessor(
		nil,
		peerAdapters,
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
	)

	assert.NotNil(t, peerProcessor)
	assert.Nil(t, err)
}

func TestPeerProcessor_LoadInitialStateErrOnInvalidNode(t *testing.T) {
	peerAdapters := make(peer.ShardedPeerAdapters, 0)
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, _ := peer.NewValidatorStatisticsProcessor(
		nil,
		peerAdapters,
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
	)

	initialNodes := []*sharding.InitialNode{{PubKey:"", Address: ""}}
	err := peerProcessor.LoadInitialState(initialNodes)

	assert.Equal(t, process.ErrInvalidInitialNodesState, err)
}

func TestPeerProcessor_LoadInitialStateErrOnWrongAddressConverter(t *testing.T) {
	peerAdapters := make(peer.ShardedPeerAdapters, 0)
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	addressErr := errors.New("hex address error")
	addressConverter := &mock.AddressConverterStub{
		CreateAddressFromHexCalled: func(hexAddress string) (container state.AddressContainer, e error) {
			return nil, addressErr
		},
	}
	peerProcessor, _ := peer.NewValidatorStatisticsProcessor(
		nil,
		peerAdapters,
		addressConverter,
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
	)

	initialNodes := []*sharding.InitialNode{{PubKey:"aaaa", Address: "aaaa"}}
	err := peerProcessor.LoadInitialState(initialNodes)

	assert.Equal(t, addressErr, err)
}

func TestPeerProcessor_LoadInitialStateErrOnInvalidAssignedShard(t *testing.T) {
	peerAdapters := make(peer.ShardedPeerAdapters, 0)
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	addressConverter := &mock.AddressConverterStub{
		CreateAddressFromHexCalled: func(hexAddress string) (container state.AddressContainer, e error) {
			return &mock.AddressMock{}, nil
		},
	}
	initialNodes := []*sharding.InitialNode{{PubKey:"aaaa", Address: "aaaa",}}
	_, err := peer.NewValidatorStatisticsProcessor(
		initialNodes,
		peerAdapters,
		addressConverter,
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
	)

	assert.Equal(t, process.ErrNilPeerAccountsAdapter, err)
}

func TestPeerProcessor_LoadInitialStateErrOnGetAccountFail(t *testing.T) {
	adapterError := errors.New("account error")
	peerAdapters := make(peer.ShardedPeerAdapters, 0)
	peerAdapters[0] = &mock.AccountsStub{
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
	_, err := peer.NewValidatorStatisticsProcessor(
		initialNodes,
		peerAdapters,
		addressConverter,
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
	)

	assert.Equal(t, adapterError, err)
}

func TestPeerProcessor_LoadInitialStateGetAccountReturnsInvalid(t *testing.T) {
	peerAdapters := make(peer.ShardedPeerAdapters, 0)
	peerAdapters[0] = &mock.AccountsStub{
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
	_, err := peer.NewValidatorStatisticsProcessor(
		initialNodes,
		peerAdapters,
		addressConverter,
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
	)

	assert.Equal(t, process.ErrInvalidPeerAccount, err)
}

func TestPeerProcessor_LoadInitialStateSetAddressErrors(t *testing.T) {
	saveAccountError := errors.New("save account error")
	peerAdapters := make(peer.ShardedPeerAdapters, 0)
	peerAccount, _ := state.NewPeerAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {

		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return saveAccountError
		},
	})
	peerAdapters[0] = &mock.AccountsStub{
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
	_, err := peer.NewValidatorStatisticsProcessor(
		initialNodes,
		peerAdapters,
		addressConverter,
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
	)

	assert.Equal(t, saveAccountError, err)
}

func TestPeerProcessor_LoadInitialStateCommitErrors(t *testing.T) {
	commitError := errors.New("commit error")
	peerAdapters := make(peer.ShardedPeerAdapters, 0)
	peerAccount, _ := state.NewPeerAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {

		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return nil
		},
	})
	peerAdapters[0] = &mock.AccountsStub{
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
	_, err := peer.NewValidatorStatisticsProcessor(
		initialNodes,
		peerAdapters,
		addressConverter,
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
	)

	assert.Equal(t, commitError, err)
}

func TestPeerProcessor_LoadInitialStateCommit(t *testing.T) {
	peerAdapters := make(peer.ShardedPeerAdapters, 0)
	peerAccount, _ := state.NewPeerAccount(&mock.AddressMock{}, &mock.AccountTrackerStub{
		JournalizeCalled: func(entry state.JournalEntry) {

		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return nil
		},
	})
	peerAdapters[0] = &mock.AccountsStub{
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
	_, err := peer.NewValidatorStatisticsProcessor(
		initialNodes,
		peerAdapters,
		addressConverter,
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
	)

	assert.Nil(t, err)
}

func TestPeerProcess_IsNodeValidEmptyAddressShoudErr(t *testing.T) {
	peerAdapters := make(peer.ShardedPeerAdapters, 0)
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, _ := peer.NewValidatorStatisticsProcessor(
		nil,
		peerAdapters,
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
	)

	isValid := peerProcessor.IsNodeValid(&sharding.InitialNode{Address: "", PubKey: "aaaaa"})
	assert.False(t, isValid)
}

func TestPeerProcess_IsNodeValidEmptyPubKeyShoudErr(t *testing.T) {
	peerAdapters := make(peer.ShardedPeerAdapters, 0)
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, _ := peer.NewValidatorStatisticsProcessor(
		nil,
		peerAdapters,
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
	)

	isValid := peerProcessor.IsNodeValid(&sharding.InitialNode{Address: "aaaaa", PubKey: ""})
	assert.False(t, isValid)
}

func TestPeerProcess_IsNodeValid(t *testing.T) {
	peerAdapters := make(peer.ShardedPeerAdapters, 0)
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, _ := peer.NewValidatorStatisticsProcessor(
		nil,
		peerAdapters,
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
	)

	isValid := peerProcessor.IsNodeValid(&sharding.InitialNode{Address: "aaaaa", PubKey: "aaaaaa"})
	assert.True(t, isValid)
}

func TestPeerProcess_UpdatePeerStateComputeValidatorErrShouldError(t *testing.T) {
	computeValidatorsErr := errors.New("compute validators error")
	peerAdapters := make(peer.ShardedPeerAdapters, 0)
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, _ := peer.NewValidatorStatisticsProcessor(
		nil,
		peerAdapters,
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{
			ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
				return nil, computeValidatorsErr
			},
		},
		shardCoordinatorMock,
		&mock.StorerStub{},
	)

	header := getHeaderHandler([]byte("header"))
	prevHEader := getHeaderHandler([]byte("prevheader"))
	err := peerProcessor.UpdatePeerState(header, prevHEader)

	assert.Equal(t, computeValidatorsErr, err)
}

func getHeaderHandler(randSeed []byte) *mock.HeaderHandlerStub {
	return &mock.HeaderHandlerStub{
		GetPrevRandSeedCalled: func() []byte {
			return randSeed
		},
	}
}
