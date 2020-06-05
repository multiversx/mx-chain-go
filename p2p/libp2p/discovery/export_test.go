package discovery

import (
	"time"
)

const KadDhtName = kadDhtName
const NullName = nilName

//------- ContinuousKadDhtDiscoverer

func (ckdd *ContinuousKadDhtDiscoverer) ConnectToOnePeerFromInitialPeersList(
	durationBetweenAttempts time.Duration,
	initialPeersList []string) <-chan struct{} {

	return ckdd.connectToOnePeerFromInitialPeersList(durationBetweenAttempts, initialPeersList)
}

func (ckdd *ContinuousKadDhtDiscoverer) StopDHT() error {
	ckdd.mutKadDht.Lock()
	err := ckdd.stopDHT()
	ckdd.mutKadDht.Unlock()

	return err
}
