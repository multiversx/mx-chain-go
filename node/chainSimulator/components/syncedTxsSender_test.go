package components

import (
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/process"
	processMock "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
)

func createMockSyncedTxsSenderArgs() ArgsSyncedTxsSender {
	return ArgsSyncedTxsSender{
		Marshaller:       &marshallerMock.MarshalizerMock{},
		ShardCoordinator: testscommon.NewMultiShardsCoordinatorMock(3),
		NetworkMessenger: &p2pmocks.MessengerStub{},
		DataPacker:       &mock.DataPackerStub{},
	}
}

func TestNewSyncedTxsSender(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSyncedTxsSenderArgs()
		args.Marshaller = nil
		sender, err := NewSyncedTxsSender(args)

		assert.Equal(t, process.ErrNilMarshalizer, err)
		assert.Nil(t, sender)
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSyncedTxsSenderArgs()
		args.ShardCoordinator = nil
		sender, err := NewSyncedTxsSender(args)

		assert.Equal(t, process.ErrNilShardCoordinator, err)
		assert.Nil(t, sender)
	})
	t.Run("nil network messenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSyncedTxsSenderArgs()
		args.NetworkMessenger = nil
		sender, err := NewSyncedTxsSender(args)

		assert.Equal(t, process.ErrNilMessenger, err)
		assert.Nil(t, sender)
	})
	t.Run("nil data packer should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSyncedTxsSenderArgs()
		args.DataPacker = nil
		sender, err := NewSyncedTxsSender(args)

		assert.Equal(t, dataRetriever.ErrNilDataPacker, err)
		assert.Nil(t, sender)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockSyncedTxsSenderArgs()
		sender, err := NewSyncedTxsSender(args)

		assert.Nil(t, err)
		assert.NotNil(t, sender)
	})
}

func TestSyncedTxsSender_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *syncedTxsSender
	assert.True(t, instance.IsInterfaceNil())

	instance = &syncedTxsSender{}
	assert.False(t, instance.IsInterfaceNil())
}

func TestSyncedTxsSender_Close(t *testing.T) {
	t.Parallel()

	args := createMockSyncedTxsSenderArgs()
	sender, _ := NewSyncedTxsSender(args)

	err := sender.Close()
	assert.Nil(t, err)
}

func TestSyncedTxsSender_SendBulkTransactions(t *testing.T) {
	t.Parallel()

	senderAShard0 := []byte("sender A shard 0")
	senderBShard1 := []byte("sender B shard 1")
	senderCShard0 := []byte("sender C shard 0")
	senderDShard1 := []byte("sender D shard 1")
	testTransactions := []*transaction.Transaction{
		{
			SndAddr: senderAShard0,
		},
		{
			SndAddr: senderBShard1,
		},
		{
			SndAddr: senderCShard0,
		},
		{
			SndAddr: senderDShard1,
		},
	}
	marshaller := &marshallerMock.MarshalizerMock{}

	marshalledTxs := make([][]byte, 0, len(testTransactions))
	for _, tx := range testTransactions {
		buff, _ := marshaller.Marshal(tx)
		marshalledTxs = append(marshalledTxs, buff)
	}

	mockShardCoordinator := &processMock.ShardCoordinatorStub{
		ComputeIdCalled: func(address []byte) uint32 {
			addrString := string(address)
			if strings.Contains(addrString, "shard 0") {
				return 0
			}
			if strings.Contains(addrString, "shard 1") {
				return 1
			}

			return core.MetachainShardId
		},
		SelfIdCalled: func() uint32 {
			return 1
		},
		CommunicationIdentifierCalled: func(destShardID uint32) string {
			if destShardID == 1 {
				return "_1"
			}
			if destShardID < 1 {
				return fmt.Sprintf("_%d_1", destShardID)
			}

			return fmt.Sprintf("_1_%d", destShardID)
		},
	}
	sentData := make(map[string][][]byte)
	netMessenger := &p2pmocks.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			sentData[topic] = append(sentData[topic], buff)
		},
	}
	mockDataPacker := &mock.DataPackerStub{
		PackDataInChunksCalled: func(data [][]byte, limit int) ([][]byte, error) {
			return data, nil
		},
	}

	t.Run("no transactions provided should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSyncedTxsSenderArgs()
		sender, _ := NewSyncedTxsSender(args)

		num, err := sender.SendBulkTransactions(nil)
		assert.Equal(t, process.ErrNoTxToProcess, err)
		assert.Zero(t, num)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := ArgsSyncedTxsSender{
			Marshaller:       marshaller,
			ShardCoordinator: mockShardCoordinator,
			NetworkMessenger: netMessenger,
			DataPacker:       mockDataPacker,
		}
		sender, _ := NewSyncedTxsSender(args)

		num, err := sender.SendBulkTransactions(testTransactions)
		assert.Nil(t, err)
		assert.Equal(t, uint64(4), num)

		expectedSentSliceForShard0 := make([][]byte, 0)
		expectedSentSliceForShard0 = append(expectedSentSliceForShard0, marshalledTxs[0])
		expectedSentSliceForShard0 = append(expectedSentSliceForShard0, marshalledTxs[2])

		expectedSentSliceForShard1 := make([][]byte, 0)
		expectedSentSliceForShard1 = append(expectedSentSliceForShard1, marshalledTxs[1])
		expectedSentSliceForShard1 = append(expectedSentSliceForShard1, marshalledTxs[3])

		expectedSentMap := map[string][][]byte{
			"transactions_1":   expectedSentSliceForShard1,
			"transactions_0_1": expectedSentSliceForShard0,
		}

		assert.Equal(t, expectedSentMap, sentData)

	})
}
