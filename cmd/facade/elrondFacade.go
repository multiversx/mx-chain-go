package facade

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

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
}

func (ef *ElrondFacade) CreateNode(maxAllowedPeers, port int) {
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
	)
}

func (ef *ElrondFacade) StartNode(initialAddresses []string) error {
	err := ef.node.Start()
	if err != nil {
		return err
	}
	err = ef.node.ConnectToAddresses(initialAddresses)
	if err != nil {
		return err
	}
	err = ef.node.StartConsensus()
	return err
}

func (ef *ElrondFacade) StartNTP(clockSyncPeriod int) {
	ef.syncTime = ntp.NewSyncTime(time.Second*time.Duration(clockSyncPeriod), func(host string) (response *beevikntp.Response, e error) {
		return nil, errors.New("this should be implemented")
	})
}

func (ef *ElrondFacade) WaitForStartTime(t time.Time, log *logger.Logger) {
	if !ef.syncTime.CurrentTime(ef.syncTime.ClockOffset()).After(t) {
		diff := t.Sub(ef.syncTime.CurrentTime(ef.syncTime.ClockOffset())).Seconds()
		log.Info("Elrond protocol not started yet, waiting " + strconv.Itoa(int(diff)) + " seconds")
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

func (ef *ElrondFacade) startRest(wg *sync.WaitGroup) {
	// TODO: Next task will refactor the api. We are not able to boot it here right now,
	//  but it will be possible after the changes are implemented. We should do that then
	wg.Done()
}
