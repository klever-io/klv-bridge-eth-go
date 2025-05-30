//go:build slow

// To run these slow tests, simply add the slow tag on the go test command. Also, provide a chain simulator instance on the 8085 port
// example: go test -tags slow

package slowTests

import (
	"math/big"
	"strings"
	"testing"

	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	"github.com/stretchr/testify/require"
)

func TestRelayersShouldExecuteTransfersWithRefund(t *testing.T) {
	t.Run("unknown marker and malformed SC call data should refund", func(t *testing.T) {
		callData := []byte{5, 4, 55}
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].KlvSCCallData = callData
		usdcToken.TestOperations[2].KlvFaultySCCall = true
		usdcToken.EthTestAddrExtraBalance = big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50 - 1000 + 950) // -(eth->kda) + (kda->eth) - fees + revert after bad SC call
		usdcToken.KDASafeExtraBalance = big.NewInt(150)                                                  // extra is just for the fees for the 2 transfers kda->eth and the failed eth->kda that needed refund

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].KlvSCCallData = callData
		memeToken.TestOperations[2].KlvFaultySCCall = true

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("unknown marker and malformed SC call data should refund with MEX", func(t *testing.T) {
		callData := []byte{5, 4, 55}
		mexToken := GenerateTestMEXToken()
		mexToken.TestOperations[2].KlvSCCallData = callData
		mexToken.TestOperations[2].KlvFaultySCCall = true

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			mexToken,
		)
	})
	t.Run("malformed SC call data should refund", func(t *testing.T) {
		callData := []byte{bridgeCore.DataPresentProtocolMarker, 4, 55}
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].KlvSCCallData = callData
		usdcToken.TestOperations[2].KlvFaultySCCall = true
		usdcToken.EthTestAddrExtraBalance = big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50 - 1000 + 950) // -(eth->kda) + (kda->eth) - fees + revert after bad SC call
		usdcToken.KDASafeExtraBalance = big.NewInt(150)                                                  // extra is just for the fees for the 2 transfers kda->eth and the failed eth->kda that needed refund

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].KlvSCCallData = callData
		memeToken.TestOperations[2].KlvFaultySCCall = true

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("unknown function should refund", func(t *testing.T) {
		callData := createScCallData("unknownFunction", 50000000)
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].KlvSCCallData = callData
		usdcToken.TestOperations[2].KlvFaultySCCall = true
		usdcToken.EthTestAddrExtraBalance = big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50 - 1000 + 950) // -(eth->kda) + (kda->eth) - fees + revert after bad SC call
		usdcToken.KDASafeExtraBalance = big.NewInt(150)                                                  // extra is just for the fees for the 2 transfers kda->eth and the failed eth->kda that needed refund

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].KlvSCCallData = callData
		memeToken.TestOperations[2].KlvFaultySCCall = true

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("wrong deposit with empty sc call data should refund", func(t *testing.T) {
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].KlvSCCallData = nil
		usdcToken.TestOperations[2].KlvFaultySCCall = true
		usdcToken.TestOperations[2].KlvForceSCCall = true
		usdcToken.EthTestAddrExtraBalance = big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50 - 1000 + 950) // -(eth->kda) + (kda->eth) - fees + revert after bad SC call
		usdcToken.KDASafeExtraBalance = big.NewInt(150)                                                  // extra is just for the fees for the 2 transfers kda->eth and the failed eth->kda that needed refund

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].KlvSCCallData = nil
		memeToken.TestOperations[2].KlvFaultySCCall = true
		usdcToken.TestOperations[2].KlvForceSCCall = true

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("0 gas limit should refund", func(t *testing.T) {
		callData := createScCallData("callPayable", 0)
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].KlvSCCallData = callData
		usdcToken.TestOperations[2].KlvFaultySCCall = true
		usdcToken.EthTestAddrExtraBalance = big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50 - 1000 + 950) // -(eth->kda) + (kda->eth) - fees + revert after bad SC call
		usdcToken.KDASafeExtraBalance = big.NewInt(150)                                                  // extra is just for the fees for the 2 transfers kda->eth and the failed eth->kda that needed refund

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].KlvSCCallData = callData
		memeToken.TestOperations[2].KlvFaultySCCall = true

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("small gas limit should refund", func(t *testing.T) {
		callData := createScCallData("callPayable", 2000)
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].KlvSCCallData = callData
		usdcToken.TestOperations[2].KlvFaultySCCall = true
		usdcToken.EthTestAddrExtraBalance = big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50 - 1000 + 950) // -(eth->kda) + (kda->eth) - fees + revert after bad SC call
		usdcToken.KDASafeExtraBalance = big.NewInt(150)                                                  // extra is just for the fees for the 2 transfers kda->eth and the failed eth->kda that needed refund

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].KlvSCCallData = callData
		memeToken.TestOperations[2].KlvFaultySCCall = true

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("extra parameter should refund", func(t *testing.T) {
		callData := createScCallData("callPayable", 50000000, "extra parameter")
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].KlvSCCallData = callData
		usdcToken.TestOperations[2].KlvFaultySCCall = true
		usdcToken.EthTestAddrExtraBalance = big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50 - 1000 + 950) // -(eth->kda) + (kda->eth) - fees + revert after bad SC call
		usdcToken.KDASafeExtraBalance = big.NewInt(150)                                                  // extra is just for the fees for the 2 transfers kda->eth and the failed eth->kda that needed refund

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].KlvSCCallData = callData
		memeToken.TestOperations[2].KlvFaultySCCall = true

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("no arguments should refund", func(t *testing.T) {
		callData := createScCallData("callPayableWithParams", 50000000)
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].KlvSCCallData = callData
		usdcToken.TestOperations[2].KlvFaultySCCall = true
		usdcToken.EthTestAddrExtraBalance = big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50 - 1000 + 950) // -(eth->kda) + (kda->eth) - fees + revert after bad SC call
		usdcToken.KDASafeExtraBalance = big.NewInt(150)                                                  // extra is just for the fees for the 2 transfers kda->eth and the failed eth->kda that needed refund

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].KlvSCCallData = callData
		memeToken.TestOperations[2].KlvFaultySCCall = true

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("wrong number of arguments should refund", func(t *testing.T) {
		callData := createScCallData("callPayableWithParams", 50000000, string([]byte{37}))
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].KlvSCCallData = callData
		usdcToken.TestOperations[2].KlvFaultySCCall = true
		usdcToken.EthTestAddrExtraBalance = big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50 - 1000 + 950) // -(eth->kda) + (kda->eth) - fees + revert after bad SC call
		usdcToken.KDASafeExtraBalance = big.NewInt(150)                                                  // extra is just for the fees for the 2 transfers kda->eth and the failed eth->kda that needed refund

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].KlvSCCallData = callData
		memeToken.TestOperations[2].KlvFaultySCCall = true

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("not an uint64 argument should refund", func(t *testing.T) {
		malformedUint64String := string([]byte{37, 36, 35, 34, 33, 32, 31, 32, 33}) // 9 bytes instead of 8
		dummyAddress := strings.Repeat("2", 32)

		callData := createScCallData("callPayableWithParams", 50000000, malformedUint64String, dummyAddress)
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].KlvSCCallData = callData
		usdcToken.TestOperations[2].KlvFaultySCCall = true
		usdcToken.EthTestAddrExtraBalance = big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50 - 1000 + 950) // -(eth->kda) + (kda->eth) - fees + revert after bad SC call
		usdcToken.KDASafeExtraBalance = big.NewInt(150)                                                  // extra is just for the fees for the 2 transfers kda->eth and the failed eth->kda that needed refund

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].KlvSCCallData = callData
		memeToken.TestOperations[2].KlvFaultySCCall = true

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("wrong arguments encoding should refund", func(t *testing.T) {
		callData := createScCallData("callPayableWithParams", 50000000)
		// the last byte is the data missing marker, we will replace that
		callData[len(callData)-1] = bridgeCore.DataPresentProtocolMarker
		// add garbage data
		callData = append(callData, []byte{5, 4, 55}...)

		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].KlvSCCallData = callData
		usdcToken.TestOperations[2].KlvFaultySCCall = true
		usdcToken.EthTestAddrExtraBalance = big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50 - 1000 + 950) // -(eth->kda) + (kda->eth) - fees + revert after bad SC call
		usdcToken.KDASafeExtraBalance = big.NewInt(150)                                                  // extra is just for the fees for the 2 transfers kda->eth and the failed eth->kda that needed refund

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].KlvSCCallData = callData
		memeToken.TestOperations[2].KlvFaultySCCall = true

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
}

func testRelayersWithChainSimulatorAndTokensAndRefund(tb testing.TB, manualStopChan chan error, tokens ...framework.TestTokenParams) {
	startsFromEthFlow, startsFromKlvFlow := createFlowsBasedOnToken(tb, tokens...)

	setupFunc := func(tb testing.TB, setup *framework.TestSetup) {
		startsFromKlvFlow.setup = setup
		startsFromEthFlow.setup = setup

		setup.IssueAndConfigureTokens(tokens...)
		setup.KleverchainHandler.CheckForZeroBalanceOnReceivers(setup.Ctx, tokens...)
		if len(startsFromEthFlow.tokens) > 0 {
			setup.EthereumHandler.CreateBatchOnEthereum(setup.Ctx, setup.KleverchainHandler.TestCallerAddress, startsFromEthFlow.tokens...)
		}
		if len(startsFromKlvFlow.tokens) > 0 {
			setup.CreateBatchOnKleverchain(startsFromKlvFlow.tokens...)
		}
	}

	processFunc := func(tb testing.TB, setup *framework.TestSetup) bool {
		if startsFromEthFlow.process() && startsFromKlvFlow.process() && startsFromKlvFlow.areTokensFullyRefunded() {
			return true
		}

		// commit blocks in order to execute incoming txs from relayers
		setup.EthereumHandler.SimulatedChain.Commit()
		setup.ChainSimulator.GenerateBlocks(setup.Ctx, 1)
		require.LessOrEqual(tb, setup.ScCallerModuleInstance.GetNumSentTransaction(), setup.GetNumScCallsOperations())

		return false
	}

	_ = testRelayersWithChainSimulator(tb,
		setupFunc,
		processFunc,
		manualStopChan,
	)
}
