package ex01ChatLibP2P

import (
	"context"
	"fmt"
	"testing"
	"time"

	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"

	host "github.com/libp2p/go-libp2p-host"
	swarmt "github.com/libp2p/go-libp2p-swarm/testing"

	pstore "github.com/libp2p/go-libp2p-peerstore"

	mdns "github.com/libp2p/go-libp2p/p2p/discovery"
)

type DiscoveryNotifee struct {
	h host.Host
}

func (n *DiscoveryNotifee) HandlePeerFound(pi pstore.PeerInfo) {
	n.h.Connect(context.Background(), pi)
	fmt.Printf("connected by %v\n", pi.ID.Pretty())
}

func TestMdnsDiscovery(t *testing.T) {
	//TODO: re-enable when the new lib will get integrated
	//t.Skip("TestMdnsDiscovery fails randomly with current lib")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a := bhost.New(swarmt.GenSwarm(t, ctx))
	b := bhost.New(swarmt.GenSwarm(t, ctx))

	fmt.Printf("generated a: %v\n", a.ID().Pretty())
	fmt.Printf("generated b: %v\n", b.ID().Pretty())

	sa, err := mdns.NewMdnsService(ctx, a, time.Second, "someTag")
	if err != nil {
		t.Fatal(err)
	}

	sb, err := mdns.NewMdnsService(ctx, b, time.Second, "someTag")
	if err != nil {
		t.Fatal(err)
	}

	_ = sb

	n := &DiscoveryNotifee{a}
	m := &DiscoveryNotifee{b}

	sa.RegisterNotifee(n)
	sa.RegisterNotifee(m)

	time.Sleep(time.Second * 2)

	err = a.Connect(ctx, pstore.PeerInfo{ID: b.ID()})
	if err != nil {
		t.Fatal(err)
	}
}
