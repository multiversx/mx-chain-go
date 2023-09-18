package processor

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	sovereignData "github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/stretchr/testify/require"
)

func createSovHdrProcArgs() *ArgsSovereignHeaderInterceptorProcessor {
	return &ArgsSovereignHeaderInterceptorProcessor{
		BlockBlackList:           &testscommon.TimeCacheStub{},
		Hasher:                   &hashingMocks.HasherMock{},
		Marshaller:               &marshallerMock.MarshalizerMock{},
		IncomingHeaderSubscriber: &sovereign.IncomingHeaderSubscriberStub{},
		HeadersPool:              &mock.HeadersCacherStub{},
	}
}

func TestNewSovereignHdrInterceptorProcessor(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		args := createSovHdrProcArgs()
		sovHdrProc, err := NewSovereignHdrInterceptorProcessor(args)
		require.Nil(t, err)
		require.False(t, sovHdrProc.IsInterfaceNil())
	})
	t.Run("nil args, should return error", func(t *testing.T) {
		sovHdrProc, err := NewSovereignHdrInterceptorProcessor(nil)
		require.Equal(t, process.ErrNilArgumentStruct, err)
		require.Nil(t, sovHdrProc)
	})
	t.Run("nil black list cache, should return error", func(t *testing.T) {
		args := createSovHdrProcArgs()
		args.BlockBlackList = nil
		sovHdrProc, err := NewSovereignHdrInterceptorProcessor(args)
		require.Equal(t, process.ErrNilBlackListCacher, err)
		require.Nil(t, sovHdrProc)
	})
	t.Run("nil hasher, should return error", func(t *testing.T) {
		args := createSovHdrProcArgs()
		args.Hasher = nil
		sovHdrProc, err := NewSovereignHdrInterceptorProcessor(args)
		require.Equal(t, errors.ErrNilHasher, err)
		require.Nil(t, sovHdrProc)
	})
	t.Run("nil marshaller, should return error", func(t *testing.T) {
		args := createSovHdrProcArgs()
		args.Marshaller = nil
		sovHdrProc, err := NewSovereignHdrInterceptorProcessor(args)
		require.Equal(t, errors.ErrNilMarshalizer, err)
		require.Nil(t, sovHdrProc)
	})
	t.Run("nil incoming header subscriber, should return error", func(t *testing.T) {
		args := createSovHdrProcArgs()
		args.IncomingHeaderSubscriber = nil
		sovHdrProc, err := NewSovereignHdrInterceptorProcessor(args)
		require.Equal(t, errors.ErrNilIncomingHeaderSubscriber, err)
		require.Nil(t, sovHdrProc)
	})
	t.Run("nil headers pool, should return error", func(t *testing.T) {
		args := createSovHdrProcArgs()
		args.HeadersPool = nil
		sovHdrProc, err := NewSovereignHdrInterceptorProcessor(args)
		require.Equal(t, process.ErrNilHeadersDataPool, err)
		require.Nil(t, sovHdrProc)
	})
}

func TestSovereignHeaderInterceptorProcessor_Validate(t *testing.T) {
	t.Parallel()

	extendedHdrHash := []byte("hash")
	wasHashComputed := false

	args := createSovHdrProcArgs()
	args.Hasher = &testscommon.HasherStub{
		ComputeCalled: func(s string) []byte {
			wasHashComputed = true
			return extendedHdrHash
		},
	}

	extendedHeader := &block.ShardHeaderExtended{}
	interceptedHdr := &testscommon.SovereignInterceptedExtendedHeaderDataStub{
		HashCalled: func() []byte {
			return extendedHdrHash
		},
		GetExtendedHeaderCalled: func() data.ShardHeaderExtendedHandler {
			return extendedHeader
		},
	}
	args.IncomingHeaderSubscriber = &sovereign.IncomingHeaderSubscriberStub{
		CreateExtendedHeaderCalled: func(header sovereignData.IncomingHeaderHandler) (data.ShardHeaderExtendedHandler, error) {
			require.Equal(t, extendedHeader, header)
			return extendedHeader, nil
		},
	}

	sovHdrProc, _ := NewSovereignHdrInterceptorProcessor(args)
	err := sovHdrProc.Validate(interceptedHdr, "")
	require.Nil(t, err)
	require.True(t, wasHashComputed)
}
