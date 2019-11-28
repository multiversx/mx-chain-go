package discovery

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/event"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	protocol "github.com/libp2p/go-libp2p-core/protocol"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multiaddr"
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
		{h: hst, ctx: nil, cps: 0, to: time.Second, expDecNil: true, expErrNil: false, expCps: 0, name: "Nil Host"},
		{h: hst, ctx: ctx, cps: 0, to: -1, expDecNil: true, expErrNil: false, expCps: 0, name: "Zero timeout"},
		{h: hst, ctx: ctx, cps: cpsLow, to: 0, expDecNil: false, expErrNil: true, expCps: hardCPSLimitLow, name: "Low cps"},
		{h: hst, ctx: ctx, cps: cpsHigh, to: 0, expDecNil: false, expErrNil: true, expCps: hardCPSLimitHigh, name: "High cps"},
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

func TestDecoratorCallRelay(t *testing.T) {
	ctx := context.Background()

	callCounters := struct {
		EventBusCalled              int
		IDCalled                    int
		PeerstoreCalled             int
		AddrsCalled                 int
		NetworkCalled               int
		MuxCalled                   int
		ConnectCalled               int
		SetStreamHandlerCalled      int
		SetStreamHandlerMatchCalled int
		RemoveStreamHandlerCalled   int
		NewStreamCalled             int
		CloseCalled                 int
		ConnManagerCalled           int
	}{}

	hst := &mock.ConnectableHostStub{
		EventBusCalled: func() event.Bus {
			callCounters.EventBusCalled++
			return nil
		},
		IDCalled: func() peer.ID {
			callCounters.IDCalled++
			return peer.ID("")
		},
		PeerstoreCalled: func() peerstore.Peerstore {
			callCounters.PeerstoreCalled++
			return nil
		},
		AddrsCalled: func() []multiaddr.Multiaddr {
			callCounters.AddrsCalled++
			return nil
		},
		NetworkCalled: func() network.Network {
			callCounters.NetworkCalled++
			return nil
		},
		MuxCalled: func() protocol.Switch {
			callCounters.MuxCalled++
			return nil
		},
		ConnectCalled: func(context.Context, peer.AddrInfo) error {
			callCounters.ConnectCalled++
			return nil
		},
		SetStreamHandlerCalled: func(protocol.ID, network.StreamHandler) {
			callCounters.SetStreamHandlerCalled++
		},
		SetStreamHandlerMatchCalled: func(protocol.ID, func(string) bool, network.StreamHandler) {
			callCounters.SetStreamHandlerMatchCalled++
		},
		RemoveStreamHandlerCalled: func(protocol.ID) {
			callCounters.RemoveStreamHandlerCalled++
		},
		NewStreamCalled: func(context.Context, peer.ID, ...protocol.ID) (network.Stream, error) {
			callCounters.NewStreamCalled++
			return nil, nil
		},
		CloseCalled: func() error {
			callCounters.CloseCalled++
			return nil
		},
		ConnManagerCalled: func() connmgr.ConnManager {
			callCounters.ConnManagerCalled++
			return nil
		},
	}

	d, err := NewHostDecorator(hst, ctx, hardCPSLimitLow, 0)
	assert.NotNil(t, d)
	assert.Nil(t, err)

	tcs := []struct {
		counter *int
		call    func()
		name    string
	}{
		{counter: &callCounters.EventBusCalled, call: func() { d.EventBus() }, name: "EventBus"},
		{counter: &callCounters.IDCalled, call: func() { d.ID() }, name: "ID"},
		{counter: &callCounters.PeerstoreCalled, call: func() { d.Peerstore() }, name: "Peerstore"},
		{counter: &callCounters.AddrsCalled, call: func() { d.Addrs() }, name: "Addrs"},
		{counter: &callCounters.NetworkCalled, call: func() { d.Network() }, name: "Network"},
		{counter: &callCounters.MuxCalled, call: func() { d.Mux() }, name: "Mux"},
		{counter: &callCounters.ConnectCalled, call: func() { d.Connect(ctx, peer.AddrInfo{}) }, name: "Connect"},
		{counter: &callCounters.SetStreamHandlerCalled, call: func() { d.SetStreamHandler("", nil) }, name: "SetStreamHandler"},
		{counter: &callCounters.SetStreamHandlerMatchCalled,
			call: func() { d.SetStreamHandlerMatch("", func(string) bool { return true }, nil) },
			name: "SetStreamHandlerMatch"},
		{counter: &callCounters.RemoveStreamHandlerCalled,
			call: func() { d.RemoveStreamHandler("") },
			name: "RemoveStreamHandler"},
		{counter: &callCounters.NewStreamCalled, call: func() { d.NewStream(ctx, "") }, name: "NewStream"},
		{counter: &callCounters.CloseCalled, call: func() { d.Close() }, name: "Close"},
		{counter: &callCounters.ConnManagerCalled, call: func() { d.ConnManager() }, name: "ConnManager"},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, *tc.counter, 0)
			tc.call()
			assert.Equal(t, *tc.counter, 1)
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
		go dec.Connect(ctx, peer.AddrInfo{})
	}
	time.Sleep(time.Second)
	dur := time.Since(start)

	mtx.Lock()
	avgConns := float64(connCount) / (float64(dur) / float64(time.Second))
	mtx.Unlock()

	assert.True(t, avgConns < float64(testCPSVal))
	go dec.Connect(ctx, peer.AddrInfo{})
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
