package antiflood

import (
	"context"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

const sizeUint32 = 4
const sizeUint64 = 8
const sizeBool = 1
const newLineChar = "\n"
const minIntervalInSeconds = 1
const maxSequencesToPrint = 5
const moreSequencesPresent = "..."

var log = logger.GetOrCreate("debug/antiflood")

type event struct {
	pid           core.PeerID
	sequences     map[uint64]struct{}
	topic         string
	sizeRejected  uint64
	numRejected   uint32
	isBlackListed bool
}

// Size returns the size of an event instance
func (ev *event) Size() int {
	return len(ev.pid) + len(ev.topic) + sizeUint32 + sizeUint64 + sizeBool
}

func (ev *event) String() string {
	sequences := make([]string, 0, len(ev.sequences))
	for seq := range ev.sequences {
		sequences = append(sequences, fmt.Sprintf("%d", seq))
	}

	sort.Slice(sequences, func(i, j int) bool {
		return sequences[i] < sequences[j]
	})

	if len(sequences) > maxSequencesToPrint {
		sequences = append([]string{moreSequencesPresent}, sequences[len(sequences)-maxSequencesToPrint:]...)
	}

	return fmt.Sprintf("pid: %s; topic: %s; num rejected: %d; size rejected: %d; sequences: %s; is blacklisted: %v",
		ev.pid.Pretty(), ev.topic, ev.numRejected, ev.sizeRejected, strings.Join(sequences, ", "), ev.isBlackListed)
}

type debugger struct {
	mut               sync.RWMutex
	cache             storage.Cacher
	intervalAutoPrint time.Duration
	printEventFunc    func(data string)
	cancelFunc        func()
}

// NewAntifloodDebugger creates a new antiflood debugger able to hold antiflood events
func NewAntifloodDebugger(config config.AntifloodDebugConfig) (*debugger, error) {
	cache, err := lrucache.NewCache(config.CacheSize)
	if err != nil {
		return nil, fmt.Errorf("%w when creating NewAntifloodDebugger", err)
	}
	if config.IntervalAutoPrintInSeconds < minIntervalInSeconds {
		return nil, fmt.Errorf("%w for IntervalAutoPrintInSeconds, minimum is %d", debug.ErrInvalidValue, minIntervalInSeconds)
	}

	d := &debugger{
		cache:             cache,
		intervalAutoPrint: time.Second * time.Duration(config.IntervalAutoPrintInSeconds),
	}

	d.printEventFunc = d.printEvent
	ctx, cancelFunc := context.WithCancel(context.Background())
	d.cancelFunc = cancelFunc
	go d.printContinuously(ctx)

	return d, nil
}

// AddData adds the data in cache
func (d *debugger) AddData(
	pid core.PeerID,
	topic string,
	numRejected uint32,
	sizeRejected uint64,
	sequence []byte,
	isBlacklisted bool,
) {
	identifier := d.computeIdentifier(pid, topic)

	seqVal := uint64(0)
	if len(sequence) >= 8 {
		seqVal = binary.BigEndian.Uint64(sequence)
	}

	d.mut.Lock()
	defer d.mut.Unlock()

	obj, ok := d.cache.Get(identifier)
	if !ok {
		obj = &event{
			pid:           pid,
			topic:         topic,
			numRejected:   0,
			sizeRejected:  0,
			isBlackListed: false,
			sequences:     map[uint64]struct{}{seqVal: {}},
		}
	}

	ev, ok := obj.(*event)
	if !ok {
		ev = &event{
			pid:           pid,
			topic:         topic,
			numRejected:   0,
			sizeRejected:  0,
			isBlackListed: false,
			sequences:     map[uint64]struct{}{seqVal: {}},
		}
	}

	ev.numRejected += numRejected
	ev.sizeRejected += sizeRejected
	ev.isBlackListed = isBlacklisted
	ev.sequences[seqVal] = struct{}{}

	d.cache.Put(identifier, ev, ev.Size())
}

func (d *debugger) computeIdentifier(pid core.PeerID, topic string) []byte {
	return []byte(string(pid) + topic)
}

func (d *debugger) printContinuously(ctx context.Context) {
	for {
		select {
		case <-time.After(d.intervalAutoPrint):
		case <-ctx.Done():
			log.Debug("antiflood debugger printContinuously go routine is stopping...")
			return
		}

		events := []string{"antiflood events:"}

		d.mut.Lock()
		events = append(events, d.getStringEvents()...)
		d.cache.Clear()
		d.mut.Unlock()

		if len(events) == 1 {
			continue
		}

		stringEvent := strings.Join(events, newLineChar)
		d.printEventFunc(stringEvent)
	}
}

func (d *debugger) getStringEvents() []string {
	keys := d.cache.Keys()
	strs := make([]string, 0, len(keys))
	for _, key := range keys {
		element, ok := d.cache.Get(key)
		if !ok {
			continue
		}

		ev, ok := element.(*event)
		if !ok {
			continue
		}

		strs = append(strs, ev.String())
	}

	return strs
}

func (d *debugger) printEvent(data string) {
	log.Trace(data)
}

// Close will clean up this instance
func (d *debugger) Close() error {
	d.cancelFunc()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *debugger) IsInterfaceNil() bool {
	return d == nil
}
