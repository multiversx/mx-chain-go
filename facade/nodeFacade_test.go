package facade

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	atomicCore "github.com/multiversx/mx-chain-core-go/core/atomic"
	nodeData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/debug"
	"github.com/multiversx/mx-chain-go/facade/mock"
	"github.com/multiversx/mx-chain-go/heartbeat/data"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/process"
	txSimData "github.com/multiversx/mx-chain-go/process/transactionEvaluator/data"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/testscommon"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedErr = errors.New("expected error")

func createMockArguments() ArgNodeFacade {
	return ArgNodeFacade{
		Node:                   &mock.NodeStub{},
		ApiResolver:            &mock.ApiResolverStub{},
		RestAPIServerDebugMode: false,
		WsAntifloodConfig: config.WebServerAntifloodConfig{
			SimultaneousRequests:               1,
			SameSourceRequests:                 1,
			SameSourceResetIntervalInSec:       1,
			TrieOperationsDeadlineMilliseconds: 1,
		},
		FacadeConfig: config.FacadeConfig{
			RestApiInterface: "127.0.0.1:8080",
			PprofEnabled:     false,
		},
		ApiRoutesConfig: config.ApiRoutesConfig{APIPackages: map[string]config.APIPackageConfig{
			"node": {
				Routes: []config.RouteConfig{
					{Name: "status"},
				},
			},
		}},
		AccountsState: &stateMock.AccountsStub{},
		PeerState:     &stateMock.AccountsStub{},
		Blockchain: &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() nodeData.HeaderHandler {
				return &block.Header{}
			},
			GetCurrentBlockRootHashCalled: func() []byte {
				return []byte("root hash")
			},
		},
	}
}

func TestNewNodeFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil Node should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.Node = nil
		nf, err := NewNodeFacade(arg)

		require.Nil(t, nf)
		require.Equal(t, ErrNilNode, err)
	})
	t.Run("nil ApiResolver should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.ApiResolver = nil
		nf, err := NewNodeFacade(arg)

		require.Nil(t, nf)
		require.Equal(t, ErrNilApiResolver, err)
	})
	t.Run("invalid ApiRoutesConfig should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.ApiRoutesConfig = config.ApiRoutesConfig{}
		nf, err := NewNodeFacade(arg)

		require.Nil(t, nf)
		require.True(t, errors.Is(err, ErrNoApiRoutesConfig))
	})
	t.Run("invalid SimultaneousRequests should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.WsAntifloodConfig.WebServerAntifloodEnabled = true
		arg.WsAntifloodConfig.SimultaneousRequests = 0
		nf, err := NewNodeFacade(arg)

		require.Nil(t, nf)
		require.True(t, errors.Is(err, ErrInvalidValue))
	})
	t.Run("invalid SameSourceRequests should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.WsAntifloodConfig.WebServerAntifloodEnabled = true
		arg.WsAntifloodConfig.SameSourceRequests = 0
		nf, err := NewNodeFacade(arg)

		require.Nil(t, nf)
		require.True(t, errors.Is(err, ErrInvalidValue))
	})
	t.Run("invalid SameSourceResetIntervalInSec should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.WsAntifloodConfig.WebServerAntifloodEnabled = true
		arg.WsAntifloodConfig.SameSourceResetIntervalInSec = 0
		nf, err := NewNodeFacade(arg)

		require.Nil(t, nf)
		require.True(t, errors.Is(err, ErrInvalidValue))
	})
	t.Run("invalid TrieOperationsDeadlineMilliseconds should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.WsAntifloodConfig.WebServerAntifloodEnabled = true
		arg.WsAntifloodConfig.TrieOperationsDeadlineMilliseconds = 0
		nf, err := NewNodeFacade(arg)

		require.Nil(t, nf)
		require.True(t, errors.Is(err, ErrInvalidValue))
	})
	t.Run("nil AccountsState should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.WsAntifloodConfig.WebServerAntifloodEnabled = true // coverage
		arg.AccountsState = nil
		nf, err := NewNodeFacade(arg)

		require.Nil(t, nf)
		require.Equal(t, ErrNilAccountState, err)
	})
	t.Run("nil PeerState should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.PeerState = nil
		nf, err := NewNodeFacade(arg)

		require.Nil(t, nf)
		require.Equal(t, ErrNilPeerState, err)
	})
	t.Run("nil Blockchain should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.Blockchain = nil
		nf, err := NewNodeFacade(arg)

		require.Nil(t, nf)
		require.Equal(t, ErrNilBlockchain, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.WsAntifloodConfig.EndpointsThrottlers = []config.EndpointsThrottlersConfig{
			{
				Endpoint:         "endpoint_1",
				MaxNumGoRoutines: 10,
			}, {
				Endpoint:         "endpoint_2",
				MaxNumGoRoutines: 0, // NewNumGoRoutinesThrottler fails for coverage
			},
		}
		nf, err := NewNodeFacade(arg)

		require.NotNil(t, nf)
		require.NoError(t, err)
	})
}

func TestNodeFacade_GetBalanceWithValidAddressShouldReturnBalance(t *testing.T) {
	t.Parallel()

	balance := big.NewInt(10)
	addr := "testAddress"
	node := &mock.NodeStub{
		GetBalanceCalled: func(address string, _ api.AccountQueryOptions) (*big.Int, api.BlockInfo, error) {
			if addr == address {
				return balance, api.BlockInfo{}, nil
			}
			return big.NewInt(0), api.BlockInfo{}, nil
		},
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	amount, _, err := nf.GetBalance(addr, api.AccountQueryOptions{})

	require.NoError(t, err)
	require.Equal(t, balance, amount)
}

func TestNodeFacade_GetBalanceWithUnknownAddressShouldReturnZeroBalance(t *testing.T) {
	t.Parallel()

	balance := big.NewInt(10)
	addr := "testAddress"
	unknownAddr := "unknownAddr"
	zeroBalance := big.NewInt(0)

	node := &mock.NodeStub{
		GetBalanceCalled: func(address string, _ api.AccountQueryOptions) (*big.Int, api.BlockInfo, error) {
			if addr == address {
				return balance, api.BlockInfo{}, nil
			}
			return big.NewInt(0), api.BlockInfo{}, nil
		},
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	amount, _, err := nf.GetBalance(unknownAddr, api.AccountQueryOptions{})
	require.NoError(t, err)
	require.Equal(t, zeroBalance, amount)
}

func TestNodeFacade_GetBalanceWithErrorOnNodeShouldReturnZeroBalanceAndError(t *testing.T) {
	t.Parallel()

	addr := "testAddress"
	zeroBalance := big.NewInt(0)

	node := &mock.NodeStub{
		GetBalanceCalled: func(address string, _ api.AccountQueryOptions) (*big.Int, api.BlockInfo, error) {
			return big.NewInt(0), api.BlockInfo{}, errors.New("error on getBalance on node")
		},
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	amount, _, err := nf.GetBalance(addr, api.AccountQueryOptions{})
	require.NotNil(t, err)
	require.Equal(t, zeroBalance, amount)
}

func TestNodeFacade_GetTransactionWithValidInputsShouldNotReturnError(t *testing.T) {
	t.Parallel()

	testHash := "testHash"
	testTx := &transaction.ApiTransactionResult{}
	node := &mock.NodeStub{}

	arg := createMockArguments()
	arg.ApiResolver = &mock.ApiResolverStub{
		GetTransactionHandler: func(hash string, withEvents bool) (*transaction.ApiTransactionResult, error) {
			if hash == testHash {
				return testTx, nil
			}
			return nil, nil
		},
	}
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	tx, err := nf.GetTransaction(testHash, false)
	require.NoError(t, err)
	require.Equal(t, testTx, tx)
}

func TestNodeFacade_GetTransactionWithUnknowHashShouldReturnNilAndNoError(t *testing.T) {
	t.Parallel()

	testHash := "testHash"
	testTx := &transaction.ApiTransactionResult{}
	arg := createMockArguments()
	arg.ApiResolver = &mock.ApiResolverStub{
		GetTransactionHandler: func(hash string, withEvents bool) (*transaction.ApiTransactionResult, error) {
			if hash == testHash {
				return testTx, nil
			}
			return nil, nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	tx, err := nf.GetTransaction("unknownHash", false)
	require.NoError(t, err)
	require.Nil(t, tx)
}

func TestNodeFacade_SetSyncer(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	nf, _ := NewNodeFacade(arg)

	sync := &mock.SyncTimerMock{}
	nf.SetSyncer(sync)
	require.Equal(t, sync, nf.GetSyncer())
}

func TestNodeFacade_GetAccount(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.Node = &mock.NodeStub{
			GetAccountCalled: func(_ string, _ api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
				return api.AccountResponse{}, api.BlockInfo{}, expectedErr
			},
		}
		nf, _ := NewNodeFacade(arg)

		_, _, err := nf.GetAccount("test", api.AccountQueryOptions{})
		require.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		getAccountCalled := false
		node := &mock.NodeStub{}
		node.GetAccountCalled = func(address string, _ api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
			getAccountCalled = true
			return api.AccountResponse{}, api.BlockInfo{}, nil
		}

		arg := createMockArguments()
		arg.Node = node
		nf, _ := NewNodeFacade(arg)

		_, _, _ = nf.GetAccount("test", api.AccountQueryOptions{})
		require.True(t, getAccountCalled)
	})
}

func TestNodeFacade_GetAccounts(t *testing.T) {
	t.Parallel()

	t.Run("too many addresses in bulk", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.WsAntifloodConfig.GetAddressesBulkMaxSize = 1
		nf, _ := NewNodeFacade(arg)

		resp, blockInfo, err := nf.GetAccounts([]string{"test1", "test2"}, api.AccountQueryOptions{})
		require.Nil(t, resp)
		require.Empty(t, blockInfo)
		require.Error(t, err)
		require.Equal(t, "too many addresses in the bulk request (provided: 2, maximum: 1)", err.Error())
	})

	t.Run("node responds with error, should err", func(t *testing.T) {
		t.Parallel()

		node := &mock.NodeStub{}
		node.GetAccountCalled = func(address string, _ api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
			return api.AccountResponse{}, api.BlockInfo{}, expectedErr
		}

		arg := createMockArguments()
		arg.Node = node
		arg.WsAntifloodConfig.GetAddressesBulkMaxSize = 2
		nf, _ := NewNodeFacade(arg)

		resp, blockInfo, err := nf.GetAccounts([]string{"test"}, api.AccountQueryOptions{})
		require.Nil(t, resp)
		require.Empty(t, blockInfo)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedAcount := api.AccountResponse{Address: "test"}
		node := &mock.NodeStub{}
		node.GetAccountCalled = func(address string, _ api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
			return expectedAcount, api.BlockInfo{}, nil
		}

		arg := createMockArguments()
		arg.Node = node
		arg.WsAntifloodConfig.GetAddressesBulkMaxSize = 1
		nf, _ := NewNodeFacade(arg)

		resp, blockInfo, err := nf.GetAccounts([]string{"test"}, api.AccountQueryOptions{})
		require.NoError(t, err)
		require.Empty(t, blockInfo)
		require.Equal(t, &expectedAcount, resp["test"])
	})
}

func TestNodeFacade_GetUsername(t *testing.T) {
	t.Parallel()

	expectedUsername := "username"
	node := &mock.NodeStub{}
	node.GetUsernameCalled = func(address string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
		return expectedUsername, api.BlockInfo{}, nil
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	username, _, err := nf.GetUsername("test", api.AccountQueryOptions{})
	require.NoError(t, err)
	require.Equal(t, expectedUsername, username)
}

func TestNodeFacade_GetCodeHash(t *testing.T) {
	t.Parallel()

	expectedCodeHash := []byte("hash")
	node := &mock.NodeStub{}
	node.GetCodeHashCalled = func(address string, _ api.AccountQueryOptions) ([]byte, api.BlockInfo, error) {
		return expectedCodeHash, api.BlockInfo{}, nil
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	codeHash, _, err := nf.GetCodeHash("test", api.AccountQueryOptions{})
	require.NoError(t, err)
	require.Equal(t, expectedCodeHash, codeHash)
}

func TestNodeFacade_GetHeartbeatsReturnsNilShouldErr(t *testing.T) {
	t.Parallel()

	node := &mock.NodeStub{
		GetHeartbeatsHandler: func() []data.PubKeyHeartbeat {
			return nil
		},
	}
	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	result, err := nf.GetHeartbeats()

	require.Nil(t, result)
	require.Equal(t, ErrHeartbeatsNotActive, err)
}

func TestNodeFacade_GetHeartbeats(t *testing.T) {
	t.Parallel()

	node := &mock.NodeStub{
		GetHeartbeatsHandler: func() []data.PubKeyHeartbeat {
			return []data.PubKeyHeartbeat{
				{
					PublicKey:       "pk1",
					TimeStamp:       time.Now(),
					IsActive:        true,
					ReceivedShardID: uint32(0),
				},
				{
					PublicKey:       "pk2",
					TimeStamp:       time.Now(),
					IsActive:        true,
					ReceivedShardID: uint32(0),
				},
			}
		},
	}
	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	result, err := nf.GetHeartbeats()

	require.NoError(t, err)
	fmt.Println(result)
}

func TestNodeFacade_GetDataValue(t *testing.T) {
	t.Parallel()

	wasCalled := false
	arg := createMockArguments()
	arg.ApiResolver = &mock.ApiResolverStub{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (*vmcommon.VMOutput, common.BlockInfo, error) {
			wasCalled = true
			return &vmcommon.VMOutput{}, nil, nil
		},
	}
	nf, err := NewNodeFacade(arg)
	require.NoError(t, err)

	_, _, _ = nf.ExecuteSCQuery(nil)
	require.True(t, wasCalled)
}

func TestNodeFacade_EmptyRestInterface(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.FacadeConfig.RestApiInterface = ""
	nf, _ := NewNodeFacade(arg)

	require.Equal(t, DefaultRestInterface, nf.RestApiInterface())
}

func TestNodeFacade_RestInterface(t *testing.T) {
	t.Parallel()

	intf := "localhost:1111"
	arg := createMockArguments()
	arg.FacadeConfig.RestApiInterface = intf
	nf, _ := NewNodeFacade(arg)

	require.Equal(t, intf, nf.RestApiInterface())
}

func TestNodeFacade_ValidatorStatisticsApi(t *testing.T) {
	t.Parallel()

	mapToRet := make(map[string]*accounts.ValidatorApiResponse)
	mapToRet["test"] = &accounts.ValidatorApiResponse{NumLeaderFailure: 5}
	node := &mock.NodeStub{
		ValidatorStatisticsApiCalled: func() (map[string]*accounts.ValidatorApiResponse, error) {
			return mapToRet, nil
		},
	}
	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	res, err := nf.ValidatorStatisticsApi()
	require.NoError(t, err)
	require.Equal(t, mapToRet, res)
}

func TestNodeFacade_SendBulkTransactions(t *testing.T) {
	t.Parallel()

	expectedNumOfSuccessfulTxs := uint64(1)
	sendBulkTxsWasCalled := false
	node := &mock.NodeStub{
		SendBulkTransactionsHandler: func(txs []*transaction.Transaction) (uint64, error) {
			sendBulkTxsWasCalled = true
			return expectedNumOfSuccessfulTxs, nil
		},
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	txs := make([]*transaction.Transaction, 0)
	txs = append(txs, &transaction.Transaction{Nonce: 1})

	res, err := nf.SendBulkTransactions(txs)
	require.NoError(t, err)
	require.Equal(t, expectedNumOfSuccessfulTxs, res)
	require.True(t, sendBulkTxsWasCalled)
}

func TestNodeFacade_StatusMetrics(t *testing.T) {
	t.Parallel()

	apiResolverMetricsRequested := false
	apiResStub := &mock.ApiResolverStub{
		StatusMetricsHandler: func() external.StatusMetricsHandler {
			apiResolverMetricsRequested = true
			return nil
		},
	}

	arg := createMockArguments()
	arg.ApiResolver = apiResStub
	nf, _ := NewNodeFacade(arg)

	_ = nf.StatusMetrics()

	require.True(t, apiResolverMetricsRequested)
}

func TestNodeFacade_PprofEnabled(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.FacadeConfig.PprofEnabled = true
	nf, _ := NewNodeFacade(arg)

	require.True(t, nf.PprofEnabled())
}

func TestNodeFacade_RestAPIServerDebugMode(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.RestAPIServerDebugMode = true
	nf, _ := NewNodeFacade(arg)

	require.True(t, nf.RestAPIServerDebugMode())
}

func TestNodeFacade_CreateTransaction(t *testing.T) {
	t.Parallel()

	nodeCreateTxWasCalled := false
	node := &mock.NodeStub{
		CreateTransactionHandler: func(txArgs *external.ArgsCreateTransaction) (*transaction.Transaction, []byte, error) {
			nodeCreateTxWasCalled = true
			return nil, nil, nil
		},
	}
	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	_, _, _ = nf.CreateTransaction(&external.ArgsCreateTransaction{})

	require.True(t, nodeCreateTxWasCalled)
}

func TestNodeFacade_Trigger(t *testing.T) {
	t.Parallel()

	wasCalled := false
	arg := createMockArguments()
	epoch := uint32(4638)
	recoveredEpoch := uint32(0)
	recoveredWithEarlyEndOfEpoch := atomicCore.Flag{}
	arg.Node = &mock.NodeStub{
		DirectTriggerCalled: func(epoch uint32, withEarlyEndOfEpoch bool) error {
			wasCalled = true
			atomic.StoreUint32(&recoveredEpoch, epoch)
			recoveredWithEarlyEndOfEpoch.SetValue(withEarlyEndOfEpoch)

			return expectedErr
		},
	}
	nf, _ := NewNodeFacade(arg)

	err := nf.Trigger(epoch, true)

	require.True(t, wasCalled)
	require.Equal(t, expectedErr, err)
	require.Equal(t, epoch, atomic.LoadUint32(&recoveredEpoch))
	require.True(t, recoveredWithEarlyEndOfEpoch.IsSet())
}

func TestNodeFacade_IsSelfTrigger(t *testing.T) {
	t.Parallel()

	wasCalled := false
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		IsSelfTriggerCalled: func() bool {
			wasCalled = true
			return true
		},
	}
	nf, _ := NewNodeFacade(arg)

	isSelf := nf.IsSelfTrigger()

	require.True(t, wasCalled)
	require.True(t, isSelf)
}

func TestNodeFacade_EncodeDecodeAddressPubkey(t *testing.T) {
	t.Parallel()

	buff := []byte("abcdefg")
	arg := createMockArguments()
	nf, _ := NewNodeFacade(arg)
	encoded, err := nf.EncodeAddressPubkey(buff)
	require.NoError(t, err)

	recoveredBytes, err := nf.DecodeAddressPubkey(encoded)

	require.NoError(t, err)
	require.Equal(t, buff, recoveredBytes)
}

func TestNodeFacade_GetQueryHandler(t *testing.T) {
	t.Parallel()

	wasCalled := false
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetQueryHandlerCalled: func(name string) (handler debug.QueryHandler, err error) {
			wasCalled = true

			return nil, nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	qh, err := nf.GetQueryHandler("")

	require.Nil(t, qh)
	require.NoError(t, err)
	require.True(t, wasCalled)
}

func TestNodeFacade_GetPeerInfo(t *testing.T) {
	t.Parallel()

	pinfo := core.QueryP2PPeerInfo{
		Pid: "pid",
	}
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetPeerInfoCalled: func(pid string) ([]core.QueryP2PPeerInfo, error) {
			return []core.QueryP2PPeerInfo{pinfo}, nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	val, err := nf.GetPeerInfo("")

	require.NoError(t, err)
	require.Equal(t, []core.QueryP2PPeerInfo{pinfo}, val)
}

func TestNodeFacade_GetThrottlerForEndpointAntifloodDisabledShouldReturnDisabled(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	nf, _ := NewNodeFacade(arg)

	thr, ok := nf.GetThrottlerForEndpoint("any-endpoint")
	require.NotNil(t, thr)
	require.True(t, ok)
	require.Equal(t, "*disabled.disabledThrottler", fmt.Sprintf("%T", thr))
}

func TestNodeFacade_GetThrottlerForEndpointNoConfigShouldReturnNilAndFalse(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.WsAntifloodConfig.WebServerAntifloodEnabled = true
	arg.WsAntifloodConfig.EndpointsThrottlers = []config.EndpointsThrottlersConfig{} // ensure it is empty
	nf, _ := NewNodeFacade(arg)

	thr, ok := nf.GetThrottlerForEndpoint("any-endpoint")

	require.Nil(t, thr)
	require.False(t, ok)
}

func TestNodeFacade_GetThrottlerForEndpointNotFoundShouldReturnNilAndFalse(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.WsAntifloodConfig.WebServerAntifloodEnabled = true
	arg.WsAntifloodConfig.EndpointsThrottlers = []config.EndpointsThrottlersConfig{
		{
			Endpoint:         "endpoint",
			MaxNumGoRoutines: 10,
		},
	}
	nf, _ := NewNodeFacade(arg)

	thr, ok := nf.GetThrottlerForEndpoint("different-endpoint")

	require.Nil(t, thr)
	require.False(t, ok)
}

func TestNodeFacade_GetThrottlerForEndpointShouldFindAndReturn(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.WsAntifloodConfig.WebServerAntifloodEnabled = true
	arg.WsAntifloodConfig.EndpointsThrottlers = []config.EndpointsThrottlersConfig{
		{
			Endpoint:         "endpoint",
			MaxNumGoRoutines: 10,
		},
	}
	nf, _ := NewNodeFacade(arg)

	thr, ok := nf.GetThrottlerForEndpoint("endpoint")

	require.NotNil(t, thr)
	require.True(t, ok)
}

func TestNodeFacade_GetKeyValuePairs(t *testing.T) {
	t.Parallel()

	expectedPairs := map[string]string{"k": "v"}
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetKeyValuePairsCalled: func(address string, _ api.AccountQueryOptions, _ context.Context) (map[string]string, api.BlockInfo, error) {
			return expectedPairs, api.BlockInfo{}, nil
		},
	}

	nf, _ := NewNodeFacade(arg)

	res, _, err := nf.GetKeyValuePairs("addr", api.AccountQueryOptions{})
	require.NoError(t, err)
	require.Equal(t, expectedPairs, res)
}

func TestNodeFacade_GetGuardianData(t *testing.T) {
	t.Parallel()
	arg := createMockArguments()

	emptyGuardianData := api.GuardianData{}
	testAddress := "test address"

	expectedGuardianData := api.GuardianData{
		ActiveGuardian: &api.Guardian{
			Address:         "guardian1",
			ActivationEpoch: 0,
		},
		PendingGuardian: &api.Guardian{
			Address:         "guardian2",
			ActivationEpoch: 10,
		},
		Guarded: true,
	}
	arg.Node = &mock.NodeStub{
		GetGuardianDataCalled: func(address string, options api.AccountQueryOptions) (api.GuardianData, api.BlockInfo, error) {
			if testAddress == address {
				return expectedGuardianData, api.BlockInfo{}, nil
			}
			return emptyGuardianData, api.BlockInfo{}, expectedErr
		},
	}

	t.Run("with error", func(t *testing.T) {
		nf, _ := NewNodeFacade(arg)
		res, _, err := nf.GetGuardianData("", api.AccountQueryOptions{})
		require.Equal(t, expectedErr, err)
		require.Equal(t, emptyGuardianData, res)
	})
	t.Run("ok", func(t *testing.T) {
		nf, _ := NewNodeFacade(arg)
		res, _, err := nf.GetGuardianData(testAddress, api.AccountQueryOptions{})
		require.NoError(t, err)
		require.Equal(t, expectedGuardianData, res)
	})
}

func TestNodeFacade_GetAllESDTTokens(t *testing.T) {
	t.Parallel()

	expectedTokens := map[string]*esdt.ESDigitalToken{
		"token0": {Value: big.NewInt(10)},
		"token1": {TokenMetaData: &esdt.MetaData{Name: []byte("name1")}},
	}
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetAllESDTTokensCalled: func(_ string, _ api.AccountQueryOptions, _ context.Context) (map[string]*esdt.ESDigitalToken, api.BlockInfo, error) {
			return expectedTokens, api.BlockInfo{}, nil
		},
	}

	nf, _ := NewNodeFacade(arg)

	res, _, err := nf.GetAllESDTTokens("addr", api.AccountQueryOptions{})
	require.NoError(t, err)
	require.Equal(t, expectedTokens, res)
}

func TestNodeFacade_GetESDTData(t *testing.T) {
	t.Parallel()

	expectedData := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{Name: []byte("name1")},
	}
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetESDTDataCalled: func(_ string, _ string, _ uint64, _ api.AccountQueryOptions) (*esdt.ESDigitalToken, api.BlockInfo, error) {
			return expectedData, api.BlockInfo{}, nil
		},
	}

	nf, _ := NewNodeFacade(arg)

	res, _, err := nf.GetESDTData("addr", "tkn", 0, api.AccountQueryOptions{})
	require.NoError(t, err)
	require.Equal(t, expectedData, res)
}

func TestNodeFacade_GetValueForKey(t *testing.T) {
	t.Parallel()

	expectedValue := "value"
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetValueForKeyCalled: func(_ string, _ string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
			return expectedValue, api.BlockInfo{}, nil
		},
	}

	nf, _ := NewNodeFacade(arg)

	res, _, err := nf.GetValueForKey("addr", "key", api.AccountQueryOptions{})
	require.NoError(t, err)
	require.Equal(t, expectedValue, res)
}

func TestNodeFacade_GetAllIssuedESDTs(t *testing.T) {
	t.Parallel()

	expectedValue := []string{"value"}
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetAllIssuedESDTsCalled: func(_ string, _ context.Context) ([]string, error) {
			return expectedValue, nil
		},
	}

	nf, _ := NewNodeFacade(arg)

	res, err := nf.GetAllIssuedESDTs("")
	require.NoError(t, err)
	require.Equal(t, expectedValue, res)
}

func TestNodeFacade_GetESDTsWithRole(t *testing.T) {
	t.Parallel()

	expectedResponse := []string{"ABC-1q2w3e", "PPP-sc78gh"}
	args := createMockArguments()

	args.Node = &mock.NodeStub{
		GetESDTsWithRoleCalled: func(address string, role string, _ api.AccountQueryOptions, _ context.Context) ([]string, api.BlockInfo, error) {
			return expectedResponse, api.BlockInfo{}, nil
		},
	}

	nf, _ := NewNodeFacade(args)

	res, _, err := nf.GetESDTsWithRole("address", "role", api.AccountQueryOptions{})
	require.NoError(t, err)
	require.Equal(t, expectedResponse, res)
}

func TestNodeFacade_GetNFTTokenIDsRegisteredByAddress(t *testing.T) {
	t.Parallel()

	expectedResponse := []string{"ABC-1q2w3e", "PPP-sc78gh"}
	args := createMockArguments()

	args.Node = &mock.NodeStub{
		GetNFTTokenIDsRegisteredByAddressCalled: func(address string, _ api.AccountQueryOptions, _ context.Context) ([]string, api.BlockInfo, error) {
			return expectedResponse, api.BlockInfo{}, nil
		},
	}

	nf, _ := NewNodeFacade(args)

	res, _, err := nf.GetNFTTokenIDsRegisteredByAddress("address", api.AccountQueryOptions{})
	require.NoError(t, err)
	require.Equal(t, expectedResponse, res)
}

func TestNodeFacade_GetAllIssuedESDTsWithError(t *testing.T) {
	t.Parallel()

	localErr := errors.New("local")
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetAllIssuedESDTsCalled: func(_ string, _ context.Context) ([]string, error) {
			return nil, localErr
		},
	}

	nf, _ := NewNodeFacade(arg)

	_, err := nf.GetAllIssuedESDTs("")
	require.Equal(t, err, localErr)
}

func TestNodeFacade_ValidateTransactionForSimulation(t *testing.T) {
	t.Parallel()

	called := false
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		ValidateTransactionForSimulationCalled: func(tx *transaction.Transaction, bypassSignature bool) error {
			called = true
			return nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	err := nf.ValidateTransactionForSimulation(&transaction.Transaction{}, false)
	require.NoError(t, err)
	require.True(t, called)
}

func TestNodeFacade_GetTotalStakedValue(t *testing.T) {
	t.Parallel()

	called := false
	arg := createMockArguments()
	arg.ApiResolver = &mock.ApiResolverStub{
		GetTotalStakedValueHandler: func(ctx context.Context) (*api.StakeValues, error) {
			called = true
			return nil, nil
		},
	}
	nf, _ := NewNodeFacade(arg)
	_, err := nf.GetTotalStakedValue()

	require.NoError(t, err)
	require.True(t, called)
}

func TestNodeFacade_GetDelegatorsList(t *testing.T) {
	t.Parallel()

	called := false
	arg := createMockArguments()
	arg.ApiResolver = &mock.ApiResolverStub{
		GetDelegatorsListHandler: func(ctx context.Context) ([]*api.Delegator, error) {
			called = true
			return nil, nil
		},
	}
	nf, _ := NewNodeFacade(arg)
	_, err := nf.GetDelegatorsList()

	require.NoError(t, err)
	require.True(t, called)
}

func TestNodeFacade_GetDirectStakedList(t *testing.T) {
	t.Parallel()

	called := false
	arg := createMockArguments()
	arg.ApiResolver = &mock.ApiResolverStub{
		GetDirectStakedListHandler: func(ctx context.Context) ([]*api.DirectStakedValue, error) {
			called = true
			return nil, nil
		},
	}
	nf, _ := NewNodeFacade(arg)
	_, err := nf.GetDirectStakedList()

	require.NoError(t, err)
	require.True(t, called)
}

func TestNodeFacade_GetProofCurrentRootHashIsEmptyShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.Blockchain = &testscommon.ChainHandlerStub{
		GetCurrentBlockRootHashCalled: func() []byte {
			return nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	response, err := nf.GetProofCurrentRootHash("addr")
	require.Nil(t, response)
	require.Equal(t, ErrEmptyRootHash, err)
}

func TestNodeFacade_GetProof(t *testing.T) {
	t.Parallel()

	expectedResponse := &common.GetProofResponse{
		Proof:    [][]byte{[]byte("valid"), []byte("proof")},
		Value:    []byte("value"),
		RootHash: "rootHash",
	}
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetProofCalled: func(_ string, _ string) (*common.GetProofResponse, error) {
			return expectedResponse, nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	response, err := nf.GetProof("hash", "addr")
	require.NoError(t, err)
	require.Equal(t, expectedResponse, response)
}

func TestNodeFacade_GetProofCurrentRootHash(t *testing.T) {
	t.Parallel()

	expectedResponse := &common.GetProofResponse{
		Proof:    [][]byte{[]byte("valid"), []byte("proof")},
		Value:    []byte("value"),
		RootHash: "rootHash",
	}
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetProofCalled: func(_ string, _ string) (*common.GetProofResponse, error) {
			return expectedResponse, nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	response, err := nf.GetProofCurrentRootHash("addr")
	require.NoError(t, err)
	require.Equal(t, expectedResponse, response)
}

func TestNodeFacade_GetProofDataTrie(t *testing.T) {
	t.Parallel()

	expectedResponseMainTrie := &common.GetProofResponse{
		Proof:    [][]byte{[]byte("valid"), []byte("proof"), []byte("mainTrie")},
		Value:    []byte("accountBytes"),
		RootHash: "rootHash",
	}
	expectedResponseDataTrie := &common.GetProofResponse{
		Proof:    [][]byte{[]byte("valid"), []byte("proof"), []byte("dataTrie")},
		Value:    []byte("value"),
		RootHash: "dataTrieRootHash",
	}
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetProofDataTrieCalled: func(_ string, _ string, _ string) (*common.GetProofResponse, *common.GetProofResponse, error) {
			return expectedResponseMainTrie, expectedResponseDataTrie, nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	mainTrieResponse, dataTrieResponse, err := nf.GetProofDataTrie("hash", "addr", "key")
	require.NoError(t, err)
	require.Equal(t, expectedResponseMainTrie, mainTrieResponse)
	require.Equal(t, expectedResponseDataTrie, dataTrieResponse)
}

func TestNodeFacade_VerifyProof(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		VerifyProofCalled: func(_ string, _ string, _ [][]byte) (bool, error) {
			return true, nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	response, err := nf.VerifyProof("hash", "addr", [][]byte{[]byte("proof")})
	require.NoError(t, err)
	require.True(t, response)
}

func TestNodeFacade_IsDataTrieMigrated(t *testing.T) {
	t.Parallel()

	t.Run("should return false if trie is not migrated", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.Node = &mock.NodeStub{
			IsDataTrieMigratedCalled: func(_ string, _ api.AccountQueryOptions) (bool, error) {
				return false, nil
			},
		}
		nf, _ := NewNodeFacade(arg)

		isMigrated, err := nf.IsDataTrieMigrated("address", api.AccountQueryOptions{})
		assert.Nil(t, err)
		assert.False(t, isMigrated)
	})

	t.Run("should return true if trie is migrated", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.Node = &mock.NodeStub{
			IsDataTrieMigratedCalled: func(_ string, _ api.AccountQueryOptions) (bool, error) {
				return true, nil
			},
		}
		nf, _ := NewNodeFacade(arg)

		isMigrated, err := nf.IsDataTrieMigrated("address", api.AccountQueryOptions{})
		assert.Nil(t, err)
		assert.True(t, isMigrated)
	})

	t.Run("should return error if node returns err", func(t *testing.T) {
		t.Parallel()

		expectedErr := fmt.Errorf(" expected error")
		arg := createMockArguments()
		arg.Node = &mock.NodeStub{
			IsDataTrieMigratedCalled: func(_ string, _ api.AccountQueryOptions) (bool, error) {
				return false, expectedErr
			},
		}
		nf, _ := NewNodeFacade(arg)

		isMigrated, err := nf.IsDataTrieMigrated("address", api.AccountQueryOptions{})
		assert.Equal(t, expectedErr, err)
		assert.False(t, isMigrated)
	})
}

func TestNodeFacade_ExecuteSCQuery(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.ApiResolver = &mock.ApiResolverStub{
			ExecuteSCQueryHandler: func(query *process.SCQuery) (*vmcommon.VMOutput, common.BlockInfo, error) {
				return nil, nil, expectedErr
			},
		}

		nf, _ := NewNodeFacade(arg)

		_, _, err := nf.ExecuteSCQuery(&process.SCQuery{})
		require.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		executeScQueryHandlerWasCalled := false
		arg := createMockArguments()

		expectedAddress := []byte("addr")
		expectedBalance := big.NewInt(37)
		expectedVmOutput := &vmcommon.VMOutput{
			ReturnData: [][]byte{[]byte("test return data")},
			ReturnCode: vmcommon.AccountCollision,
			OutputAccounts: map[string]*vmcommon.OutputAccount{
				"key0": {
					Address: expectedAddress,
					Balance: expectedBalance,
				},
			},
		}
		arg.ApiResolver = &mock.ApiResolverStub{
			ExecuteSCQueryHandler: func(query *process.SCQuery) (*vmcommon.VMOutput, common.BlockInfo, error) {
				executeScQueryHandlerWasCalled = true
				return expectedVmOutput, nil, nil
			},
		}

		nf, _ := NewNodeFacade(arg)

		apiVmOutput, _, err := nf.ExecuteSCQuery(&process.SCQuery{})
		require.NoError(t, err)
		require.True(t, executeScQueryHandlerWasCalled)
		require.Equal(t, expectedVmOutput.ReturnData, apiVmOutput.ReturnData)
		require.Equal(t, expectedVmOutput.ReturnCode.String(), apiVmOutput.ReturnCode)
		require.Equal(t, 1, len(apiVmOutput.OutputAccounts))

		outputAccount := apiVmOutput.OutputAccounts[hex.EncodeToString([]byte("key0"))]
		require.Equal(t, expectedBalance, outputAccount.Balance)
		require.Equal(t, hex.EncodeToString(expectedAddress), outputAccount.Address)
	})
}

func TestNodeFacade_GetBlockByRoundShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	blk := &api.Block{
		Round: 1,
		Nonce: 2,
	}

	arg.ApiResolver = &mock.ApiResolverStub{
		GetBlockByRoundCalled: func(_ uint64, _ api.BlockQueryOptions) (*api.Block, error) {
			return blk, nil
		},
	}

	nf, _ := NewNodeFacade(arg)
	ret, err := nf.GetBlockByRound(0, api.BlockQueryOptions{})

	require.NoError(t, err)
	require.Equal(t, ret, blk)
}

// ---- MetaBlock

func TestNodeFacade_GetInternalMetaBlockByNonceShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	blk := &block.MetaBlock{
		Round: 0,
		Nonce: 0,
	}

	arg.ApiResolver = &mock.ApiResolverStub{
		GetInternalMetaBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return blk, nil
		},
	}

	nf, _ := NewNodeFacade(arg)
	ret, err := nf.GetInternalMetaBlockByNonce(common.ApiOutputFormatProto, 0)

	require.NoError(t, err)
	require.Equal(t, ret, blk)
}

func TestNodeFacade_GetInternalMetaBlockByRoundShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	blk := &block.MetaBlock{
		Round: 0,
		Nonce: 0,
	}

	arg.ApiResolver = &mock.ApiResolverStub{
		GetInternalMetaBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return blk, nil
		},
	}

	nf, _ := NewNodeFacade(arg)
	ret, err := nf.GetInternalMetaBlockByRound(common.ApiOutputFormatProto, 0)

	require.NoError(t, err)
	require.Equal(t, ret, blk)
}

func TestNodeFacade_GetInternalMetaBlockByHashShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	blk := &block.MetaBlock{
		Round: 0,
		Nonce: 0,
	}

	arg.ApiResolver = &mock.ApiResolverStub{
		GetInternalMetaBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
			return blk, nil
		},
	}

	nf, _ := NewNodeFacade(arg)
	ret, err := nf.GetInternalMetaBlockByHash(common.ApiOutputFormatProto, "dummyhash")

	require.NoError(t, err)
	require.Equal(t, ret, blk)
}

// ---- ShardBlock

func TestNodeFacade_GetInternalShardBlockByNonceShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	blk := &block.Header{
		Round: 0,
		Nonce: 0,
	}

	arg.ApiResolver = &mock.ApiResolverStub{
		GetInternalShardBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return blk, nil
		},
	}

	nf, _ := NewNodeFacade(arg)
	ret, err := nf.GetInternalShardBlockByNonce(common.ApiOutputFormatProto, 0)

	require.NoError(t, err)
	require.Equal(t, ret, blk)
}

func TestNodeFacade_GetInternalShardBlockByRoundShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	blk := &block.Header{
		Round: 0,
		Nonce: 0,
	}

	arg.ApiResolver = &mock.ApiResolverStub{
		GetInternalShardBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return blk, nil
		},
	}

	nf, _ := NewNodeFacade(arg)
	ret, err := nf.GetInternalShardBlockByRound(common.ApiOutputFormatProto, 0)

	require.NoError(t, err)
	require.Equal(t, ret, blk)
}

func TestNodeFacade_GetInternalShardBlockByHashShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	blk := &block.Header{
		Round: 0,
		Nonce: 0,
	}

	arg.ApiResolver = &mock.ApiResolverStub{
		GetInternalShardBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
			return blk, nil
		},
	}

	nf, _ := NewNodeFacade(arg)
	ret, err := nf.GetInternalShardBlockByHash(common.ApiOutputFormatProto, "dummyhash")

	require.NoError(t, err)
	require.Equal(t, ret, blk)
}

func TestNodeFacade_GetInternalMiniBlockByHashShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	blk := &block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
	}

	arg.ApiResolver = &mock.ApiResolverStub{
		GetInternalMiniBlockCalled: func(_ common.ApiOutputFormat, _ string, epoch uint32) (interface{}, error) {
			return blk, nil
		},
	}

	nf, _ := NewNodeFacade(arg)
	ret, err := nf.GetInternalMiniBlockByHash(common.ApiOutputFormatProto, "dummyhash", 1)

	require.NoError(t, err)
	require.Equal(t, ret, blk)
}

func TestFacade_convertVmOutputToApiResponse(t *testing.T) {
	arg := createMockArguments()
	nf, _ := NewNodeFacade(arg)

	convertAddressFunc := func(input []byte) string {
		// NodeStub uses hex.EncodeToString for the stubbed EncodeAddressPubkey() function
		return hex.EncodeToString(input)
	}

	retData := [][]byte{[]byte("ret_data_0")}
	outAcc, outAccStorageKey, outAccOffset := []byte("addr0"), []byte("out_acc_storage_key"), []byte("offset")
	outAccTransferSndrAddr := []byte("addr1")
	logId, logAddr, logTopics, logData := []byte("log_id"), []byte("log_addr"), [][]byte{[]byte("log_topic")}, []byte("log_data")
	vmInput := vmcommon.VMOutput{
		ReturnData: retData,
		OutputAccounts: map[string]*vmcommon.OutputAccount{
			string(outAcc): {
				Address: outAcc,
				StorageUpdates: map[string]*vmcommon.StorageUpdate{
					string(outAccStorageKey): {
						Offset: outAccOffset,
					},
				},
				OutputTransfers: []vmcommon.OutputTransfer{
					{
						SenderAddress: outAccTransferSndrAddr,
					},
				},
			},
		},
		Logs: []*vmcommon.LogEntry{
			{
				Identifier: logId,
				Address:    logAddr,
				Topics:     logTopics,
				Data:       logData,
			},
		},
	}

	expectedOutputAccounts := map[string]*vm.OutputAccountApi{
		convertAddressFunc(outAcc): {
			Address: convertAddressFunc(outAcc),
			StorageUpdates: map[string]*vm.StorageUpdateApi{
				hex.EncodeToString(outAccStorageKey): {
					Offset: outAccOffset,
				},
			},
			OutputTransfers: []vm.OutputTransferApi{
				{
					SenderAddress: convertAddressFunc(outAccTransferSndrAddr),
				},
			},
		},
	}

	expectedLogs := []*vm.LogEntryApi{
		{
			Identifier: logId,
			Address:    convertAddressFunc(logAddr),
			Topics:     logTopics,
			Data:       logData,
		},
	}

	res := nf.convertVmOutputToApiResponse(&vmInput)
	require.Equal(t, retData, res.ReturnData)
	require.Equal(t, expectedOutputAccounts, res.OutputAccounts)
	require.Equal(t, expectedLogs, res.Logs)
}

func TestNodeFacade_GetTransactionsPool(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.ApiResolver = &mock.ApiResolverStub{
			GetTransactionsPoolCalled: func(fields string) (*common.TransactionsPoolAPIResponse, error) {
				return nil, expectedErr
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetTransactionsPool("")
		require.Nil(t, res)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		expectedPool := &common.TransactionsPoolAPIResponse{
			RegularTransactions: []common.Transaction{
				{
					TxFields: map[string]interface{}{
						"hash": "tx0",
					},
				},
				{
					TxFields: map[string]interface{}{
						"hash": "tx1",
					},
				},
			},
			SmartContractResults: []common.Transaction{
				{
					TxFields: map[string]interface{}{
						"hash": "tx2",
					},
				},
				{
					TxFields: map[string]interface{}{
						"hash": "tx3",
					},
				},
			},
			Rewards: []common.Transaction{
				{
					TxFields: map[string]interface{}{
						"hash": "tx4",
					},
				},
			},
		}
		arg.ApiResolver = &mock.ApiResolverStub{
			GetTransactionsPoolCalled: func(fields string) (*common.TransactionsPoolAPIResponse, error) {
				return expectedPool, nil
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetTransactionsPool("")
		require.NoError(t, err)
		require.Equal(t, expectedPool, res)
	})
}

func TestNodeFacade_GetGenesisNodesPubKeys(t *testing.T) {
	t.Parallel()

	t.Run("should return error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.ApiResolver = &mock.ApiResolverStub{
			GetGenesisNodesPubKeysCalled: func() (map[uint32][]string, map[uint32][]string) {
				return nil, nil
			},
		}

		nf, _ := NewNodeFacade(arg)
		eligible, waiting, err := nf.GetGenesisNodesPubKeys()
		require.Nil(t, eligible)
		require.Nil(t, waiting)
		require.Equal(t, ErrNilGenesisNodes, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedEligible := map[uint32][]string{
			0: {"pk1", "pk2"},
		}
		providedWaiting := map[uint32][]string{
			1: {"pk3", "pk4"},
		}

		arg := createMockArguments()
		arg.ApiResolver = &mock.ApiResolverStub{
			GetGenesisNodesPubKeysCalled: func() (map[uint32][]string, map[uint32][]string) {
				return providedEligible, providedWaiting
			},
		}

		nf, _ := NewNodeFacade(arg)
		eligible, waiting, err := nf.GetGenesisNodesPubKeys()
		require.NoError(t, err)
		require.Equal(t, providedEligible, eligible)
		require.Equal(t, providedWaiting, waiting)
	})
}

func TestNodeFacade_GetGenesisBalances(t *testing.T) {
	t.Parallel()

	t.Run("GetGenesisBalances error should return error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.ApiResolver = &mock.ApiResolverStub{
			GetGenesisBalancesCalled: func() ([]*common.InitialAccountAPI, error) {
				return nil, expectedErr
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetGenesisBalances()
		require.Nil(t, res)
		require.Equal(t, expectedErr, err)
	})
	t.Run("GetGenesisBalances returns empty initial accounts should return error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.ApiResolver = &mock.ApiResolverStub{
			GetGenesisBalancesCalled: func() ([]*common.InitialAccountAPI, error) {
				return nil, nil
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetGenesisBalances()
		require.Nil(t, res)
		require.Equal(t, ErrNilGenesisBalances, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		expectedBalances := []*common.InitialAccountAPI{
			{
				Address: "addr",
			},
		}
		arg.ApiResolver = &mock.ApiResolverStub{
			GetGenesisBalancesCalled: func() ([]*common.InitialAccountAPI, error) {
				return expectedBalances, nil
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetGenesisBalances()
		require.NoError(t, err)
		require.Equal(t, expectedBalances, res)
	})
}

func TestGetGasConfigs(t *testing.T) {
	t.Parallel()

	t.Run("empty gas configs map", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()

		wasCalled := false
		arg.ApiResolver = &mock.ApiResolverStub{
			GetGasConfigsCalled: func() map[string]map[string]uint64 {
				wasCalled = true
				return make(map[string]map[string]uint64)
			},
		}

		nf, _ := NewNodeFacade(arg)
		_, err := nf.GetGasConfigs()
		require.Equal(t, ErrEmptyGasConfigs, err)
		require.True(t, wasCalled)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()

		providedMap := map[string]map[string]uint64{
			"map1": {
				"test1": 1,
			},
		}
		wasCalled := false
		arg.ApiResolver = &mock.ApiResolverStub{
			GetGasConfigsCalled: func() map[string]map[string]uint64 {
				wasCalled = true
				return providedMap
			},
		}

		nf, _ := NewNodeFacade(arg)
		gasConfigMap, err := nf.GetGasConfigs()
		require.NoError(t, err)
		require.Equal(t, providedMap, gasConfigMap)
		require.True(t, wasCalled)
	})
}

func TestNodeFacade_GetTransactionsPoolForSender(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.ApiResolver = &mock.ApiResolverStub{
			GetTransactionsPoolForSenderCalled: func(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error) {
				return nil, expectedErr
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetTransactionsPoolForSender("", "")
		require.Nil(t, res)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		expectedSender := "alice"
		providedParameters := "sender,hash,receiver"
		expectedResponse := &common.TransactionsPoolForSenderApiResponse{
			Transactions: []common.Transaction{
				{
					TxFields: map[string]interface{}{
						"hash":     "txhash1",
						"sender":   expectedSender,
						"receiver": "receiver1",
					},
				},
				{
					TxFields: map[string]interface{}{
						"hash":     "txhash2",
						"sender":   expectedSender,
						"receiver": "receiver2",
					},
				},
			},
		}
		arg.ApiResolver = &mock.ApiResolverStub{
			GetTransactionsPoolForSenderCalled: func(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error) {
				require.Equal(t, expectedSender, sender)
				require.Equal(t, providedParameters, fields)
				return expectedResponse, nil
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetTransactionsPoolForSender(expectedSender, providedParameters)
		require.NoError(t, err)
		require.Equal(t, expectedResponse, res)
	})
}

func TestNodeFacade_GetLastPoolNonceForSender(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.ApiResolver = &mock.ApiResolverStub{
			GetLastPoolNonceForSenderCalled: func(sender string) (uint64, error) {
				return 0, expectedErr
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetLastPoolNonceForSender("")
		require.Equal(t, uint64(0), res)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		expectedSender := "alice"
		expectedNonce := uint64(33)
		arg.ApiResolver = &mock.ApiResolverStub{
			GetLastPoolNonceForSenderCalled: func(sender string) (uint64, error) {
				return expectedNonce, nil
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetLastPoolNonceForSender(expectedSender)
		require.NoError(t, err)
		require.Equal(t, expectedNonce, res)
	})
}

func TestNodeFacade_GetTransactionsPoolNonceGapsForSender(t *testing.T) {
	t.Parallel()

	t.Run("GetAccount error should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.Node = &mock.NodeStub{
			GetAccountCalled: func(address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
				return api.AccountResponse{}, api.BlockInfo{}, expectedErr
			},
		}
		arg.ApiResolver = &mock.ApiResolverStub{
			GetTransactionsPoolNonceGapsForSenderCalled: func(sender string, senderAccountNonce uint64) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
				require.Fail(t, "should have not been called")
				return nil, nil
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetTransactionsPoolNonceGapsForSender("")
		require.Equal(t, &common.TransactionsPoolNonceGapsForSenderApiResponse{}, res)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.Node = &mock.NodeStub{
			GetAccountCalled: func(address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
				return api.AccountResponse{}, api.BlockInfo{}, nil
			},
		}
		arg.ApiResolver = &mock.ApiResolverStub{
			GetTransactionsPoolNonceGapsForSenderCalled: func(sender string, senderAccountNonce uint64) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
				return nil, expectedErr
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetTransactionsPoolNonceGapsForSender("")
		require.Nil(t, res)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		expectedSender := "alice"
		expectedNonceGaps := &common.TransactionsPoolNonceGapsForSenderApiResponse{
			Sender: expectedSender,
			Gaps: []common.NonceGapApiResponse{
				{
					From: 33,
					To:   60,
				},
			},
		}
		providedNonce := uint64(10)
		arg.Node = &mock.NodeStub{
			GetAccountCalled: func(address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
				return api.AccountResponse{Nonce: providedNonce}, api.BlockInfo{}, nil
			},
		}
		arg.ApiResolver = &mock.ApiResolverStub{
			GetTransactionsPoolNonceGapsForSenderCalled: func(sender string, senderAccountNonce uint64) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
				require.Equal(t, providedNonce, senderAccountNonce)
				return expectedNonceGaps, nil
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetTransactionsPoolNonceGapsForSender(expectedSender)
		require.NoError(t, err)
		require.Equal(t, expectedNonceGaps, res)
	})
}

func TestNodeFacade_InternalValidatorsInfo(t *testing.T) {
	t.Parallel()

	t.Run("should fail on facade error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		arg.ApiResolver = &mock.ApiResolverStub{
			GetInternalStartOfEpochValidatorsInfoCalled: func(epoch uint32) ([]*state.ShardValidatorInfo, error) {
				return nil, expectedErr
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetInternalStartOfEpochValidatorsInfo(0)
		require.Nil(t, res)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()

		wasCalled := false
		arg.ApiResolver = &mock.ApiResolverStub{
			GetInternalStartOfEpochValidatorsInfoCalled: func(epoch uint32) ([]*state.ShardValidatorInfo, error) {
				wasCalled = true
				return make([]*state.ShardValidatorInfo, 0), nil
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetInternalStartOfEpochValidatorsInfo(0)
		require.NotNil(t, res)
		require.NoError(t, err)
		require.True(t, wasCalled)
	})
}

func TestNodeFacade_GetESDTsRoles(t *testing.T) {
	t.Parallel()

	expectedResponse := map[string][]string{
		"key": {"val1", "val2"},
	}
	args := createMockArguments()
	args.WsAntifloodConfig.WebServerAntifloodEnabled = true // coverage

	args.Node = &mock.NodeStub{
		GetESDTsRolesCalled: func(address string, options api.AccountQueryOptions, ctx context.Context) (map[string][]string, api.BlockInfo, error) {
			return expectedResponse, api.BlockInfo{}, nil
		},
	}

	nf, _ := NewNodeFacade(args)

	res, _, err := nf.GetESDTsRoles("address", api.AccountQueryOptions{})
	require.NoError(t, err)
	require.Equal(t, expectedResponse, res)
}

func TestNodeFacade_GetTokenSupply(t *testing.T) {
	t.Parallel()

	providedResponse := &api.ESDTSupply{
		Supply: "1000",
		Burned: "500",
		Minted: "1500",
	}
	args := createMockArguments()
	args.Node = &mock.NodeStub{
		GetTokenSupplyCalled: func(token string) (*api.ESDTSupply, error) {
			return providedResponse, nil
		},
	}

	nf, _ := NewNodeFacade(args)

	response, err := nf.GetTokenSupply("token")
	require.NoError(t, err)
	require.Equal(t, providedResponse, response)
}

func TestNodeFacade_ValidateTransaction(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	wasCalled := false
	args.Node = &mock.NodeStub{
		ValidateTransactionHandler: func(tx *transaction.Transaction) error {
			wasCalled = true
			return nil
		},
	}

	nf, _ := NewNodeFacade(args)

	err := nf.ValidateTransaction(&transaction.Transaction{})
	require.NoError(t, err)
	require.True(t, wasCalled)
}

func TestNodeFacade_SimulateTransactionExecution(t *testing.T) {
	t.Parallel()

	providedResponse := &txSimData.SimulationResultsWithVMOutput{
		SimulationResults: transaction.SimulationResults{
			Status:     "ok",
			FailReason: "no reason",
			ScResults:  nil,
			Receipts:   nil,
			Hash:       "hash",
		},
	}
	args := createMockArguments()
	args.ApiResolver = &mock.ApiResolverStub{
		SimulateTransactionExecutionHandler: func(tx *transaction.Transaction) (*txSimData.SimulationResultsWithVMOutput, error) {
			return providedResponse, nil
		},
	}

	nf, _ := NewNodeFacade(args)

	response, err := nf.SimulateTransactionExecution(&transaction.Transaction{})
	require.NoError(t, err)
	require.Equal(t, providedResponse, response)
}

func TestNodeFacade_ComputeTransactionGasLimit(t *testing.T) {
	t.Parallel()

	providedResponse := &transaction.CostResponse{
		GasUnits: 10,
	}
	args := createMockArguments()
	args.ApiResolver = &mock.ApiResolverStub{
		ComputeTransactionGasLimitHandler: func(tx *transaction.Transaction) (*transaction.CostResponse, error) {
			return providedResponse, nil
		},
	}

	nf, _ := NewNodeFacade(args)

	response, err := nf.ComputeTransactionGasLimit(&transaction.Transaction{})
	require.NoError(t, err)
	require.Equal(t, providedResponse, response)
}

func TestNodeFacade_GetEpochStartDataAPI(t *testing.T) {
	t.Parallel()

	providedResponse := &common.EpochStartDataAPI{
		Nonce:     1,
		Round:     2,
		Shard:     3,
		Timestamp: 4,
	}
	args := createMockArguments()
	args.Node = &mock.NodeStub{
		GetEpochStartDataAPICalled: func(epoch uint32) (*common.EpochStartDataAPI, error) {
			return providedResponse, nil
		},
	}
	nf, _ := NewNodeFacade(args)

	response, err := nf.GetEpochStartDataAPI(0)
	require.NoError(t, err)
	require.Equal(t, providedResponse, response)
}

func TestNodeFacade_GetConnectedPeersRatingsOnMainNetwork(t *testing.T) {
	t.Parallel()

	providedResponse := "ratings"
	args := createMockArguments()
	args.Node = &mock.NodeStub{
		GetConnectedPeersRatingsOnMainNetworkCalled: func() (string, error) {
			return providedResponse, nil
		},
	}
	nf, _ := NewNodeFacade(args)

	response, err := nf.GetConnectedPeersRatingsOnMainNetwork()
	require.NoError(t, err)
	require.Equal(t, providedResponse, response)
}

func TestNodeFacade_GetBlockByHash(t *testing.T) {
	t.Parallel()

	providedResponse := &api.Block{
		Nonce: 123,
		Round: 321,
	}
	args := createMockArguments()
	args.ApiResolver = &mock.ApiResolverStub{
		GetBlockByHashCalled: func(hash string, options api.BlockQueryOptions) (*api.Block, error) {
			return providedResponse, nil
		},
	}
	nf, _ := NewNodeFacade(args)

	response, err := nf.GetBlockByHash("hash", api.BlockQueryOptions{})
	require.NoError(t, err)
	require.Equal(t, providedResponse, response)
}

func TestNodeFacade_GetBlockByNonce(t *testing.T) {
	t.Parallel()

	providedResponse := &api.Block{
		Nonce: 123,
		Round: 321,
	}
	args := createMockArguments()
	args.ApiResolver = &mock.ApiResolverStub{
		GetBlockByNonceCalled: func(nonce uint64, options api.BlockQueryOptions) (*api.Block, error) {
			return providedResponse, nil
		},
	}
	nf, _ := NewNodeFacade(args)

	response, err := nf.GetBlockByNonce(0, api.BlockQueryOptions{})
	require.NoError(t, err)
	require.Equal(t, providedResponse, response)
}

func TestNodeFacade_GetAlteredAccountsForBlock(t *testing.T) {
	t.Parallel()

	providedResponse := []*alteredAccount.AlteredAccount{
		{
			Nonce:   123,
			Address: "address",
		},
	}
	args := createMockArguments()
	args.ApiResolver = &mock.ApiResolverStub{
		GetAlteredAccountsForBlockCalled: func(options api.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error) {
			return providedResponse, nil
		},
	}
	nf, _ := NewNodeFacade(args)

	response, err := nf.GetAlteredAccountsForBlock(api.GetAlteredAccountsForBlockOptions{})
	require.NoError(t, err)
	require.Equal(t, providedResponse, response)
}

func TestNodeFacade_GetInternalStartOfEpochMetaBlock(t *testing.T) {
	t.Parallel()

	providedResponse := "meta block"
	args := createMockArguments()
	args.ApiResolver = &mock.ApiResolverStub{
		GetInternalStartOfEpochMetaBlockCalled: func(format common.ApiOutputFormat, epoch uint32) (interface{}, error) {
			return providedResponse, nil
		},
	}
	nf, _ := NewNodeFacade(args)

	response, err := nf.GetInternalStartOfEpochMetaBlock(0, 0)
	require.NoError(t, err)
	require.Equal(t, providedResponse, response)
}

func TestNodeFacade_Close(t *testing.T) {
	t.Parallel()

	nf, _ := NewNodeFacade(createMockArguments())
	require.NoError(t, nf.Close())
}

func TestNodeFacade_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var nf *nodeFacade
	require.True(t, nf.IsInterfaceNil())

	nf, _ = NewNodeFacade(createMockArguments())
	require.False(t, nf.IsInterfaceNil())
}
