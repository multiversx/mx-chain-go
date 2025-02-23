import re

import pytest
from sample_data import log1, log2
from logsToJsonConverter import LogsToJsonConverter


def normalize_spacing(log_entry):
    """Replaces multiple spaces with exactly one space."""
    return re.sub(r' {2,}', ' ', log_entry)


class TestLogsParsing:
    def test_general(self):
        unexpected = []
        i = 0
        for log in [log1, log2]:
            parser = LogsToJsonConverter('node1')
            parser.parse(log)

            for log_entry in re.split('\n', log):
                if not log_entry:
                    continue
                normalized_entry = normalize_spacing(log_entry).strip()
                if parser.log_content and normalized_entry != str(parser.log_content[i]).strip():
                    unexpected.append(f"{normalized_entry}\n{str(parser.log_content[i]).strip()}\n")
                i += 1

        if unexpected:
            pytest.fail("Unexpected log entries found:\n\n" + "\n".join(unexpected))
