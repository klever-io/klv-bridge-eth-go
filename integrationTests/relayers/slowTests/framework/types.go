package framework

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// IssueTokenParams the parameters when issuing a new token
type IssueTokenParams struct {
	InitialSupplyParams
	AbstractTokenIdentifier string

	// Klever Blockchain
	NumOfDecimalsUniversal           int
	NumOfDecimalsChainSpecific       byte
	KlvUniversalTokenTicker          string
	KlvChainSpecificTokenTicker      string
	KlvUniversalTokenDisplayName     string
	KlvChainSpecificTokenDisplayName string
	ValueToMintOnKlv                 string
	IsMintBurnOnKlv                  bool
	IsNativeOnKlv                    bool
	HasChainSpecificToken            bool

	// Ethereum
	EthTokenName     string
	EthTokenSymbol   string
	ValueToMintOnEth string
	IsMintBurnOnEth  bool
	IsNativeOnEth    bool
}

// InitialSupplyParams represents the initial supply parameters
type InitialSupplyParams struct {
	InitialSupplyValue string
}

// TokenOperations defines a token operation in a test. Usually this can define one or to deposits in a batch
type TokenOperations struct {
	ValueToTransferToKlv *big.Int
	ValueToSendFromKlv   *big.Int
	KlvSCCallData        []byte
	KlvFaultySCCall      bool
	KlvForceSCCall       bool
}

// TestTokenParams defines a token collection of operations in one or 2 batches
type TestTokenParams struct {
	IssueTokenParams
	TestOperations          []TokenOperations
	KDASafeExtraBalance     *big.Int
	EthTestAddrExtraBalance *big.Int
}

// TokenData represents a test token data
type TokenData struct {
	AbstractTokenIdentifier string

	KlvUniversalTokenTicker     string
	KlvChainSpecificTokenTicker string
	EthTokenName                string
	EthTokenSymbol              string

	KlvUniversalToken     string
	KlvChainSpecificToken string
	EthErc20Address       common.Address
	EthErc20Contract      ERC20Contract
}
