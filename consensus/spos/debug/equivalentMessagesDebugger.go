package debug

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/sharding"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("debug/equivalentmessages")

type proofsPoolHandler interface {
	GetAllNotarizedProofs(shardID uint32) (map[string]data.HeaderProofHandler, error)
	GetNotarizedProof(shardID uint32, headerHash []byte) (data.HeaderProofHandler, error)
	IsInterfaceNil() bool
}

type equivalentMessagesDebugger struct {
	proofsPool       proofsPoolHandler
	shardCoordinator sharding.Coordinator

	shouldProcessDataFunc func() bool

	mutEquivalentMessages sync.RWMutex
	msgCounters           map[string]uint64
}

// NewEquivalentMessagesDebugger returns a new instance of equivalentMessagesDebugger
func NewEquivalentMessagesDebugger(proofsPool proofsPoolHandler, shardCoordinator sharding.Coordinator) (*equivalentMessagesDebugger, error) {
	if check.IfNil(proofsPool) {
		return nil, spos.ErrNilProofPool
	}
	if check.IfNil(shardCoordinator) {
		return nil, spos.ErrNilShardCoordinator
	}

	return &equivalentMessagesDebugger{
		proofsPool:            proofsPool,
		shardCoordinator:      shardCoordinator,
		shouldProcessDataFunc: isLogTrace,
		msgCounters:           make(map[string]uint64),
	}, nil
}

func (debugger *equivalentMessagesDebugger) ResetEquivalentMessages() {
	debugger.msgCounters = make(map[string]uint64)
}

func (debugger *equivalentMessagesDebugger) UpsertEquivalentMessage(
	headerHash []byte,
) {
	debugger.mutEquivalentMessages.Lock()
	defer debugger.mutEquivalentMessages.Unlock()

	_, ok := debugger.msgCounters[string(headerHash)]
	if !ok {
		debugger.msgCounters[string(headerHash)] = 0
	}
	debugger.msgCounters[string(headerHash)]++
}

func (debugger *equivalentMessagesDebugger) DeleteEquivalentMessage(headerHash []byte) {
	debugger.mutEquivalentMessages.Lock()
	defer debugger.mutEquivalentMessages.Unlock()

	delete(debugger.msgCounters, string(headerHash))
}

// DisplayEquivalentMessagesStatistics prints all the equivalent messages
func (debugger *equivalentMessagesDebugger) DisplayEquivalentMessagesStatistics() {
	if !debugger.shouldProcessDataFunc() {
		return
	}

	dataAsStr := debugger.dataToString()

	log.Trace(fmt.Sprintf("Equivalent messages statistics for current round\n%s", dataAsStr))
}

func (debugger *equivalentMessagesDebugger) dataToString() string {
	debugger.mutEquivalentMessages.RLock()
	defer debugger.mutEquivalentMessages.RUnlock()

	header := []string{
		"Block header hash",
		"Equivalent messages received",
		"Aggregated signature",
		"Pubkeys Bitmap",
	}

	lines := make([]*display.LineData, 0, len(debugger.msgCounters))
	idx := 0
	for hash, numMessages := range debugger.msgCounters {
		sig, bitmap := make([]byte, 0), make([]byte, 0)
		proof, err := debugger.proofsPool.GetNotarizedProof(debugger.shardCoordinator.SelfId(), []byte(hash))
		if err == nil {
			sig, bitmap = proof.GetAggregatedSignature(), proof.GetPubKeysBitmap()
		}

		horizontalLineAfter := idx == len(debugger.msgCounters)
		line := []string{
			hash,
			fmt.Sprintf("%d", numMessages),
			string(sig),
			string(bitmap),
		}
		lines = append(lines, display.NewLineData(horizontalLineAfter, line))
		idx++
	}

	table, err := display.CreateTableString(header, lines)
	if err != nil {
		return "error creating equivalent proofs stats table: " + err.Error()
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
