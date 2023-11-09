
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
   --key-type value  What kind of keys should generate. Available options: validator, wallet, p2p, both, mined-wallet (default: "validator")
   --console-out     Boolean option that will enable printing the generated keys directly on the console
   --no-split        Boolean option that will make each generated key added in the same file
   --shard value     integer option that will make each generated wallet key allocated to the desired shard (affects suffix of the key)
available patterns: -1, [0-2] (default: -1)
   --hex-key-prefix value  only used for special patterns in key. Available options: nopattern, [0-f]+ (default: "nopattern")
   --help, -h              show help
   --version, -v           print the version
   

```

