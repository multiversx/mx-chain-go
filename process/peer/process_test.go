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
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, peerProcessor)
	assert.Equal(t, process.ErrNilPeerAccountsAdapter, err)
}

func TestNewPeerProcessor_NilAddressConverterShouldErr(t *testing.T) {
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, err := peer.NewValidatorStatisticsProcessor(
		nil,
		getAccountsMock(),
		nil,
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, peerProcessor)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewPeerProcessor_NilNodesCoordinatorShouldErr(t *testing.T) {
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, err := peer.NewValidatorStatisticsProcessor(
		nil,
		getAccountsMock(),
		&mock.AddressConverterMock{},
		nil,
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, peerProcessor)
	assert.Equal(t, process.ErrNilNodesCoordinator, err)
}

func TestNewPeerProcessor_NilShardCoordinatorShouldErr(t *testing.T) {
	peerProcessor, err := peer.NewValidatorStatisticsProcessor(
		nil,
		getAccountsMock(),
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{},
		nil,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, peerProcessor)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewPeerProcessor_NilShardHeaderStorageShouldErr(t *testing.T) {
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, err := peer.NewValidatorStatisticsProcessor(
		nil,
		getAccountsMock(),
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		nil,
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, peerProcessor)
	assert.Equal(t, process.ErrNilShardHeaderStorage, err)
}

func TestNewPeerProcessor_NilMarshalizerShouldErr(t *testing.T) {
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, err := peer.NewValidatorStatisticsProcessor(
		nil,
		getAccountsMock(),
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
		nil,
	)

	assert.Nil(t, peerProcessor)
	assert.Equal(t, process.ErrNilShardHeaderStorage, err)
}

func TestNewPeerProcessor(t *testing.T) {
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, err := peer.NewValidatorStatisticsProcessor(
		nil,
		getAccountsMock(),
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.NotNil(t, peerProcessor)
	assert.Nil(t, err)
}

func TestPeerProcessor_LoadInitialStateErrOnInvalidNode(t *testing.T) {
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, _ := peer.NewValidatorStatisticsProcessor(
		nil,
		getAccountsMock(),
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	initialNodes := []*sharding.InitialNode{{PubKey:"", Address: ""}}
	err := peerProcessor.LoadInitialState(initialNodes)

	assert.Equal(t, process.ErrInvalidInitialNodesState, err)
}

func TestPeerProcessor_LoadInitialStateErrOnWrongAddressConverter(t *testing.T) {
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()

	addressErr := errors.New("hex address error")
	addressConverter := &mock.AddressConverterStub{
		CreateAddressFromHexCalled: func(hexAddress string) (container state.AddressContainer, e error) {
			return nil, addressErr
		},
	}
	peerProcessor, _ := peer.NewValidatorStatisticsProcessor(
		nil,
		getAccountsMock(),
		addressConverter,
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	initialNodes := []*sharding.InitialNode{{PubKey:"aaaa", Address: "aaaa"}}
	err := peerProcessor.LoadInitialState(initialNodes)

	assert.Equal(t, addressErr, err)
}

func TestPeerProcessor_LoadInitialStateErrOnGetAccountFail(t *testing.T) {
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
	_, err := peer.NewValidatorStatisticsProcessor(
		initialNodes,
		peerAdapters,
		addressConverter,
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, adapterError, err)
}

func TestPeerProcessor_LoadInitialStateGetAccountReturnsInvalid(t *testing.T) {
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
	_, err := peer.NewValidatorStatisticsProcessor(
		initialNodes,
		peerAdapter,
		addressConverter,
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrInvalidPeerAccount, err)
}

func TestPeerProcessor_LoadInitialStateSetAddressErrors(t *testing.T) {
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
	_, err := peer.NewValidatorStatisticsProcessor(
		initialNodes,
		peerAdapter,
		addressConverter,
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, saveAccountError, err)
}

func TestPeerProcessor_LoadInitialStateCommitErrors(t *testing.T) {
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
	_, err := peer.NewValidatorStatisticsProcessor(
		initialNodes,
		peerAdapter,
		addressConverter,
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, commitError, err)
}

func TestPeerProcessor_LoadInitialStateCommit(t *testing.T) {
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
	_, err := peer.NewValidatorStatisticsProcessor(
		initialNodes,
		peerAdapter,
		addressConverter,
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
}

func TestPeerProcess_IsNodeValidEmptyAddressShoudErr(t *testing.T) {
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, _ := peer.NewValidatorStatisticsProcessor(
		nil,
		getAccountsMock(),
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	isValid := peerProcessor.IsNodeValid(&sharding.InitialNode{Address: "", PubKey: "aaaaa"})
	assert.False(t, isValid)
}

func TestPeerProcess_IsNodeValidEmptyPubKeyShoudErr(t *testing.T) {
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, _ := peer.NewValidatorStatisticsProcessor(
		nil,
		getAccountsMock(),
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	isValid := peerProcessor.IsNodeValid(&sharding.InitialNode{Address: "aaaaa", PubKey: ""})
	assert.False(t, isValid)
}

func TestPeerProcess_IsNodeValid(t *testing.T) {
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, _ := peer.NewValidatorStatisticsProcessor(
		nil,
		getAccountsMock(),
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	isValid := peerProcessor.IsNodeValid(&sharding.InitialNode{Address: "aaaaa", PubKey: "aaaaaa"})
	assert.True(t, isValid)
}

func TestPeerProcess_UpdatePeerStateComputeValidatorErrShouldError(t *testing.T) {
	computeValidatorsErr := errors.New("compute validators error")
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, _ := peer.NewValidatorStatisticsProcessor(
		nil,
		getAccountsMock(),
		&mock.AddressConverterMock{},
		&mock.NodesCoordinatorMock{
			ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
				return nil, computeValidatorsErr
			},
		},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	header := getHeaderHandler([]byte("header"))
	prevHEader := getHeaderHandler([]byte("prevheader"))
	err := peerProcessor.UpdatePeerState(header, prevHEader)

	assert.Equal(t, computeValidatorsErr, err)
}

func TestPeerProcess_UpdatePeerStateCreateAddressFromPublicKeyBytesErr(t *testing.T) {
	createAddressErr := errors.New("create address error")
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, _ := peer.NewValidatorStatisticsProcessor(
		nil,
		getAccountsMock(),
		&mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
				return nil, createAddressErr
			},
		},
		&mock.NodesCoordinatorMock{
			ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
				return []sharding.Validator{&mock.ValidatorMock{}}, nil
			},
		},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	header := getHeaderHandler([]byte("header"))
	prevHEader := getHeaderHandler([]byte("prevheader"))
	err := peerProcessor.UpdatePeerState(header, prevHEader)

	assert.Equal(t, createAddressErr, err)
}

func TestPeerProcess_UpdatePeerStateGetExistingAccountErr(t *testing.T) {
	existingAccountErr := errors.New("existing account err")
	adapter := getAccountsMock()
	adapter.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return nil, existingAccountErr
	}
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, _ := peer.NewValidatorStatisticsProcessor(
		nil,
		adapter,
		&mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
				return &mock.AddressMock{}, nil
			},
		},
		&mock.NodesCoordinatorMock{
			ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
				return []sharding.Validator{&mock.ValidatorMock{}}, nil
			},
		},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	header := getHeaderHandler([]byte("header"))
	prevHEader := getHeaderHandler([]byte("prevheader"))
	err := peerProcessor.UpdatePeerState(header, prevHEader)

	assert.Equal(t, existingAccountErr, err)
}

func TestPeerProcess_UpdatePeerStateGetExistingAccountInvalidType(t *testing.T) {
	adapter := getAccountsMock()
	adapter.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &mock.AccountWrapMock{}, nil
	}
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, _ := peer.NewValidatorStatisticsProcessor(
		nil,
		adapter,
		&mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
				return &mock.AddressMock{}, nil
			},
		},
		&mock.NodesCoordinatorMock{
			ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
				return []sharding.Validator{&mock.ValidatorMock{}}, nil
			},
		},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	header := getHeaderHandler([]byte("header"))
	prevHEader := getHeaderHandler([]byte("prevheader"))
	err := peerProcessor.UpdatePeerState(header, prevHEader)

	assert.Equal(t, process.ErrInvalidPeerAccount, err)
}

func TestPeerProcess_UpdatePeerStateCallsSuccessForLeader(t *testing.T) {
	adapter := getAccountsMock()
	increaseLeaderCalled := false
	decreaseLeaderCalled := false
	increaseValidatorCalled := false
	decreaseValidatorCalled := false
	adapter.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &mock.PeerAccountHandlerMock{
			IncreaseLeaderSuccessRateWithJournalCalled: func() error {
				increaseLeaderCalled = true
				return nil
			},
			DecreaseLeaderSuccessRateWithJournalCalled: func() error {
				decreaseLeaderCalled = true
				return nil
			},
			IncreaseValidatorSuccessRateWithJournalCalled: func() error {
				increaseValidatorCalled = true
				return nil
			},
			DecreaseValidatorSuccessRateWithJournalCalled: func() error {
				decreaseValidatorCalled = true
				return nil
			},
		}, nil
	}
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, _ := peer.NewValidatorStatisticsProcessor(
		nil,
		adapter,
		&mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
				return &mock.AddressMock{}, nil
			},
		},
		&mock.NodesCoordinatorMock{
			ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
				return []sharding.Validator{&mock.ValidatorMock{}, &mock.ValidatorMock{}}, nil
			},
		},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	header := getHeaderHandler([]byte("header"))
	prevHEader := getHeaderHandler([]byte("prevheader"))
	// {0} - nobody signed
	// {1} - leader signed, validator did not
	// {2} - validator signed, leader didn't
	// {3} - everybody signed
	bitmaps := [][]byte{{0}, {1}, {2}, {3}}
	for _, bitmap := range bitmaps {
		increaseLeaderCalled = false
		decreaseLeaderCalled = false
		increaseValidatorCalled = false
		decreaseValidatorCalled = false

		increaseLeaderShouldBeCalled := false
		decreaseLeaderShouldBeCalled := false
		increaseValidatorShouldBeCalled := false
		decreaseValidatorShouldBeCalled := false

		prevHEader.GetPubKeysBitmapCalled = func() []byte {
			return bitmap
		}
		err := peerProcessor.UpdatePeerState(header, prevHEader)

		leaderSigned := validatorSigned(0, bitmap)
		validatorSigned := validatorSigned(1, bitmap)

		if leaderSigned {
			increaseLeaderShouldBeCalled = true
		} else {
			decreaseLeaderShouldBeCalled = true
		}

		if validatorSigned {
			increaseValidatorShouldBeCalled = true
		} else {
			decreaseValidatorShouldBeCalled = true
		}

		assert.Nil(t, err)
		assert.Equal(t, increaseLeaderCalled, increaseLeaderShouldBeCalled)
		assert.Equal(t, decreaseLeaderCalled, decreaseLeaderShouldBeCalled)
		assert.Equal(t, increaseValidatorCalled, increaseValidatorShouldBeCalled)
		assert.Equal(t, decreaseValidatorCalled, decreaseValidatorShouldBeCalled)
	}
}

func TestPeerProcess_RevertPeerStateCallsSuccessForLeader(t *testing.T) {
	adapter := getAccountsMock()
	increaseLeaderCalled := false
	decreaseLeaderCalled := false
	increaseValidatorCalled := false
	decreaseValidatorCalled := false
	adapter.GetExistingAccountCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &mock.PeerAccountHandlerMock{
			IncreaseLeaderSuccessRateWithJournalCalled: func() error {
				increaseLeaderCalled = true
				return nil
			},
			DecreaseLeaderSuccessRateWithJournalCalled: func() error {
				decreaseLeaderCalled = true
				return nil
			},
			IncreaseValidatorSuccessRateWithJournalCalled: func() error {
				increaseValidatorCalled = true
				return nil
			},
			DecreaseValidatorSuccessRateWithJournalCalled: func() error {
				decreaseValidatorCalled = true
				return nil
			},
		}, nil
	}
	shardCoordinatorMock := mock.NewOneShardCoordinatorMock()
	peerProcessor, _ := peer.NewValidatorStatisticsProcessor(
		nil,
		adapter,
		&mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
				return &mock.AddressMock{}, nil
			},
		},
		&mock.NodesCoordinatorMock{
			ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32) (validatorsGroup []sharding.Validator, err error) {
				return []sharding.Validator{&mock.ValidatorMock{}, &mock.ValidatorMock{}}, nil
			},
		},
		shardCoordinatorMock,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	header := getHeaderHandler([]byte("header"))
	prevHEader := getHeaderHandler([]byte("prevheader"))
	// {0} - nobody signed
	// {1} - leader signed, validator did not
	// {2} - validator signed, leader didn't
	// {3} - everybody signed
	bitmaps := [][]byte{{0}, {1}, {2}, {3}}
	for _, bitmap := range bitmaps {
		increaseLeaderCalled = false
		decreaseLeaderCalled = false
		increaseValidatorCalled = false
		decreaseValidatorCalled = false

		increaseLeaderShouldBeCalled := false
		decreaseLeaderShouldBeCalled := false
		increaseValidatorShouldBeCalled := false
		decreaseValidatorShouldBeCalled := false

		prevHEader.GetPubKeysBitmapCalled = func() []byte {
			return bitmap
		}
		err := peerProcessor.RevertPeerState(header, prevHEader)

		leaderSigned := validatorSigned(0, bitmap)
		validatorSigned := validatorSigned(1, bitmap)

		if leaderSigned {
			decreaseLeaderShouldBeCalled = true
		} else {
			increaseLeaderShouldBeCalled = true
		}

		if validatorSigned {
			decreaseValidatorShouldBeCalled = true
		} else {
			increaseValidatorShouldBeCalled = true
		}

		assert.Nil(t, err)
		assert.Equal(t, increaseLeaderCalled, increaseLeaderShouldBeCalled)
		assert.Equal(t, decreaseLeaderCalled, decreaseLeaderShouldBeCalled)
		assert.Equal(t, increaseValidatorCalled, increaseValidatorShouldBeCalled)
		assert.Equal(t, decreaseValidatorCalled, decreaseValidatorShouldBeCalled)
	}
}

func validatorSigned(i int, bitmap []byte) bool {
	return (bitmap[i/8] & (1 << (uint16(i) % 8))) != 0
}

func getHeaderHandler(randSeed []byte) *mock.HeaderHandlerStub {
	return &mock.HeaderHandlerStub{
		GetPrevRandSeedCalled: func() []byte {
			return randSeed
		},
		GetPubKeysBitmapCalled: func() []byte {
			return randSeed
		},
		GetShardIDCalled: func() uint32 {
			return 0
		},
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
