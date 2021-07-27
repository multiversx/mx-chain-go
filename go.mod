module github.com/ElrondNetwork/elrond-go

go 1.13

require (
	github.com/ElrondNetwork/arwen-wasm-vm/v1_2 v1.2.28
	github.com/ElrondNetwork/arwen-wasm-vm/v1_3 v1.3.27
	github.com/ElrondNetwork/arwen-wasm-vm/v1_4 v1.4.3
	github.com/ElrondNetwork/concurrent-map v0.1.3
	github.com/ElrondNetwork/elastic-indexer-go v1.0.8
	github.com/ElrondNetwork/elrond-go-core v1.0.1-0.20210723124835-0f85becd722a
	github.com/ElrondNetwork/elrond-go-logger v1.0.5
	github.com/ElrondNetwork/elrond-vm-common v1.1.1-0.20210722121034-4894ec636075
	github.com/beevik/ntp v0.3.0
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/davecgh/go-spew v1.1.1
	github.com/elastic/go-elasticsearch/v7 v7.12.0
	github.com/gin-contrib/cors v0.0.0-20190301062745-f9e10995c85a
	github.com/gin-contrib/pprof v1.3.0
	github.com/gin-gonic/gin v1.7.2
	github.com/gizak/termui/v3 v3.1.0
	github.com/gogo/protobuf v1.3.2
	github.com/google/gops v0.3.18
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/golang-lru v0.5.4
	github.com/herumi/bls-go-binary v1.0.0
	github.com/ipfs/go-log v1.0.5
	github.com/jbenet/goprocess v0.1.4
	github.com/libp2p/go-libp2p v0.14.3
	github.com/libp2p/go-libp2p-core v0.8.5
	github.com/libp2p/go-libp2p-kad-dht v0.12.2
	github.com/libp2p/go-libp2p-kbucket v0.4.7
	github.com/libp2p/go-libp2p-pubsub v0.4.1
	github.com/libp2p/go-libp2p-transport-upgrader v0.4.2
	github.com/libp2p/go-tcp-transport v0.2.3
	github.com/mitchellh/mapstructure v1.4.1
	github.com/multiformats/go-multiaddr v0.3.3
	github.com/pelletier/go-toml v1.9.3
	github.com/pkg/errors v0.9.1
	github.com/shirou/gopsutil v3.21.6+incompatible
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20190318030020-c3a204f8e965
	github.com/urfave/cli v1.22.5
	github.com/whyrusleeping/timecache v0.0.0-20160911033111-cfcb2f1abfee
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	golang.org/x/net v0.0.0-20210423184538-5f58ad60dda6
	gopkg.in/go-playground/validator.v8 v8.18.2
)

replace github.com/gogo/protobuf => github.com/ElrondNetwork/protobuf v1.3.2

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_2 v1.2.28 => github.com/ElrondNetwork/arwen-wasm-vm v1.2.28

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_3 v1.3.27 => github.com/ElrondNetwork/arwen-wasm-vm v1.3.27

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_4 v1.4.3 => github.com/ElrondNetwork/arwen-wasm-vm v1.4.3
