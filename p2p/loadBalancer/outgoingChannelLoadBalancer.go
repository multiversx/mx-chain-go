package loadBalancer

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

const defaultSendChannel = "default send channel"

// OutgoingChannelLoadBalancer is a component that evenly balances requests to be sent
type OutgoingChannelLoadBalancer struct {
	mut      sync.RWMutex
	chans    []chan *p2p.SendableData
	mainChan chan *p2p.SendableData
	names    []string
	//namesChans is defined only for performance purposes as to fast search by name
	//iteration is done directly on slices as that is used very often and is about 50x
	//faster then an iteration over a map
	namesChans map[string]chan *p2p.SendableData
}

// NewOutgoingChannelLoadBalancer creates a new instance of a ChannelLoadBalancer instance
func NewOutgoingChannelLoadBalancer() *OutgoingChannelLoadBalancer {
	oclb := &OutgoingChannelLoadBalancer{
		chans:      make([]chan *p2p.SendableData, 0),
		names:      make([]string, 0),
		namesChans: make(map[string]chan *p2p.SendableData),
		mainChan:   make(chan *p2p.SendableData),
	}

	oclb.appendChannel(defaultSendChannel)

	return oclb
}

func (oplb *OutgoingChannelLoadBalancer) appendChannel(channel string) {
	oplb.names = append(oplb.names, channel)
	ch := make(chan *p2p.SendableData)
	oplb.chans = append(oplb.chans, ch)
	oplb.namesChans[channel] = ch

	go func() {
		for {
			obj, ok := <-ch

			if !ok {
				//channel closed, close the go routine
				return
			}

			oplb.mainChan <- obj
		}
	}()
}

// AddChannel adds a new channel to the throttler
func (oplb *OutgoingChannelLoadBalancer) AddChannel(channel string) error {
	oplb.mut.Lock()
	defer oplb.mut.Unlock()

	for _, name := range oplb.names {
		if name == channel {
			return p2p.ErrChannelAlreadyExists
		}
	}

	oplb.appendChannel(channel)

	return nil
}

// RemoveChannel removes an existing channel from the throttler
func (oplb *OutgoingChannelLoadBalancer) RemoveChannel(channel string) error {
	if channel == defaultSendChannel {
		return p2p.ErrChannelCanNotBeDeleted
	}

	oplb.mut.Lock()
	defer oplb.mut.Unlock()

	index := -1

	for idx, name := range oplb.names {
		if name == channel {
			index = idx
			break
		}
	}

	if index == -1 {
		return p2p.ErrChannelDoesNotExist
	}

	sendableChan := oplb.chans[index]

	//remove the index-th element in the chan slice
	copy(oplb.chans[index:], oplb.chans[index+1:])
	oplb.chans[len(oplb.chans)-1] = nil
	oplb.chans = oplb.chans[:len(oplb.chans)-1]

	//remove the index-th element in the names slice
	copy(oplb.names[index:], oplb.names[index+1:])
	oplb.names = oplb.names[:len(oplb.names)-1]

	close(sendableChan)

	delete(oplb.namesChans, channel)

	return nil
}

// GetChannelOrDefault fetches the required channel or the default if the channel is not present
func (oplb *OutgoingChannelLoadBalancer) GetChannelOrDefault(channel string) chan *p2p.SendableData {
	oplb.mut.RLock()
	defer oplb.mut.RUnlock()

	ch, _ := oplb.namesChans[channel]
	if ch != nil {
		return ch
	}

	return oplb.chans[0]
}

// CollectOneElementFromChannels gets the waiting object from mainChan. It is a blocking call.
func (oplb *OutgoingChannelLoadBalancer) CollectOneElementFromChannels() *p2p.SendableData {
	obj := <-oplb.mainChan
	return obj
}
