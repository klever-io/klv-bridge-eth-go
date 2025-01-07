package mock

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever"
)

// TODO: check if the tokenID in the Safe contract uses a ticker value or the token contract address.
type tokenMock struct {
	EthAddress      string
	TokenId         string
	IsNativeToken   bool
	IsMintBurnToken bool
	TotalBalance    *big.Int
	MintBalance     *big.Int
	BurnBalance     *big.Int
}

var tokenList = []tokenMock{
	{
		EthAddress:      "0x82afDD299Adf4b1e6E101BF6508b52384aa0714f",
		TokenId:         "tck-000001",
		IsNativeToken:   false,
		IsMintBurnToken: true,
		BurnBalance:     big.NewInt(0),
		MintBalance:     big.NewInt(0),
		TotalBalance:    big.NewInt(0),
	},
}

// TODO: remove CreateMockProxyKLV when real proxy is implemented.
// CreateMockProxyKLV returns a Proxy interface, to be used as a klever client proxy
// while real proxy implementation is in development.
func CreateMockProxyKLV() klever.Proxy {
	proxyMock := NewKleverChainMock()

	for _, token := range tokenList {
		ethTokenAddress := common.HexToAddress(token.EthAddress)
		proxyMock.AddTokensPair(ethTokenAddress, token.TokenId, token.IsNativeToken, token.IsMintBurnToken, token.TotalBalance, token.MintBalance, token.BurnBalance)
	}

	return proxyMock
}
