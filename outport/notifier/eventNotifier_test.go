package notifier_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/outport/mock"
	"github.com/multiversx/mx-chain-go/outport/notifier"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	outportStub "github.com/multiversx/mx-chain-go/testscommon/outport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockEventNotifierArgs() notifier.ArgsEventNotifier {
	return notifier.ArgsEventNotifier{
		HttpClient:     &mock.HTTPClientStub{},
		Marshaller:     &marshallerMock.MarshalizerMock{},
		BlockContainer: &outportStub.BlockContainerStub{},
	}
}

func TestNewEventNotifier(t *testing.T) {
	t.Parallel()

	t.Run("nil http client", func(t *testing.T) {
		t.Parallel()

		args := createMockEventNotifierArgs()
		args.HttpClient = nil

		en, err := notifier.NewEventNotifier(args)
		require.Nil(t, en)
		require.Equal(t, notifier.ErrNilHTTPClientWrapper, err)
	})

	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		args := createMockEventNotifierArgs()
		args.Marshaller = nil

		en, err := notifier.NewEventNotifier(args)
		require.Nil(t, en)
		require.Equal(t, notifier.ErrNilMarshaller, err)
	})

	t.Run("nil block container", func(t *testing.T) {
		t.Parallel()

		args := createMockEventNotifierArgs()
		args.BlockContainer = nil

		en, err := notifier.NewEventNotifier(args)
		require.Nil(t, en)
		require.Equal(t, notifier.ErrNilBlockContainerHandler, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		en, err := notifier.NewEventNotifier(createMockEventNotifierArgs())
		require.Nil(t, err)
		require.NotNil(t, en)
	})
}

func TestSaveBlock(t *testing.T) {
	t.Parallel()

	args := createMockEventNotifierArgs()

	txHash1 := "txHash1"
	scrHash1 := "scrHash1"

	wasCalled := false
	args.HttpClient = &mock.HTTPClientStub{
		PostCalled: func(route string, payload interface{}) error {
			saveBlockData := payload.(*outport.OutportBlock)

			require.Equal(t, saveBlockData.TransactionPool.Logs[0].TxHash, txHash1)
			for txHash := range saveBlockData.TransactionPool.Transactions {
				require.Equal(t, txHash1, txHash)
			}

			for scrHash := range saveBlockData.TransactionPool.SmartContractResults {
				require.Equal(t, scrHash1, scrHash)
			}

			wasCalled = true
			return nil
		},
	}

	en, _ := notifier.NewEventNotifier(args)

	saveBlockData := &outport.OutportBlock{
		BlockData: &outport.BlockData{
			HeaderHash: []byte{},
		},
		TransactionPool: &outport.TransactionPool{
			Transactions: map[string]*outport.TxInfo{
				txHash1: nil,
			},
			SmartContractResults: map[string]*outport.SCRInfo{
				scrHash1: nil,
			},
			Logs: []*outport.LogData{
				{
					TxHash: txHash1,
					Log:    &transaction.Log{},
				},
			},
		},
	}

	err := en.SaveBlock(saveBlockData)
	require.Nil(t, err)

	require.True(t, wasCalled)
}

func TestRevertIndexedBlock(t *testing.T) {
	t.Parallel()

	args := createMockEventNotifierArgs()

	header := &block.Header{
		Nonce: 1,
		Round: 2,
		Epoch: 3,
	}
	headerBytes, _ := args.Marshaller.Marshal(header)

	blockData := &outport.BlockData{
		HeaderBytes: headerBytes,
		Body:        &block.Body{},
		HeaderType:  string(core.ShardHeaderV1),
	}

	wasCalled := false
	args.HttpClient = &mock.HTTPClientStub{
		PostCalled: func(route string, payload interface{}) error {
			require.Equal(t, blockData, payload)
			wasCalled = true
			return nil
		},
	}
	en, _ := notifier.NewEventNotifier(args)

	err := en.RevertIndexedBlock(blockData)
	require.Nil(t, err)
	require.True(t, wasCalled)
}

func TestFinalizedBlock(t *testing.T) {
	t.Parallel()

	args := createMockEventNotifierArgs()

	wasCalled := false
	args.HttpClient = &mock.HTTPClientStub{
		PostCalled: func(route string, payload interface{}) error {
			wasCalled = true
			return nil
		},
	}

	en, _ := notifier.NewEventNotifier(args)

	hash := []byte("headerHash")
	err := en.FinalizedBlock(&outport.FinalizedBlock{HeaderHash: hash})
	require.Nil(t, err)

	require.True(t, wasCalled)
}

func TestMockFunctions(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	en, err := notifier.NewEventNotifier(createMockEventNotifierArgs())
	require.Nil(t, err)
	require.False(t, en.IsInterfaceNil())

	err = en.SaveRoundsInfo(nil)
	require.Nil(t, err)

	err = en.SaveValidatorsRating(nil)
	require.Nil(t, err)

	err = en.SaveValidatorsPubKeys(nil)
	require.Nil(t, err)

	err = en.SaveAccounts(nil)
	require.Nil(t, err)

	err = en.Close()
	require.Nil(t, err)
}
