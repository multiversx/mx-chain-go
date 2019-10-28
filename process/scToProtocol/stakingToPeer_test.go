package scToProtocol

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/vm/factory"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func createMockArgumentsNewStakingToPeer() ArgStakingToPeer {
	return ArgStakingToPeer{
		AdrConv:      &mock.AddressConverterMock{},
		Hasher:       &mock.HasherMock{},
		Marshalizer:  &mock.MarshalizerStub{},
		PeerState:    &mock.AccountsStub{},
		BaseState:    &mock.AccountsStub{},
		ArgParser:    &mock.ArgumentParserMock{},
		CurrTxs:      &mock.TxForCurrentBlockMock{},
		ScDataGetter: &mock.ScDataGetterMock{},
	}
}

func createBlockBody() block.Body {
	return block.Body{
		{
			TxHashes:        [][]byte{[]byte("hash1"), []byte("hash2")},
			ReceiverShardID: 0,
			SenderShardID:   0,
			Type:            block.SmartContractResultBlock,
		},
	}
}

func TestNewStakingToPeerNilAddrConverterShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.AdrConv = nil

	_, err := NewStakingToPeer(arguments)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewStakingToPeerNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.Hasher = nil

	_, err := NewStakingToPeer(arguments)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewStakingToPeerNilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.Marshalizer = nil

	_, err := NewStakingToPeer(arguments)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewStakingToPeerNilPeerAccountAdapterShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.PeerState = nil

	_, err := NewStakingToPeer(arguments)
	assert.Equal(t, process.ErrNilPeerAccountsAdapter, err)
}

func TestNewStakingToPeerNilBaseAccountAdapterShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.BaseState = nil

	_, err := NewStakingToPeer(arguments)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewStakingToPeerNilArgumentParserShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.ArgParser = nil

	_, err := NewStakingToPeer(arguments)
	assert.Equal(t, process.ErrNilArgumentParser, err)
}

func TestNewStakingToPeerNilCurrentBlockHeaderShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.CurrTxs = nil

	_, err := NewStakingToPeer(arguments)
	assert.Equal(t, process.ErrNilTxForCurrentBlockHandler, err)
}

func TestNewStakingToPeerNilScDataGetterShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.ScDataGetter = nil

	_, err := NewStakingToPeer(arguments)
	assert.Equal(t, process.ErrNilSCDataGetter, err)
}

func TestNewStakingToPeer_ShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()

	stakingToPeer, err := NewStakingToPeer(arguments)
	assert.NotNil(t, stakingToPeer)
	assert.Nil(t, err)
}

func TestStakingToPeer_UpdateProtocolCannotGetTxShouldErr(t *testing.T) {
	t.Parallel()

	testError := errors.New("error")
	currTx := &mock.TxForCurrentBlockMock{}
	currTx.GetTxCalled = func(txHash []byte) (handler data.TransactionHandler, e error) {
		return nil, testError
	}

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.CurrTxs = currTx
	blockBody := createBlockBody()
	stakingToPeer, _ := NewStakingToPeer(arguments)

	err := stakingToPeer.UpdateProtocol(blockBody, 0)
	assert.Equal(t, testError, err)
}

func TestStakingToPeer_UpdateProtocolWrongTransactionTypeShouldErr(t *testing.T) {
	t.Parallel()

	currTx := &mock.TxForCurrentBlockMock{}
	currTx.GetTxCalled = func(txHash []byte) (handler data.TransactionHandler, e error) {
		return &transaction.Transaction{
			RcvAddr: factory.StakingSCAddress,
		}, nil
	}

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.CurrTxs = currTx
	blockBody := createBlockBody()
	stakingToPeer, _ := NewStakingToPeer(arguments)

	err := stakingToPeer.UpdateProtocol(blockBody, 0)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestStakingToPeer_UpdateProtocolCannotGetStorageUpdatesShouldErr(t *testing.T) {
	t.Parallel()

	testError := errors.New("error")
	currTx := &mock.TxForCurrentBlockMock{}
	currTx.GetTxCalled = func(txHash []byte) (handler data.TransactionHandler, e error) {
		return &smartContractResult.SmartContractResult{
			RcvAddr: factory.StakingSCAddress,
		}, nil
	}

	argParser := &mock.ArgumentParserMock{}
	argParser.GetStorageUpdatesCalled = func(data string) (updates []*vmcommon.StorageUpdate, e error) {
		return nil, testError
	}

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.ArgParser = argParser
	arguments.CurrTxs = currTx
	blockBody := createBlockBody()
	stakingToPeer, _ := NewStakingToPeer(arguments)

	err := stakingToPeer.UpdateProtocol(blockBody, 0)
	assert.Equal(t, testError, err)
}

func TestStakingToPeer_UpdateProtocolWrongAccountShouldErr(t *testing.T) {
	t.Parallel()

	currTx := &mock.TxForCurrentBlockMock{}
	currTx.GetTxCalled = func(txHash []byte) (handler data.TransactionHandler, e error) {
		return &smartContractResult.SmartContractResult{
			RcvAddr: factory.StakingSCAddress,
		}, nil
	}

	argParser := &mock.ArgumentParserMock{}
	argParser.GetStorageUpdatesCalled = func(data string) (updates []*vmcommon.StorageUpdate, e error) {
		return []*vmcommon.StorageUpdate{
			{Offset: []byte("off1"), Data: []byte("data1")},
		}, nil
	}

	peerState := &mock.AccountsStub{}
	peerState.GetAccountWithJournalCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &mock.AccountWrapMock{}, nil
	}

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.ArgParser = argParser
	arguments.CurrTxs = currTx
	arguments.PeerState = peerState
	blockBody := createBlockBody()
	stakingToPeer, _ := NewStakingToPeer(arguments)

	err := stakingToPeer.UpdateProtocol(blockBody, 0)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestStakingToPeer_UpdateProtocol(t *testing.T) {
	t.Parallel()

	currTx := &mock.TxForCurrentBlockMock{}
	currTx.GetTxCalled = func(txHash []byte) (handler data.TransactionHandler, e error) {
		return &smartContractResult.SmartContractResult{
			RcvAddr: factory.StakingSCAddress,
		}, nil
	}

	argParser := &mock.ArgumentParserMock{}
	argParser.GetStorageUpdatesCalled = func(data string) (updates []*vmcommon.StorageUpdate, e error) {
		return []*vmcommon.StorageUpdate{
			{Offset: []byte("off1"), Data: []byte("data1")},
		}, nil
	}

	peerState := &mock.AccountsStub{}
	peerState.GetAccountWithJournalCalled = func(addressContainer state.AddressContainer) (handler state.AccountHandler, e error) {
		return &state.PeerAccount{
			Address:      []byte("addr"),
			BLSPublicKey: []byte("BlsAddr"),
			Stake:        big.NewInt(100),
		}, nil
	}

	marshalizer := &mock.MarshalizerStub{}
	marshalizer.MarshalCalled = func(obj interface{}) (bytes []byte, e error) {
		return []byte("mashalizedData"), nil
	}

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.ArgParser = argParser
	arguments.CurrTxs = currTx
	arguments.PeerState = peerState
	arguments.Marshalizer = marshalizer
	blockBody := createBlockBody()
	stakingToPeer, _ := NewStakingToPeer(arguments)

	err := stakingToPeer.UpdateProtocol(blockBody, 0)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}
