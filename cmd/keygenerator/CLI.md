
# Keygenerator CLI

The **Key generation Tool** exposes the following Command Line Interface:

```
$ keygenerator --help

NAME:
   Key generation Tool - This binary will generate a validatorKey.pem and walletKey.pem, each containing private key(s)
USAGE:
   keygenerator [global options]
   
AUTHOR:
   The Elrond Team <contact@elrond.com>
   
GLOBAL OPTIONS:
   --num-keys value  How many keys should generate. Example: 1 (default: 1)
   --key-type value  What king of keys should generate. Available options: validator, wallet, both (default: "validator")
   --help, -h        show help
   --version, -v     print the version
   --hex-key-prefix value special pattern for the hex key prefix: Example 00f00 (default nopattern)
   --shard-id value  Which shard should manage the address: Example: 0 (default -1 - no prefference)
```
