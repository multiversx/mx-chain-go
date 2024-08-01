package nodesExporter

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("nodesExporter")

var NodesExporterInstance = NodesExporter{}

type NodesExporter struct {
	apiPort int
	shardId uint32
}

func (n *NodesExporter) SetApiPort(port string) {
	splt := strings.Split(port, ":")
	if len(splt) != 2 {
		n.apiPort = 8080
	} else {
		number, err := strconv.Atoi(splt[1])
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		n.apiPort = number
	}
}

func (n *NodesExporter) SetSelfShardId(shardId uint32) {
	n.shardId = shardId
}

func (n *NodesExporter) ExportNqtForEpoch(epoch uint32) {
	if n.shardId != core.MetachainShardId {
		log.Info("NQT will only be extracted from metachain nodes")
		return
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/validator/auction", n.apiPort)
	resp, err := http.Get(url)
	if err != nil {
		log.Error("cannot fetch validation auction API", "error", err, "epoch", epoch)
		return
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("cannot read validator auction response body", "error", err, "epoch", epoch)
		return
	}

	var auctionList ValidatorAuctionResponse

	err = json.Unmarshal(respBody, &auctionList)
	if err != nil {
		log.Error("cannot unmarshal validator auction response", "error", err, "epoch", epoch)
		return
	}

	minTopUp := big.NewInt(0)
	for _, auctions := range auctionList.Data.AuctionList {
		for _, node := range auctions.Nodes {
			if !node.Qualified {
				continue
			}

			minTopUpBI, _ := big.NewInt(0).SetString(auctions.QualifiedTopUp, 10)
			if minTopUpBI.Cmp(minTopUp) < 0 {
				minTopUp = minTopUpBI
			}
		}
	}

	baseStakeBi, _ := big.NewInt(0).SetString("2500000000000000000000", 10)
	nqt := big.NewInt(0).Add(baseStakeBi, minTopUp)
	fileContent := fmt.Sprintf("nqt:%s,topUp:%s\n", nqt.String(), minTopUp.String())
	err = os.WriteFile(fmt.Sprintf("nqt_epoch%d.txt", epoch), []byte(fileContent), 0644)
	if err != nil {
		log.Error("cannot save NQT to file.", "error", err, "epoch", epoch, "nqt", nqt.String(), "top up", minTopUp.String())
		return
	}

	log.Info("[NodesExporter] Successfully saved nqt data", "epoch", epoch, "nqt", nqt.String())
}

func (n *NodesExporter) ExportHeartbeatForEpoch(epoch uint32) {
	url := fmt.Sprintf("http://127.0.0.1:%d/node/heartbeatstatus", n.apiPort)
	resp, err := http.Get(url)
	if err != nil {
		log.Error("cannot fetch node heartbeat status API", "error", err, "epoch", epoch)
		return
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("cannot read node heartbeat status response body", "error", err, "epoch", epoch)
		return
	}

	err = os.WriteFile(fmt.Sprintf("heartbeat_epoch%d_shard%d.json", epoch, n.shardId), respBody, 0644)
	if err != nil {
		log.Error("cannot save heartbeat to file.", "error", err, "epoch", epoch)
		return
	}

	log.Info("[NodesExporter] Successfully saved heartbeat data", "epoch", epoch)
}
