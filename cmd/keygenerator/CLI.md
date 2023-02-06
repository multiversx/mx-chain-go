
# Keygenerator CLI

The **Key generation Tool** exposes the following Command Line Interface:

```
$ keygenerator --help

NAME:
   Key generation Tool - This binary will generate a validatorKey.pem and walletKey.pem, each containing private key(s)
USAGE:
   keygenerator [global options]
   
AUTHOR:
   The MultiversX Team <contact@multiversx.com>
   
GLOBAL OPTIONS:
   --num-keys value  How many keys should generate. Example: 1 (default: 1)
   --key-type value  What kind of keys should generate. Available options: validator, wallet, both (default: "validator")
   --help, -h        Show help
   --version, -v     Print the version
   --hex-key-prefix  Value special pattern for the hex key prefix: Example 0f00 (default nopattern). Give even number of hex chars.
   --shard value     Which shard should manage the address: Example: 0 (default -1 - no prefference)
```
