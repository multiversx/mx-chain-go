package facade

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/api"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"

	beevikntp "github.com/beevik/ntp"
)

type ElrondFacade struct {
	node     *node.Node
	syncTime *ntp.SyncTime
	log *logger.Logger
}

func (ef *ElrondFacade) SetLogger(log *logger.Logger) {
	ef.log = log
}

func (ef *ElrondFacade) CreateNode(maxAllowedPeers, port int, initialNodeAddresses []string) {
	appContext := context.Background()
	hasher := sha256.Sha256{}
	marshalizer := marshal.JsonMarshalizer{}
	ef.node = node.NewNode(
		node.WithHasher(hasher),
		node.WithContext(appContext),
		node.WithMarshalizer(marshalizer),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithMaxAllowedPeers(maxAllowedPeers),
		node.WithPort(port),
		node.WithInitialNodeAddresses(initialNodeAddresses),
	)
}

func (ef *ElrondFacade) StartNode() error {
	err := ef.node.Start()
	if err != nil {
		return err
	}
	err = ef.node.ConnectToInitialAddresses()
	if err != nil {
		return err
	}
	err = ef.node.StartConsensus()
	return err
}

func (ef *ElrondFacade) StopNode() error {
	return ef.StopNode()
}

func (ef *ElrondFacade) StartNTP(clockSyncPeriod int) {
	ef.syncTime = ntp.NewSyncTime(time.Second*time.Duration(clockSyncPeriod), func(host string) (response *beevikntp.Response, e error) {
		return nil, errors.New("this should be implemented")
	})
}

func (ef *ElrondFacade) WaitForStartTime(t time.Time) {
	if !ef.syncTime.CurrentTime(ef.syncTime.ClockOffset()).After(t) {
		diff := t.Sub(ef.syncTime.CurrentTime(ef.syncTime.ClockOffset())).Seconds()
		ef.log.Info("Elrond protocol not started yet, waiting " + strconv.Itoa(int(diff)) + " seconds")
	}
	for {
		if ef.syncTime.CurrentTime(ef.syncTime.ClockOffset()).After(t) {
			break
		}
		time.Sleep(time.Duration(5 * time.Millisecond))
	}
}

func (ef *ElrondFacade) StartBackgroundServices(wg *sync.WaitGroup) {
	wg.Add(1)
	go ef.startRest(wg)
}

func (ef *ElrondFacade) IsNodeRunning() bool {
	return ef.node.IsRunning()
}

func (ef *ElrondFacade) startRest(wg *sync.WaitGroup) {
	ef.log.Info("Starting web server...")
	err := api.Start(ef)
	if err != nil {
		ef.log.Error("Could not start webserver", err.Error())
	}
	wg.Done()
}