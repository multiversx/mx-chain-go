package ex09BenchmarkDirectSendMultiple

import (
	"context"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/service"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-libp2p-peerstore"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var records []int

var stats map[string][]int
var statsRcv []string
var mut sync.Mutex

func PrepareBuff(sender string, nums int) string {
	if nums < 100 {
		nums = 100
	}

	tme := time.Now()

	unix := tme.UnixNano()

	buff := make([]rune, nums)

	run := []rune{'A'}

	for i := range buff {
		buff[i] = run[0]
	}

	return fmt.Sprintf("%d|%d|%s|%s", unix/time.Second.Nanoseconds(),
		unix%time.Second.Nanoseconds(), sender, string(buff))
}

func recv(sender *p2p.Node, peerID string, m *p2p.Message) {
	splt := strings.Split(string(m.Payload), "|")

	if splt[0] == "STAT" {

		fmt.Printf("Got stat msg: %v\n", string(m.Payload))

		if len(splt) != 5 {
			fmt.Printf("Malformed stats message received %v\n!", string(m.Payload))
			return
		}

		vals := []int{}

		i := 2
		for i < 5 {
			val, err := strconv.Atoi(splt[i])

			if err != nil {
				fmt.Printf("Malformed stats message received %v [%v]\n!", m.Payload, err)
				return
			}

			vals = append(vals, val)
			i++
		}

		mut.Lock()
		defer mut.Unlock()

		stats[splt[1]] = vals
		statsRcv = append(statsRcv, splt[1])
		return

	}

	if len(splt) != 4 {
		fmt.Printf("Malformed message received %v!\n", m.Payload)
		return
	}

	tstamp1, _ := strconv.Atoi(splt[0])
	tstamp2, _ := strconv.Atoi(splt[1])
	//senderMsg := splt[2]
	//num := len(message)

	duration := time.Now().Sub(time.Unix(int64(tstamp1), int64(tstamp2)))

	records = append(records, int(duration.Nanoseconds()/1000/1000))
	//fmt.Printf("Got from %v a payload of %v bytes in %v\n", senderMsg, num, duration)

}

func shutdownHandler(node *p2p.Node, peerID string, chanStop chan os.Signal) {
	defer func() {
		os.Exit(1)
	}()

	<-chanStop

	if len(statsRcv) > 0 {
		min := 2147483647
		max := 0
		avg := 0

		mut.Lock()
		defer mut.Unlock()

		cnt := 0

		for _, pid := range statsRcv {

			vals, ok := stats[pid]

			if !ok {
				panic("Internal error, not found peerID:" + pid)
			}

			if vals[0] < min {
				min = vals[0]
			}

			if vals[1] > max {
				max = vals[1]
			}

			avg = avg + vals[2]

			fmt.Printf("Stat: %v| min: %5d | max: %5d | avg: %5d\n", pid, vals[0], vals[1], vals[2])

			cnt++
		}

		fmt.Printf("TOTAL: %v peers >>> min: %5d | max: %5d | avg: %5d\n", cnt, min, max, avg/len(statsRcv))

		time.Sleep(time.Second)

		return
	}

	if len(records) == 0 {
		fmt.Println("No statistics available!")
		return
	}

	min := 2147483647
	max := 0
	var avg int64 = 0

	for _, val := range records {
		if min > val {
			min = val
		}

		if max < val {
			max = val
		}

		avg += int64(val)
	}

	avg = avg / int64(len(records))

	//fmt.Printf("Min: %v ms, max: %v ms, avg: %v ms, recvs: %v\n", min, max, avg, len(records))
	str := fmt.Sprintf("STAT|%v|%v|%v|%v", node.P2pNode.ID().Pretty(), min, max, avg)
	fmt.Println(str)

	if peerID != "" {
		addr, err := ipfsaddr.ParseString(peerID)
		if err != nil {
			panic(err)
		}
		pinfo, _ := peerstore.InfoFromP2pAddr(addr.Multiaddr())

		node.SendDirectString(pinfo.ID.Pretty(), str)

		time.Sleep(time.Second * 20)
	}

}

func Main() {

	durationWork := time.Second * 60
	durationWaitConnect := time.Second * 30
	durationWaitSend := time.Second * 5

	chanStop := make(chan os.Signal, 1)
	signal.Notify(chanStop, os.Interrupt, syscall.SIGTERM)

	records = []int{}
	stats = make(map[string][]int)
	statsRcv = []string{}

	mut = sync.Mutex{}

	filename, _ := filepath.Abs("configP2P.json")

	fmt.Printf("Node started...reading '%s' config file\n", filename)

	argsWithoutProg := os.Args[1:]

	fmt.Printf("Config data read as: %v\n", argsWithoutProg)

	if len(argsWithoutProg) < 2 {
		panic("Should have been minimum 2 parameters: max_packet_size and port")
	}

	maxPakSize, err := strconv.Atoi(argsWithoutProg[0])
	if err != nil {
		panic("Max packet size incorrectly defined as" + argsWithoutProg[1])
	}
	if maxPakSize < 100 {
		maxPakSize = 100
	}

	port, err := strconv.Atoi(argsWithoutProg[1])
	if err != nil {
		panic("Port incorrectly defined as" + argsWithoutProg[1])
	}

	peers := []string{}

	i := 2

	masterPeerID := ""

	for i < len(argsWithoutProg) {
		masterPeerID = argsWithoutProg[2]

		peers = append(peers, argsWithoutProg[i])

		i++
	}

	node, err := p2p.CreateNewNode(context.Background(), port, peers, service.GetMarshalizerService())
	if err != nil {
		panic(err)
	}
	node.OnMsgRecv = recv

	go shutdownHandler(node, masterPeerID, chanStop)

	if len(peers) == 0 {
		fmt.Printf("Waiting %v for other peers to connect...\n", durationWaitConnect)
		time.Sleep(durationWaitConnect)
		fmt.Printf("Got a peerstore of %v peers (including self)\n", len(node.P2pNode.Peerstore().Peers()))
	}

	//only seeder sends
	for len(peers) == 0 {
		time.Sleep(durationWaitSend)

		data := PrepareBuff(node.P2pNode.ID().Pretty(), maxPakSize)

		node.BroadcastString(data, []string{})

		fmt.Printf("Sent %v bytes...\n", len(data))
	}

	fmt.Printf("Waiting %v...\n", durationWork)

	time.Sleep(durationWork)

	chanStop <- os.Interrupt

	select {}
}
