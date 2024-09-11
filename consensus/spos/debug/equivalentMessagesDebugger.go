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
	Proof       data.HeaderProofHandler
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

func (debugger *equivalentMessagesDebugger) ResetEquivalentMessages() {
	debugger.equivalentMessages = make(map[string]*equivalentMessageDebugInfo)
}

func (debugger *equivalentMessagesDebugger) SetValidEquivalentProof(
	headerHash []byte,
	proof data.HeaderProofHandler,
) {
	debugger.mutEquivalentMessages.Lock()
	defer debugger.mutEquivalentMessages.Unlock()

	equivalentMessage, ok := debugger.equivalentMessages[string(headerHash)]
	if !ok {
		equivalentMessage = &equivalentMessageDebugInfo{
			NumMessages: 1,
		}
		debugger.equivalentMessages[string(headerHash)] = equivalentMessage
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
		equivalentMessage = &equivalentMessageDebugInfo{
			NumMessages: 0,
			Validated:   false,
		}
		debugger.equivalentMessages[string(headerHash)] = equivalentMessage
	}
	equivalentMessage.NumMessages++
}

func (debugger *equivalentMessagesDebugger) GetEquivalentMessages() map[string]*equivalentMessageDebugInfo {
	debugger.mutEquivalentMessages.Lock()
	defer debugger.mutEquivalentMessages.Unlock()

	return debugger.equivalentMessages
}

func (debugger *equivalentMessagesDebugger) DeleteEquivalentMessage(headerHash []byte) {
	debugger.mutEquivalentMessages.Lock()
	defer debugger.mutEquivalentMessages.Unlock()

	delete(debugger.equivalentMessages, string(headerHash))
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
			string(info.Proof.GetAggregatedSignature()),
			string(info.Proof.GetPubKeysBitmap()),
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
