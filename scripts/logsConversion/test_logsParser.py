import re

import pytest

from scripts.logsConversion.logsToJsonConverter import LogsToJsonConverter
from scripts.logsConversion.sample_data import (
    comma_separated_log, common_entries_log, contract_deployed_log,
    empty_value_log, initialized_chainParametersHolder_log,
    misc_no_ascii_chars_log, monitor_new_cache_logs, network_connection_log,
    new_sharded_data_log, scheduled_txs_log, several_words_in_key_log,
    special_chars_in_parameters_log, start_time_log, statistics_entries_log,
    total_transactions_in_pool_logs, transactions_processed_table_log,
    trie_statistics_log, vm_mettering_log)


def normalize_spacing(log_entry):
    """Replaces multiple spaces with exactly one space."""
    return re.sub(r' {2,}', ' ', log_entry)


class TestLogsParsing:

    def test_regular_entries(self):
        unexpected = []

        for log in [common_entries_log, empty_value_log]:
            parser = LogsToJsonConverter('node1')
            parser.parse(log.split('\n'))
            i = 0
            for log_entry in re.split('\n', log):
                if not log_entry:
                    continue
                normalized_entry = normalize_spacing(log_entry).strip()
                if parser.log_content and normalized_entry != str(parser.log_content[i]).strip():
                    unexpected.append(f"{normalized_entry}\n{str(parser.log_content[i]).strip()}\n")
                i += 1

        if unexpected:
            pytest.fail("Unexpected log entries found:\n\n" + "\n".join(unexpected))

    def test_multiple_words_key(self):
        converter = LogsToJsonConverter('node1')

        expected_result = ['{"v": "node1", "t": 1740146957.199, "l": 1, "n": "process/sync", "s": "0", "e": "9", "r": "1008", "sr": "END_ROUND", "m": "forkDetector.appendHeaderInfo", "a": {"round": "1008", "nonce": "1006", "hash": "88b984b5cc8757fc02281e4024314ebcc5edf28552debfa4f4382d4a490b8570", "state": "2", "probable highest nonce": "1006", "last checkpoint nonce": "1006", "final checkpoint nonce": "1006", "has proof": "true"}}',
                           '{"v": "node1", "t": 1738912460.145, "l": 1, "n": "..ever/proofscache", "s": "1", "e": "0", "r": "859", "sr": "START_ROUND", "m": "added proof to pool", "a": {"header hash": "efaf3cf089a64298423e968f43d856b0caac244a5df8cea82d39506402362c05", "epoch": "4", "nonce": "849", "shardID": "4294967295", "pubKeys bitmap": "bd01", "round": "857", "isStartOfEpoch": "false"}}',
                           '{"v": "node1", "t": 1738912453.702, "l": 1, "n": "..block/preprocess", "s": "1", "e": "0", "r": "857", "sr": "BLOCK", "m": "scheduledTxsExecution.RollBackToBlock", "a": {"header hash": "5774fe20bee691ac4cded77bcde0b4d2c07584ec33b08933f65786ec920add67", "scheduled root hash": "33b86cbcf34f512c97faed4dc46d9955562dde22df91928808a977e3b37ebaa7", "num of scheduled mbs": "0", "num of scheduled intermediate txs": "0", "accumulatedFees": "0", "developerFees": "0", "gasProvided": "0", "gasPenalized": "0", "gasRefunded": "0"}}'
                           ]
        converter.parse(several_words_in_key_log.split('\n'))
        assert len(converter.log_content) == 3
        for line in converter.log_content:
            assert (line.to_dict() in expected_result)

    def test_network_connection(self):
        converter = LogsToJsonConverter('node1')

        converter.parse(network_connection_log.split('\n'))
        assert len(converter.log_content) == 2
        assert (converter.log_content[0].to_dict() == '{"v": "node1", "t": 1740156794.625, "l": 1, "n": "main/p2p", "s": "0", "e": "26", "r": "2647", "sr": "END_ROUND", "m": "network connection status", "a": {"network": "main", "known peers": "55", "connected peers": "28", "intra shard validators": "2", "intra shard observers": "7", "cross shard validators": "5", "cross shard observers": "13", "unknown": "0", "seeders": "1", "current shard": "0", "validators histogram": {"shard 0": "2", "shard 1": "2", "shard 2": "2", "meta": "1"}, "observers histogram": {"shard 0": "7", "shard 1": "4", "shard 2": "2", "meta": "7"}, "preferred peers histogram": {}}}')
        assert (converter.log_content[1].to_dict() == '{"v": "node1", "t": 1740156794.625, "l": 1, "n": "main/p2p", "s": "0", "e": "26", "r": "2647", "sr": "END_ROUND", "m": "network connection metrics", "a": {"network": "main", "connections/s": "3", "disconnections/s": "3", "connections": "63", "disconnections": "63", "time": "20s"}}')

    def test_monitor_new_cache_logs(self):
        converter = LogsToJsonConverter('node1')
        expected_result = [
            '{"v": "node1", "t": 1741683641.557, "l": 1, "n": "storage", "s": "", "e": 0, "r": 0, "sr": "", "m": "MonitorNewCache", "a": {"name": "Antiflood", "capacity": "0B", "cumulated": "0B"}}',
            '{"v": "node1", "t": 1741683641.557, "l": 1, "n": "storage", "s": "", "e": 0, "r": 0, "sr": "", "m": "MonitorNewCache", "a": {"name": "PeerHonesty", "capacity": "0B", "cumulated": "0B"}}',
            '{"v": "node1", "t": 1741683642.561, "l": 1, "n": "storage", "s": "", "e": 0, "r": 0, "sr": "", "m": "MonitorNewCache", "a": {"name": "TrieNodesDataPool", "capacity": "100.00MB", "cumulated": "700.00MB"}}',
            '{"v": "node1", "t": 1741683642.561, "l": 1, "n": "storage", "s": "", "e": 0, "r": 0, "sr": "", "m": "MonitorNewCache", "a": {"name": "SmartContractDataPool", "capacity": "200.00MB", "cumulated": "900.00MB"}}',
            '{"v": "node1", "t": 1741683642.561, "l": 1, "n": "storage", "s": "", "e": 0, "r": 0, "sr": "", "m": "MonitorNewCache", "a": {"name": "HeartbeatPool", "capacity": "300.00MB", "cumulated": "1.17GB"}}'
        ]
        converter.parse(monitor_new_cache_logs.split('\n'))
        assert len(converter.log_content) == 5
        for line in converter.log_content:
            assert (line.to_dict() in expected_result)

    def test_vm_mettering(self):
        converter = LogsToJsonConverter('node1')
        expected_result = [
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/storage", "s": "", "e": 0, "r": 0, "sr": "", "m": "storage added", "a": {"key": "6e5f626c6f636b735f6265666f72655f756e626f6e64", "value": "030d40"}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/host", "s": "", "e": 0, "r": 0, "sr": "", "m": "checkFinalGasAfterExit", "a": {}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/metering", "s": "", "e": 0, "r": 0, "sr": "", "m": "UpdateGasStateOnSuccess", "a": {}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/metering", "s": "", "e": 0, "r": 0, "sr": "", "m": "check gas", "a": {}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/metering", "s": "", "e": 0, "r": 0, "sr": "", "m": "metering state", "a": {"e2948c2d2d2d2d2d2d2d2d2d2d2020202020202020202020207363": "00000000000000000500ed8e25a94efa837aae0e593112cfbb01b448755069e1"}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/metering", "s": "", "e": 0, "r": 0, "sr": "", "m": "metering state", "a": {"function": ""}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/metering", "s": "", "e": 0, "r": 0, "sr": "", "m": "metering state", "a": {"initial provided": "9223372036854775807"}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/metering", "s": "", "e": 0, "r": 0, "sr": "", "m": "metering state", "a": {"initial cost": "17487900"}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/metering", "s": "", "e": 0, "r": 0, "sr": "", "m": "metering state", "a": {"gas for exec": "9223372036837287907"}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/metering", "s": "", "e": 0, "r": 0, "sr": "", "m": "metering state", "a": {"instance gas": "2008429"}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/metering", "s": "", "e": 0, "r": 0, "sr": "", "m": "metering state", "a": {"gas left": "9223372036835279478"}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/metering", "s": "", "e": 0, "r": 0, "sr": "", "m": "metering state", "a": {"gas spent by sc": "19496329"}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/metering", "s": "", "e": 0, "r": 0, "sr": "", "m": "metering state", "a": {"gas transferred": "0"}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/metering", "s": "", "e": 0, "r": 0, "sr": "", "m": "metering state", "a": {"gas used by others": "0"}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/metering", "s": "", "e": 0, "r": 0, "sr": "", "m": "metering state", "a": {"adjusted gas used by sc": "19496329"}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/metering", "s": "", "e": 0, "r": 0, "sr": "", "m": "metering state", "a": {"gas per acct": "19496329", "key": "00000000000000000500ed8e25a94efa837aae0e593112cfbb01b448755069e1"}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/metering", "s": "", "e": 0, "r": 0, "sr": "", "m": "metering state", "a": {"e2949420737461636b2073697a65": "0"}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/metering", "s": "", "e": 0, "r": 0, "sr": "", "m": "UpdateGasStateOnSuccess", "a": {"vmOutput.GasRemaining": "9223372036835279478"}}',
            '{"v": "node1", "t": 1741643946.489, "l": 0, "n": "vm/metering", "s": "", "e": 0, "r": 0, "sr": "", "m": "UpdateGasStateOnSuccess", "a": {"instance gas left": "9223372036835279478"}}'
        ]
        converter.parse(vm_mettering_log.split('\n'))
        for line in converter.log_content:
            assert (line.to_dict() in expected_result)

    def test_scheduled_txs(self):
        converter = LogsToJsonConverter('node1')

        expected_result = [
            '{"v": "node1", "t": 1740141083.039, "l": 1, "n": "..block/preprocess", "s": "0", "e": "0", "r": "29", "sr": "END_ROUND", "m": "scheduledTxsExecution.GetScheduledRootHash scheduled root hash =", "a": {}}',
            '{"v": "node1", "t": 1740141083.039, "l": 1, "n": "..block/preprocess", "s": "0", "e": "0", "r": "29", "sr": "END_ROUND", "m": "scheduledTxsExecution.GetScheduledIntermediateTxs num of scheduled intermediate", "a": {"txs": "0"}}',
            '{"v": "node1", "t": 1740141083.04, "l": 1, "n": "..block/preprocess", "s": "0", "e": "0", "r": "29", "sr": "END_ROUND", "m": "scheduledTxsExecution.GetScheduledGasAndFees", "a": {"accumulatedFees": "0", "developerFees": "0", "gasProvided": "0", "gasPenalized": "0", "gasRefunded": "0"}}',
            '{"v": "node1", "t": 1740141083.04, "l": 1, "n": "..block/preprocess", "s": "0", "e": "0", "r": "29", "sr": "END_ROUND", "m": "scheduledTxsExecution.SaveStateIfNeeded", "a": {"header hash": "eab8f3f0197e0ec68b71a6cf2fed98811dd740906595ae7a3d189d0c75e35a7e", "scheduled root hash": "_", "num of scheduled txs": "0", "num of scheduled intermediate txs": "0", "accumulatedFees": "0", "developerFees": "0", "gasProvided": "0", "gasPenalized": "0", "gasRefunded": "0"}}',
            '{"v": "node1", "t": 1740156809.296, "l": 1, "n": "..block/preprocess", "s": "0", "e": "26", "r": "2650", "sr": "END_ROUND", "m": "scheduledTxsExecution.GetScheduledRootHash scheduled root", "a": {"hash": "39012320e3d45859b7bf97f2b89a1d0ed0fdde20176cbb23970c0ac3199a5149"}}',
            '{"v": "node1", "t": 1740156809.296, "l": 1, "n": "..block/preprocess", "s": "0", "e": "26", "r": "2650", "sr": "END_ROUND", "m": "scheduledTxsExecution.GetScheduledIntermediateTxs num of scheduled intermediate", "a": {"txs": "0"}}',
            '{"v": "node1", "t": 1740156809.296, "l": 1, "n": "..block/preprocess", "s": "0", "e": "26", "r": "2650", "sr": "END_ROUND", "m": "scheduledTxsExecution.GetScheduledGasAndFees", "a": {"accumulatedFees": "0", "developerFees": "0", "gasProvided": "0", "gasPenalized": "0", "gasRefunded": "0"}}',
            '{"v": "node1", "t": 1740156809.296, "l": 1, "n": "..block/preprocess", "s": "0", "e": "26", "r": "2650", "sr": "END_ROUND", "m": "scheduledTxsExecution.SaveStateIfNeeded", "a": {"header hash": "cc77686defc6f5423759185f1cb07dbfb0cf10a2472d7d77558f94482831c0b3", "scheduled root hash": "39012320e3d45859b7bf97f2b89a1d0ed0fdde20176cbb23970c0ac3199a5149", "num of scheduled txs": "0", "num of scheduled intermediate txs": "0", "accumulatedFees": "0", "developerFees": "0", "gasProvided": "0", "gasPenalized": "0", "gasRefunded": "0"}}'
        ]
        converter.parse(scheduled_txs_log.split('\n'))
        assert len(converter.log_content) == 8
        for line in converter.log_content:
            assert (line.to_dict() in expected_result)

    def test_start_time(self):
        converter = LogsToJsonConverter('node1')
        converter.parse(start_time_log.split('\n'))
        assert len(converter.log_content) == 1
        entry = converter.log_content[0]
        assert entry.to_dict() == '{"v": "node1", "t": 1740140913.738, "l": 2, "n": "factory", "s": "", "e": 0, "r": 0, "sr": "", "m": "start time", "a": {"formatted": "Fri Feb 21 14:28:29 UTC 2025", "seconds": "1740148109"}}'

    def test_block_header_proof_sent(self):
        block_header_proof_sent_log = 'DEBUG[2025-02-21 18:53:23.109] [..nsus/spos/bls/v2] [0/26/2649/(END_ROUND)] step 3: block header proof has been sent PubKeysBitmap = fd02 AggregateSignature = d94a7f18409f9d931a8f96691ea9f2b875281125bf71f9956f5223ccb6fb8d89bfd5f0e8abfeb4975e126c4e250d0988 proof sender = f547e115b1ada7cf9b8aeef45ee0d9ec4b206315ef44be706d994a0571688cd96291d1ab6c3761df29d00a2ba290a3185e4796bc49891906f86e16da01af3fd52320944b96b60e679ac8e686d4819e97e15e5fe46503c556b4acdd8079624005 '
        converter = LogsToJsonConverter('node1')
        converter.parse(block_header_proof_sent_log.split('\n'))
        assert len(converter.log_content) == 1
        entry = converter.log_content[0]
        assert entry.to_dict() == '{"v": "node1", "t": 1740156803.109, "l": 1, "n": "..nsus/spos/bls/v2", "s": "0", "e": "26", "r": "2649", "sr": "END_ROUND", "m": "step 3: block header proof has been sent", "a": {"PubKeysBitmap": "fd02", "AggregateSignature": "d94a7f18409f9d931a8f96691ea9f2b875281125bf71f9956f5223ccb6fb8d89bfd5f0e8abfeb4975e126c4e250d0988", "proof sender": "f547e115b1ada7cf9b8aeef45ee0d9ec4b206315ef44be706d994a0571688cd96291d1ab6c3761df29d00a2ba290a3185e4796bc49891906f86e16da01af3fd52320944b96b60e679ac8e686d4819e97e15e5fe46503c556b4acdd8079624005"}}'

    def test_trie_statistics_ignore(self):
        'num of nodes = 282160 total size = 33.37 MB num tries by type = dataTrie: 8, mainTrie: 1 num main trie leaves = 191807 max depth main trie = 8'
        converter = LogsToJsonConverter('node1')
        converter.parse(trie_statistics_log.split('\n'))
        assert len(converter.log_content) == 1
        for entry in converter.log_content:
            assert entry.to_dict() == '{"v": "node1", "t": 1740156709.437, "l": 1, "n": "trieStatistics", "s": "0", "e": "26", "r": "2633", "sr": "END_ROUND", "m": "tries statistics", "a": {"num of nodes": "282160", "total size": "33.37MB", "num tries by type": "dataTrie:", "1 num main trie leaves": "191807", "max depth main trie": "8 8, mainTrie:"}}'
            # TODO fix this

    def test_transactions_processed(self):
        # ignore the table, take the transaction processed entry as parameters
        converter = LogsToJsonConverter('node1')
        converter.parse(transactions_processed_table_log.split('\n'))
        for entry in converter.log_content:
            assert entry.to_dict() == '{"v": "node1", "t": 1738912460.175, "l": 1, "n": "process/block", "s": "1", "e": "0", "r": "859", "sr": "START_ROUND", "m": "header hash: c509d85b6a913143f3de41312b862725b60ff5e586f937311b0bb0700801b0f3", "a": {"total txs processed": "0", "block txs processed": "0", "num shards": "3", "shard": "1"}}'

    def test_comma_separated_parameters(self):

        conv = LogsToJsonConverter('node1')
        conv.parse(comma_separated_log.split('\n'))

        assert len(conv.log_content) == 1
        for entry in conv.log_content:
            assert entry.to_dict() == '{"v": "node1", "t": 1740142140.451, "l": 1, "n": "debug/handler", "s": "0", "e": "2", "r": "205", "sr": "END_ROUND", "m": "Requests pending and resolver fails:", "a": {"type": "resolve", "topic": "metachainBlocks_REQUEST", "hash": "00000000000000cb", "numReqIntra": "0", "numReqCross": "0", "numReceived": "3", "numProcessed": "0", "last err": "cannot find header in cache", "query time": "2025-02-21 14:48:44.000"}}'

    def test_comma_separated_parameters_version2(self):
        expected_result = [
            '{"v": "node1", "t": 1740140913.737, "l": 1, "n": "..uffledOutTrigger", "s": "", "e": 0, "r": 0, "sr": "", "m": "initialized chainParametersHolder with the values:", "a": {"enable epoch": "4", " round duration": "6000", " hysteresis": "0.20", " shard consensus group size": "10", " shard min nodes": "10", " meta consensus group size": "10", " meta min nodes": "10", " adaptivity": "false"}}',
            '{"v": "node1", "t": 1740140913.737, "l": 1, "n": "..uffledOutTrigger", "s": "", "e": 0, "r": 0, "sr": "", "m": "initialized chainParametersHolder with the values:", "a": {"enable epoch": "0", " round duration": "6000", " hysteresis": "0.20", " shard consensus group size": "7", " shard min nodes": "10", " meta consensus group size": "10", " meta min nodes": "10", " adaptivity": "false"}}'
        ]

        conv = LogsToJsonConverter('node1')
        conv.parse(initialized_chainParametersHolder_log.split('\n'))

        assert len(conv.log_content) == 2
        for entry in conv.log_content:
            assert entry.to_dict() in expected_result

    def test_transactions_in_pool_entries(self):
        expected_result = [
            '{"v": "node1", "t": 1740156809.211, "l": 1, "n": "process/block", "s": "0", "e": "26", "r": "2650", "sr": "BLOCK", "m": "total txs in pool", "a": {"counts": "Total:217 (51.37 KB); [0]=68 (15.94 KB); [1_0]=92 (21.83 KB); [2_0]=57 (13.60 KB);"}}',
            '{"v": "node1", "t": 1740156809.211, "l": 1, "n": "process/block", "s": "0", "e": "26", "r": "2650", "sr": "BLOCK", "m": "total txs in rewards pool", "a": {"counts": "Total:0 (0 B); [4294967295_0]=0 (0 B);"}}',
            '{"v": "node1", "t": 1740156809.211, "l": 1, "n": "process/block", "s": "0", "e": "26", "r": "2650", "sr": "BLOCK", "m": "total txs in unsigned pool", "a": {"counts": "Total:425 (137.92 KB); [0_1]=57 (9.13 KB); [0_2]=368 (128.79 KB); [1_0]=0 (0 B); [2_0]=0 (0 B);"}}'
        ]

        conv = LogsToJsonConverter('node1')
        conv.parse(total_transactions_in_pool_logs.split('\n'))

        assert len(conv.log_content) == 3
        for entry in conv.log_content:
            assert entry.to_dict() in expected_result

    def test_statistics_entries(self):
        expected_result = [
            '{"v": "node1", "t": 1740156666.706, "l": 1, "n": "common/statistics", "s": "0", "e": "25", "r": "2626", "sr": "END_ROUND", "m": "node statistics", "a": {"uptime": "4h22m33s", "timestamp": "1740163866", "num go": "847", "heap alloc": "1.52 GB", "heap idle": "589.41 MB", "heap inuse": "1.59 GB", "heap sys": "2.17 GB", "heap num objs": "13699793", "sys mem": "2.26 GB", "num GC": "242", "os level stats": "{VmPeak: 22.06 GB, VmSize: 22.06 GB, VmLck: 0 B, VmPin: 0 B, VmHWM: 2.23 GB, VmRSS: 1.89 GB, RssAnon: 1.83 GB, RssFile: 59.21 MB, RssShmem: 4.00 KB, VmData: 2.49 GB, VmStk: 132.00 KB, VmExe: 27.89 MB, VmLib: 10.29 MB, VmPTE: 4.96 MB, VmSwap: 0 B}", "FDs": "284", "num opened files": "283", "num conns": "32", "cpuPercent": "13.32", "cpuLoadAveragePercent": "{avg1min: 0.17, avg5min: 0.26, avg15min: 0.20}", "ioStatsString": "{readCount: 7192811, writeCount: 6761406, readBytes: 4.64 MB, writeBytes: 2.51 GB}", "host network data size in epoch": "{total received: 55.76 MB, total sent: 54.63 MB}"}}',
            '{"v": "node1", "t": 1740156696.711, "l": 1, "n": "common/statistics", "s": "0", "e": "26", "r": "2631", "sr": "END_ROUND", "m": "node statistics", "a": {"uptime": "4h23m3s", "timestamp": "1740163896", "num go": "842", "heap alloc": "1.20 GB", "heap idle": "738.32 MB", "heap inuse": "1.45 GB", "heap sys": "2.17 GB", "heap num objs": "10008737", "sys mem": "2.26 GB", "num GC": "243", "os level stats": "{VmPeak: 22.06 GB, VmSize: 22.06 GB, VmLck: 0 B, VmPin: 0 B, VmHWM: 2.23 GB, VmRSS: 1.94 GB, RssAnon: 1.88 GB, RssFile: 59.21 MB, RssShmem: 4.00 KB, VmData: 2.49 GB, VmStk: 132.00 KB, VmExe: 27.89 MB, VmLib: 10.29 MB, VmPTE: 4.96 MB, VmSwap: 0 B}", "FDs": "282", "num opened files": "281", "num conns": "31", "cpuPercent": "13.34", "cpuLoadAveragePercent": "{avg1min: 0.16, avg5min: 0.25, avg15min: 0.19}", "ioStatsString": "{readCount: 7206533, writeCount: 6777547, readBytes: 4.75 MB, writeBytes: 2.52 GB}", "host network data size in epoch": "{total received: 2.09 MB, total sent: 1.97 MB}"}}'
        ]

        conv = LogsToJsonConverter('node1')
        conv.parse(statistics_entries_log.split('\n'))

        assert len(conv.log_content) == 2

        for entry in conv.log_content:
            assert entry.to_dict() in expected_result

    def test_contract_deployed_log(self):
        expected_entries = [
            '{"v": "node1", "t": 1740140920.086, "l": 1, "n": "..ss/smartcontract", "s": "", "e": 0, "r": 0, "sr": "", "m": "SmartContract deployed", "a": {"owner": "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u", "SC address(es)": "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"}}',
            '{"v": "node1", "t": 1740141797.076, "l": 1, "n": "..ss/smartcontract", "s": "0", "e": "1", "r": "148", "sr": "BLOCK", "m": "SmartContract deployed", "a": {"owner": "erd13utlnu8haswaxcyy0qcz4h36ac00ljwcswh99296ttky26c9squq5kvhnw", "SC address(es)": "erd1qqqqqqqqqqqqqpgq757dnjww7m08velt2udhwkgyzh5dv5z6squqpxgcgs"}}']
        conv = LogsToJsonConverter('node1')
        conv.parse(contract_deployed_log.split('\n'))
        assert len(conv.log_content) == 2

        for entry in conv.log_content:
            assert entry.to_dict() in expected_entries

    def test_special_characters_in_parameters(self):
        expected_result = [
            '{"v": "node1", "t": 1740142739.305, "l": 1, "n": "storage/leveldb", "s": "0", "e": "3", "r": "305", "sr": "END_ROUND", "m": "processLoop - closing the leveldb process loop", "a": {"path": "/home/ubuntu/go/src/github.com/multiversx/mx-chain-go/cmd/node/db/1/Epoch_0/Shard_0/BootstrapData"}}',
            '{"v": "node1", "t": 1740140913.737, "l": 1, "n": "node", "s": "", "e": 0, "r": 0, "sr": "", "m": "read enable epoch for max nodes change", "a": {"epoch": "[{0 48 4} {1 64 2} {3 56 2}]"}}',
            '{"v": "node1", "t": 1740140913.737, "l": 1, "n": "node", "s": "", "e": 0, "r": 0, "sr": "", "m": "read enable epoch for gas schedule directories paths", "a": {"epoch": "[{0 gasScheduleV8.toml}]"}}',
            '{"v": "node1", "t": 1740140913.738, "l": 1, "n": "common/enablers", "s": "", "e": 0, "r": 0, "sr": "", "m": "loaded round config", "a": {"name": "DisableAsyncCallV1", "round": "100", "options": "[]"}}',
            '{"v": "node1", "t": 1740141527.46, "l": 1, "n": "trieStatistics", "s": "0", "e": "1", "r": "103", "sr": "END_ROUND", "m": "migration stats for mainTrie", "a": {"num leaves with not specified version": "771"}}',
            '{"v": "node1", "t": 1740141527.46, "l": 1, "n": "trieStatistics", "s": "0", "e": "1", "r": "103", "sr": "END_ROUND", "m": "migration stats for dataTrie", "a": {"num leaves with not specified version": "55"}}',
            '{"v": "node1", "t": 1740142133.177, "l": 1, "n": "storage/pruning", "s": "0", "e": "2", "r": "204", "sr": "END_ROUND", "m": "PruningStorer.processPersistersToClose", "a": {"epochs to close": "[]", "before process": "[2 1 0]", "after process": "[2 1 0]"}}',
            '{"v": "node1", "t": 1740156709.438, "l": 1, "n": "trieStatistics", "s": "0", "e": "26", "r": "2633", "sr": "END_ROUND", "m": "migration stats for dataTrie", "a": {"num leaves with auto balanced version": "18306", "num leaves with not specified version": "55"}}',
            '{"v": "node1", "t": 1740156709.438, "l": 1, "n": "trieStatistics", "s": "0", "e": "26", "r": "2633", "sr": "END_ROUND", "m": "migration stats for mainTrie", "a": {"num leaves with not specified version": "191807"}}'
        ]
        conv = LogsToJsonConverter('node1')
        conv.parse(special_chars_in_parameters_log.split('\n'))

        assert len(conv.log_content) == 9

        for entry in conv.log_content:
            assert entry.to_dict() in expected_result

    def test_new_sharded_data(self):
        conv = LogsToJsonConverter('node1')
        conv.parse(new_sharded_data_log.split('\n'))

        assert len(conv.log_content) == 7

        for entry in conv.log_content:
            # assert entry.to_dict() in expected_result
            print(entry.to_dict())

    def test_miscelanious(self):
        conv = LogsToJsonConverter('node1')
        conv.parse(misc_no_ascii_chars_log.split('\n'))
        expected_result = [
            '{"v": "node1", "t": 1741683641.52, "l": 2, "n": "node", "s": "", "e": 0, "r": 0, "sr": "", "m": "starting node", "a": {"version": "logs-to-json-conversion-00208db9bb/go1.20.7/linux-amd64/70b02c7576", "pid": "8203"}}',
            '{"v": "node1", "t": 1741683661.022, "l": 1, "n": "node", "s": "", "e": 0, "r": 0, "sr": "", "m": "starting node... executeOneComponentCreationCycle", "a": {}}',
            '{"v": "node1", "t": 1741692751.99, "l": 1, "n": "node", "s": "0", "e": "14", "r": "2839", "sr": "END_ROUND", "m": "starting node... executeOneComponentCreationCycle", "a": {}}',
            '{"v": "node1", "t": 1741692720.918, "l": 2, "n": "main/p2p", "s": "1", "e": "14", "r": "2839", "sr": "END_ROUND", "m": "listening on addresses", "a": {"addr0": "/ip4/10.0.0.143/tcp/37870/p2p/16Uiu2HAmCRqUbbixX67JHPLuH2h6dSqTqt961CdLRXMSzsDXsd23", "addr1": "/ip4/127.0.0.1/tcp/37870/p2p/16Uiu2HAmCRqUbbixX67JHPLuH2h6dSqTqt961CdLRXMSzsDXsd23"}}',
            '{"v": "node1", "t": 1741683641.557, "l": 1, "n": "..rottle/antiflood", "s": "", "e": 0, "r": 0, "sr": "", "m": "SetMaxMessagesForTopic", "a": {"topic": "shardBlocks*", "num messages": "30"}}',
            '{"v": "node1", "t": 1741683661.024, "l": 1, "n": "process/sync", "s": "", "e": 0, "r": 0, "sr": "", "m": "storageBootstrapper.loadBlocks", "a": {"LastHeader": {"shard": "1", "nonce": "1171", "epoch": "5", "hash": "25ff478ad6b9f8ab78f28570e04b93b90931522c4da2927be5b9361ef47bc101"}, "LastCrossNotarizedHeaders": {"shard": "4294967295", "nonce": "1191", "epoch": "0", "hash": "f5e0f02a598bc292a55e10f19b04bda72198fee0cba7c220f2dd21f2dbf365a9"}, "LastSelfNotarizedHeaders": {"shard": "1", "nonce": "1171", "epoch": "5", "hash": "25ff478ad6b9f8ab78f28570e04b93b90931522c4da2927be5b9361ef47bc101"}, "HighestFinalBlockNonce": "1171", "NodesCoordinatorConfigKey": "6839b5f19d9a0544b91f3f141af7212c5e380aef0ec9523012c2a21b01c7a07d220e37973ddd0f1454298e633e5b8b0a", "EpochStartTriggerConfigKey": "31323032"}}'
        ]
        assert len(conv.log_content) == 6

        for entry in conv.log_content:
            assert entry.to_dict() in expected_result
