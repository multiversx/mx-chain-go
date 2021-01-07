package genesisState

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/parsing"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/stretchr/testify/assert"
)

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

	genesisFile := "genesisEdgeCase.json"

	accountsParser, err := parsing.NewAccountsParser(
		genesisFile,
		big.NewInt(6000000000),
		integrationTests.TestAddressPubkeyConverter,
		&mock.KeyGenMock{},
	)
	assert.Nil(t, err)

	fmt.Printf("Loaded %d entries...\n", len(accountsParser.InitialAccounts()))

	referenceRootHash, adbReference := getRootHashByRunningInitialAccounts(accountsParser.InitialAccounts())
	fmt.Printf("Root hash: %s\n", base64.StdEncoding.EncodeToString(referenceRootHash))

	_, _ = adbReference.RootHash()

	noOfTests := 1000
	for i := 0; i < noOfTests; i++ {
		rootHash, adb := getRootHashByRunningInitialAccounts(accountsParser.InitialAccounts())
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

	tr1 := integrationTests.CreateNewDefaultTrie()
	tr2 := integrationTests.CreateNewDefaultTrie()

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

	tr1 := integrationTests.CreateNewDefaultTrie()
	tr2 := integrationTests.CreateNewDefaultTrie()

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

func TestExtensiveUpdatesAndRemovesWithConsistencyBetweenCyclesWith32byteSlices(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	t.Parallel()

	totalPairs, totalPairsIdx, removablePairsIdx := generateTestData(
		1000,
		500,
		generate32ByteSlices,
	)

	numTests := 1000
	referenceTrie := integrationTests.CreateNewDefaultTrie()
	refAfterAddRootHash, refFinalRootHash, refTotalPairsIdx, refRemovablePairsIdx := execute(
		referenceTrie,
		totalPairs,
		totalPairsIdx,
		removablePairsIdx,
	)

	for i := 0; i < numTests; i++ {
		tr := integrationTests.CreateNewDefaultTrie()
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

func TestExtensiveUpdatesAndRemovesWithConsistencyBetweenCyclesWith32HexByteSlices(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	t.Parallel()

	totalPairs, totalPairsIdx, removablePairsIdx := generateTestData(
		1000,
		500,
		generate32HexByteSlices,
	)

	numTests := 1000
	referenceTrie := integrationTests.CreateNewDefaultTrie()
	refAfterAddRootHash, refFinalRootHash, refTotalPairsIdx, refRemovablePairsIdx := execute(
		referenceTrie,
		totalPairs,
		totalPairsIdx,
		removablePairsIdx,
	)

	for i := 0; i < numTests; i++ {
		tr := integrationTests.CreateNewDefaultTrie()

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

func getRootHashByRunningInitialAccounts(initialAccounts []genesis.InitialAccountHandler) ([]byte, state.AccountsAdapter) {
	trieStorage, _ := integrationTests.CreateTrieStorageManager(integrationTests.CreateMemUnit())
	adb, _ := integrationTests.CreateAccountsDB(0, trieStorage)

	uniformIndexes := make([]int, len(initialAccounts))
	for i := 0; i < len(initialAccounts); i++ {
		uniformIndexes[i] = i
	}
	randomIndexes, _ := fisherYatesShuffle(uniformIndexes)

	for _, idx := range randomIndexes {
		ia := initialAccounts[idx]
		decoded, _ := integrationTests.TestAddressPubkeyConverter.Decode(ia.GetAddress())
		integrationTests.MintAddress(adb, decoded, ia.GetBalanceValue())
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
	randomRemovablePairsIdx, _ := fisherYatesShuffle(removablePairsIdx)

	for _, idx := range randomTotalPairsIdx {
		tPair := totalPairs[idx]

		_ = tr.Update(tPair.key, tPair.val)
	}
	afterAddRootHash, _ := tr.Root()

	for _, idx := range randomRemovablePairsIdx {
		tPair := totalPairs[idx]

		_ = tr.Delete(tPair.key)
	}
	finalRootHash, _ := tr.Root()

	return afterAddRootHash, finalRootHash, randomTotalPairsIdx, randomRemovablePairsIdx
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
				key: integrationTests.GenerateRandomSlice(sizeBuff),
				val: integrationTests.GenerateRandomSlice(sizeBuff),
			}
		case generate32HexByteSlices:
			sizeBuff := 16
			totalPairs[i] = &testPair{
				key: []byte(hex.EncodeToString(integrationTests.GenerateRandomSlice(sizeBuff))),
				val: []byte(hex.EncodeToString(integrationTests.GenerateRandomSlice(sizeBuff))),
			}
		}

		totalPairsIndexes[i] = i

		if i < numRemovablePairs {
			removablePairsIndexes[i] = i
		}
	}

	return totalPairs, totalPairsIndexes, removablePairsIndexes
}
