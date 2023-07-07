
# MultiversX TermUI CLI

The **MultiversX Term UI** exposes the following Command Line Interface:

```
$ termui --help

NAME:
   MultiversX Terminal UI App - Terminal UI application used to display metrics from the node
USAGE:
   termui [global options]
   
AUTHOR:
   The MultiversX Team <contact@multiversx.com>
   
GLOBAL OPTIONS:
   --address value       Address and port number on which the application will try to connect to the mx-chain-go node (default: "127.0.0.1:8080")
   --log-level level(s)  This flag specifies the logger level(s). It can contain multiple comma-separated value. For example, if set to *:INFO the logs for all packages will have the INFO level. However, if set to *:INFO,api:DEBUG the logs for all packages will have the INFO level, excepting the api package which will receive a DEBUG log level. (default: "*:INFO ")
   --log-correlation     Boolean option for enabling log correlation elements.
   --log-logger-name     Boolean option for logger name in the logs.
   --interval value      This flag specifies the duration in milliseconds until new data is fetched from the node (default: 1000)
   --use-wss             Will use wss instead of ws when creating the web socket
   --help, -h            show help
   --version, -v         print the version
   

```

