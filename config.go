package clob

import (
	"fmt"

	"github.com/jeffWu90/go-clob-client-v2/types"
)

// ContractConfig holds the on-chain contract addresses for a given network.
type ContractConfig struct {
	Exchange          string
	NegRiskAdapter    string
	NegRiskExchange   string
	Collateral        string
	ConditionalTokens string

	ExchangeV2        string
	NegRiskExchangeV2 string
}

// CollateralTokenDecimals is the number of decimals on the USDC collateral token.
const CollateralTokenDecimals = 6

// ConditionalTokenDecimals is the number of decimals on the CTF ERC1155 outcome tokens.
const ConditionalTokenDecimals = 6

var (
	amoyContracts = ContractConfig{
		Exchange:          "0xdFE02Eb6733538f8Ea35D585af8DE5958AD99E40",
		NegRiskAdapter:    "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
		NegRiskExchange:   "0xC5d563A36AE78145C45a50134d48A1215220f80a",
		Collateral:        "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB",
		ConditionalTokens: "0x69308FB512518e39F9b16112fA8d994F4e2Bf8bB",
		ExchangeV2:        "0xE111180000d2663C0091e4f400237545B87B996B",
		NegRiskExchangeV2: "0xe2222d279d744050d28e00520010520000310F59",
	}

	maticContracts = ContractConfig{
		Exchange:          "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E",
		NegRiskAdapter:    "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
		NegRiskExchange:   "0xC5d563A36AE78145C45a50134d48A1215220f80a",
		Collateral:        "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB",
		ConditionalTokens: "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045",
		ExchangeV2:        "0xE111180000d2663C0091e4f400237545B87B996B",
		NegRiskExchangeV2: "0xe2222d279d744050d28e00520010520000310F59",
	}
)

// GetContractConfig returns the contract addresses for the given chain id.
// Returns an error for any chain other than Polygon (137) or Amoy (80002).
func GetContractConfig(chainID types.Chain) (ContractConfig, error) {
	switch chainID {
	case types.ChainPolygon:
		return maticContracts, nil
	case types.ChainAmoy:
		return amoyContracts, nil
	default:
		return ContractConfig{}, fmt.Errorf("invalid network: %d", chainID)
	}
}
