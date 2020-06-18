package p2p

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/display"
)

var log = logger.GetOrCreate("debug/p2p")

const printTimeOneSecond = time.Second //if this needs to be changed, remember to divide the values when computing metrics

type metric struct {
	topic string

	incomingSize         uint64
	incomingNum          uint32
	incomingRejectedSize uint64
	incomingRejectedNum  uint32

	outgoingSize         uint64
	outgoingNum          uint32
	outgoingRejectedSize uint64
	outgoingRejectedNum  uint32
}

func (m *metric) stringify() []string {
	return []string{
		m.topic,
		fmt.Sprintf("%d / %s/s", m.incomingNum, core.ConvertBytes(m.incomingSize)),
		fmt.Sprintf("%d / %s/s", m.incomingRejectedNum, core.ConvertBytes(m.incomingRejectedSize)),
		fmt.Sprintf("%d / %s/s", m.outgoingNum, core.ConvertBytes(m.outgoingSize)),
		fmt.Sprintf("%d / %s/s", m.outgoingRejectedNum, core.ConvertBytes(m.outgoingRejectedSize)),
	}
}

type p2pDebugger struct {
	selfPeerId          core.PeerID
	mut                 sync.Mutex
	data                map[string]*metric
	cancelFunc          func()
	shouldProcessDataFn func() bool
	printStringFn       func(data string)
}

// NewP2PDebugger creates a new p2p debug instance
func NewP2PDebugger(selfPeerId core.PeerID) *p2pDebugger {
	pd := &p2pDebugger{
		selfPeerId: selfPeerId,
		data:       make(map[string]*metric),
	}
	pd.shouldProcessDataFn = pd.checkLogTrace
	pd.printStringFn = pd.printLog

	ctx, cancelFunc := context.WithCancel(context.Background())
	pd.cancelFunc = cancelFunc

	go pd.doStats(ctx)

	return pd
}

func (pd *p2pDebugger) checkLogTrace() bool {
	return log.GetLevel() == logger.LogTrace
}

func (pd *p2pDebugger) printLog(data string) {
	log.Trace(fmt.Sprintf("p2p topic stats for %s\n", pd.selfPeerId.Pretty()) + data)
}

// AddIncomingMessage adds a new incoming message stats in metrics structs
func (pd *p2pDebugger) AddIncomingMessage(topic string, size uint64, isRejected bool) {
	if !pd.shouldProcessDataFn() {
		return
	}

	pd.mut.Lock()
	defer pd.mut.Unlock()

	m := pd.getMetric(topic)
	m.incomingNum++
	m.incomingSize += size
	if isRejected {
		m.incomingRejectedNum++
		m.incomingRejectedSize += size
	}
}

// AddOutgoingMessage adds a new outgoing message stats in metrics structs
func (pd *p2pDebugger) AddOutgoingMessage(topic string, size uint64, isRejected bool) {
	if !pd.shouldProcessDataFn() {
		return
	}

	pd.mut.Lock()
	defer pd.mut.Unlock()

	m := pd.getMetric(topic)
	m.outgoingNum++
	m.outgoingSize += size
	if isRejected {
		m.outgoingRejectedNum++
		m.outgoingRejectedSize += size
	}
}

func (pd *p2pDebugger) getMetric(topic string) *metric {
	m, ok := pd.data[topic]
	if !ok {
		m = &metric{
			topic: topic,
		}
		pd.data[topic] = m
	}

	return m
}

func (pd *p2pDebugger) doStats(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(printTimeOneSecond):
		}

		if !pd.shouldProcessDataFn() {
			continue
		}

		str := pd.doStatsString()
		pd.printStringFn(str)
	}
}

func (pd *p2pDebugger) doStatsString() string {
	header := []string{
		"Topic",
		"Incoming (num / size)",
		"Incoming rejected (num / size)",
		"Outgoing (num / size)",
		"Outgoing rejected (num / size)",
	}

	pd.mut.Lock()
	defer pd.mut.Unlock()

	metrics := make([]*metric, 0, len(pd.data))
	total := &metric{
		topic: "TOTAL",
	}
	for _, m := range pd.data {
		metrics = append(metrics, m)

		total.incomingSize += m.incomingSize
		total.incomingNum += m.incomingNum
		total.incomingRejectedSize += m.incomingRejectedSize
		total.incomingRejectedNum += m.incomingRejectedNum
		total.outgoingSize += m.outgoingSize
		total.outgoingNum += m.outgoingNum
		total.outgoingRejectedSize += m.outgoingRejectedSize
		total.outgoingRejectedNum += m.outgoingRejectedNum
	}

	sort.Slice(metrics, func(i, j int) bool {
		//sort descending by incomingSize + outgoingSize and alphabetically
		mi := metrics[i]
		mj := metrics[j]

		miSize := mi.outgoingSize + mi.incomingSize
		mjSize := mj.outgoingSize + mj.incomingSize

		if miSize == mjSize {
			return mi.topic < mj.topic
		}

		return miSize > mjSize
	})

	lines := make([]*display.LineData, 0, len(metrics)+1)
	for idx, m := range metrics {
		horizontalLineAfter := idx == len(metrics)-1
		lines = append(lines, display.NewLineData(horizontalLineAfter, m.stringify()))
	}
	lines = append(lines, display.NewLineData(false, total.stringify()))

	pd.data = make(map[string]*metric)

	tab, err := display.CreateTableString(header, lines)
	if err != nil {
		return "error creating p2p stats table: " + err.Error()
	}

	return tab
}

// Close will stop any go routines launched by this instance
func (pd *p2pDebugger) Close() error {
	pd.cancelFunc()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pd *p2pDebugger) IsInterfaceNil() bool {
	return pd == nil
}
