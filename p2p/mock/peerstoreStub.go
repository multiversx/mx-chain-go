package mock

import (
	"context"
	"time"

	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

// PeerstoreStub -
type PeerstoreStub struct {
	CloseCalled                  func() error
	AddAddrCalled                func(p peer.ID, addr multiaddr.Multiaddr, ttl time.Duration)
	AddAddrsCalled               func(p peer.ID, addrs []multiaddr.Multiaddr, ttl time.Duration)
	SetAddrCalled                func(p peer.ID, addr multiaddr.Multiaddr, ttl time.Duration)
	SetAddrsCalled               func(p peer.ID, addrs []multiaddr.Multiaddr, ttl time.Duration)
	UpdateAddrsCalled            func(p peer.ID, oldTTL time.Duration, newTTL time.Duration)
	AddrsCalled                  func(p peer.ID) []multiaddr.Multiaddr
	AddrStreamCalled             func(ctx context.Context, id peer.ID) <-chan multiaddr.Multiaddr
	ClearAddrsCalled             func(p peer.ID)
	PeersWithAddrsCalled         func() peer.IDSlice
	PubKeyCalled                 func(id peer.ID) libp2pCrypto.PubKey
	AddPubKeyCalled              func(id peer.ID, key libp2pCrypto.PubKey) error
	PrivKeyCalled                func(id peer.ID) libp2pCrypto.PrivKey
	AddPrivKeyCalled             func(id peer.ID, key libp2pCrypto.PrivKey) error
	PeersWithKeysCalled          func() peer.IDSlice
	GetCalled                    func(p peer.ID, key string) (interface{}, error)
	PutCalled                    func(p peer.ID, key string, val interface{}) error
	RecordLatencyCalled          func(id peer.ID, duration time.Duration)
	LatencyEWMACalled            func(id peer.ID) time.Duration
	GetProtocolsCalled           func(id peer.ID) ([]string, error)
	AddProtocolsCalled           func(id peer.ID, s ...string) error
	SetProtocolsCalled           func(id peer.ID, s ...string) error
	RemoveProtocolsCalled        func(id peer.ID, s ...string) error
	SupportsProtocolsCalled      func(id peer.ID, s ...string) ([]string, error)
	FirstSupportedProtocolCalled func(id peer.ID, s ...string) (string, error)
	PeerInfoCalled               func(id peer.ID) peer.AddrInfo
	PeersCalled                  func() peer.IDSlice
	RemovePeerCalled             func(id peer.ID)
}

// Close -
func (ps *PeerstoreStub) Close() error {
	if ps.CloseCalled != nil {
		return ps.CloseCalled()
	}

	return nil
}

// AddAddr -
func (ps *PeerstoreStub) AddAddr(p peer.ID, addr multiaddr.Multiaddr, ttl time.Duration) {
	if ps.AddAddrCalled != nil {
		ps.AddAddrCalled(p, addr, ttl)
	}
}

// AddAddrs -
func (ps *PeerstoreStub) AddAddrs(p peer.ID, addrs []multiaddr.Multiaddr, ttl time.Duration) {
	if ps.AddAddrsCalled != nil {
		ps.AddAddrsCalled(p, addrs, ttl)
	}
}

// SetAddr -
func (ps *PeerstoreStub) SetAddr(p peer.ID, addr multiaddr.Multiaddr, ttl time.Duration) {
	if ps.SetAddrCalled != nil {
		ps.SetAddrCalled(p, addr, ttl)
	}
}

// SetAddrs -
func (ps *PeerstoreStub) SetAddrs(p peer.ID, addrs []multiaddr.Multiaddr, ttl time.Duration) {
	if ps.SetAddrsCalled != nil {
		ps.SetAddrsCalled(p, addrs, ttl)
	}
}

// UpdateAddrs -
func (ps *PeerstoreStub) UpdateAddrs(p peer.ID, oldTTL time.Duration, newTTL time.Duration) {
	if ps.UpdateAddrsCalled != nil {
		ps.UpdateAddrsCalled(p, oldTTL, newTTL)
	}
}

// Addrs -
func (ps *PeerstoreStub) Addrs(p peer.ID) []multiaddr.Multiaddr {
	if ps.AddrsCalled != nil {
		return ps.AddrsCalled(p)
	}

	return nil
}

// AddrStream -
func (ps *PeerstoreStub) AddrStream(ctx context.Context, id peer.ID) <-chan multiaddr.Multiaddr {
	if ps.AddrStreamCalled != nil {
		return ps.AddrStreamCalled(ctx, id)
	}

	return nil
}

// ClearAddrs -
func (ps *PeerstoreStub) ClearAddrs(p peer.ID) {
	if ps.ClearAddrsCalled != nil {
		ps.ClearAddrsCalled(p)
	}
}

// PeersWithAddrs -
func (ps *PeerstoreStub) PeersWithAddrs() peer.IDSlice {
	if ps.PeersWithAddrsCalled != nil {
		return ps.PeersWithAddrsCalled()
	}

	return nil
}

// PubKey -
func (ps *PeerstoreStub) PubKey(id peer.ID) libp2pCrypto.PubKey {
	if ps.PubKeyCalled != nil {
		return ps.PubKeyCalled(id)
	}

	return nil
}

// AddPubKey -
func (ps *PeerstoreStub) AddPubKey(id peer.ID, key libp2pCrypto.PubKey) error {
	if ps.AddPubKeyCalled != nil {
		return ps.AddPubKeyCalled(id, key)
	}

	return nil
}

// PrivKey -
func (ps *PeerstoreStub) PrivKey(id peer.ID) libp2pCrypto.PrivKey {
	if ps.PrivKeyCalled != nil {
		return ps.PrivKeyCalled(id)
	}

	return nil
}

// AddPrivKey -
func (ps *PeerstoreStub) AddPrivKey(id peer.ID, key libp2pCrypto.PrivKey) error {
	if ps.AddPrivKeyCalled != nil {
		return ps.AddPrivKeyCalled(id, key)
	}

	return nil
}

// PeersWithKeys -
func (ps *PeerstoreStub) PeersWithKeys() peer.IDSlice {
	if ps.PeersWithKeysCalled != nil {
		return ps.PeersWithKeysCalled()
	}

	return nil
}

// Get -
func (ps *PeerstoreStub) Get(p peer.ID, key string) (interface{}, error) {
	if ps.GetCalled != nil {
		return ps.GetCalled(p, key)
	}

	return nil, nil
}

// Put -
func (ps *PeerstoreStub) Put(p peer.ID, key string, val interface{}) error {
	if ps.PutCalled != nil {
		return ps.PutCalled(p, key, val)
	}

	return nil
}

// RecordLatency -
func (ps *PeerstoreStub) RecordLatency(id peer.ID, duration time.Duration) {
	if ps.RecordLatencyCalled != nil {
		ps.RecordLatencyCalled(id, duration)
	}
}

// LatencyEWMA -
func (ps *PeerstoreStub) LatencyEWMA(id peer.ID) time.Duration {
	if ps.LatencyEWMACalled != nil {
		return ps.LatencyEWMACalled(id)
	}

	return 0
}

// GetProtocols -
func (ps *PeerstoreStub) GetProtocols(id peer.ID) ([]string, error) {
	if ps.GetProtocolsCalled != nil {
		return ps.GetProtocolsCalled(id)
	}

	return nil, nil
}

// AddProtocols -
func (ps *PeerstoreStub) AddProtocols(id peer.ID, s ...string) error {
	if ps.AddProtocolsCalled != nil {
		return ps.AddProtocolsCalled(id, s...)
	}

	return nil
}

// SetProtocols -
func (ps *PeerstoreStub) SetProtocols(id peer.ID, s ...string) error {
	if ps.SetProtocolsCalled != nil {
		return ps.SetProtocolsCalled(id, s...)
	}

	return nil
}

// RemoveProtocols -
func (ps *PeerstoreStub) RemoveProtocols(id peer.ID, s ...string) error {
	if ps.RemoveProtocolsCalled != nil {
		return ps.RemoveProtocolsCalled(id, s...)
	}

	return nil
}

// SupportsProtocols -
func (ps *PeerstoreStub) SupportsProtocols(id peer.ID, s ...string) ([]string, error) {
	if ps.SupportsProtocolsCalled != nil {
		return ps.SupportsProtocolsCalled(id, s...)
	}

	return nil, nil
}

// FirstSupportedProtocol -
func (ps *PeerstoreStub) FirstSupportedProtocol(id peer.ID, s ...string) (string, error) {
	if ps.FirstSupportedProtocolCalled != nil {
		return ps.FirstSupportedProtocolCalled(id, s...)
	}

	return "", nil
}

// PeerInfo -
func (ps *PeerstoreStub) PeerInfo(id peer.ID) peer.AddrInfo {
	if ps.PeerInfoCalled != nil {
		return ps.PeerInfoCalled(id)
	}

	return peer.AddrInfo{}
}

// Peers -
func (ps *PeerstoreStub) Peers() peer.IDSlice {
	if ps.PeersCalled != nil {
		return ps.PeersCalled()
	}

	return nil
}

// RemovePeer -
func (ps *PeerstoreStub) RemovePeer(id peer.ID) {
	if ps.RemovePeerCalled != nil {
		ps.RemovePeerCalled(id)
	}
}
