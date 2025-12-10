package mock

import (
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klv-bridge-eth-go/integrationTests"
)

// DecimalConversionConfig holds the multiplier and divisor for decimal conversion
type DecimalConversionConfig struct {
	Multiplier *big.Int
	Divisor    *big.Int
}

// tokensRegistryMock is not concurrent safe
type tokensRegistryMock struct {
	ethToKC           map[common.Address]string
	kcToEth           map[string]common.Address
	mintBurnTokens    map[string]bool
	nativeTokens      map[string]bool
	totalBalances     map[string]*big.Int
	mintBalances      map[string]*big.Int
	burnBalances      map[string]*big.Int
	decimalConversion map[string]*DecimalConversionConfig // key is hex-encoded ticker
}

func (mock *tokensRegistryMock) addTokensPair(erc20Address common.Address, ticker string, isNativeToken, isMintBurnToken bool, totalBalance, mintBalances, burnBalances *big.Int) {
	integrationTests.Log.Info("added tokens pair", "ticker", ticker,
		"erc20 address", erc20Address.String(), "is native token", isNativeToken, "is mint burn token", isMintBurnToken,
		"total balance", totalBalance, "mint balances", mintBalances, "burn balances", burnBalances)

	mock.ethToKC[erc20Address] = ticker

	hexedTicker := hex.EncodeToString([]byte(ticker))
	mock.kcToEth[hexedTicker] = erc20Address

	if isNativeToken {
		mock.nativeTokens[hexedTicker] = true
	}
	if isMintBurnToken {
		mock.mintBurnTokens[hexedTicker] = true
		mock.mintBalances[hexedTicker] = mintBalances
		mock.burnBalances[hexedTicker] = burnBalances
	} else {
		mock.totalBalances[hexedTicker] = totalBalance
	}
}

// setDecimalConversion sets the decimal conversion configuration for a token
// The conversion is calculated as: convertedAmount = (amount * multiplier) / divisor
// For example, ETH 18 decimals to KDA 6 decimals: multiplier=1, divisor=10^12
func (mock *tokensRegistryMock) setDecimalConversion(ticker string, multiplier, divisor *big.Int) {
	hexedTicker := hex.EncodeToString([]byte(ticker))
	mock.decimalConversion[hexedTicker] = &DecimalConversionConfig{
		Multiplier: multiplier,
		Divisor:    divisor,
	}
}

// getConvertedAmount converts an amount using the decimal conversion config for the token
// If no conversion is configured, returns the same amount
func (mock *tokensRegistryMock) getConvertedAmount(hexedTicker string, amount *big.Int) *big.Int {
	config, exists := mock.decimalConversion[hexedTicker]
	if !exists || config == nil {
		return amount
	}
	// Defensive guards to avoid panics
	if config.Multiplier == nil || config.Divisor == nil || config.Divisor.Sign() == 0 {
		return amount
	}
	// convertedAmount = (amount * multiplier) / divisor
	result := new(big.Int).Mul(amount, config.Multiplier)
	result.Div(result, config.Divisor)
	return result
}

func (mock *tokensRegistryMock) clearTokens() {
	mock.ethToKC = make(map[common.Address]string)
	mock.kcToEth = make(map[string]common.Address)
	mock.mintBurnTokens = make(map[string]bool)
	mock.nativeTokens = make(map[string]bool)
	mock.totalBalances = make(map[string]*big.Int)
	mock.mintBalances = make(map[string]*big.Int)
	mock.burnBalances = make(map[string]*big.Int)
	mock.decimalConversion = make(map[string]*DecimalConversionConfig)
}

func (mock *tokensRegistryMock) getTicker(erc20Address common.Address) string {
	ticker, found := mock.ethToKC[erc20Address]
	if !found {
		panic("tiker for erc20 address " + erc20Address.String() + " not found")
	}

	return ticker
}

func (mock *tokensRegistryMock) getErc20Address(ticker string) common.Address {
	addr, found := mock.kcToEth[ticker]
	if !found {
		panic("erc20 address for ticker " + ticker + " not found")
	}

	return addr
}

func (mock *tokensRegistryMock) isMintBurnToken(ticker string) bool {
	_, found := mock.mintBurnTokens[ticker]

	return found
}

func (mock *tokensRegistryMock) isNativeToken(ticker string) bool {
	_, found := mock.nativeTokens[ticker]

	return found
}

func (mock *tokensRegistryMock) getTotalBalances(ticker string) *big.Int {
	return mock.totalBalances[ticker]
}

func (mock *tokensRegistryMock) getMintBalances(ticker string) *big.Int {
	return mock.mintBalances[ticker]
}

func (mock *tokensRegistryMock) getBurnBalances(ticker string) *big.Int {
	return mock.burnBalances[ticker]
}
