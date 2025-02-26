import re

import pytest
from scripts.logsConversion.sample_data import common_entries_log, empty_value_log, several_words_in_key_log, transactions_processed_table_log
from scripts.logsConversion.logsToJsonConverter import LogsToJsonConverter


def normalize_spacing(log_entry):
    """Replaces multiple spaces with exactly one space."""
    return re.sub(r' {2,}', ' ', log_entry)


class TestLogsParsing:
    def test_general(self):
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

    def test_multiple_lines_entry(self):
        # ignore the table, take the transaction processed entry as parameters
        converter = LogsToJsonConverter('node1')
        converter.parse(transactions_processed_table_log.split('\n'))
        for entry in converter.log_content:
            assert entry.to_dict() == '{"v": "node1", "t": 1738912460.175, "l": 1, "n": "process/block", "s": "1", "e": "0", "r": "859", "sr": "START_ROUND", "m": "header hash: c509d85b6a913143f3de41312b862725b60ff5e586f937311b0bb0700801b0f3", "a": {"total txs processed": "0", "block txs processed": "0", "num shards": "3", "shard": "1"}}'

    def test_colon_separated_key_values_entry(self):
        pass

    def test_transactions_in_pool_entries(self):
        pass

    def test_statistics_entries(self):
        pass
