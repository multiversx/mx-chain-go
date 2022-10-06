module github.com/ElrondNetwork/elrond-go

go 1.15

require (
	github.com/ElrondNetwork/arwen-wasm-vm/v1_2 v1.2.42-0.20220729115258-b9f2fb2f6568
	github.com/ElrondNetwork/arwen-wasm-vm/v1_3 v1.3.42-0.20220729115131-85ecca868e90
	github.com/ElrondNetwork/arwen-wasm-vm/v1_4 v1.4.59-0.20220729115431-a6c93119bdda
	github.com/ElrondNetwork/covalent-indexer-go v1.0.7-0.20220922092743-7f3a26b4001a
	github.com/ElrondNetwork/elastic-indexer-go v1.2.40-0.20220922074555-961c00b8bfce
	github.com/ElrondNetwork/elrond-go-core v1.1.21-0.20221006120212-5c0ba882cdd0
	github.com/ElrondNetwork/elrond-go-crypto v1.2.1
	github.com/ElrondNetwork/elrond-go-logger v1.0.8-0.20220901113442-d5ad00505a23
	github.com/ElrondNetwork/elrond-go-p2p v1.0.2
	github.com/ElrondNetwork/elrond-go-storage v1.0.1
	github.com/ElrondNetwork/elrond-vm-common v1.3.18-0.20220921081708-baae086376d6
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

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_2 v1.2.42-0.20220729115258-b9f2fb2f6568 => github.com/ElrondNetwork/arwen-wasm-vm v1.2.42-0.20220729115258-b9f2fb2f6568

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_3 v1.3.42-0.20220729115131-85ecca868e90 => github.com/ElrondNetwork/arwen-wasm-vm v1.3.42-0.20220729115131-85ecca868e90

replace github.com/ElrondNetwork/arwen-wasm-vm/v1_4 v1.4.59-0.20220729115431-a6c93119bdda => github.com/ElrondNetwork/arwen-wasm-vm v1.4.59-0.20220729115431-a6c93119bdda
