package headerForBlock_test

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	retriever "github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/headerForBlock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
	"github.com/stretchr/testify/require"
)

func createMockArgs() headerForBlock.ArgHeadersForBlock {
	return headerForBlock.ArgHeadersForBlock{
		DataPool:            dataRetriever.NewPoolsHolderMock(),
		RequestHandler:      &testscommon.RequestHandlerStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ShardCoordinator:    &testscommon.ShardsCoordinatorMock{},
		BlockTracker:        &mock.BlockTrackerStub{},
		TxCoordinator:       &testscommon.TransactionCoordinatorMock{},
		RoundHandler:        &testscommon.RoundHandlerMock{},
		ExtraDelayForRequestBlockInfoInMilliseconds: 10,
		GenesisNonce: 12345,
	}
}

func TestNewHeadersForBlock(t *testing.T) {
	t.Parallel()

	t.Run("nil DataPool", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.DataPool = nil

		hfb, err := headerForBlock.NewHeadersForBlock(args)
		require.Equal(t, process.ErrNilDataPoolHolder, err)
		require.Nil(t, hfb)
	})
	t.Run("nil Headers", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.DataPool = &dataRetriever.PoolsHolderStub{}

		hfb, err := headerForBlock.NewHeadersForBlock(args)
		require.Equal(t, process.ErrNilHeadersDataPool, err)
		require.Nil(t, hfb)
	})
	t.Run("nil Proofs", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.DataPool = &dataRetriever.PoolsHolderStub{
			HeadersCalled: func() retriever.HeadersPool {
				return &pool.HeadersPoolStub{}
			},
		}

		hfb, err := headerForBlock.NewHeadersForBlock(args)
		require.Equal(t, process.ErrNilProofsPool, err)
		require.Nil(t, hfb)
	})
	t.Run("nil RequestHandler", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RequestHandler = nil

		hfb, err := headerForBlock.NewHeadersForBlock(args)
		require.Equal(t, process.ErrNilRequestHandler, err)
		require.Nil(t, hfb)
	})
	t.Run("nil EnableEpochsHandler", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.EnableEpochsHandler = nil

		hfb, err := headerForBlock.NewHeadersForBlock(args)
		require.Equal(t, process.ErrNilEnableEpochsHandler, err)
		require.Nil(t, hfb)
	})
	t.Run("nil ShardCoordinator", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.ShardCoordinator = nil

		hfb, err := headerForBlock.NewHeadersForBlock(args)
		require.Equal(t, process.ErrNilShardCoordinator, err)
		require.Nil(t, hfb)
	})
	t.Run("nil BlockTracker", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.BlockTracker = nil

		hfb, err := headerForBlock.NewHeadersForBlock(args)
		require.Equal(t, process.ErrNilBlockTracker, err)
		require.Nil(t, hfb)
	})
	t.Run("nil TxCoordinator", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.TxCoordinator = nil

		hfb, err := headerForBlock.NewHeadersForBlock(args)
		require.Equal(t, process.ErrNilTransactionCoordinator, err)
		require.Nil(t, hfb)
	})
	t.Run("nil RoundHandler", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RoundHandler = nil

		hfb, err := headerForBlock.NewHeadersForBlock(args)
		require.Error(t, err)
		require.Nil(t, hfb)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		hfb, err := headerForBlock.NewHeadersForBlock(createMockArgs())
		require.NoError(t, err)
		require.NotNil(t, hfb)
	})
}

func TestHeadersForBlock_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	args.DataPool = nil
	hfb, _ := headerForBlock.NewHeadersForBlock(args)
	require.True(t, hfb.IsInterfaceNil())

	hfb, _ = headerForBlock.NewHeadersForBlock(createMockArgs())
	require.False(t, hfb.IsInterfaceNil())
}

func TestHeadersForBlock_AddHeaderUsedInBlock(t *testing.T) {
	t.Parallel()

	hfb, err := headerForBlock.NewHeadersForBlock(createMockArgs())
	require.NoError(t, err)

	hfb.AddHeaderUsedInBlock(
		"hash1",
		&testscommon.HeaderHandlerStub{},
	)

	hi, found := hfb.GetHeaderInfo("hash1")
	require.True(t, found)
	require.NotNil(t, hi.GetHeader())
}

func TestHeadersForBlock_GetHeadersInfoMap(t *testing.T) {
	t.Parallel()

	hfb, err := headerForBlock.NewHeadersForBlock(createMockArgs())
	require.NoError(t, err)

	hfb.AddHeaderUsedInBlock(
		"hash1",
		&testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 1
			},
		},
	)
	hfb.AddHeaderUsedInBlock(
		"hash2",
		&testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 2
			},
		},
	)

	infoMap := hfb.GetHeadersInfoMap()
	require.Len(t, infoMap, 2)
	info, ok := infoMap["hash1"]
	require.True(t, ok)
	require.Equal(t, uint64(1), info.GetHeader().GetNonce())
	info, ok = infoMap["hash2"]
	require.True(t, ok)
	require.Equal(t, uint64(2), info.GetHeader().GetNonce())
}

func TestHeadersForBlock_GetHeadersMap(t *testing.T) {
	t.Parallel()

	hfb, err := headerForBlock.NewHeadersForBlock(createMockArgs())
	require.NoError(t, err)

	hfb.AddHeaderUsedInBlock(
		"hash1",
		&testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 1
			},
		},
	)
	hfb.AddHeaderUsedInBlock(
		"hash2",
		&testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 2
			},
		},
	)

	headersMap := hfb.GetHeadersMap()
	require.Len(t, headersMap, 2)
	header, ok := headersMap["hash1"]
	require.True(t, ok)
	require.Equal(t, uint64(1), header.GetNonce())
	header, ok = headersMap["hash2"]
	require.True(t, ok)
	require.Equal(t, uint64(2), header.GetNonce())
}

func TestHeadersForBlock_Reset(t *testing.T) {
	t.Parallel()

	hfb, err := headerForBlock.NewHeadersForBlock(createMockArgs())
	require.NoError(t, err)

	hfb.AddHeaderUsedInBlock("hash1", &testscommon.HeaderHandlerStub{})
	hfb.AddHeaderUsedInBlock("hash2", &testscommon.HeaderHandlerStub{})

	hfb.Reset()

	hfbMap := hfb.GetHeadersInfoMap()
	require.Empty(t, hfbMap)
}

func TestHeadersForBlock_RequestAndWaitHeaders(t *testing.T) {
	t.Parallel()

	t.Run("request meta headers should work", testRequestAndWaitHeaders(true))
	t.Run("request shard headers should work", testRequestAndWaitHeaders(false))
}

func testRequestAndWaitHeaders(requestMetaHeaders bool) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		// - one header exists in cache with proof (idx 1)
		// - one header exists in cache without proof (idx 2)
		// - one proof exists in cache without header (idx 3)
		// - one header is missing completely (idx 4)

		td := createTestData(5, requestMetaHeaders)

		args := createMockArgs()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return true // return true so finality is based on proofs
			},
		}
		poolsHolder, ok := args.DataPool.(*dataRetriever.PoolsHolderMock)
		require.True(t, ok)

		headersPoolStub := createPoolsHolderForHeaderRequests()
		poolsHolder.SetHeadersPool(headersPoolStub)
		headersPool := poolsHolder.Headers()
		// adding the existing headers
		headersPool.AddHeader(td[1].headerHash, td[1].header)
		headersPool.AddHeader(td[2].headerHash, td[2].header)

		proofsPoolStub := createProofsPoolForHeaderRequests()
		poolsHolder.SetProofsPool(proofsPoolStub)
		proofsPool := poolsHolder.Proofs()
		// adding existing proofs
		proofsPool.AddProof(&block.HeaderProof{
			HeaderHash: td[1].headerHash,
		})
		proofsPool.AddProof(&block.HeaderProof{
			HeaderHash: td[3].headerHash,
		})

		args.BlockTracker = &mock.BlockTrackerStub{
			GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
				return td[0].header, nil, nil
			},
		}

		hfb, err := headerForBlock.NewHeadersForBlock(args)
		require.NoError(t, err)

		if requestMetaHeaders {
			referencedMetaHeaders := [][]byte{td[1].headerHash, td[2].headerHash, td[3].headerHash, td[4].headerHash}
			header := &block.HeaderV2{
				Header: &block.Header{
					MetaBlockHashes: referencedMetaHeaders,
				},
			}
			hfb.RequestMetaHeaders(&block.HeaderV2{}) // for coverage only, no shard data
			hfb.RequestMetaHeaders(header)
		} else {
			referencedHeaders := []*headerData{td[1], td[2], td[3], td[4]}
			shardInfo := createShardInfo(referencedHeaders)
			metaBlock := &block.MetaBlock{
				ShardInfo: shardInfo,
			}
			hfb.RequestShardHeaders(&block.MetaBlock{}) // for coverage only, no shard data
			hfb.RequestShardHeaders(metaBlock)
		}

		missingHdrs, missingProofs, missingFinalityAttesting := hfb.GetMissingData()
		require.Equal(t, uint32(2), missingHdrs)
		require.Equal(t, uint32(1), missingProofs) // only one proof missing here as the second one will be observed when header is received
		require.Zero(t, missingFinalityAttesting)

		go func() {
			time.Sleep(time.Millisecond * 200)
			// simulate receiving missing stuff
			proofsPool.AddProof(&block.HeaderProof{
				HeaderHash: td[2].headerHash,
			})

			headersPool.AddHeader(td[3].headerHash, td[3].header)

			headersPool.AddHeader(td[4].headerHash, td[4].header)
			proofsPool.AddProof(&block.HeaderProof{
				HeaderHash: td[4].headerHash,
			})
		}()

		err = hfb.WaitForHeadersIfNeeded(func() time.Duration {
			return time.Second * 2
		})
		require.NoError(t, err)

		missingHdrs, missingProofs, missingFinalityAttesting = hfb.GetMissingData()
		require.Zero(t, missingHdrs)
		require.Zero(t, missingProofs)
		require.Zero(t, missingFinalityAttesting)
	}
}

func TestHeadersForBlock_GetHeaderInfo(t *testing.T) {
	t.Parallel()

	td := createTestData(3, false)

	args := createMockArgs()
	args.ShardCoordinator = &testscommon.ShardsCoordinatorMock{
		NoShards: 1,
	}
	poolsHolder, ok := args.DataPool.(*dataRetriever.PoolsHolderMock)
	require.True(t, ok)

	headersPoolStub := createPoolsHolderForHeaderRequests()
	poolsHolder.SetHeadersPool(headersPoolStub)
	headersPool := poolsHolder.Headers()
	// adding the existing header
	headersPool.AddHeader(td[1].headerHash, td[1].header)

	args.BlockTracker = &mock.BlockTrackerStub{
		GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return td[0].header, nil, nil
		},
	}

	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)

	referencedHeaders := []*headerData{td[1]}
	shardInfo := createShardInfo(referencedHeaders)
	metaBlock := &block.MetaBlock{
		ShardInfo: shardInfo,
	}
	hfb.RequestShardHeaders(metaBlock)

	missingHdrs, missingProofs, missingFinalityAttesting := hfb.GetMissingData()
	require.Zero(t, missingHdrs)
	require.Zero(t, missingProofs)
	require.Equal(t, uint32(1), missingFinalityAttesting)

	go func() {
		time.Sleep(time.Millisecond * 200)
		// simulate receiving missing stuff
		headersPool.AddHeader(td[2].headerHash, td[2].header)
	}()

	err = hfb.WaitForHeadersIfNeeded(func() time.Duration {
		return time.Second * 2
	})
	require.NoError(t, err)

	missingHdrs, missingProofs, missingFinalityAttesting = hfb.GetMissingData()
	require.Zero(t, missingHdrs)
	require.Zero(t, missingProofs)
	require.Zero(t, missingFinalityAttesting)
}

func TestBaseProcessor_FilterHeadersWithoutProofs(t *testing.T) {
	t.Parallel()

	headersForCurrentBlock := map[string]data.HeaderHandler{
		"hash0": &testscommon.HeaderHandlerStub{
			EpochField: 12,
			GetNonceCalled: func() uint64 {
				return 1
			},
			GetShardIDCalled: func() uint32 {
				return 0
			},
		},
		"hash1": &testscommon.HeaderHandlerStub{
			EpochField: 12,
			GetNonceCalled: func() uint64 {
				return 1
			},
			GetShardIDCalled: func() uint32 {
				return 1
			},
		},
		"hash2": &testscommon.HeaderHandlerStub{
			EpochField: 12, // no proof for this one, should be marked for deletion
			GetNonceCalled: func() uint64 {
				return 2
			},
			GetShardIDCalled: func() uint32 {
				return 0
			},
		},
		"hash3": &testscommon.HeaderHandlerStub{
			EpochField: 1, // flag not active, for coverage only
			GetNonceCalled: func() uint64 {
				return 2
			},
			GetShardIDCalled: func() uint32 {
				return 1
			},
		},
	}

	args := createMockArgs()
	args.ShardCoordinator = &testscommon.ShardsCoordinatorMock{
		NoShards: 2,
	}
	args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return epoch == 12
		},
	}
	poolsHolder, ok := args.DataPool.(*dataRetriever.PoolsHolderMock)
	require.True(t, ok)
	proofsPoolStub := &dataRetriever.ProofsPoolMock{
		HasProofCalled: func(shardID uint32, headerHash []byte) bool {
			return string(headerHash) != "hash2"
		},
	}
	poolsHolder.SetProofsPool(proofsPoolStub)

	hfb, _ := headerForBlock.NewHeadersForBlock(args)

	for hash, header := range headersForCurrentBlock {
		hfb.AddHeaderUsedInBlock(hash, header)
	}

	// this call should fail because header with nonce 2 from shard 0 (hash2) does not have proof
	// and there is no other header with the same nonce and proof
	headersWithProofs, err := hfb.FilterHeadersWithoutProofs()
	require.True(t, errors.Is(err, process.ErrMissingHeaderProof))
	require.Nil(t, headersWithProofs)

	// add one more header with same nonce as hash2, but this one has proof
	hfb.AddHeaderUsedInBlock(
		"hash4",
		&testscommon.HeaderHandlerStub{
			EpochField: 12, // same nonce as above, but this one has proof
			GetNonceCalled: func() uint64 {
				return 2
			},
			GetShardIDCalled: func() uint32 {
				return 0
			},
		},
	)

	// this call should succeed, as for nonce 2 in shard 0 we have 2 headers, hash2 and hash4, but hash4 has proof
	headersWithProofs, err = hfb.FilterHeadersWithoutProofs()
	require.NoError(t, err)
	require.Equal(t, 4, len(headersWithProofs))

	returnedHashes := make([]string, 0, len(headersWithProofs))
	for hash := range headersWithProofs {
		returnedHashes = append(returnedHashes, hash)
	}
	slices.Sort(returnedHashes)

	expectedSortedHashes := []string{"hash0", "hash1", "hash3", "hash4"}
	require.Equal(t, expectedSortedHashes, returnedHashes)
}

type headerData struct {
	header     data.HeaderHandler
	headerHash []byte
}

func createPoolsHolderForHeaderRequests() retriever.HeadersPool {
	headersInPool := make(map[string]data.HeaderHandler)
	mutHeadersInPool := sync.RWMutex{}
	errNotFound := errors.New("header not found")

	handlers := make([]func(header data.HeaderHandler, shardHeaderHash []byte), 0)
	mutHandlers := sync.RWMutex{}

	return &pool.HeadersPoolStub{
		AddCalled: func(headerHash []byte, header data.HeaderHandler) {
			mutHeadersInPool.Lock()
			headersInPool[string(headerHash)] = header
			mutHeadersInPool.Unlock()

			mutHandlers.RLock()
			defer mutHandlers.RUnlock()
			for _, handler := range handlers {
				handler(header, headerHash)
			}
		},
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			mutHeadersInPool.RLock()
			defer mutHeadersInPool.RUnlock()
			if h, ok := headersInPool[string(hash)]; ok {
				return h, nil
			}
			return nil, errNotFound
		},
		GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
			mutHeadersInPool.RLock()
			defer mutHeadersInPool.RUnlock()
			for hash, h := range headersInPool {
				if h.GetNonce() == hdrNonce && h.GetShardID() == shardId {
					return []data.HeaderHandler{h}, [][]byte{[]byte(hash)}, nil
				}
			}
			return nil, nil, errNotFound
		},
		RegisterHandlerCalled: func(handler func(header data.HeaderHandler, shardHeaderHash []byte)) {
			mutHandlers.Lock()
			defer mutHandlers.Unlock()
			handlers = append(handlers, handler)
		},
	}
}

func createProofsPoolForHeaderRequests() retriever.ProofsPool {
	proofsInPool := make(map[string]data.HeaderProofHandler)
	mutProofsInPool := sync.RWMutex{}

	handlers := make([]func(header data.HeaderProofHandler), 0)
	mutHandlers := sync.RWMutex{}

	return &dataRetriever.ProofsPoolMock{
		AddProofCalled: func(headerProof data.HeaderProofHandler) bool {
			mutProofsInPool.Lock()
			proofsInPool[string(headerProof.GetHeaderHash())] = headerProof
			mutProofsInPool.Unlock()

			mutHandlers.RLock()
			defer mutHandlers.RUnlock()
			for _, handler := range handlers {
				handler(headerProof)
			}

			return true
		},
		HasProofCalled: func(_ uint32, headerHash []byte) bool {
			mutProofsInPool.RLock()
			defer mutProofsInPool.RUnlock()
			_, ok := proofsInPool[string(headerHash)]
			return ok
		},
		RegisterHandlerCalled: func(handler func(headerProof data.HeaderProofHandler)) {
			mutHandlers.Lock()
			defer mutHandlers.Unlock()
			handlers = append(handlers, handler)
		},
	}
}

func createTestData(numHeaders uint32, requestMetaHeaders bool) []*headerData {
	testData := make([]*headerData, numHeaders)
	for i := uint32(0); i < numHeaders; i++ {
		if requestMetaHeaders {
			testData[i] = &headerData{
				header: &block.MetaBlock{
					Round: 100,
					Nonce: uint64(i),
				},
				headerHash: []byte(fmt.Sprintf("hash%d", i)),
			}

			continue
		}

		testData[i] = &headerData{
			header: &block.HeaderV2{
				Header: &block.Header{
					ShardID: 0,
					Round:   100,
					Nonce:   uint64(i),
				},
			},
			headerHash: []byte(fmt.Sprintf("hash%d", i)),
		}
	}

	return testData
}

func createShardInfo(referencedHeaders []*headerData) []block.ShardData {
	shardData := make([]block.ShardData, len(referencedHeaders))
	for i, h := range referencedHeaders {
		shardData[i] = block.ShardData{
			HeaderHash: h.headerHash,
			Round:      h.header.GetRound(),
			PrevHash:   h.header.GetPrevHash(),
			Nonce:      h.header.GetNonce(),
			ShardID:    h.header.GetShardID(),
		}
	}

	return shardData
}
