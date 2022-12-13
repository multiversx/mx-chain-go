package common

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go-core/data"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

// vars
var (
	mutData                         sync.Mutex
	mapAddresses                    map[string]uint64
	currentProcessingBlockTimestamp uint64
	currentShardID                  uint32
	log                             = logger.GetOrCreate("common")
	pkConverter, _                  = pubkeyConverter.NewBech32PubkeyConverter(32, log)
	wasHeaderWritten                bool
	buf                             = &bytes.Buffer{}
	sortedKeys                      = make([]string, 0)
	needToWrite                     bool
	file                            *os.File
)

func init() {
	mutData.Lock()
	defer mutData.Unlock()

	mapAddresses = map[string]uint64{
		"erd1qqqqqqqqqqqqqpgqyg62a3tswuzun2wzp8a83sjkc709wwjt2jpssphddm": 1636119990,
		"erd1qqqqqqqqqqqqqpgqe9pvzd3yssaexdg3v8tqr7fh60utpz0h2jpsksum32": 1666615248,
		"erd1qqqqqqqqqqqqqpgq666ta9cwtmjn6dtzv9fhfcmahkjmzhg62jps49k7k8": 1658148870,
		"erd1qqqqqqqqqqqqqpgqye633y7k0zd7nedfnp3m48h24qygm5jl2jpslxallh": 1645436808,
		"erd1qqqqqqqqqqqqqpgqsw9pssy8rchjeyfh8jfafvl3ynum0p9k2jps6lwewp": 1645436868,
		"erd1qqqqqqqqqqqqqpgq389gc8qnqy0ksha948jf72c9cks9g3rf2jpsl5z0aa": 1648546542,
		"erd1qqqqqqqqqqqqqpgqdt892e0aflgm0xwryhhegsjhw0zru60m2jps959w5z": 1650480906,
		"erd1qqqqqqqqqqqqqpgqs0jjyjmx0cvek4p8yj923q5yreshtpa62jpsz6vt84": 1650477984,
		"erd1qqqqqqqqqqqqqpgqjpt0qqgsrdhp2xqygpjtfrpwf76f9nvg2jpsg4q7th": 1647357960,
		"erd1qqqqqqqqqqqqqpgq5e2m9df5yxxkmr86rusejc979arzayjk2jpsz2q43s": 1639604028,
		"erd1qqqqqqqqqqqqqpgq50dge6rrpcra4tp9hl57jl0893a4r2r72jpsk39rjj": 1651744326,
		"erd1qqqqqqqqqqqqqpgqhe8t5jewej70zupmh44jurgn29psua5l2jps3ntjj3": 1654626522,
		"erd1qqqqqqqqqqqqqpgqc3p5xkj656h2ts649h4f447se4clns662jpsccq7rx": 1658314308,
		"erd1qqqqqqqqqqqqqpgqgqfrrkvvuu677u7lz74fhyyyhz7w63kz2jpsuh2wpz": 1666615182,
		"erd1qqqqqqqqqqqqqpgqz55dum74jl0nqxgs7z5d84f448tx0zwp2jpser4d8k": 1651056630,
		"erd1qqqqqqqqqqqqqpgqxrkd0t4lzq6rfdfg03g2ymc7ygv3ayqc2jpsezhzt9": 1651056720,
		"erd1qqqqqqqqqqqqqpgqv692rgsd6teaq78nufvgkgjnmca6anqn2jpsphphj6": 1653918258,
		"erd1qqqqqqqqqqqqqpgq6zat6jefz9z8n54yag8xmz7jw20vlv3x2jpsxms0sd": 1653918342,
		"erd1qqqqqqqqqqqqqpgqv4ks4nzn2cw96mm06lt7s2l3xfrsznmp2jpsszdry5": 1645436910,
		"erd1qqqqqqqqqqqqqpgqyawg3d9r4l27zue7e9sz7djf7p9aj3sz2jpsm070jf": 1651481868,
		"erd1qqqqqqqqqqqqqpgqkxtzg0cd02eahr3tac9efl25cfr9yj4t2jps3xnwda": 1658748972,
		"erd1qqqqqqqqqqqqqpgqj7yq0qysvn2xye0ywzkl0qn95wz8ttcr2jpscpfazv": 1654623150,
		"erd1qqqqqqqqqqqqqpgqq66xk9gfr4esuhem3jru86wg5hvp33a62jps2fy57p": 1658751528,
		"erd1qqqqqqqqqqqqqpgqr7kdhagkqgxvjrsk7s5333l9wwnenr9g2jps8puq33": 1650021420,
		"erd1qqqqqqqqqqqqqpgqzps75vsk97w9nsx2cenv2r2tyxl4fl402jpsx78m9j": 1650475200,
		"erd1qqqqqqqqqqqqqpgq45zs77q884ts6y9zj4jyqfn6ydev8ruv2jps3tteqq": 1652271846,
		"erd1qqqqqqqqqqqqqpgqp2wfzvkhlpkwcdxx25qzznx33345979w2jpsl3gflj": 1658147208,
		"erd1qqqqqqqqqqqqqpgqjdlnu9ggwfc79pygn5fjjgmdm6d7vu5e2jpsw59amp": 1666614324,
		"erd1qqqqqqqqqqqqqpgqs2puacesqywt84jawqenmmree20uj80d2jpsefuka9": 1639603572,
		"erd1qqqqqqqqqqqqqpgqnqvjnn4haygsw2hls2k9zjjadnjf9w7g2jpsmc60a4": 1646310102,
		"erd1qqqqqqqqqqqqqpgqutddd7dva0x4xmehyljp7wh7ecynag0u2jpskxx6xt": 1646310582,
		"erd1qqqqqqqqqqqqqpgq6xp2uvdk22frgqkwm9pwa564gqg2gxfm2jpszggzgw": 1650480612,
		"erd1qqqqqqqqqqqqqpgq38cmvwwkujae3c0xmc9p9554jjctkv4w2jps83r492": 1650881628,
		"erd1qqqqqqqqqqqqqpgqe9v45fnpkv053fj0tk7wvnkred9pms892jps9lkqrn": 1651507524,
		"erd1qqqqqqqqqqqqqpgqmaj6qc4c6f9ldkuwllqh7pyd8e9yu3m82jpscc6f9s": 1653915318,
		"erd1qqqqqqqqqqqqqpgqawujux7w60sjhm8xdx3n0ed8v9h7kpqu2jpsecw6ek": 1666523334,
		"erd1qqqqqqqqqqqqqpgqzy0qwv0jf0yk8tysll0y6ea0fhpfqlp32jpsder9ne": 1666613388,
		"erd1qqqqqqqqqqqqqpgqq93kw6jngqjugfyd7pf555lwurfg2z5e2jpsnhhywk": 1666720350,
		"erd1qqqqqqqqqqqqqpgqdt9aady5jp7av97m7rqxh6e5ywyqsplz2jps5mw02n": 1669651842,
		"erd1qqqqqqqqqqqqqpgqwtzqckt793q8ggufxxlsv3za336674qq2jpszzgqra": 1646310684,
		"erd1qqqqqqqqqqqqqpgq7qhsw8kffad85jtt79t9ym0a4ycvan9a2jps0zkpen": 1651506174,
		"erd1qqqqqqqqqqqqqpgq3f8jfeg34ujzy0muhe9dvn5yngwgud022jpsscjxgl": 1653915534,
		"erd1qqqqqqqqqqqqqpgqcejzfjfmmgvq7yjch2tqnhhsngr7hqyw2jpshce5u2": 1658314380,
		"erd1qqqqqqqqqqqqqpgqlu6vrtkgfjh68nycf5sdw4nyzudncns42jpsr7szch": 1666613460,
		"erd1qqqqqqqqqqqqqpgqs2mmvzpu6wz83z3vthajl4ncpwz67ctu2jpsrcl0ct": 1639603890,
		"erd1qqqqqqqqqqqqqpgq2ymdr66nk5hx32j2tqdeqv9ajm9dj9uu2jps3dtv6v": 1648546488,
		"erd1qqqqqqqqqqqqqpgqt7tyyswqvplpcqnhwe20xqrj7q7ap27d2jps7zczse": 1647357594,
		"erd1qqqqqqqqqqqqqpgqcg4knmq77wxlnkzjgqpjjyk582smte7e2jps3px0sk": 1658749032,
		"erd1qqqqqqqqqqqqqpgqrc4pg2xarca9z34njcxeur622qmfjp8w2jps89fxnl": 1651507662,
		"erd1qqqqqqqqqqqqqpgqmqq78c5htmdnws8hm5u4suvags36eq092jpsaxv3e7": 1648545870,
		"erd1qqqqqqqqqqqqqpgqcedkmj8ezme6mtautj79ngv7fez978le2jps8jtawn": 1653917046,
		"erd1qqqqqqqqqqqqqpgq0tajepcazernwt74820t8ef7t28vjfgukp2sw239f3": 1670600000,
		"erd1qqqqqqqqqqqqqpgqenvn0s3ldc94q2mlkaqx4arj3zfnvnmakp2sxca2h9": 1670600000,
		"erd1qqqqqqqqqqqqqpgq6v5ta4memvrhjs4x3gqn90c3pujc77takp2sqhxm9j": 1670600000,
		"erd1qqqqqqqqqqqqqpgqmgd0eu4z9kzvrputt4l4kw4fqf2wcjsekp2sftan7s": 1670600000,
		"erd1qqqqqqqqqqqqqpgq97fctvqlskzu0str05muew2qyjttp8kfkp2sl9k0gx": 1670600000,
		"erd1qqqqqqqqqqqqqpgqxv0y4p6vvszrknaztatycac77yvsxqrrkp2sghd86c": 1670600000,
		"erd1qqqqqqqqqqqqqpgqv0pz5z3fkz54nml6pkzzdruuf020gqzykp2sya7kkv": 1670600000,
		"erd1qqqqqqqqqqqqqpgquhqfknr9dg5t0xma0zcc55dm8gwv5rpkkp2sq8f40p": 1670600000,
		"erd1qqqqqqqqqqqqqpgqjvcrk4tun6l438crr0jcz7q3yzp66smwkp2shzmazs": 1670600000,
		"erd1qqqqqqqqqqqqqpgqx6zhjfjnwxpzdcj594htl0y6w60aa3a8kp2se2pdms": 1670600000,
		"erd1qqqqqqqqqqqqqpgqg48z8qfm028ze6y3d6tamdfaw6vn774tkp2shg5dt5": 1670600000,
		"erd1qqqqqqqqqqqqqpgqjsnxqprks7qxfwkcg2m2v9hxkrchgm9akp2segrswt": 1670600000,
		"erd1qqqqqqqqqqqqqpgqt6ltx52ukss9d2qag2k67at28a36xc9lkp2sr06394": 1670600000,
		"erd1qqqqqqqqqqqqqpgqrdq6zvepdxg36rey8pmluwur43q4hcx3kp2su4yltq": 1670600000,
		"erd1qqqqqqqqqqqqqpgq4acurmluezvmhna8tztgcrnwh0l70a2wkp2sfh6jkp": 1670600000,
		"erd1qqqqqqqqqqqqqpgqduftj7u5u5d4ql2znchcmrcet9kcze7ckp2sk4474f": 1670600000,
		"erd1qqqqqqqqqqqqqpgqapxdp9gjxtg60mjwhle3n6h88zch9e7kkp2s8aqhkg": 1670600000,
		"erd1qqqqqqqqqqqqqpgqxwakt2g7u9atsnr03gqcgmhcv38pt7mkd94q6shuwt": 1646326740,
		"erd1qqqqqqqqqqqqqpgq37fmv57uqkayxctplh4swkw9vfz2jawvqqyqdugcju": 1646241714,
		"erd1qqqqqqqqqqqqqpgqnhvsujzd95jz6fyv3ldmynlf97tscs9nqqqq49en6w": 1646241714,
		"erd1qqqqqqqqqqqqqpgqed4re4upmy8qmtpwfqk7jfkepduyn7y0yfkq4ew7s6": 1652876700,
		"erd1qqqqqqqqqqqqqpgq305jfaqrdxpzjgf9y5gvzh60mergh866yfkqzqjv2h": 1652876754,
		"erd1qqqqqqqqqqqqqpgqhxkc48lt5uv2hejj4wtjqvugfm4wgv6gyfkqw0uuxl": 1652876802,
		"erd1qqqqqqqqqqqqqpgqsevgnqk0vu424sg866nh8k2v2hdekrwdyfkqh2mdwr": 1652876838,
		"erd1qqqqqqqqqqqqqpgqxexs26vrvhwh2m4he62d6y3jzmv3qkujyfkq8yh4z2": 1652876868,
	}

	log.Info("loaded MapAddresses", "num records", len(mapAddresses))
}

// OutputTxCounters -
func OutputTxCounters(tx data.TransactionHandler, countersMap map[string]uint64, txHash []byte) {
	receiver := pkConverter.Encode(tx.GetRcvAddr())

	mutData.Lock()
	defer mutData.Unlock()

	if !wasHeaderWritten {
		needToWrite = true

		for key := range countersMap {
			sortedKeys = append(sortedKeys, key)
		}

		sort.Slice(sortedKeys, func(i, j int) bool {
			return sortedKeys[i] < sortedKeys[j]
		})

		_, _ = buf.WriteString(`"Shard ID","Tx hash","Receiver","Method"`)
		for _, key := range sortedKeys {
			_, _ = buf.WriteString(",\"" + key + "\"")
		}
		_, _ = buf.WriteString("\n")

		wasHeaderWritten = true
	}

	timeStamp, found := mapAddresses[receiver]
	if !found {
		return
	}
	if currentProcessingBlockTimestamp < timeStamp {
		return
	}

	needToWrite = true
	method := "<unknown>"
	splts := strings.Split(string(tx.GetData()), "@")
	if len(splts) > 1 {
		method = splts[0]
	}
	method = processMethod(method, splts)

	_, _ = buf.WriteString(fmt.Sprintf(`"%d","%x","%s","%s"`,
		currentShardID, txHash, receiver, method))
	writeCounters(buf, countersMap)
	_, _ = buf.WriteString("\n")
}

func processMethod(method string, splts []string) string {
	if method == "ESDTNFTTransfer" {
		return processESDTNFTTransfer(method, splts)
	}
	if method == "MultiESDTNFTTransfer" {
		return processMultiESDTNFTTransfer(method, splts)
	}

	return method
}

func processESDTNFTTransfer(method string, splts []string) string {
	if len(splts) <= 5 {
		return method + " ???"
	}

	newMethod, err := hex.DecodeString(splts[5])
	if err != nil {
		return method + " ???"
	}

	return string(newMethod)
}

func processMultiESDTNFTTransfer(method string, splts []string) string {
	if len(splts) <= 2 {
		return method + " ???"
	}

	numIndex := 1
	if len(splts[1]) == 32 {
		numIndex++
	}

	numTransfersBytes, err := hex.DecodeString(splts[numIndex])
	if err != nil {
		return method + " ???"
	}

	numTransfers := big.NewInt(0).SetBytes(numTransfersBytes).Uint64()
	methodIndex := numIndex + 1 + int(numTransfers)*3
	if len(splts) <= methodIndex {
		return method + " ???"
	}

	newMethod, err := hex.DecodeString(splts[methodIndex])
	if err != nil {
		return method + " ???"
	}

	return string(newMethod)
}

func writeCounters(buff *bytes.Buffer, counters map[string]uint64) {
	for _, key := range sortedKeys {
		buff.WriteString(fmt.Sprintf(",\"%d\"", counters[key]))
	}
}

// SetHeaderData -
func SetHeaderData(timestamp uint64, shardID uint32) {
	mutData.Lock()
	currentProcessingBlockTimestamp = timestamp
	currentShardID = shardID
	mutData.Unlock()
}

// WriteCSV -
func WriteCSV() {
	mutData.Lock()
	defer mutData.Unlock()

	if !needToWrite {
		return
	}

	var err error
	if file == nil {
		args := core.ArgCreateFileArgument{
			Directory:     "",
			Prefix:        "data",
			FileExtension: "csv",
		}

		file, err = core.CreateFile(args)
		if err != nil {
			log.Error("error creating file", "error", err)
			return
		}
	}

	_, err = file.Write(buf.Bytes())
	log.LogIfError(err)

	buf.Reset()
}

// CloseCSV -
func CloseCSV() {
	mutData.Lock()
	defer mutData.Unlock()

	if file == nil {
		return
	}

	err := file.Close()
	log.LogIfError(err)
}
