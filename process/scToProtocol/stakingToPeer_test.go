package scToProtocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func createMockArgumentsNewStakingToPeer() ArgStakingToPeer {
	return ArgStakingToPeer{
		PubkeyConv:    mock.NewPubkeyConverterMock(32),
		Hasher:        &mock.HasherMock{},
		Marshalizer:   &mock.MarshalizerStub{},
		PeerState:     &mock.AccountsStub{},
		BaseState:     &mock.AccountsStub{},
		ArgParser:     &mock.ArgumentParserMock{},
		CurrTxs:       &mock.TxForCurrentBlockStub{},
		ScQuery:       &mock.ScQueryStub{},
		RatingsData:   &mock.RatingsInfoMock{},
		EpochNotifier: &mock.EpochNotifierStub{},
	}
}

func createBlockBody() *block.Body {
	return &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes:        [][]byte{[]byte("hash1"), []byte("hash2")},
				ReceiverShardID: core.MetachainShardId,
				SenderShardID:   core.MetachainShardId,
				Type:            block.SmartContractResultBlock,
			},
		},
	}
}

func TestNewStakingToPeerNilAddrConverterShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.PubkeyConv = nil

	stp, err := NewStakingToPeer(arguments)
	assert.Nil(t, stp)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewStakingToPeerNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.Hasher = nil

	stp, err := NewStakingToPeer(arguments)
	assert.Nil(t, stp)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewStakingToPeerNilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.Marshalizer = nil

	stp, err := NewStakingToPeer(arguments)
	assert.Nil(t, stp)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewStakingToPeerNilPeerAccountAdapterShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.PeerState = nil

	stp, err := NewStakingToPeer(arguments)
	assert.Nil(t, stp)
	assert.Equal(t, process.ErrNilPeerAccountsAdapter, err)
}

func TestNewStakingToPeerNilBaseAccountAdapterShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.BaseState = nil

	stp, err := NewStakingToPeer(arguments)
	assert.Nil(t, stp)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewStakingToPeerNilArgumentParserShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.ArgParser = nil

	stp, err := NewStakingToPeer(arguments)
	assert.Nil(t, stp)
	assert.Equal(t, process.ErrNilArgumentParser, err)
}

func TestNewStakingToPeerNilCurrentBlockHeaderShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.CurrTxs = nil

	stp, err := NewStakingToPeer(arguments)
	assert.Nil(t, stp)
	assert.Equal(t, process.ErrNilTxForCurrentBlockHandler, err)
}

func TestNewStakingToPeerNilScDataGetterShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.ScQuery = nil

	stp, err := NewStakingToPeer(arguments)
	assert.Nil(t, stp)
	assert.Equal(t, process.ErrNilSCDataGetter, err)
}

func TestNewStakingToPeer_ShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsNewStakingToPeer()

	stp, err := NewStakingToPeer(arguments)
	assert.NotNil(t, stp)
	assert.Nil(t, err)
}

func TestStakingToPeer_UpdateProtocolCannotGetTxShouldErr(t *testing.T) {
	t.Parallel()

	called := false
	testError := errors.New("error")
	currTx := &mock.TxForCurrentBlockStub{}
	currTx.GetTxCalled = func(txHash []byte) (handler data.TransactionHandler, e error) {
		called = true
		return nil, testError
	}

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.CurrTxs = currTx
	stp, _ := NewStakingToPeer(arguments)

	blockBody := createBlockBody()
	err := stp.UpdateProtocol(blockBody, 0)
	assert.Nil(t, err)
	assert.True(t, called)
}

func TestStakingToPeer_UpdateProtocolWrongTransactionTypeShouldErr(t *testing.T) {
	t.Parallel()

	currTx := &mock.TxForCurrentBlockStub{}
	currTx.GetTxCalled = func(txHash []byte) (handler data.TransactionHandler, e error) {
		return &transaction.Transaction{
			RcvAddr: vm.StakingSCAddress,
		}, nil
	}

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.CurrTxs = currTx
	stp, _ := NewStakingToPeer(arguments)

	blockBody := createBlockBody()
	err := stp.UpdateProtocol(blockBody, 0)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestStakingToPeer_UpdateProtocolCannotGetStorageUpdatesShouldErr(t *testing.T) {
	t.Parallel()

	testError := errors.New("error")
	currTx := &mock.TxForCurrentBlockStub{}
	currTx.GetTxCalled = func(txHash []byte) (handler data.TransactionHandler, e error) {
		return &smartContractResult.SmartContractResult{
			RcvAddr: vm.StakingSCAddress,
		}, nil
	}

	argParser := &mock.ArgumentParserMock{}
	argParser.GetStorageUpdatesCalled = func(data string) (updates []*vmcommon.StorageUpdate, e error) {
		return nil, testError
	}

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.ArgParser = argParser
	arguments.CurrTxs = currTx
	stp, _ := NewStakingToPeer(arguments)

	blockBody := createBlockBody()
	err := stp.UpdateProtocol(blockBody, 0)
	assert.Nil(t, err)
}

func TestStakingToPeer_UpdateProtocolRemoveAccountShouldReturnNil(t *testing.T) {
	t.Parallel()

	currTx := &mock.TxForCurrentBlockStub{}
	currTx.GetTxCalled = func(txHash []byte) (handler data.TransactionHandler, e error) {
		return &smartContractResult.SmartContractResult{
			RcvAddr: vm.StakingSCAddress,
		}, nil
	}

	argParser := &mock.ArgumentParserMock{}
	argParser.GetStorageUpdatesCalled = func(data string) (updates []*vmcommon.StorageUpdate, e error) {
		return []*vmcommon.StorageUpdate{
			{Offset: []byte("aabbcc"), Data: []byte("data1")},
		}, nil
	}

	peerState := &mock.AccountsStub{}
	peerState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		peerAcc, _ := state.NewPeerAccount(address)
		_ = peerAcc.SetRewardAddress([]byte("addr"))
		_ = peerAcc.SetBLSPublicKey([]byte("BlsAddr"))

		return peerAcc, nil
	}
	peerState.RemoveAccountCalled = func(address []byte) error {
		return nil
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
	stp, _ := NewStakingToPeer(arguments)

	blockBody := createBlockBody()
	err := stp.UpdateProtocol(blockBody, 0)
	assert.Nil(t, err)
}

func TestStakingToPeer_UpdateProtocolCannotSetRewardAddressShouldErr(t *testing.T) {
	t.Parallel()

	currTx := &mock.TxForCurrentBlockStub{}
	currTx.GetTxCalled = func(txHash []byte) (handler data.TransactionHandler, e error) {
		return &smartContractResult.SmartContractResult{
			RcvAddr: vm.StakingSCAddress,
		}, nil
	}

	arguments := createMockArgumentsNewStakingToPeer()
	offset := make([]byte, 0, arguments.PubkeyConv.Len())
	for i := 0; i < arguments.PubkeyConv.Len(); i++ {
		offset = append(offset, 99)
	}

	argParser := &mock.ArgumentParserMock{}
	argParser.GetStorageUpdatesCalled = func(data string) (updates []*vmcommon.StorageUpdate, e error) {
		return []*vmcommon.StorageUpdate{
			{Offset: offset, Data: []byte("data1")},
		}, nil
	}

	peerState := &mock.AccountsStub{}
	peerState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		peerAcc, _ := state.NewPeerAccount(address)
		_ = peerAcc.SetRewardAddress([]byte("key"))

		return peerAcc, nil
	}

	stakingData := systemSmartContracts.StakedDataV2{
		StakeValue: big.NewInt(100),
	}
	marshalizer := &mock.MarshalizerMock{}

	scDataGetter := &mock.ScQueryStub{}
	scDataGetter.ExecuteQueryCalled = func(query *process.SCQuery) (output *vmcommon.VMOutput, e error) {
		retData, _ := json.Marshal(&stakingData)
		return &vmcommon.VMOutput{ReturnData: [][]byte{retData}}, nil
	}

	arguments.ArgParser = argParser
	arguments.CurrTxs = currTx
	arguments.PeerState = peerState
	arguments.Marshalizer = marshalizer
	arguments.ScQuery = scDataGetter
	stp, _ := NewStakingToPeer(arguments)

	blockBody := createBlockBody()
	err := stp.UpdateProtocol(blockBody, 0)
	assert.Equal(t, state.ErrEmptyAddress, err)
}

func TestStakingToPeer_UpdateProtocolCannotSaveAccountShouldErr(t *testing.T) {
	t.Parallel()

	testError := errors.New("error")
	address := "address"
	currTx := &mock.TxForCurrentBlockStub{}
	currTx.GetTxCalled = func(txHash []byte) (handler data.TransactionHandler, e error) {
		return &smartContractResult.SmartContractResult{
			RcvAddr: vm.StakingSCAddress,
		}, nil
	}

	arguments := createMockArgumentsNewStakingToPeer()
	offset := make([]byte, 0, arguments.PubkeyConv.Len())
	for i := 0; i < arguments.PubkeyConv.Len(); i++ {
		offset = append(offset, 99)
	}

	argParser := &mock.ArgumentParserMock{}
	argParser.GetStorageUpdatesCalled = func(data string) (updates []*vmcommon.StorageUpdate, e error) {
		return []*vmcommon.StorageUpdate{
			{Offset: offset, Data: []byte("data1")},
		}, nil
	}

	peerState := &mock.AccountsStub{
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return testError
		},
	}

	peerState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		peerAccount, _ := state.NewPeerAccount(address)
		peerAccount.RewardAddress = address
		return peerAccount, nil
	}

	stakingData := systemSmartContracts.StakedDataV2{
		StakeValue:    big.NewInt(100),
		RewardAddress: []byte(address),
	}
	marshalizer := &mock.MarshalizerMock{}

	scDataGetter := &mock.ScQueryStub{}
	scDataGetter.ExecuteQueryCalled = func(query *process.SCQuery) (output *vmcommon.VMOutput, e error) {
		retData, _ := json.Marshal(&stakingData)
		return &vmcommon.VMOutput{ReturnData: [][]byte{retData}}, nil
	}

	arguments.ArgParser = argParser
	arguments.CurrTxs = currTx
	arguments.PeerState = peerState
	arguments.Marshalizer = marshalizer
	arguments.ScQuery = scDataGetter
	stp, _ := NewStakingToPeer(arguments)

	blockBody := createBlockBody()
	err := stp.UpdateProtocol(blockBody, 0)
	assert.Equal(t, testError, err)
}

func TestStakingToPeer_UpdateProtocolCannotSaveAccountNonceShouldErr(t *testing.T) {
	t.Parallel()

	testError := errors.New("error")
	address := "address"
	currTx := &mock.TxForCurrentBlockStub{}
	currTx.GetTxCalled = func(txHash []byte) (handler data.TransactionHandler, e error) {
		return &smartContractResult.SmartContractResult{
			RcvAddr: vm.StakingSCAddress,
		}, nil
	}

	arguments := createMockArgumentsNewStakingToPeer()
	offset := make([]byte, 0, arguments.PubkeyConv.Len())
	for i := 0; i < arguments.PubkeyConv.Len(); i++ {
		offset = append(offset, 99)
	}

	argParser := &mock.ArgumentParserMock{}
	argParser.GetStorageUpdatesCalled = func(data string) (updates []*vmcommon.StorageUpdate, e error) {
		return []*vmcommon.StorageUpdate{
			{Offset: offset, Data: []byte("data1")},
		}, nil
	}

	peerState := &mock.AccountsStub{
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return testError
		},
	}
	peerState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		peerAccount, _ := state.NewPeerAccount([]byte("1234"))
		peerAccount.BLSPublicKey = address
		peerAccount.Nonce = 1
		return peerAccount, nil
	}

	stakingData := systemSmartContracts.StakedDataV2{
		StakeValue:    big.NewInt(100),
		RewardAddress: []byte(address),
	}
	marshalizer := &mock.MarshalizerMock{}

	scDataGetter := &mock.ScQueryStub{}
	scDataGetter.ExecuteQueryCalled = func(query *process.SCQuery) (output *vmcommon.VMOutput, e error) {
		retData, _ := json.Marshal(&stakingData)
		return &vmcommon.VMOutput{ReturnData: [][]byte{retData}}, nil
	}

	arguments.ArgParser = argParser
	arguments.CurrTxs = currTx
	arguments.PeerState = peerState
	arguments.Marshalizer = marshalizer
	arguments.ScQuery = scDataGetter
	stp, _ := NewStakingToPeer(arguments)

	blockBody := createBlockBody()
	err := stp.UpdateProtocol(blockBody, 0)
	assert.Equal(t, testError, err)
}

func TestStakingToPeer_UpdateProtocol(t *testing.T) {
	t.Parallel()

	address := "address"
	currTx := &mock.TxForCurrentBlockStub{}
	currTx.GetTxCalled = func(txHash []byte) (handler data.TransactionHandler, e error) {
		return &smartContractResult.SmartContractResult{
			RcvAddr: vm.StakingSCAddress,
		}, nil
	}

	arguments := createMockArgumentsNewStakingToPeer()
	offset := make([]byte, 0, arguments.PubkeyConv.Len())
	for i := 0; i < arguments.PubkeyConv.Len(); i++ {
		offset = append(offset, 99)
	}

	argParser := &mock.ArgumentParserMock{}
	argParser.GetStorageUpdatesCalled = func(data string) (updates []*vmcommon.StorageUpdate, e error) {
		return []*vmcommon.StorageUpdate{
			{Offset: offset, Data: []byte("data1")},
		}, nil
	}

	peerState := &mock.AccountsStub{
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return nil
		},
	}
	peerState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		peerAccount, _ := state.NewPeerAccount([]byte("1234"))
		peerAccount.BLSPublicKey = address
		peerAccount.Nonce = 1
		return peerAccount, nil
	}

	stakingData := systemSmartContracts.StakedDataV2{
		StakeValue:    big.NewInt(100),
		RewardAddress: []byte(address),
	}
	marshalizer := &mock.MarshalizerMock{}

	scDataGetter := &mock.ScQueryStub{}
	scDataGetter.ExecuteQueryCalled = func(query *process.SCQuery) (output *vmcommon.VMOutput, e error) {
		retData, _ := json.Marshal(&stakingData)
		return &vmcommon.VMOutput{ReturnData: [][]byte{retData}}, nil
	}

	arguments.ArgParser = argParser
	arguments.CurrTxs = currTx
	arguments.PeerState = peerState
	arguments.Marshalizer = marshalizer
	arguments.ScQuery = scDataGetter
	stp, _ := NewStakingToPeer(arguments)

	blockBody := createBlockBody()
	err := stp.UpdateProtocol(blockBody, 0)
	assert.Nil(t, err)
}

func TestStakingToPeer_UpdateProtocolCannotSaveUnStakedNonceShouldErr(t *testing.T) {
	t.Parallel()

	testError := errors.New("error")
	address := "address"
	currTx := &mock.TxForCurrentBlockStub{}
	currTx.GetTxCalled = func(txHash []byte) (handler data.TransactionHandler, e error) {
		return &smartContractResult.SmartContractResult{
			RcvAddr: vm.StakingSCAddress,
		}, nil
	}

	arguments := createMockArgumentsNewStakingToPeer()
	offset := make([]byte, 0, arguments.PubkeyConv.Len())
	for i := 0; i < arguments.PubkeyConv.Len(); i++ {
		offset = append(offset, 99)
	}

	argParser := &mock.ArgumentParserMock{}
	argParser.GetStorageUpdatesCalled = func(data string) (updates []*vmcommon.StorageUpdate, e error) {
		return []*vmcommon.StorageUpdate{
			{Offset: offset, Data: []byte("data1")},
		}, nil
	}

	peerState := &mock.AccountsStub{
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return testError
		},
	}
	peerState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		peerAccount, _ := state.NewPeerAccount([]byte("1234"))
		peerAccount.BLSPublicKey = address
		peerAccount.IndexInList = 1
		return peerAccount, nil
	}

	stakingData := systemSmartContracts.StakedDataV2{
		StakeValue:    big.NewInt(100),
		RewardAddress: []byte(address),
	}
	marshalizer := &mock.MarshalizerMock{}

	scDataGetter := &mock.ScQueryStub{}
	scDataGetter.ExecuteQueryCalled = func(query *process.SCQuery) (output *vmcommon.VMOutput, e error) {
		retData, _ := json.Marshal(&stakingData)
		return &vmcommon.VMOutput{ReturnData: [][]byte{retData}}, nil
	}

	arguments.ArgParser = argParser
	arguments.CurrTxs = currTx
	arguments.PeerState = peerState
	arguments.Marshalizer = marshalizer
	arguments.ScQuery = scDataGetter
	stp, _ := NewStakingToPeer(arguments)

	blockBody := createBlockBody()
	err := stp.UpdateProtocol(blockBody, 0)
	assert.Equal(t, testError, err)
}

func TestStakingToPeer_UpdatePeerState(t *testing.T) {
	t.Parallel()

	var peerAccount state.PeerAccountHandler
	peerAccount = state.NewEmptyPeerAccount()
	peerAccountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (state.AccountHandler, error) {
			return peerAccount, nil
		},
	}

	arguments := createMockArgumentsNewStakingToPeer()
	arguments.PeerState = peerAccountsDB
	stp, _ := NewStakingToPeer(arguments)

	stakingData := systemSmartContracts.StakedDataV2{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		UnStakedEpoch: core.DefaultUnstakedEpoch,
		RewardAddress: []byte("rwd"),
		StakeValue:    big.NewInt(0),
		JailedRound:   0,
		JailedNonce:   0,
		UnJailedNonce: 0,
		StakedNonce:   math.MaxUint64,
	}

	blsPubKey := []byte("key")
	nonce := uint64(1)
	err := stp.updatePeerState(stakingData, blsPubKey, nonce)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(peerAccount.GetRewardAddress()))

	stakingData.Staked = true
	stakingData.StakedNonce = nonce
	err = stp.updatePeerState(stakingData, blsPubKey, nonce)
	assert.NoError(t, err)
	assert.True(t, bytes.Equal(blsPubKey, peerAccount.GetBLSPublicKey()))
	assert.True(t, bytes.Equal(stakingData.RewardAddress, peerAccount.GetRewardAddress()))
	assert.Equal(t, string(core.NewList), peerAccount.GetList())

	stakingData.UnStakedNonce = 11
	_ = stp.updatePeerState(stakingData, blsPubKey, stakingData.UnStakedNonce)
	assert.Equal(t, string(core.LeavingList), peerAccount.GetList())

	peerAccount.SetListAndIndex(0, string(core.EligibleList), 5)
	stakingData.JailedNonce = 12
	_ = stp.updatePeerState(stakingData, blsPubKey, stakingData.JailedNonce)
	assert.Equal(t, string(core.LeavingList), peerAccount.GetList())

	// it is still jailed - no change allowed
	stakingData.RegisterNonce = 13
	_ = stp.updatePeerState(stakingData, blsPubKey, stakingData.RegisterNonce)
	assert.Equal(t, string(core.LeavingList), peerAccount.GetList())

	stakingData.UnJailedNonce = 14
	_ = stp.updatePeerState(stakingData, blsPubKey, stakingData.UnJailedNonce)
	assert.Equal(t, string(core.NewList), peerAccount.GetList())

	stakingData.UnStakedNonce = 15
	_ = stp.updatePeerState(stakingData, blsPubKey, stakingData.UnStakedNonce)
	assert.Equal(t, string(core.LeavingList), peerAccount.GetList())
}
