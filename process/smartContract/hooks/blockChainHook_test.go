package hooks_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	genesisMock "github.com/ElrondNetwork/elrond-go/genesis/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon/trie"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	vmcommonBuiltInFunctions "github.com/ElrondNetwork/elrond-vm-common/builtInFunctions"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockBlockChainHookArgs() hooks.ArgBlockChainHook {
	datapool := dataRetrieverMock.NewPoolsHolderMock()
	arguments := hooks.ArgBlockChainHook{
		Accounts: &stateMock.AccountsStub{
			GetExistingAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
				return &mock.AccountWrapMock{}, nil
			},
		},
		PubkeyConv:         mock.NewPubkeyConverterMock(32),
		StorageService:     &mock.ChainStorerMock{},
		BlockChain:         &testscommon.ChainHandlerStub{},
		ShardCoordinator:   mock.NewOneShardCoordinatorMock(),
		Marshalizer:        &mock.MarshalizerMock{},
		Uint64Converter:    &mock.Uint64ByteSliceConverterMock{},
		BuiltInFunctions:   vmcommonBuiltInFunctions.NewBuiltInFunctionContainer(),
		NFTStorageHandler:  &testscommon.SimpleNFTStorageHandlerStub{},
		DataPool:           datapool,
		CompiledSCPool:     datapool.SmartContracts(),
		EpochNotifier:      &epochNotifier.EpochNotifierStub{},
		NilCompiledSCStore: true,
		EnableEpochs: config.EnableEpochs{
			DoNotReturnOldBlockInBlockchainHookEnableEpoch: math.MaxUint32,
		},
	}
	return arguments
}

func createContractCallInput(function string, sender, receiver []byte) *vmcommon.ContractCallInput {
	return &vmcommon.ContractCallInput{
		Function:      function,
		RecipientAddr: receiver,
		VMInput: vmcommon.VMInput{
			CallerAddr: sender,
		},
	}
}

func TestNewBlockChainHookImpl(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args        func() hooks.ArgBlockChainHook
		expectedErr error
	}{
		{
			args: func() hooks.ArgBlockChainHook {
				args := createMockBlockChainHookArgs()
				args.Accounts = nil
				return args
			},
			expectedErr: process.ErrNilAccountsAdapter,
		},
		{
			args: func() hooks.ArgBlockChainHook {
				args := createMockBlockChainHookArgs()
				args.PubkeyConv = nil
				return args
			},
			expectedErr: process.ErrNilPubkeyConverter,
		},
		{
			args: func() hooks.ArgBlockChainHook {
				args := createMockBlockChainHookArgs()
				args.StorageService = nil
				return args
			},
			expectedErr: process.ErrNilStorage,
		},
		{
			args: func() hooks.ArgBlockChainHook {
				args := createMockBlockChainHookArgs()
				args.BlockChain = nil
				return args
			},
			expectedErr: process.ErrNilBlockChain,
		},
		{
			args: func() hooks.ArgBlockChainHook {
				args := createMockBlockChainHookArgs()
				args.ShardCoordinator = nil
				return args
			},
			expectedErr: process.ErrNilShardCoordinator,
		},
		{
			args: func() hooks.ArgBlockChainHook {
				args := createMockBlockChainHookArgs()
				args.Marshalizer = nil
				return args
			},
			expectedErr: process.ErrNilMarshalizer,
		},
		{
			args: func() hooks.ArgBlockChainHook {
				args := createMockBlockChainHookArgs()
				args.Uint64Converter = nil
				return args
			},
			expectedErr: process.ErrNilUint64Converter,
		},
		{
			args: func() hooks.ArgBlockChainHook {
				args := createMockBlockChainHookArgs()
				args.BuiltInFunctions = nil
				return args
			},
			expectedErr: process.ErrNilBuiltInFunction,
		},
		{
			args: func() hooks.ArgBlockChainHook {
				args := createMockBlockChainHookArgs()
				args.CompiledSCPool = nil
				return args
			},
			expectedErr: process.ErrNilCacher,
		},
		{
			args: func() hooks.ArgBlockChainHook {
				args := createMockBlockChainHookArgs()
				args.NFTStorageHandler = nil
				return args
			},
			expectedErr: process.ErrNilNFTStorageHandler,
		},
		{
			args: func() hooks.ArgBlockChainHook {
				args := createMockBlockChainHookArgs()
				args.EpochNotifier = nil
				return args
			},
			expectedErr: process.ErrNilEpochNotifier,
		},
		{
			args: func() hooks.ArgBlockChainHook {
				args := createMockBlockChainHookArgs()
				args.NilCompiledSCStore = false
				args.ConfigSCStorage = config.StorageConfig{
					Cache: config.CacheConfig{
						Capacity: 1,
					},
					DB: config.DBConfig{
						MaxBatchSize: 100,
					},
				}
				return args
			},
			expectedErr: storage.ErrCacheSizeIsLowerThanBatchSize,
		},
		{
			args: func() hooks.ArgBlockChainHook {
				return createMockBlockChainHookArgs()
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		bh, err := hooks.NewBlockChainHookImpl(test.args())
		require.Equal(t, test.expectedErr, err)

		if test.expectedErr != nil {
			require.Nil(t, bh)
		} else {
			require.NotNil(t, bh)
		}
	}
}

func TestBlockChainHookImpl_GetCode(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	t.Run("nil account expect nil code", func(t *testing.T) {
		t.Parallel()

		bh, _ := hooks.NewBlockChainHookImpl(args)
		code := bh.GetCode(nil)
		require.Nil(t, code)
	})

	t.Run("expect correct returned code", func(t *testing.T) {
		t.Parallel()

		expectedCodeHash := []byte("codeHash")
		expectedCode := []byte("code")

		args.Accounts = &stateMock.AccountsStub{
			GetCodeCalled: func(codeHash []byte) []byte {
				require.Equal(t, expectedCodeHash, codeHash)
				return expectedCode
			},
		}
		bh, _ := hooks.NewBlockChainHookImpl(args)

		account, _ := state.NewUserAccount([]byte("address"))
		account.SetCodeHash(expectedCodeHash)

		code := bh.GetCode(account)
		require.Equal(t, expectedCode, code)
	})
}

func TestBlockChainHookImpl_GetUserAccountNotASystemAccountInCrossShard(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	args.ShardCoordinator = &mock.ShardCoordinatorStub{
		ComputeIdCalled: func(address []byte) uint32 {
			return 0
		},
		SelfIdCalled: func() uint32 {
			return 1
		},
	}
	args.Accounts = &stateMock.AccountsStub{}
	bh, _ := hooks.NewBlockChainHookImpl(args)
	addr := bytes.Repeat([]byte{0}, 32)
	_, err := bh.GetUserAccount(addr)
	assert.Equal(t, state.ErrAccNotFound, err)
}

func TestBlockChainHookImpl_GetUserAccountGetAccFromAddressErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")

	args := createMockBlockChainHookArgs()
	args.Accounts = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			return nil, errExpected
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)
	_, err := bh.GetUserAccount(make([]byte, 0))
	assert.Equal(t, errExpected, err)
}

func TestBlockChainHookImpl_GetUserAccountWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	args.Accounts = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			return &mock.PeerAccountHandlerMock{}, nil
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)
	_, err := bh.GetUserAccount(make([]byte, 0))
	assert.Equal(t, state.ErrWrongTypeAssertion, err)
}

func TestBlockChainHookImpl_GetUserAccount(t *testing.T) {
	t.Parallel()

	expectedAccount, _ := state.NewUserAccount([]byte("1234"))
	args := createMockBlockChainHookArgs()
	args.Accounts = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			return expectedAccount, nil
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)
	acc, err := bh.GetUserAccount(expectedAccount.Address)

	assert.Nil(t, err)
	assert.Equal(t, expectedAccount, acc)
}

func TestBlockChainHookImpl_GetStorageDataAccountNotFoundExpectEmptyStorage(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	address := []byte("address")
	args.Accounts = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
			require.Equal(t, address, addressContainer)
			return nil, state.ErrAccNotFound
		},
	}

	bh, _ := hooks.NewBlockChainHookImpl(args)
	storageData, err := bh.GetStorageData(address, []byte("index"))
	require.Equal(t, []byte{}, storageData)
	require.Nil(t, err)
}

func TestBlockChainHookImpl_GetStorageDataCannotRetrieveAccountValueExpectError(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	address := []byte("address")
	index := []byte("i")
	expectedErr := errors.New("error retrieving value")

	dataTrieStub := &trie.DataTrieTrackerStub{
		RetrieveValueCalled: func(key []byte) ([]byte, error) {
			require.Equal(t, index, key)
			return nil, expectedErr
		},
	}
	account := &mock.AccountWrapMock{
		AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
			return dataTrieStub
		},
	}
	args.Accounts = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
			require.Equal(t, address, addressContainer)
			return account, nil
		},
	}

	bh, _ := hooks.NewBlockChainHookImpl(args)
	storageData, err := bh.GetStorageData(address, index)
	require.Nil(t, storageData)
	require.Equal(t, expectedErr, err)
}

func TestBlockChainHookImpl_GetStorageDataErrorsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")

	args := createMockBlockChainHookArgs()
	args.Accounts = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			return nil, errExpected
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)

	value, err := bh.GetStorageData(make([]byte, 0), make([]byte, 0))

	assert.Equal(t, errExpected, err)
	assert.Nil(t, value)
}

func TestBlockChainHookImpl_GetStorageDataShouldWork(t *testing.T) {
	t.Parallel()

	variableIdentifier := []byte("variable")
	variableValue := []byte("value")
	accnt := mock.NewAccountWrapMock(nil)
	_ = accnt.DataTrieTracker().SaveKeyValue(variableIdentifier, variableValue)

	args := createMockBlockChainHookArgs()
	args.Accounts = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			return accnt, nil
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)

	value, err := bh.GetStorageData(make([]byte, 0), variableIdentifier)

	assert.Nil(t, err)
	assert.Equal(t, variableValue, value)
}

func TestBlockChainHookImpl_NewAddressLengthNoGood(t *testing.T) {
	t.Parallel()

	acnts := &stateMock.AccountsStub{}
	acnts.GetExistingAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		return state.NewUserAccount(address)
	}
	args := createMockBlockChainHookArgs()
	args.Accounts = acnts
	bh, _ := hooks.NewBlockChainHookImpl(args)

	address := []byte("test")
	nonce := uint64(10)

	scAddress, err := bh.NewAddress(address, nonce, []byte("00"))
	assert.Equal(t, hooks.ErrAddressLengthNotCorrect, err)
	assert.Nil(t, scAddress)

	address = []byte("1234567890123456789012345678901234567890")
	scAddress, err = bh.NewAddress(address, nonce, []byte("00"))
	assert.Equal(t, hooks.ErrAddressLengthNotCorrect, err)
	assert.Nil(t, scAddress)
}

func TestBlockChainHookImpl_NewAddressVMTypeTooLong(t *testing.T) {
	t.Parallel()

	acnts := &stateMock.AccountsStub{}
	acnts.GetExistingAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		return state.NewUserAccount(address)
	}
	args := createMockBlockChainHookArgs()
	args.Accounts = acnts
	bh, _ := hooks.NewBlockChainHookImpl(args)

	address := []byte("01234567890123456789012345678900")
	nonce := uint64(10)

	vmType := []byte("010")
	scAddress, err := bh.NewAddress(address, nonce, vmType)
	assert.Equal(t, hooks.ErrVMTypeLengthIsNotCorrect, err)
	assert.Nil(t, scAddress)
}

func TestBlockChainHookImpl_NewAddress(t *testing.T) {
	t.Parallel()

	acnts := &stateMock.AccountsStub{}
	acnts.GetExistingAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		return state.NewUserAccount(address)
	}
	args := createMockBlockChainHookArgs()
	args.Accounts = acnts
	bh, _ := hooks.NewBlockChainHookImpl(args)

	address := []byte("01234567890123456789012345678900")
	nonce := uint64(10)

	vmType := []byte("11")
	scAddress1, err := bh.NewAddress(address, nonce, vmType)
	assert.Nil(t, err)

	for i := 0; i < 8; i++ {
		assert.Equal(t, scAddress1[i], uint8(0))
	}
	assert.True(t, bytes.Equal(vmType, scAddress1[8:10]))

	nonce++
	scAddress2, err := bh.NewAddress(address, nonce, []byte("00"))
	assert.Nil(t, err)

	assert.False(t, bytes.Equal(scAddress1, scAddress2))

	fmt.Printf("%s \n%s \n", hex.EncodeToString(scAddress1), hex.EncodeToString(scAddress2))
}

func TestBlockChainHookImpl_GetBlockhashNilBlockHeaderExpectError(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	args.BlockChain = &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return nil
		},
	}

	bh, _ := hooks.NewBlockChainHookImpl(args)
	hash, err := bh.GetBlockhash(0)
	require.Nil(t, hash)
	require.Equal(t, process.ErrNilBlockHeader, err)
}

func TestBlockChainHookImpl_GetBlockhashInvalidNonceExpectError(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	args.BlockChain = &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 1}
		},
	}

	bh, _ := hooks.NewBlockChainHookImpl(args)
	hash, err := bh.GetBlockhash(2)
	require.Nil(t, hash)
	require.Equal(t, process.ErrInvalidNonceRequest, err)
}

func TestBlockChainHookImpl_GetBlockhashShouldReturnCurrentBlockHeaderHash(t *testing.T) {
	t.Parallel()

	hdrToRet := &block.Header{Nonce: 2}
	hashToRet := []byte("hash")
	args := createMockBlockChainHookArgs()
	args.BlockChain = &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return hdrToRet
		},
		GetCurrentBlockHeaderHashCalled: func() []byte {
			return hashToRet
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)

	hash, err := bh.GetBlockhash(2)
	assert.Nil(t, err)
	assert.Equal(t, hashToRet, hash)
}

func TestBlockChainHookImpl_GetBlockhashFromStorerErrorReadingFromStorage(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	args.BlockChain = &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 10}
		},
	}
	storer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			return nil, errors.New("local error")
		},
	}
	args.StorageService = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return storer
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)

	hash, err := bh.GetBlockhash(2)
	require.Nil(t, hash)
	require.Equal(t, process.ErrMissingHashForHeaderNonce, err)
}

func TestBlockChainHookImpl_GetBlockhashFromStorerInSameEpoch(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	hash := []byte("hash")
	nonce := uint64(10)
	header := &block.Header{Nonce: nonce}
	shardID := args.ShardCoordinator.SelfId()
	nonceToByteSlice := args.Uint64Converter.ToByteSlice(nonce)
	marshalledHeader, _ := args.Marshalizer.Marshal(header)

	args.BlockChain = &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return header
		},
	}

	storerBlockHeader := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			require.Equal(t, hash, key)
			return marshalledHeader, nil
		},
	}
	storerShardHdrNonceHash := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			require.Equal(t, nonceToByteSlice, key)
			return hash, nil
		},
	}
	args.StorageService = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			switch unitType {
			case dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardID):
				return storerShardHdrNonceHash
			case dataRetriever.BlockHeaderUnit:
				return storerBlockHeader
			default:
				require.Fail(t, "should not search in another storer")
				return nil
			}
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)

	actualHash, err := bh.GetBlockhash(nonce - 1)
	require.Nil(t, err)
	require.Equal(t, hash, actualHash)
}

func TestBlockChainHookImpl_GetBlockhashFromStorerInSameEpochWithFlagEnabled(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	args.EnableEpochs.DoNotReturnOldBlockInBlockchainHookEnableEpoch = 0
	nonce := uint64(10)
	header := &block.Header{Nonce: nonce}
	shardID := args.ShardCoordinator.SelfId()

	args.BlockChain = &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return header
		},
	}

	storerBlockHeader := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			require.Fail(t, "should have not called Get operation")
			return nil, nil
		},
	}
	storerShardHdrNonceHash := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			require.Fail(t, "should have not called Get operation")
			return nil, nil
		},
	}
	args.StorageService = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			switch unitType {
			case dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardID):
				return storerShardHdrNonceHash
			case dataRetriever.BlockHeaderUnit:
				return storerBlockHeader
			default:
				require.Fail(t, "should not search in another storer")
				return nil
			}
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)

	hash, err := bh.GetBlockhash(nonce - 1)
	require.Equal(t, process.ErrInvalidNonceRequest, err)
	require.Nil(t, hash)
}

func TestBlockChainHookImpl_GetBlockhashFromOldEpochExpectError(t *testing.T) {
	t.Parallel()

	hdrToRet := &block.Header{Nonce: 2, Epoch: 2}
	hashToRet := []byte("hash")
	args := createMockBlockChainHookArgs()

	marshaledData, _ := args.Marshalizer.Marshal(hdrToRet)

	args.BlockChain = &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 10, Epoch: 10}
		},
	}
	args.StorageService = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			if uint8(unitType) >= uint8(dataRetriever.ShardHdrNonceHashDataUnit) {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						return hashToRet, nil
					},
				}
			}

			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return marshaledData, nil
				},
			}
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)

	_, err := bh.GetBlockhash(2)
	assert.Equal(t, err, process.ErrInvalidBlockRequestOldEpoch)
}

func TestBlockChainHookImpl_GettersFromBlockchainCurrentHeader(t *testing.T) {
	t.Parallel()

	t.Run("nil header, expect default values", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		args.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return nil
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		assert.Equal(t, uint64(0), bh.LastNonce())
		assert.Equal(t, uint64(0), bh.LastRound())
		assert.Equal(t, uint64(0), bh.LastTimeStamp())
		assert.Equal(t, uint32(0), bh.LastEpoch())
		assert.Equal(t, []byte{}, bh.LastRandomSeed())
		assert.Equal(t, []byte{}, bh.GetStateRootHash())
	})
	t.Run("custom header, expect correct values are returned", func(t *testing.T) {
		t.Parallel()

		nonce := uint64(37)
		round := uint64(5)
		timestamp := uint64(1234)
		randSeed := []byte("a")
		rootHash := []byte("b")
		epoch := uint32(7)
		hdrToRet := &block.Header{
			Nonce:     nonce,
			Round:     round,
			TimeStamp: timestamp,
			RandSeed:  randSeed,
			RootHash:  rootHash,
			Epoch:     epoch,
		}

		args := createMockBlockChainHookArgs()
		args.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return hdrToRet
			},
			GetCurrentBlockRootHashCalled: func() []byte {
				return hdrToRet.RootHash
			},
		}
		bh, _ := hooks.NewBlockChainHookImpl(args)

		assert.Equal(t, nonce, bh.LastNonce())
		assert.Equal(t, round, bh.LastRound())
		assert.Equal(t, timestamp, bh.LastTimeStamp())
		assert.Equal(t, epoch, bh.LastEpoch())
		assert.Equal(t, randSeed, bh.LastRandomSeed())
		assert.Equal(t, rootHash, bh.GetStateRootHash())
	})
	t.Run("custom header, do not return old block is set, expect default values", func(t *testing.T) {
		t.Parallel()

		nonce := uint64(37)
		round := uint64(5)
		timestamp := uint64(1234)
		randSeed := []byte("a")
		rootHash := []byte("b")
		epoch := uint32(7)
		hdrToRet := &block.Header{
			Nonce:     nonce,
			Round:     round,
			TimeStamp: timestamp,
			RandSeed:  randSeed,
			RootHash:  rootHash,
			Epoch:     epoch,
		}

		args := createMockBlockChainHookArgs()
		args.EnableEpochs.DoNotReturnOldBlockInBlockchainHookEnableEpoch = 0
		args.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return hdrToRet
			},
			GetCurrentBlockRootHashCalled: func() []byte {
				return hdrToRet.RootHash
			},
		}
		bh, _ := hooks.NewBlockChainHookImpl(args)

		assert.Equal(t, nonce, bh.LastNonce())
		assert.Equal(t, round, bh.LastRound())
		assert.Equal(t, timestamp, bh.LastTimeStamp())
		assert.Equal(t, randSeed, bh.LastRandomSeed())
		assert.Equal(t, epoch, bh.LastEpoch())
		assert.Equal(t, rootHash, bh.GetStateRootHash())
	})
}

func TestBlockChainHookImpl_GettersFromCurrentHeader(t *testing.T) {
	t.Parallel()

	nonce := uint64(37)
	round := uint64(5)
	timestamp := uint64(1234)
	randSeed := []byte("a")
	epoch := uint32(7)
	hdr := &block.Header{
		Nonce:     nonce,
		Round:     round,
		TimeStamp: timestamp,
		RandSeed:  randSeed,
		Epoch:     epoch,
	}

	args := createMockBlockChainHookArgs()
	bh, _ := hooks.NewBlockChainHookImpl(args)

	bh.SetCurrentHeader(hdr)
	assert.Equal(t, nonce, bh.CurrentNonce())
	assert.Equal(t, round, bh.CurrentRound())
	assert.Equal(t, timestamp, bh.CurrentTimeStamp())
	assert.Equal(t, epoch, bh.CurrentEpoch())
	assert.Equal(t, randSeed, bh.CurrentRandomSeed())

	bh.SetCurrentHeader(nil)
	assert.Equal(t, nonce, bh.CurrentNonce())
	assert.Equal(t, round, bh.CurrentRound())
	assert.Equal(t, timestamp, bh.CurrentTimeStamp())
	assert.Equal(t, epoch, bh.CurrentEpoch())
	assert.Equal(t, randSeed, bh.CurrentRandomSeed())
}

func TestBlockChainHookImpl_SaveNFTMetaDataToSystemAccount(t *testing.T) {
	t.Parallel()

	expectedTx := &transaction.Transaction{Nonce: 1}

	args := createMockBlockChainHookArgs()
	args.NFTStorageHandler = &testscommon.SimpleNFTStorageHandlerStub{
		SaveNFTMetaDataToSystemAccountCalled: func(tx data.TransactionHandler) error {
			require.Equal(t, expectedTx, tx)
			return nil
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)
	err := bh.SaveNFTMetaDataToSystemAccount(expectedTx)
	require.Nil(t, err)
}

func TestBlockChainHookImpl_GetShardOfAddress(t *testing.T) {
	t.Parallel()

	expectedAddr := []byte("address")
	expectedShardID := uint32(444)

	args := createMockBlockChainHookArgs()
	args.ShardCoordinator = &testscommon.ShardsCoordinatorMock{
		ComputeIdCalled: func(address []byte) uint32 {
			require.Equal(t, expectedAddr, address)
			return expectedShardID
		},
	}

	bh, _ := hooks.NewBlockChainHookImpl(args)
	shardID := bh.GetShardOfAddress(expectedAddr)
	require.Equal(t, expectedShardID, shardID)
}

func TestBlockChainHookImpl_IsPayableNormalAccount(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	bh, _ := hooks.NewBlockChainHookImpl(args)
	isPayable, err := bh.IsPayable([]byte("address"), []byte("address"))
	assert.True(t, isPayable)
	assert.Nil(t, err)
}

func TestBlockChainHookImpl_IsPayableSCNonPayable(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	args.Accounts = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			acc := &mock.AccountWrapMock{}
			acc.SetCodeMetadata([]byte{0, 0})
			return acc, nil
		},
	}
	bh, _ := hooks.NewBlockChainHookImpl(args)
	isPayable, err := bh.IsPayable([]byte("address"), make([]byte, 32))
	assert.False(t, isPayable)
	assert.Nil(t, err)
}

func TestBlockChainHookImpl_IsPayablePayable(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	args.Accounts = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			acc := &mock.AccountWrapMock{}
			acc.SetCodeMetadata([]byte{0, vmcommon.MetadataPayable})
			return acc, nil
		},
	}

	bh, _ := hooks.NewBlockChainHookImpl(args)
	isPayable, err := bh.IsPayable([]byte("address"), make([]byte, 32))
	assert.True(t, isPayable)
	assert.Nil(t, err)

	isPayable, err = bh.IsPayable(make([]byte, 32), make([]byte, 32))
	assert.True(t, isPayable)
	assert.Nil(t, err)
}

func TestBlockChainHookImpl_IsPayablePayableBySC(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	args.Accounts = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(address []byte) (handler vmcommon.AccountHandler, e error) {
			acc := &mock.AccountWrapMock{}
			acc.SetCodeMetadata([]byte{0, vmcommon.MetadataPayableBySC})
			return acc, nil
		},
	}

	bh, _ := hooks.NewBlockChainHookImpl(args)
	isPayable, err := bh.IsPayable(make([]byte, 32), make([]byte, 32))
	assert.True(t, isPayable)
	assert.Nil(t, err)
}

func TestBlockChainHookImpl_IsPayableReceiverIsSystemAccountNotPayable(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	bh, _ := hooks.NewBlockChainHookImpl(args)

	receiver := make([]byte, 32)
	copy(receiver, core.SystemAccountAddress)

	isPayable, err := bh.IsPayable(make([]byte, 32), receiver)
	require.False(t, isPayable)
	require.Nil(t, err)
}

func TestTestBlockChainHookImpl_IsPayableReceiverIsCrossShardNotPayable(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	receiver := make([]byte, 32)
	args.ShardCoordinator = &testscommon.ShardsCoordinatorMock{
		ComputeIdCalled: func(address []byte) uint32 {
			require.Equal(t, receiver, address)
			return 1
		},
		SelfIDCalled: func() uint32 {
			return 0
		},
	}

	bh, _ := hooks.NewBlockChainHookImpl(args)
	isPayable, err := bh.IsPayable([]byte("sender"), receiver)
	require.True(t, isPayable)
	require.Nil(t, err)
}

func TestTestBlockChainHookImpl_IsPayableReceiverNotFoundNotPayable(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	receiver := make([]byte, 32)
	args.Accounts = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
			require.Equal(t, addressContainer, receiver)
			return nil, state.ErrAccNotFound
		},
	}

	bh, _ := hooks.NewBlockChainHookImpl(args)
	isPayable, err := bh.IsPayable([]byte("sender"), receiver)
	require.False(t, isPayable)
	require.Nil(t, err)
}

func TestTestBlockChainHookImpl_IsPayableErrorGettingReceiverNotPayable(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	receiver := make([]byte, 32)
	errGetAccount := errors.New("error getting account")
	args.Accounts = &stateMock.AccountsStub{
		GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
			require.Equal(t, addressContainer, receiver)
			return nil, errGetAccount
		},
	}

	bh, _ := hooks.NewBlockChainHookImpl(args)
	isPayable, err := bh.IsPayable([]byte("sender"), receiver)
	require.False(t, isPayable)
	require.Equal(t, errGetAccount, err)
}

func TestBlockChainHookImpl_GetBuiltinFunctionNamesAndContainer(t *testing.T) {
	t.Parallel()

	builtInFunctionContainer := vmcommonBuiltInFunctions.NewBuiltInFunctionContainer()
	_ = builtInFunctionContainer.Add("func1", &mock.BuiltInFunctionStub{})
	_ = builtInFunctionContainer.Add("func2", &mock.BuiltInFunctionStub{})

	args := createMockBlockChainHookArgs()
	args.BuiltInFunctions = builtInFunctionContainer

	bh, _ := hooks.NewBlockChainHookImpl(args)
	funcNames := bh.GetBuiltinFunctionNames()
	expectedFuncNames := vmcommon.FunctionNames{
		"func1": {},
		"func2": {},
	}
	require.Equal(t, expectedFuncNames, funcNames)
	require.Equal(t, builtInFunctionContainer, bh.GetBuiltinFunctionsContainer())
}

func TestBlockChainHookImpl_NumberOfShards(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	args.ShardCoordinator = &testscommon.ShardsCoordinatorMock{NoShards: 4}

	bh, _ := hooks.NewBlockChainHookImpl(args)
	require.Equal(t, uint32(4), bh.NumberOfShards())
}

func TestBlockChainHookImpl_SaveCompiledCode(t *testing.T) {
	t.Parallel()

	code := []byte("code")
	codeHash := []byte("codeHash")

	t.Run("get compiled code from compiled sc pool", func(t *testing.T) {
		args := createMockBlockChainHookArgs()

		wasCodeSavedInPool := &atomic.Flag{}
		args.CompiledSCPool = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				require.Equal(t, codeHash, key)
				return code, true
			},
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				wasCodeSavedInPool.SetValue(true)
				return false
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		found, actualCode := bh.GetCompiledCode(codeHash)
		require.True(t, found)
		require.Equal(t, code, actualCode)
		require.False(t, wasCodeSavedInPool.IsSet())
	})

	t.Run("compiled code found in compiled sc pool, but not as byte slice, error getting it from storage", func(t *testing.T) {
		args := createMockBlockChainHookArgs()
		args.NilCompiledSCStore = true

		wasCodeSavedInPool := &atomic.Flag{}
		args.CompiledSCPool = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				require.Equal(t, codeHash, key)
				return struct{}{}, true
			},
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				wasCodeSavedInPool.SetValue(true)
				return false
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		found, actualCode := bh.GetCompiledCode(codeHash)
		require.False(t, found)
		require.Nil(t, actualCode)
		require.False(t, wasCodeSavedInPool.IsSet())
	})

	t.Run("compiled code found in storage, but nil", func(t *testing.T) {
		args := createMockBlockChainHookArgs()
		args.NilCompiledSCStore = false
		args.ConfigSCStorage = config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10,
				Type:     string(storageUnit.LRUCache),
			},
			DB: config.DBConfig{
				FilePath:     "test1",
				Type:         string(storageUnit.MemoryDB),
				MaxBatchSize: 1,
				MaxOpenFiles: 10,
			},
		}
		wasCodeSavedInPool := &atomic.Flag{}
		args.CompiledSCPool = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				require.Equal(t, codeHash, key)
				return nil, false
			},
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				wasCodeSavedInPool.SetValue(true)
				return false
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		bh.SaveCompiledCode(codeHash, nil)

		wasCodeSavedInPool.Reset()
		found, actualCode := bh.GetCompiledCode(codeHash)
		require.False(t, found)
		require.Nil(t, actualCode)
		require.False(t, wasCodeSavedInPool.IsSet())

		_ = bh.Close()
	})

	t.Run("compiled code not found in compiled sc pool, get it from storage", func(t *testing.T) {
		args := createMockBlockChainHookArgs()
		args.ConfigSCStorage = config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10,
				Type:     string(storageUnit.LRUCache),
			},
			DB: config.DBConfig{
				FilePath:     "test2",
				Type:         string(storageUnit.MemoryDB),
				MaxBatchSize: 1,
				MaxOpenFiles: 10,
			},
		}
		args.NilCompiledSCStore = false
		args.CompiledSCPool = &testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				require.Equal(t, codeHash, key)
				return nil, false
			},
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				require.Equal(t, codeHash, key)
				require.Equal(t, code, value)
				require.Equal(t, len(code), sizeInBytes)
				return false
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)

		bh.SaveCompiledCode(codeHash, code)
		found, actualCode := bh.GetCompiledCode(codeHash)
		require.True(t, found)
		require.Equal(t, code, actualCode)

		bh.DeleteCompiledCode(codeHash)
		found, actualCode = bh.GetCompiledCode(codeHash)
		require.False(t, found)
		require.Nil(t, actualCode)

		_ = bh.Close()
	})
}

func TestBlockChainHookImpl_GetSnapshot(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	args.Accounts = &stateMock.AccountsStub{
		JournalLenCalled: func() int {
			return 444
		},
	}

	bh, _ := hooks.NewBlockChainHookImpl(args)
	require.Equal(t, 444, bh.GetSnapshot())
}

func TestBlockChainHookImpl_RevertToSnapshot(t *testing.T) {
	t.Parallel()

	args := createMockBlockChainHookArgs()
	args.Accounts = &stateMock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			require.Equal(t, 444, snapshot)
			return nil
		},
	}

	bh, _ := hooks.NewBlockChainHookImpl(args)
	require.Nil(t, bh.RevertToSnapshot(444))
}

func TestBlockChainHookImpl_ProcessBuiltInFunction(t *testing.T) {
	t.Parallel()

	funcName := "func"
	builtInFunctionsContainer := vmcommonBuiltInFunctions.NewBuiltInFunctionContainer()
	_ = builtInFunctionsContainer.Add(funcName, &mock.BuiltInFunctionStub{})

	addrSender := []byte("addr sender")
	addrReceiver := []byte("addr receiver")

	errGetAccount := errors.New("cannot get account")
	errSaveAccount := errors.New("error saving account")

	t.Run("nil input, expect error", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		bh, _ := hooks.NewBlockChainHookImpl(args)

		output, err := bh.ProcessBuiltInFunction(nil)
		require.Nil(t, output)
		require.Equal(t, process.ErrNilVmInput, err)
	})

	t.Run("no function set in built in function container, expect error", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		bh, _ := hooks.NewBlockChainHookImpl(args)

		output, err := bh.ProcessBuiltInFunction(&vmcommon.ContractCallInput{})
		require.Nil(t, output)
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), vmcommonBuiltInFunctions.ErrInvalidContainerKey.Error()))
	})

	t.Run("cannot get sender account, expect error", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		args.BuiltInFunctions = builtInFunctionsContainer
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, addrSender, addressContainer)
				return nil, errGetAccount
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		input := createContractCallInput(funcName, addrSender, addrReceiver)
		output, err := bh.ProcessBuiltInFunction(input)

		require.Nil(t, output)
		require.Equal(t, errGetAccount, err)
	})

	t.Run("cannot convert sender account to user account, expect error", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		args.BuiltInFunctions = builtInFunctionsContainer
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, addrSender, addressContainer)
				return &genesisMock.BaseAccountMock{}, nil
			},
		}
		bh, _ := hooks.NewBlockChainHookImpl(args)
		input := createContractCallInput(funcName, addrSender, addrReceiver)
		output, err := bh.ProcessBuiltInFunction(input)

		require.Nil(t, output)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
	})

	t.Run("cannot load destination account, expect error", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		args.BuiltInFunctions = builtInFunctionsContainer
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, addrSender, addressContainer)
				return mock.NewAccountWrapMock(addrSender), nil
			},

			LoadAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, addrReceiver, addressContainer)
				return nil, errGetAccount
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		input := createContractCallInput(funcName, addrSender, addrReceiver)
		output, err := bh.ProcessBuiltInFunction(input)

		require.Nil(t, output)
		require.Equal(t, errGetAccount, err)
	})

	t.Run("cannot convert destination account to user account, expect error", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		args.BuiltInFunctions = builtInFunctionsContainer
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, addrSender, addressContainer)
				return mock.NewAccountWrapMock(addrSender), nil
			},

			LoadAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, addrReceiver, addressContainer)
				return &genesisMock.BaseAccountMock{}, nil
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		input := createContractCallInput(funcName, addrSender, addrReceiver)
		output, err := bh.ProcessBuiltInFunction(input)

		require.Nil(t, output)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
	})

	t.Run("cannot process new built in function, expect error", func(t *testing.T) {
		t.Parallel()

		newFunc := "newFunc"
		input := createContractCallInput(newFunc, addrSender, addrReceiver)

		errProcessBuiltInFunc := errors.New("error processing builtin func")
		newBuiltInFunc := &mock.BuiltInFunctionStub{
			ProcessBuiltinFunctionCalled: func(acntSnd, acntDst vmcommon.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
				require.Equal(t, addrSender, acntSnd.AddressBytes())
				require.Equal(t, addrReceiver, acntDst.AddressBytes())
				require.Equal(t, input, vmInput)

				return nil, errProcessBuiltInFunc
			},
		}
		newBuiltInFuncContainer := vmcommonBuiltInFunctions.NewBuiltInFunctionContainer()
		_ = newBuiltInFuncContainer.Add(newFunc, newBuiltInFunc)

		args := createMockBlockChainHookArgs()
		args.BuiltInFunctions = newBuiltInFuncContainer
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, addrSender, addressContainer)
				return mock.NewAccountWrapMock(addrSender), nil
			},

			LoadAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, addrReceiver, addressContainer)
				return mock.NewAccountWrapMock(addrReceiver), nil
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		output, err := bh.ProcessBuiltInFunction(input)
		require.Nil(t, output)
		require.Equal(t, errProcessBuiltInFunc, err)
	})

	t.Run("sender and receiver not in same shard, expect they are not saved", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		args.BuiltInFunctions = builtInFunctionsContainer

		args.ShardCoordinator = &testscommon.ShardsCoordinatorMock{
			ComputeIdCalled: func(address []byte) uint32 {
				isSender := bytes.Equal(addrSender, address)
				isReceiver := bytes.Equal(addrReceiver, address)
				require.True(t, isSender || isReceiver)

				return 0
			},
			SelfIDCalled: func() uint32 {
				return 1
			},
		}

		getSenderAccountCalled := &atomic.Flag{}
		getReceiverAccountCalled := &atomic.Flag{}
		saveAccountCalled := &atomic.Flag{}
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				getSenderAccountCalled.SetValue(true)
				return nil, nil
			},

			LoadAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				getReceiverAccountCalled.SetValue(true)
				return nil, nil
			},
			SaveAccountCalled: func(account vmcommon.AccountHandler) error {
				saveAccountCalled.SetValue(true)
				return nil
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		input := createContractCallInput(funcName, addrSender, addrReceiver)
		output, err := bh.ProcessBuiltInFunction(input)

		require.Nil(t, err)
		require.Equal(t, &vmcommon.VMOutput{}, output)
		require.False(t, getSenderAccountCalled.IsSet())
		require.False(t, getReceiverAccountCalled.IsSet())
		require.False(t, saveAccountCalled.IsSet())
	})

	t.Run("sender and receiver same shard, expect accounts saved", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		args.BuiltInFunctions = builtInFunctionsContainer

		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, addrSender, addressContainer)
				return mock.NewAccountWrapMock(addrSender), nil
			},

			LoadAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, addrReceiver, addressContainer)
				return mock.NewAccountWrapMock(addrReceiver), nil
			},
			SaveAccountCalled: func(account vmcommon.AccountHandler) error {
				isSender := bytes.Equal(addrSender, account.AddressBytes())
				isReceiver := bytes.Equal(addrReceiver, account.AddressBytes())

				require.True(t, isSender || isReceiver)
				return nil
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		input := createContractCallInput(funcName, addrSender, addrReceiver)
		output, err := bh.ProcessBuiltInFunction(input)

		require.Nil(t, err)
		require.Equal(t, &vmcommon.VMOutput{}, output)
	})

	t.Run("sender and receiver same shard, sender = receiver, expect only one account is saved", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		args.BuiltInFunctions = builtInFunctionsContainer

		getReceiverAccountCalled := &atomic.Flag{}
		ctSaveAccount := &atomic.Counter{}
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, addrSender, addressContainer)
				return mock.NewAccountWrapMock(addrSender), nil
			},

			LoadAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				getReceiverAccountCalled.SetValue(true)
				return nil, nil
			},
			SaveAccountCalled: func(account vmcommon.AccountHandler) error {
				require.Equal(t, addrSender, account.AddressBytes())
				ctSaveAccount.Increment()
				return nil
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		input := createContractCallInput(funcName, addrSender, addrSender)
		output, err := bh.ProcessBuiltInFunction(input)

		require.Nil(t, err)
		require.Equal(t, &vmcommon.VMOutput{}, output)
		require.Equal(t, int64(1), ctSaveAccount.Get())
		require.False(t, getReceiverAccountCalled.IsSet())
	})

	t.Run("cannot save sender account, expect error", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		args.BuiltInFunctions = builtInFunctionsContainer

		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, addrSender, addressContainer)
				return mock.NewAccountWrapMock(addrSender), nil
			},
			SaveAccountCalled: func(account vmcommon.AccountHandler) error {
				require.Equal(t, addrSender, account.AddressBytes())
				return errSaveAccount
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		input := createContractCallInput(funcName, addrSender, addrSender)
		output, err := bh.ProcessBuiltInFunction(input)

		require.Nil(t, output)
		require.Equal(t, errSaveAccount, err)
	})

	t.Run("cannot save receiver account, expect error", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		args.BuiltInFunctions = builtInFunctionsContainer

		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, addrSender, addressContainer)
				return mock.NewAccountWrapMock(addrSender), nil
			},

			LoadAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, addrReceiver, addressContainer)
				return mock.NewAccountWrapMock(addrReceiver), nil
			},
			SaveAccountCalled: func(account vmcommon.AccountHandler) error {
				isSender := bytes.Equal(addrSender, account.AddressBytes())
				isReceiver := bytes.Equal(addrReceiver, account.AddressBytes())
				require.True(t, isSender || isReceiver)

				if isSender {
					return nil
				}
				return errSaveAccount

			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		input := createContractCallInput(funcName, addrSender, addrReceiver)
		output, err := bh.ProcessBuiltInFunction(input)

		require.Nil(t, output)
		require.Equal(t, errSaveAccount, err)
	})
}

func TestBlockChainHookImpl_GetESDTToken(t *testing.T) {
	t.Parallel()

	address := []byte("address")
	token := []byte("tkn")
	nonce := uint64(0)
	emptyESDTData := &esdt.ESDigitalToken{Value: big.NewInt(0)}
	expectedErr := errors.New("expected error")
	completeEsdtTokenKey := []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier + string(token))
	testESDTData := &esdt.ESDigitalToken{
		Type:       uint32(core.Fungible),
		Value:      big.NewInt(1),
		Properties: []byte("properties"),
		TokenMetaData: &esdt.MetaData{
			Nonce:      1,
			Name:       []byte("name"),
			Creator:    []byte("creator"),
			Royalties:  2,
			Hash:       []byte("hash"),
			URIs:       [][]byte{[]byte("uri1"), []byte("uri2")},
			Attributes: []byte("attributes"),
		},
		Reserved: []byte("reserved"),
	}

	t.Run("account not found returns an empty esdt data", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
				return nil, state.ErrAccNotFound
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		esdtData, err := bh.GetESDTToken(address, token, nonce)
		assert.Nil(t, err)
		require.NotNil(t, esdtData)
		assert.Equal(t, emptyESDTData, esdtData)
	})
	t.Run("error unmarshal", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		invalidUnmarshalledData := []byte("invalid data")
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				require.Equal(t, address, addressContainer)
				account := mock.NewAccountWrapMock(address)
				_ = account.DataTrieTracker().SaveKeyValue(completeEsdtTokenKey, invalidUnmarshalledData)

				return account, nil
			},
		}
		errMarshaller := errors.New("error marshaller")
		args.Marshalizer = &testscommon.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				require.Equal(t, emptyESDTData, obj)
				require.Equal(t, invalidUnmarshalledData, buff)
				return errMarshaller
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		bh.SetFlagOptimizeNFTStore(false)

		esdtData, err := bh.GetESDTToken(address, token, nonce)
		require.Nil(t, esdtData)
		require.Equal(t, errMarshaller, err)
	})
	t.Run("accountsDB errors returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
				return nil, expectedErr
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		esdtData, err := bh.GetESDTToken(address, token, nonce)
		assert.Nil(t, esdtData)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("backwards compatibility - retrieve value errors, should return error", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				addressHandler := mock.NewAccountWrapMock(address)
				addressHandler.SetDataTrie(nil)

				return addressHandler, nil
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		bh.SetFlagOptimizeNFTStore(false)

		esdtData, err := bh.GetESDTToken(address, token, nonce)
		assert.Nil(t, esdtData)
		assert.Equal(t, state.ErrNilTrie, err)
	})
	t.Run("backwards compatibility - empty byte slice should return empty esdt token", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				addressHandler := mock.NewAccountWrapMock(address)
				addressHandler.SetDataTrie(&trie.TrieStub{
					GetCalled: func(key []byte) ([]byte, error) {
						return make([]byte, 0), nil
					},
				})

				return addressHandler, nil
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		bh.SetFlagOptimizeNFTStore(false)

		esdtData, err := bh.GetESDTToken(address, token, nonce)
		assert.Equal(t, emptyESDTData, esdtData)
		assert.Nil(t, err)
	})
	t.Run("backwards compatibility - should load the esdt data in case of an NFT", func(t *testing.T) {
		t.Parallel()

		nftNonce := uint64(44)
		args := createMockBlockChainHookArgs()
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				addressHandler := mock.NewAccountWrapMock(address)
				buffToken, _ := args.Marshalizer.Marshal(testESDTData)
				key := append(completeEsdtTokenKey, big.NewInt(0).SetUint64(nftNonce).Bytes()...)
				_ = addressHandler.DataTrieTracker().SaveKeyValue(key, buffToken)

				return addressHandler, nil
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		bh.SetFlagOptimizeNFTStore(false)

		esdtData, err := bh.GetESDTToken(address, token, nftNonce)
		assert.Equal(t, testESDTData, esdtData)
		assert.Nil(t, err)
	})
	t.Run("backwards compatibility - should load the esdt data", func(t *testing.T) {
		t.Parallel()

		args := createMockBlockChainHookArgs()
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				addressHandler := mock.NewAccountWrapMock(address)
				buffToken, _ := args.Marshalizer.Marshal(testESDTData)
				_ = addressHandler.DataTrieTracker().SaveKeyValue(completeEsdtTokenKey, buffToken)

				return addressHandler, nil
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)
		bh.SetFlagOptimizeNFTStore(false)

		esdtData, err := bh.GetESDTToken(address, token, nonce)
		assert.Equal(t, testESDTData, esdtData)
		assert.Nil(t, err)
	})
	t.Run("new optimized implementation - NFTStorageHandler errors", func(t *testing.T) {
		t.Parallel()

		nftNonce := uint64(44)
		args := createMockBlockChainHookArgs()
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				return mock.NewAccountWrapMock(address), nil
			},
		}
		args.NFTStorageHandler = &testscommon.SimpleNFTStorageHandlerStub{
			GetESDTNFTTokenOnDestinationCalled: func(accnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64) (*esdt.ESDigitalToken, bool, error) {
				assert.Equal(t, completeEsdtTokenKey, esdtTokenKey)
				assert.Equal(t, nftNonce, nonce)

				return nil, false, expectedErr
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)

		esdtData, err := bh.GetESDTToken(address, token, nftNonce)
		assert.Nil(t, esdtData)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("new optimized implementation - should return the esdt by calling NFTStorageHandler", func(t *testing.T) {
		t.Parallel()

		nftNonce := uint64(44)
		args := createMockBlockChainHookArgs()
		args.Accounts = &stateMock.AccountsStub{
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				return mock.NewAccountWrapMock(address), nil
			},
		}
		args.NFTStorageHandler = &testscommon.SimpleNFTStorageHandlerStub{
			GetESDTNFTTokenOnDestinationCalled: func(accnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64) (*esdt.ESDigitalToken, bool, error) {
				assert.Equal(t, completeEsdtTokenKey, esdtTokenKey)
				assert.Equal(t, nftNonce, nonce)
				copyToken := *testESDTData

				return &copyToken, false, nil
			},
		}

		bh, _ := hooks.NewBlockChainHookImpl(args)

		esdtData, err := bh.GetESDTToken(address, token, nftNonce)
		assert.Equal(t, testESDTData, esdtData)
		assert.Nil(t, err)
	})
}
