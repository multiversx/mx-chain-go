package replaymetrics

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"
)

func TestRecorderSeparatesAttemptsAndSummarizesEpochs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "replay.jsonl")
	r, err := Start("", path, "test-build")
	require.NoError(t, err)
	t.Cleanup(func() {
		if Enabled() {
			require.NoError(t, r.Close(nil))
		}
	})

	header := &block.Header{ShardID: 1, Epoch: 7, Nonce: 40, TxCount: 3}
	a := Begin(header, []byte{1}, "legacy")
	require.NotNil(t, a)
	Capture(User).Add(NodesResolved, 3)
	Capture(Data).Add(RootOpens, 2)
	RecordItem(Transaction)
	RecordItem(Transaction)
	RecordStateCommit(header, 500*time.Nanosecond)
	a.Finish("executed", []byte{2}, nil)

	Capture(Data).Add(RootOpens, 500)
	h2 := &block.Header{ShardID: 1, Epoch: 8, Nonce: 41, TxCount: 1}
	a = Begin(h2, []byte{3}, "v3")
	Capture(Data).Add(RootOpens, 1)
	RecordItem(Transaction)
	a.Finish("failed", nil, io.ErrUnexpectedEOF)
	a = Begin(h2, []byte{3}, "v3")
	Capture(Data).Add(RootOpens, 4)
	RecordItem(Transaction)
	a.Finish("executed", []byte{4}, nil)
	require.NoError(t, r.Close(nil))

	file, err := os.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, file.Close()) })
	decoder := json.NewDecoder(file)
	type result struct {
		Type           string
		Outcome        string
		Epoch          uint32
		Attempts       uint64
		ExecutedBlocks uint64                       `json:"executed_blocks"`
		StateCommitNs  uint64                       `json:"state_commit_ns"`
		StateRoot      string                       `json:"state_root"`
		Tries          map[string]map[string]uint64 `json:"tries"`
		ExecutionItems map[string]uint64            `json:"execution_items"`
	}
	var attempts, epochs []result
	kinds := make(map[string]int)
	for {
		var record result
		err = decoder.Decode(&record)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		kinds[record.Type]++
		switch record.Type {
		case "block":
			attempts = append(attempts, record)
		case "epoch":
			epochs = append(epochs, record)
		}
	}
	require.Equal(t, 1, kinds["run_start"])
	require.Equal(t, 1, kinds["run_end"])
	require.Positive(t, kinds["resource"])
	require.Len(t, attempts, 3)
	require.Equal(t, "02", attempts[0].StateRoot)
	require.Equal(t, uint64(500), attempts[0].StateCommitNs)
	require.Equal(t, "failed", attempts[1].Outcome)
	require.Len(t, epochs, 2)
	require.Equal(t, uint32(7), epochs[0].Epoch)
	require.Equal(t, uint64(3), epochs[0].Tries["main"]["nodes_resolved"])
	require.Equal(t, uint64(2), epochs[0].Tries["data"]["root_opens"])
	require.Equal(t, uint64(2), epochs[0].ExecutionItems["transaction_calls"])
	require.Equal(t, uint32(8), epochs[1].Epoch)
	require.Equal(t, uint64(2), epochs[1].Attempts)
	require.Equal(t, uint64(1), epochs[1].ExecutedBlocks)
	require.Equal(t, uint64(5), epochs[1].Tries["data"]["root_opens"])
	require.Equal(t, uint64(1), epochs[1].ExecutionItems["transaction_calls"])
}

func TestRecorderCreatesUniqueReportsUnderWorkingDirectory(t *testing.T) {
	workingDir := t.TempDir()
	first, err := Start(workingDir, "", "master")
	require.NoError(t, err)
	require.NoError(t, first.Close(nil))

	second, err := Start(workingDir, "", "feature")
	require.NoError(t, err)
	require.NoError(t, second.Close(nil))

	for _, path := range []string{first.Path(), second.Path()} {
		require.Equal(t, filepath.Join(workingDir, "stats"), filepath.Dir(path))
		require.Regexp(t, `^import-db-[0-9]{8}T[0-9]{6}Z-[0-9]+\.jsonl$`, filepath.Base(path))
		content, readErr := os.ReadFile(path)
		require.NoError(t, readErr)
		require.Contains(t, string(content), `"type":"run_end"`)
	}
	require.NotEqual(t, first.Path(), second.Path())
}
