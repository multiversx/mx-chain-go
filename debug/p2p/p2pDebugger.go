package p2p

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/display"
	logger "github.com/multiversx/mx-chain-logger-go"
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

	duplicateSize uint64
	duplicateNum  uint32

	ignoredSize uint64
	ignoredNum  uint32
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

	m.duplicateSize = uint64(float32(m.duplicateSize) / divideValue)
	m.duplicateNum = uint32(float32(m.duplicateNum) / divideValue)

	m.ignoredSize = uint64(float32(m.ignoredSize) / divideValue)
	m.ignoredNum = uint32(float32(m.ignoredNum) / divideValue)
}

func (m *metric) stringify() []string {
	return []string{
		m.topic,
		fmt.Sprintf("%d / %s/s", m.incomingNum, core.ConvertBytes(m.incomingSize)),
		fmt.Sprintf("%d / %s/s", m.incomingRejectedNum, core.ConvertBytes(m.incomingRejectedSize)),
		fmt.Sprintf("%d / %s/s", m.duplicateNum, core.ConvertBytes(m.duplicateSize)),
		fmt.Sprintf("%d / %s/s", m.ignoredNum, core.ConvertBytes(m.ignoredSize)),
		fmt.Sprintf("%d / %s/s", m.outgoingNum, core.ConvertBytes(m.outgoingSize)),
		fmt.Sprintf("%d / %s/s", m.outgoingRejectedNum, core.ConvertBytes(m.outgoingRejectedSize)),
	}
}

type rpcMetric struct {
	topic string

	publishedInSize  uint64
	publishedInNum   uint32
	publishedOutSize uint64
	publishedOutNum  uint32

	controlInSize  uint64
	controlInNum   uint32
	controlOutSize uint64
	controlOutNum  uint32
}

func (m *rpcMetric) divideValues(divideValue float32) {
	m.publishedInSize = uint64(float32(m.publishedInSize) / divideValue)
	m.publishedInNum = uint32(float32(m.publishedInNum) / divideValue)
	m.publishedOutSize = uint64(float32(m.publishedOutSize) / divideValue)
	m.publishedOutNum = uint32(float32(m.publishedOutNum) / divideValue)

	m.controlInSize = uint64(float32(m.controlInSize) / divideValue)
	m.controlInNum = uint32(float32(m.controlInNum) / divideValue)
	m.controlOutSize = uint64(float32(m.controlOutSize) / divideValue)
	m.controlOutNum = uint32(float32(m.controlOutNum) / divideValue)
}

func (m *rpcMetric) stringify() []string {
	return []string{
		m.topic,
		fmt.Sprintf("%d / %s/s", m.publishedInNum, core.ConvertBytes(m.publishedInSize)),
		fmt.Sprintf("%d / %s/s", m.publishedOutNum, core.ConvertBytes(m.publishedOutSize)),
		fmt.Sprintf("%d / %s/s", m.controlInNum, core.ConvertBytes(m.controlInSize)),
		fmt.Sprintf("%d / %s/s", m.controlOutNum, core.ConvertBytes(m.controlOutSize)),
	}
}

type p2pDebugger struct {
	selfPeerId          core.PeerID
	mut                 sync.Mutex
	data                map[string]*metric
	rpcData             map[string]*rpcMetric
	cancelFunc          func()
	shouldProcessDataFn func() bool
	printStringFn       func(data string)
}

// NewP2PDebugger creates a new p2p debug instance
func NewP2PDebugger(selfPeerId core.PeerID) *p2pDebugger {
	pd := &p2pDebugger{
		selfPeerId: selfPeerId,
		data:       make(map[string]*metric),
		rpcData:    make(map[string]*rpcMetric),
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

// AddDuplicateMessage adds a new duplicated (already seen) message stats in metrics structs
func (pd *p2pDebugger) AddDuplicateMessage(topic string, size uint64) {
	if !pd.shouldProcessDataFn() {
		return
	}

	pd.mut.Lock()
	defer pd.mut.Unlock()

	m := pd.getMetric(topic)
	m.duplicateNum++
	m.duplicateSize += size
}

// AddIgnoredMessage adds a new message that was received but ignored by validation, so not propagated further
func (pd *p2pDebugger) AddIgnoredMessage(topic string, size uint64) {
	if !pd.shouldProcessDataFn() {
		return
	}

	pd.mut.Lock()
	defer pd.mut.Unlock()

	m := pd.getMetric(topic)
	m.ignoredNum++
	m.ignoredSize += size
}

// IsRecording returns true if the statistics are gathered, it gates the accounting done on the RPC hot paths
func (pd *p2pDebugger) IsRecording() bool {
	return pd.shouldProcessDataFn()
}

// AddRPCPublishedMessage adds a message carried by an RPC, as it travels on the wire
func (pd *p2pDebugger) AddRPCPublishedMessage(topic string, size uint64, isIncoming bool) {
	if !pd.shouldProcessDataFn() {
		return
	}

	pd.mut.Lock()
	defer pd.mut.Unlock()

	m := pd.getRPCMetric(topic)
	if isIncoming {
		m.publishedInNum++
		m.publishedInSize += size
		return
	}

	m.publishedOutNum++
	m.publishedOutSize += size
}

// AddRPCControlMessage adds a gossip control message carried by an RPC
func (pd *p2pDebugger) AddRPCControlMessage(topic string, size uint64, isIncoming bool) {
	if !pd.shouldProcessDataFn() {
		return
	}

	pd.mut.Lock()
	defer pd.mut.Unlock()

	m := pd.getRPCMetric(topic)
	if isIncoming {
		m.controlInNum++
		m.controlInSize += size
		return
	}

	m.controlOutNum++
	m.controlOutSize += size
}

func (pd *p2pDebugger) getRPCMetric(topic string) *rpcMetric {
	m, ok := pd.rpcData[topic]
	if !ok {
		m = &rpcMetric{
			topic: topic,
		}
		pd.rpcData[topic] = m
	}

	return m
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
		"Incoming duplicates (num / size)",
		"Incoming ignored (num / size)",
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
		total.duplicateSize += m.duplicateSize
		total.duplicateNum += m.duplicateNum
		total.ignoredSize += m.ignoredSize
		total.ignoredNum += m.ignoredNum
	}

	sort.Slice(metrics, func(i, j int) bool {
		// sort descending by the bytes that crossed the wire, then alphabetically. The duplicates are added as they
		// never reach the topic validator, while the ignored ones are already counted in incoming.
		mi := metrics[i]
		mj := metrics[j]

		miSize := mi.outgoingSize + mi.incomingSize + mi.duplicateSize
		mjSize := mj.outgoingSize + mj.incomingSize + mj.duplicateSize

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

	return tab + pd.rpcStatsToString(divideSeconds)
}

// must be called under the mut lock, held by statsToString
func (pd *p2pDebugger) rpcStatsToString(divideSeconds float32) string {
	header := []string{
		"Topic",
		"RPC messages in (num / size)",
		"RPC messages out (num / size)",
		"RPC control in (num / size)",
		"RPC control out (num / size)",
	}

	metrics := make([]*rpcMetric, 0, len(pd.rpcData))
	total := &rpcMetric{
		topic: "TOTAL",
	}
	for _, m := range pd.rpcData {
		m.divideValues(divideSeconds)
		metrics = append(metrics, m)

		total.publishedInSize += m.publishedInSize
		total.publishedInNum += m.publishedInNum
		total.publishedOutSize += m.publishedOutSize
		total.publishedOutNum += m.publishedOutNum
		total.controlInSize += m.controlInSize
		total.controlInNum += m.controlInNum
		total.controlOutSize += m.controlOutSize
		total.controlOutNum += m.controlOutNum
	}

	sort.Slice(metrics, func(i, j int) bool {
		mi := metrics[i]
		mj := metrics[j]

		miSize := mi.publishedInSize + mi.publishedOutSize + mi.controlInSize + mi.controlOutSize
		mjSize := mj.publishedInSize + mj.publishedOutSize + mj.controlInSize + mj.controlOutSize

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

	pd.rpcData = make(map[string]*rpcMetric)

	tab, err := display.CreateTableString(header, lines)
	if err != nil {
		return "error creating p2p RPC stats table: " + err.Error()
	}

	return "\n" + tab
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
