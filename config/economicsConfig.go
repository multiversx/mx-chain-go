package config

// GlobalSettings will hold general economic values
type GlobalSettings struct {
	GenesisTotalSupply string
	MinimumInflation   float64
	YearSettings       []*YearSetting
	Denomination       int
}

// MaxInflationInYear will hold the maximum inflation rate for year
type YearSetting struct {
	Year             uint32
	MaximumInflation float64
}

// RewardsSettings will hold economics rewards settings
type RewardsSettings struct {
	LeaderPercentage    float64
	DeveloperPercentage float64
	CommunityPercentage float64
	CommunityAddress    string
}

// FeeSettings will hold economics fee settings
type FeeSettings struct {
	MaxGasLimitPerBlock     string
	MaxGasLimitPerMetaBlock string
	GasPerDataByte          string
	DataLimitForBaseCalc    string
	MinGasPrice             string
	MinGasLimit             string
}

// ValidatorSettings will hold the validator settings
type ValidatorSettings struct {
	GenesisNodePrice                     string
	UnBondPeriod                         string
	TotalSupply                          string
	MinStepValue                         string
	AuctionEnableNonce                   string
	StakeEnableNonce                     string
	NumRoundsWithoutBleed                string
	MaximumPercentageToBleed             string
	BleedPercentagePerRound              string
	UnJailValue                          string
	ActivateBLSPubKeyMessageVerification bool
}

// EconomicsConfig will hold economics config
type EconomicsConfig struct {
	GlobalSettings    GlobalSettings
	RewardsSettings   RewardsSettings
	FeeSettings       FeeSettings
	ValidatorSettings ValidatorSettings
}
