package integrationTests

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/leveldb"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

const batchDelaySeconds = 10
const maxBatchSize = 30000
const maxOpenFiles = 10

// TestStorage represents a container type of class used in integration tests for storage
type TestStorage struct {
	rcvAddr        []byte
	sndAddr        []byte
	sig            []byte
	nbTxs          int
	mapRemovedKeys sync.Map
	maxWritten     *uint64
}

// NewTestStorage return an object of type TestStorage
func NewTestStorage() TestStorage {
	testStorage := new(TestStorage)

	testStorage.rcvAddr = make([]byte, 32)
	testStorage.sndAddr = make([]byte, 32)
	testStorage.sig = make([]byte, 64)

	_, _ = rand.Reader.Read(testStorage.rcvAddr)
	_, _ = rand.Reader.Read(testStorage.sndAddr)
	_, _ = rand.Reader.Read(testStorage.sig)

	return *testStorage
}

// InitAdditionalFieldsForStorageOperations init additional structure fields to can do storage operations
func (ts *TestStorage) InitAdditionalFieldsForStorageOperations(nbTxsWrite int, mapRemovedKeys sync.Map, maxWritten *uint64) {
	if ts == nil {
		return
	}

	ts.nbTxs = nbTxsWrite
	ts.mapRemovedKeys = mapRemovedKeys
	ts.maxWritten = maxWritten
}

// CreateStoredData creates stored data
func (ts *TestStorage) CreateStoredData(nonce uint64) ([]byte, []byte) {
	tx := &transaction.Transaction{
		Nonce:     nonce,
		GasLimit:  0,
		GasPrice:  0,
		RcvAddr:   ts.rcvAddr,
		SndAddr:   ts.sndAddr,
		Signature: ts.sig,
		Value:     big.NewInt(1),
	}
	txBuff, _ := TestMarshalizer.Marshal(tx)
	txHash := TestHasher.Compute(string(txBuff))

	return txHash, txBuff
}

// CreateStorageLevelDB creates a storage levelDB
func (ts *TestStorage) CreateStorageLevelDB() storage.Storer {
	db, _ := leveldb.NewDB("Transactions", batchDelaySeconds, maxBatchSize, maxOpenFiles)
	cacher, _ := lrucache.NewCache(50000)
	store, _ := storageUnit.NewStorageUnit(
		cacher,
		db,
	)

	return store
}

// CreateStorageLevelDBSerial creates a storage levelDB serial
func (ts *TestStorage) CreateStorageLevelDBSerial() storage.Storer {
	db, _ := leveldb.NewSerialDB("Transactions", batchDelaySeconds, maxBatchSize, maxOpenFiles)
	cacher, _ := lrucache.NewCache(50000)
	store, _ := storageUnit.NewStorageUnit(
		cacher,
		db,
	)

	return store
}

// WriteMultipleWithNotif write multiple data in storage without notification
func (ts *TestStorage) WriteMultipleWithNotif(
	store storage.Storer,
	wg *sync.WaitGroup,
	chWriteDone chan struct{},
	nbNotifDone int,
	errors *int32,
) {
	defer wg.Done()

	written := 10000
	initTime := time.Now()
	startTime := time.Now()

	for counter := 1; counter <= ts.nbTxs; counter++ {
		if counter%written == 0 {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			endTime := time.Now()
			diff := endTime.Sub(startTime)
			cumul := endTime.Sub(initTime)
			writesPerSecond := float64(counter) / cumul.Seconds()

			fmt.Printf("Written %d, total %d in %f s\nCumulativeWriteTime %f writes/s:%f\n",
				written,
				counter,
				diff.Seconds(),
				cumul.Seconds(),
				writesPerSecond)

			fmt.Printf("timestamp: %d, num go: %d, go mem: %s, sys mem: %s, total mem: %s, num GC: %d\n",
				time.Now().Unix(),
				runtime.NumGoroutine(),
				core.ConvertBytes(memStats.Alloc),
				core.ConvertBytes(memStats.Sys),
				core.ConvertBytes(memStats.TotalAlloc),
				memStats.NumGC,
			)
			startTime = time.Now()
		}

		key, val := ts.CreateStoredData(uint64(counter))
		errPut := store.Put(key, val)
		if errPut != nil {
			fmt.Print(errPut.Error())
			atomic.AddInt32(errors, 1)
			return
		}

		atomic.StoreUint64(ts.maxWritten, uint64(counter))
	}

	fmt.Println("Done Writing!")

	for i := 0; i < nbNotifDone; i++ {
		chWriteDone <- struct{}{}
	}
}

// RemoveMultiple remove multiple data from storage
func (ts *TestStorage) RemoveMultiple(
	store storage.Storer,
	wg *sync.WaitGroup,
	chEndRemove chan struct{},
	errors *int32,
) {
	defer wg.Done()

	for {
		select {
		case <-chEndRemove:
			fmt.Println("Done Removing!")
			return
		case <-time.After(time.Millisecond * 100):
			//remove happen less often than writes
		}

		if atomic.LoadUint64(ts.maxWritten) == 0 {
			//not written yet
			continue
		}

		maxWrittenUint64 := atomic.LoadUint64(ts.maxWritten)
		maxWrittenBigInt := big.NewInt(0).SetUint64(maxWrittenUint64)
		existingNonce, _ := rand.Int(rand.Reader, maxWrittenBigInt)
		key, _ := ts.CreateStoredData(existingNonce.Uint64())
		ts.mapRemovedKeys.Store(string(key), struct{}{})

		errRemove := store.Remove(key)
		if errRemove != nil {
			fmt.Println(errRemove.Error())
			atomic.AddInt32(errors, 1)
			return
		}

		if maxWrittenUint64 == uint64(ts.nbTxs) {
			fmt.Println("Done Removing!")
			return
		}
	}
}

// ReadMultiple  read multiple data from storage
func (ts *TestStorage) ReadMultiple(
	store storage.Storer,
	wg *sync.WaitGroup,
	chStartTrigger chan struct{},
	errors *int32,
) {
	defer wg.Done()

	<-chStartTrigger

	read := uint64(10000)
	initTime := time.Now()
	startTime := time.Now()
	maxRoutines := make(chan struct{}, 5000)
	actualRead := uint64(0)
	wgRead := &sync.WaitGroup{}
	wgRead.Add(ts.nbTxs)

	for cnt := 1; cnt <= ts.nbTxs; cnt++ {
		maxRoutines <- struct{}{}
		go func(count uint64) {
			defer func() {
				<-maxRoutines
				wgRead.Done()
			}()

			var maxWrittenUint64 uint64

			for {
				maxWrittenUint64 = atomic.LoadUint64(ts.maxWritten)
				if count <= maxWrittenUint64 {
					break
				}

				<-time.After(time.Microsecond)
			}

			key, val := ts.CreateStoredData(count)
			v, errGet := store.Get(key)
			_, ok := ts.mapRemovedKeys.Load(string(key))
			if !ok && errGet != nil {
				fmt.Printf("Not getting tx with nonce %d\n", count)
				atomic.AddInt32(errors, 1)
				return
			}

			if !ok && !bytes.Equal(val, v) {
				fmt.Printf("Not equal values with nonce %d\n", count)
				atomic.AddInt32(errors, 1)
				return
			}

			aRead := atomic.AddUint64(&actualRead, 1)
			if aRead%read == 0 {
				var memStats runtime.MemStats
				runtime.ReadMemStats(&memStats)
				endTime := time.Now()
				diff := endTime.Sub(startTime)
				cumul := endTime.Sub(initTime)
				readsPerSecond := float64(aRead) / cumul.Seconds()

				fmt.Printf("Read %d, total %d in %f s\nCumulativeReadTime %f reads/s %f\n",
					read,
					aRead,
					diff.Seconds(),
					cumul.Seconds(),
					readsPerSecond,
				)

				fmt.Printf("timestamp: %d, num go: %d, go mem: %s, sys mem: %s, total mem: %s, num GC: %d\n",
					time.Now().Unix(),
					runtime.NumGoroutine(),
					core.ConvertBytes(memStats.Alloc),
					core.ConvertBytes(memStats.Sys),
					core.ConvertBytes(memStats.TotalAlloc),
					memStats.NumGC,
				)
				startTime = time.Now()
			}
		}(uint64(cnt))
	}
	wgRead.Wait()
	fmt.Println("Done Reading!")
}
