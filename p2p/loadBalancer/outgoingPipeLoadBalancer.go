package loadBalancer

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

const defaultSendPipe = "default send pipe"

// OutgoingPipeLoadBalancer is a component that evenly balances requests to be sent
type OutgoingPipeLoadBalancer struct {
	mut   sync.RWMutex
	chans []chan *p2p.SendableData
	names []string
	//namesChans is defined only for performance purposes as to fast search by name
	//iteration is done directly on slices as that is used very often and is about 50x
	//faster then an iteration over a map
	namesChans map[string]chan *p2p.SendableData
}

// NewOutgoingPipeLoadBalancer creates a new instance of a SendDataThrottle instance
func NewOutgoingPipeLoadBalancer() *OutgoingPipeLoadBalancer {
	sdt := &OutgoingPipeLoadBalancer{
		chans:      make([]chan *p2p.SendableData, 0),
		names:      make([]string, 0),
		namesChans: make(map[string]chan *p2p.SendableData),
	}

	sdt.appendPipe(defaultSendPipe)

	return sdt
}

func (oplb *OutgoingPipeLoadBalancer) appendPipe(pipe string) {
	oplb.names = append(oplb.names, pipe)
	ch := make(chan *p2p.SendableData)
	oplb.chans = append(oplb.chans, ch)
	oplb.namesChans[pipe] = ch
}

// AddPipe adds a new pipe to the throttler
func (oplb *OutgoingPipeLoadBalancer) AddPipe(pipe string) error {
	oplb.mut.Lock()
	defer oplb.mut.Unlock()

	for _, name := range oplb.names {
		if name == pipe {
			return p2p.ErrPipeAlreadyExists
		}
	}

	oplb.appendPipe(pipe)

	return nil
}

// RemovePipe removes an existing pipe from the throttler
func (oplb *OutgoingPipeLoadBalancer) RemovePipe(pipe string) error {
	if pipe == defaultSendPipe {
		return p2p.ErrPipeCanNotBeDeleted
	}

	oplb.mut.Lock()
	defer oplb.mut.Unlock()

	index := -1

	for idx, name := range oplb.names {
		if name == pipe {
			index = idx
			break
		}
	}

	if index == -1 {
		return p2p.ErrPipeDoNotExists
	}

	//remove the index-th element in the chan slice
	copy(oplb.chans[index:], oplb.chans[index+1:])
	oplb.chans[len(oplb.chans)-1] = nil
	oplb.chans = oplb.chans[:len(oplb.chans)-1]

	//remove the index-th element in the names slice
	copy(oplb.names[index:], oplb.names[index+1:])
	oplb.names = oplb.names[:len(oplb.names)-1]

	delete(oplb.namesChans, pipe)

	return nil
}

// GetChannelOrDefault fetches the required pipe or the default if the pipe is not present
func (oplb *OutgoingPipeLoadBalancer) GetChannelOrDefault(pipe string) chan *p2p.SendableData {
	oplb.mut.RLock()
	defer oplb.mut.RUnlock()

	ch, _ := oplb.namesChans[pipe]
	if ch != nil {
		return ch
	}

	return oplb.chans[0]
}

// CollectFromPipes gets the waiting object found in the non buffered chans (it iterates through whole collection of
// defined pipes) and returns a list of those objects. If a chan do not have a waiting object to be fetch, that chan will
// be skipped. Method always returns a valid slice object and in the case that no chan has an object to be sent,
// the slice will be empty
func (oplb *OutgoingPipeLoadBalancer) CollectFromPipes() []*p2p.SendableData {
	oplb.mut.RLock()
	defer oplb.mut.RUnlock()

	collectedData := make([]*p2p.SendableData, 0)

	for _, channel := range oplb.chans {
		select {
		case sendable := <-channel:
			collectedData = append(collectedData, sendable)
		default:
		}
	}

	return collectedData
}
