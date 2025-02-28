import re

import pytest

from scripts.logsConversion.logsToJsonConverter import LogsToJsonConverter
from scripts.logsConversion.sample_data import (
    comma_separated_log, common_entries_log, empty_value_log,
    several_words_in_key_log, special_chars_in_parameters_log,
    statistics_entries_log, total_transactions_in_pool_logs,
    transactions_processed_table_log)


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
        for line in converter.log_content:
            assert (line.to_dict() in expected_result)

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

    def test_speial_characters_in_parameters(self):
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
            print(entry.to_dict())
            assert entry.to_dict() in expected_result
