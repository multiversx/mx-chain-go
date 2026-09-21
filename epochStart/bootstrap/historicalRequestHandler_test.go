package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func TestHistoricalRequestHandler_EpochPairs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		install  func(*testscommon.RequestHandlerStub, func(uint32))
		request  func(process.RequestHandler)
		explicit func(process.RequestHandler, uint32)
	}{
		{name: "MiniBlock", install: func(stub *testscommon.RequestHandlerStub, record func(uint32)) {
			stub.RequestMiniBlockForEpochCalled = func(shardID uint32, hash []byte, epoch uint32) { record(epoch) }
		}, request: func(h process.RequestHandler) { h.RequestMiniBlock(2, []byte("data")) },
			explicit: func(h process.RequestHandler, epoch uint32) { h.RequestMiniBlockForEpoch(2, []byte("data"), epoch) }},
		{name: "MiniBlocks", install: func(stub *testscommon.RequestHandlerStub, record func(uint32)) {
			stub.RequestMiniBlocksForEpochCalled = func(shardID uint32, hashes [][]byte, epoch uint32) { record(epoch) }
		}, request: func(h process.RequestHandler) { h.RequestMiniBlocks(2, [][]byte{[]byte("data")}) },
			explicit: func(h process.RequestHandler, epoch uint32) {
				h.RequestMiniBlocksForEpoch(2, [][]byte{[]byte("data")}, epoch)
			}},
		{name: "ShardHeader", install: func(stub *testscommon.RequestHandlerStub, record func(uint32)) {
			stub.RequestShardHeaderForEpochCalled = func(shardID uint32, hash []byte, epoch uint32) { record(epoch) }
		}, request: func(h process.RequestHandler) { h.RequestShardHeader(2, []byte("data")) },
			explicit: func(h process.RequestHandler, epoch uint32) { h.RequestShardHeaderForEpoch(2, []byte("data"), epoch) }},
		{name: "MetaHeader", install: func(stub *testscommon.RequestHandlerStub, record func(uint32)) {
			stub.RequestMetaHeaderForEpochCalled = func(hash []byte, epoch uint32) { record(epoch) }
		}, request: func(h process.RequestHandler) { h.RequestMetaHeader([]byte("data")) },
			explicit: func(h process.RequestHandler, epoch uint32) { h.RequestMetaHeaderForEpoch([]byte("data"), epoch) }},
		{name: "ShardHeaderByNonce", install: func(stub *testscommon.RequestHandlerStub, record func(uint32)) {
			stub.RequestShardHeaderByNonceForEpochCalled = func(shardID uint32, nonce uint64, epoch uint32) { record(epoch) }
		}, request: func(h process.RequestHandler) { h.RequestShardHeaderByNonce(2, 42) },
			explicit: func(h process.RequestHandler, epoch uint32) { h.RequestShardHeaderByNonceForEpoch(2, 42, epoch) }},
		{name: "MetaHeaderByNonce", install: func(stub *testscommon.RequestHandlerStub, record func(uint32)) {
			stub.RequestMetaHeaderByNonceForEpochCalled = func(nonce uint64, epoch uint32) { record(epoch) }
		}, request: func(h process.RequestHandler) { h.RequestMetaHeaderByNonce(42) },
			explicit: func(h process.RequestHandler, epoch uint32) { h.RequestMetaHeaderByNonceForEpoch(42, epoch) }},
		{name: "Transactions", install: func(stub *testscommon.RequestHandlerStub, record func(uint32)) {
			stub.RequestTransactionsForEpochCalled = func(shardID uint32, hashes [][]byte, epoch uint32) { record(epoch) }
		}, request: func(h process.RequestHandler) { h.RequestTransactions(2, [][]byte{[]byte("data")}) },
			explicit: func(h process.RequestHandler, epoch uint32) {
				h.RequestTransactionsForEpoch(2, [][]byte{[]byte("data")}, epoch)
			}},
		{name: "UnsignedTransactions", install: func(stub *testscommon.RequestHandlerStub, record func(uint32)) {
			stub.RequestScrHandlerForEpochCalled = func(shardID uint32, hashes [][]byte, epoch uint32) { record(epoch) }
		}, request: func(h process.RequestHandler) { h.RequestUnsignedTransactions(2, [][]byte{[]byte("data")}) },
			explicit: func(h process.RequestHandler, epoch uint32) {
				h.RequestUnsignedTransactionsForEpoch(2, [][]byte{[]byte("data")}, epoch)
			}},
		{name: "RewardTransactions", install: func(stub *testscommon.RequestHandlerStub, record func(uint32)) {
			stub.RequestRewardTxHandlerForEpochCalled = func(shardID uint32, hashes [][]byte, epoch uint32) { record(epoch) }
		}, request: func(h process.RequestHandler) { h.RequestRewardTransactions(2, [][]byte{[]byte("data")}) },
			explicit: func(h process.RequestHandler, epoch uint32) {
				h.RequestRewardTransactionsForEpoch(2, [][]byte{[]byte("data")}, epoch)
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var epochs, forwarded []uint32
			stub := &testscommon.RequestHandlerStub{SetEpochCalled: func(epoch uint32) { forwarded = append(forwarded, epoch) }}
			tt.install(stub, func(epoch uint32) { epochs = append(epochs, epoch) })
			provider := &epochStartBootstrap{flagsConfig: config.ContextFlagsConfig{StartInEpochOffset: 2}, requestHandler: stub}
			provider.enableHistoricalRequests()
			h := provider.requestHandler
			for _, epoch := range []uint32{2240, 2239, 0} {
				h.SetEpoch(epoch)
				epochs = nil
				tt.request(h)
				expected := []uint32{epoch}
				if epoch > 0 {
					expected = append(expected, epoch-1)
				}
				require.Equal(t, expected, epochs)
			}
			require.Equal(t, []uint32{2240, 2239, 0}, forwarded)
			for _, epoch := range []uint32{2237, 0} {
				epochs = nil
				tt.explicit(h, epoch)
				expected := []uint32{epoch}
				if epoch > 0 {
					expected = append(expected, epoch-1)
				}
				require.Equal(t, expected, epochs)
			}
			// Explicit requests must not change the current request epoch.
			epochs = nil
			tt.request(h)
			require.Equal(t, []uint32{0}, epochs)
		})
	}
}

func TestHistoricalRequestHandler_NormalBootstrapUnchanged(t *testing.T) {
	t.Parallel()
	stub := &testscommon.RequestHandlerStub{}
	provider := &epochStartBootstrap{requestHandler: stub}
	provider.enableHistoricalRequests()
	require.Same(t, stub, provider.requestHandler)
}
