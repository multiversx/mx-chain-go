package processor_test

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/interceptors/processor"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/stretchr/testify/assert"
)

func createMockHdrArgument() *processor.ArgHdrInterceptorProcessor {
	arg := &processor.ArgHdrInterceptorProcessor{
		Headers:             &mock.HeadersCacherStub{},
		Proofs:              &dataRetriever.ProofsPoolMock{},
		BlockBlackList:      &testscommon.TimeCacheStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}

	return arg
}

// ------- NewHdrInterceptorProcessor

func TestNewHdrInterceptorProcessor_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	hip, err := processor.NewHdrInterceptorProcessor(nil)

	assert.Nil(t, hip)
	assert.Equal(t, process.ErrNilArgumentStruct, err)
}

func TestNewHdrInterceptorProcessor_NilHeadersShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	arg.Headers = nil
	hip, err := processor.NewHdrInterceptorProcessor(arg)

	assert.Nil(t, hip)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestNewHdrInterceptorProcessor_NilBlackListHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	arg.BlockBlackList = nil
	hip, err := processor.NewHdrInterceptorProcessor(arg)

	assert.Nil(t, hip)
	assert.Equal(t, process.ErrNilBlackListCacher, err)
}

func TestNewHdrInterceptorProcessor_NilProofsPoolShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	arg.Proofs = nil
	hip, err := processor.NewHdrInterceptorProcessor(arg)

	assert.Nil(t, hip)
	assert.Equal(t, process.ErrNilProofsPool, err)
}

func TestNewHdrInterceptorProcessor_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	arg.EnableEpochsHandler = nil
	hip, err := processor.NewHdrInterceptorProcessor(arg)

	assert.Nil(t, hip)
	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
}

func TestNewHdrInterceptorProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	hip, err := processor.NewHdrInterceptorProcessor(arg)

	assert.False(t, check.IfNil(hip))
	assert.Nil(t, err)
	assert.False(t, hip.IsInterfaceNil())
}

// ------- Validate

func TestHdrInterceptorProcessor_ValidateNilHdrShouldErr(t *testing.T) {
	t.Parallel()

	hip, _ := processor.NewHdrInterceptorProcessor(createMockHdrArgument())

	err := hip.Validate(nil, "")

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestHdrInterceptorProcessor_ValidateHeaderIsBlackListedShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	arg.BlockBlackList = &testscommon.TimeCacheStub{
		HasCalled: func(key string) bool {
			return true
		},
	}
	hip, _ := processor.NewHdrInterceptorProcessor(arg)

	hdrInterceptedData := &struct {
		testscommon.InterceptedDataStub
		mock.GetHdrHandlerStub
	}{
		InterceptedDataStub: testscommon.InterceptedDataStub{
			HashCalled: func() []byte {
				return make([]byte, 0)
			},
		},
	}
	err := hip.Validate(hdrInterceptedData, "")

	assert.Equal(t, process.ErrHeaderIsBlackListed, err)
}

func TestHdrInterceptorProcessor_ValidateReturnsNil(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	arg.BlockBlackList = &testscommon.TimeCacheStub{}
	hip, _ := processor.NewHdrInterceptorProcessor(arg)

	hdrInterceptedData := &struct {
		testscommon.InterceptedDataStub
		mock.GetHdrHandlerStub
	}{
		InterceptedDataStub: testscommon.InterceptedDataStub{
			HashCalled: func() []byte {
				return make([]byte, 0)
			},
		},
	}
	hdrInterceptedData.GetHdrHandlerStub.HeaderHandlerCalled = func() data.HeaderHandler {
		return &block.Header{}
	}
	err := hip.Validate(hdrInterceptedData, "")

	assert.Nil(t, err)
}

func TestHdrInterceptorProcessor_ValidateInExcludedIntervals(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	arg.BlockBlackList = &testscommon.TimeCacheStub{}
	hip, _ := processor.NewHdrInterceptorProcessor(arg)

	t.Run("shard 0", func(t *testing.T) {
		hdr := &block.Header{
			ShardID: 0,
		}

		hdrInterceptedData := &struct {
			testscommon.InterceptedDataStub
			mock.GetHdrHandlerStub
		}{
			InterceptedDataStub: testscommon.InterceptedDataStub{
				HashCalled: func() []byte {
					return make([]byte, 0)
				},
			},
		}
		hdrInterceptedData.GetHdrHandlerStub.HeaderHandlerCalled = func() data.HeaderHandler {
			return hdr
		}

		t.Run("round == 0 should not error", func(t *testing.T) {
			hdr.Round = 0
			err := hip.Validate(hdrInterceptedData, "")
			assert.Nil(t, err)
		})
		t.Run("round == 6076599 should not error", func(t *testing.T) {
			hdr.Round = 6076599
			err := hip.Validate(hdrInterceptedData, "")
			assert.Nil(t, err)
		})
		t.Run("round == 6076600 should error", func(t *testing.T) {
			hdr.Round = 6076600
			err := hip.Validate(hdrInterceptedData, "")
			assert.NotNil(t, err)
			assert.Equal(t, "header is in excluded range, shard 0, round 6076600, low 6076600, high 6630115", err.Error())
		})
		t.Run("round == 6300000 should error", func(t *testing.T) {
			hdr.Round = 6300000
			err := hip.Validate(hdrInterceptedData, "")
			assert.NotNil(t, err)
			assert.Equal(t, "header is in excluded range, shard 0, round 6300000, low 6076600, high 6630115", err.Error())
		})
		t.Run("round == 6630115 should error", func(t *testing.T) {
			hdr.Round = 6630115
			err := hip.Validate(hdrInterceptedData, "")
			assert.NotNil(t, err)
			assert.Equal(t, "header is in excluded range, shard 0, round 6630115, low 6076600, high 6630115", err.Error())
		})
		t.Run("round == 6630116 should not error", func(t *testing.T) {
			hdr.Round = 6630116
			err := hip.Validate(hdrInterceptedData, "")
			assert.Nil(t, err)
		})
		t.Run("round == 99999999 should not error", func(t *testing.T) {
			hdr.Round = 99999999
			err := hip.Validate(hdrInterceptedData, "")
			assert.Nil(t, err)
		})
	})

	t.Run("shard 1", func(t *testing.T) {
		hdr := &block.Header{
			ShardID: 1,
		}

		hdrInterceptedData := &struct {
			testscommon.InterceptedDataStub
			mock.GetHdrHandlerStub
		}{
			InterceptedDataStub: testscommon.InterceptedDataStub{
				HashCalled: func() []byte {
					return make([]byte, 0)
				},
			},
		}
		hdrInterceptedData.GetHdrHandlerStub.HeaderHandlerCalled = func() data.HeaderHandler {
			return hdr
		}

		t.Run("round == 0 should not error", func(t *testing.T) {
			hdr.Round = 0
			err := hip.Validate(hdrInterceptedData, "")
			assert.Nil(t, err)
		})
		t.Run("round == 6076599 should not error", func(t *testing.T) {
			hdr.Round = 6076599
			err := hip.Validate(hdrInterceptedData, "")
			assert.Nil(t, err)
		})
		t.Run("round == 6076600 should error", func(t *testing.T) {
			hdr.Round = 6076600
			err := hip.Validate(hdrInterceptedData, "")
			assert.NotNil(t, err)
			assert.Equal(t, "header is in excluded range, shard 1, round 6076600, low 6076600, high 6630115", err.Error())
		})
		t.Run("round == 6300000 should error", func(t *testing.T) {
			hdr.Round = 6300000
			err := hip.Validate(hdrInterceptedData, "")
			assert.NotNil(t, err)
			assert.Equal(t, "header is in excluded range, shard 1, round 6300000, low 6076600, high 6630115", err.Error())
		})
		t.Run("round == 6630115 should error", func(t *testing.T) {
			hdr.Round = 6630115
			err := hip.Validate(hdrInterceptedData, "")
			assert.NotNil(t, err)
			assert.Equal(t, "header is in excluded range, shard 1, round 6630115, low 6076600, high 6630115", err.Error())
		})
		t.Run("round == 6630116 should not error", func(t *testing.T) {
			hdr.Round = 6630116
			err := hip.Validate(hdrInterceptedData, "")
			assert.Nil(t, err)
		})
		t.Run("round == 99999999 should not error", func(t *testing.T) {
			hdr.Round = 99999999
			err := hip.Validate(hdrInterceptedData, "")
			assert.Nil(t, err)
		})
	})

	t.Run("shard 2", func(t *testing.T) {
		hdr := &block.Header{
			ShardID: 2,
		}

		hdrInterceptedData := &struct {
			testscommon.InterceptedDataStub
			mock.GetHdrHandlerStub
		}{
			InterceptedDataStub: testscommon.InterceptedDataStub{
				HashCalled: func() []byte {
					return make([]byte, 0)
				},
			},
		}
		hdrInterceptedData.GetHdrHandlerStub.HeaderHandlerCalled = func() data.HeaderHandler {
			return hdr
		}

		t.Run("round == 0 should not error", func(t *testing.T) {
			hdr.Round = 0
			err := hip.Validate(hdrInterceptedData, "")
			assert.Nil(t, err)
		})
		t.Run("round == 6076599 should not error", func(t *testing.T) {
			hdr.Round = 6076599
			err := hip.Validate(hdrInterceptedData, "")
			assert.Nil(t, err)
		})
		t.Run("round == 6076600 should error", func(t *testing.T) {
			hdr.Round = 6076600
			err := hip.Validate(hdrInterceptedData, "")
			assert.NotNil(t, err)
			assert.Equal(t, "header is in excluded range, shard 2, round 6076600, low 6076600, high 6630115", err.Error())
		})
		t.Run("round == 6300000 should error", func(t *testing.T) {
			hdr.Round = 6300000
			err := hip.Validate(hdrInterceptedData, "")
			assert.NotNil(t, err)
			assert.Equal(t, "header is in excluded range, shard 2, round 6300000, low 6076600, high 6630115", err.Error())
		})
		t.Run("round == 6630115 should error", func(t *testing.T) {
			hdr.Round = 6630115
			err := hip.Validate(hdrInterceptedData, "")
			assert.NotNil(t, err)
			assert.Equal(t, "header is in excluded range, shard 2, round 6630115, low 6076600, high 6630115", err.Error())
		})
		t.Run("round == 6630116 should not error", func(t *testing.T) {
			hdr.Round = 6630116
			err := hip.Validate(hdrInterceptedData, "")
			assert.Nil(t, err)
		})
		t.Run("round == 99999999 should not error", func(t *testing.T) {
			hdr.Round = 99999999
			err := hip.Validate(hdrInterceptedData, "")
			assert.Nil(t, err)
		})
	})

	t.Run("shard meta", func(t *testing.T) {
		hdr := &block.MetaBlock{}

		hdrInterceptedData := &struct {
			testscommon.InterceptedDataStub
			mock.GetHdrHandlerStub
		}{
			InterceptedDataStub: testscommon.InterceptedDataStub{
				HashCalled: func() []byte {
					return make([]byte, 0)
				},
			},
		}
		hdrInterceptedData.GetHdrHandlerStub.HeaderHandlerCalled = func() data.HeaderHandler {
			return hdr
		}

		t.Run("round == 0 should not error", func(t *testing.T) {
			hdr.Round = 0
			err := hip.Validate(hdrInterceptedData, "")
			assert.Nil(t, err)
		})
		t.Run("round == 6076599 should not error", func(t *testing.T) {
			hdr.Round = 6076599
			err := hip.Validate(hdrInterceptedData, "")
			assert.Nil(t, err)
		})
		t.Run("round == 6076600 should error", func(t *testing.T) {
			hdr.Round = 6076600
			err := hip.Validate(hdrInterceptedData, "")
			assert.NotNil(t, err)
			assert.Equal(t, "header is in excluded range, shard 4294967295, round 6076600, low 6076600, high 6630115", err.Error())
		})
		t.Run("round == 6300000 should error", func(t *testing.T) {
			hdr.Round = 6300000
			err := hip.Validate(hdrInterceptedData, "")
			assert.NotNil(t, err)
			assert.Equal(t, "header is in excluded range, shard 4294967295, round 6300000, low 6076600, high 6630115", err.Error())
		})
		t.Run("round == 6630115 should error", func(t *testing.T) {
			hdr.Round = 6630115
			err := hip.Validate(hdrInterceptedData, "")
			assert.NotNil(t, err)
			assert.Equal(t, "header is in excluded range, shard 4294967295, round 6630115, low 6076600, high 6630115", err.Error())
		})
		t.Run("round == 6630116 should not error", func(t *testing.T) {
			hdr.Round = 6630116
			err := hip.Validate(hdrInterceptedData, "")
			assert.Nil(t, err)
		})
		t.Run("round == 99999999 should not error", func(t *testing.T) {
			hdr.Round = 99999999
			err := hip.Validate(hdrInterceptedData, "")
			assert.Nil(t, err)
		})
	})

}

// ------- Save

func TestHdrInterceptorProcessor_SaveNilDataShouldErr(t *testing.T) {
	t.Parallel()

	hip, _ := processor.NewHdrInterceptorProcessor(createMockHdrArgument())

	_, err := hip.Save(nil, "", "", "")

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestHdrInterceptorProcessor_SaveShouldWork(t *testing.T) {
	t.Parallel()

	minNonceWithProof := uint64(2)
	hdrInterceptedData := &struct {
		testscommon.InterceptedDataStub
		mock.GetHdrHandlerStub
	}{
		InterceptedDataStub: testscommon.InterceptedDataStub{
			HashCalled: func() []byte {
				return []byte("hash")
			},
		},
		GetHdrHandlerStub: mock.GetHdrHandlerStub{
			HeaderHandlerCalled: func() data.HeaderHandler {
				return &testscommon.HeaderHandlerStub{
					GetNonceCalled: func() uint64 {
						return minNonceWithProof
					},
				}
			},
		},
	}

	wasAddedHeaders := false

	arg := createMockHdrArgument()
	arg.Headers = &mock.HeadersCacherStub{
		AddCalled: func(headerHash []byte, header data.HeaderHandler) {
			wasAddedHeaders = true
		},
	}
	arg.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.AndromedaFlag
		},
	}

	hip, _ := processor.NewHdrInterceptorProcessor(arg)
	chanCalled := make(chan struct{}, 1)
	hip.RegisterHandler(func(topic string, hash []byte, data interface{}) {
		chanCalled <- struct{}{}
	})

	_, err := hip.Save(hdrInterceptedData, "", "", "")

	assert.Nil(t, err)
	assert.True(t, wasAddedHeaders)

	timeout := time.Second * 2
	select {
	case <-chanCalled:
	case <-time.After(timeout):
		assert.Fail(t, "save did not notify handler in a timely fashion")
	}
}

func TestHdrInterceptorProcessor_RegisterHandlerNilHandler(t *testing.T) {
	t.Parallel()

	arg := createMockHdrArgument()
	hip, _ := processor.NewHdrInterceptorProcessor(arg)

	hip.RegisterHandler(nil)
	assert.Equal(t, 0, len(hip.RegisteredHandlers()))
}

// ------- IsInterfaceNil

func TestHdrInterceptorProcessor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var hip *processor.HdrInterceptorProcessor

	assert.True(t, check.IfNil(hip))
}
