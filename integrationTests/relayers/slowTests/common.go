//go:build slow

package slowTests

import (
	"math/big"

	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	"github.com/klever-io/klv-bridge-eth-go/parsers"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var (
	log = logger.GetOrCreate("integrationTests/relayers/slowTests")
)

// GenerateTestUSDCToken will generate a test USDC token
func GenerateTestUSDCToken() framework.TestTokenParams {
	// USDC is ethNative = true, ethMintBurn = false, kdaNative = false, kdaMintBurn = true
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "USDC",
			NumOfDecimalsUniversal:           6,
			NumOfDecimalsChainSpecific:       6,
			KlvUniversalTokenTicker:          "USDC",
			KlvChainSpecificTokenTicker:      "ETHUSDC",
			KlvUniversalTokenDisplayName:     "WrappedUSDC",
			KlvChainSpecificTokenDisplayName: "EthereumWrappedUSDC",
			ValueToMintOnKlv:                 "10000000000",
			IsMintBurnOnKlv:                  true,
			IsNativeOnKlv:                    false,
			HasChainSpecificToken:            true,
			EthTokenName:                     "EthUSDC",
			EthTokenSymbol:                   "USDC",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  false,
			IsNativeOnEth:                    true,
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToKlv: big.NewInt(5000),
				ValueToSendFromKlv:   big.NewInt(2500),
			},
			{
				ValueToTransferToKlv: big.NewInt(7000),
				ValueToSendFromKlv:   big.NewInt(300),
			},
			{
				ValueToTransferToKlv: big.NewInt(1000),
				ValueToSendFromKlv:   nil,
				KlvSCCallData:        createScCallData("callPayable", 50000000),
			},
		},
		KDASafeExtraBalance:     big.NewInt(100),                                        // extra is just for the fees for the 2 transfers kda->eth
		EthTestAddrExtraBalance: big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50 - 1000), // -(eth->kda) + (kda->eth) - fees
	}
}

// GenerateTestMEMEToken will generate a test MEME token
func GenerateTestMEMEToken() framework.TestTokenParams {
	//MEME is ethNative = false, ethMintBurn = true, kdaNative = true, kdaMintBurn = false
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "MEME",
			NumOfDecimalsUniversal:           1,
			NumOfDecimalsChainSpecific:       1,
			KlvUniversalTokenTicker:          "MEME",
			KlvChainSpecificTokenTicker:      "ETHMEME",
			KlvUniversalTokenDisplayName:     "WrappedMEME",
			KlvChainSpecificTokenDisplayName: "EthereumWrappedMEME",
			ValueToMintOnKlv:                 "10000000000",
			IsMintBurnOnKlv:                  false,
			IsNativeOnKlv:                    true,
			HasChainSpecificToken:            true,
			EthTokenName:                     "EthMEME",
			EthTokenSymbol:                   "MEME",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  true,
			IsNativeOnEth:                    false,
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToKlv: big.NewInt(2400),
				ValueToSendFromKlv:   big.NewInt(4000),
			},
			{
				ValueToTransferToKlv: big.NewInt(200),
				ValueToSendFromKlv:   big.NewInt(6000),
			},
			{
				ValueToTransferToKlv: big.NewInt(1000),
				ValueToSendFromKlv:   big.NewInt(2000),
				KlvSCCallData:        createScCallData("callPayable", 50000000),
			},
		},
		KDASafeExtraBalance:     big.NewInt(4000 + 6000 + 2000), // everything is locked in the safe kda contract
		EthTestAddrExtraBalance: big.NewInt(4000 - 50 + 6000 - 50 + 2000 - 50),
	}
}

// GenerateTestEUROCToken will generate a test EUROC token
func GenerateTestEUROCToken() framework.TestTokenParams {
	//EUROC is ethNative = true, ethMintBurn = true, kdaNative = false, kdaMintBurn = true
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "EUROC",
			NumOfDecimalsUniversal:           6,
			NumOfDecimalsChainSpecific:       6,
			KlvUniversalTokenTicker:          "EUROC",
			KlvChainSpecificTokenTicker:      "EUROC",
			KlvUniversalTokenDisplayName:     "TestEUROC",
			KlvChainSpecificTokenDisplayName: "TestEUROC",
			ValueToMintOnKlv:                 "10000000000",
			IsMintBurnOnKlv:                  true,
			IsNativeOnKlv:                    false,
			HasChainSpecificToken:            false,
			EthTokenName:                     "EthEuroC",
			EthTokenSymbol:                   "EUROC",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  true,
			IsNativeOnEth:                    true,
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToKlv: big.NewInt(5010),
				ValueToSendFromKlv:   big.NewInt(2510),
			},
			{
				ValueToTransferToKlv: big.NewInt(7010),
				ValueToSendFromKlv:   big.NewInt(310),
			},
			{
				ValueToTransferToKlv: big.NewInt(1010),
				ValueToSendFromKlv:   nil,
				KlvSCCallData:        createScCallData("callPayable", 50000000),
			},
		},
		KDASafeExtraBalance:     big.NewInt(100),                                        // extra is just for the fees for the 2 transfers kda->eth
		EthTestAddrExtraBalance: big.NewInt(-5010 + 2510 - 50 - 7010 + 310 - 50 - 1010), // -(eth->kda) + (kda->eth) - fees
	}
}

// GenerateTestMEXToken will generate a test EUROC token
func GenerateTestMEXToken() framework.TestTokenParams {
	//MEX is ethNative = false, ethMintBurn = true, kdaNative = true, kdaMintBurn = true
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "MEX",
			NumOfDecimalsUniversal:           2,
			NumOfDecimalsChainSpecific:       2,
			KlvUniversalTokenTicker:          "MEX",
			KlvChainSpecificTokenTicker:      "MEX",
			KlvUniversalTokenDisplayName:     "TestMEX",
			KlvChainSpecificTokenDisplayName: "TestMEX",
			ValueToMintOnKlv:                 "10000000000",
			IsMintBurnOnKlv:                  true,
			IsNativeOnKlv:                    true,
			HasChainSpecificToken:            false,
			EthTokenName:                     "EthMex",
			EthTokenSymbol:                   "MEX",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  true,
			IsNativeOnEth:                    false,
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToKlv: big.NewInt(2410),
				ValueToSendFromKlv:   big.NewInt(4010),
			},
			{
				ValueToTransferToKlv: big.NewInt(210),
				ValueToSendFromKlv:   big.NewInt(6010),
			},
			{
				ValueToTransferToKlv: big.NewInt(1010),
				ValueToSendFromKlv:   big.NewInt(2010),
				KlvSCCallData:        createScCallData("callPayable", 50000000),
			},
		},
		KDASafeExtraBalance:     big.NewInt(150), // just the fees should be collected in KDA safe
		EthTestAddrExtraBalance: big.NewInt(4010 - 50 + 6010 - 50 + 2010 - 50),
	}
}

func createScCallData(function string, gasLimit uint64, args ...string) []byte {
	codec := testsCommon.TestKcCodec{}
	callData := parsers.CallData{
		Type:      bridgeCore.DataPresentProtocolMarker,
		Function:  function,
		GasLimit:  gasLimit,
		Arguments: args,
	}

	return codec.EncodeCallDataStrict(callData)
}
