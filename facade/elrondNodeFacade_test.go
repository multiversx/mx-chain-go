package facade

import (
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/facade/mock"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func createElrondNodeFacadeWithMockNodeAndResolver() *ElrondNodeFacade {
	return NewElrondNodeFacade(&mock.NodeMock{}, &mock.ApiResolverStub{}, false)
}

func createElrondNodeFacadeWithMockResolver(node *mock.NodeMock) *ElrondNodeFacade {
	return NewElrondNodeFacade(node, &mock.ApiResolverStub{}, false)
}

func TestNewElrondFacade_FromValidNodeShouldReturnNotNil(t *testing.T) {
	ef := createElrondNodeFacadeWithMockNodeAndResolver()
	assert.NotNil(t, ef)
}

func TestNewElrondFacade_FromNilNodeShouldReturnNil(t *testing.T) {
	ef := NewElrondNodeFacade(nil, &mock.ApiResolverStub{}, false)
	assert.Nil(t, ef)
}

func TestNewElrondFacade_FromNilApiResolverShouldReturnNil(t *testing.T) {
	ef := NewElrondNodeFacade(&mock.NodeMock{}, nil, false)
	assert.Nil(t, ef)
}

func TestElrondFacade_StartNodeWithNodeNotNullShouldNotReturnError(t *testing.T) {
	started := false
	node := &mock.NodeMock{
		StartHandler: func() error {
			started = true
			return nil
		},
		P2PBootstrapHandler: func() error {
			return nil
		},
		IsRunningHandler: func() bool {
			return started
		},
		StartConsensusHandler: func() error {
			return nil
		},
	}

	ef := createElrondNodeFacadeWithMockResolver(node)

	err := ef.StartNode()
	assert.Nil(t, err)

	isRunning := ef.IsNodeRunning()
	assert.True(t, isRunning)
}

func TestElrondFacade_StartNodeWithErrorOnStartNodeShouldReturnError(t *testing.T) {
	started := false
	node := &mock.NodeMock{
		StartHandler: func() error {
			return fmt.Errorf("error on start node")
		},
		IsRunningHandler: func() bool {
			return started
		},
	}

	ef := createElrondNodeFacadeWithMockResolver(node)

	err := ef.StartNode()
	assert.NotNil(t, err)

	isRunning := ef.IsNodeRunning()
	assert.False(t, isRunning)
}

func TestElrondFacade_StartNodeWithErrorOnStartConsensusShouldReturnError(t *testing.T) {
	started := false
	node := &mock.NodeMock{
		StartHandler: func() error {
			started = true
			return nil
		},
		P2PBootstrapHandler: func() error {
			return nil
		},
		IsRunningHandler: func() bool {
			return started
		},
		StartConsensusHandler: func() error {
			started = false
			return fmt.Errorf("error on StartConsensus")
		},
	}

	ef := createElrondNodeFacadeWithMockResolver(node)

	err := ef.StartNode()
	assert.NotNil(t, err)

	isRunning := ef.IsNodeRunning()
	assert.False(t, isRunning)
}

func TestElrondFacade_GetBalanceWithValidAddressShouldReturnBalance(t *testing.T) {
	balance := big.NewInt(10)
	addr := "testAddress"
	node := &mock.NodeMock{
		GetBalanceHandler: func(address string) (*big.Int, error) {
			if addr == address {
				return balance, nil
			}
			return big.NewInt(0), nil
		},
	}

	ef := createElrondNodeFacadeWithMockResolver(node)

	amount, err := ef.GetBalance(addr)
	assert.Nil(t, err)
	assert.Equal(t, balance, amount)
}

func TestElrondFacade_GetBalanceWithUnknownAddressShouldReturnZeroBalance(t *testing.T) {
	balance := big.NewInt(10)
	addr := "testAddress"
	unknownAddr := "unknownAddr"
	zeroBalance := big.NewInt(0)

	node := &mock.NodeMock{
		GetBalanceHandler: func(address string) (*big.Int, error) {
			if addr == address {
				return balance, nil
			}
			return big.NewInt(0), nil
		},
	}

	ef := createElrondNodeFacadeWithMockResolver(node)

	amount, err := ef.GetBalance(unknownAddr)
	assert.Nil(t, err)
	assert.Equal(t, zeroBalance, amount)
}

func TestElrondFacade_GetBalanceWithErrorOnNodeShouldReturnZeroBalanceAndError(t *testing.T) {
	addr := "testAddress"
	zeroBalance := big.NewInt(0)

	node := &mock.NodeMock{
		GetBalanceHandler: func(address string) (*big.Int, error) {
			return big.NewInt(0), errors.New("error on getBalance on node")
		},
	}

	ef := createElrondNodeFacadeWithMockResolver(node)

	amount, err := ef.GetBalance(addr)
	assert.NotNil(t, err)
	assert.Equal(t, zeroBalance, amount)
}

func TestElrondFacade_GetTransactionWithValidInputsShouldNotReturnError(t *testing.T) {
	testHash := "testHash"
	testTx := &transaction.Transaction{}
	//testTx.
	node := &mock.NodeMock{
		GetTransactionHandler: func(hash string) (*transaction.Transaction, error) {
			if hash == testHash {
				return testTx, nil
			}
			return nil, nil
		},
	}

	ef := createElrondNodeFacadeWithMockResolver(node)

	tx, err := ef.GetTransaction(testHash)
	assert.Nil(t, err)
	assert.Equal(t, testTx, tx)
}

func TestElrondFacade_GetTransactionWithUnknowHashShouldReturnNilAndNoError(t *testing.T) {
	testHash := "testHash"
	testTx := &transaction.Transaction{}
	node := &mock.NodeMock{
		GetTransactionHandler: func(hash string) (*transaction.Transaction, error) {
			if hash == testHash {
				return testTx, nil
			}
			return nil, nil
		},
	}

	ef := createElrondNodeFacadeWithMockResolver(node)

	tx, err := ef.GetTransaction("unknownHash")
	assert.Nil(t, err)
	assert.Nil(t, tx)
}

func TestElrondNodeFacade_SetLogger(t *testing.T) {
	node := &mock.NodeMock{}

	ef := createElrondNodeFacadeWithMockResolver(node)
	log := logger.DefaultLogger()
	ef.SetLogger(log)
	assert.Equal(t, log, ef.GetLogger())
}

func TestElrondNodeFacade_SetSyncer(t *testing.T) {
	node := &mock.NodeMock{}

	ef := createElrondNodeFacadeWithMockResolver(node)
	sync := &mock.SyncTimerMock{}
	ef.SetSyncer(sync)
	assert.Equal(t, sync, ef.GetSyncer())
}

func TestElrondNodeFacade_SendTransaction(t *testing.T) {
	called := 0
	node := &mock.NodeMock{}
	node.SendTransactionHandler = func(nonce uint64, sender string, receiver string, amount string, code string, signature []byte) (string, error) {
		called++
		return "", nil
	}
	ef := createElrondNodeFacadeWithMockResolver(node)
	_, _ = ef.SendTransaction(1, "test", "test", "0", 0, 0, "code", []byte{})
	assert.Equal(t, called, 1)
}

func TestElrondNodeFacade_GetAccount(t *testing.T) {
	called := 0
	node := &mock.NodeMock{}
	node.GetAccountHandler = func(address string) (account *state.Account, e error) {
		called++
		return nil, nil
	}
	ef := createElrondNodeFacadeWithMockResolver(node)
	_, _ = ef.GetAccount("test")
	assert.Equal(t, called, 1)
}

func TestElrondNodeFacade_GetHeartbeatsReturnsNilShouldErr(t *testing.T) {
	node := &mock.NodeMock{
		GetHeartbeatsHandler: func() []heartbeat.PubKeyHeartbeat {
			return nil
		},
	}
	ef := createElrondNodeFacadeWithMockResolver(node)

	result, err := ef.GetHeartbeats()

	assert.Nil(t, result)
	assert.Equal(t, ErrHeartbeatsNotActive, err)
}

func TestElrondNodeFacade_GetHeartbeats(t *testing.T) {
	node := &mock.NodeMock{
		GetHeartbeatsHandler: func() []heartbeat.PubKeyHeartbeat {
			return []heartbeat.PubKeyHeartbeat{
				{
					HexPublicKey:    "pk1",
					TimeStamp:       time.Now(),
					MaxInactiveTime: heartbeat.Duration{Duration: 0},
					IsActive:        true,
					ReceivedShardID: uint32(0),
				},
				{
					HexPublicKey:    "pk2",
					TimeStamp:       time.Now(),
					MaxInactiveTime: heartbeat.Duration{Duration: 0},
					IsActive:        true,
					ReceivedShardID: uint32(0),
				},
			}
		},
	}
	ef := createElrondNodeFacadeWithMockResolver(node)

	result, err := ef.GetHeartbeats()

	assert.Nil(t, err)
	fmt.Println(result)
}

func TestElrondNodeFacade_GetDataValue(t *testing.T) {
	t.Parallel()

	wasCalled := false
	ef := NewElrondNodeFacade(
		&mock.NodeMock{},
		&mock.ApiResolverStub{
			ExecuteSCQueryHandler: func(query *process.SCQuery) (*vmcommon.VMOutput, error) {
				wasCalled = true
				return &vmcommon.VMOutput{}, nil
			},
		},
		false,
	)

	_, _ = ef.ExecuteSCQuery(nil)
	assert.True(t, wasCalled)
}

func TestElrondNodeFacade_RestApiPortNilConfig(t *testing.T) {
	ef := createElrondNodeFacadeWithMockNodeAndResolver()
	ef.SetConfig(nil)

	assert.Equal(t, DefaultRestPort, ef.RestApiPort())
}

func TestElrondNodeFacade_RestApiPortEmptyPortSpecified(t *testing.T) {
	ef := createElrondNodeFacadeWithMockNodeAndResolver()
	ef.SetConfig(&config.FacadeConfig{
		RestApiPort: "",
	})

	assert.Equal(t, DefaultRestPort, ef.RestApiPort())
}

func TestElrondNodeFacade_RestApiPortInvalidPortSpecified(t *testing.T) {
	ef := createElrondNodeFacadeWithMockNodeAndResolver()
	ef.SetConfig(&config.FacadeConfig{
		RestApiPort: "abc123",
	})

	assert.Equal(t, DefaultRestPort, ef.RestApiPort())
}

func TestElrondNodeFacade_RestApiPortCorrectPortSpecified(t *testing.T) {
	ef := createElrondNodeFacadeWithMockNodeAndResolver()
	port := "1111"
	ef.SetConfig(&config.FacadeConfig{
		RestApiPort: port,
	})

	assert.Equal(t, port, ef.RestApiPort())
}
