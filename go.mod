module github.com/ElrondNetwork/elrond-go

go 1.15

require (
	github.com/ElrondNetwork/arwen-wasm-vm/v1_2 v1.2.42-0.20220729115258-b9f2fb2f6568
	github.com/ElrondNetwork/arwen-wasm-vm/v1_3 v1.3.42-0.20220729115131-85ecca868e90
	github.com/ElrondNetwork/arwen-wasm-vm/v1_4 v1.4.59-0.20220729115431-a6c93119bdda
	github.com/ElrondNetwork/covalent-indexer-go v1.0.6
	github.com/ElrondNetwork/elastic-indexer-go v1.2.42-0.20220921110140-860b4b8c7fc3
	github.com/ElrondNetwork/elrond-go-core v1.1.20-0.20220922083155-e3a18647a2d1
	github.com/ElrondNetwork/elrond-go-crypto v1.0.1
	github.com/ElrondNetwork/elrond-go-logger v1.0.7
	github.com/ElrondNetwork/elrond-go-storage v1.0.1
	github.com/ElrondNetwork/elrond-vm-common v1.3.17
	github.com/ElrondNetwork/go-libp2p-pubsub v0.6.1-rc1
	github.com/beevik/ntp v0.3.0
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/davecgh/go-spew v1.1.1
	github.com/elastic/go-elasticsearch/v7 v7.12.0
	github.com/gin-contrib/cors v0.0.0-20190301062745-f9e10995c85a
	github.com/gin-contrib/pprof v1.4.0
	github.com/gin-gonic/gin v1.8.1
	github.com/gizak/termui/v3 v3.1.0
	github.com/gogo/protobuf v1.3.2
	github.com/google/gops v0.3.18
	github.com/gorilla/websocket v1.5.0
	github.com/ipfs/go-log v1.0.5
	github.com/jbenet/goprocess v0.1.4
	github.com/libp2p/go-libp2p v0.19.3
	github.com/libp2p/go-libp2p-core v0.15.1
	github.com/libp2p/go-libp2p-kad-dht v0.15.0
	github.com/libp2p/go-libp2p-kbucket v0.4.7
	github.com/mitchellh/mapstructure v1.5.0
	github.com/multiformats/go-multiaddr v0.5.0
	github.com/pelletier/go-toml v1.9.3
	github.com/pkg/errors v0.9.1
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/stretchr/testify v1.7.1
	github.com/urfave/cli v1.22.10
	github.com/whyrusleeping/timecache v0.0.0-20160911033111-cfcb2f1abfee
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4
	golang.org/x/net v0.0.0-20220418201149-a630d4f3e7a2
	gopkg.in/go-playground/validator.v8 v8.18.2
// test point 3 for custom profiler
)

replace github.com/gogo/protobuf => github.com/ElrondNetwork/protobuf v1.3.2

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_2 v1.2.42-0.20220729115258-b9f2fb2f6568 => github.com/ElrondNetwork/arwen-wasm-vm v1.2.42-0.20220729115258-b9f2fb2f6568

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_3 v1.3.42-0.20220729115131-85ecca868e90 => github.com/ElrondNetwork/arwen-wasm-vm v1.3.42-0.20220729115131-85ecca868e90

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_4 v1.4.59-0.20220729115431-a6c93119bdda => github.com/ElrondNetwork/arwen-wasm-vm v1.4.59-0.20220729115431-a6c93119bdda
