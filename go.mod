module github.com/ElrondNetwork/elrond-go

go 1.15

require (
	github.com/ElrondNetwork/covalent-indexer-go v1.0.6
	github.com/ElrondNetwork/elastic-indexer-go v1.2.44
	github.com/ElrondNetwork/elrond-go-core v1.1.21-0.20221019080438-384a03841a75
	github.com/ElrondNetwork/elrond-go-crypto v1.2.1
	github.com/ElrondNetwork/elrond-go-logger v1.0.9
	github.com/ElrondNetwork/elrond-go-p2p v1.0.3
	github.com/ElrondNetwork/elrond-go-storage v1.0.1
	github.com/ElrondNetwork/elrond-vm-common v1.3.22
	github.com/ElrondNetwork/wasm-vm v1.4.59
	github.com/ElrondNetwork/wasm-vm-v1_2 v1.2.42
	github.com/ElrondNetwork/wasm-vm-v1_3 v1.3.42
	github.com/beevik/ntp v0.3.0
	github.com/davecgh/go-spew v1.1.1
	github.com/elastic/go-elasticsearch/v7 v7.12.0
	github.com/gin-contrib/cors v0.0.0-20190301062745-f9e10995c85a
	github.com/gin-contrib/pprof v1.4.0
	github.com/gin-gonic/gin v1.8.1
	github.com/gizak/termui/v3 v3.1.0
	github.com/gogo/protobuf v1.3.2
	github.com/google/gops v0.3.18
	github.com/gorilla/websocket v1.5.0
	github.com/libp2p/go-libp2p-core v0.15.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pelletier/go-toml v1.9.3
	github.com/pkg/errors v0.9.1
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/stretchr/testify v1.7.1
	github.com/urfave/cli v1.22.10
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4
	golang.org/x/net v0.0.0-20220418201149-a630d4f3e7a2
	gopkg.in/go-playground/validator.v8 v8.18.2
// test point 3 for custom profiler
)

replace github.com/gogo/protobuf => github.com/ElrondNetwork/protobuf v1.3.2
