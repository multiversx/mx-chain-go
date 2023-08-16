
# Logviewer App

The **MultiversX Logviewer App** exposes the following Command Line Interface:

```
$ logviewer --help

NAME:
   MultiversX Logviewer App - Logviewer application used to communicate with mx-chain-go node to log the message lines
USAGE:
   logviewer [global options]
   
AUTHOR:
   The MultiversX Team <contact@multiversx.com>
   
GLOBAL OPTIONS:
   --address value            Address and port number on which the application will try to connect to the mx-chain-go node (default: "127.0.0.1:8080")
   --log-level level(s)       This flag specifies the logger level(s). It can contain multiple comma-separated value. For example, if set to *:INFO the logs for all packages will have the INFO level. However, if set to *:INFO,api:DEBUG the logs for all packages will have the INFO level, excepting the api package which will receive a DEBUG log level. (default: "*:INFO ")
   --log-save                 Boolean option for enabling log saving. If set, it will automatically save all the logs into a file.
   --working-directory value  The application will store here the logs in a subfolder.
   --use-wss                  Will use wss instead of ws when creating the web socket
   --log-correlation          Boolean option for enabling log correlation elements.
   --log-logger-name          Boolean option for logger name in the logs.
   --help, -h                 show help
   --version, -v              print the version
   

```

