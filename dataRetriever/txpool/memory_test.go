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
	pool := newPool()
	journal := analyzeMemoryFootprint(t, pool, "0_300x1x1048576", 300, 1, core.MegabyteSize, "0")
	assert.True(t, journal.txsFootprintIsBetweenMb(300, 300))
	assert.True(t, journal.poolStructuresFootprintIsBetweenMb(0, 1))

	holdPoolInMemory(pool)
}

func TestShardedTxPool_MemoryFootprint_SourceIsMe_10x1000x30720(t *testing.T) {
	pool := newPool()

	journal := analyzeMemoryFootprint(t, pool, "0_10x1000x30720", 10, 1000, 30720, "0")
	assert.True(t, journal.txsFootprintIsBetweenMb(300, 315))
	assert.True(t, journal.poolStructuresFootprintIsBetweenMb(1, 4))

	holdPoolInMemory(pool)
}

func TestShardedTxPool_MemoryFootprint_SourceIsMe_10000x1x1024(t *testing.T) {
	pool := newPool()

	journal := analyzeMemoryFootprint(t, pool, "0_10000x1x1024", 10000, 1, 1024, "0")
	assert.True(t, journal.txsFootprintIsBetweenMb(10, 16))
	assert.True(t, journal.poolStructuresFootprintIsBetweenMb(6, 10))

	holdPoolInMemory(pool)
}

func TestShardedTxPool_MemoryFootprint_SourceIsMe_1x60000x1024(t *testing.T) {
	pool := newPool()

	journal := analyzeMemoryFootprint(t, pool, "0_1x60000x1024", 1, 60000, 256, "0")
	assert.True(t, journal.txsFootprintIsBetweenMb(30, 32))
	assert.True(t, journal.poolStructuresFootprintIsBetweenMb(10, 16))

	holdPoolInMemory(pool)
}

func TestShardedTxPool_MemoryFootprint_SourceIsMe_10x10000x100(t *testing.T) {
	pool := newPool()

	journal := analyzeMemoryFootprint(t, pool, "0_10x10000x100", 10, 10000, 100, "0")
	assert.True(t, journal.txsFootprintIsBetweenMb(36, 40))
	assert.True(t, journal.poolStructuresFootprintIsBetweenMb(20, 24))

	holdPoolInMemory(pool)
}

func TestShardedTxPool_MemoryFootprint_SourceIsMe_100000x1x1024(t *testing.T) {
	pool := newPool()

	journal := analyzeMemoryFootprint(t, pool, "0_100000x1x1024", 100000, 1, 1024, "0")
	assert.True(t, journal.txsFootprintIsBetweenMb(120, 128))
	assert.True(t, journal.poolStructuresFootprintIsBetweenMb(56, 60))

	holdPoolInMemory(pool)
}

// Many transactions per sender result in the largest memory footprint.
func TestShardedTxPool_MemoryFootprint_SourceIsMe_20x20000x100(t *testing.T) {
	pool := newPool()

	journal := analyzeMemoryFootprint(t, pool, "0_20x20000x100", 20, 20000, 100, "0")
	assert.True(t, journal.txsFootprintIsBetweenMb(150, 150))
	assert.True(t, journal.poolStructuresFootprintIsBetweenMb(82, 90))

	holdPoolInMemory(pool)
}

func TestShardedTxPool_MemoryFootprint_DestinationIsMe_150x1x1048576(t *testing.T) {
	pool := newPool()

	journal := analyzeMemoryFootprint(t, pool, "1_to_0_150x1x1048576", 150, 1, core.MegabyteSize, "1_0")
	assert.True(t, journal.txsFootprintIsBetweenMb(148, 150))
	assert.True(t, journal.poolStructuresFootprintIsBetweenMb(0, 1))

	journal = analyzeMemoryFootprint(t, pool, "4294967295_to_0_150x1x1048576", 150, 1, core.MegabyteSize, "4294967295_0")
	assert.True(t, journal.txsFootprintIsBetweenMb(148, 150))
	assert.True(t, journal.poolStructuresFootprintIsBetweenMb(0, 1))

	holdPoolInMemory(pool)
}

func analyzeMemoryFootprint(t *testing.T, pool *shardedTxPool, scenario string, numSenders int, numTxsPerSender int, payloadLengthPerTx int, cacheID string) *memoryFootprintJournal {
	journal := &memoryFootprintJournal{}

	journal.beforeGenerate = getMemStats()

	txs := generateTxs(numSenders, numTxsPerSender, payloadLengthPerTx)

	journal.afterGenerate = getMemStats()

	pprofCPU(scenario, "addition", func() {
		for _, tx := range txs {
			pool.AddData(tx.hash, tx, tx.Size(), cacheID)
		}
	})

	require.Equal(t, numSenders*numTxsPerSender, len(pool.ShardDataStore(cacheID).Keys()))

	pprofHeap(scenario, "afterAddition")
	journal.afterAddition = getMemStats()

	journal.display()

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
	runtime.GC()

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

type memoryFootprintJournal struct {
	beforeGenerate runtime.MemStats
	afterGenerate  runtime.MemStats
	afterAddition  runtime.MemStats
}

func (journal *memoryFootprintJournal) txsFootprint() uint64 {
	return uint64(core.MaxInt(0, int(journal.afterGenerate.HeapInuse)-int(journal.beforeGenerate.HeapInuse)))
}

func (journal *memoryFootprintJournal) txsFootprintIsBetweenMb(lower int, upper int) bool {
	return lower <= bToMb(journal.txsFootprint()) && bToMb(journal.txsFootprint()) <= upper
}

func (journal *memoryFootprintJournal) poolStructuresFootprint() uint64 {
	return uint64(core.MaxInt(0, int(journal.afterAddition.HeapInuse)-int(journal.afterGenerate.HeapInuse)))
}

func (journal *memoryFootprintJournal) poolStructuresFootprintIsBetweenMb(lower int, upper int) bool {
	return lower <= bToMb(journal.poolStructuresFootprint()) && bToMb(journal.poolStructuresFootprint()) <= upper
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
		Shards:               1,
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
	fmt.Println("[0]:", len(pool.ShardDataStore("0").Keys()))
	fmt.Println("[1 -> 0]:", len(pool.ShardDataStore("1_0").Keys()))
	fmt.Println("[4294967295 -> 0]:", len(pool.ShardDataStore("4294967295_0").Keys()))
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
