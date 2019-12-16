package discovery

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

const (
	testCPSVal = 10
)

// don't break the interface
var (
	_ host.Host = &hostDecorator{}
)

func TestDecoratorConstructor(t *testing.T) {
	ctx := context.Background()
	hst := &mock.ConnectableHostStub{}

	cpsLow := uint32(hardCPSLimitLow - 1)
	cpsHigh := uint32(hardCPSLimitHigh + 1)
	cpsAvg := uint32((hardCPSLimitLow + hardCPSLimitHigh) / 2)

	testTbl := []struct {
		h   host.Host
		ctx context.Context
		cps uint32
		to  time.Duration

		expDecNil bool
		expErrNil bool
		expCps    int
		name      string
	}{
		{h: nil, ctx: nil, cps: 0, to: 0, expDecNil: true, expErrNil: false, expCps: 0, name: "All zero"},
		{h: nil, ctx: ctx, cps: 0, to: time.Second, expDecNil: true, expErrNil: false, expCps: 0, name: "Nil Host"},
		{h: hst, ctx: nil, cps: 0, to: time.Second, expDecNil: true, expErrNil: false, expCps: 0, name: "Nil Context"},
		{h: hst, ctx: ctx, cps: 0, to: -1, expDecNil: true, expErrNil: false, expCps: 0, name: "Bad timeout"},
		{h: hst, ctx: ctx, cps: cpsLow, to: 0, expDecNil: true, expErrNil: false, expCps: hardCPSLimitLow, name: "Low cps"},
		{h: hst, ctx: ctx, cps: cpsHigh, to: 0, expDecNil: true, expErrNil: false, expCps: hardCPSLimitHigh, name: "High cps"},
		{h: hst, ctx: ctx, cps: cpsAvg, to: 0, expDecNil: false, expErrNil: true, expCps: int(cpsAvg), name: "All good"},
	}

	for _, tc := range testTbl {
		t.Run(tc.name, func(t *testing.T) {
			d, err := NewHostDecorator(tc.h, tc.ctx, tc.cps, tc.to)

			if tc.expDecNil {
				assert.Nil(t, d)
			} else {
				assert.NotNil(t, d)
				assert.Equal(t, cap(d.acceptedChan), tc.expCps)
			}

			if tc.expErrNil {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestLimiting(t *testing.T) {

	connCount := 0
	mtx := sync.Mutex{}

	host := &mock.ConnectableHostStub{
		ConnectCalled: func(context.Context, peer.AddrInfo) error {
			mtx.Lock()
			defer mtx.Unlock()
			connCount++
			return nil
		},
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	start := time.Now()

	dec, err := NewHostDecorator(host, ctx, testCPSVal, 100*time.Microsecond)
	assert.Nil(t, err)
	assert.NotNil(t, dec)

	for i := 0; i < 100; i++ {
		go func() { _ = dec.Connect(ctx, peer.AddrInfo{}) }()
	}
	time.Sleep(time.Second)
	dur := time.Since(start)

	mtx.Lock()
	avgConns := float64(connCount) / (float64(dur) / float64(time.Second))
	mtx.Unlock()

	assert.True(t, avgConns < float64(testCPSVal))
	go func() { _ = dec.Connect(ctx, peer.AddrInfo{}) }()
	cancelCtx()
}

func TestPauseResume(t *testing.T) {

	connCount := 0

	host := &mock.ConnectableHostStub{
		ConnectCalled: func(context.Context, peer.AddrInfo) error {
			connCount++
			return nil
		},
	}

	ctx, cancelCtx := context.WithCancel(context.Background())

	dec, err := NewHostDecorator(host, ctx, testCPSVal, 100*time.Microsecond)
	assert.Nil(t, err)
	assert.NotNil(t, dec)
	assert.True(t, len(dec.acceptedChan) < testCPSVal)
	time.Sleep(time.Second)
	assert.Equal(t, len(dec.acceptedChan), testCPSVal)
	time.Sleep(time.Second)
	assert.Equal(t, len(dec.acceptedChan), testCPSVal)
	cancelCtx()
}
