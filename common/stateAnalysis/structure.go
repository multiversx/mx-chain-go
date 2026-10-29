package stateAnalysis

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/bits"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
)

const (
	numStructDepth   = 32
	numFanOut        = 18
	numKeyLen        = 66
	numLog2          = 40
	numLeafVersions  = 3
	hashSize         = 32
	autoBalanceMeta  = 6 + hashSize + hashSize // proto tags+lengths, Key, Address
	compactBranchHdr = 4                       // bitmap + version byte + type byte
)

// StructureCollector is implemented by snapshot statistics that also collect node structure
type StructureCollector interface {
	AddBranchStructure(level int, size int, numChildren int)
	AddExtensionStructure(level int, size int, keyLen int)
	AddLeafStructure(level int, size int, keyLen int, valueLen int, version core.TrieNodeVersion)
}

type trieKindStructure struct {
	Nodes         [numNodeTypes]atomic.Int64
	Bytes         [numNodeTypes]atomic.Int64
	DepthNodes    [numNodeTypes][numStructDepth]atomic.Int64
	DepthBytes    [numNodeTypes][numStructDepth]atomic.Int64
	FanOut        [numFanOut]atomic.Int64
	ExtKeyLen     [numKeyLen]atomic.Int64
	LeafKeyLen    [numKeyLen]atomic.Int64
	LeafKeyBytes  atomic.Int64
	LeafValBytes  atomic.Int64
	LeafValLog2   [numLog2]atomic.Int64
	LeafVersion   [numLeafVersions]atomic.Int64
	LeafMetaEst   atomic.Int64
	Compact       atomic.Int64
	Tries         atomic.Int64
	TrieBytesLog2 [numLog2]atomic.Int64
	TrieNodesLog2 [numLog2]atomic.Int64
	TrieDepth     [numStructDepth]atomic.Int64
}

type accountStructure struct {
	Accounts       atomic.Int64
	WithCode       atomic.Int64
	WithDataTrie   atomic.Int64
	WithBoth       atomic.Int64
	AccountBytes   atomic.Int64
	CodeEntries    atomic.Int64
	CodeBytes      atomic.Int64
	CodeReferences atomic.Int64
	LeafValueBytes atomic.Int64
	OtherLeaves    atomic.Int64
}

type structureStats struct {
	kinds    [numKinds]*trieKindStructure
	accounts accountStructure
	started  atomic.Int64
}

func newStructureStats() *structureStats {
	s := &structureStats{}
	for i := range s.kinds {
		s.kinds[i] = &trieKindStructure{}
	}
	return s
}

func log2Bucket(v int64) int {
	if v <= 0 {
		return 0
	}
	b := bits.Len64(uint64(v))
	if b >= numLog2 {
		b = numLog2 - 1
	}
	return b
}

func clampIdx(v, n int) int {
	if v < 0 {
		return 0
	}
	if v >= n {
		return n - 1
	}
	return v
}

func (s *trieKindStructure) addNode(t, level, size int) {
	s.Nodes[t].Add(1)
	s.Bytes[t].Add(int64(size))
	d := clampIdx(level, numStructDepth)
	s.DepthNodes[t][d].Add(1)
	s.DepthBytes[t][d].Add(int64(size))
}

// snapshotStats wraps the trie statistics handler of one snapshotted trie
type snapshotStats struct {
	common.TrieStatisticsHandler
	unit int
	kind int
}

// WrapSnapshotStats returns a statistics handler that also records node structure for the given unit and kind
func WrapSnapshotStats(stats common.TrieStatisticsHandler, identifier string, isDataTrie bool) common.TrieStatisticsHandler {
	if !enabled.Load() {
		return stats
	}
	unit := UnitFromIdentifier(identifier)
	kind := KindMain
	if unit == UnitPeer {
		kind = KindPeer
	} else if isDataTrie {
		kind = KindData
	}
	return &snapshotStats{TrieStatisticsHandler: stats, unit: unit, kind: kind}
}

// UnwrapSnapshotStats returns the wrapped statistics handler
func UnwrapSnapshotStats(stats common.TrieStatisticsHandler) common.TrieStatisticsHandler {
	if ss, ok := stats.(*snapshotStats); ok {
		return ss.TrieStatisticsHandler
	}
	return stats
}

func (ss *snapshotStats) structure() *trieKindStructure {
	return global.structure[ss.unit].kinds[ss.kind]
}

// AddBranchStructure records a branch node
func (ss *snapshotStats) AddBranchStructure(level int, size int, numChildren int) {
	s := ss.structure()
	s.addNode(NodeBranch, level, size)
	s.FanOut[clampIdx(numChildren, numFanOut)].Add(1)
	s.Compact.Add(int64(compactBranchHdr + numChildren*hashSize))
}

// AddExtensionStructure records an extension node
func (ss *snapshotStats) AddExtensionStructure(level int, size int, keyLen int) {
	s := ss.structure()
	s.addNode(NodeExtension, level, size)
	s.ExtKeyLen[clampIdx(keyLen, numKeyLen)].Add(1)
	s.Compact.Add(int64(2 + (keyLen+1)/2 + hashSize))
}

// AddLeafStructure records a leaf node
func (ss *snapshotStats) AddLeafStructure(level int, size int, keyLen int, valueLen int, version core.TrieNodeVersion) {
	s := ss.structure()
	s.addNode(NodeLeaf, level, size)
	s.LeafKeyLen[clampIdx(keyLen, numKeyLen)].Add(1)
	s.LeafKeyBytes.Add(int64(keyLen))
	s.LeafValBytes.Add(int64(valueLen))
	s.LeafValLog2[log2Bucket(int64(valueLen))].Add(1)
	s.LeafVersion[clampIdx(int(version), numLeafVersions)].Add(1)
	meta := 0
	if ss.kind == KindData && version == core.AutoBalanceEnabled {
		meta = autoBalanceMeta
	}
	s.LeafMetaEst.Add(int64(meta))
	s.Compact.Add(int64(2 + (keyLen+1)/2 + valueLen - meta))
}

// ObserveSnapshotTrie records the size of one snapshotted trie
func ObserveSnapshotTrie(stats common.TrieStatisticsHandler) {
	ss, ok := stats.(*snapshotStats)
	if !ok || !enabled.Load() {
		return
	}
	s := ss.structure()
	s.Tries.Add(1)
	s.TrieBytesLog2[log2Bucket(int64(ss.GetTotalNodesSize()))].Add(1)
	s.TrieNodesLog2[log2Bucket(int64(ss.GetTotalNumNodes()))].Add(1)
	s.TrieDepth[clampIdx(int(ss.GetMaxTrieDepth()), numStructDepth)].Add(1)
}

// ObserveSnapshotAccount records one main trie leaf seen during the snapshot of the given unit
func ObserveSnapshotAccount(identifier string, leafBytes int, isAccount bool, hasCode bool, hasDataTrie bool) {
	if !enabled.Load() {
		return
	}
	a := &global.structure[UnitFromIdentifier(identifier)].accounts
	a.LeafValueBytes.Add(int64(leafBytes))
	if !isAccount {
		a.OtherLeaves.Add(1)
		return
	}
	a.Accounts.Add(1)
	a.AccountBytes.Add(int64(leafBytes))
	if hasCode {
		a.WithCode.Add(1)
	}
	if hasDataTrie {
		a.WithDataTrie.Add(1)
	}
	if hasCode && hasDataTrie {
		a.WithBoth.Add(1)
	}
}

// ObserveSnapshotCodeEntry records one code entry leaf
func ObserveSnapshotCodeEntry(identifier string, codeBytes int, numReferences uint32) {
	if !enabled.Load() {
		return
	}
	a := &global.structure[UnitFromIdentifier(identifier)].accounts
	a.CodeEntries.Add(1)
	a.CodeBytes.Add(int64(codeBytes))
	a.CodeReferences.Add(int64(numReferences))
}

// SnapshotStarted marks the start of a state snapshot for the given unit
func SnapshotStarted(identifier string) {
	if !enabled.Load() {
		return
	}
	global.structure[UnitFromIdentifier(identifier)].started.Store(time.Now().UnixNano())
}

func snapshotHeader() []string {
	return []string{"shard", "label", "unit", "epoch", "root_hash", "duration_s", "wall_ns", "num_nodes", "throttler_waits", "had_error",
		"main_nodes", "main_bytes", "data_tries", "data_nodes", "data_bytes", "accounts", "with_code", "with_datatrie", "code_entries", "code_bytes"}
}

// SnapshotCompleted writes the snapshot row and the structure file for the given unit and epoch, then resets the structure
func SnapshotCompleted(identifier string, epoch uint32, rootHash []byte, durationSec int64, numNodes uint64, throttlerWaits int64, hadError bool) {
	if !enabled.Load() {
		return
	}
	c := global
	unit := UnitFromIdentifier(identifier)
	s := c.structure[unit]
	wall := int64(0)
	if started := s.started.Load(); started != 0 {
		wall = time.Now().UnixNano() - started
	}
	main := s.kinds[KindMain]
	if unit == UnitPeer {
		main = s.kinds[KindPeer]
	}
	data := s.kinds[KindData]
	row := []string{
		itoa(int(c.shardID)), c.cfg.Label, unitNames[unit], itoa(int(epoch)), hex.EncodeToString(rootHash),
		i64s(durationSec), i64s(wall), u64s(numNodes), i64s(throttlerWaits), fmt.Sprintf("%t", hadError),
		i64s(sumNodes(main)), i64s(sumBytes(main)), i64s(data.Tries.Load()), i64s(sumNodes(data)), i64s(sumBytes(data)),
		i64s(s.accounts.Accounts.Load()), i64s(s.accounts.WithCode.Load()), i64s(s.accounts.WithDataTrie.Load()),
		i64s(s.accounts.CodeEntries.Load()), i64s(s.accounts.CodeBytes.Load()),
	}
	_ = c.snapshotsFile.write(row)

	path := filepath.Join(c.dir, "structure", fmt.Sprintf("epoch_%d_%s.json", epoch, unitNames[unit]))
	err := writeStructureFile(path, c.shardID, c.cfg.Label, unit, epoch, s)
	if err != nil {
		log.Warn("stateAnalysis: cannot write structure file", "err", err)
	}
	c.structure[unit] = newStructureStats()
}

func sumNodes(s *trieKindStructure) int64 {
	var total int64
	for i := range s.Nodes {
		total += s.Nodes[i].Load()
	}
	return total
}

func sumBytes(s *trieKindStructure) int64 {
	var total int64
	for i := range s.Bytes {
		total += s.Bytes[i].Load()
	}
	return total
}

type trieKindJSON struct {
	Nodes           map[string]int64   `json:"nodes"`
	Bytes           map[string]int64   `json:"bytes"`
	DepthNodes      map[string][]int64 `json:"depthNodes"`
	DepthBytes      map[string][]int64 `json:"depthBytes"`
	BranchFanOut    []int64            `json:"branchFanOut"`
	ExtensionKeyLen []int64            `json:"extensionKeyLen"`
	LeafKeyLen      []int64            `json:"leafKeyLen"`
	LeafKeyBytes    int64              `json:"leafKeyBytes"`
	LeafValueBytes  int64              `json:"leafValueBytes"`
	LeafValueLog2   []int64            `json:"leafValueLog2"`
	LeafVersions    []int64            `json:"leafVersions"`
	LeafMetadataEst int64              `json:"leafMetadataEstimate"`
	CompactBytesEst int64              `json:"compactEncodingEstimate"`
	Tries           int64              `json:"tries"`
	TrieBytesLog2   []int64            `json:"trieBytesLog2"`
	TrieNodesLog2   []int64            `json:"trieNodesLog2"`
	TrieMaxDepth    []int64            `json:"trieMaxDepth"`
}

type structureJSON struct {
	Shard    uint32                  `json:"shard"`
	Label    string                  `json:"label"`
	Unit     string                  `json:"unit"`
	Epoch    uint32                  `json:"epoch"`
	Tries    map[string]trieKindJSON `json:"tries"`
	Accounts map[string]int64        `json:"accounts"`
}

func loadArr[T any](arr []T, load func(*T) int64) []int64 {
	out := make([]int64, len(arr))
	for i := range arr {
		out[i] = load(&arr[i])
	}
	return out
}

func loadAtomic(a *atomic.Int64) int64 { return a.Load() }

func toJSON(s *trieKindStructure) trieKindJSON {
	j := trieKindJSON{
		Nodes:      make(map[string]int64),
		Bytes:      make(map[string]int64),
		DepthNodes: make(map[string][]int64),
		DepthBytes: make(map[string][]int64),
	}
	for t := 0; t < numNodeTypes; t++ {
		j.Nodes[nodeNames[t]] = s.Nodes[t].Load()
		j.Bytes[nodeNames[t]] = s.Bytes[t].Load()
		j.DepthNodes[nodeNames[t]] = loadArr(s.DepthNodes[t][:], loadAtomic)
		j.DepthBytes[nodeNames[t]] = loadArr(s.DepthBytes[t][:], loadAtomic)
	}
	j.BranchFanOut = loadArr(s.FanOut[:], loadAtomic)
	j.ExtensionKeyLen = loadArr(s.ExtKeyLen[:], loadAtomic)
	j.LeafKeyLen = loadArr(s.LeafKeyLen[:], loadAtomic)
	j.LeafKeyBytes = s.LeafKeyBytes.Load()
	j.LeafValueBytes = s.LeafValBytes.Load()
	j.LeafValueLog2 = loadArr(s.LeafValLog2[:], loadAtomic)
	j.LeafVersions = loadArr(s.LeafVersion[:], loadAtomic)
	j.LeafMetadataEst = s.LeafMetaEst.Load()
	j.CompactBytesEst = s.Compact.Load()
	j.Tries = s.Tries.Load()
	j.TrieBytesLog2 = loadArr(s.TrieBytesLog2[:], loadAtomic)
	j.TrieNodesLog2 = loadArr(s.TrieNodesLog2[:], loadAtomic)
	j.TrieMaxDepth = loadArr(s.TrieDepth[:], loadAtomic)
	return j
}

func writeStructureFile(path string, shard uint32, label string, unit int, epoch uint32, s *structureStats) error {
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return err
	}
	out := structureJSON{
		Shard: shard,
		Label: label,
		Unit:  unitNames[unit],
		Epoch: epoch,
		Tries: make(map[string]trieKindJSON),
		Accounts: map[string]int64{
			"accounts":       s.accounts.Accounts.Load(),
			"withCode":       s.accounts.WithCode.Load(),
			"withDataTrie":   s.accounts.WithDataTrie.Load(),
			"withBoth":       s.accounts.WithBoth.Load(),
			"accountBytes":   s.accounts.AccountBytes.Load(),
			"codeEntries":    s.accounts.CodeEntries.Load(),
			"codeBytes":      s.accounts.CodeBytes.Load(),
			"codeReferences": s.accounts.CodeReferences.Load(),
			"leafValueBytes": s.accounts.LeafValueBytes.Load(),
			"otherLeaves":    s.accounts.OtherLeaves.Load(),
		},
	}
	for k := 0; k < numKinds; k++ {
		if sumNodes(s.kinds[k]) == 0 {
			continue
		}
		out.Tries[kindNames[k]] = toJSON(s.kinds[k])
	}
	data, err := json.MarshalIndent(out, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
