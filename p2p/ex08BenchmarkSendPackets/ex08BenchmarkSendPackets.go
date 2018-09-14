package ex08BenchmarkSendPackets

import (
	"context"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/config"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/service"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var records []int

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

	if len(splt) != 4 {
		fmt.Println("Malformed message received!")
		return
	}

	tstamp1, _ := strconv.Atoi(splt[0])
	tstamp2, _ := strconv.Atoi(splt[1])
	senderMsg := splt[2]
	num := len(m.Payload)

	duration := time.Now().Sub(time.Unix(int64(tstamp1), int64(tstamp2)))

	records = append(records, int(duration.Nanoseconds()/1000/1000))

	fmt.Printf("Got from %v a payload of %v bytes in %v\n", senderMsg, num, duration)
}

func Main() {
	chanStop := make(chan os.Signal, 1)
	signal.Notify(chanStop, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer func() {
			os.Exit(1)
		}()

		<-chanStop

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

		fmt.Printf("Min: %v ms, max: %v ms, avg: %v ms, recvs: %v\n", min, max, avg, len(records))
	}()

	records = []int{}

	filename, _ := filepath.Abs("configP2P.json")

	fmt.Printf("Node started...reading '%s' config file\n", filename)

	var config config.JsonConfig

	err := config.ReadFromFile(filename)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Config data read as: %v\n", config)

	node, err := p2p.CreateNewNode(context.Background(), config.Port, config.Peers, service.GetMarshalizerService())
	if err != nil {
		panic(err)
	}
	node.OnMsgRecv = recv

	//only seeder sends
	for len(config.Peers) == 0 {
		time.Sleep(time.Second * 2)

		data := PrepareBuff(node.P2pNode.ID().Pretty(), config.Size)

		node.BroadcastString(data, []string{})

		fmt.Printf("Sent %v bytes...\n", len(data))
	}

	fmt.Println("Waiting...")

	select {}
}
