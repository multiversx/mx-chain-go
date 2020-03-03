package blackList_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/blackList"
	"github.com/stretchr/testify/assert"
)

//-------- NewP2PQuotaBlacklistProcessor

func TestNewP2PQuotaBlacklistProcessor_NilCacherShouldErr(t *testing.T) {
	t.Parallel()

	pbp, err := blackList.NewP2PBlackListProcessor(
		nil,
		&mock.BlackListHandlerStub{},
		1,
		1,
		1,
	)

	assert.True(t, check.IfNil(pbp))
	assert.True(t, errors.Is(err, process.ErrNilCacher))
}

func TestNewP2PQuotaBlacklistProcessor_NilBlackListHandlerShouldErr(t *testing.T) {
	t.Parallel()

	pbp, err := blackList.NewP2PBlackListProcessor(
		&mock.CacherStub{},
		nil,
		1,
		1,
		1,
	)

	assert.True(t, check.IfNil(pbp))
	assert.True(t, errors.Is(err, process.ErrNilBlackListHandler))
}

func TestNewP2PQuotaBlacklistProcessor_InvalidThresholdNumReceivedFloodShouldErr(t *testing.T) {
	t.Parallel()

	pbp, err := blackList.NewP2PBlackListProcessor(
		&mock.CacherStub{},
		&mock.BlackListHandlerStub{},
		0,
		1,
		1,
	)

	assert.True(t, check.IfNil(pbp))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewP2PQuotaBlacklistProcessor_InvalidThresholdSizeReceivedFloodShouldErr(t *testing.T) {
	t.Parallel()

	pbp, err := blackList.NewP2PBlackListProcessor(
		&mock.CacherStub{},
		&mock.BlackListHandlerStub{},
		1,
		0,
		1,
	)

	assert.True(t, check.IfNil(pbp))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewP2PQuotaBlacklistProcessor_InvalidNumFloodingRoundsShouldErr(t *testing.T) {
	t.Parallel()

	pbp, err := blackList.NewP2PBlackListProcessor(
		&mock.CacherStub{},
		&mock.BlackListHandlerStub{},
		1,
		1,
		0,
	)

	assert.True(t, check.IfNil(pbp))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewP2PQuotaBlacklistProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	pbp, err := blackList.NewP2PBlackListProcessor(
		&mock.CacherStub{},
		&mock.BlackListHandlerStub{},
		1,
		1,
		1,
	)

	assert.False(t, check.IfNil(pbp))
	assert.Nil(t, err)
}

//------- AddQuota

func TestP2PQuotaBlacklistProcessor_AddQuotaUnderThresholdShouldNotCallGetOrPut(t *testing.T) {
	t.Parallel()

	thresholdNum := uint32(10)
	thresholdSize := uint64(20)

	pbp, _ := blackList.NewP2PBlackListProcessor(
		&mock.CacherStub{
			GetCalled: func(key []byte) (interface{}, bool) {
				assert.Fail(t, "should not have called get")
				return nil, false
			},
			PutCalled: func(key []byte, value interface{}) (evicted bool) {
				assert.Fail(t, "should not have called put")
				return false
			},
		},
		&mock.BlackListHandlerStub{},
		thresholdNum,
		thresholdSize,
		1,
	)

	pbp.AddQuota("identifier", thresholdNum-1, thresholdSize-1, 1, 1)
}

func TestP2PQuotaBlacklistProcessor_AddQuotaOverThresholdInexistentDataOnGetShouldPutOne(t *testing.T) {
	t.Parallel()

	thresholdNum := uint32(10)
	thresholdSize := uint64(20)

	putCalled := false
	identifier := "identifier"
	pbp, _ := blackList.NewP2PBlackListProcessor(
		&mock.CacherStub{
			GetCalled: func(key []byte) (interface{}, bool) {
				return nil, false
			},
			PutCalled: func(key []byte, value interface{}) (evicted bool) {
				putCalled = true
				assert.Equal(t, uint32(1), value)
				assert.Equal(t, identifier, string(key))

				return false
			},
		},
		&mock.BlackListHandlerStub{},
		thresholdNum,
		thresholdSize,
		1,
	)

	pbp.AddQuota(identifier, thresholdNum, thresholdSize, 1, 1)

	assert.True(t, putCalled)
}

func TestP2PQuotaBlacklistProcessor_AddQuotaOverThresholdDataNotValidOnGetShouldPutOne(t *testing.T) {
	t.Parallel()

	thresholdNum := uint32(10)
	thresholdSize := uint64(20)

	putCalled := false
	identifier := "identifier"
	pbp, _ := blackList.NewP2PBlackListProcessor(
		&mock.CacherStub{
			GetCalled: func(key []byte) (interface{}, bool) {
				return "invalid data", true
			},
			PutCalled: func(key []byte, value interface{}) (evicted bool) {
				putCalled = true
				assert.Equal(t, uint32(1), value)
				assert.Equal(t, identifier, string(key))

				return false
			},
		},
		&mock.BlackListHandlerStub{},
		thresholdNum,
		thresholdSize,
		1,
	)

	pbp.AddQuota(identifier, thresholdNum, thresholdSize, 1, 1)

	assert.True(t, putCalled)
}

func TestP2PQuotaBlacklistProcessor_AddQuotaShouldIncrement(t *testing.T) {
	t.Parallel()

	thresholdNum := uint32(10)
	thresholdSize := uint64(20)

	putCalled := false
	identifier := "identifier"
	existingValue := uint32(445)
	pbp, _ := blackList.NewP2PBlackListProcessor(
		&mock.CacherStub{
			GetCalled: func(key []byte) (interface{}, bool) {
				return existingValue, true
			},
			PutCalled: func(key []byte, value interface{}) (evicted bool) {
				putCalled = true
				assert.Equal(t, existingValue+1, value)
				assert.Equal(t, identifier, string(key))

				return false
			},
		},
		&mock.BlackListHandlerStub{},
		thresholdNum,
		thresholdSize,
		1,
	)

	pbp.AddQuota(identifier, thresholdNum, thresholdSize, 1, 1)

	assert.True(t, putCalled)
}

//------- ResetStatistics

func TestP2PQuotaBlacklistProcessor_ResetStatisticsRemoveNilValueKey(t *testing.T) {
	t.Parallel()

	thresholdNum := uint32(10)
	thresholdSize := uint64(20)

	nilValKey := "nil val key"
	removedCalled := false
	pbp, _ := blackList.NewP2PBlackListProcessor(
		&mock.CacherStub{
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
		&mock.BlackListHandlerStub{},
		thresholdNum,
		thresholdSize,
		1,
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
		&mock.CacherStub{
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
		&mock.BlackListHandlerStub{},
		thresholdNum,
		thresholdSize,
		1,
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
	addToBlacklistCalled := false
	pbp, _ := blackList.NewP2PBlackListProcessor(
		&mock.CacherStub{
			KeysCalled: func() [][]byte {
				return [][]byte{[]byte(key)}
			},
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				return numFloodingRounds - 1, true
			},
			RemoveCalled: func(key []byte) {
				removedCalled = true
			},
		},
		&mock.BlackListHandlerStub{
			AddCalled: func(key string) error {
				addToBlacklistCalled = true

				return nil
			},
		},
		thresholdNum,
		thresholdSize,
		numFloodingRounds,
	)

	pbp.ResetStatistics()

	assert.False(t, removedCalled)
	assert.False(t, addToBlacklistCalled)
}

func TestP2PQuotaBlacklistProcessor_ResetStatisticsOverNumFloodingRoundsShouldBlackList(t *testing.T) {
	t.Parallel()

	thresholdNum := uint32(10)
	thresholdSize := uint64(20)
	numFloodingRounds := uint32(30)

	key := "key"
	removedCalled := false
	addToBlacklistCalled := false
	pbp, _ := blackList.NewP2PBlackListProcessor(
		&mock.CacherStub{
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
		&mock.BlackListHandlerStub{
			AddCalled: func(key string) error {
				addToBlacklistCalled = true

				return nil
			},
		},
		thresholdNum,
		thresholdSize,
		numFloodingRounds,
	)

	pbp.ResetStatistics()

	assert.True(t, removedCalled)
	assert.True(t, addToBlacklistCalled)
}
