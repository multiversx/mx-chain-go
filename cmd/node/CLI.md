
# Node CLI

The **Elrond Node** exposes the following Command Line Interface:

```
$ node --help

NAME:
   Elrond Node CLI App - This is the entry point for starting a new Elrond node - the app will start after the genesis timestamp
USAGE:
   node [global options]
   
AUTHOR:
   The Elrond Team <contact@elrond.com>
   
GLOBAL OPTIONS:
   --genesis-file [path]                  The [path] for the genesis file. This JSON file contains initial data to bootstrap from, such as initial balances for accounts. (default: "./config/genesis.json")
   --smart-contracts-file [path]          The [path] for the initial smart contracts file. This JSON file contains data used to deploy initial smart contracts such as delegation smart contracts (default: "./config/genesisSmartContracts.json")
   --nodes-setup-file [path]              The [path] for the nodes setup. This JSON file contains initial nodes info, such as consensus group size, round duration, validators public keys and so on. (default: "./config/nodesSetup.json")
   --config [path]                        The [path] for the main configuration file. This TOML file contain the main configurations such as storage setups, epoch duration and so on. (default: "./config/config.toml")
   --config-api [path]                    The [path] for the api configuration file. This TOML file contains all available routes for Rest API and options to enable or disable them. (default: "./config/api.toml")
   --config-economics [path]              The [path] for the economics configuration file. This TOML file contains economics configurations such as minimum gas price for a transactions and so on. (default: "./config/economics.toml")
   --config-systemSmartContracts [path]   The [path] for the system smart contracts configuration file. (default: "./config/systemSmartContractsConfig.toml")
   --config-ratings value                 The ratings configuration file to load (default: "./config/ratings.toml")
   --config-preferences [path]            The [path] for the preferences configuration file. This TOML file contains preferences configurations, such as the node display name or the shard to start in when starting as observer (default: "./config/prefs.toml")
   --config-external [path]               The [path] for the external configuration file. This TOML file contains external configurations such as ElasticSearch's URL and login information (default: "./config/external.toml")
   --p2p-config [path]                    The [path] for the p2p configuration file. This TOML file contains peer-to-peer configurations such as port, target peer count or KadDHT settings (default: "./config/p2p.toml")
   --gas-costs-config [path]              The [path] for the gas costs configuration file. This TOML file contains gas costs used in SmartContract execution (default: "./config/gasSchedule.toml")
   --sk-index value                       The index in the PEM file of the private key to be used by the node. (default: 0)
   --validator-key-pem-file filepath      The filepath for the PEM file which contains the secret keys for the validator key. (default: "./config/validatorKey.pem")
   --port [p2p port]                      The [p2p port] number on which the application will start. Can use single values such as `0, 10230, 15670` or range of ports such as `5000-10000` (default: "0")
   --profile-mode                         Boolean option for enabling the profiling mode. If set, the /debug/pprof routes will be available on the node for profiling the application.
   --use-health-service                   Boolean option for enabling the health service.
   --storage-cleanup                      Boolean option for starting the node with clean storage. If set, the Node will empty its storage before starting, otherwise it will start from the last state stored on disk..
   --gops-enable                          Boolean option for enabling gops over the process. If set, stack can be viewed by calling 'gops stack <pid>'.
   --display-name value                   The user-friendly name for the node, appearing in the public monitoring tools. Will override the name set in the preferences TOML file.
   --keybase-identity value               The keybase's identity. If set, will override the one set in the preferences TOML file.
   --rest-api-interface address and port  The interface address and port to which the REST API will attempt to bind. To bind to all available interfaces, set this flag to :8080 (default: "localhost:8080")
   --rest-api-debug                       Boolean option for starting the Rest API in debug mode.
   --disable-ansi-color                   Boolean option for disabling ANSI colors in the logging system.
   --log-level level(s)                   This flag specifies the logger level(s). It can contain multiple comma-separated value. For example, if set to *:INFO the logs for all packages will have the INFO level. However, if set to *:INFO,api:DEBUG the logs for all packages will have the INFO level, excepting the api package which will receive a DEBUG log level. (default: "*:INFO ")
   --log-save                             Boolean option for enabling log saving. If set, it will automatically save all the logs into a file.
   --log-correlation                      Boolean option for enabling log correlation elements.
   --log-logger-name                      Boolean option for logger name in the logs.
   --use-log-view                         Boolean option for enabling the simple node's interface. If set, the node will not enable the user-friendly terminal view of the node.
   --bootstrap-round-index index          This flag specifies the round index from which node should bootstrap from storage. (default: 18446744073709551615)
   --working-directory directory          This flag specifies the directory where the node will store databases, logs and statistics.
   --destination-shard-as-observer value  This flag specifies the shard to start in when running as an observer. It will override the configuration set in the preferences TOML config file.
   --keep-old-epochs-data                 Boolean option for enabling a node to keep old epochs data. If set, the node won't remove any database and will have a full history over epochs.
   --num-epochs-to-keep value             This flag represents the number of epochs which will kept in the databases. It is relevant only if the full archive flag is not set. (default: 2)
   --num-active-persisters value          This flag represents the number of databases (1 database = 1 epoch) which are kept open at a moment. It is relevant even if the node is full archive or not. (default: 2)
   --start-in-epoch                       Boolean option for enabling a node the fast bootstrap mechanism from the network.Should be enabled if data is not available in local disk.
   --help, -h                             show help
   --version, -v                          print the version
   

```

