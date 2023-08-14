
# MultiversX SeedNode CLI

The **MultiversX SeedNode** exposes the following Command Line Interface:

```
$ seednode --help

NAME:
   SeedNode CLI App - This is the entry point for starting a new seed node - the app will help bootnodes connect to the network
USAGE:
   seednode [global options]
   
AUTHOR:
   The MultiversX Team <contact@multiversx.com>
   
GLOBAL OPTIONS:
   --port [p2p port]                      The [p2p port] number on which the application will start. Can use single values such as `0, 10230, 15670` or range of ports such as `5000-10000` (default: "10000")
   --rest-api-interface address and port  The interface address and port to which the REST API will attempt to bind. To bind to all available interfaces, set this flag to :8080. If set to `off` then the API won't be available (default: "localhost:8080")
   --log-level level(s)                   This flag specifies the logger level(s). It can contain multiple comma-separated value. For example, if set to *:INFO the logs for all packages will have the INFO level. However, if set to *:INFO,api:DEBUG the logs for all packages will have the INFO level, excepting the api package which will receive a DEBUG log level. (default: "*:INFO ")
   --log-save                             Boolean option for enabling log saving. If set, it will automatically save all the logs into a file.
   --config [path]                        The [path] for the main configuration file. This TOML file contain the main configurations such as the marshalizer type (default: "./config/config.toml")
   --p2p-key-pem-file filepath            The filepath for the PEM file which contains the secret keys for the p2p key. If this is not specified a new key will be generated (internally) by default. (default: "./config/p2pKey.pem")
   --help, -h                             show help
   --version, -v                          print the version
   

```

