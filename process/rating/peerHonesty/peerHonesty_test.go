package peerHonesty

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

// createMockPeerHonestyConfig creates a peer honesty config with reasonable values
func createMockPeerHonestyConfig() config.PeerHonestyConfig {
	return config.PeerHonestyConfig{
		DecayCoefficient:             0.9779,
		DecayUpdateIntervalInSeconds: 10,
		MaxScore:                     100,
		MinScore:                     -100,
		BadPeerThreshold:             -80,
		UnitValue:                    1.0,
	}
}

func TestNewP2pPeerHonesty_NilCacheShouldErr(t *testing.T) {
	t.Parallel()

	pph, err := NewP2pPeerHonesty(
		createMockPeerHonestyConfig(),
		&testscommon.TimeCacheStub{},
		nil,
	)

	assert.True(t, check.IfNil(pph))
	assert.True(t, errors.Is(err, process.ErrNilCacher))
}

func TestNewP2pPeerHonesty_NilBlacklistedPkCacheShouldErr(t *testing.T) {
	t.Parallel()

	pph, err := NewP2pPeerHonesty(
		createMockPeerHonestyConfig(),
		nil,
		&testscommon.CacherStub{},
	)

	assert.True(t, check.IfNil(pph))
	assert.True(t, errors.Is(err, process.ErrNilBlackListedPkCache))
}

func TestNewP2pPeerHonesty_InvalidDecayCoefficientShouldErr(t *testing.T) {
	t.Parallel()

	cfg := createMockPeerHonestyConfig()
	cfg.DecayCoefficient = -0.1
	pph, err := NewP2pPeerHonesty(
		cfg,
		&testscommon.TimeCacheStub{},
		&testscommon.CacherStub{},
	)

	assert.True(t, check.IfNil(pph))
	assert.True(t, errors.Is(err, process.ErrInvalidDecayCoefficient))
}

func TestNewP2pPeerHonesty_InvalidDecayUpdateIntervalShouldErr(t *testing.T) {
	t.Parallel()

	cfg := createMockPeerHonestyConfig()
	cfg.DecayUpdateIntervalInSeconds = 0
	pph, err := NewP2pPeerHonesty(
		cfg,
		&testscommon.TimeCacheStub{},
		&testscommon.CacherStub{},
	)

	assert.True(t, check.IfNil(pph))
	assert.True(t, errors.Is(err, process.ErrInvalidDecayIntervalInSeconds))
}

func TestNewP2pPeerHonesty_InvalidMinScoreShouldErr(t *testing.T) {
	t.Parallel()

	cfg := createMockPeerHonestyConfig()
	cfg.MinScore = 1
	pph, err := NewP2pPeerHonesty(
		cfg,
		&testscommon.TimeCacheStub{},
		&testscommon.CacherStub{},
	)

	assert.True(t, check.IfNil(pph))
	assert.True(t, errors.Is(err, process.ErrInvalidMinScore))
}

func TestNewP2pPeerHonesty_InvalidMaxScoreShouldErr(t *testing.T) {
	t.Parallel()

	cfg := createMockPeerHonestyConfig()
	cfg.MaxScore = -1
	pph, err := NewP2pPeerHonesty(
		cfg,
		&testscommon.TimeCacheStub{},
		&testscommon.CacherStub{},
	)

	assert.True(t, check.IfNil(pph))
	assert.True(t, errors.Is(err, process.ErrInvalidMaxScore))
}

func TestNewP2pPeerHonesty_InvalidUnitValueShouldErr(t *testing.T) {
	t.Parallel()

	cfg := createMockPeerHonestyConfig()
	cfg.UnitValue = -1
	pph, err := NewP2pPeerHonesty(
		cfg,
		&testscommon.TimeCacheStub{},
		&testscommon.CacherStub{},
	)

	assert.True(t, check.IfNil(pph))
	assert.True(t, errors.Is(err, process.ErrInvalidUnitValue))
}

func TestNewP2pPeerHonesty_InvalidBadPeerThresholdShouldErr(t *testing.T) {
	t.Parallel()

	cfg := createMockPeerHonestyConfig()
	cfg.BadPeerThreshold = cfg.MinScore
	pph, err := NewP2pPeerHonesty(
		cfg,
		&testscommon.TimeCacheStub{},
		&testscommon.CacherStub{},
	)

	assert.True(t, check.IfNil(pph))
	assert.True(t, errors.Is(err, process.ErrInvalidBadPeerThreshold))
}

func TestNewP2pPeerHonesty_ShouldWork(t *testing.T) {
	t.Parallel()

	cfg := createMockPeerHonestyConfig()
	pph, err := NewP2pPeerHonesty(
		cfg,
		&testscommon.TimeCacheStub{},
		&testscommon.CacherStub{},
	)

	assert.False(t, check.IfNil(pph))
	assert.Nil(t, err)
}

func TestP2pPeerHonesty_Close(t *testing.T) {
	t.Parallel()

	cfg := createMockPeerHonestyConfig()
	cfg.DecayUpdateIntervalInSeconds = 1
	numCalls := int32(0)
	handler := func() {
		atomic.AddInt32(&numCalls, 1)
	}
	pph, _ := NewP2pPeerHonestyWithCustomExecuteDelayFunction(
		cfg,
		&testscommon.TimeCacheStub{},
		&testscommon.CacherStub{},
		handler,
	)

	time.Sleep(time.Second*2 + time.Millisecond*100) // this will call the handler 3 times

	err := pph.Close()
	assert.Nil(t, err)

	time.Sleep(time.Second*2 + time.Millisecond*100)
	calls := atomic.LoadInt32(&numCalls)
	assert.Equal(t, int32(2), calls)
}

func TestP2pPeerHonesty_ChangeScoreShouldWork(t *testing.T) {
	t.Parallel()

	cfg := createMockPeerHonestyConfig()
	cfg.UnitValue = 4
	pph, _ := NewP2pPeerHonesty(
		cfg,
		&testscommon.TimeCacheStub{},
		testscommon.NewCacherMock(),
	)

	pk := "pk"
	topic := "topic"
	units := 2
	pph.ChangeScore(pk, topic, units)

	ps := pph.Get(pk)
	assert.Equal(t, 1, len(ps.scoresByTopic))
	assert.Equal(t, float64(units)*cfg.UnitValue, ps.scoresByTopic[topic])
}

func TestP2pPeerHonesty_DoubleChangeScoreShouldWork(t *testing.T) {
	t.Parallel()

	cfg := createMockPeerHonestyConfig()
	cfg.UnitValue = 4
	pph, _ := NewP2pPeerHonesty(
		cfg,
		&testscommon.TimeCacheStub{},
		testscommon.NewCacherMock(),
	)

	pk := "pk"
	topic := "topic"
	units := 2
	pph.ChangeScore(pk, topic, units)
	pph.ChangeScore(pk, topic, units)

	ps := pph.Get(pk)
	assert.Equal(t, 1, len(ps.scoresByTopic))
	assert.Equal(t, float64(units+units)*cfg.UnitValue, ps.scoresByTopic[topic])
}

func TestP2pPeerHonesty_CheckBlacklistNotBlacklisted(t *testing.T) {
	t.Parallel()

	cfg := createMockPeerHonestyConfig()
	cfg.UnitValue = 4
	hasCalled := false
	upsertCalled := false
	pph, _ := NewP2pPeerHonesty(
		cfg,
		&testscommon.TimeCacheStub{
			HasCalled: func(key string) bool {
				hasCalled = true
				return false
			},
			UpsertCalled: func(key string, span time.Duration) error {
				upsertCalled = true
				return nil
			},
		},
		testscommon.NewCacherMock(),
	)

	pk := "pk"
	topic := "topic"
	units := 2
	pph.ChangeScore(pk, topic, units)
	pph.ChangeScore(pk, topic, units)

	assert.False(t, hasCalled)
	assert.False(t, upsertCalled)
}

func TestP2pPeerHonesty_CheckBlacklistMaxScoreReached(t *testing.T) {
	t.Parallel()

	cfg := createMockPeerHonestyConfig()
	cfg.UnitValue = 4
	hasCalled := false
	upsertCalled := false
	pph, _ := NewP2pPeerHonesty(
		cfg,
		&testscommon.TimeCacheStub{
			HasCalled: func(key string) bool {
				hasCalled = true
				return false
			},
			UpsertCalled: func(key string, span time.Duration) error {
				upsertCalled = true
				return nil
			},
		},
		testscommon.NewCacherMock(),
	)

	pk := "pk"
	topic := "topic"
	units := int(cfg.MaxScore) + 1
	pph.ChangeScore(pk, topic, units)

	ps := pph.Get(pk)
	assert.Equal(t, 1, len(ps.scoresByTopic))
	assert.Equal(t, cfg.MaxScore, ps.scoresByTopic[topic])

	assert.False(t, hasCalled)
	assert.False(t, upsertCalled)
}

func TestP2pPeerHonesty_CheckBlacklistMinScoreReached(t *testing.T) {
	t.Parallel()

	cfg := createMockPeerHonestyConfig()
	cfg.UnitValue = 4
	hasCalled := false
	upsertCalled := false
	pph, _ := NewP2pPeerHonesty(
		cfg,
		&testscommon.TimeCacheStub{
			HasCalled: func(key string) bool {
				hasCalled = true
				return false
			},
			UpsertCalled: func(key string, span time.Duration) error {
				upsertCalled = true
				return nil
			},
		},
		testscommon.NewCacherMock(),
	)

	pk := "pk"
	topic := "topic"
	units := int(cfg.MinScore) - 1
	pph.ChangeScore(pk, topic, units)

	ps := pph.Get(pk)
	assert.Equal(t, 1, len(ps.scoresByTopic))
	assert.Equal(t, cfg.MinScore, ps.scoresByTopic[topic])

	assert.True(t, hasCalled)
	assert.True(t, upsertCalled)
}

func TestP2pPeerHonesty_CheckBlacklistHasShouldNotCallUpsert(t *testing.T) {
	t.Parallel()

	cfg := createMockPeerHonestyConfig()
	cfg.UnitValue = 4
	hasCalled := false
	upsertCalled := false
	pph, _ := NewP2pPeerHonesty(
		cfg,
		&testscommon.TimeCacheStub{
			HasCalled: func(key string) bool {
				hasCalled = true
				return true
			},
			UpsertCalled: func(key string, span time.Duration) error {
				upsertCalled = true
				return nil
			},
		},
		testscommon.NewCacherMock(),
	)

	pk := "pk"
	topic := "topic"
	units := int(cfg.MinScore) - 1
	pph.ChangeScore(pk, topic, units)

	assert.True(t, hasCalled)
	assert.False(t, upsertCalled)
}

func TestP2pPeerHonesty_CheckBlacklistUpsertErrorsShouldWork(t *testing.T) {
	t.Parallel()

	cfg := createMockPeerHonestyConfig()
	cfg.UnitValue = 4
	upsertCalled := false
	pph, _ := NewP2pPeerHonesty(
		cfg,
		&testscommon.TimeCacheStub{
			HasCalled: func(key string) bool {
				return false
			},
			UpsertCalled: func(key string, span time.Duration) error {
				upsertCalled = true
				return errors.New("expected error")
			},
		},
		testscommon.NewCacherMock(),
	)

	pk := "pk"
	topic := "topic"
	units := int(cfg.MinScore) - 1
	pph.ChangeScore(pk, topic, units)

	assert.True(t, upsertCalled)
}

func TestP2pPeerHonesty_ApplyDecay(t *testing.T) {
	t.Parallel()

	cfg := createMockPeerHonestyConfig()
	pph, _ := NewP2pPeerHonesty(
		cfg,
		&testscommon.TimeCacheStub{},
		testscommon.NewCacherMock(),
	)

	pks := []string{"pkMin", "pkMax", "pkNearZero", "pkZero", "pkValue"}
	initial := []float64{cfg.MinScore, cfg.MaxScore, approximateZero / 2, 0, 28}
	topic := "topic"

	for idx := range pks {
		pph.Put(pks[idx], topic, initial[idx])
	}

	pph.applyDecay()

	checkScore(t, pph, pks[0], topic, cfg.MinScore*cfg.DecayCoefficient)
	checkScore(t, pph, pks[1], topic, cfg.MaxScore*cfg.DecayCoefficient)
	checkScore(t, pph, pks[2], topic, 0)
	checkScore(t, pph, pks[3], topic, 0)
	checkScore(t, pph, pks[4], topic, initial[4]*cfg.DecayCoefficient)
}

func TestP2pPeerHonesty_ApplyDecayWillEventuallyGoTheScoreToZero(t *testing.T) {
	t.Parallel()

	cfg := createMockPeerHonestyConfig()
	cfg.MaxScore = 100
	cfg.UnitValue = 1
	cfg.DecayCoefficient = 0.9779
	pph, _ := NewP2pPeerHonesty(
		cfg,
		&testscommon.TimeCacheStub{},
		testscommon.NewCacherMock(),
	)

	pk := "pk"
	topic := "topic"
	pph.Put(pk, topic, cfg.MaxScore)

	expectedDecaysToBeZero := 722 // (at 10 seconds decay interval this will be ~2h)
	for i := 0; i < expectedDecaysToBeZero; i++ {
		pph.applyDecay()
	}

	checkScore(t, pph, pk, topic, 0)
}

func checkScore(t *testing.T, pph *p2pPeerHonesty, pk string, topic string, value float64) {
	ps := pph.Get(pk)
	assert.Equal(t, value, ps.scoresByTopic[topic])
}
