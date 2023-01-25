package antiflooding

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process/throttle/antiflood/blackList"
	"github.com/multiversx/mx-chain-go/process/throttle/antiflood/factory"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

var log = logger.GetOrCreate("integrationtests/longtests/antiflood") //nolint

//nolint
func createWorkableConfig() config.Config {
	return config.Config{
		Antiflood: config.AntifloodConfig{
			Enabled: true,
			Cache: config.CacheConfig{
				Type:     "LRU",
				Capacity: 5000,
				Shards:   16,
			},
			FastReacting: config.FloodPreventerConfig{
				IntervalInSeconds: 1,
				ReservedPercent:   20,
				PeerMaxInput: config.AntifloodLimitsConfig{
					BaseMessagesPerInterval: 75,
					TotalSizePerInterval:    2097152,
				},
				BlackList: config.BlackListConfig{
					ThresholdNumMessagesPerInterval: 480,
					ThresholdSizePerInterval:        5242880,
					NumFloodingRounds:               10,
					PeerBanDurationInSeconds:        300,
				},
			},
			SlowReacting: config.FloodPreventerConfig{
				IntervalInSeconds: 30,
				ReservedPercent:   20,
				PeerMaxInput: config.AntifloodLimitsConfig{
					BaseMessagesPerInterval: 2500,
					TotalSizePerInterval:    15728640,
				},
				BlackList: config.BlackListConfig{
					ThresholdNumMessagesPerInterval: 6000,
					ThresholdSizePerInterval:        37748736,
					NumFloodingRounds:               2,
					PeerBanDurationInSeconds:        3600,
				},
			},
			OutOfSpecs: config.FloodPreventerConfig{
				IntervalInSeconds: 1,
				ReservedPercent:   0,
				PeerMaxInput: config.AntifloodLimitsConfig{
					BaseMessagesPerInterval: 1000,
					TotalSizePerInterval:    8388608,
				},
				BlackList: config.BlackListConfig{
					ThresholdNumMessagesPerInterval: 1500,
					ThresholdSizePerInterval:        10485760,
					NumFloodingRounds:               2,
					PeerBanDurationInSeconds:        3600,
				},
			},
			Topic: config.TopicAntifloodConfig{
				DefaultMaxMessagesPerSec: 10000,
			},
		},
	}
}

//nolint
func createDisabledConfig() config.Config {
	return config.Config{
		Antiflood: config.AntifloodConfig{
			Enabled: false,
		},
	}
}

func TestAntifloodingForLargerPeriodOfTime(t *testing.T) {
	t.Skip("this is a long and harsh test")

	peers, err := integrationTests.CreateFixedNetworkOf8Peers()
	assert.Nil(t, err)

	defer func() {
		integrationTests.ClosePeers(peers)
	}()

	//the network has 8 peers (check integrationTests.CreateFixedNetworkOf7Peers function)
	//nodes 2, 4, 6 decide to flood the network, but they will keep their messages/sec under the threshold
	topic := "test_topic"
	idxGoodPeers := []int{0, 1, 3, 5, 7}
	idxBadPeers := []int{2, 4, 6}

	processors := createProcessors(peers, topic, idxBadPeers, idxGoodPeers)

	go startFlooding(peers, topic, idxBadPeers, 1*1024*1024, 256*1024)

	for i := 0; i < 1000; i++ {
		displayProcessors(processors, idxBadPeers, i)
	}
}

//nolint
func createProcessors(peers []p2p.Messenger, topic string, idxBadPeers []int, idxGoodPeers []int) []*messageProcessor {
	processors := make([]*messageProcessor, 0, len(peers))
	ctx := context.Background()
	for i := 0; i < len(peers); i++ {
		var antifloodComponents *factory.AntiFloodComponents
		var err error

		if intInSlice(i, idxBadPeers) {
			antifloodComponents, err = factory.NewP2PAntiFloodComponents(ctx, createDisabledConfig(), &statusHandlerMock.AppStatusHandlerStub{}, peers[i].ID())
			log.LogIfError(err)
		}

		if intInSlice(i, idxGoodPeers) {
			statusHandler := &statusHandlerMock.AppStatusHandlerStub{}
			antifloodComponents, err = factory.NewP2PAntiFloodComponents(ctx, createWorkableConfig(), statusHandler, peers[i].ID())
			log.LogIfError(err)
		}

		pde, _ := blackList.NewPeerDenialEvaluator(
			antifloodComponents.BlacklistHandler,
			antifloodComponents.PubKeysCacher,
			&mock.PeerShardMapperStub{},
		)

		err = peers[i].SetPeerDenialEvaluator(pde)
		log.LogIfError(err)

		proc := NewMessageProcessor(antifloodComponents.AntiFloodHandler, peers[i])
		processors = append(processors, proc)

		err = proc.messenger.CreateTopic(topic, true)
		log.LogIfError(err)

		err = proc.messenger.RegisterMessageProcessor(topic, "test", proc)
		log.LogIfError(err)
	}

	return processors
}

//nolint
func intInSlice(searchFor int, slice []int) bool {
	for _, val := range slice {
		if searchFor == val {
			return true
		}
	}

	return false
}

//nolint
func displayProcessors(processors []*messageProcessor, idxBadPeers []int, idxRound int) {
	header := []string{"idx", "pid", "received", "processed", "received/s", "connections"}
	data := make([]*display.LineData, 0, len(processors))
	timeBetweenPrints := time.Second
	for idx, p := range processors {
		mark := ""
		if intInSlice(idx, idxBadPeers) {
			mark = " *"
		}

		val := []string{
			fmt.Sprintf("%d%s", idx, mark),
			p.Messenger().ID().Pretty(),
			fmt.Sprintf("%d / %s", p.NumMessagesReceived(), core.ConvertBytes(p.SizeMessagesReceived())),
			fmt.Sprintf("%d / %s", p.NumMessagesProcessed(), core.ConvertBytes(p.SizeMessagesProcessed())),
			fmt.Sprintf("%d/s / %s/s",
				p.NumMessagesReceivedPerInterval(timeBetweenPrints),
				core.ConvertBytes(p.SizeMessagesReceivedPerInterval(timeBetweenPrints)),
			),
			fmt.Sprintf("%d", len(p.Messenger().ConnectedPeers())),
		}

		line := display.NewLineData(false, val)
		data = append(data, line)
	}

	tbl, _ := display.CreateTableString(header, data)

	log.Info(fmt.Sprintf("Test round %d\n", idxRound) + tbl)
	time.Sleep(timeBetweenPrints)
}

//nolint
func startFlooding(peers []p2p.Messenger, topic string, idxBadPeers []int, maxSize int, msgSize int) {
	lastUpdated := time.Now()
	m := make(map[core.PeerID]int)

	for {
		for idx, p := range peers {
			time.Sleep(time.Millisecond)
			if !intInSlice(idx, idxBadPeers) {
				continue
			}

			if time.Since(lastUpdated) > time.Second {
				m = make(map[core.PeerID]int)
				//comment the following line to make the test generate a large number of messages/sec
				lastUpdated = time.Now()
			}

			size := m[p.ID()]
			if size >= maxSize {
				continue
			}

			m[p.ID()] += msgSize
			buff := make([]byte, msgSize)

			p.Broadcast(topic, buff)
		}
	}
}
