package facade_test

import (
	"errors"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/cmd/facade"
	"github.com/ElrondNetwork/elrond-go-sandbox/cmd/facade/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestNewElrondFacade_FromValidNodeShouldReturnNotNil(t *testing.T) {
	node := &mock.NodeMock{}
	ef := facade.NewElrondNodeFacade(node)
	assert.NotNil(t, ef)
}

func TestNewElrondFacade_FromNullNodeShouldReturnNil(t *testing.T) {
	ef := facade.NewElrondNodeFacade(nil)
	assert.Nil(t, ef)
}

func TestElrondFacade_StartNode_WithNodeNotNull_ShouldNotReturnError(t *testing.T) {
	started := false
	node := &mock.NodeMock{
		StartCalled: func() error {
			started = true
			return nil
		},
		IsRunningCalled: func() bool {
			return started
		},
		ConnectToInitialAddressesCalled: func() error {
			return nil
		},
		StartConsensusCalled: func() error {
			return nil
		},
	}

	ef := facade.NewElrondNodeFacade(node)

	err := ef.StartNode()
	assert.Nil(t, err)

	isRunning := ef.IsNodeRunning()
	assert.True(t, isRunning)
}

func TestElrondFacade_StartNode_WithErrorOnStartNode_ShouldReturnError(t *testing.T) {
	started := false
	node := &mock.NodeMock{
		StartCalled: func() error {
			return fmt.Errorf("error on start node")
		},
		IsRunningCalled: func() bool {
			return started
		},
	}

	ef := facade.NewElrondNodeFacade(node)

	err := ef.StartNode()
	assert.NotNil(t, err)

	isRunning := ef.IsNodeRunning()
	assert.False(t, isRunning)
}

func TestElrondFacade_StartNode_WithErrorOnConnectToInitialAddresses_ShouldReturnError(t *testing.T) {
	started := false
	node := &mock.NodeMock{
		StartCalled: func() error {
			started = true
			return nil
		},
		IsRunningCalled: func() bool {
			return started
		},
		ConnectToInitialAddressesCalled: func() error {
			started = false
			return fmt.Errorf("error on connecting to initial addresses")
		},
	}

	ef := facade.NewElrondNodeFacade(node)

	err := ef.StartNode()
	assert.NotNil(t, err)

	isRunning := ef.IsNodeRunning()
	assert.False(t, isRunning)
}

func TestElrondFacade_StartNode_WithErrorOnStartConsensus_ShouldReturnError(t *testing.T) {
	started := false
	node := &mock.NodeMock{
		StartCalled: func() error {
			started = true
			return nil
		},
		IsRunningCalled: func() bool {
			return started
		},
		ConnectToInitialAddressesCalled: func() error {
			return nil
		},
		StartConsensusCalled: func() error {
			started = false
			return fmt.Errorf("error on StartConsensus")
		},
	}

	ef := facade.NewElrondNodeFacade(node)

	err := ef.StartNode()
	assert.NotNil(t, err)

	isRunning := ef.IsNodeRunning()
	assert.False(t, isRunning)
}

func TestElrondFacade_StopNode_WithNodeNotNull_ShouldNotReturnError(t *testing.T) {
	started := true
	node := &mock.NodeMock{
		StopCalled: func() error {
			started = false
			return nil
		},
		IsRunningCalled: func() bool {
			return started
		},
	}

	ef := facade.NewElrondNodeFacade(node)

	err := ef.StopNode()
	assert.Nil(t, err)

	isRunning := ef.IsNodeRunning()
	assert.False(t, isRunning)
}

func TestElrondFacade_StopNode_WithNodeNull_ShouldReturnError(t *testing.T) {
	started := true
	node := &mock.NodeMock{
		StopCalled: func() error {
			started = false
			return errors.New("failed to stop node")
		},
		IsRunningCalled: func() bool {
			return started
		},
	}

	ef := facade.NewElrondNodeFacade(node)

	err := ef.StopNode()
	assert.NotNil(t, err)

	isRunning := ef.IsNodeRunning()
	assert.False(t, isRunning)
}

func TestElrondFacade_GetBalance_WithValidAddress_ShouldReturnBalance(t *testing.T) {
	balance := big.NewInt(10)
	addr := "testAddress"
	node := &mock.NodeMock{
		GetBalanceCalled: func(address string) (*big.Int, error) {
			if addr == address {
				return balance, nil
			}
			return big.NewInt(0), nil
		},
	}

	ef := facade.NewElrondNodeFacade(node)

	amount, err := ef.GetBalance(addr)
	assert.Nil(t, err)
	assert.Equal(t, balance, amount)
}

func TestElrondFacade_GetBalance_WithUnknownAddress_ShouldReturnZeroBalance(t *testing.T) {
	balance := big.NewInt(10)
	addr := "testAddress"
	unknownAddr := "unknownAddr"
	zeroBalance := big.NewInt(0)

	node := &mock.NodeMock{
		GetBalanceCalled: func(address string) (*big.Int, error) {
			if addr == address {
				return balance, nil
			}
			return big.NewInt(0), nil
		},
	}

	ef := facade.NewElrondNodeFacade(node)

	amount, err := ef.GetBalance(unknownAddr)
	assert.Nil(t, err)
	assert.Equal(t, zeroBalance, amount)
}

func TestElrondFacade_GetBalance_WithErrorOnNode_ShouldReturnZeroBalanceAndError(t *testing.T) {
	addr := "testAddress"
	zeroBalance := big.NewInt(0)

	node := &mock.NodeMock{
		GetBalanceCalled: func(address string) (*big.Int, error) {
			return big.NewInt(0), errors.New("error on getBalance on node")
		},
	}

	ef := facade.NewElrondNodeFacade(node)

	amount, err := ef.GetBalance(addr)
	assert.NotNil(t, err)
	assert.Equal(t, zeroBalance, amount)
}

func TestElrondFacade_GenerateTransaction_WithCorrectInputs_ShouldReturnNoError(t *testing.T) {
	sender := "sender"
	receiver := "receiver"
	amount := *big.NewInt(10)
	code := "code"

	hash := sender + receiver + amount.String() + code

	node := &mock.NodeMock{
		GenerateTransactionCalled: func(sender string, receiver string, amount big.Int, code string) (string, error) {
			return sender + receiver + amount.String() + code, nil
		},
	}

	ef := facade.NewElrondNodeFacade(node)

	generatedHash, err := ef.GenerateTransaction(sender, receiver, amount, code)
	assert.Nil(t, err)
	assert.Equal(t, hash, generatedHash)
}

func TestElrondFacade_GenerateTransaction_WithNilSender_ShouldReturnError(t *testing.T) {
	receiver := "receiver"
	amount := *big.NewInt(10)
	code := "code"

	node := &mock.NodeMock{
		GenerateTransactionCalled: func(sender string, receiver string, amount big.Int, code string) (string, error) {
			if sender == "" {
				return "", errors.New("nil sender")
			}
			return sender + receiver + amount.String() + code, nil
		},
	}

	ef := facade.NewElrondNodeFacade(node)

	generatedHash, err := ef.GenerateTransaction("", receiver, amount, code)
	assert.NotNil(t, err)
	assert.Equal(t, "", generatedHash)
}

func TestElrondFacade_GenerateTransaction_WithNilReceiver_ShouldReturnError(t *testing.T) {
	sender := "sender"
	amount := *big.NewInt(10)
	code := "code"

	node := &mock.NodeMock{
		GenerateTransactionCalled: func(sender string, receiver string, amount big.Int, code string) (string, error) {
			if receiver == "" {
				return "", errors.New("nil receiver")
			}
			return sender + receiver + amount.String() + code, nil
		},
	}

	ef := facade.NewElrondNodeFacade(node)

	generatedHash, err := ef.GenerateTransaction(sender, "", amount, code)
	assert.NotNil(t, err)
	assert.Equal(t, "", generatedHash)
}

func TestElrondFacade_GenerateTransaction_WithZeroAmount_ShouldReturnError(t *testing.T) {
	sender := "sender"
	receiver := "receiver"
	amount := *big.NewInt(0)
	code := "code"

	node := &mock.NodeMock{
		GenerateTransactionCalled: func(sender string, receiver string, amount big.Int, code string) (string, error) {
			if amount.Cmp(big.NewInt(0)) == 0 {
				return "", errors.New("zero amount")
			}
			return sender + receiver + amount.String() + code, nil
		},
	}

	ef := facade.NewElrondNodeFacade(node)

	generatedHash, err := ef.GenerateTransaction(sender, receiver, amount, code)
	assert.NotNil(t, err)
	assert.Equal(t, "", generatedHash)
}

func TestElrondFacade_GenerateTransaction_WithNegativeAmount_ShouldReturnError(t *testing.T) {
	sender := "sender"
	receiver := "receiver"
	amount := *big.NewInt(-2)
	code := "code"

	node := &mock.NodeMock{
		GenerateTransactionCalled: func(sender string, receiver string, amount big.Int, code string) (string, error) {
			if amount.Cmp(big.NewInt(0)) < 0 {
				return "", errors.New("negative amount")
			}
			return sender + receiver + amount.String() + code, nil
		},
	}

	ef := facade.NewElrondNodeFacade(node)

	generatedHash, err := ef.GenerateTransaction(sender, receiver, amount, code)
	assert.NotNil(t, err)
	assert.Equal(t, "", generatedHash)
}

func TestElrondFacade_GetTransaction_WithValidInputs_ShouldNotReturnError(t *testing.T) {
	testHash := "testHash"
	testTx := &transaction.Transaction{}
	//testTx.
	node := &mock.NodeMock{
		GetTransactionCalled: func(hash string) (*transaction.Transaction, error) {
			if hash == testHash {
				return testTx, nil
			}
			return nil, nil
		},
	}

	ef := facade.NewElrondNodeFacade(node)

	tx, err := ef.GetTransaction(testHash)
	assert.Nil(t, err)
	assert.Equal(t, testTx, tx)
}

func TestElrondFacade_GetTransaction_WithUnknowHash_ShouldReturnNilAndNoError(t *testing.T) {
	testHash := "testHash"
	testTx := &transaction.Transaction{}
	node := &mock.NodeMock{
		GetTransactionCalled: func(hash string) (*transaction.Transaction, error) {
			if hash == testHash {
				return testTx, nil
			}
			return nil, nil
		},
	}

	ef := facade.NewElrondNodeFacade(node)

	tx, err := ef.GetTransaction("unknownHash")
	assert.Nil(t, err)
	assert.Nil(t, tx)
}
