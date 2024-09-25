package debug

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/sharding"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("debug/equivalentmessages")

type equivalentMessagesDebugger struct {
	proofsPool       consensus.EquivalentProofsPool
	shardCoordinator sharding.Coordinator

	shouldProcessDataFunc func() bool

	mutEquivalentMessages sync.RWMutex
	msgCounters           map[string]uint64
}

// NewEquivalentMessagesDebugger returns a new instance of equivalentMessagesDebugger
func NewEquivalentMessagesDebugger(proofsPool consensus.EquivalentProofsPool, shardCoordinator sharding.Coordinator) (*equivalentMessagesDebugger, error) {
	if check.IfNil(proofsPool) {
		return nil, spos.ErrNilEquivalentProofPool
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

// ResetEquivalentMessages will reset messages counters
func (debugger *equivalentMessagesDebugger) ResetEquivalentMessages() {
	debugger.mutEquivalentMessages.Lock()
	defer debugger.mutEquivalentMessages.Unlock()

	debugger.msgCounters = make(map[string]uint64)
}

// UpsertEquivalentMessage will insert or update messages counter for provided header hash
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

// DeleteEquivalentMessage will delete equivalent message counter
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
		var sig, bitmap []byte
		proof, err := debugger.proofsPool.GetProof(debugger.shardCoordinator.SelfId(), []byte(hash))
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
