package state

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/stretchr/testify/assert"
)

var log = logger.DefaultLogger()
var genesisFile = "genesisEdgeCase.json"

type InitialBalance struct {
	PubKey  string `json:"pubkey"`
	Balance string `json:"balance"`
}

type Genesis struct {
	InitialBalances []*InitialBalance `json:"initialBalances"`
}

type testPair struct {
	key []byte
	val []byte
}

const generate32ByteSlices = 0
const generate32HexByteSlices = 1

func TestCreationOfTheGenesisState(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	t.Parallel()

	genesisBalances := &Genesis{}
	err := core.LoadJsonFile(genesisBalances, genesisFile, log)

	assert.Nil(t, err)

	fmt.Printf("Loaded %d entries...\n", len(genesisBalances.InitialBalances))

	referenceRootHash, adbReference := getRootHashByRunningInitialBalances(genesisBalances.InitialBalances)
	fmt.Printf("Root hash: %s\n", base64.StdEncoding.EncodeToString(referenceRootHash))

	_, _ = adbReference.RootHash()

	noOfTests := 1000
	for i := 0; i < noOfTests; i++ {
		rootHash, adb := getRootHashByRunningInitialBalances(genesisBalances.InitialBalances)
		if !bytes.Equal(rootHash, referenceRootHash) {
			_, _ = adb.RootHash()
			fmt.Printf("**** Wrong root hash on iteration %d: %s\n", i, base64.StdEncoding.EncodeToString(rootHash))
			assert.Fail(t, "wrong root hash encountered")
			return
		}
	}
}

func TestExtensionNodeToBranchEdgeCaseSet1(t *testing.T) {
	t.Parallel()

	marsh := &marshal.JsonMarshalizer{}
	hasher := sha256.Sha256{}

	tr1, _ := trie.NewTrie(integrationTests.CreateMemUnit(), marsh, hasher)
	tr2, _ := trie.NewTrie(integrationTests.CreateMemUnit(), marsh, hasher)

	key1 := "e353dd9e3da522eb366e751346195a10"
	key2 := "a5dfc2ec3b0607e820ad375c5074c510"
	key3 := "eb6d6e15652c0c4d1f73490e12c8b310"
	val := "value"

	_ = tr1.Update([]byte(key1), []byte(val))
	_ = tr1.Update([]byte(key2), []byte(val))
	_ = tr1.Update([]byte(key3), []byte(val))

	fmt.Println()
	strTr1 := tr1.String()
	fmt.Println(strTr1)

	hash1, _ := tr1.Root()
	fmt.Printf("root hash1: %s\n", base64.StdEncoding.EncodeToString(hash1))

	_ = tr2.Update([]byte(key1), []byte(val))
	_ = tr2.Update([]byte(key3), []byte(val))
	_ = tr2.Update([]byte(key2), []byte(val))
	hash2, _ := tr2.Root()
	fmt.Printf("root hash2: %s\n", base64.StdEncoding.EncodeToString(hash2))

	fmt.Println()
	strTr2 := tr2.String()
	fmt.Println(strTr2)

	assert.Equal(t, hash1, hash2)
}

func TestExtensionNodeToBranchEdgeCaseSet2(t *testing.T) {
	t.Parallel()

	marsh := &marshal.JsonMarshalizer{}
	hasher := sha256.Sha256{}

	tr1, _ := trie.NewTrie(integrationTests.CreateMemUnit(), marsh, hasher)
	tr2, _ := trie.NewTrie(integrationTests.CreateMemUnit(), marsh, hasher)

	key1 := "e353dd9e3da522eb366e751346195a10"
	key2 := "eb6d6e15652c0c4d1f73490e12c8b310"
	key3 := "6f1d4baa654871c773d0af8a4ccf0410"
	key4 := "cdb0c9d63e94c56f18a75d1d186cac10"
	key5 := "176aa1b6ff17951ba202c6e8bcdaf410"
	key6 := "a5dfc2ec3b0607e820ad375c5074c510"
	val := "value"

	_ = tr1.Update([]byte(key5), []byte(val))
	_ = tr1.Update([]byte(key3), []byte(val))
	_ = tr1.Update([]byte(key1), []byte(val))
	_ = tr1.Update([]byte(key6), []byte(val))
	_ = tr1.Update([]byte(key2), []byte(val))
	_ = tr1.Update([]byte(key4), []byte(val))

	fmt.Println()
	strTr1 := tr1.String()
	fmt.Println(strTr1)

	hash1, _ := tr1.Root()
	fmt.Printf("root hash1: %s\n", base64.StdEncoding.EncodeToString(hash1))

	_ = tr2.Update([]byte(key1), []byte(val))
	_ = tr2.Update([]byte(key2), []byte(val))
	_ = tr2.Update([]byte(key3), []byte(val))
	_ = tr2.Update([]byte(key4), []byte(val))
	_ = tr2.Update([]byte(key5), []byte(val))
	_ = tr2.Update([]byte(key6), []byte(val))

	fmt.Println()
	strTr2 := tr2.String()
	fmt.Println(strTr2)

	hash2, _ := tr2.Root()
	fmt.Printf("root hash2: %s\n", base64.StdEncoding.EncodeToString(hash2))

	assert.Equal(t, hash1, hash2)
}

func TestExtensiveUpdatesAndRemovesWithConsistencyBetweenCylcesWith32byteSlices(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	t.Parallel()

	marsh := &marshal.JsonMarshalizer{}
	hasher := sha256.Sha256{}

	totalPairs, totalPairsIdx, removablePairsIdx := generateTestData(
		1000,
		500,
		generate32ByteSlices,
	)

	numTests := 1000
	referenceTrie, _ := trie.NewTrie(integrationTests.CreateMemUnit(), marsh, hasher)
	refAfterAddRootHash, refFinalRootHash, refTotalPairsIdx, refRemovablePairsIdx := execute(
		referenceTrie,
		totalPairs,
		totalPairsIdx,
		removablePairsIdx,
	)

	for i := 0; i < numTests; i++ {
		tr, _ := trie.NewTrie(integrationTests.CreateMemUnit(), marsh, hasher)

		afterAddRootHash, finalRootHash, totalPairsIdx, removablePairsIdx := execute(
			tr,
			totalPairs,
			totalPairsIdx,
			removablePairsIdx,
		)

		if !bytes.Equal(afterAddRootHash, refAfterAddRootHash) ||
			!bytes.Equal(finalRootHash, refFinalRootHash) {

			assert.Fail(t, "mismatched root hashes")
			printTestDebugLines(
				refAfterAddRootHash,
				afterAddRootHash,
				refFinalRootHash,
				finalRootHash,
				totalPairs,
				refTotalPairsIdx,
				totalPairsIdx,
				refRemovablePairsIdx,
				removablePairsIdx,
				referenceTrie,
				tr,
			)

			return
		}
	}

	fmt.Printf("Completed %d iterations\n", numTests)
}

func TestExtensiveUpdatesAndRemovesWithConsistencyBetweenCylcesWith32HexByteSlices(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	t.Parallel()

	marsh := &marshal.JsonMarshalizer{}
	hasher := sha256.Sha256{}

	totalPairs, totalPairsIdx, removablePairsIdx := generateTestData(
		1000,
		500,
		generate32HexByteSlices,
	)

	numTests := 1000
	referenceTrie, _ := trie.NewTrie(integrationTests.CreateMemUnit(), marsh, hasher)
	refAfterAddRootHash, refFinalRootHash, refTotalPairsIdx, refRemovablePairsIdx := execute(
		referenceTrie,
		totalPairs,
		totalPairsIdx,
		removablePairsIdx,
	)

	for i := 0; i < numTests; i++ {
		tr, _ := trie.NewTrie(integrationTests.CreateMemUnit(), marsh, hasher)

		afterAddRootHash, finalRootHash, totalPairsIdx, removablePairsIdx := execute(
			tr,
			totalPairs,
			totalPairsIdx,
			removablePairsIdx,
		)

		if !bytes.Equal(afterAddRootHash, refAfterAddRootHash) ||
			!bytes.Equal(finalRootHash, refFinalRootHash) {

			assert.Fail(t, "mismatched root hashes")
			printTestDebugLines(
				refAfterAddRootHash,
				afterAddRootHash,
				refFinalRootHash,
				finalRootHash,
				totalPairs,
				refTotalPairsIdx,
				totalPairsIdx,
				refRemovablePairsIdx,
				removablePairsIdx,
				referenceTrie,
				tr,
			)

			return
		}
	}

	fmt.Printf("Completed %d iterations\n", numTests)
}

func printTestDebugLines(
	referenceAfterAddRootHash []byte,
	afterAddRootHash []byte,
	referenceFinalRootHash []byte,
	finalRootHash []byte,
	totalPairs []*testPair,
	referenceTotalPairsIdx []int,
	totalPairsIdx []int,
	referenceRemovablePairs []int,
	removablePairs []int,
	referenceTrie data.Trie,
	tr data.Trie,
) {

	fmt.Printf("expected after add root hash: %s\n", base64.StdEncoding.EncodeToString(referenceAfterAddRootHash))
	fmt.Printf("actual after add root hash: %s\n", base64.StdEncoding.EncodeToString(afterAddRootHash))
	fmt.Printf("expected final add root hash: %s\n", base64.StdEncoding.EncodeToString(referenceFinalRootHash))
	fmt.Printf("actual final root hash: %s\n", base64.StdEncoding.EncodeToString(finalRootHash))

	fmt.Println()
	fmt.Println("Test pairs as hex:")
	for _, tp := range totalPairs {
		fmt.Printf("%s %s\n", hex.EncodeToString(tp.key), hex.EncodeToString(tp.val))
	}

	fmt.Println()
	fmt.Println("Reference total pairs idx, test total pairs idx:")
	for idx := range referenceTotalPairsIdx {
		fmt.Printf("%d %d\n", referenceTotalPairsIdx[idx], totalPairsIdx[idx])
	}

	fmt.Println()
	fmt.Println("Reference removable pairs idx, test removable pairs idx:")
	for idx := range referenceRemovablePairs {
		fmt.Printf("%d %d\n", referenceRemovablePairs[idx], removablePairs[idx])
	}

	fmt.Println()
	fmt.Println("Reference trie:")
	strRefTrie := referenceTrie.String()
	fmt.Println(strRefTrie)

	fmt.Println()
	fmt.Println("Actual trie:")
	strTr := tr.String()
	fmt.Println(strTr)
}

func getRootHashByRunningInitialBalances(initialBalances []*InitialBalance) ([]byte, state.AccountsAdapter) {
	adb, _ := integrationTests.CreateAccountsDB(nil)
	addrConv, _ := addressConverters.NewPlainAddressConverter(32, "")

	uniformIndexes := make([]int, len(initialBalances))
	for i := 0; i < len(initialBalances); i++ {
		uniformIndexes[i] = i
	}
	randomIndexes, _ := fisherYatesShuffle(uniformIndexes)

	for _, idx := range randomIndexes {
		ib := initialBalances[idx]
		balance, _ := big.NewInt(0).SetString(ib.Balance, 10)

		addr, _ := addrConv.CreateAddressFromPublicKeyBytes([]byte(ib.PubKey))
		accnt, _ := adb.GetAccountWithJournal(addr)
		shardAccount := accnt.(*state.Account)
		_ = shardAccount.SetBalanceWithJournal(balance)
	}

	rootHash, _ := adb.Commit()
	return rootHash, adb
}

func fisherYatesShuffle(indexes []int) ([]int, error) {
	newIndexes := make([]int, len(indexes))
	copy(newIndexes, indexes)

	for i := len(newIndexes) - 1; i > 0; i-- {
		buffRand := make([]byte, 8)
		_, _ = rand.Reader.Read(buffRand)
		randUint64 := binary.BigEndian.Uint64(buffRand)
		idx := randUint64 % uint64(i+1)

		newIndexes[i], newIndexes[idx] = newIndexes[idx], newIndexes[i]
	}

	return newIndexes, nil
}

func execute(
	tr data.Trie,
	totalPairs []*testPair,
	totalPairsIdx []int,
	removablePairsIdx []int,
) ([]byte, []byte, []int, []int) {

	randomTotalPairsIdx, _ := fisherYatesShuffle(totalPairsIdx)
	randomRemovablePirsIdx, _ := fisherYatesShuffle(removablePairsIdx)

	for _, idx := range randomTotalPairsIdx {
		tPair := totalPairs[idx]

		_ = tr.Update(tPair.key, tPair.val)
	}
	afterAddRootHash, _ := tr.Root()

	for _, idx := range randomRemovablePirsIdx {
		tPair := totalPairs[idx]

		_ = tr.Delete(tPair.key)
	}
	finalRootHash, _ := tr.Root()

	return afterAddRootHash, finalRootHash, randomTotalPairsIdx, randomRemovablePirsIdx
}

func generateTestData(numTotalPairs int, numRemovablePairs int, generationMethod int) ([]*testPair, []int, []int) {
	totalPairs := make([]*testPair, numTotalPairs)
	totalPairsIndexes := make([]int, numTotalPairs)
	removablePairsIndexes := make([]int, numRemovablePairs)

	for i := 0; i < numTotalPairs; i++ {
		switch generationMethod {
		case generate32ByteSlices:
			sizeBuff := 32
			totalPairs[i] = &testPair{
				key: generateRandomSlice(sizeBuff),
				val: generateRandomSlice(sizeBuff),
			}
		case generate32HexByteSlices:
			sizeBuff := 16
			totalPairs[i] = &testPair{
				key: []byte(hex.EncodeToString(generateRandomSlice(sizeBuff))),
				val: []byte(hex.EncodeToString(generateRandomSlice(sizeBuff))),
			}
		}

		totalPairsIndexes[i] = i

		if i < numRemovablePairs {
			removablePairsIndexes[i] = i
		}
	}

	return totalPairs, totalPairsIndexes, removablePairsIndexes
}

func generateRandomSlice(size int) []byte {
	buff := make([]byte, size)
	_, _ = rand.Reader.Read(buff)

	return buff
}
