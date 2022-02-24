package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/require"
)

var maxValues = big.NewInt(256 * 256 * 256)
var randomizer = rand.Reader
var numESDTs = 1000000
var numAddresses = 1000000

func TestESDTStorageAndBasicProcessing_Setup(t *testing.T) {
	// TODO skip this test

	log.Info("application is now running")

	esdts := createESDTs(numESDTs)

	addresses := generateAddresses(numAddresses)

	log.Info("generated addresses", "num", numAddresses)

	storer := createTestStorer(t)
	defer func() {
		_ = storer.Close()
	}()

	numESDTsPerWallet := 1000
	for i, address := range addresses {
		if i%100 == 0 {
			log.Info(fmt.Sprint(i))
		}
		startBigInt, _ := rand.Int(randomizer, big.NewInt(int64(numESDTs-numESDTsPerWallet)))
		start := int(startBigInt.Int64())
		entries := make([]*ESDTEntry, 0, numESDTsPerWallet)
		data := make([]byte, numESDTsPerWallet)
		_, _ = randomizer.Read(data)
		balance, _ := rand.Int(randomizer, big.NewInt(1000000))
		balance.Mul(balance, big.NewInt(1000000))
		for i := start; i < start+numESDTsPerWallet; i++ {
			entries = append(entries, &ESDTEntry{
				Name:    []byte(esdts[i]),
				Data:    data,
				Balance: balance,
			})
		}

		info := ESDTInfo{
			Entry: entries,
		}
		infom, _ := info.Marshal()
		_ = storer.Put(address, infom)

	}
}

func TestESDTStorageAndBasicProcessing_Size(t *testing.T) {

	storer := createTestStorer(t)
	defer func() {
		_ = storer.Close()
	}()

	i := 0
	storer.RangeKeys(func(address []byte, infom []byte) bool {
		i++
		if i%1000 == 0 {
			log.Info(fmt.Sprint(i))
		}
		return true
	})
	log.Info(fmt.Sprint(i))
}

func TestESDTStorageAndBasicProcessing_Processing(t *testing.T) {

	storer := createTestStorer(t)
	defer func() {
		_ = storer.Close()
	}()

	log.Info("start updating...")
	i := 0

	storer.RangeKeys(func(address []byte, infom []byte) bool {
		info := ESDTInfo{}
		_ = info.Unmarshal(infom)
		adressEsdts := info.GetEntry()

		esdtIndexToModify := len(adressEsdts)
		if len(adressEsdts) > 1 {
			randomIndex, _ := rand.Int(randomizer, big.NewInt(int64(len(adressEsdts)-1)))
			esdtIndexToModify = int(randomIndex.Int64())
		}

		switch true {
		case esdtIndexToModify <= 1:
			newEsdt := generateESDT(5)
			data := make([]byte, 100)
			_, _ = randomizer.Read(data)
			balance, _ := rand.Int(randomizer, big.NewInt(1000000))
			balance.Mul(balance, big.NewInt(1000000))
			adressEsdts = append(adressEsdts, &ESDTEntry{
				Name:    []byte(newEsdt),
				Data:    data,
				Balance: balance,
			})
			break
		case i%5 == 0: //add a new esdt
			newEsdt := generateESDT(5)
			data := make([]byte, 100)
			_, _ = randomizer.Read(data)
			balance, _ := rand.Int(randomizer, big.NewInt(1000000))
			balance.Mul(balance, big.NewInt(1000000))
			adressEsdts = append(adressEsdts[:esdtIndexToModify+1], adressEsdts[esdtIndexToModify:]...)
			adressEsdts[esdtIndexToModify] = &ESDTEntry{
				Name:    []byte(newEsdt),
				Data:    data,
				Balance: balance,
			}
			break
		case i%4 == 0: // delete an existing esdt
			adressEsdts = append(adressEsdts[:esdtIndexToModify], adressEsdts[esdtIndexToModify+1:]...)
			break
		case i%3 == 0: // lower the esdt balance
			esdt := adressEsdts[esdtIndexToModify]
			newBalance, _ := rand.Int(randomizer, esdt.Balance)
			esdt.Balance.Sub(esdt.Balance, newBalance)
			break
		case i%2 == 0: // increase the esdt balance
			esdt := adressEsdts[esdtIndexToModify]
			newBalance, _ := rand.Int(randomizer, esdt.Balance)
			esdt.Balance.Add(esdt.Balance, newBalance)
			break
		default:
			//do nothing but get
		}
		info.Entry = adressEsdts
		infom, _ = info.Marshal()
		_ = storer.Put(address, infom)
		i++
		if i%1000 == 0 {
			log.Info(fmt.Sprint(i))
		}
		if i == 10000 {
			return false
		}
		return true
	})
	log.Info("updating ended...")
}

func createESDTs(numESDTs int) []string {
	log.Info("generating ESDTs...")
	esdts := make([]string, numESDTs)
	for i := 0; i < numESDTs; i++ {
		esdts[i] = generateESDT(i + 1)
	}

	return esdts
}

func generateESDT(counter int) string {
	num, _ := rand.Int(randomizer, maxValues)

	return "LKMEX-" + hex.EncodeToString(num.Bytes()) + "-" + hex.EncodeToString(big.NewInt(int64(counter)).Bytes())
}

func generateAddresses(numAddresses int) [][]byte {
	log.Info("generating addresses...")

	addresses := make([][]byte, 0, numAddresses)
	for i := 0; i < numAddresses; i++ {
		addr := make([]byte, 32)
		_, _ = randomizer.Read(addr)

		addresses = append(addresses, addr)
	}

	return addresses
}

func createTestStorer(t *testing.T) storage.Storer {
	cacheConfig := storageUnit.CacheConfig{
		Name:        "store",
		Type:        "SizeLRU",
		SizeInBytes: 209715200, // 200MB
		Capacity:    500000,
	}
	cache, err := storageUnit.NewCache(cacheConfig)
	require.Nil(t, err)

	argDB := storageUnit.ArgDB{
		DBType:            "LvlDBSerial",
		Path:              "store",
		BatchDelaySeconds: 1,
		MaxBatchSize:      100,
		MaxOpenFiles:      10,
	}

	persister, err := storageUnit.NewDB(argDB)
	require.Nil(t, err)

	storer, err := storageUnit.NewStorageUnit(cache, persister)
	require.Nil(t, err)

	return storer
}
