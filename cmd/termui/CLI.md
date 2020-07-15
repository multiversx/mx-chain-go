
# Elrond TermUI CLI

The **Elrond Term UI** exposes the following Command Line Interface:

```
$ termui --help

NAME:
   Elrond Terminal UI App - Terminal UI application used to display metrics from the node
USAGE:
   termui [global options]
   
AUTHOR:
   The Elrond Team <contact@elrond.com>
   
GLOBAL OPTIONS:
   --address value    Address and port number on which the application will try to connect to the elrond-go node (default: "127.0.0.1:8080")
   --log-level value  This flag specifies the logger level (default: "*:INFO ")
   --log-correlation  Will include log correlation elements
   --log-logger-name  Will include logger name
   --interval value   This flag specifies the duration in seconds until new data is fetched from the node (default: 2)
   --use-wss          Will use wss instead of ws when creating the web socket
   --help, -h         show help
   --version, -v      print the version
   

```

