package debug

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/display"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("debug/equivalentmessages")

// EquivalentMessageInfo holds information about an equivalent message
type equivalentMessageDebugInfo struct {
	NumMessages uint64
	Validated   bool
	Proof       data.HeaderProof
}

type equivalentMessagesDebugger struct {
	shouldProcessDataFunc func() bool

	mutEquivalentMessages sync.RWMutex
	equivalentMessages    map[string]*equivalentMessageDebugInfo
}

// NewEquivalentMessagesDebugger returns a new instance of equivalentMessagesDebugger
func NewEquivalentMessagesDebugger() *equivalentMessagesDebugger {
	debugger := &equivalentMessagesDebugger{
		shouldProcessDataFunc: isLogTrace,
		equivalentMessages:    make(map[string]*equivalentMessageDebugInfo),
	}

	return debugger
}

func (debugger *equivalentMessagesDebugger) resetEquivalentMessages() {
	debugger.mutEquivalentMessages.Lock()
	defer debugger.mutEquivalentMessages.Unlock()

	debugger.equivalentMessages = make(map[string]*equivalentMessageDebugInfo)
}

func (debugger *equivalentMessagesDebugger) SetValidEquivalentProof(
	headerHash []byte,
	proof data.HeaderProof,
) {
	debugger.mutEquivalentMessages.Lock()
	defer debugger.mutEquivalentMessages.Unlock()

	equivalentMessage, ok := debugger.equivalentMessages[string(headerHash)]
	if !ok {
		debugger.equivalentMessages[string(headerHash)] = &equivalentMessageDebugInfo{
			NumMessages: 1,
		}
	}
	equivalentMessage.Validated = true
	equivalentMessage.Proof = proof
}

func (debugger *equivalentMessagesDebugger) UpsertEquivalentMessage(
	headerHash []byte,
) {
	debugger.mutEquivalentMessages.Lock()
	defer debugger.mutEquivalentMessages.Unlock()

	equivalentMessage, ok := debugger.equivalentMessages[string(headerHash)]
	if !ok {
		debugger.equivalentMessages[string(headerHash)] = &equivalentMessageDebugInfo{
			NumMessages: 0,
			Validated:   false,
		}
	}
	equivalentMessage.NumMessages++
}

// DisplayEquivalentMessagesStatistics prints all the equivalent messages
func (debugger *equivalentMessagesDebugger) DisplayEquivalentMessagesStatistics() {
	if !debugger.shouldProcessDataFunc() {
		return
	}

	debugger.mutEquivalentMessages.Lock()
	defer debugger.mutEquivalentMessages.Unlock()

	dataMap := debugger.equivalentMessages

	log.Trace(fmt.Sprintf("Equivalent messages statistics for current round\n%s", dataToString(dataMap)))

	debugger.resetEquivalentMessages()
}

func dataToString(data map[string]*equivalentMessageDebugInfo) string {
	header := []string{
		"Block header hash",
		"Equivalent messages received",
		"Validated",
		"Aggregated signature",
		"Pubkeys Bitmap",
	}

	lines := make([]*display.LineData, 0, len(data))
	idx := 0
	for hash, info := range data {
		horizontalLineAfter := idx == len(data)
		line := []string{
			hash,
			fmt.Sprintf("%d", info.NumMessages),
			fmt.Sprintf("%t", info.Validated),
			string(info.Proof.AggregatedSignature),
			string(info.Proof.PubKeysBitmap),
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
