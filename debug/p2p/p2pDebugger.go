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

const printInterval = time.Second

type metric struct {
	topic string

	incomingSize         uint64
	incomingRejectedSize uint64
	incomingNum          uint32
	incomingRejectedNum  uint32

	outgoingSize         uint64
	outgoingRejectedSize uint64
	outgoingNum          uint32
	outgoingRejectedNum  uint32
}

func (m *metric) divideValues(divideValue float32) {
	m.incomingSize = uint64(float32(m.incomingSize) / divideValue)
	m.incomingNum = uint32(float32(m.incomingNum) / divideValue)
	m.incomingRejectedSize = uint64(float32(m.incomingRejectedSize) / divideValue)
	m.incomingRejectedNum = uint32(float32(m.incomingRejectedNum) / divideValue)

	m.outgoingSize = uint64(float32(m.outgoingSize) / divideValue)
	m.outgoingNum = uint32(float32(m.outgoingNum) / divideValue)
	m.outgoingRejectedSize = uint64(float32(m.outgoingRejectedSize) / divideValue)
	m.outgoingRejectedNum = uint32(float32(m.outgoingRejectedNum) / divideValue)
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
	pd.shouldProcessDataFn = pd.isLogTrace
	pd.printStringFn = pd.printLog

	ctx, cancelFunc := context.WithCancel(context.Background())
	pd.cancelFunc = cancelFunc

	go pd.continuouslyPrintStatistics(ctx)

	return pd
}

func (pd *p2pDebugger) isLogTrace() bool {
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

func (pd *p2pDebugger) continuouslyPrintStatistics(ctx context.Context) {
	divideSeconds := float32(printInterval) / float32(time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Debug("p2p debugger continuouslyPrintStatistics go routine is stopping...")
			return
		case <-time.After(printInterval):
		}

		if !pd.shouldProcessDataFn() {
			continue
		}

		str := pd.statsToString(divideSeconds)
		pd.printStringFn(str)
	}
}

func (pd *p2pDebugger) statsToString(divideSeconds float32) string {
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
		m.divideValues(divideSeconds)
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
