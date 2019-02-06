package dataThrottle

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

const defaultSendPipe = "default send pipe"

// SendDataThrottle is a software device that evenly balances requests to be send
type SendDataThrottle struct {
	mut   sync.RWMutex
	chans []chan *p2p.SendableData
	names []string
	//namesChans is defined only for performance purposes as to fast search by name
	//iteration is done directly on slices as that is used very often and is about 50x
	//faster then an iteration over a map
	namesChans map[string]chan *p2p.SendableData
}

// NewSendDataThrottle creates a new instance of a SendDataThrottle instance
func NewSendDataThrottle() *SendDataThrottle {
	sdt := &SendDataThrottle{
		chans:      make([]chan *p2p.SendableData, 0),
		names:      make([]string, 0),
		namesChans: make(map[string]chan *p2p.SendableData),
	}

	sdt.appendPipe(defaultSendPipe)

	return sdt
}

func (sdt *SendDataThrottle) appendPipe(pipe string) {
	sdt.names = append(sdt.names, pipe)
	ch := make(chan *p2p.SendableData)
	sdt.chans = append(sdt.chans, ch)
	sdt.namesChans[pipe] = ch
}

// AddPipe adds a new pipe to the throttler
func (sdt *SendDataThrottle) AddPipe(pipe string) error {
	sdt.mut.Lock()
	defer sdt.mut.Unlock()

	for _, name := range sdt.names {
		if name == pipe {
			return p2p.ErrPipeAlreadyExists
		}
	}

	sdt.appendPipe(pipe)

	return nil
}

// RemovePipe removes an existing pipe from the throttler
func (sdt *SendDataThrottle) RemovePipe(pipe string) error {
	if pipe == defaultSendPipe {
		return p2p.ErrPipeCanNotBeDeleted
	}

	sdt.mut.Lock()
	defer sdt.mut.Unlock()

	index := -1

	for idx, name := range sdt.names {
		if name == pipe {
			index = idx
			break
		}
	}

	if index == -1 {
		return p2p.ErrPipeDoNotExists
	}

	//remove the index-th element in the chan slice
	copy(sdt.chans[index:], sdt.chans[index+1:])
	sdt.chans[len(sdt.chans)-1] = nil
	sdt.chans = sdt.chans[:len(sdt.chans)-1]

	//remove the index-th element in the names slice
	copy(sdt.names[index:], sdt.names[index+1:])
	sdt.names = sdt.names[:len(sdt.names)-1]

	delete(sdt.namesChans, pipe)

	return nil
}

// GetChannelOrDefault fetches the required pipe or the default if the pipe is not present
func (sdt *SendDataThrottle) GetChannelOrDefault(pipe string) chan *p2p.SendableData {
	sdt.mut.RLock()
	defer sdt.mut.RUnlock()

	ch, _ := sdt.namesChans[pipe]
	if ch != nil {
		return ch
	}

	return sdt.chans[0]
}

// CollectFromPipes gets the waiting object found in the non buffered chans (it iterates through whole collection of
// defined pipes) and returns a list of those objects. If a chan do not have a waiting object to be fetch, that chan will
// be skipped. Method always returns a valid slice object and in the case that no chan has an object to be sent,
// the slice will be empty
func (sdt *SendDataThrottle) CollectFromPipes() []*p2p.SendableData {
	sdt.mut.RLock()
	defer sdt.mut.RUnlock()

	collectedData := make([]*p2p.SendableData, 0)

	for _, channel := range sdt.chans {
		select {
		case sendable := <-channel:
			collectedData = append(collectedData, sendable)
		default:
		}
	}

	return collectedData
}
