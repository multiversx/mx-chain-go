package nodesExporter

type Node struct {
	BlsKey    string `json:"blsKey"`
	Qualified bool   `json:"qualified"`
}

type Auction struct {
	Owner          string `json:"owner"`
	NumStakedNodes int    `json:"numStakedNodes"`
	TotalTopUp     string `json:"totalTopUp"`
	TopUpPerNode   string `json:"topUpPerNode"`
	QualifiedTopUp string `json:"qualifiedTopUp"`
	Nodes          []Node `json:"nodes"`
}

type ValidatorAuctionData struct {
	AuctionList []Auction `json:"auctionList"`
}

type ValidatorAuctionResponse struct {
	Data ValidatorAuctionData `json:"data"`
}
