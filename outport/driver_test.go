package outport

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/outport/marshaling"
	"github.com/ElrondNetwork/elrond-go/outport/messages"
	"github.com/ElrondNetwork/elrond-go/outport/mock"
	"github.com/stretchr/testify/require"
)

func TestNewOutportDriver(t *testing.T) {
	config := createConfig()
	txCoordinator := mock.NewTxCoordinatorMock()
	logsProcessor := mock.NewTxLogsProcessorMock()
	marshalizer := marshaling.CreateMarshalizer(marshaling.JSON)
	sender := NewSender(nil, marshalizer)

	driver, err := newOutportDriver(config, nil, logsProcessor, sender)
	require.Nil(t, driver)
	require.Equal(t, ErrNilTxCoordinator, err)

	driver, err = newOutportDriver(config, txCoordinator, nil, sender)
	require.Nil(t, driver)
	require.Equal(t, ErrNilLogsProcessor, err)

	driver, err = newOutportDriver(config, txCoordinator, logsProcessor, nil)
	require.Nil(t, driver)
	require.Equal(t, ErrNilSender, err)
}

func TestOutportDriver_DigestCommittedBlock(t *testing.T) {
	config := createConfig()
	txCoordinator := mock.NewTxCoordinatorMock()
	logsProcessor := mock.NewTxLogsProcessorMock()
	sender := mock.NewSenderMock()

	driver, err := newOutportDriver(config, txCoordinator, logsProcessor, sender)
	require.Nil(t, err)
	require.NotNil(t, driver)

	txCoordinator.AddTx(block.TxBlock, "tx1", &transaction.Transaction{SndAddr: []byte("alice"), Nonce: 42})
	txCoordinator.AddTx(block.TxBlock, "tx2", &transaction.Transaction{SndAddr: []byte("alice"), Nonce: 43})
	txCoordinator.AddTx(block.TxBlock, "tx3", &transaction.Transaction{SndAddr: []byte("alice"), Nonce: 44})

	header := &block.Header{Nonce: 300}
	driver.DigestCommittedBlock([]byte("foo"), header)

	sentMessage := sender.GetLatestMessage().(*messages.MessageCommittedBlock)
	require.Equal(t, 300, int(sentMessage.Header.Nonce))
	require.Len(t, sentMessage.RegularTransactions.Keys, 3)
	require.Len(t, sentMessage.RegularTransactions.Values, 3)
}

func createConfig() config.OutportConfig {
	result := config.OutportConfig{}
	result.Enabled = true
	result.Filter = config.OutportFilterConfig{
		WithRegularTransactions:  true,
		WithSmartContractResults: true,
		WithRewardTransactions:   true,
		WithInvalidTransactions:  true,
		WithReceipts:             true,
		WithSmartContractLogs:    true,
	}

	return result
}
