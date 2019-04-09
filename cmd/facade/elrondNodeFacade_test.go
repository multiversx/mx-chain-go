package facade_test

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/cmd/facade"
	"github.com/ElrondNetwork/elrond-go-sandbox/cmd/facade/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/stretchr/testify/assert"
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

	ef := facade.NewElrondNodeFacade(node)

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

	ef := facade.NewElrondNodeFacade(node)

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

	ef := facade.NewElrondNodeFacade(node)

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

	ef := facade.NewElrondNodeFacade(node)

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

	ef := facade.NewElrondNodeFacade(node)

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

	ef := facade.NewElrondNodeFacade(node)

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

	ef := facade.NewElrondNodeFacade(node)

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

	ef := facade.NewElrondNodeFacade(node)

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

	ef := facade.NewElrondNodeFacade(node)

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

	ef := facade.NewElrondNodeFacade(node)

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

	ef := facade.NewElrondNodeFacade(node)

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

	ef := facade.NewElrondNodeFacade(node)

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

	ef := facade.NewElrondNodeFacade(node)

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

	ef := facade.NewElrondNodeFacade(node)

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

	ef := facade.NewElrondNodeFacade(node)

	tx, err := ef.GetTransaction("unknownHash")
	assert.Nil(t, err)
	assert.Nil(t, tx)
}

func TestElrondNodeFacade_SetLogger(t *testing.T) {
	node := &mock.NodeMock{}

	ef := facade.NewElrondNodeFacade(node)
	log := logger.DefaultLogger()
	ef.SetLogger(log)
	assert.Equal(t, log, ef.GetLogger())
}

func TestElrondNodeFacade_SetSyncer(t *testing.T) {
	node := &mock.NodeMock{}

	ef := facade.NewElrondNodeFacade(node)
	sync := &mock.SyncTimerMock{}
	ef.SetSyncer(sync)
	assert.Equal(t, sync, ef.GetSyncer())
}

func TestElrondNodeFacade_SendTransaction(t *testing.T) {
	called := 0
	node := &mock.NodeMock{}
	node.SendTransactionHandler = func(nonce uint64, sender string, receiver string, amount *big.Int, code string, signature []byte) (i *transaction.Transaction, e error) {
		called++
		return nil, nil
	}
	ef := facade.NewElrondNodeFacade(node)
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
	ef := facade.NewElrondNodeFacade(node)
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
	ef := facade.NewElrondNodeFacade(node)
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
	ef := facade.NewElrondNodeFacade(node)
	ef.GenerateAndSendBulkTransactions("", big.NewInt(0), 0)
	assert.Equal(t, called, 1)
}
