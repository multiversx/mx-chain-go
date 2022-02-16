package storage

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/require"
)

var maxValues = big.NewInt(256 * 256 * 256)
var randomizer = rand.Reader

func TestESDTStorageAndBasicProcessing(t *testing.T) {
	// TODO skip this test

	log.Info("application is now running")

	numESDTs := 1000000
	esdts := createESDTs(numESDTs)
	log.Info("generated ESDTs", "num", numESDTs, "first 100 ESDTs", strings.Join(esdts[:100], " "))

	numAddresses := 1000000
	addresses := generateAddresses(numAddresses)

	first100addresses := make([]string, 0, 100)
	for i := 0; i < 100; i++ {
		first100addresses = append(first100addresses, integrationTests.TestAddressPubkeyConverter.Encode(addresses[i]))
	}

	log.Info("generated addresses", "num", numAddresses, "first 100 addresses", strings.Join(first100addresses, " "))

	storer := createTestStorer(t)
	defer func() {
		_ = storer.DestroyUnit()
	}()

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
		BatchDelaySeconds: 2,
		MaxBatchSize:      45000,
		MaxOpenFiles:      10,
	}

	persister, err := storageUnit.NewDB(argDB)
	require.Nil(t, err)

	storer, err := storageUnit.NewStorageUnit(cache, persister)
	require.Nil(t, err)

	return storer
}

func createMapOfESDTsOnAddresses(addresses [][]byte, esdts []string) map[string]*