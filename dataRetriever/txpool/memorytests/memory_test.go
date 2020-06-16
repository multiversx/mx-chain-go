package memorytests

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
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/txpool"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/require"
)

var usePprof = false

// We run all scenarios within a single test so that we minimize memory interferences (of tests running in parallel)
func TestShardedTxPool_MemoryFootprint(t *testing.T) {

	// Scenarios where source == me

	scenario := newScenario(300, 1, core.MegabyteSize, "0")
	pool := newPool()
	journal := analyzeMemoryFootprint(t, pool, scenario)
	require.True(t, journal.payloadFootprintBetween(300, 300))
	require.True(t, journal.structuralFootprintBetween(0, 1))
	keepPoolInMemoryUpToThisPoint(pool)

	scenario = newScenario(10, 1000, 30720, "0")
	pool = newPool()
	journal = analyzeMemoryFootprint(t, pool, scenario)
	require.True(t, journal.payloadFootprintBetween(300, 315))
	require.True(t, journal.structuralFootprintBetween(1, 4))
	keepPoolInMemoryUpToThisPoint(pool)

	scenario = newScenario(10000, 1, 1024, "0")
	pool = newPool()
	journal = analyzeMemoryFootprint(t, pool, scenario)
	require.True(t, journal.payloadFootprintBetween(10, 16))
	require.True(t, journal.structuralFootprintBetween(4, 10))
	keepPoolInMemoryUpToThisPoint(pool)

	scenario = newScenario(1, 60000, 256, "0")
	pool = newPool()
	journal = analyzeMemoryFootprint(t, pool, scenario)
	require.True(t, journal.payloadFootprintBetween(30, 32))
	require.True(t, journal.structuralFootprintBetween(10, 16))
	keepPoolInMemoryUpToThisPoint(pool)

	scenario = newScenario(10, 10000, 100, "0")
	pool = newPool()
	journal = analyzeMemoryFootprint(t, pool, scenario)
	require.True(t, journal.payloadFootprintBetween(36, 40))
	require.True(t, journal.structuralFootprintBetween(16, 24))
	keepPoolInMemoryUpToThisPoint(pool)

	scenario = newScenario(100000, 1, 1024, "0")
	pool = newPool()
	journal = analyzeMemoryFootprint(t, pool, scenario)
	require.True(t, journal.payloadFootprintBetween(120, 128))
	require.True(t, journal.structuralFootprintBetween(56, 60))
	keepPoolInMemoryUpToThisPoint(pool)

	// With largest memory footprint

	scenario = newScenario(100000, 6, 128, "0")
	pool = newPool()
	journal = analyzeMemoryFootprint(t, pool, scenario)
	require.True(t, journal.payloadFootprintBetween(230, 240))
	require.True(t, journal.structuralFootprintBetween(165, 185))
	keepPoolInMemoryUpToThisPoint(pool)

	scenario = newScenario(30, 20000, 128, "0")
	pool = newPool()
	journal = analyzeMemoryFootprint(t, pool, scenario)
	require.True(t, journal.payloadFootprintBetween(230, 240))
	require.True(t, journal.structuralFootprintBetween(120, 136))
	keepPoolInMemoryUpToThisPoint(pool)

	scenario = newScenario(600, 1000, 128, "0")
	pool = newPool()
	journal = analyzeMemoryFootprint(t, pool, scenario)
	require.True(t, journal.payloadFootprintBetween(230, 240))
	require.True(t, journal.structuralFootprintBetween(120, 136))
	keepPoolInMemoryUpToThisPoint(pool)

	// Scenarios where destination == me

	scenario = newScenario(150, 1, core.MegabyteSize, "1_0")
	pool = newPool()
	journal = analyzeMemoryFootprint(t, pool, scenario)
	require.True(t, journal.payloadFootprintBetween(148, 150))
	require.True(t, journal.structuralFootprintBetween(0, 1))
	scenario = newScenario(150, 1, core.MegabyteSize, "4294967295_0")
	journal = analyzeMemoryFootprint(t, pool, scenario)
	require.True(t, journal.payloadFootprintBetween(148, 150))
	require.True(t, journal.structuralFootprintBetween(0, 1))
	keepPoolInMemoryUpToThisPoint(pool)

	scenario = newScenario(10000, 1, 10240, "1_0")
	pool = newPool()
	journal = analyzeMemoryFootprint(t, pool, scenario)
	require.True(t, journal.payloadFootprintBetween(96, 128))
	require.True(t, journal.structuralFootprintBetween(0, 4))
	scenario = newScenario(10000, 1, 10240, "4294967295_0")
	journal = analyzeMemoryFootprint(t, pool, scenario)
	require.True(t, journal.payloadFootprintBetween(96, 128))
	require.True(t, journal.structuralFootprintBetween(0, 4))
	keepPoolInMemoryUpToThisPoint(pool)

	scenario = newScenario(10, 10000, 1024, "1_0")
	pool = newPool()
	journal = analyzeMemoryFootprint(t, pool, scenario)
	require.True(t, journal.payloadFootprintBetween(96, 128))
	require.True(t, journal.structuralFootprintBetween(16, 24))
	scenario = newScenario(10, 10000, 1024, "4294967295_0")
	journal = analyzeMemoryFootprint(t, pool, scenario)
	require.True(t, journal.payloadFootprintBetween(96, 128))
	require.True(t, journal.structuralFootprintBetween(16, 24))
	keepPoolInMemoryUpToThisPoint(pool)
}

type scenario struct {
	name               string
	numSenders         int
	numTxsPerSender    int
	payloadLengthPerTx int
	cacheID            string
}

func newScenario(numSenders int, numTxsPerSender int, payloadLengthPerTx int, cacheID string) *scenario {
	name := fmt.Sprintf("%dx%dx%d_%s", numSenders, numTxsPerSender, payloadLengthPerTx, cacheID)

	return &scenario{
		name:               name,
		numSenders:         numSenders,
		numTxsPerSender:    numTxsPerSender,
		payloadLengthPerTx: payloadLengthPerTx,
		cacheID:            cacheID,
	}
}

func (scenario *scenario) numTxs() int {
	return scenario.numSenders * scenario.numTxsPerSender
}

func analyzeMemoryFootprint(t *testing.T, pool dataRetriever.ShardedDataCacherNotifier, scenario *scenario) *memoryFootprintJournal {
	fmt.Println("analyzeMemoryFootprint", scenario.name)

	journal := &memoryFootprintJournal{}

	journal.beforeGenerate = getMemStats()
	txs := generateTxs(scenario.numSenders, scenario.numTxsPerSender, scenario.payloadLengthPerTx)
	journal.afterGenerate = getMemStats()

	addTxs := func() {
		for _, tx := range txs {
			pool.AddData(tx.hash, tx, tx.Size(), scenario.cacheID)
		}
	}

	if usePprof {
		pprofCPU(scenario, "addition", addTxs)
	} else {
		addTxs()
	}

	require.Equal(t, len(pool.ShardDataStore(scenario.cacheID).Keys()), scenario.numTxs())

	if usePprof {
		pprofHeap(scenario, "afterAddition")
	}

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

func pprofHeap(scenario *scenario, step string) {
	runtime.GC()

	filename := path.Join(".", "pprofoutput", fmt.Sprintf("%s_%s.pprof", scenario.name, step))
	file, err := os.Create(filename)
	if err != nil {
		panic(fmt.Sprintf("pprofHeap: %s", err))
	}

	defer file.Close()

	err = pprof.WriteHeapProfile(file)
	if err != nil {
		panic(fmt.Sprintf("pprofHeap: %s", err))
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

func (journal *memoryFootprintJournal) payloadFootprintBetween(lower int, upper int) bool {
	return lower <= bToMb(journal.txsFootprint()) && bToMb(journal.txsFootprint()) <= upper
}

func (journal *memoryFootprintJournal) poolStructuresFootprint() uint64 {
	return uint64(core.MaxInt(0, int(journal.afterAddition.HeapInuse)-int(journal.afterGenerate.HeapInuse)))
}

func (journal *memoryFootprintJournal) structuralFootprintBetween(lower int, upper int) bool {
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

func newPool() dataRetriever.ShardedDataCacherNotifier {
	config := storageUnit.CacheConfig{
		Capacity:             900000,
		SizePerSender:        60000,
		SizeInBytes:          500 * core.MegabyteSize,
		SizeInBytesPerSender: 32 * core.MegabyteSize,
		Shards:               1,
	}

	args := txpool.ArgShardedTxPool{Config: config, MinGasPrice: 200000000000, NumberOfShards: 2, SelfShardID: 0}
	pool, err := txpool.NewShardedTxPool(args)
	if err != nil {
		panic(fmt.Sprintf("newPool: %s", err))
	}

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

func keepPoolInMemoryUpToThisPoint(pool dataRetriever.ShardedDataCacherNotifier) {
	fmt.Println("[0]:", len(pool.ShardDataStore("0").Keys()))
	fmt.Println("[1 -> 0]:", len(pool.ShardDataStore("1_0").Keys()))
	fmt.Println("[4294967295 -> 0]:", len(pool.ShardDataStore("4294967295_0").Keys()))
}

func pprofCPU(scenario *scenario, step string, function func()) {
	filename := path.Join(".", "pprofoutput", fmt.Sprintf("%s_%s.CPU.pprof", scenario.name, step))
	file, err := os.Create(filename)
	if err != nil {
		panic(fmt.Sprintf("pprofCPU: %s", err))
	}

	defer file.Close()

	err = pprof.StartCPUProfile(file)
	if err != nil {
		panic(fmt.Sprintf("pprofCPU: %s", err))
	}

	function()

	pprof.StopCPUProfile()
	convertPprofToHumanReadable(filename)
}
