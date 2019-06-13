package storage

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/leveldb"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

var testMarshalizer = &marshal.JsonMarshalizer{}
var testHasher = sha256.Sha256{}

var rcvAddr = make([]byte, 32)
var sndAddr = make([]byte, 32)
var sig = make([]byte, 64)

const batchDelaySeconds = 10
const maxBatchSize = 30000

func createStoredData(nonce uint64) ([]byte, []byte) {
	tx := &transaction.Transaction{
		Nonce:     nonce,
		GasLimit:  0,
		GasPrice:  0,
		RcvAddr:   rcvAddr,
		SndAddr:   sndAddr,
		Signature: sig,
		Value:     big.NewInt(1),
	}

	txBuff, _ := testMarshalizer.Marshal(tx)
	txHash := testHasher.Compute(string(txBuff))

	return txHash, txBuff
}

func createStorage() storage.Storer {
	db, _ := leveldb.NewSerialDB("Transactions", batchDelaySeconds, maxBatchSize)
	cacher, _ := lrucache.NewCache(50000)
	store, _ := storageUnit.NewStorageUnit(
		cacher,
		db,
	)

	return store
}

func TestWriteContinously(t *testing.T) {
	t.Skip("this is not a short test")
	rand.Reader.Read(rcvAddr)
	rand.Reader.Read(sndAddr)
	rand.Reader.Read(sig)

	nbTxsWrite := 1000000
	store := createStorage()
	defer func() {
		store.DestroyUnit()
	}()

	startTime := time.Now()
	for i := 1; i <= nbTxsWrite; i++ {
		written := 10000
		if i%written == 0 {
			endTime := time.Now()
			diff := endTime.Sub(startTime)
			fmt.Printf("Written %d, total %d in %f s\n", written, i, diff.Seconds())
			startTime = time.Now()
		}

		key, val := createStoredData(uint64(i))

		err := store.Put(key, val)
		assert.Nil(t, err)
	}
}

func TestWriteReadAnDelete(t *testing.T) {
	t.Skip("this is not a short test")
	wg := &sync.WaitGroup{}
	errors := int32(0)

	rand.Reader.Read(rcvAddr)
	rand.Reader.Read(sndAddr)
	rand.Reader.Read(sig)

	store := createStorage()
	defer func() {
		store.DestroyUnit()
	}()

	nbTxsWrite := 1000000
	var maxWritten = uint64(0)

	wg.Add(3)
	chWriteDone := make(chan struct{})

	//write anonymous func
	go func(wg *sync.WaitGroup, chWriteDone chan struct{}) {
		defer wg.Done()

		written := 10000
		initTime := time.Now()
		startTime := time.Now()
		for counter := 1; counter <= nbTxsWrite; counter++ {
			if counter%written == 0 {
				endTime := time.Now()
				diff := endTime.Sub(startTime)
				cumul := endTime.Sub(initTime)
				fmt.Printf("Written %d, total %d in %f s\nCumulativeWriteTime %f\n", written, counter, diff.Seconds(), cumul.Seconds())
				var memStats runtime.MemStats
				runtime.ReadMemStats(&memStats)

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

			key, val := createStoredData(uint64(counter))

			errPut := store.Put(key, val)
			if errPut != nil {
				fmt.Print(errPut.Error())
				atomic.AddInt32(&errors, 1)
				return
			}

			atomic.StoreUint64(&maxWritten, uint64(counter))
		}
		fmt.Println("Done Writing!")
		chWriteDone <- struct{}{}
		chWriteDone <- struct{}{}
	}(wg, chWriteDone)

	mapRemovedKeys := sync.Map{}

	// remove
	go func(wg *sync.WaitGroup, chWriteDone chan struct{}) {
		defer wg.Done()

		for {
			select {
			case <-chWriteDone:
				fmt.Println("Done Removing!")
				return
			case <-time.After(time.Millisecond * 100):
				//remove happen less often than writes
			}

			if atomic.LoadUint64(&maxWritten) == 0 {
				//not written yet
				continue
			}

			maxWrittenUint64 := atomic.LoadUint64(&maxWritten)
			maxWrittenBigInt := big.NewInt(0).SetUint64(maxWrittenUint64)
			existingNonce, _ := rand.Int(rand.Reader, maxWrittenBigInt)
			key, _ := createStoredData(existingNonce.Uint64())
			errRemove := store.Remove(key)

			mapRemovedKeys.Store(string(key), struct{}{})

			if errRemove != nil {
				fmt.Println(errRemove.Error())
				atomic.AddInt32(&errors, 1)
				return
			}

			if maxWrittenUint64 == uint64(nbTxsWrite) {
				fmt.Println("Done Removing!")
				return
			}
		}
	}(wg, chWriteDone)

	// read
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		read := uint64(10000)

		<-chWriteDone

		initTime := time.Now()
		startTime := time.Now()

		maxRoutines := make(chan struct{}, 5000)
		actualRead := uint64(0)

		wgRead := &sync.WaitGroup{}
		wgRead.Add(nbTxsWrite)
		for cnt := 1; cnt <= nbTxsWrite; cnt++ {
			maxRoutines <- struct{}{}
			go func(count uint64) {
				defer func() {
					<-maxRoutines
					wgRead.Done()
				}()

				var maxWrittenUint64 uint64

				for {
					maxWrittenUint64 = atomic.LoadUint64(&maxWritten)
					if count <= maxWrittenUint64 {
						break
					}

					select {
					case <-time.After(time.Microsecond):
					}
				}

				key, val := createStoredData(count)
				v, errGet := store.Get(key)
				_, ok := mapRemovedKeys.Load(string(key))

				if !ok && errGet != nil {
					fmt.Printf("Not getting tx with nonce %d\n", count)
					atomic.AddInt32(&errors, 1)
					return
				}

				if !ok && !bytes.Equal(val, v) {
					fmt.Printf("Not equal values with nonce %d\n", count)
					atomic.AddInt32(&errors, 1)
					return
				}

				aRead := atomic.AddUint64(&actualRead, 1)

				if aRead%read == 0 {
					endTime := time.Now()
					diff := endTime.Sub(startTime)
					cumul := endTime.Sub(initTime)
					fmt.Printf("Read %d, total %d in %f s\nCumulativeReadTime %f\n", read, aRead, diff.Seconds(), cumul.Seconds())
					var memStats runtime.MemStats
					runtime.ReadMemStats(&memStats)

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
	}(wg)

	wg.Wait()
	assert.Equal(t, int32(0), errors)
}
