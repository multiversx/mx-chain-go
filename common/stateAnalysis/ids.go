package stateAnalysis

import (
	"strconv"
	"strings"
)

// CounterID identifies a monotonically increasing counter. Per-block deltas are exported.
type CounterID int32

// GaugeID identifies a sampled value. The last value set before a row is written is exported.
type GaugeID int32

// MaxID identifies a running maximum that is reset at every block end.
type MaxID int32

// Trie kinds used to attribute node-level counters
const (
	KindMain = iota
	KindData
	KindPeer
	numKinds
)

// Storage units used to attribute storage-level counters
const (
	UnitAccounts = iota
	UnitPeer
	UnitOther
	numUnits
)

// reuse-distance / hot-set families, independent of the trie kinds
const (
	ReuseAccount = iota
	ReuseDataTrieKey
	ReuseDataTrie
	numReuse
)

// Adapters used to attribute accountsDB-level counters
const (
	AdapterUser = iota
	AdapterPeer
	numAdapters
)

// Node types
const (
	NodeBranch = iota
	NodeExtension
	NodeLeaf
	numNodeTypes
)

const (
	numDepthBuckets     = 16
	numPersisterBuckets = 6
)

var (
	kindNames      = [numKinds]string{"main", "data", "peer"}
	unitNames      = [numUnits]string{"accounts", "peer", "other"}
	reuseKindNames = [numReuse]string{"account", "datatrie_key", "datatrie"}
	adapterNames   = [numAdapters]string{"user", "peer"}
	nodeNames      = [numNodeTypes]string{"branch", "ext", "leaf"}
)

type registry struct {
	counterNames []string
	counterMax   []bool
	gaugeNames   []string
	maxNames     []string
}

var reg = &registry{}

func counter(name string) CounterID {
	reg.counterNames = append(reg.counterNames, name)
	reg.counterMax = append(reg.counterMax, strings.HasSuffix(name, "_ns"))
	return CounterID(len(reg.counterNames) - 1)
}

func counterWithMax(name string) CounterID {
	id := counter(name)
	reg.counterMax[id] = true
	return id
}

func gauge(name string) GaugeID {
	reg.gaugeNames = append(reg.gaugeNames, name)
	return GaugeID(len(reg.gaugeNames) - 1)
}

func maxID(name string) MaxID {
	reg.maxNames = append(reg.maxNames, name)
	return MaxID(len(reg.maxNames) - 1)
}

// consecutive counters, addressed as base+i
func counterSeq(prefix string, labels []string) CounterID {
	register := counter
	if strings.HasSuffix(prefix, "_ns") {
		register = counterWithMax
	}
	base := register(prefix + "_" + labels[0])
	for i := 1; i < len(labels); i++ {
		register(prefix + "_" + labels[i])
	}
	return base
}

func counterBuckets(prefix string, n int) CounterID {
	labels := make([]string, n)
	for i := range labels {
		labels[i] = "b" + itoa(i)
	}
	return counterSeq(prefix, labels)
}

func gaugeSeq(prefix string, labels []string) GaugeID {
	base := gauge(prefix + "_" + labels[0])
	for i := 1; i < len(labels); i++ {
		gauge(prefix + "_" + labels[i])
	}
	return base
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

// ---------------------------------------------------------------------------
// Block level (Q1, Q10)
// ---------------------------------------------------------------------------

var (
	CBlocks            = counter("blk_count")
	CBlocksAborted     = counter("blk_aborted")
	CBlockWallNs       = counter("blk_wall_ns")
	CBlockProcessNs    = counter("blk_process_ns")
	CBlockCommitNs     = counter("blk_commit_state_ns")
	CBlockCPUNs        = counter("blk_cpu_ns")
	CBlockSnapshotWait = counter("blk_snapshot_wait_ns")
	CBlockTxs          = counter("blk_txs")
	CBlockStartOfEpoch = counter("blk_start_of_epoch")
)

// ---------------------------------------------------------------------------
// AccountsDB level, per adapter (Q2, Q4, Q9, Q10)
// ---------------------------------------------------------------------------

// adapterCounter registers one counter per adapter and returns the base ID
func adapterCounter(name string) CounterID {
	return counterSeq("adb_"+name, adapterNames[:])
}

func adapterCounterWithMax(name string) CounterID {
	base := adapterCounter(name)
	for i := 0; i < numAdapters; i++ {
		reg.counterMax[int(base)+i] = true
	}
	return base
}

var (
	CAdbLoadAccount            = adapterCounter("load_account")
	CAdbGetExisting            = adapterCounter("get_existing")
	CAdbGetExistingMissing     = adapterCounter("get_existing_missing")
	CAdbGetAccountNs           = adapterCounter("get_account_ns")
	CAdbGetAccountFound        = adapterCounter("get_account_found")
	CAdbGetAccountBytes        = adapterCounter("get_account_bytes")
	CAdbUnmarshalNs            = adapterCounter("account_unmarshal_ns")
	CAdbSaveAccount            = adapterCounter("save_account")
	CAdbSaveNewAccount         = adapterCounter("save_new_account")
	CAdbSaveRereadNs           = adapterCounter("save_reread_ns")
	CAdbSaveUnchanged          = adapterCounter("save_unchanged_bytes")
	CAdbSaveMarshalNs          = adapterCounter("save_marshal_ns")
	CAdbSaveBytes              = adapterCounter("save_bytes")
	CAdbSaveNs                 = adapterCounter("save_account_ns")
	CAdbSaveRepeatAddr         = adapterCounter("save_repeat_addr")
	CAdbRemoveAccount          = adapterCounter("remove_account")
	CAdbGetCode                = adapterCounter("get_code")
	CAdbGetCodeBytes           = adapterCounter("get_code_bytes")
	CAdbGetCodeNs              = adapterCounter("get_code_ns")
	CAdbCodeEntryGet           = adapterCounter("code_entry_get")
	CAdbCodeEntrySave          = adapterCounter("code_entry_save")
	CAdbDataTrieHolderHit      = adapterCounter("datatrie_holder_hit")
	CAdbDataTrieRecreate       = adapterCounter("datatrie_recreate")
	CAdbDataTrieRecreateNs     = adapterCounter("datatrie_recreate_ns")
	CAdbAccountsWithRootHash   = adapterCounter("accounts_with_roothash_read")
	CAdbSaveDataTrieDirty      = adapterCounter("save_datatrie_dirty")
	CAdbSaveDataTrieNs         = adapterCounter("save_datatrie_ns")
	CAdbSaveDataTrieRootHashNs = adapterCounter("save_datatrie_roothash_ns")
	CAdbJournalize             = adapterCounter("journalize")
	CAdbJournalLenAtCommit     = adapterCounterWithMax("journal_len_at_commit")
	CAdbRevertPartial          = adapterCounter("revert_partial")
	CAdbRevertFull             = adapterCounter("revert_full")
	CAdbRevertEntries          = adapterCounter("revert_entries")
	CAdbRecreateTrie           = adapterCounter("recreate_trie")
	CAdbRemoveDataTrie         = adapterCounter("remove_datatrie")
	CAdbRemoveDataTrieHashes   = adapterCounter("remove_datatrie_hashes")
	CAdbCommit                 = adapterCounter("commit")
	CAdbCommitNs               = adapterCounter("commit_ns")
	CAdbCommitDataTriesNs      = adapterCounter("commit_datatries_ns")
	CAdbCommitDataTries        = adapterCounterWithMax("commit_datatries")
	CAdbCommitMainTrieNs       = adapterCounter("commit_maintrie_ns")
	CAdbCommitEvictNs          = adapterCounter("commit_evict_ns")
	CAdbCommitStateAccNs       = adapterCounter("commit_stateaccesses_ns")
	CAdbCommitOldHashes        = adapterCounter("commit_old_hashes")
	CAdbCommitNewHashes        = adapterCounter("commit_new_hashes")
	CAdbUniqueAccounts         = adapterCounterWithMax("unique_accounts_block")
	CAdbUniqueAccountsPrev     = adapterCounter("unique_accounts_seen_prev_block")
	CAdbUniqueDataTries        = adapterCounterWithMax("unique_datatries_block")
	CAdbUniqueDataTriesPrev    = adapterCounter("unique_datatries_seen_prev_block")
	CAdbSnapshotStateCalls     = adapterCounter("snapshot_state_calls")
)

// ---------------------------------------------------------------------------
// Trackable data trie (Q2)
// ---------------------------------------------------------------------------

var (
	CDtRetrieve         = counter("dt_retrieve")
	CDtRetrieveDirtyHit = counter("dt_retrieve_dirty_hit")
	CDtRetrieveNs       = counter("dt_retrieve_ns")
	CDtKeyHashComputed  = counter("dt_key_hash_computed")
	CDtLookupHashed     = counter("dt_lookup_hashed")
	CDtLookupHashedHit  = counter("dt_lookup_hashed_hit")
	CDtLookupRaw        = counter("dt_lookup_raw")
	CDtLookupRawHit     = counter("dt_lookup_raw_hit")
	CDtLookupMiss       = counter("dt_lookup_miss")
	CDtSaveKeyValue     = counter("dt_save_key_value")
	CDtUpdateTrieCalls  = counter("dt_update_trie_calls")
	CDtUpdateDirtyKeys  = counterWithMax("dt_update_dirty_keys")
	CDtUpdateNs         = counter("dt_update_ns")
	CDtUpdateOldReadNs  = counter("dt_update_old_value_read_ns")
	CDtUpdateNoop       = counter("dt_update_noop")
	CDtUpdateInsert     = counter("dt_update_insert")
	CDtUpdateModify     = counter("dt_update_modify")
	CDtUpdateDelete     = counter("dt_update_delete")
	CDtUpdateDeleteMiss = counter("dt_update_delete_missing")
	CDtUpdateMigrate    = counter("dt_update_migrate")
	CDtUpdateBytesNew   = counter("dt_update_bytes_new")
	CDtUpdateBytesOld   = counter("dt_update_bytes_old")
	CDtRollback         = counter("dt_rollback")
)

// ---------------------------------------------------------------------------
// Trie level, per kind (Q3, Q5)
// ---------------------------------------------------------------------------

func kindCounter(name string) CounterID {
	return counterSeq("trie_"+name, kindNames[:])
}

func kindCounterWithMax(name string) CounterID {
	base := kindCounter(name)
	for i := 0; i < numKinds; i++ {
		reg.counterMax[int(base)+i] = true
	}
	return base
}

func kindNodeCounter(name string) CounterID {
	labels := make([]string, 0, numKinds*numNodeTypes)
	for k := 0; k < numKinds; k++ {
		for n := 0; n < numNodeTypes; n++ {
			labels = append(labels, kindNames[k]+"_"+nodeNames[n])
		}
	}
	return counterSeq("trie_"+name, labels)
}

func kindBucketCounter(name string, n int) CounterID {
	labels := make([]string, 0, numKinds*n)
	for k := 0; k < numKinds; k++ {
		for b := 0; b < n; b++ {
			labels = append(labels, kindNames[k]+"_b"+itoa(b))
		}
	}
	return counterSeq("trie_"+name, labels)
}

var (
	CTrieGet              = kindCounter("get")
	CTrieGetNs            = kindCounter("get_ns")
	CTrieGetFound         = kindCounter("get_found")
	CTrieGetDepth         = kindBucketCounter("get_depth", numDepthBuckets)
	CTrieUpdate           = kindCounter("update")
	CTrieUpdateNs         = kindCounter("update_ns")
	CTrieDelete           = kindCounter("delete")
	CTrieRootHashCalls    = kindCounter("roothash_calls")
	CTrieRootHashComputed = kindCounter("roothash_computed")
	CTrieRootHashNs       = kindCounter("roothash_ns")
	CTrieCommit           = kindCounter("commit")
	CTrieCommitClean      = kindCounter("commit_clean")
	CTrieCommitHashNs     = kindCounter("commit_hash_ns")
	CTrieCommitWriteNs    = kindCounter("commit_write_ns")
	CTrieRecreate         = kindCounter("recreate")
	CTrieRecreateNs       = kindCounter("recreate_ns")
	CTrieResidentHit      = kindCounter("resident_hit")
	CTrieNodeLoad         = kindNodeCounter("node_load")
	CTrieNodeLoadNs       = kindCounter("node_load_ns")
	CTrieNodeLoadBytes    = kindCounter("node_load_bytes")
	CTrieLoadGetDepth     = kindBucketCounter("node_load_get_depth", numDepthBuckets)
	CTrieLoadInGet        = kindCounter("node_load_in_get")
	CTrieDecodeNs         = kindCounter("decode_ns")
	CTrieEncode           = kindNodeCounter("encode")
	CTrieEncodeBytes      = kindCounter("encode_bytes")
	CTrieEncodeNs         = kindCounter("encode_ns")
	CTrieHashCompute      = kindCounter("hash_compute")
	CTrieChildReencode    = kindCounter("child_reencode")
	CTrieSetHashCached    = kindCounter("sethash_cached")
	CTrieDirtyCommitted   = kindNodeCounter("dirty_committed")
	CTrieDirtyBytes       = kindCounter("dirty_committed_bytes")
	CTrieCollapseAtMax    = kindCounter("collapse_at_max_level")
	CTrieInsertNewLeaf    = kindCounter("insert_new_leaf")
)

// ---------------------------------------------------------------------------
// Storage level, per unit (Q5, Q9)
// ---------------------------------------------------------------------------

func unitCounter(name string) CounterID {
	return counterSeq("st_"+name, unitNames[:])
}

func unitBucketCounter(name string, n int) CounterID {
	labels := make([]string, 0, numUnits*n)
	for u := 0; u < numUnits; u++ {
		for b := 0; b < n; b++ {
			labels = append(labels, unitNames[u]+"_b"+itoa(b))
		}
	}
	return counterSeq("st_"+name, labels)
}

var (
	CStTsmGet          = unitCounter("tsm_get")
	CStTsmGetNs        = unitCounter("tsm_get_ns")
	CStTsmGetBytes     = unitCounter("tsm_get_bytes")
	CStTsmGetMiss      = unitCounter("tsm_get_miss")
	CStTsmPut          = unitCounter("tsm_put")
	CStTsmPutNs        = unitCounter("tsm_put_ns")
	CStTsmPutBytes     = unitCounter("tsm_put_bytes")
	CStPsCacheHit      = unitCounter("ps_cache_hit")
	CStPsPersisterHit  = unitBucketCounter("ps_persister_hit", numPersisterBuckets)
	CStPsMiss          = unitCounter("ps_miss")
	CStPsPut           = unitCounter("ps_put")
	CStDbGet           = unitCounter("db_get")
	CStDbGetHit        = unitCounter("db_get_hit")
	CStDbGetNs         = unitCounter("db_get_ns")
	CStDbGetBytes      = unitCounter("db_get_bytes")
	CStDbGetSlow       = unitCounter("db_get_slow")
	CStDbPut           = unitCounter("db_put")
	CStDbPutNs         = unitCounter("db_put_ns")
	CStDbPutBytes      = unitCounter("db_put_bytes")
	CStSnapGet         = unitCounter("snap_get")
	CStSnapGetNs       = unitCounter("snap_get_ns")
	CStSnapGetBytes    = unitCounter("snap_get_bytes")
	CStSnapPut         = unitCounter("snap_put")
	CStSnapPutBytes    = unitCounter("snap_put_bytes")
	CStSnapPutPrev     = unitCounter("snap_put_prev_epoch")
	CStSnapPutPrevSkip = unitCounter("snap_put_prev_epoch_present")
)

// ---------------------------------------------------------------------------
// Pruning (Q9)
// ---------------------------------------------------------------------------

var (
	CPruneEwlPut           = counter("prune_ewl_put")
	CPruneEwlHashes        = counter("prune_ewl_hashes")
	CPruneEwlWipes         = counter("prune_ewl_wipes")
	CPruneEwlHashesDropped = counter("prune_ewl_hashes_dropped")
	CPruneBufferAdd        = counter("prune_buffer_add")
	CPruneBufferOverflow   = counter("prune_buffer_overflow")
	CPruneCalls            = counter("prune_calls")
	CPruneNodesRemoved     = counter("prune_nodes_removed")
	CPruneRemoveNs         = counter("prune_remove_ns")
	CPruneCancel           = counter("prune_cancel")
)

// ---------------------------------------------------------------------------
// Gauges (Q6)
// ---------------------------------------------------------------------------

func kindNodeGauge(name string) GaugeID {
	labels := make([]string, 0, numKinds*numNodeTypes)
	for k := 0; k < numKinds; k++ {
		for n := 0; n < numNodeTypes; n++ {
			labels = append(labels, kindNames[k]+"_"+nodeNames[n])
		}
	}
	return gaugeSeq("trie_"+name, labels)
}

var (
	GMemHeapAlloc       = gauge("mem_heap_alloc")
	GMemHeapInuse       = gauge("mem_heap_inuse")
	GMemHeapObjects     = gauge("mem_heap_objects")
	GMemSys             = gauge("mem_sys")
	GMemRSS             = gauge("mem_rss")
	GMemNumGC           = gauge("mem_num_gc")
	GMemGCPauseTotal    = gauge("mem_gc_pause_total_ns")
	GTrieResidentNodes  = kindNodeGauge("resident_nodes")
	GTrieResidentBytes  = kindNodeGauge("resident_bytes")
	GTrieResidentWalkNs = gauge("trie_resident_walk_ns")
	GHolderDataTries    = gaugeSeq("adb_holder_datatries", adapterNames[:])
	GStCacheEntries     = gaugeSeq("st_cache_entries", unitNames[:])
	GStCacheBytes       = gaugeSeq("st_cache_bytes", unitNames[:])
	GStActivePersisters = gaugeSeq("st_active_persisters", unitNames[:])
	GPruneEwlLen        = gauge("prune_ewl_len")
	GPruneBufferLen     = gauge("prune_buffer_len")
	GTrackedKeys        = gaugeSeq("tracked_keys", reuseKindNames[:])
)

// ---------------------------------------------------------------------------
// Running maxima, reset at every block
// ---------------------------------------------------------------------------

var (
	MStDbGetNs   = maxID("st_db_get_ns_max")
	MTrieLoadNs  = maxID("trie_node_load_ns_max")
	MDtDirtyKeys = maxID("dt_dirty_keys_per_save_max")
	MAdbSaveNs   = maxID("adb_save_account_ns_max")
)
