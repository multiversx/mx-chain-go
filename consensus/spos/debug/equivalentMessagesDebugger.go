package debug

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-go/consensus"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("debug/equivalentmessages")

type equivalentMessagesDebugger struct {
	shouldProcessDataFunc func() bool
}

// NewEquivalentMessagesDebugger returns a new instance of equivalentMessagesDebugger
func NewEquivalentMessagesDebugger() *equivalentMessagesDebugger {
	debugger := &equivalentMessagesDebugger{
		shouldProcessDataFunc: isLogTrace,
	}

	return debugger
}

// DisplayEquivalentMessagesStatistics prints all the equivalent messages
func (debugger *equivalentMessagesDebugger) DisplayEquivalentMessagesStatistics(getDataHandler func() map[string]*consensus.EquivalentMessageInfo) {
	if !debugger.shouldProcessDataFunc() {
		return
	}
	if getDataHandler == nil {
		return
	}

	dataMap := getDataHandler()
	log.Trace(fmt.Sprintf("Equivalent messages statistics for current round\n%s", dataToString(dataMap)))
}

func dataToString(data map[string]*consensus.EquivalentMessageInfo) string {
	header := []string{
		"Block header hash",
		"Equivalent messages received",
		"Validated",
		"Previous aggregated signature",
		"Previous Pubkeys Bitmap",
	}

	lines := make([]*display.LineData, 0, len(data))
	idx := 0
	for hash, info := range data {
		horizontalLineAfter := idx == len(data)
		line := []string{
			hash,
			fmt.Sprintf("%d", info.NumMessages),
			fmt.Sprintf("%t", info.Validated),
			string(info.PreviousAggregateSignature),
			string(info.PreviousPubkeysBitmap),
		}
		lines = append(lines, display.NewLineData(horizontalLineAfter, line))
		idx++
	}

	table, err := display.CreateTableString(header, lines)
	if err != nil {
		return "error creating p2p stats table: " + err.Error()
	}

	return table
}

func isLogTrace() bool {
	return log.GetLevel() == logger.LogTrace
}

// IsInterfaceNil returns true if there is no value under the interface
func (debugger *equivalentMessagesDebugger) IsInterfaceNil() bool {
	return debugger == nil
}
