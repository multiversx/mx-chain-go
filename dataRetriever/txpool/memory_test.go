package txpool

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"runtime/pprof"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardedTxPool_MemoryFootprint_SourceIsMe_300x1x1048576(t *testing.T) {
	journal := analyzeMemoryFootprint(t, "SourceIsMe_300x1x1048576", 300, 1, core.MegabyteSize, "0")
	assert.LessOrEqual(t, bToMb(journal.txsFootprint()), 300)
	assert.LessOrEqual(t, bToMb(journal.poolStructuresFootprint()), 1)
}

func TestShardedTxPool_MemoryFootprint_SourceIsMe_10x1000x30720(t *testing.T) {
	journal := analyzeMemoryFootprint(t, "SourceIsMe_10x1000x30720", 10, 1000, 30720, "0")
	assert.LessOrEqual(t, bToMb(journal.txsFootprint()), 315)
	assert.LessOrEqual(t, bToMb(journal.poolStructuresFootprint()), 4)
}

func TestShardedTxPool_MemoryFootprint_SourceIsMe_10000x1x1024(t *testing.T) {
	journal := analyzeMemoryFootprint(t, "SourceIsMe_10000x1x1024", 10000, 1, 1024, "0")
	assert.LessOrEqual(t, bToMb(journal.txsFootprint()), 16)
	assert.LessOrEqual(t, bToMb(journal.poolStructuresFootprint()), 10)
}

func TestShardedTxPool_MemoryFootprint_SourceIsMe_1x60000x1024(t *testing.T) {
	journal := analyzeMemoryFootprint(t, "SourceIsMe_1x60000x1024", 1, 60000, 256, "0")
	assert.LessOrEqual(t, bToMb(journal.txsFootprint()), 32)
	assert.LessOrEqual(t, bToMb(journal.poolStructuresFootprint()), 15)
}

func TestShardedTxPool_MemoryFootprint_SourceIsMe_10x10000x100(t *testing.T) {
	journal := analyzeMemoryFootprint(t, "SourceIsMe_10x10000x100", 10, 10000, 100, "0")
	assert.LessOrEqual(t, bToMb(journal.txsFootprint()), 40)
	assert.LessOrEqual(t, bToMb(journal.poolStructuresFootprint()), 25)
}

func TestShardedTxPool_MemoryFootprint_SourceIsMe_100000x1x1024(t *testing.T) {
	journal := analyzeMemoryFootprint(t, "SourceIsMe_100000x1x1024", 100000, 1, 1024, "0")
	assert.LessOrEqual(t, bToMb(journal.txsFootprint()), 125)
	assert.LessOrEqual(t, bToMb(journal.poolStructuresFootprint()), 60)
}

// Many transactions per sender result in the largest memory footprint.
func TestShardedTxPool_MemoryFootprint_SourceIsMe_20x20000x100(t *testing.T) {
	journal := analyzeMemoryFootprint(t, "SourceIsMe_20x20000x100", 20, 20000, 100, "0")
	assert.LessOrEqual(t, bToMb(journal.txsFootprint()), 150)
	assert.LessOrEqual(t, bToMb(journal.poolStructuresFootprint()), 100)
}

func analyzeMemoryFootprint(t *testing.T, scenario string, numSenders int, numTxsPerSender int, payloadLengthPerTx int, cacheID string) *memoryFootprintJournal {
	journal := &memoryFootprintJournal{}

	journal.beforeGenerate = getMemStats()

	txs := generateTxs(numSenders, numTxsPerSender, payloadLengthPerTx)

	journal.afterGenerate = getMemStats()

	pool := newPool()

	pprofCPU(scenario, "addition", func() {
		for _, tx := range txs {
			pool.AddData(tx.hash, tx, tx.Size(), cacheID)
		}
	})

	require.Equal(t, numSenders*numTxsPerSender, len(pool.getTxCache(cacheID).Keys()))

	pprofHeap(scenario, "afterAddition")
	journal.afterAddition = getMemStats()

	journal.display()

	holdPoolInMemory(pool)
	return journal
}

func generateTxs(numSenders int, numTxsPerSender int, payloadLengthPerTx int) []*dummyTx {
	txs := make([]*dummyTx, 0, numSenders*numTxsPerSender)
	for senderTag := 0; senderTag < numSenders; senderTag++ {
		for nonce := 0; nonce < numTxsPerSender; nonce++ {
			tx := createTxWithPayload(senderTag, nonce, payloadLengthPerTx)
			txs = append(txs, tx)
		}
	}

	return txs
}

func getMemStats() runtime.MemStats {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats
}

func pprofHeap(scenario string, step string) {
	filename := path.Join(".", "pprofoutput", fmt.Sprintf("%s_%s.pprof", scenario, step))
	file, err := os.Create(filename)
	if err != nil {
		panic("could not create pprof file")
	}

	defer file.Close()
	runtime.GC()

	err = pprof.WriteHeapProfile(file)
	if err != nil {
		panic("could not write memory profile")
	}

	convertPprofToHumanReadable(filename)
}

func createTxWithPayload(senderTag int, nonce int, payloadLength int) *dummyTx {
	sender := createFakeSenderAddress(senderTag)
	hash := createFakeTxHash(sender, nonce)

	return &dummyTx{
		Transaction: transaction.Transaction{
			SndAddr: []byte(sender),
			Nonce:   uint64(nonce),
			Data:    make([]byte, payloadLength),
		},
		hash: []byte(hash),
	}
}

func createFakeSenderAddress(senderTag int) []byte {
	bytes := make([]byte, 32)
	binary.LittleEndian.PutUint32(bytes, uint32(senderTag))
	binary.LittleEndian.PutUint32(bytes[28:], uint32(senderTag))
	return bytes
}

func createFakeTxHash(fakeSenderAddress []byte, nonce int) []byte {
	bytes := make([]byte, 32)
	copy(bytes, fakeSenderAddress)
	binary.LittleEndian.PutUint32(bytes[4:], uint32(nonce))
	binary.LittleEndian.PutUint32(bytes[24:], uint32(nonce))
	return bytes
}

type dummyTx struct {
	transaction.Transaction
	hash []byte
}

func (dummyTx *dummyTx) GetGasPrice() uint64 {
	return 0
}

type memoryFootprintJournal struct {
	beforeGenerate runtime.MemStats
	afterGenerate  runtime.MemStats
	afterAddition  runtime.MemStats
}

func (journal *memoryFootprintJournal) txsFootprint() uint64 {
	return journal.afterGenerate.HeapInuse - journal.beforeGenerate.HeapInuse
}

func (journal *memoryFootprintJournal) poolStructuresFootprint() uint64 {
	return journal.afterAddition.HeapInuse - journal.afterGenerate.HeapInuse
}

func (journal *memoryFootprintJournal) display() {
	// See: https://golang.org/pkg/runtime/#MemStats

	fmt.Printf("beforeGenerate:")
	fmt.Printf("\tHeapAlloc = %v MiB", bToMb(journal.beforeGenerate.HeapAlloc))
	fmt.Printf("\tHeapInUse = %v MiB", bToMb(journal.beforeGenerate.HeapInuse))
	fmt.Println()

	fmt.Printf("afterGenerate:")
	fmt.Printf("\tHeapAlloc = %v MiB", bToMb(journal.afterGenerate.HeapAlloc))
	fmt.Printf("\tHeapInUse = %v MiB", bToMb(journal.afterGenerate.HeapInuse))
	fmt.Println()

	fmt.Printf("afterAddition:")
	fmt.Printf("\tHeapAlloc = %v MiB", bToMb(journal.afterAddition.HeapAlloc))
	fmt.Printf("\tHeapInUse = %v MiB", bToMb(journal.afterAddition.HeapInuse))
	fmt.Println()

	fmt.Println("Txs footprint:", bToMb(journal.txsFootprint()))
	fmt.Println("Pool structures footprint:", bToMb(journal.poolStructuresFootprint()))
}

func bToMb(b uint64) int {
	return int(b / 1024 / 1024)
}

func newPool() *shardedTxPool {
	config := storageUnit.CacheConfig{
		Capacity:             900000,
		SizePerSender:        60000,
		SizeInBytes:          500 * core.MegabyteSize,
		SizeInBytesPerSender: 32 * core.MegabyteSize,
		Shards:               16,
	}

	args := ArgShardedTxPool{Config: config, MinGasPrice: 200000000000, NumberOfShards: 2, SelfShardID: 0}
	poolAsInterface, err := NewShardedTxPool(args)
	if err != nil {
		panic("newMainnetPool")
	}

	pool := poolAsInterface.(*shardedTxPool)
	return pool
}

func convertPprofToHumanReadable(filename string) {
	cmd := exec.Command("go", "tool", "pprof", "-png", "-output", fmt.Sprintf("%s.png", filename), filename)
	err := cmd.Run()
	if err != nil {
		panic(fmt.Sprintf("convertPprofToHumanReadable: %v", err))
	}

	cmd = exec.Command("go", "tool", "pprof", "-text", "-output", fmt.Sprintf("%s.txt", filename), filename)
	err = cmd.Run()
	if err != nil {
		panic(fmt.Sprintf("convertPprofToHumanReadable: %v", err))
	}
}

func holdPoolInMemory(pool *shardedTxPool) {
	fmt.Println(pool.GetCounts().String())
}

func pprofCPU(scenario string, step string, function func()) {
	filename := path.Join(".", "pprofoutput", fmt.Sprintf("%s_%s.CPU.pprof", scenario, step))
	file, err := os.Create(filename)
	if err != nil {
		panic("could not create pprof file")
	}

	defer file.Close()

	err = pprof.StartCPUProfile(file)
	if err != nil {
		panic("could not start CPU profile")
	}

	function()

	pprof.StopCPUProfile()
	convertPprofToHumanReadable(filename)
}
