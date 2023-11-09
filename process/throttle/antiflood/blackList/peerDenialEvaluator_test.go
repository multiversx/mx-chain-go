package blackList

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewPeerDenialEvaluator_NilPeerTimeCacherShouldErr(t *testing.T) {
	t.Parallel()

	pdc, err := NewPeerDenialEvaluator(
		nil,
		&testscommon.TimeCacheStub{},
		&mock.PeerShardMapperStub{},
	)

	assert.True(t, errors.Is(err, process.ErrNilBlackListCacher))
	assert.True(t, check.IfNil(pdc))
}

func TestNewPeerDenialEvaluator_NilTimeCacherShouldErr(t *testing.T) {
	t.Parallel()

	pdc, err := NewPeerDenialEvaluator(
		&mock.PeerBlackListHandlerStub{},
		nil,
		&mock.PeerShardMapperStub{},
	)

	assert.True(t, errors.Is(err, process.ErrNilBlackListCacher))
	assert.True(t, check.IfNil(pdc))
}

func TestNewPeerDenialEvaluator_NilPeerShardMapperShouldErr(t *testing.T) {
	t.Parallel()

	pdc, err := NewPeerDenialEvaluator(
		&mock.PeerBlackListHandlerStub{},
		&testscommon.TimeCacheStub{},
		nil,
	)

	assert.True(t, errors.Is(err, process.ErrNilPeerShardMapper))
	assert.True(t, check.IfNil(pdc))
}

func TestNewPeerDenialEvaluator_ShouldWork(t *testing.T) {
	t.Parallel()

	pdc, err := NewPeerDenialEvaluator(
		&mock.PeerBlackListHandlerStub{},
		&testscommon.TimeCacheStub{},
		&mock.PeerShardMapperStub{},
	)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(pdc))
}

func TestPeerDenialEvaluator_IsDeniedShouldWorkIfFoundInPids(t *testing.T) {
	t.Parallel()

	pdc, _ := NewPeerDenialEvaluator(
		&mock.PeerBlackListHandlerStub{
			HasCalled: func(pid core.PeerID) bool {
				return true
			},
		},
		&testscommon.TimeCacheStub{
			HasCalled: func(key string) bool {
				assert.Fail(t, "should have not reached this point")
				return false
			},
		},
		&mock.PeerShardMapperStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				assert.Fail(t, "should have not reached this point")
				return core.P2PPeerInfo{}
			},
		},
	)

	assert.True(t, pdc.IsDenied(""))
}

func TestPeerDenialEvaluator_IsDeniedShouldWorkIfNotFoundInPidsNorInPeerShardMapper(t *testing.T) {
	t.Parallel()

	pdc, _ := NewPeerDenialEvaluator(
		&mock.PeerBlackListHandlerStub{
			HasCalled: func(pid core.PeerID) bool {
				return false
			},
		},
		&testscommon.TimeCacheStub{
			HasCalled: func(key string) bool {
				assert.Fail(t, "should have not reached this point")
				return false
			},
		},
		&mock.PeerShardMapperStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				return core.P2PPeerInfo{}
			},
		},
	)

	assert.False(t, pdc.IsDenied(""))
}

func TestPeerDenialEvaluator_IsDeniedShouldWorkIfFoundInPk(t *testing.T) {
	t.Parallel()

	pdc, _ := NewPeerDenialEvaluator(
		&mock.PeerBlackListHandlerStub{
			HasCalled: func(pid core.PeerID) bool {
				return false
			},
		},
		&testscommon.TimeCacheStub{
			HasCalled: func(key string) bool {
				return true
			},
		},
		&mock.PeerShardMapperStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				return core.P2PPeerInfo{
					PkBytes: []byte("pk"),
				}
			},
		},
	)

	assert.True(t, pdc.IsDenied(""))
}

func TestPeerDenialEvaluator_UpsertPeerID(t *testing.T) {
	t.Parallel()

	upsertCalled := false
	pdc, _ := NewPeerDenialEvaluator(
		&mock.PeerBlackListHandlerStub{
			UpsertCalled: func(pid core.PeerID, span time.Duration) error {
				upsertCalled = true
				return nil
			},
		},
		&testscommon.TimeCacheStub{},
		&mock.PeerShardMapperStub{},
	)

	err := pdc.UpsertPeerID("", time.Second)
	assert.Nil(t, err)
	assert.True(t, upsertCalled)
}
