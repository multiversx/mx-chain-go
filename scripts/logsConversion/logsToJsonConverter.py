import argparse
import json
import os
import re
from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from typing import Any

OUTPUT_FOLDER = 'output'
LOGS_DATE_FORMAT = '%Y-%m-%d %H:%M:%S.%f'


def dict_copy(d: dict[str, Any]) -> dict[str, Any]:
    result = {}
    for k in d.keys():
        result[k] = d[k]
    return result


def load_json(file_path: str) -> dict[str, Any]:
    """Load JSON data from a file."""
    with open(file_path) as f:
        return json.load(f)


def normalize_spacing(text):
    return re.sub(r' {2,}', ' ', text)


@dataclass
class LogEntry:
    v: str              # node name
    t: float            # timestamp
    l: int              # log level
    n: str              # logger name
    s: str              # shard
    e: int              # epoch
    r: int              # round
    sr: str             # subround
    m: str              # message
    a: list[str]        # parameters

    def to_dict(self):
        return json.dumps({attr: getattr(self, attr) for attr in vars(self)})

    # for testing purposes - reconstruct the log entry
    def __str__(self):
        parameters = " ".join(f"{k} = {v}" for k, v in self.a.items())
        return f'{LoggingLevel.find_by_value(self.l).name}[{datetime.fromtimestamp(self.t).strftime(LOGS_DATE_FORMAT)[:-3]}] [{self.n}] [{self.s}/{self.e}/{self.r}/({self.sr})] {normalize_spacing(self.m)} {parameters}'

    def copy(self) -> 'LogEntry':
        return LogEntry(**vars(self))


class LoggingLevel (Enum):
    TRACE = 0
    DEBUG = 1
    INFO = 2
    WARN = 3
    ERROR = 4
    NONE = 5

    @staticmethod
    def find(logging_level_name: str):
        return next(item for item in LoggingLevel if item.name == logging_level_name)

    @staticmethod
    def find_by_value(value: int):
        return next(item for item in LoggingLevel if item.value == value)


class LogsToJsonConverter:
    def __init__(self, node_name: str, input_path: str = '', output_path: str = OUTPUT_FOLDER, log_content: list[LogEntry] = None):
        self.node_name = node_name
        self.input_path = input_path
        self.output_path = output_path
        self.log_content = log_content if log_content else []

    def write_to_file(self, output_file_name: str):
        output_file_path = self.output_path + '/' + f'{self.node_name}_{output_file_name}'
        directory = os.path.dirname(output_file_path)
        os.makedirs(directory, exist_ok=True)

        with open(output_file_path, 'w') as f:
            for entry in self.log_content:
                f.write(entry.to_dict())
                f.write('\n')
        return output_file_path

    def is_block_table(self, line: str) -> bool:
        table_strings = ['+-', '| ']
        return line[:2] in table_strings

    def is_trie_statistics(self, line: str) -> bool:
        trie_statistics_entries_types = ['top 10 tries by size', 'top 10 tries by depth', '= address', 'address']
        return self.log_content and self.log_content[-1].m.startswith('tries statistics') and line.strip().startswith(tuple(trie_statistics_entries_types))

    # main add parameters function - also checks for particular cases
    def add_parameters(self, target_dict: dict[str, Any], param_str: str):
        if not param_str:
            return

        # empty scheduled root hash
        if param_str.strip().endswith('scheduled root hash =') or 'scheduled root hash =  num of scheduled txs' in param_str:
            param_str = param_str.replace('scheduled root hash =', 'scheduled root hash = _')

        # transaction in pool entries
        if param_str.strip().startswith('counts = Total:'):
            target_dict['counts'] = param_str.strip().replace('counts = ', '')
            return
        else:
            param_str = param_str.strip().replace(' MB', 'MB').replace(' GB', 'GB').replace(' KB', 'KB').replace(' B', 'B')

        # migration stats entries
        if param_str.strip().startswith('stats = ['):
            param_str = param_str.strip().replace('stats = [', '').replace(']', '')

        pattern = r"([\w\s]+)\s*=\s*(\[.*?\])" if param_str.strip().startswith('epochs to close') else r"([\w\s\[\]().,-/]+?)\s*=\s*(\{.*\}|\S+)"

        matches = re.findall(pattern, param_str)
        remaining_text = param_str

        for k, v in matches:
            value = v.strip()
            target_dict[k.strip()] = (json.loads(value) if value.startswith("{") and '"' in value else value)

            remaining_text = re.sub(re.escape(f"{k.strip()} = {v.strip()}"), "", remaining_text, count=1).strip()

        # unmatched text after parsing all key-value pairs
        if remaining_text:
            if '=' in remaining_text:
                splitted = remaining_text.split('=', 1)
                target_dict[splitted[0].strip()] = splitted[1].strip()
                # unhandled cases
                if remaining_text.count('=') != 1:
                    print('MORE THAN ONE = : ', param_str)
                    print()

            else:
                # add it to the last key (useful if the last value contains spaces)
                if target_dict:
                    k = list(target_dict.keys())[-1]
                    target_dict[k] += ' ' + remaining_text.strip()
                else:
                    # should not happen
                    print('EMPTY PARAM DICT. COULDNT ADD PARAMS', param_str)

    # add parameters for node statistics entries
    def add_parameters_for_node_statistics(self, target_dict: dict[str, Any], param_str: str):
        if not param_str:
            return
        pattern = r"([\w\s]+?)\s*=\s*(\{(?:[^{}]+|(?:\{[^{}]*\}))*\}|\S+)"

        matches = re.findall(pattern, param_str)
        # fix matching
        stats = recursive_post_process_keys_for_statistics('', list(matches), {})
        for key in stats.keys():
            target_dict[key.strip()] = stats[key].strip()

    # add parameters for coma separated entries in subsequent rows
    def add_parameters_for_coma_separatated_entries(self, target_dict: dict[str, Any], param_str: str):
        if not param_str:
            return
        pattern = re.compile(r'(\w+(?:\s+\w+)?): ([^,]+)' if ':' in param_str else r'([\w\s]+)=([\w.\d-]+)')

        for key in dict(pattern.findall(param_str)):
            target_dict[key] = dict(pattern.findall(param_str))[key]

    def add_parameters_for_paranthesis_in_message(self, target_dict: dict[str, Any], param_str: str):
        pattern = r"(?P<key>[\w\s]+)\s*=\s*(?P<value>\[.*?\])"

        matches = re.findall(pattern, param_str)
        for key, value in matches:
            target_dict[key.strip()] = value.strip()

    def add_parameters_for_start_time(self, target_dict: dict[str, Any], param_str: str):
        pattern = r"formatted = (.+?) seconds = (\d+)"
        matches = re.search(pattern, param_str)
        if matches:
            target_dict['formatted'] = matches.group(1).strip()
            target_dict['seconds'] = matches.group(2).strip()

    def add_parameters_for_storage_bootstraper(self, target_dict: dict[str, Any], param_str: str):
        # normalize LastHeader section to also fit a key = [...] format
        text = param_str.strip().replace(' = shard', ' = [shard').replace(' LastCrossNotarizedHeaders', '] LastCrossNotarizedHeaders')
        # Match key-value pairs and list sections
        pattern = re.compile(r'(\w+(?:\s\w+)*)\s*=\s*(?:\[(.*?)\]|(\S+))')

        matches = pattern.findall(text)

        for key, list_content, value in matches:
            if list_content:
                # Parse items inside the brackets as key-value pairs
                items = re.findall(r'(\w+)\s+([\w\d]+)', list_content)
                target_dict[key] = {k: v for k, v in items}
            else:
                target_dict[key] = value

    def parse(self, log_lines: list[str]):
        print(f'Parsing content of log file {self.input_path} for {self.node_name}')
        pattern = re.compile(
            r'^(?P<log_level>DEBUG|INFO|TRACE|WARN|ERROR)\s*\['         # Log level
            r'(?P<timestamp>[^\]]+)\]\s*'              # Timestamp
            r'\[(?P<module>[^\]]+)\]\s*'               # Logger name
            r'\[(?P<context>[^\]]*)\]\s*'              # context inside the third bracket
            r'(?P<message>.*)$'                        # The rest of the log message
        )

        # context_pattern = re.compile(r'(?P<shard>\d+)/(?P<epoch>\d+)/(?P<round>\d+)/\((?P<subround>[^)]+)\)')
        context_pattern = re.compile(r'(?P<shard>\S+)/(?P<epoch>\d+)/(?P<round>\d+)(?:/\((?P<subround>[^\)]+)\))?/?')

        no_space_in_message = []
        first_params = []
        for line in log_lines:
            param_dict = {}
            match = pattern.match(line)
            network_connection_parts = []

            # normal log line, completely formed, that starts with log level etc
            if match:
                logger_level = match.group('log_level').strip()
                date = match.group('timestamp').strip()
                logger_name = match.group('module').strip()
                context = match.group('context').strip()
                raw_message = match.group('message')  # The rest of the log message

                # parse shard, epoch, round, subround
                context_match = context_pattern.match(context)
                if context_match:
                    tmp_subround = context_match.group('subround')
                    shard, epoch, round, sub_round = context_match.group('shard').strip(), context_match.group('epoch').strip(), context_match.group('round').strip(), tmp_subround.strip() if tmp_subround else ''
                else:
                    shard, epoch, round, sub_round = '', 0, 0, ''

                # network connection statistics - histograms are separated from the rest of the entry and handled at a later stage
                if "network connection status" in raw_message:
                    network_connection_parts = re.split(r'validators histogram = |observers histogram = |preferred peers histogram = ', raw_message)
                    raw_message = network_connection_parts[0]

                # handle message and parameters
                if raw_message and ' = ' in raw_message:

                    # vm/mettering consequent rows: no message, parameters might start with '|'
                    if logger_name == 'vm/metering' and (len(raw_message.strip().replace(' = ', ' ').split(' ')) == 2 or raw_message.strip().startswith('|')):
                        message = 'metering state'
                        parameters_str = raw_message.replace('|', '').strip()
                        self.add_parameters(param_dict, parameters_str if len(parameters_str) > 1 else '_')

                    # message & parameters together: take last word before the first '=' as first key
                    elif '  ' not in raw_message:
                        parts = raw_message.split(' = ', 1)
                        if len(parts) != 2:
                            parts[1] = '_'
                        message, first_label = parts[0].rsplit(" ", 1)
                        self.add_parameters(param_dict, first_label.strip() + ' = ' + parts[1].strip())
                        # log the result
                        if parts[0] not in no_space_in_message:
                            no_space_in_message.append(parts[0])
                            first_params.append(param_dict)

                    # message & parameters separated by at least double space
                    else:
                        splitted_message = re.split('  ', raw_message, 1)
                        message = splitted_message[0]
                        parameters_str = splitted_message[1].strip()
                        if 'node statistics' in message:
                            self.add_parameters_for_node_statistics(param_dict, parameters_str if len(parameters_str) > 1 else '')
                        elif message.strip() == 'start time':
                            self.add_parameters_for_start_time(param_dict, parameters_str)
                        elif 'storageBootstrapper.loadBlocks' in message:
                            self.add_parameters_for_storage_bootstraper(param_dict, parameters_str)
                        else:
                            self.add_parameters(param_dict, parameters_str if len(parameters_str) > 1 else '')

                    # network connection statistics - parse the histograms
                    if network_connection_parts:
                        for i, part in enumerate(network_connection_parts[1:]):
                            key = 'validators histogram' if i == 0 else 'observers histogram' if i == 1 else 'preferred peers histogram'
                            param_dict[key] = {}
                            self.add_parameters_for_coma_separatated_entries(param_dict[key], part.strip())

                # no parameters
                else:
                    message = raw_message.strip() if raw_message else ''

                self.log_content.append(LogEntry(
                    v=self.node_name,
                    t=datetime.strptime(date, LOGS_DATE_FORMAT).timestamp(),
                    l=LoggingLevel.find(logger_level).value,
                    n=logger_name,
                    s=shard,
                    e=epoch,
                    r=round,
                    sr=sub_round,
                    m=message.strip(),
                    a=dict_copy(param_dict)
                ))

            # message or parameters on subsequent entry/entries - add to the last fully formed entry
            elif line and not self.is_block_table(line) and not self.is_trie_statistics(line):
                # message in subsequent rows
                if '===' in line:
                    self.log_content[-1].m += line.strip()

                # comma separated parameters line on subsequent row (<key>: <value>, <key>: <value>)
                elif ',' in line and ':' in line:
                    if self.log_content:
                        self.add_parameters_for_coma_separatated_entries(self.log_content[-1].a, line.strip())

                # multiple comma separated parameters lines on subsequent rows (<key>=<value>, <key>=<value>)
                elif ',' in line and '=' in line:
                    if self.log_content:
                        param_dict = {}
                        self.add_parameters_for_coma_separatated_entries(param_dict, line.strip())

                        # if it is not the first parameters row, add a new entry with the same basic values as the last one
                        if self.log_content[-1].a:
                            self.log_content.append(self.log_content[-1].copy())
                        self.log_content[-1].a = param_dict

                # regular parameters line on subsequent entry/entries (<key> = <value> <key> = <value>)
                elif self.log_content and '=' in line:
                    self.add_parameters(self.log_content[-1].a, line.strip())

                else:
                    pass
                    # irrelevant line.

        # write the no_space_in_message to a file for debugging purposes
        with open('no_space.txt', 'w') as f:
            for item in no_space_in_message:
                f.write(f'{item} => {first_params[no_space_in_message.index(item)]}\n')

    @staticmethod
    def from_logs(path: str, node_name: str, output_path: str = OUTPUT_FOLDER):
        result = LogsToJsonConverter(node_name, path, output_path)
        log_content = load_json(path)
        result.parse(log_content)


def recursive_post_process_keys_for_statistics(txt: str, matches: list[tuple[str, str]], stats_dict: dict[str, str]) -> dict[str, str]:
    if not matches:
        return stats_dict

    key, value = matches.pop()
    key = key.strip()
    value = value.strip() + (' ' + txt.strip() if txt else '')

    txt = key[:3] if key.startswith('GB ') or key.startswith('MB ') or key.startswith('KB ') else key[:2] if key.startswith('B ') else ''
    stats_dict[key.replace('GB ', '').replace('MB ', '').replace('KB ', '').replace('B ', '')] = value
    recursive_post_process_keys_for_statistics(txt, matches, stats_dict)
    return {k: stats_dict[k] for k in reversed(list(stats_dict.keys()))}


def main():
    parser = argparse.ArgumentParser(
        description='''
        Runs node log conversion to JSON format. Example script:

            python logsToJsonConverter --node_name=ovh-p03-validator-7 --path=logsPath/mx-chain-go-2024-12-10-10-14-21.log
        ''',
        epilog='\n',
        formatter_class=argparse.RawTextHelpFormatter
    )

    parser.add_argument(
        '--node_name',
        required=True,
        type=str,
        help='Name of the node.'
    )
    parser.add_argument(
        '--path',
        required=True,
        type=validate_file_path,
        help='Path to the node logs file.'
    )

    args = parser.parse_args()

    converter = LogsToJsonConverter(args.node_name)
    with open(args.path, "r") as file:
        log = file.readlines()
        converter.parse(log)
    output_file_name = args.path.split('/')[-1].replace('.log', '.jsonl')
    output_file_path = converter.write_to_file(output_file_name)
    print('Output file name: ', output_file_path)


def validate_file_path(path: str):
    if not os.path.isfile(path):
        raise argparse.ArgumentTypeError(f"File '{path}' does not exist.")
    return path


if __name__ == "__main__":
    main()
