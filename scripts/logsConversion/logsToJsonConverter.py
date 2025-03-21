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

        # smart contract deployed
        # param_str = param_str.replace('SC address(es)', 'SC addresses')

        # migration stats entries
        if param_str.strip().startswith('stats = ['):
            param_str = param_str.strip().replace('stats = [', '').replace(']', '')

        pattern = r"([\w\s]+)\s*=\s*(\[.*?\])" if param_str.strip().startswith('epochs to close') else r"([\w\s\[\]().,-]+?)\s*=\s*([\w().,-]+)"
        # r"([\w\s\[\].]+?)\s*=\s*([\w().,-]+)"

        matches = re.findall(pattern, param_str)
        remaining_text = param_str

        for k, v in matches:
            target_dict[k.strip()] = v.strip()
            remaining_text = re.sub(re.escape(f"{k.strip()} = {v.strip()}"), "", remaining_text, count=1).strip()

        # unmatched text after parsing all key-value pairs
        if remaining_text:
            if '=' in remaining_text:
                splitted = remaining_text.split('=', 1)
                target_dict[splitted[0].strip()] = splitted[1].strip()
                if remaining_text.count('=') != 1:
                    print('MORE THAN ONE = : ', param_str)
                    print(remaining_text)
                    print(splitted)
                    print()

            else:
                # add it to the last key (useful if the last value contains spaces)
                if target_dict:
                    k = list(target_dict.keys())[-1]
                    target_dict[k] += ' ' + remaining_text.strip()
                else:
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

    def parse(self, log_lines: list[str]):
        print(f'Parsing content of log file {self.input_path} for {self.node_name}')
        pattern = re.compile(
            r'^(?P<log_level>DEBUG|INFO|TRACE|WARN|ERROR)\s*\['         # Log level
            r'(?P<timestamp>[^\]]+)\]\s*'              # Timestamp
            r'\[(?P<module>[^\]]+)\]\s*'               # Logger name
            r'\[(?P<context>[^\]]*)\]\s*'              # context inside the third bracket
            r'(?P<message>.*)$'                        # The rest of the log message
        )

        context_pattern = re.compile(r'(?P<shard>\d+)/(?P<epoch>\d+)/(?P<round>\d+)/\((?P<subround>[^)]+)\)')

        no_space_in_message = []
        first_params = []
        for line in log_lines:
            param_dict = {}
            match = pattern.match(line)

            # normal log line, completely formed, that starts with log level etc
            if match:
                logger_level = match.group('log_level').strip()
                date = match.group('timestamp').strip()
                logger_name = match.group('module').strip()
                context = match.group('context').strip()
                raw_message = match.group('message')  # The rest of the log message

                # parse parameters
                context_match = context_pattern.match(context)
                if context_match:
                    shard, epoch, round, sub_round = context_match.group('shard').strip(), context_match.group('epoch').strip(), context_match.group('round').strip(), context_match.group('subround').strip()
                else:
                    shard, epoch, round, sub_round = '', 0, 0, ''

                # handle message and parameters
                if raw_message and '=' in raw_message:
                    # message & parameters together: take last word before the first '=' as first key
                    if '  ' not in raw_message:
                        parts = raw_message.split(' = ', 1)
                        message, first_label = parts[0].rsplit(" ", 1)
                        self.add_parameters(param_dict, first_label.strip() + ' = ' + parts[1].strip())
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
                        else:
                            self.add_parameters(param_dict, parameters_str if len(parameters_str) > 1 else '')

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

            # parameters line on subsequent entry/entries - add to the last fully formed entry
            elif line and not self.is_block_table(line) and not self.is_trie_statistics(line):

                # comma separated parameters line on subsequent row (<key>: <value>, <key>: <value>)
                if ',' in line and ':' in line:
                    if self.log_content:
                        self.add_parameters_for_coma_separatated_entries(self.log_content[-1].a, line.strip())

                # several comma separated parameters lines on subsequent rows (<key>=<value>, <key>=<value>)
                if ',' in line and '=' in line:
                    if self.log_content:
                        param_dict = {}
                        self.add_parameters_for_coma_separatated_entries(param_dict, line.strip())

                        # if it is not the first parameters row, add a new entry with the same parameters
                        if self.log_content[-1].a:
                            self.log_content.append(self.log_content[-1].copy())
                        self.log_content[-1].a = param_dict

                # regular parameters line on subsequent entry/entries (<key> = <value> <key> = <value>)
                elif self.log_content and '=' in line:
                    self.add_parameters(self.log_content[-1].a, line.strip())

                else:
                    pass
                    # TODO if logging is enabled add this to the log file
                    # potentialy irrelevant line. display to check if it should be handled
                    # elif '=' not in line:
                    #    print('***', line)
        print('no space in message')
        for item in no_space_in_message:
            print(f'{item} => {first_params[no_space_in_message.index(item)]}')

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

    txt = key[:3] if key.startswith('GB ') or key.startswith('MB ') or key.startswith('KB ') else ''
    stats_dict[key.replace('GB ', '').replace('MB ', '').replace('KB ', '')] = value
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
    # log_file = '~/Downloads/perf-deg-andromeda/OVH-P04--Shard-0--4cd5fb6017a3--172.30.40.76--ovh-p04-validator-26/logs/logs/mx-chain-go-2025-02-21-14-28-33.log'
    with open(args.path, "r") as file:
        log = file.readlines()
        converter.parse(log)

    output_file_name = OUTPUT_FOLDER + '/' + f'{args.node_name}_{args.path.split('/')[-1].replace('.log', '.jsonl')}'
    directory = os.path.dirname(output_file_name)
    os.makedirs(directory, exist_ok=True)
    print('Output file name: ', output_file_name)
    with open(output_file_name, 'w') as f:
        for entry in converter.log_content:
            f.write(entry.to_dict())
            f.write('\n')


def validate_file_path(path: str):
    if not os.path.isfile(path):
        raise argparse.ArgumentTypeError(f"File '{path}' does not exist.")
    return path


if __name__ == "__main__":
    main()
