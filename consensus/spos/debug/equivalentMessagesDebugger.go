package debug

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/display"
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
func (debugger *equivalentMessagesDebugger) DisplayEquivalentMessagesStatistics(getDataHandler func() map[string]uint64) {
	if !debugger.shouldProcessDataFunc() {
		return
	}
	if getDataHandler == nil {
		return
	}

	dataMap := getDataHandler()
	log.Trace(fmt.Sprintf("Equivalent messages statistics for current round\n%s", dataToString(dataMap)))
}

func dataToString(data map[string]uint64) string {
	header := []string{
		"Block header hash",
		"Equivalent messages received",
	}

	lines := make([]*display.LineData, 0, len(data))
	idx := 0
	for hash, cnt := range data {
		horizontalLineAfter := idx == len(data)
		line := []string{
			hash,
			fmt.Sprintf("%d", cnt),
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
