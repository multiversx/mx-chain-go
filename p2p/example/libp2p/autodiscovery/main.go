package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/dataThrottle"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	cr "github.com/libp2p/go-libp2p-crypto"
)

var r *rand.Rand

// LineData represents a displayable table line
type lineData struct {
	Values              []string
	HorizontalRuleAfter bool
}

//The purpose of this example program is to show what happens if a peer connects to a network of 100 peers
//and has its connection manager set that will trim its connections to a number of 10
func main() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
	startingPort := 32000

	advertiser, _ := libp2p.NewSocketLibp2pMessenger(
		context.Background(),
		startingPort,
		genPrivKey(),
		nil,
		dataThrottle.NewSendDataThrottle(),
	)
	startingPort++
	fmt.Printf("advertiser is %s\n", getConnectableAddress(advertiser))
	peers := make([]p2p.Messenger, 0)
	go func() {
		_ = advertiser.DiscoverNewPeers()
		time.Sleep(time.Second)
	}()

	for i := 0; i < 99; i++ {
		netPeer, _ := libp2p.NewSocketLibp2pMessenger(
			context.Background(),
			startingPort,
			genPrivKey(),
			nil,
			dataThrottle.NewSendDataThrottle(),
		)
		startingPort++

		fmt.Printf("%s connecting to %s...\n",
			getConnectableAddress(netPeer),
			getConnectableAddress(advertiser))

		_ = netPeer.ConnectToPeer(getConnectableAddress(advertiser))
		_ = netPeer.DiscoverNewPeers()

		peers = append(peers, netPeer)

		go func() {
			_ = netPeer.DiscoverNewPeers()
			time.Sleep(time.Second)
		}()
	}

	//display func
	go func() {
		for {
			time.Sleep(time.Second)
			showConnections(advertiser, peers)
		}
	}()

	time.Sleep(time.Second * 15)

	_ = advertiser.Close()
	for _, peer := range peers {
		if peer == nil {
			continue
		}

		_ = peer.Close()
	}
}

func getConnectableAddress(peer p2p.Messenger) string {
	for _, adr := range peer.Addresses() {
		if strings.Contains(adr, "127.0.0.1") {
			return adr
		}
	}

	return ""
}

func genPrivKey() cr.PrivKey {
	prv, _, _ := cr.GenerateKeyPairWithReader(cr.Ed25519, 0, r)
	return prv
}

func showConnections(advertiser p2p.Messenger, peers []p2p.Messenger) {
	header := []string{"Node", "Address", "No. of conns"}

	lines := make([]*lineData, 0)
	lines = append(lines, createDataLine(advertiser, advertiser, peers))

	for i := 0; i < len(peers); i++ {
		lines = append(lines, createDataLine(peers[i], advertiser, peers))
	}

	table, _ := createTableString(header, lines)

	fmt.Println(table)
}

func createDataLine(peer p2p.Messenger, advertiser p2p.Messenger, peers []p2p.Messenger) *lineData {
	ld := &lineData{}

	if peer == nil {
		ld.Values = []string{"<NIL>", "<NIL>", "0"}
		return ld
	}

	nodeName := "Peer"
	if advertiser == peer {
		nodeName = "Advertiser"
	}

	ld.Values = []string{nodeName,
		getConnectableAddress(peer),
		strconv.Itoa(computeConnectionsCount(peer, advertiser, peers))}

	return ld
}

func computeConnectionsCount(peer p2p.Messenger, advertiser p2p.Messenger, peers []p2p.Messenger) int {
	if peer == nil {
		return 0
	}

	knownPeers := 0
	if peer.IsConnected(advertiser.ID()) {
		knownPeers++
	}

	for i := 0; i < len(peers); i++ {
		p := peers[i]

		if p == nil {
			continue
		}

		if peer.IsConnected(peers[i].ID()) {
			knownPeers++
		}
	}

	return knownPeers
}

// CreateTableString creates an ASCII table having header as table header and a LineData slice as table rows
// It automatically resize itself based on the lengths of every cell
func createTableString(header []string, data []*lineData) (string, error) {
	err := checkValidity(header, data)

	if err != nil {
		return "", err
	}

	columnsWidths := computeColumnsWidths(header, data)

	builder := &strings.Builder{}
	drawHorizontalRule(builder, columnsWidths)

	drawLine(builder, columnsWidths, header)
	drawHorizontalRule(builder, columnsWidths)

	lastLineHadHR := false

	for i := 0; i < len(data); i++ {
		drawLine(builder, columnsWidths, data[i].Values)
		lastLineHadHR = data[i].HorizontalRuleAfter

		if data[i].HorizontalRuleAfter {
			drawHorizontalRule(builder, columnsWidths)
		}
	}

	if !lastLineHadHR {
		drawHorizontalRule(builder, columnsWidths)
	}

	return builder.String(), nil
}

func checkValidity(header []string, data []*lineData) error {
	if header == nil {
		return errors.New("nil header")
	}

	if data == nil {
		return errors.New("nil data lines")
	}

	if len(data) == 0 && len(header) == 0 {
		return errors.New("empty slices")
	}

	maxElemFound := len(header)

	for _, ld := range data {
		if ld == nil {
			return errors.New("nil linedata in slice")
		}

		if ld.Values == nil {
			return errors.New("nil values of linedata in slice")
		}

		if maxElemFound < len(ld.Values) {
			maxElemFound = len(ld.Values)
		}
	}

	if maxElemFound == 0 {
		return errors.New("empty slices")
	}

	return nil
}

func computeColumnsWidths(header []string, data []*lineData) []int {
	widths := make([]int, len(header))

	for i := 0; i < len(header); i++ {
		widths[i] = len(header[i])
	}

	for i := 0; i < len(data); i++ {
		line := data[i]

		for j := 0; j < len(line.Values); j++ {
			if len(widths) <= j {
				widths = append(widths, len(line.Values[j]))
				continue
			}

			if widths[j] < len(line.Values[j]) {
				widths[j] = len(line.Values[j])
			}
		}
	}

	return widths
}

func drawHorizontalRule(builder *strings.Builder, columnsWidths []int) {
	builder.WriteByte('+')
	for i := 0; i < len(columnsWidths); i++ {
		for j := 0; j < columnsWidths[i]+2; j++ {
			builder.WriteByte('-')
		}
		builder.WriteByte('+')
	}

	builder.Write([]byte{'\r', '\n'})
}

func drawLine(builder *strings.Builder, columnsWidths []int, strings []string) {
	builder.WriteByte('|')

	for i := 0; i < len(columnsWidths); i++ {
		builder.WriteByte(' ')

		lenStr := 0

		if i < len(strings) {
			lenStr = len(strings[i])
			builder.WriteString(strings[i])
		}

		for j := lenStr; j < columnsWidths[i]; j++ {
			builder.WriteByte(' ')
		}

		builder.Write([]byte{' ', '|'})
	}

	builder.Write([]byte{'\r', '\n'})
}
