package trieExport

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/keyValStorage"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/require"
)

//TODO add unit tests

func TestTrieExport_ExportTrieShouldExportNodesSetupJson(t *testing.T) {
	t.Parallel()

	testFolderName := "testFilesExportNodes"
	_ = os.Mkdir(testFolderName, 0777)

	defer func() {
		_ = os.RemoveAll(testFolderName)
	}()

	hs := &mock.HardforkStorerStub{
		WriteCalled: func(identifier string, key []byte, value []byte) error {
			return nil
		},
	}

	pubKeyConv := &mock.PubkeyConverterStub{
		EncodeCalled: func(pkBytes []byte) string {
			return string(pkBytes)
		},
	}

	trie := &mock.TrieStub{
		RootCalled: func() ([]byte, error) {
			return []byte{}, nil
		},
		GetAllLeavesOnChannelCalled: func(rootHash []byte) (chan core.KeyValueHolder, error) {
			ch := make(chan core.KeyValueHolder)

			mm := &mock.MarshalizerMock{}
			valInfo := &state.ValidatorInfo{List: string(core.EligibleList)}
			pacB, _ := mm.Marshal(valInfo)

			go func() {
				ch <- keyValStorage.NewKeyValStorage([]byte("test"), pacB)
				close(ch)
			}()

			return ch, nil
		},
	}

	stateExporter, err := NewTrieExport(
		testFolderName,
		mock.NewOneShardCoordinatorMock(),
		&mock.MarshalizerMock{},
		hs,
		pubKeyConv,
		pubKeyConv,
		&mock.GenesisNodesSetupHandlerStub{},
	)
	require.NoError(t, err)

	require.False(t, check.IfNil(stateExporter))

	err = stateExporter.ExportValidatorTrie(trie, context.Background())
	require.NoError(t, err)
}

func TestStateExport_ExportNodesSetupJsonShouldExportKeysInAlphabeticalOrder(t *testing.T) {
	t.Parallel()

	testFolderName := "testFilesExportNodes2"
	_ = os.Mkdir(testFolderName, 0777)

	defer func() {
		_ = os.RemoveAll(testFolderName)
	}()

	hs := &mock.HardforkStorerStub{
		WriteCalled: func(identifier string, key []byte, value []byte) error {
			return nil
		},
	}

	pubKeyConv := &mock.PubkeyConverterStub{
		EncodeCalled: func(pkBytes []byte) string {
			return string(pkBytes)
		},
	}

	stateExporter, err := NewTrieExport(
		testFolderName,
		mock.NewOneShardCoordinatorMock(),
		&mock.MarshalizerMock{},
		hs,
		pubKeyConv,
		pubKeyConv,
		&mock.GenesisNodesSetupHandlerStub{},
	)
	require.NoError(t, err)

	require.False(t, check.IfNil(stateExporter))

	vals := make(map[uint32][]*state.ValidatorInfo)
	val50 := &state.ValidatorInfo{ShardId: 5, PublicKey: []byte("aaa"), List: string(core.EligibleList)}
	val51 := &state.ValidatorInfo{ShardId: 5, PublicKey: []byte("bbb"), List: string(core.EligibleList)}
	val10 := &state.ValidatorInfo{ShardId: 5, PublicKey: []byte("ccc"), List: string(core.EligibleList)}
	val11 := &state.ValidatorInfo{ShardId: 5, PublicKey: []byte("ddd"), List: string(core.EligibleList)}
	val00 := &state.ValidatorInfo{ShardId: 5, PublicKey: []byte("aaaaaa"), List: string(core.EligibleList)}
	val01 := &state.ValidatorInfo{ShardId: 5, PublicKey: []byte("bbbbbb"), List: string(core.EligibleList)}
	vals[1] = []*state.ValidatorInfo{val50, val51}
	vals[0] = []*state.ValidatorInfo{val00, val01}
	vals[2] = []*state.ValidatorInfo{val10, val11}
	err = stateExporter.exportNodesSetupJson(vals)
	require.Nil(t, err)

	var nodesSetup sharding.NodesSetup

	nsBytes, err := ioutil.ReadFile(filepath.Join(testFolderName, core.NodesSetupJsonFileName))
	require.NoError(t, err)

	err = json.Unmarshal(nsBytes, &nodesSetup)
	require.NoError(t, err)

	initialNodes := nodesSetup.InitialNodes

	// results should be in alphabetical order, sorted by public key
	require.Equal(t, string(val50.PublicKey), initialNodes[0].PubKey) // aaa
	require.Equal(t, string(val00.PublicKey), initialNodes[1].PubKey) // aaaaaa
	require.Equal(t, string(val51.PublicKey), initialNodes[2].PubKey) // bbb
	require.Equal(t, string(val01.PublicKey), initialNodes[3].PubKey) // bbbbbb
	require.Equal(t, string(val10.PublicKey), initialNodes[4].PubKey) // ccc
	require.Equal(t, string(val11.PublicKey), initialNodes[5].PubKey) // ddd
}
