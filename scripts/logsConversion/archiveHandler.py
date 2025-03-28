
import argparse
import json
import re
import tarfile
import zipfile
from typing import IO, Any

from scripts.logsConversion.logsToJsonConverter import OUTPUT_FOLDER, LogsToJsonConverter, validate_file_path


class ArchiveHandler:
    def __init__(self, logs_path: str):
        self.logs_path = logs_path

        zip_name_pattern = r'.*/(.*?).zip'
        match = re.match(zip_name_pattern, self.logs_path)
        self.run_name = match.group(1) if match else 'unknown-zip-name'

    # loop through nodes and process logs

    def parse_logs(self):
        # Open the zip file and process tar.gz files inside it that each correspond to a node

        with zipfile.ZipFile(self.logs_path, 'r') as zip_file:
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
                        member_name = member.name.split('/')[-1]
                        # Decode the content to text
                        log_content = f.read().decode('utf-8')
                        print('***' + member_name)
                        converter = LogsToJsonConverter(node_name=member_name,
                                                        output_path=OUTPUT_FOLDER + f'/{tar_gz_path.replace(".tar.gz", "")}',
                                                        )
                        converter.parse(log_content.split('\n'))
                        converter.write_to_file(member_name.replace('.log', '.jsonl'))


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description='''
        Runs node log conversion to JSON format. Example script:

            python logsToJsonConverter --node_name=ovh-p03-validator-7 --path=logsPath/mx-chain-go-2024-12-10-10-14-21.log
        ''',
        epilog='\n',
        formatter_class=argparse.RawTextHelpFormatter
    )

    parser.add_argument(
        '--path',
        required=True,
        type=validate_file_path,
        help='Path to the run zip file.'
    )

    args = parser.parse_args()

    handler = ArchiveHandler(args.path)
    handler.parse_logs()
    print(f'Output folder: {OUTPUT_FOLDER}' / f'{handler.run_name}')
