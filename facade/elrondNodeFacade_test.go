package facade

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/facade/mock"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/stretchr/testify/assert"
)

func createElrondNodeFacadeWithMockNodeAndResolver() *ElrondNodeFacade {
	return NewElrondNodeFacade(&mock.NodeMock{}, &mock.ExternalResolverStub{})
}

func createElrondNodeFacadeWithMockResolver(node *mock.NodeMock) *ElrondNodeFacade {
	return NewElrondNodeFacade(node, &mock.ExternalResolverStub{})
}

func TestNewElrondFacade_FromValidNodeShouldReturnNotNil(t *testing.T) {
	ef := createElrondNodeFacadeWithMockNodeAndResolver()
	assert.NotNil(t, ef)
}

func TestNewElrondFacade_FromNilNodeShouldReturnNil(t *testing.T) {
	ef := NewElrondNodeFacade(nil, &mock.ExternalResolverStub{})
	assert.Nil(t, ef)
}

func TestNewElrondFacade_FromNilExternalResolverShouldReturnNil(t *testing.T) {
	ef := NewElrondNodeFacade(&mock.NodeMock{}, nil)
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

func TestElrondFacade_StopNodeWithNodeNotNullShouldNotReturnError(t *testing.T) {
	started := true
	node := &mock.NodeMock{
		StopHandler: func() error {
			started = false
			return nil
		},
		IsRunningHandler: func() bool {
			return started
		},
	}

	ef := createElrondNodeFacadeWithMockResolver(node)

	err := ef.StopNode()
	assert.Nil(t, err)

	isRunning := ef.IsNodeRunning()
	assert.False(t, isRunning)
}

func TestElrondFacade_StopNodeWithNodeNullShouldReturnError(t *testing.T) {
	started := true
	node := &mock.NodeMock{
		StopHandler: func() error {
			started = false
			return errors.New("failed to stop node")
		},
		IsRunningHandler: func() bool {
			return started
		},
	}

	ef := createElrondNodeFacadeWithMockResolver(node)

	err := ef.StopNode()
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

func TestElrondFacade_GenerateTransactionWithCorrectInputsShouldReturnNoError(t *testing.T) {
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "code"

	tr := &transaction.Transaction{
		SndAddr: []byte(sender),
		RcvAddr: []byte(receiver),
		Data:    []byte(data),
		Value:   value}

	node := &mock.NodeMock{
		GenerateTransactionHandler: func(sender string, receiver string, value *big.Int,
			data string) (*transaction.Transaction, error) {
			return &transaction.Transaction{
				SndAddr: []byte(sender),
				RcvAddr: []byte(receiver),
				Data:    []byte(data),
				Value:   value,
			}, nil
		},
	}

	ef := createElrondNodeFacadeWithMockResolver(node)

	generatedTx, err := ef.GenerateTransaction(sender, receiver, value, data)
	assert.Nil(t, err)
	assert.Equal(t, tr, generatedTx)
}

func TestElrondFacade_GenerateTransactionWithNilSenderShouldReturnError(t *testing.T) {
	receiver := "receiver"
	amount := big.NewInt(10)
	code := "code"

	node := &mock.NodeMock{
		GenerateTransactionHandler: func(sender string, receiver string, amount *big.Int,
			code string) (*transaction.Transaction, error) {
			if sender == "" {
				return nil, errors.New("nil sender")
			}
			return nil, nil
		},
	}

	ef := createElrondNodeFacadeWithMockResolver(node)

	generatedTx, err := ef.GenerateTransaction("", receiver, amount, code)
	assert.NotNil(t, err)
	assert.Nil(t, generatedTx)
}

func TestElrondFacade_GenerateTransactionWithNilReceiverShouldReturnError(t *testing.T) {
	sender := "sender"
	amount := big.NewInt(10)
	code := "code"

	node := &mock.NodeMock{
		GenerateTransactionHandler: func(sender string, receiver string, amount *big.Int,
			code string) (*transaction.Transaction, error) {
			if receiver == "" {
				return nil, errors.New("nil receiver")
			}
			return nil, nil
		},
	}

	ef := createElrondNodeFacadeWithMockResolver(node)

	generatedTx, err := ef.GenerateTransaction(sender, "", amount, code)
	assert.NotNil(t, err)
	assert.Nil(t, generatedTx)
}

func TestElrondFacade_GenerateTransactionWithZeroAmountShouldReturnError(t *testing.T) {
	sender := "sender"
	receiver := "receiver"
	amount := big.NewInt(0)
	code := "code"

	node := &mock.NodeMock{
		GenerateTransactionHandler: func(sender string, receiver string, amount *big.Int,
			code string) (*transaction.Transaction, error) {
			if amount.Cmp(big.NewInt(0)) == 0 {
				return nil, errors.New("zero amount")
			}
			return nil, nil
		},
	}

	ef := createElrondNodeFacadeWithMockResolver(node)

	generatedTx, err := ef.GenerateTransaction(sender, receiver, amount, code)
	assert.NotNil(t, err)
	assert.Nil(t, generatedTx)
}

func TestElrondFacade_GenerateTransactionWithNegativeAmountShouldReturnError(t *testing.T) {
	sender := "sender"
	receiver := "receiver"
	amount := big.NewInt(-2)
	code := "code"

	node := &mock.NodeMock{
		GenerateTransactionHandler: func(sender string, receiver string, amount *big.Int,
			code string) (*transaction.Transaction, error) {
			if amount.Cmp(big.NewInt(0)) < 0 {
				return nil, errors.New("negative amount")
			}
			return nil, nil
		},
	}

	ef := createElrondNodeFacadeWithMockResolver(node)

	generatedTx, err := ef.GenerateTransaction(sender, receiver, amount, code)
	assert.NotNil(t, err)
	assert.Nil(t, generatedTx)
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
	node.SendTransactionHandler = func(nonce uint64, sender string, receiver string, amount *big.Int, code string, signature []byte) (string, error) {
		called++
		return "", nil
	}
	ef := createElrondNodeFacadeWithMockResolver(node)
	ef.SendTransaction(1, "test", "test", big.NewInt(0), "code", []byte{})
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
	ef.GetAccount("test")
	assert.Equal(t, called, 1)
}

func TestElrondNodeFacade_GetCurrentPublicKey(t *testing.T) {
	called := 0
	node := &mock.NodeMock{}
	node.GetCurrentPublicKeyHandler = func() string {
		called++
		return ""
	}
	ef := createElrondNodeFacadeWithMockResolver(node)
	ef.GetCurrentPublicKey()
	assert.Equal(t, called, 1)
}

func TestElrondNodeFacade_GenerateAndSendBulkTransactions(t *testing.T) {
	called := 0
	node := &mock.NodeMock{}
	node.GenerateAndSendBulkTransactionsHandler = func(destination string, value *big.Int, nrTransactions uint64) error {
		called++
		return nil
	}
	ef := createElrondNodeFacadeWithMockResolver(node)
	ef.GenerateAndSendBulkTransactions("", big.NewInt(0), 0)
	assert.Equal(t, called, 1)
}

func TestElrondNodeFacade_GenerateAndSendBulkTransactionsOneByOne(t *testing.T) {
	called := 0
	node := &mock.NodeMock{}
	node.GenerateAndSendBulkTransactionsOneByOneHandler = func(destination string, value *big.Int, nrTransactions uint64) error {
		called++
		return nil
	}
	ef := createElrondNodeFacadeWithMockResolver(node)
	ef.GenerateAndSendBulkTransactionsOneByOne("", big.NewInt(0), 0)
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
					HexPublicKey: "pk1",
					PeerHeartBeats: []heartbeat.PeerHeartbeat{
						{
							P2PAddress: "addr1",
							IsActive:   true,
						},
						{
							P2PAddress: "addr2",
						},
					},
				},
				{
					HexPublicKey: "pk2",
					PeerHeartBeats: []heartbeat.PeerHeartbeat{
						{
							P2PAddress: "addr3",
							IsActive:   true,
						},
						{
							P2PAddress: "addr4",
						},
					},
				},
			}
		},
	}
	ef := createElrondNodeFacadeWithMockResolver(node)

	result, err := ef.GetHeartbeats()

	assert.Nil(t, err)
	fmt.Println(result)
}
