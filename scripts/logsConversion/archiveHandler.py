
import json
import re
import tarfile
import zipfile
from typing import IO, Any

from scripts.logsConversion.logsToJsonConverter import LogsToJsonConverter


class ArchiveHandler:
    def __init__(self, logs_path: str):
        self.logs_path = logs_path

        zip_name_pattern = r'.*/(.*?).zip'
        match = re.match(zip_name_pattern, self.logs_path)
        self.run_name = match.group(1) if match else 'unknown-zip-name'

    # loop through nodes and process logs

    def parse_logs(self, zip_path: str):
        # Open the zip file and process tar.gz files inside it that each correspond to a node

        with zipfile.ZipFile(zip_path, 'r') as zip_file:
            # List all files inside the zip
            file_list = zip_file.namelist()

            for file_name in file_list:
                if file_name.endswith(".tar.gz"):
                    print(f"Processing {file_name}")

                    # Open the tar.gz file as bytes
                    with zip_file.open(file_name) as tar_file_io:
                        rounds_dict = self.process_node_logs(file_name, tar_file_io)

                    with open('report.json', "w") as json_file:
                        json.dump(rounds_dict, json_file, indent=4)

    # loop through log files in a particular node's logs and parse data

    def process_node_logs(self, tar_gz_path: str, tar_gz_contents: IO[bytes]) -> list[dict[str, Any]]:
        with tarfile.open(fileobj=tar_gz_contents, mode='r:gz') as tar:
            # sort logs in chronological order
            sorted_members = sorted(
                tar.getmembers(),
                key=lambda member: member.name
            )

            # process all log files
            for member in sorted_members:
                if member.name.startswith('logs/logs/') and member.name.endswith('.log'):
                    raw_data = tar.extractfile(member)
                    if not raw_data:
                        continue

                    with raw_data as f:
                        # Decode the content to text
                        log_content = f.read().decode('utf-8')
                        print('***' + member.name.split('/')[-1])
                        converter = LogsToJsonConverter(tar_gz_path)
                        converter.parse(log_content.split('\n'))


if __name__ == "__main__":

    zip_pathes = [
        # '/home/mihaela/Downloads/perf-report_mainnet6L_ovh-p03_19-11-2024_16-21-20.zip',
        # '/home/mihaela/Downloads/perf-report_mainnet7L_ovh-p03_19-11-2024_21-38-47.zip',
        # '/home/mihaela/Downloads/perf-report_spica1L_ovh-p03_27-11-2024_22-20-35.zip',
        # '/home/mihaela/Downloads/perf-report_spica2L_ovh-p03_28-11-2024_10-55-21.zip',
        # '/home/mihaela/Downloads/perf-report_spica3L_ovh-p03_28-11-2024_10-55-21.zip',
        # '/home/mihaela/Downloads/perf-report_mempool1L_ovh-p03_04-12-2024_10-55-21.zip',
        # '/home/mihaela/Downloads/perf-report_mempool3L_ovh-p03_06-12-2024_10-55-21.zip',
        # '/home/mihaela/Downloads/perf-report_mempool5L_do-ams_06-12-2024_10-55-21.zip',
        '/home/mihaela/Downloads/perf-report_mempool6L_ovh-p03_06-12-2024_10-55-21.zip',
        # '/home/mihaela/Downloads/perf-report_mempool2L_ovh-p03_28-11-2024_10-55-21.zip'
    ]

    for zip_path in zip_pathes:
        parser = ArchiveHandler(zip_path)
        parser.parse_logs(zip_path)
