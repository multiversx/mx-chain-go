package facade

import (
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	atomicCore "github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/ElrondNetwork/elrond-go/facade/mock"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//TODO increase code coverage

func createMockArguments() ArgNodeFacade {
	return ArgNodeFacade{
		Node:                   &mock.NodeStub{},
		ApiResolver:            &mock.ApiResolverStub{},
		RestAPIServerDebugMode: false,
		TxSimulatorProcessor:   &mock.TxExecutionSimulatorStub{},
		WsAntifloodConfig: config.WebServerAntifloodConfig{
			SimultaneousRequests:         1,
			SameSourceRequests:           1,
			SameSourceResetIntervalInSec: 1,
		},
		FacadeConfig: config.FacadeConfig{
			RestApiInterface: "127.0.0.1:8080",
			PprofEnabled:     false,
		},
		ApiRoutesConfig: config.ApiRoutesConfig{APIPackages: map[string]config.APIPackageConfig{
			"node": {
				[]config.RouteConfig{
					{Name: "status"},
				},
			},
		}},
		AccountsState: &mock.AccountsStub{},
		PeerState:     &mock.AccountsStub{},
	}
}

//------- NewNodeFacade

func TestNewNodeFacade_WithNilNodeShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.Node = nil
	nf, err := NewNodeFacade(arg)

	assert.True(t, check.IfNil(nf))
	assert.Equal(t, ErrNilNode, err)
}

func TestNewNodeFacade_WithNilApiResolverShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.ApiResolver = nil
	nf, err := NewNodeFacade(arg)

	assert.True(t, check.IfNil(nf))
	assert.Equal(t, ErrNilApiResolver, err)
}

func TestNewNodeFacade_WithInvalidSimultaneousRequestsShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.WsAntifloodConfig.SimultaneousRequests = 0
	nf, err := NewNodeFacade(arg)

	assert.True(t, check.IfNil(nf))
	assert.True(t, errors.Is(err, ErrInvalidValue))
}

func TestNewNodeFacade_WithInvalidSameSourceResetIntervalInSecShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.WsAntifloodConfig.SameSourceResetIntervalInSec = 0
	nf, err := NewNodeFacade(arg)

	assert.True(t, check.IfNil(nf))
	assert.True(t, errors.Is(err, ErrInvalidValue))
}

func TestNewNodeFacade_WithInvalidSameSourceRequestsShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.WsAntifloodConfig.SameSourceRequests = 0
	nf, err := NewNodeFacade(arg)

	assert.True(t, check.IfNil(nf))
	assert.True(t, errors.Is(err, ErrInvalidValue))
}

func TestNewNodeFacade_WithInvalidApiRoutesConfigShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.ApiRoutesConfig = config.ApiRoutesConfig{}
	nf, err := NewNodeFacade(arg)

	assert.True(t, check.IfNil(nf))
	assert.True(t, errors.Is(err, ErrNoApiRoutesConfig))
}

func TestNewNodeFacade_WithValidNodeShouldReturnNotNil(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	nf, err := NewNodeFacade(arg)

	assert.False(t, check.IfNil(nf))
	assert.Nil(t, err)
}

//------- Methods

func TestNodeFacade_GetBalanceWithValidAddressShouldReturnBalance(t *testing.T) {
	t.Parallel()

	balance := big.NewInt(10)
	addr := "testAddress"
	node := &mock.NodeStub{
		GetBalanceHandler: func(address string) (*big.Int, error) {
			if addr == address {
				return balance, nil
			}
			return big.NewInt(0), nil
		},
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	amount, err := nf.GetBalance(addr)

	assert.Nil(t, err)
	assert.Equal(t, balance, amount)
}

func TestNodeFacade_GetBalanceWithUnknownAddressShouldReturnZeroBalance(t *testing.T) {
	t.Parallel()

	balance := big.NewInt(10)
	addr := "testAddress"
	unknownAddr := "unknownAddr"
	zeroBalance := big.NewInt(0)

	node := &mock.NodeStub{
		GetBalanceHandler: func(address string) (*big.Int, error) {
			if addr == address {
				return balance, nil
			}
			return big.NewInt(0), nil
		},
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	amount, err := nf.GetBalance(unknownAddr)
	assert.Nil(t, err)
	assert.Equal(t, zeroBalance, amount)
}

func TestNodeFacade_GetBalanceWithErrorOnNodeShouldReturnZeroBalanceAndError(t *testing.T) {
	t.Parallel()

	addr := "testAddress"
	zeroBalance := big.NewInt(0)

	node := &mock.NodeStub{
		GetBalanceHandler: func(address string) (*big.Int, error) {
			return big.NewInt(0), errors.New("error on getBalance on node")
		},
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	amount, err := nf.GetBalance(addr)
	assert.NotNil(t, err)
	assert.Equal(t, zeroBalance, amount)
}

func TestNodeFacade_GetTransactionWithValidInputsShouldNotReturnError(t *testing.T) {
	t.Parallel()

	testHash := "testHash"
	testTx := &transaction.ApiTransactionResult{}
	node := &mock.NodeStub{
		GetTransactionHandler: func(hash string, withEvents bool) (*transaction.ApiTransactionResult, error) {
			if hash == testHash {
				return testTx, nil
			}
			return nil, nil
		},
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	tx, err := nf.GetTransaction(testHash, false)
	assert.Nil(t, err)
	assert.Equal(t, testTx, tx)
}

func TestNodeFacade_SetAndGetTpsBenchmark(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	nf, _ := NewNodeFacade(arg)

	tpsBench, _ := statistics.NewTPSBenchmark(2, 5)
	nf.SetTpsBenchmark(tpsBench)
	assert.Equal(t, tpsBench, nf.TpsBenchmark())

}

func TestNodeFacade_GetTransactionWithUnknowHashShouldReturnNilAndNoError(t *testing.T) {
	t.Parallel()

	testHash := "testHash"
	testTx := &transaction.ApiTransactionResult{}
	node := &mock.NodeStub{
		GetTransactionHandler: func(hash string, withEvents bool) (*transaction.ApiTransactionResult, error) {
			if hash == testHash {
				return testTx, nil
			}
			return nil, nil
		},
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	tx, err := nf.GetTransaction("unknownHash", false)
	assert.Nil(t, err)
	assert.Nil(t, tx)
}

func TestNodeFacade_SetSyncer(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	nf, _ := NewNodeFacade(arg)

	sync := &mock.SyncTimerMock{}
	nf.SetSyncer(sync)
	assert.Equal(t, sync, nf.GetSyncer())
}

func TestNodeFacade_GetAccount(t *testing.T) {
	t.Parallel()

	called := 0
	node := &mock.NodeStub{}
	node.GetAccountHandler = func(address string) (state.UserAccountHandler, error) {
		called++
		return nil, nil
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	_, _ = nf.GetAccount("test")
	assert.Equal(t, called, 1)
}

func TestNodeFacade_GetUsername(t *testing.T) {
	t.Parallel()

	expectedUsername := "username"
	node := &mock.NodeStub{}
	node.GetUsernameCalled = func(address string) (string, error) {
		return expectedUsername, nil
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	username, err := nf.GetUsername("test")
	assert.NoError(t, err)
	assert.Equal(t, expectedUsername, username)
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

	assert.Nil(t, result)
	assert.Equal(t, ErrHeartbeatsNotActive, err)
}

func TestNodeFacade_GetHeartbeats(t *testing.T) {
	t.Parallel()

	node := &mock.NodeStub{
		GetHeartbeatsHandler: func() []data.PubKeyHeartbeat {
			return []data.PubKeyHeartbeat{
				{
					PublicKey:       "pk1",
					TimeStamp:       time.Now(),
					MaxInactiveTime: data.Duration{Duration: 0},
					IsActive:        true,
					ReceivedShardID: uint32(0),
				},
				{
					PublicKey:       "pk2",
					TimeStamp:       time.Now(),
					MaxInactiveTime: data.Duration{Duration: 0},
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

	assert.Nil(t, err)
	fmt.Println(result)
}

func TestNodeFacade_GetDataValue(t *testing.T) {
	t.Parallel()

	wasCalled := false
	arg := createMockArguments()
	arg.ApiResolver = &mock.ApiResolverStub{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (*vmcommon.VMOutput, error) {
			wasCalled = true
			return &vmcommon.VMOutput{}, nil
		},
	}
	nf, err := NewNodeFacade(arg)
	require.NoError(t, err)

	_, _ = nf.ExecuteSCQuery(nil)
	assert.True(t, wasCalled)
}

func TestNodeFacade_EmptyRestInterface(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.FacadeConfig.RestApiInterface = ""
	nf, _ := NewNodeFacade(arg)

	assert.Equal(t, DefaultRestInterface, nf.RestApiInterface())
}

func TestNodeFacade_RestInterface(t *testing.T) {
	t.Parallel()

	intf := "localhost:1111"
	arg := createMockArguments()
	arg.FacadeConfig.RestApiInterface = intf
	nf, _ := NewNodeFacade(arg)

	assert.Equal(t, intf, nf.RestApiInterface())
}

func TestNodeFacade_ValidatorStatisticsApi(t *testing.T) {
	t.Parallel()

	mapToRet := make(map[string]*state.ValidatorApiResponse)
	mapToRet["test"] = &state.ValidatorApiResponse{NumLeaderFailure: 5}
	node := &mock.NodeStub{
		ValidatorStatisticsApiCalled: func() (map[string]*state.ValidatorApiResponse, error) {
			return mapToRet, nil
		},
	}
	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	res, err := nf.ValidatorStatisticsApi()
	assert.Nil(t, err)
	assert.Equal(t, mapToRet, res)
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
	assert.Nil(t, err)
	assert.Equal(t, expectedNumOfSuccessfulTxs, res)
	assert.True(t, sendBulkTxsWasCalled)
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

	assert.True(t, apiResolverMetricsRequested)
}

func TestNodeFacade_PprofEnabled(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.FacadeConfig.PprofEnabled = true
	nf, _ := NewNodeFacade(arg)

	assert.True(t, nf.PprofEnabled())
}

func TestNodeFacade_RestAPIServerDebugMode(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.RestAPIServerDebugMode = true
	nf, _ := NewNodeFacade(arg)

	assert.True(t, nf.RestAPIServerDebugMode())
}

func TestNodeFacade_CreateTransaction(t *testing.T) {
	t.Parallel()

	nodeCreateTxWasCalled := false
	node := &mock.NodeStub{
		CreateTransactionHandler: func(_ uint64, _ string, _ string, _ string, _ uint64, _ uint64, _ []byte, _ string, _ string, _ uint32) (*transaction.Transaction, []byte, error) {
			nodeCreateTxWasCalled = true
			return nil, nil, nil
		},
	}
	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	_, _, _ = nf.CreateTransaction(0, "0", "0", "0", 0, 0, []byte("0"), "0", "chainID", 1)

	assert.True(t, nodeCreateTxWasCalled)
}

func TestNodeFacade_Trigger(t *testing.T) {
	t.Parallel()

	wasCalled := false
	expectedErr := errors.New("expected err")
	arg := createMockArguments()
	epoch := uint32(4638)
	recoveredEpoch := uint32(0)
	recoveredWithEarlyEndOfEpoch := atomicCore.Flag{}
	arg.Node = &mock.NodeStub{
		DirectTriggerCalled: func(epoch uint32, withEarlyEndOfEpoch bool) error {
			wasCalled = true
			atomic.StoreUint32(&recoveredEpoch, epoch)
			recoveredWithEarlyEndOfEpoch.Toggle(withEarlyEndOfEpoch)

			return expectedErr
		},
	}
	nf, _ := NewNodeFacade(arg)

	err := nf.Trigger(epoch, true)

	assert.True(t, wasCalled)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, epoch, atomic.LoadUint32(&recoveredEpoch))
	assert.True(t, recoveredWithEarlyEndOfEpoch.IsSet())
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

	assert.True(t, wasCalled)
	assert.True(t, isSelf)
}

func TestNodeFacade_EncodeDecodeAddressPubkey(t *testing.T) {
	t.Parallel()

	buff := []byte("abcdefg")
	arg := createMockArguments()
	nf, _ := NewNodeFacade(arg)
	encoded, err := nf.EncodeAddressPubkey(buff)
	assert.Nil(t, err)

	recoveredBytes, err := nf.DecodeAddressPubkey(encoded)

	assert.Nil(t, err)
	assert.Equal(t, buff, recoveredBytes)
}

func TestElrondNodeFacade_GetQueryHandler(t *testing.T) {
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

	assert.Nil(t, qh)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
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

	assert.Nil(t, err)
	assert.Equal(t, []core.QueryP2PPeerInfo{pinfo}, val)
}

func TestNodeFacade_GetThrottlerForEndpointNoConfigShouldReturnNilAndFalse(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.WsAntifloodConfig.EndpointsThrottlers = []config.EndpointsThrottlersConfig{} //ensure it is empty
	nf, _ := NewNodeFacade(arg)

	thr, ok := nf.GetThrottlerForEndpoint("any-endpoint")

	assert.Nil(t, thr)
	assert.False(t, ok)
}

func TestNodeFacade_GetThrottlerForEndpointNotFoundShouldReturnNilAndFalse(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.WsAntifloodConfig.EndpointsThrottlers = []config.EndpointsThrottlersConfig{
		{
			Endpoint:         "endpoint",
			MaxNumGoRoutines: 10,
		},
	}
	nf, _ := NewNodeFacade(arg)

	thr, ok := nf.GetThrottlerForEndpoint("different-endpoint")

	assert.Nil(t, thr)
	assert.False(t, ok)
}

func TestNodeFacade_GetThrottlerForEndpointShouldFindAndReturn(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.WsAntifloodConfig.EndpointsThrottlers = []config.EndpointsThrottlersConfig{
		{
			Endpoint:         "endpoint",
			MaxNumGoRoutines: 10,
		},
	}
	nf, _ := NewNodeFacade(arg)

	thr, ok := nf.GetThrottlerForEndpoint("endpoint")

	assert.NotNil(t, thr)
	assert.True(t, ok)
}
