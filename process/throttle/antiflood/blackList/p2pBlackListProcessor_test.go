package blackList_test

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common/configs/dto"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/throttle/antiflood/blackList"
	"github.com/multiversx/mx-chain-go/testscommon/cache"

	"github.com/stretchr/testify/assert"
)

const selfPid = "current pid"

// -------- NewP2PQuotaBlacklistProcessor

func TestNewP2PQuotaBlacklistProcessor_NilCacherShouldErr(t *testing.T) {
	t.Parallel()

	pbp, err := blackList.NewP2PBlackListProcessor(
		nil,
		&mock.PeerBlackListHandlerStub{},
		1,
		1,
		dto.NumFloodingRoundsFastReacting,
		time.Second,
		"",
		selfPid,
		&testscommon.ProcessConfigsHandlerStub{},
	)

	assert.True(t, check.IfNil(pbp))
	assert.True(t, errors.Is(err, process.ErrNilCacher))
}

func TestNewP2PQuotaBlacklistProcessor_NilBlackListHandlerShouldErr(t *testing.T) {
	t.Parallel()

	pbp, err := blackList.NewP2PBlackListProcessor(
		cache.NewCacherStub(),
		nil,
		1,
		1,
		dto.NumFloodingRoundsFastReacting,
		time.Second,
		"",
		selfPid,
		&testscommon.ProcessConfigsHandlerStub{},
	)

	assert.True(t, check.IfNil(pbp))
	assert.True(t, errors.Is(err, process.ErrNilBlackListCacher))
}

func TestNewP2PQuotaBlacklistProcessor_InvalidThresholdNumReceivedFloodShouldErr(t *testing.T) {
	t.Parallel()

	pbp, err := blackList.NewP2PBlackListProcessor(
		cache.NewCacherStub(),
		&mock.PeerBlackListHandlerStub{},
		0,
		1,
		dto.NumFloodingRoundsFastReacting,
		time.Second,
		"",
		selfPid,
		&testscommon.ProcessConfigsHandlerStub{},
	)

	assert.True(t, check.IfNil(pbp))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewP2PQuotaBlacklistProcessor_InvalidThresholdSizeReceivedFloodShouldErr(t *testing.T) {
	t.Parallel()

	pbp, err := blackList.NewP2PBlackListProcessor(
		cache.NewCacherStub(),
		&mock.PeerBlackListHandlerStub{},
		1,
		0,
		dto.NumFloodingRoundsFastReacting,
		time.Second,
		"",
		selfPid,
		&testscommon.ProcessConfigsHandlerStub{},
	)

	assert.True(t, check.IfNil(pbp))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewP2PQuotaBlacklistProcessor_NilProcessConfigsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	pbp, err := blackList.NewP2PBlackListProcessor(
		cache.NewCacherStub(),
		&mock.PeerBlackListHandlerStub{},
		1,
		0,
		dto.NumFloodingRoundsFastReacting,
		time.Second,
		"",
		selfPid,
		nil,
	)

	require.True(t, check.IfNil(pbp))
	require.True(t, errors.Is(err, process.ErrNilProcessConfigsHandler))
}

func TestNewP2PQuotaBlacklistProcessor_EmptyVarNameForNumFloodingRoundsShouldErr(t *testing.T) {
	t.Parallel()

	pbp, err := blackList.NewP2PBlackListProcessor(
		cache.NewCacherStub(),
		&mock.PeerBlackListHandlerStub{},
		1,
		0,
		"",
		time.Second,
		"",
		selfPid,
		&testscommon.ProcessConfigsHandlerStub{},
	)

	require.True(t, check.IfNil(pbp))
	require.True(t, errors.Is(err, blackList.ErrEmptyConfigVarNameForNumFloodingRounds))
}

func TestNewP2PQuotaBlacklistProcessor_InvalidBanDurationShouldErr(t *testing.T) {
	t.Parallel()

	pbp, err := blackList.NewP2PBlackListProcessor(
		cache.NewCacherStub(),
		&mock.PeerBlackListHandlerStub{},
		1,
		1,
		dto.NumFloodingRoundsFastReacting,
		time.Millisecond,
		"",
		selfPid,
		&testscommon.ProcessConfigsHandlerStub{},
	)

	assert.True(t, check.IfNil(pbp))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewP2PQuotaBlacklistProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	pbp, err := blackList.NewP2PBlackListProcessor(
		cache.NewCacherStub(),
		&mock.PeerBlackListHandlerStub{},
		1,
		1,
		dto.NumFloodingRoundsFastReacting,
		time.Second,
		"",
		selfPid,
		&testscommon.ProcessConfigsHandlerStub{},
	)

	assert.False(t, check.IfNil(pbp))
	assert.Nil(t, err)
}

// ------- AddQuota

func TestP2PQuotaBlacklistProcessor_AddQuotaUnderThresholdShouldNotCallGetOrPut(t *testing.T) {
	t.Parallel()

	thresholdNum := uint32(10)
	thresholdSize := uint64(20)

	pbp, _ := blackList.NewP2PBlackListProcessor(
		&cache.CacherStub{
			GetCalled: func(key []byte) (interface{}, bool) {
				assert.Fail(t, "should not have called get")
				return nil, false
			},
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				assert.Fail(t, "should not have called put")
				return false
			},
		},
		&mock.PeerBlackListHandlerStub{},
		thresholdNum,
		thresholdSize,
		dto.NumFloodingRoundsFastReacting,
		time.Second,
		"",
		selfPid,
		&testscommon.ProcessConfigsHandlerStub{},
	)

	pbp.AddQuota("identifier", thresholdNum-1, thresholdSize-1, 1, 1)
}

func TestP2PQuotaBlacklistProcessor_AddQuotaOverThresholdInexistentDataOnGetShouldPutOne(t *testing.T) {
	t.Parallel()

	thresholdNum := uint32(10)
	thresholdSize := uint64(20)

	putCalled := false
	identifier := core.PeerID("identifier")
	pbp, _ := blackList.NewP2PBlackListProcessor(
		&cache.CacherStub{
			GetCalled: func(key []byte) (interface{}, bool) {
				return nil, false
			},
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				putCalled = true
				assert.Equal(t, uint32(1), value)
				assert.Equal(t, identifier, core.PeerID(key))

				return false
			},
		},
		&mock.PeerBlackListHandlerStub{},
		thresholdNum,
		thresholdSize,
		dto.NumFloodingRoundsFastReacting,
		time.Second,
		"",
		selfPid,
		&testscommon.ProcessConfigsHandlerStub{},
	)

	pbp.AddQuota(identifier, thresholdNum, thresholdSize, 1, 1)

	assert.True(t, putCalled)
}

func TestP2PQuotaBlacklistProcessor_AddQuotaOverThresholdDataNotValidOnGetShouldPutOne(t *testing.T) {
	t.Parallel()

	thresholdNum := uint32(10)
	thresholdSize := uint64(20)

	putCalled := false
	identifier := core.PeerID("identifier")
	pbp, _ := blackList.NewP2PBlackListProcessor(
		&cache.CacherStub{
			GetCalled: func(key []byte) (interface{}, bool) {
				return "invalid data", true
			},
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				putCalled = true
				assert.Equal(t, uint32(1), value)
				assert.Equal(t, identifier, core.PeerID(key))

				return false
			},
		},
		&mock.PeerBlackListHandlerStub{},
		thresholdNum,
		thresholdSize,
		dto.NumFloodingRoundsFastReacting,
		time.Second,
		"",
		selfPid,
		&testscommon.ProcessConfigsHandlerStub{},
	)

	pbp.AddQuota(identifier, thresholdNum, thresholdSize, 1, 1)

	assert.True(t, putCalled)
}

func TestP2PQuotaBlacklistProcessor_AddQuotaShouldIncrement(t *testing.T) {
	t.Parallel()

	thresholdNum := uint32(10)
	thresholdSize := uint64(20)

	putCalled := false
	identifier := core.PeerID("identifier")
	existingValue := uint32(445)
	pbp, _ := blackList.NewP2PBlackListProcessor(
		&cache.CacherStub{
			GetCalled: func(key []byte) (interface{}, bool) {
				return existingValue, true
			},
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				putCalled = true
				assert.Equal(t, existingValue+1, value)
				assert.Equal(t, identifier, core.PeerID(key))

				return false
			},
		},
		&mock.PeerBlackListHandlerStub{},
		thresholdNum,
		thresholdSize,
		dto.NumFloodingRoundsFastReacting,
		time.Second,
		"",
		selfPid,
		&testscommon.ProcessConfigsHandlerStub{},
	)

	pbp.AddQuota(identifier, thresholdNum, thresholdSize, 1, 1)

	assert.True(t, putCalled)
}

func TestP2PQuotaBlacklistProcessor_AddQuotaForSelfShouldNotIncrement(t *testing.T) {
	t.Parallel()

	thresholdNum := uint32(10)
	thresholdSize := uint64(20)

	putCalled := false
	existingValue := uint32(445)
	pbp, _ := blackList.NewP2PBlackListProcessor(
		&cache.CacherStub{
			GetCalled: func(key []byte) (interface{}, bool) {
				return existingValue, true
			},
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
				putCalled = true
				return false
			},
		},
		&mock.PeerBlackListHandlerStub{},
		thresholdNum,
		thresholdSize,
		dto.NumFloodingRoundsFastReacting,
		time.Second,
		"",
		selfPid,
		&testscommon.ProcessConfigsHandlerStub{},
	)

	pbp.AddQuota(selfPid, thresholdNum, thresholdSize, 1, 1)

	assert.False(t, putCalled)
}

// ------- ResetStatistics

func TestP2PQuotaBlacklistProcessor_ResetStatisticsRemoveNilValueKey(t *testing.T) {
	t.Parallel()

	thresholdNum := uint32(10)
	thresholdSize := uint64(20)

	nilValKey := "nil val key"
	removedCalled := false
	pbp, _ := blackList.NewP2PBlackListProcessor(
		&cache.CacherStub{
			KeysCalled: func() [][]byte {
				return [][]byte{[]byte(nilValKey)}
			},
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
			RemoveCalled: func(key []byte) {
				if string(key) == nilValKey {
					removedCalled = true
				}
			},
		},
		&mock.PeerBlackListHandlerStub{},
		thresholdNum,
		thresholdSize,
		dto.NumFloodingRoundsFastReacting,
		time.Second,
		"",
		selfPid,
		&testscommon.ProcessConfigsHandlerStub{},
	)

	pbp.ResetStatistics()

	assert.True(t, removedCalled)
}

func TestP2PQuotaBlacklistProcessor_ResetStatisticsShouldRemoveInvalidValueKey(t *testing.T) {
	t.Parallel()

	thresholdNum := uint32(10)
	thresholdSize := uint64(20)

	invalidValKey := "invalid val key"
	removedCalled := false
	pbp, _ := blackList.NewP2PBlackListProcessor(
		&cache.CacherStub{
			KeysCalled: func() [][]byte {
				return [][]byte{[]byte(invalidValKey)}
			},
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				return "invalid value", true
			},
			RemoveCalled: func(key []byte) {
				if string(key) == invalidValKey {
					removedCalled = true
				}
			},
		},
		&mock.PeerBlackListHandlerStub{},
		thresholdNum,
		thresholdSize,
		dto.NumFloodingRoundsFastReacting,
		time.Second,
		"",
		selfPid,
		&testscommon.ProcessConfigsHandlerStub{},
	)

	pbp.ResetStatistics()

	assert.True(t, removedCalled)
}

func TestP2PQuotaBlacklistProcessor_ResetStatisticsUnderNumFloodingRoundsShouldNotBlackList(t *testing.T) {
	t.Parallel()

	thresholdNum := uint32(10)
	thresholdSize := uint64(20)
	numFloodingRounds := uint32(30)

	key := "key"
	removedCalled := false
	upsertCalled := false
	duration := time.Second * 3892
	pbp, _ := blackList.NewP2PBlackListProcessor(
		&cache.CacherStub{
			KeysCalled: func() [][]byte {
				return [][]byte{[]byte(key)}
			},
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				return numFloodingRounds - 2, true
			},
			RemoveCalled: func(key []byte) {
				removedCalled = true
			},
		},
		&mock.PeerBlackListHandlerStub{
			UpsertCalled: func(pid core.PeerID, span time.Duration) error {
				upsertCalled = true
				assert.Equal(t, duration, span)

				return nil
			},
		},
		thresholdNum,
		thresholdSize,
		dto.NumFloodingRoundsFastReacting,
		duration,
		"",
		selfPid,
		&testscommon.ProcessConfigsHandlerStub{
			GetValueCalled: func(variable dto.ConfigVariable) uint64 {
				return 0
			},
		},
	)

	pbp.ResetStatistics()

	assert.False(t, removedCalled)
	assert.False(t, upsertCalled)
}

func TestP2PQuotaBlacklistProcessor_ResetStatisticsOverNumFloodingRoundsShouldBlackList(t *testing.T) {
	t.Parallel()

	thresholdNum := uint32(10)
	thresholdSize := uint64(20)
	numFloodingRounds := uint32(30)

	key := "key"
	removedCalled := false
	upsertCalled := false
	duration := time.Second * 3892
	pbp, _ := blackList.NewP2PBlackListProcessor(
		&cache.CacherStub{
			KeysCalled: func() [][]byte {
				return [][]byte{[]byte(key)}
			},
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				return numFloodingRounds, true
			},
			RemoveCalled: func(key []byte) {
				removedCalled = true
			},
		},
		&mock.PeerBlackListHandlerStub{
			UpsertCalled: func(pid core.PeerID, span time.Duration) error {
				upsertCalled = true
				assert.Equal(t, duration, span)

				return nil
			},
		},
		thresholdNum,
		thresholdSize,
		dto.NumFloodingRoundsFastReacting,
		duration,
		"",
		selfPid,
		&testscommon.ProcessConfigsHandlerStub{
			GetValueCalled: func(variable dto.ConfigVariable) uint64 {
				require.Equal(t, dto.NumFloodingRoundsFastReacting, variable)
				return uint64(numFloodingRounds)
			},
		},
	)

	pbp.ResetStatistics()

	assert.True(t, removedCalled)
	assert.True(t, upsertCalled)
}
