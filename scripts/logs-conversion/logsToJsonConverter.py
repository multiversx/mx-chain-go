'''
{"t":1733818556457877716,"l":1,"n":"process/sync","s":"1","e":0,"r":872859,"sr":"(START_ROUND)","m":"computeNodeState","a":["probableHighestNonce","871289","currentBlockNonce","870772","boot.hasLastBlock","false"]}
{"t":1733818556457950125,"l":1,"n":"process/sync","s":"1","e":0,"r":872859,"sr":"(START_ROUND)","m":"computeNodeState","a":["isNodeStateCalculated","true","isNodeSynchronized","false"]}
{"t":1733818556457926413,"l":1,"n":"process/track","s":"1","e":0,"r":872859,"sr":"(START_ROUND)","m":"received shard header from network in block tracker","a":["shard","1","epoch","726","round","872358","nonce","870788","hash","cb527848d1e521e4de8f426a201f12023cbef25dab925e00b63337b95f74317a"]}
{"t":1733818556457984132,"l":1,"n":"statusHandler","s":"1","e":0,"r":872859,"sr":"(START_ROUND)","m":"processStatusHandler.SetBusy","a":["reason","shardProcessor.ProcessBlock"]}
{"t":1733818556457989706,"l":1,"n":"process/sync","s":"1","e":0,"r":872859,"sr":"(START_ROUND)","m":"forkDetector.AddHeader","a":["round","872358","nonce","870788","hash","cb527848d1e521e4de8f426a201f12023cbef25dab925e00b63337b95f74317a","state","1","probable highest nonce","871289","last checkpoint nonce","870772","final checkpoint nonce","870770"]}
Unde:
t - timestamp
l - logLevel
n - loggerName
...
'''

from dataclasses import dataclass
from datetime import datetime
from typing import Any
import re
import json
# from sample_data import log1
from sample_data import log2

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
    v: str
    t: float
    l: int
    n: str
    s: str
    e: int
    r: int
    sr: str
    m: str
    a: list[str]

    def to_dict(self):
        return json.dumps({attr: getattr(self, attr) for attr in vars(self)})

    def __str__(self):
        parameters = "".join(f"{k} = {v}" for k, v in self.a.items())
        return f'{'DEBUG' if self.l == 1 else 'INFO'}[{datetime.fromtimestamp(self.t).strftime(LOGS_DATE_FORMAT)[:-3]}] [{self.n}] [{self.s}/{self.e}/{self.r}/({self.sr})] {normalize_spacing(self.m)} {parameters}'


class LogsToJsonConverter:
    def __init__(self, node_name: str, input_path: str = '', output_path: str = OUTPUT_FOLDER, log_content: list[LogEntry] = []):
        self.node_name = node_name
        self.input_path = input_path
        self.output_path = output_path
        self.log_content = log_content

    def add_parameters(self, target_dict: dict[str, Any], param_str: str):
        if not param_str:
            return
        matches = re.findall(r"([\w\s\[\].]+?)\s*=\s*([\w.]+)", param_str)
        for k, v in matches:
            target_dict[k] = v

    def parse(self, log_lines: str):
        print(f'Parsing content of log file {self.input_path} for {self.node_name}')
        pattern = re.compile(
            r'^(DEBUG|INFO)'               # Log level (DEBUG/INFO)
            r'\[(.*?)\]'                   # Timestamp inside []
            r'\s*\[(.*?)\]'                # Next [] content
            r'\s*\[(\d+)/(\d+)/(\d+)/\((.*?)\)\]'  # Extract values inside third []
            r'\s*(.*)'                      # The rest of the message
        )

        for line in re.split('\n', log_lines):

            param_dict = {}
            match = pattern.match(line)
            if match:
                logger_level = match.group(1).strip()
                date = match.group(2).strip()
                logger_name = match.group(3).strip()
                shard, epoch, round, sub_round = match.group(4).strip(), match.group(5).strip(), match.group(6).strip(), match.group(7).strip()
                raw_message = match.group(8)
                if '=' in raw_message:
                    if '  ' not in raw_message:
                        parts = raw_message.split(' = ', 1)
                        message, first_label = parts[0].rsplit(" ", 1)
                        self.add_parameters(param_dict, first_label + ' = ' + parts[1])
                        print(param_dict)
                    else:
                        splitted_message = re.split('  ', raw_message, 1)
                        message = splitted_message[0]
                        self.add_parameters(param_dict, splitted_message[1].strip() if len(splitted_message) > 1 else '')
                else:
                    message = raw_message.strip()
                print(f'***|{message}|***')
                self.log_content.append(LogEntry(
                    v=self.node_name,
                    t=datetime.strptime(date, LOGS_DATE_FORMAT).timestamp(),
                    l=1 if logger_level == 'DEBUG' else 2 if logger_level == 'INFO' else 0,
                    n=logger_name,
                    s=shard,
                    e=epoch,
                    r=round,
                    sr=sub_round,
                    m=message.strip(),
                    a=dict_copy(param_dict)
                ))
            elif line:
                print(line)
                self.add_parameters(self.log_content[-1].a, line.strip())

        # for entry in self.log_content:
        #    print(entry.to_dict())

    @staticmethod
    def from_logs(path: str, node_name: str, output_path: str = OUTPUT_FOLDER):
        result = LogsToJsonConverter(node_name, path, output_path)
        log_content = load_json(path)
        result.parse(log_content)


if __name__ == "__main__":
    converter = LogsToJsonConverter('ovh-p03-validator-7')
    converter.parse(log2)

'''
{{"t":1733818514272014721,"l":1,"n":"consensus/chronology","s":"1","e":0,"r":0,"sr":"","m":"2024-12-10 10:15:14.293489691  ################################## ROUND 872852 BEGINS (1733818512) ##################################"}
{"t":1733818514272094361,"l":1,"n":"consensus/chronology","s":"1","e":0,"r":872852,"sr":"","m":"2024-12-10 10:15:14.293578283  ................................... SUBROUND (START_ROUND) BEGINS ..................................."}
{"t":1733818514272738491,"l":1,"n":"node","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"creating node structure"}
{"t":1733818514272797396,"l":1,"n":"heartbeat/sender","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"starting peer shard sender's goroutine"}
{"t":1733818514272784663,"l":1,"n":"heartbeat/sender","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"sending heartbeat message","a":["key","66acb1529e4511ebdcb68475f9f0cbbb1dabe49d9ac046a809050b0dc8e883d5f29f592577d6edbf3fb7bc3a2a7020170c92dd5515ae549f3bac1c1855b4da80d873236b2ff8e245e6440555d8a4ba4305d4a0b5dd035724631b9c9e5b87a613"]}
{"t":1733818514272879977,"l":1,"n":"heartbeat/sender","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"heartbeat message sent","a":["next send will be in","5m9.404678601s"]}
{"t":1733818514273270482,"l":2,"n":"node","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"consensus group","a":["size","63"]}
{"t":1733818514273325275,"l":1,"n":"node","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"updating the API service after creating the node facade"}
{"t":1733818514273333026,"l":1,"n":"node","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"creating api resolver structure"}
{"t":1733818514273361792,"l":1,"n":"trie","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"trie storage manager options","a":["trie pruning status","false","trie snapshot status","true"]}
{"t":1733818514273960387,"l":1,"n":"process/smartcontract/builtInFunctions","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"createBuiltInFunctionsFactory","a":["shardId","1","crawlerAllowedAddress","fd077340be20ee7f1e529207e1132c6e3e35ecc9908cf55cd93e8100230f6101"]}
{"t":1733818514274184094,"l":1,"n":"storage","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"MonitorNewCache","a":["name","SmartContractDataPool","capacity","200.00 MB","cumulated","5.30 GB"]}
{"t":1733818514274241296,"l":1,"n":"factory","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"apiResolver: enable epoch for sc deploy","a":["epoch","1"]}
{"t":1733818514274250988,"l":1,"n":"factory","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"apiResolver: enable epoch for ahead of time gas usage","a":["epoch","1"]}
{"t":1733818514274262380,"l":1,"n":"factory","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"apiResolver: enable epoch for repair callback","a":["epoch","1"]}
{"t":1733818514274281804,"l":1,"n":"factory","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"maximum gas per VM Query","a":["value","1500000000"]}
{"t":1733818514274294522,"l":1,"n":"vmContainerFactory","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"createInProcessWasmVMByVersion","a":["version","{1 v1.5}"]}
{"t":1733818514274305248,"l":2,"n":"vmContainerFactory","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"VM 1.5 created"}
{"t":1733818514275739896,"l":1,"n":"process/sync","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"computeNodeState","a":["probableHighestNonce","870354","currentBlockNonce","870354","boot.hasLastBlock","true"]}
{"t":1733818514275823770,"l":1,"n":"process/sync","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"node has changed its synchronized state","a":["state","true"]}
{"t":1733818514275851779,"l":1,"n":"process/sync","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"computeNodeState","a":["isNodeStateCalculated","true","isNodeSynchronized","true"]}
{"t":1733818514275892528,"l":1,"n":"consensus/spos/bls","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"random source for the next consensus group","a":["rand","076dfe5dd9d7c0480a71956aa01852e5508403ad8a53f4e45b3e906701ff38a6e391f68847e8697d848d4972b83e0717"]}
{"t":1733818514275970944,"l":1,"n":"sharding/nodesCoordinator","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"computeValidatorsGroup","a":["randomness","3837323835322d076dfe5dd9d7c0480a71956aa01852e5508403ad8a53f4e45b3e906701ff38a6e391f68847e8697d848d4972b83e0717","consensus size","63","eligible list length","400","epoch","725","round","872852","shardID","1"]}
{"t":1733818514276175394,"l":1,"n":"consensus/spos/bls","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"step 0: preparing the round","a":["leader","e501505cac55","messsage",""]}
{"t":1733818514276224692,"l":1,"n":"consensus/spos/bls","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"not in consensus group"}
{"t":1733818514276652613,"l":1,"n":"process/smartcontract/builtInFunctions","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"createBuiltInFunctionsFactory","a":["shardId","1","crawlerAllowedAddress","fd077340be20ee7f1e529207e1132c6e3e35ecc9908cf55cd93e8100230f6101"]}
{"t":1733818514276833300,"l":1,"n":"node","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"creating multiversx node facade"}
{"t":1733818514276850285,"l":1,"n":"api/gin","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"upgrading facade for gin API group","a":["group name","address"]}
{"t":1733818514276860866,"l":1,"n":"api/gin","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"upgrading facade for gin API group","a":["group name","hardfork"]}
{"t":1733818514276869474,"l":1,"n":"api/gin","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"upgrading facade for gin API group","a":["group name","proof"]}
{"t":1733818514276878321,"l":1,"n":"api/gin","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"upgrading facade for gin API group","a":["group name","vm-values"]}
{"t":1733818514276886293,"l":1,"n":"api/gin","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"upgrading facade for gin API group","a":["group name","validator"]}
{"t":1733818514276899037,"l":1,"n":"api/gin","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"upgrading facade for gin API group","a":["group name","block"]}
{"t":1733818514276907089,"l":1,"n":"api/gin","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"upgrading facade for gin API group","a":["group name","internal"]}
{"t":1733818514276915936,"l":1,"n":"api/gin","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"upgrading facade for gin API group","a":["group name","network"]}
{"t":1733818514276925116,"l":1,"n":"api/gin","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"upgrading facade for gin API group","a":["group name","node"]}
{"t":1733818514276933703,"l":1,"n":"api/gin","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"upgrading facade for gin API group","a":["group name","transaction"]}
{"t":1733818514276942559,"l":1,"n":"node","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"updated node facade"}
{"t":1733818514276949113,"l":2,"n":"node","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"application is now running"}
{"t":1733818514277334161,"l":1,"n":"consensus/chronology","s":"1","e":0,"r":872852,"sr":"(START_ROUND)","m":"2024-12-10 10:15:14.298815203  ...................................... SUBROUND (BLOCK) BEGINS ......................................"}
{"t":1733818514744202989,"l":1,"n":"consensus/spos","s":"1","e":0,"r":872852,"sr":"(BLOCK)","m":"received proposed block","a":["from","b3646fa12fce","header hash","3deca460d4bd48a3cf6df6666ef12c56a64cd08ad8ae95aebae3d76bd1ea0460","epoch","726","round","872852","nonce","871282","prev hash","8e6159a634883ffeeb3ae94af1816cb1740ff9ad4ce93f66d78465e8a4abd5ef","nbTxs","0","val stats root hash",""]}
{"t":1733818514744334034,"l":1,"n":"sharding/nodesCoordinator","s":"1","e":0,"r":872852,"sr":"(BLOCK)","m":"computeValidatorsGroup","a":["randomness","3837323835322d290e9e934a9d6e7fa8984e303a99c6d2c5eeec71cbb5fdf87336acd570341b2197db1b55219b4dcea60537e34cf24018","consensus size","63","eligible list length","400","epoch","726","round","872852","shardID","1"]}
{"t":1733818514746078693,"l":1,"n":"process/sync","s":"1","e":0,"r":872852,"sr":"(BLOCK)","m":"forkDetector.setHighestNonceReceived","a":["highest nonce received","871282"]}
{"t":1733818514746152697,"l":1,"n":"process/rating/peerhonesty","s":"1","e":0,"r":872852,"sr":"(BLOCK)","m":"p2pPeerHonesty.ChangeScore decrease","a":["pk","b3646fa12fce","current","0.00","change","-4.00"]}
{"t":1733818514746189389,"l":1,"n":"consensus/spos/bls","s":"1","e":0,"r":872852,"sr":"(BLOCK)","m":"time measurements of receivedBlockBodyAndHeader","a":["receivedBlockBodyAndHeader","0.0000s"]}
{"t":1733818517679378027,"l":1,"n":"consensus/spos","s":"1","e":0,"r":872852,"sr":"(BLOCK)","m":"extend function is called","a":["subround","(BLOCK)"]}
{"t":1733818517679481027,"l":1,"n":"consensus/spos","s":"1","e":0,"r":872852,"sr":"(BLOCK)","m":"proposed header with signatures","a":["hash","3deca460d4bd48a3cf6df6666ef12c56a64cd08ad8ae95aebae3d76bd1ea0460","sigs num","43","round","872852"]}
{"t":1733818517679617162,"l":1,"n":"process/block/preprocess","s":"1","e":0,"r":872852,"sr":"(BLOCK)","m":"scheduledTxsExecution.RollBackToBlock","a":["header hash","1a01d067bf4f4fb56b5c813b18424927efaeba2d52c234f8a7042b7bbe6ab676","scheduled root hash","6936658c3db69a32712e6758881117814cd1e9280c4a3fce455c3fd83780384c","num of scheduled mbs","0","num of scheduled intermediate txs","0","accumulatedFees","0","developerFees","0","gasProvided","0","gasPenalized","0","gasRefunded","0"]}
{"t":1733818517679676903,"l":1,"n":"process/block/preprocess","s":"1","e":0,"r":872852,"sr":"(BLOCK)","m":"scheduledTxsExecution.SetScheduledInfo","a":["scheduled root hash","6936658c3db69a32712e6758881117814cd1e9280c4a3fce455c3fd83780384c","num of scheduled mbs","0","num of scheduled intermediate txs","0","accumulatedFees","0","developerFees","0","gasProvided","0","gasPenalized","0","gasRefunded","0"]}
{"t":1733818517679721251,"l":1,"n":"consensus/spos","s":"1","e":0,"r":872852,"sr":"(BLOCK)","m":"current block is reverted"}
{"t":1733818517979454914,"l":1,"n":"consensus/chronology","s":"1","e":0,"r":872852,"sr":"(BLOCK)","m":"2024-12-10 10:15:18.000929998  ################################## ROUND 872853 BEGINS (1733818518) ##################################"}
{"t":1733818517979549442,"l":1,"n":"consensus/chronology","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"2024-12-10 10:15:18.001030711  ................................... SUBROUND (START_ROUND) BEGINS ..................................."}
{"t":1733818517981895562,"l":1,"n":"process/sync","s":"1","e":0,"r":872853,"sr":"(START_ROUND)","m":"computeNodeState","a":["probableHighestNonce","870354","currentBlockNonce","870354","boot.hasLastBlock","true"]}
{"t":1733818517982011664,"l":1,"n":"process/sync","s":"1","e":0,"r":872853,"sr":"(START_ROUND)","m":"computeNodeState","a":["isNodeStateCalculated","true","isNodeSynchronized","true"]}
{"t":1733818517982065797,"l":1,"n":"consensus/spos/bls","s":"1","e":0,"r":872853,"sr":"(START_ROUND)","m":"random source for the next consensus group","a":["rand","076dfe5dd9d7c0480a71956aa01852e5508403ad8a53f4e45b3e906701ff38a6e391f68847e8697d848d4972b83e0717"]}
{"t":1733818517982106365,"l":1,"n":"sharding/nodesCoordinator","s":"1","e":0,"r":872853,"sr":"(START_ROUND)","m":"computeValidatorsGroup","a":["randomness","3837323835332d076dfe5dd9d7c0480a71956aa01852e5508403ad8a53f4e45b3e906701ff38a6e391f68847e8697d848d4972b83e0717","consensus size","63","eligible list length","400","epoch","725","round","872853","shardID","1"]}
{"t":1733818517982378183,"l":1,"n":"consensus/spos/bls","s":"1","e":0,"r":872853,"sr":"(START_ROUND)","m":"step 0: preparing the round","a":["leader","18f74714691a","messsage",""]}
{"t":1733818517982423558,"l":1,"n":"consensus/spos/bls","s":"1","e":0,"r":872853,"sr":"(START_ROUND)","m":"not in consensus group"}
{"t":1733818517983598539,"l":1,"n":"consensus/chronology","s":"1","e":0,"r":872853,"sr":"(START_ROUND)","m":"2024-12-10 10:15:18.005075467  ...................................... SUBROUND (BLOCK) BEGINS ......................................"}
{"t":1733818518014091200,"l":1,"n":"consensus/spos","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"received proposed block","a":["from","1a9bf0a93335","header hash","0230564cb3654691fbafdfdc1c02bd22f28e27cac4fdd86fb47ab88fc61ac706","epoch","726","round","872853","nonce","871283","prev hash","3fdda7ff3a0537134a6e59130807d52d02903ccdc844229bc80965a3b2f7171b","nbTxs","0","val stats root hash",""]}
{"t":1733818518014255347,"l":1,"n":"sharding/nodesCoordinator","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"computeValidatorsGroup","a":["randomness","3837323835332d04c65acbe6d99a01ea5ac0feb6d2d9a2402a8722355605a88bea878eda1795b7703e39424d46551b6a9d187ff2b64898","consensus size","63","eligible list length","400","epoch","726","round","872853","shardID","1"]}
{"t":1733818518018312969,"l":1,"n":"process/sync","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"forkDetector.setHighestNonceReceived","a":["highest nonce received","871283"]}
{"t":1733818518018451573,"l":1,"n":"process/rating/peerhonesty","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"p2pPeerHonesty.ChangeScore decrease","a":["pk","1a9bf0a93335","current","0.00","change","-4.00"]}
{"t":1733818518018574941,"l":1,"n":"consensus/spos/bls","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"time measurements of receivedBlockBodyAndHeader","a":["receivedBlockBodyAndHeader","0.0001s"]}
{"t":1733818518292757744,"l":1,"n":"process/sync","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"forkDetector.AddHeader","a":["round","872853","nonce","871283","hash","bc021e86d2f53a7bb7edf4e6dd815ce61421ca5e00cdf24eddbc517fdd1ca131","state","0","probable highest nonce","871283","last checkpoint nonce","870354","final checkpoint nonce","870354"]}
{"t":1733818518292725449,"l":1,"n":"process/track","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"received shard header from network in block tracker","a":["shard","1","epoch","726","round","872853","nonce","871283","hash","bc021e86d2f53a7bb7edf4e6dd815ce61421ca5e00cdf24eddbc517fdd1ca131"]}
{"t":1733818519123206627,"l":1,"n":"sharding/nodesCoordinator","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"computeValidatorsGroup","a":["randomness","3837323835332d5466e08bcc03c9396cf5b43f68b837ea35d486ec641b20fa8fe4de778412d228a0388e69f2d1da42c13d096a2137e307","consensus size","400","eligible list length","400","epoch","726","round","872853","shardID","4294967295"]}
{"t":1733818519275068425,"l":1,"n":"dataretriever/requesthandlers","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"requesting peer authentication messages from network","a":["topic","peerAuthentication","shard","1","num hashes","500","epoch","726"]}
{"t":1733818521602396456,"l":1,"n":"main/p2p","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"network connection status","a":["network","main","known peers","207","connected peers","70","intra shard validators","0","intra shard observers","1","cross shard validators","1","cross shard observers","1","unknown","66","seeders","1","current shard","1","validators histogram","meta: 1","observers histogram","shard 1: 1, shard 2: 1","preferred peers histogram",""]}
{"t":1733818521602463664,"l":1,"n":"main/p2p","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"network connection metrics","a":["network","main","connections/s","3","disconnections/s","3","connections","65","disconnections","67","time","20s"]}
{"t":1733818521607992079,"l":1,"n":"common/statistics","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"node statistics","a":["uptime","1m0s","timestamp","1733818521","num go","1270","heap alloc","448.38 MB","heap idle","459.22 MB","heap inuse","465.16 MB","heap sys","924.38 MB","heap num objs","979921","sys mem","967.57 MB","num GC","29","os level stats","{VmPeak: 2.64 GB, VmSize: 2.64 GB, VmLck: 0 B, VmPin: 0 B, VmHWM: 664.98 MB, VmRSS: 517.01 MB, RssAnon: 464.39 MB, RssFile: 52.62 MB, RssShmem: 0 B, VmData: 1.09 GB, VmStk: 136.00 KB, VmExe: 27.32 MB, VmLib: 9.19 MB, VmPTE: 2.01 MB, VmSwap: 0 B}","FDs","366","num opened files","365","num conns","90","cpuPercent","25.94","cpuLoadAveragePercent","{avg1min: 1.18, avg5min: 1.16, avg15min: 1.37}","ioStatsString","{readCount: 39210, writeCount: 103539, readBytes: 6.41 MB, writeBytes: 101.64 MB}","host network data size in epoch","{total received: 497.42 KB, total sent: 289.00 KB}"]}
{"t":1733818523678849730,"l":1,"n":"consensus/spos","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"extend function is called","a":["subround","(BLOCK)"]}
{"t":1733818523678962148,"l":1,"n":"consensus/spos","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"proposed header with signatures","a":["hash","0230564cb3654691fbafdfdc1c02bd22f28e27cac4fdd86fb47ab88fc61ac706","sigs num","44","round","872853"]}
{"t":1733818523679163464,"l":1,"n":"process/block/preprocess","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"scheduledTxsExecution.RollBackToBlock","a":["header hash","1a01d067bf4f4fb56b5c813b18424927efaeba2d52c234f8a7042b7bbe6ab676","scheduled root hash","6936658c3db69a32712e6758881117814cd1e9280c4a3fce455c3fd83780384c","num of scheduled mbs","0","num of scheduled intermediate txs","0","accumulatedFees","0","developerFees","0","gasProvided","0","gasPenalized","0","gasRefunded","0"]}
{"t":1733818523679274755,"l":1,"n":"process/block/preprocess","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"scheduledTxsExecution.SetScheduledInfo","a":["scheduled root hash","6936658c3db69a32712e6758881117814cd1e9280c4a3fce455c3fd83780384c","num of scheduled mbs","0","num of scheduled intermediate txs","0","accumulatedFees","0","developerFees","0","gasProvided","0","gasPenalized","0","gasRefunded","0"]}
{"t":1733818523679346921,"l":1,"n":"consensus/spos","s":"1","e":0,"r":872853,"sr":"(BLOCK)","m":"current block is reverted"}
'''
