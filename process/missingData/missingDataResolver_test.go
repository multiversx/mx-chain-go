package missingData

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
	"github.com/multiversx/mx-chain-go/testscommon/preprocMocks"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
)

func TestNewMissingDataResolver(t *testing.T) {
	t.Parallel()

	proofsPool := &dataRetriever.ProofsPoolMock{}
	headersPool := &pool.HeadersPoolStub{}
	requestHandler := &testscommon.RequestHandlerStub{}
	blockDataRequester := &preprocMocks.BlockDataRequesterStub{}
	commonArgs := MissingDataResolverArgs{
		HeadersPool:        headersPool,
		ProofsPool:         proofsPool,
		RequestHandler:     requestHandler,
		BlockDataRequester: blockDataRequester,
	}

	t.Run("valid inputs ok", func(t *testing.T) {
		t.Parallel()

		mdr, err := NewMissingDataResolver(commonArgs)
		require.NotNil(t, mdr)
		require.Nil(t, err)
	})

	t.Run("nil headersPool should err", func(t *testing.T) {
		t.Parallel()

		commonArgsCopy := commonArgs
		commonArgsCopy.HeadersPool = nil
		mdr, err := NewMissingDataResolver(commonArgsCopy)
		require.Nil(t, mdr)
		require.Equal(t, process.ErrNilHeadersDataPool, err)
	})

	t.Run("nil proofsPool should err", func(t *testing.T) {
		t.Parallel()

		commonArgsCopy := commonArgs
		commonArgsCopy.ProofsPool = nil
		mdr, err := NewMissingDataResolver(commonArgsCopy)
		require.Nil(t, mdr)
		require.Equal(t, process.ErrNilProofsPool, err)
	})

	t.Run("nil requestHandler should err", func(t *testing.T) {
		t.Parallel()

		commonArgsCopy := commonArgs
		commonArgsCopy.RequestHandler = nil
		mdr, err := NewMissingDataResolver(commonArgsCopy)
		require.Nil(t, mdr)
		require.Equal(t, process.ErrNilRequestHandler, err)
	})
	t.Run("nil blockDataRequester should err", func(t *testing.T) {
		t.Parallel()

		commonArgsCopy := commonArgs
		commonArgsCopy.BlockDataRequester = nil
		mdr, err := NewMissingDataResolver(commonArgsCopy)
		require.Nil(t, mdr)
		require.Equal(t, process.ErrNilBlockDataRequester, err)
	})
}

func TestMissingDataResolver_AddAndMarkMissingData(t *testing.T) {
	t.Parallel()

	proofNotFoundError := errors.New("proof not found")
	headerNotFoundError := errors.New("header not found")
	headerHash := []byte("headerHash")
	proofHash := []byte("proofHash")

	proofsPool := &dataRetriever.ProofsPoolMock{
		GetProofCalled: func(_ uint32, _ []byte) (data.HeaderProofHandler, error) {
			return nil, proofNotFoundError
		},
	}
	headersPool := &pool.HeadersPoolStub{
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			return nil, headerNotFoundError
		},
	}
	requestHandler := &testscommon.RequestHandlerStub{}

	blockDataRequester := &preprocMocks.BlockDataRequesterStub{}
	commonArgs := MissingDataResolverArgs{
		HeadersPool:        headersPool,
		ProofsPool:         proofsPool,
		RequestHandler:     requestHandler,
		BlockDataRequester: blockDataRequester,
	}
	t.Run("add missing and mark header", func(t *testing.T) {
		t.Parallel()

		mdr, _ := NewMissingDataResolver(commonArgs)
		mdr.addMissingHeader(headerHash)
		require.Contains(t, mdr.missingHeaders, string(headerHash))
		require.False(t, mdr.allHeadersReceived())

		mdr.markHeaderReceived(headerHash)
		require.NotContains(t, mdr.missingHeaders, string(headerHash))
		require.True(t, mdr.allHeadersReceived())
	})

	t.Run("add missing and mark proof", func(t *testing.T) {
		t.Parallel()

		mdr, _ := NewMissingDataResolver(commonArgs)
		mdr.addMissingProof(0, proofHash)
		require.Contains(t, mdr.missingProofs, string(proofHash))
		require.False(t, mdr.allProofsReceived())

		mdr.markProofReceived(proofHash)
		require.NotContains(t, mdr.missingProofs, string(proofHash))
		require.True(t, mdr.allProofsReceived())
	})
}

func TestMissingDataResolver_requestHeaderIfNeeded(t *testing.T) {
	t.Parallel()

	existingMetaHeader := "existingMetaHeader"
	existingShardHeader := "existingShardHeader"
	notFoundError := errors.New("not found")

	headersPool := &pool.HeadersPoolStub{
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			switch string(hash) {
			case existingMetaHeader:
				return &block.MetaBlock{}, nil
			case existingShardHeader:
				return &block.HeaderV2{}, nil
			default:
				return nil, notFoundError
			}
		},
	}
	proofsPool := &dataRetriever.ProofsPoolMock{}

	requestHandler := &testscommon.RequestHandlerStub{
		RequestMetaHeaderCalled:  func(hash []byte) {},
		RequestShardHeaderCalled: func(shardID uint32, hash []byte) {},
	}
	blockDataRequester := &preprocMocks.BlockDataRequesterStub{}
	commonArgs := MissingDataResolverArgs{
		HeadersPool:        headersPool,
		ProofsPool:         proofsPool,
		RequestHandler:     requestHandler,
		BlockDataRequester: blockDataRequester,
	}

	t.Run("header already exists", func(t *testing.T) {
		t.Parallel()

		mdr, _ := NewMissingDataResolver(commonArgs)
		mdr.requestHeaderIfNeeded(core.MetachainShardId, []byte(existingShardHeader))
		require.NotContains(t, mdr.missingHeaders, existingShardHeader)
		require.True(t, mdr.allHeadersReceived())
	})

	t.Run("meta header missing", func(t *testing.T) {
		t.Parallel()

		mdr, _ := NewMissingDataResolver(commonArgs)
		mdr.requestHeaderIfNeeded(core.MetachainShardId, []byte("missingHeader"))
		require.False(t, mdr.allHeadersReceived())
	})
	t.Run("shard header missing", func(t *testing.T) {
		t.Parallel()

		mdr, _ := NewMissingDataResolver(commonArgs)
		mdr.requestHeaderIfNeeded(1, []byte("missingShardHeader"))
		require.False(t, mdr.allHeadersReceived())
	})
	t.Run("header arriving in pool after first check, should not request", func(t *testing.T) {
		t.Parallel()

		numCall := 0
		headersPool := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if string(hash) == existingMetaHeader && numCall == 0 {
					numCall++
				} else if string(hash) == existingMetaHeader && numCall == 1 {
					numCall++
					return &block.MetaBlock{}, nil
				}
				return nil, notFoundError
			},
		}
		requestHandler := &testscommon.RequestHandlerStub{
			RequestMetaHeaderCalled: func(hash []byte) {
				require.Fail(t, "RequestMetaHeader should not be called again for existing header")
			},
		}

		commonArgs := MissingDataResolverArgs{
			HeadersPool:        headersPool,
			ProofsPool:         proofsPool,
			RequestHandler:     requestHandler,
			BlockDataRequester: blockDataRequester,
		}
		mdr, _ := NewMissingDataResolver(commonArgs)
		mdr.requestHeaderIfNeeded(core.MetachainShardId, []byte(existingMetaHeader))
		require.True(t, mdr.allHeadersReceived())
	})
}

func TestMissingDataResolver_requestProofIfNeeded(t *testing.T) {
	t.Parallel()

	proofHash := []byte("proofHash")
	headersPool := &pool.HeadersPoolStub{}
	requestHandler := &testscommon.RequestHandlerStub{}
	blockDataRequester := &preprocMocks.BlockDataRequesterStub{}
	commonArgs := MissingDataResolverArgs{
		HeadersPool:        headersPool,
		RequestHandler:     requestHandler,
		BlockDataRequester: blockDataRequester,
	}
	t.Run("proof already exists", func(t *testing.T) {
		t.Parallel()
		proofsPool := &dataRetriever.ProofsPoolMock{
			HasProofCalled: func(_ uint32, _ []byte) bool {
				return true
			},
		}

		commonArgsCopy := commonArgs
		commonArgsCopy.ProofsPool = proofsPool
		mdr, _ := NewMissingDataResolver(commonArgsCopy)
		mdr.requestProofIfNeeded(0, proofHash)
		require.NotContains(t, mdr.missingProofs, string(proofHash))
		require.True(t, mdr.allProofsReceived())
	})

	t.Run("proof missing", func(t *testing.T) {
		t.Parallel()
		proofsPool := &dataRetriever.ProofsPoolMock{
			HasProofCalled: func(_ uint32, _ []byte) bool {
				return false
			},
		}
		commonArgsCopy := commonArgs
		commonArgsCopy.ProofsPool = proofsPool
		mdr, _ := NewMissingDataResolver(commonArgsCopy)
		mdr.requestProofIfNeeded(0, []byte("missingProof"))
		require.False(t, mdr.allProofsReceived())
	})
	t.Run("proof arriving in pool after first check, should not request", func(t *testing.T) {
		t.Parallel()

		numCall := 0
		proofsPool := &dataRetriever.ProofsPoolMock{
			HasProofCalled: func(_ uint32, hash []byte) bool {
				numCall++
				return numCall > 1
			},
		}

		requestHandler := &testscommon.RequestHandlerStub{
			RequestEquivalentProofByHashCalled: func(shardID uint32, hash []byte) {
				require.Fail(t, "RequestHeaderProof should not be called again for existing proof")
			},
		}
		commonArgsCopy := commonArgs
		commonArgsCopy.ProofsPool = proofsPool
		commonArgsCopy.RequestHandler = requestHandler
		mdr, _ := NewMissingDataResolver(commonArgsCopy)
		mdr.requestProofIfNeeded(0, proofHash)
		require.True(t, mdr.allProofsReceived())
	})
}

func TestMissingDataResolver_WaitForMissingData(t *testing.T) {
	t.Parallel()

	headerHash := []byte("headerHash")
	errorHeaderNotFound := errors.New("header not found")
	proofsPool := &dataRetriever.ProofsPoolMock{}
	blockDataRequester := &preprocMocks.BlockDataRequesterStub{}

	t.Run("data received before timeout", func(t *testing.T) {
		t.Parallel()

		headersPool := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if string(hash) == string(headerHash) {
					return nil, errorHeaderNotFound
				}
				return nil, nil
			},
		}
		requestHandler := &testscommon.RequestHandlerStub{}
		commonArgs := MissingDataResolverArgs{
			HeadersPool:        headersPool,
			ProofsPool:         proofsPool,
			RequestHandler:     requestHandler,
			BlockDataRequester: blockDataRequester,
		}
		mdr, _ := NewMissingDataResolver(commonArgs)
		mdr.addMissingHeader(headerHash)
		go func() {
			time.Sleep(50 * time.Millisecond)
			mdr.markHeaderReceived(headerHash)
		}()

		err := mdr.WaitForMissingData(100 * time.Millisecond)
		require.Nil(t, err)
	})

	t.Run("timeout waiting for data", func(t *testing.T) {
		t.Parallel()
		headersPool := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if string(hash) == string(headerHash) {
					return nil, errorHeaderNotFound
				}
				return nil, nil
			},
		}
		requestHandler := &testscommon.RequestHandlerStub{}
		commonArgs := MissingDataResolverArgs{
			HeadersPool:        headersPool,
			ProofsPool:         proofsPool,
			RequestHandler:     requestHandler,
			BlockDataRequester: blockDataRequester,
		}
		mdr, _ := NewMissingDataResolver(commonArgs)
		_ = mdr.addMissingHeader(headerHash)
		err := mdr.WaitForMissingData(50 * time.Millisecond)
		require.Equal(t, process.ErrTimeIsOut, err)
	})
	t.Run("timeout waiting for block data requester", func(t *testing.T) {
		t.Parallel()
		headersPool := &pool.HeadersPoolStub{}
		requestHandler := &testscommon.RequestHandlerStub{}
		blockDataRequester := &preprocMocks.BlockDataRequesterStub{
			IsDataPreparedForProcessingCalled: func(haveTime func() time.Duration) error {
				for haveTime() > 0 {
					time.Sleep(10 * time.Millisecond)
				}

				return fmt.Errorf("missing data")
			},
		}
		commonArgs := MissingDataResolverArgs{
			HeadersPool:        headersPool,
			ProofsPool:         proofsPool,
			RequestHandler:     requestHandler,
			BlockDataRequester: blockDataRequester,
		}
		mdr, _ := NewMissingDataResolver(commonArgs)
		err := mdr.WaitForMissingData(50 * time.Millisecond)
		require.NotNil(t, err)
		require.Equal(t, process.ErrTimeIsOut, err)
	})
}

func TestMissingDataResolver_MonitorReceivedData(t *testing.T) {
	t.Parallel()

	var headerReceivedHandler func(data.HeaderHandler, []byte)
	var proofReceivedHandler func(data.HeaderProofHandler)
	proofsPool := &dataRetriever.ProofsPoolMock{}
	headersPool := &pool.HeadersPoolStub{}
	requestHandler := &testscommon.RequestHandlerStub{}

	headerHash := "headerHash"
	headersPool.RegisterHandlerCalled = func(handler func(data.HeaderHandler, []byte)) {
		headerReceivedHandler = handler
	}
	proofsPool.RegisterHandlerCalled = func(handler func(data.HeaderProofHandler)) {
		proofReceivedHandler = handler
	}
	blockDataRequester := &preprocMocks.BlockDataRequesterStub{}
	commonArgs := MissingDataResolverArgs{
		HeadersPool:        headersPool,
		ProofsPool:         proofsPool,
		RequestHandler:     requestHandler,
		BlockDataRequester: blockDataRequester,
	}
	mdr, _ := NewMissingDataResolver(commonArgs)

	t.Run("monitor headers and proofs", func(t *testing.T) {
		t.Parallel()

		mdr.addMissingHeader([]byte(headerHash))
		mdr.addMissingProof(0, []byte(headerHash)) // the proof key is the header hash

		// receive both header and proof for header
		go headerReceivedHandler(&block.HeaderV2{}, []byte(headerHash))
		go proofReceivedHandler(&processMocks.HeaderProofHandlerStub{
			GetHeaderHashCalled: func() []byte {
				return []byte(headerHash)
			},
		})
		time.Sleep(50 * time.Millisecond)
		require.True(t, mdr.allDataReceived())
	})
}

func TestMissingDataResolver_RequestMissingMetaHeadersBlocking(t *testing.T) {
	t.Parallel()

	headerNotFoundErr := errors.New("header not found")
	metaHeaderHash := []byte("metaHeaderHash")
	shardHeader := &block.HeaderV2{
		Header: &block.Header{
			MetaBlockHashes: [][]byte{metaHeaderHash},
		},
	}

	proofsPool := &dataRetriever.ProofsPoolMock{
		HasProofCalled: func(shardID uint32, headerHash []byte) bool {
			return false
		},
	}
	headersPool := &pool.HeadersPoolStub{
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			if string(hash) == string(metaHeaderHash) {
				return nil, headerNotFoundErr
			}
			return nil, nil
		},
	}
	requestHandler := &testscommon.RequestHandlerStub{}
	blockDataRequester := &preprocMocks.BlockDataRequesterStub{}
	commonArgs := MissingDataResolverArgs{
		HeadersPool:        headersPool,
		ProofsPool:         proofsPool,
		RequestHandler:     requestHandler,
		BlockDataRequester: blockDataRequester,
	}
	mdr, _ := NewMissingDataResolver(commonArgs)

	t.Run("nil shard header should err", func(t *testing.T) {
		t.Parallel()

		err := mdr.RequestMissingMetaHeadersBlocking(nil, 100*time.Millisecond)
		require.Equal(t, process.ErrNilBlockHeader, err)
	})

	t.Run("request missing meta headers and proofs, none received", func(t *testing.T) {
		t.Parallel()

		err := mdr.RequestMissingMetaHeadersBlocking(shardHeader, 100*time.Millisecond)
		require.Equal(t, process.ErrTimeIsOut, err)
		require.False(t, mdr.allHeadersReceived())
		require.False(t, mdr.allProofsReceived())
	})
	t.Run("request missing meta headers and proofs, all received", func(t *testing.T) {
		t.Parallel()

		var headerReceivedHandler func(header data.HeaderHandler, _ []byte)
		var proofReceivedHandler func(proof data.HeaderProofHandler)

		proofsPool := &dataRetriever.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, headerHash []byte) bool {
				return false
			},
			RegisterHandlerCalled: func(handler func(headerProof data.HeaderProofHandler)) {
				proofReceivedHandler = handler
			},
		}
		headersPool := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if string(hash) == string(metaHeaderHash) {
					return nil, headerNotFoundErr
				}
				return nil, nil
			},
			RegisterHandlerCalled: func(handler func(header data.HeaderHandler, _ []byte)) {
				headerReceivedHandler = handler
			},
		}
		commonArgs := MissingDataResolverArgs{
			HeadersPool:        headersPool,
			ProofsPool:         proofsPool,
			RequestHandler:     requestHandler,
			BlockDataRequester: blockDataRequester,
		}
		mdr, _ := NewMissingDataResolver(commonArgs)

		go func() {
			time.Sleep(50 * time.Millisecond)
			headerReceivedHandler(&block.HeaderV2{}, metaHeaderHash)
			proofReceivedHandler(&processMocks.HeaderProofHandlerStub{
				GetHeaderHashCalled: func() []byte {
					return metaHeaderHash
				},
			})
		}()

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			err := mdr.RequestMissingMetaHeadersBlocking(shardHeader, 200*time.Millisecond)
			require.Nil(t, err)
			require.True(t, mdr.allHeadersReceived())
			require.True(t, mdr.allProofsReceived())
			wg.Done()
		}()
		wg.Wait()
	})
}

func TestMissingDataResolver_RequestBlockTransactions(t *testing.T) {
	t.Parallel()

	var called bool
	proofsPool := &dataRetriever.ProofsPoolMock{}
	headersPool := &pool.HeadersPoolStub{}
	requestHandler := &testscommon.RequestHandlerStub{}
	blockDataRequester := &preprocMocks.BlockDataRequesterStub{
		RequestBlockTransactionsCalled: func(_ *block.Body) {
			called = true
		},
	}
	commonArgs := MissingDataResolverArgs{
		HeadersPool:        headersPool,
		ProofsPool:         proofsPool,
		RequestHandler:     requestHandler,
		BlockDataRequester: blockDataRequester,
	}

	mdr, _ := NewMissingDataResolver(commonArgs)
	body := &block.Body{}
	mdr.RequestBlockTransactions(body)
	require.True(t, called)
}

func TestMissingDataResolver_RequestMiniBlocksAndTransactions(t *testing.T) {
	t.Parallel()
	var called bool
	proofsPool := &dataRetriever.ProofsPoolMock{}
	headersPool := &pool.HeadersPoolStub{}
	requestHandler := &testscommon.RequestHandlerStub{}
	blockDataRequester := &preprocMocks.BlockDataRequesterStub{
		RequestMiniBlocksAndTransactionsCalled: func(_ data.HeaderHandler) {
			called = true
		},
	}

	commonArgs := MissingDataResolverArgs{
		HeadersPool:        headersPool,
		ProofsPool:         proofsPool,
		RequestHandler:     requestHandler,
		BlockDataRequester: blockDataRequester,
	}

	mdr, _ := NewMissingDataResolver(commonArgs)
	header := &block.HeaderV2{}
	mdr.RequestMiniBlocksAndTransactions(header)
	require.True(t, called)
}

func TestMissingDataResolver_GetFinalCrossMiniBlockInfoAndRequestMissing(t *testing.T) {
	t.Parallel()

	proofsPool := &dataRetriever.ProofsPoolMock{}
	headersPool := &pool.HeadersPoolStub{}
	requestHandler := &testscommon.RequestHandlerStub{}

	expectedValue := []*data.MiniBlockInfo{
		{Hash: []byte("mbHash"), SenderShardID: 1},
		{Hash: []byte("mbHash2"), SenderShardID: 2},
	}

	blockDataRequester := &preprocMocks.BlockDataRequesterStub{
		GetFinalCrossMiniBlockInfoAndRequestMissingCalled: func(_ data.HeaderHandler) []*data.MiniBlockInfo {
			return expectedValue
		},
	}
	commonArgs := MissingDataResolverArgs{
		HeadersPool:        headersPool,
		ProofsPool:         proofsPool,
		RequestHandler:     requestHandler,
		BlockDataRequester: blockDataRequester,
	}

	mdr, _ := NewMissingDataResolver(commonArgs)
	header := &block.HeaderV2{}
	finalCrossMiniBlocks := mdr.GetFinalCrossMiniBlockInfoAndRequestMissing(header)
	require.Equal(t, expectedValue, finalCrossMiniBlocks)
}

func TestMissingDataResolver_Reset(t *testing.T) {
	t.Parallel()

	called := false
	proofsPool := &dataRetriever.ProofsPoolMock{}
	headersPool := &pool.HeadersPoolStub{}
	requestHandler := &testscommon.RequestHandlerStub{}
	blockDataRequester := &preprocMocks.BlockDataRequesterStub{
		ResetCalled: func() {
			called = true
		},
	}
	commonArgs := MissingDataResolverArgs{
		HeadersPool:        headersPool,
		ProofsPool:         proofsPool,
		RequestHandler:     requestHandler,
		BlockDataRequester: blockDataRequester,
	}

	mdr, _ := NewMissingDataResolver(commonArgs)
	mdr.addMissingHeader([]byte("headerHash"))
	mdr.addMissingProof(0, []byte("proofHash"))
	require.False(t, mdr.allDataReceived())

	mdr.Reset()
	require.True(t, mdr.allDataReceived())
	require.True(t, called)
}

func TestMissingDataResolver_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	proofsPool := &dataRetriever.ProofsPoolMock{}
	headersPool := &pool.HeadersPoolStub{}
	requestHandler := &testscommon.RequestHandlerStub{}
	blockDataRequester := &preprocMocks.BlockDataRequesterStub{}
	commonArgs := MissingDataResolverArgs{
		HeadersPool:        headersPool,
		ProofsPool:         proofsPool,
		RequestHandler:     requestHandler,
		BlockDataRequester: blockDataRequester,
	}
	t.Run("nil receiver should be nil", func(t *testing.T) {
		t.Parallel()

		var mdr *missingDataResolver
		require.True(t, check.IfNil(mdr))
	})

	t.Run("non-nil receiver should not be nil", func(t *testing.T) {
		t.Parallel()

		mdr, _ := NewMissingDataResolver(commonArgs)
		require.False(t, check.IfNil(mdr))
	})
}
