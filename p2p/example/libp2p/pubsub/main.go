package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
)

var hosts = make([]host.Host, 0)
var pubsubs = make([]*pubsub.PubSub, 0)
var startingPort = 10000
var topic = "test"

func main() {
	var err error

	err = createNodes(5)
	if err != nil {
		return
	}

	defer func() {
		for _, h := range hosts {
			_ = h.Close()
		}

		if err != nil {
			fmt.Printf("Error encountered: %v\n", err)
		}
	}()

	err = createPubsubs()
	if err != nil {
		return
	}

	err = topicRegistration()
	if err != nil {
		return
	}

	connectingNodes()
	//wait so that subscriptions on topic will be done and all peers will "know" of all other peers
	time.Sleep(time.Second * 2)

	fmt.Println("Broadcasting a message on node 0...")

	err = pubsubs[0].Publish(topic, []byte("a message"))
	if err != nil {
		fmt.Printf("Error encountered: %v\n", err)
		return
	}

	time.Sleep(time.Second)
}

func createNodes(nrNodes int) error {
	for i := 0; i < nrNodes; i++ {
		h, err := createHost(startingPort + i)
		if err != nil {
			return err
		}

		hosts = append(hosts, h)
		fmt.Printf("Node %v is %s\n", i, getLocalhostAddress(h))
	}

	return nil
}

func createPubsubs() error {
	for _, h := range hosts {
		ps, err := applyPubSub(h)
		if err != nil {
			return err
		}

		pubsubs = append(pubsubs, ps)
	}
	return nil
}

func topicRegistration() error {
	for i := 0; i < len(pubsubs); i++ {
		var subscr *pubsub.Subscription
		subscr, err := pubsubs[i].Subscribe(topic)
		if err != nil {
			return err
		}

		//just a dummy func to consume messages received by the newly created topic
		go func() {
			for {
				//here you will actually have the message received after passing all validators
				//not required since we put validators on each topic and the message has already been processed there
				_, _ = subscr.Next(context.Background())
			}
		}()

		crtHost := hosts[i]
		_ = pubsubs[i].RegisterTopicValidator(topic, func(ctx context.Context, pid peer.ID, msg *pubsub.Message) bool {
			//do the message validation
			//example: deserialize msg.Data, do checks on the message, etc.

			//processing part should be done on a go routine as the validator func should return immediately
			go func(data []byte, p peer.ID, h host.Host) {
				fmt.Printf("%s: Message: '%s' was received from %s\n", crtHost.ID(), msg.Data, pid.Pretty())
			}(msg.Data, pid, crtHost)

			//if the return value is true, the message will hit other peers
			//if the return value is false, this peer prevented message broadcasting
			//note that this topic validator will be called also for messages sent by self
			return true
		})
	}

	return nil
}

func connectingNodes() {
	//connect the nodes as following:
	//
	// node0 --------- node1
	//   |               |
	//   +------------ node2
	//   |               |
	//   |             node3
	//   |               |
	//   +------------ node4
	connectHostToPeer(hosts[1], getLocalhostAddress(hosts[0]))
	connectHostToPeer(hosts[2], getLocalhostAddress(hosts[1]))
	connectHostToPeer(hosts[2], getLocalhostAddress(hosts[0]))
	connectHostToPeer(hosts[3], getLocalhostAddress(hosts[2]))
	connectHostToPeer(hosts[4], getLocalhostAddress(hosts[3]))
	connectHostToPeer(hosts[4], getLocalhostAddress(hosts[0]))
}

func createHost(port int) (host.Host, error) {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	sk := (*libp2pCrypto.Secp256k1PrivateKey)(prvKey)

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)),
		libp2p.Identity(sk),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
	}

	h, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func applyPubSub(h host.Host) (*pubsub.PubSub, error) {
	optsPS := []pubsub.Option{
		pubsub.WithMessageSigning(true),
	}

	return pubsub.NewGossipSub(context.Background(), h, optsPS...)
}

func getLocalhostAddress(h host.Host) string {
	for _, addr := range h.Addrs() {
		if strings.Contains(addr.String(), "127.0.0.1") {
			return addr.String() + "/p2p/" + h.ID().Pretty()
		}
	}

	return ""
}

func connectHostToPeer(h host.Host, connectToAddress string) {
	multiAddr, err := multiaddr.NewMultiaddr(connectToAddress)
	if err != nil {
		fmt.Printf("Error encountered: %v\n", err)
		return
	}

	pInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)
	if err != nil {
		fmt.Printf("Error encountered: %v\n", err)
		return
	}

	err = h.Connect(context.Background(), *pInfo)
	if err != nil {
		fmt.Printf("Error encountered: %v\n", err)
	}
}
