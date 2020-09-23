module github.com/ElrondNetwork/elrond-go

go 1.13

require (
	github.com/ElrondNetwork/arwen-wasm-vm v0.3.36-0.20200918125038-6a78974dde88
	github.com/ElrondNetwork/concurrent-map v0.1.3
	github.com/ElrondNetwork/elrond-go-logger v1.0.4
	github.com/ElrondNetwork/elrond-vm-common v0.2.0
	github.com/beevik/ntp v0.2.0
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/denisbrodbeck/machineid v1.0.1
	github.com/elastic/go-elasticsearch/v7 v7.1.0
	github.com/gin-contrib/cors v0.0.0-20190301062745-f9e10995c85a
	github.com/gin-contrib/pprof v1.3.0
	github.com/gin-gonic/gin v1.6.3
	github.com/gizak/termui/v3 v3.1.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/google/gops v0.3.6
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/golang-lru v0.5.4
	github.com/herumi/bls-go-binary v0.0.0-20200324054641-17de9ae04665
	github.com/ipfs/go-log v1.0.4
	github.com/jbenet/goprocess v0.1.4
	github.com/libp2p/go-libp2p v0.10.3
	github.com/libp2p/go-libp2p-core v0.6.1
	github.com/libp2p/go-libp2p-discovery v0.5.0
	github.com/libp2p/go-libp2p-kad-dht v0.8.3
	github.com/libp2p/go-libp2p-kbucket v0.4.2
	github.com/libp2p/go-libp2p-pubsub v0.3.3
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mr-tron/base58 v1.2.0
	github.com/multiformats/go-multiaddr v0.2.2
	github.com/pelletier/go-toml v1.8.0
	github.com/pkg/errors v0.9.1
	github.com/shirou/gopsutil v0.0.0-20190731134726-d80c43f9c984
	github.com/stretchr/testify v1.6.1
	github.com/syndtr/goleveldb v1.0.1-0.20190318030020-c3a204f8e965
	github.com/urfave/cli v1.22.4
	github.com/whyrusleeping/timecache v0.0.0-20160911033111-cfcb2f1abfee
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	golang.org/x/net v0.0.0-20200519113804-d87ec0cfa476
	gopkg.in/go-playground/validator.v8 v8.18.2
)

replace github.com/gogo/protobuf => github.com/ElrondNetwork/protobuf v1.3.2

replace github.com/ElrondNetwork/arwen-wasm-vm => ../arwen-wasm-vm
