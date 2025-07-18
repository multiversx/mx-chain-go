package missingData

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
)

func TestNewMissingDataResolver(t *testing.T) {
	proofsPool := &dataRetriever.ProofsPoolMock{}
	headersPool := &pool.HeadersPoolStub{}
	requestHandler := &testscommon.RequestHandlerStub{}

	t.Run("valid inputs ok", func(t *testing.T) {
		mdr, err := NewMissingDataResolver(headersPool, proofsPool, requestHandler)
		require.NotNil(t, mdr)
		require.Nil(t, err)
	})

	t.Run("nil headersPool should err", func(t *testing.T) {
		mdr, err := NewMissingDataResolver(nil, proofsPool, requestHandler)
		require.Nil(t, mdr)
		require.Equal(t, process.ErrNilHeadersDataPool, err)
	})

	t.Run("nil proofsPool should err", func(t *testing.T) {
		mdr, err := NewMissingDataResolver(headersPool, nil, requestHandler)
		require.Nil(t, mdr)
		require.Equal(t, process.ErrNilProofsPool, err)
	})

	t.Run("nil requestHandler should err", func(t *testing.T) {
		mdr, err := NewMissingDataResolver(headersPool, proofsPool, nil)
		require.Nil(t, mdr)
		require.Equal(t, process.ErrNilRequestHandler, err)
	})
}

func TestMissingDataResolver_AddAndMarkMissingData(t *testing.T) {
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

	t.Run("add missing and mark header", func(t *testing.T) {
		mdr, _ := NewMissingDataResolver(headersPool, proofsPool, requestHandler)
		mdr.addMissingHeader(headerHash)
		require.Contains(t, mdr.missingHeaders, string(headerHash))
		require.False(t, mdr.allHeadersReceived())

		mdr.markHeaderReceived(headerHash)
		require.NotContains(t, mdr.missingHeaders, string(headerHash))
		require.True(t, mdr.allHeadersReceived())
	})

	t.Run("add missing and mark proof", func(t *testing.T) {
		mdr, _ := NewMissingDataResolver(headersPool, proofsPool, requestHandler)
		mdr.addMissingProof(0, proofHash)
		require.Contains(t, mdr.missingProofs, string(proofHash))
		require.False(t, mdr.allProofsReceived())

		mdr.markProofReceived(proofHash)
		require.NotContains(t, mdr.missingProofs, string(proofHash))
		require.True(t, mdr.allProofsReceived())
	})
}

func TestMissingDataResolver_RequestHeaderIfNeeded(t *testing.T) {
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
	mdr, _ := NewMissingDataResolver(headersPool, proofsPool, requestHandler)

	t.Run("header already exists", func(t *testing.T) {
		mdr.requestHeaderIfNeeded(core.MetachainShardId, []byte(existingShardHeader))
		require.NotContains(t, mdr.missingHeaders, existingShardHeader)
		require.True(t, mdr.allHeadersReceived())
	})

	t.Run("header missing", func(t *testing.T) {
		mdr.requestHeaderIfNeeded(core.MetachainShardId, []byte("missingHeader"))
		require.False(t, mdr.allHeadersReceived())
	})
}

func TestMissingDataResolver_WaitForMissingData(t *testing.T) {
	headerHash := []byte("headerHash")
	errorHeaderNotFound := errors.New("header not found")
	proofsPool := &dataRetriever.ProofsPoolMock{}
	headersPool := &pool.HeadersPoolStub{
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			if string(hash) == string(headerHash) {
				return nil, errorHeaderNotFound
			}
			return nil, nil
		},
	}
	requestHandler := &testscommon.RequestHandlerStub{}
	mdr, _ := NewMissingDataResolver(headersPool, proofsPool, requestHandler)

	t.Run("data received before timeout", func(t *testing.T) {
		mdr.addMissingHeader(headerHash)
		go func() {
			time.Sleep(50 * time.Millisecond)
			mdr.markHeaderReceived(headerHash)
		}()

		err := mdr.waitForMissingData(100 * time.Millisecond)
		require.Nil(t, err)
	})

	t.Run("timeout waiting for data", func(t *testing.T) {
		mdr.addMissingHeader(headerHash)
		err := mdr.waitForMissingData(50 * time.Millisecond)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "timeout waiting for missing data")
	})
}

func TestMissingDataResolver_MonitorReceivedData(t *testing.T) {
	proofsPool := &dataRetriever.ProofsPoolMock{}
	headersPool := &pool.HeadersPoolStub{}
	requestHandler := &testscommon.RequestHandlerStub{}

	headersPool.RegisterHandlerCalled = func(handler func(data.HeaderHandler, []byte)) {
		go handler(nil, []byte("headerHash"))
	}
	proofsPool.RegisterHandlerCalled = func(handler func(data.HeaderProofHandler)) {
		go handler(&processMocks.HeaderProofHandlerStub{
			GetHeaderHashCalled: func() []byte {
				return []byte("proofHash")
			},
		})
	}

	mdr, _ := NewMissingDataResolver(headersPool, proofsPool, requestHandler)

	t.Run("monitor headers and proofs", func(t *testing.T) {
		mdr.addMissingHeader([]byte("headerHash"))
		mdr.addMissingProof(0, []byte("proofHash"))
		time.Sleep(50 * time.Millisecond)
		require.True(t, mdr.allDataReceived())
	})
}
