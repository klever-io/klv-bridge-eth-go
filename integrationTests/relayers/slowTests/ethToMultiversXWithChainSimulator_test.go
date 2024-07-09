//go:build slow

// To run these slow tests, simply add the slow tag on the go test command. Also, provide a chain simulator instance on the 8085 port
// example: go test -tags slow

package slowTests

import (
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	timeout = time.Minute * 15
)

func TestRelayersShouldExecuteTransfers(t *testing.T) {
	// USDC is ethNative = true, ethMintBurn = false, mvxNative = false, mvxMintBurn = true
	usdcToken := testTokenParams{
		issueTokenParams: issueTokenParams{
			abstractTokenIdentifier:          "USDC",
			numOfDecimalsUniversal:           6,
			numOfDecimalsChainSpecific:       6,
			mvxUniversalTokenTicker:          "USDC",
			mvxChainSpecificTokenTicker:      "ETHUSDC",
			mvxUniversalTokenDisplayName:     "WrappedUSDC",
			mvxChainSpecificTokenDisplayName: "EthereumWrappedUSDC",
			valueToMintOnMvx:                 "10000000000",
			isMintBurnOnMvX:                  true,
			isNativeOnMvX:                    false,
			ethTokenName:                     "ETHTOKEN",
			ethTokenSymbol:                   "ETHT",
			valueToMintOnEth:                 "10000000000",
			isMintBurnOnEth:                  false,
			isNativeOnEth:                    true,
		},
		testOperations: []tokenOperations{
			{
				valueToTransferToMvx: big.NewInt(5000),
				valueToSendFromMvX:   big.NewInt(2500),
				ethSCCallMethod:      "",
				ethSCCallGasLimit:    0,
				ethSCCallArguments:   nil,
			},
			{
				valueToTransferToMvx: big.NewInt(7000),
				valueToSendFromMvX:   big.NewInt(300),
				ethSCCallMethod:      "",
				ethSCCallGasLimit:    0,
				ethSCCallArguments:   nil,
			},
			{
				valueToTransferToMvx: big.NewInt(1000),
				valueToSendFromMvX:   nil,
				ethSCCallMethod:      "callPayable",
				ethSCCallGasLimit:    50000000,
				ethSCCallArguments:   nil,
			},
		},
		esdtSafeExtraBalance:    big.NewInt(100),                                        // extra is just for the fees for the 2 transfers mvx->eth
		ethTestAddrExtraBalance: big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50 - 1000), // -(eth->mvx) + (mvx->eth) - fees
	}

	//MEME is ethNative = false, ethMintBurn = true, mvxNative = true, mvxMintBurn = false
	memeToken := testTokenParams{
		issueTokenParams: issueTokenParams{
			abstractTokenIdentifier:          "MEME",
			numOfDecimalsUniversal:           1,
			numOfDecimalsChainSpecific:       1,
			mvxUniversalTokenTicker:          "MEME",
			mvxChainSpecificTokenTicker:      "ETHMEME",
			mvxUniversalTokenDisplayName:     "WrappedMEME",
			mvxChainSpecificTokenDisplayName: "EthereumWrappedMEME",
			valueToMintOnMvx:                 "10000000000",
			isMintBurnOnMvX:                  false,
			isNativeOnMvX:                    true,
			ethTokenName:                     "ETHMEME",
			ethTokenSymbol:                   "ETHM",
			valueToMintOnEth:                 "10000000000",
			isMintBurnOnEth:                  true,
			isNativeOnEth:                    false,
		},
		testOperations: []tokenOperations{
			{
				valueToTransferToMvx: big.NewInt(2400),
				valueToSendFromMvX:   big.NewInt(4000),
				ethSCCallMethod:      "",
				ethSCCallGasLimit:    0,
				ethSCCallArguments:   nil,
			},
			{
				valueToTransferToMvx: big.NewInt(200),
				valueToSendFromMvX:   big.NewInt(6000),
				ethSCCallMethod:      "",
				ethSCCallGasLimit:    0,
				ethSCCallArguments:   nil,
			},
			{
				valueToTransferToMvx: big.NewInt(1000),
				valueToSendFromMvX:   big.NewInt(2000),
				ethSCCallMethod:      "callPayable",
				ethSCCallGasLimit:    50000000,
				ethSCCallArguments:   nil,
			},
		},
		esdtSafeExtraBalance:    big.NewInt(4000 + 6000 + 2000), // everything is locked in the safe esdt contract
		ethTestAddrExtraBalance: big.NewInt(4000 - 50 + 6000 - 50 + 2000 - 50),
	}

	testRelayersWithChainSimulatorAndTokens(t, usdcToken, memeToken)
}

func testRelayersWithChainSimulatorAndTokens(tb testing.TB, tokens ...testTokenParams) {
	startsFromEthFlow := &startsFromEthereumFlow{
		TB:     tb,
		tokens: make([]testTokenParams, 0, len(tokens)),
	}

	startsFromMvXFlow := &startsFromMultiversXFlow{
		TB:     tb,
		tokens: make([]testTokenParams, 0, len(tokens)),
	}

	// split the tokens from where should the bridge start
	for _, token := range tokens {
		if token.isNativeOnEth {
			startsFromEthFlow.tokens = append(startsFromEthFlow.tokens, token)
			continue
		}
		if token.isNativeOnMvX {
			startsFromMvXFlow.tokens = append(startsFromMvXFlow.tokens, token)
			continue
		}
		require.Fail(tb, "invalid setup, found a token that is not native on any chain", "abstract identifier", token.abstractTokenIdentifier)
	}

	setupFunc := func(tb testing.TB, testSetup *simulatedSetup) {
		startsFromMvXFlow.testSetup = testSetup
		startsFromEthFlow.testSetup = testSetup

		testSetup.issueAndConfigureTokens(tokens...)
		testSetup.checkForZeroBalanceOnReceivers(tokens...)
		if len(startsFromEthFlow.tokens) > 0 {
			testSetup.createBatchOnEthereum(startsFromEthFlow.tokens...)
		}
		if len(startsFromMvXFlow.tokens) > 0 {
			testSetup.createBatchOnMultiversX(startsFromMvXFlow.tokens...)
		}
	}

	processFunc := func(tb testing.TB, testSetup *simulatedSetup, stopChan chan os.Signal) bool {
		// TODO: remove select
		select {
		default:
			if startsFromEthFlow.process() && startsFromMvXFlow.process() {
				return true
			}

		case <-stopChan:
			require.Fail(tb, "signal interrupted")
			return true
		case <-time.After(timeout):
			require.Fail(tb, "time out")
			return true
		}

		// commit blocks in order to execute incoming txs from relayers
		testSetup.simulatedETHChain.Commit()
		testSetup.mvxChainSimulator.GenerateBlocks(testSetup.testContext, 1)

		return false
	}

	testRelayersWithChainSimulator(tb,
		setupFunc,
		processFunc,
	)
}

func testRelayersWithChainSimulator(tb testing.TB,
	setupFunc func(tb testing.TB, testSetup *simulatedSetup),
	processLoopFunc func(tb testing.TB, testSetup *simulatedSetup) bool,
) {
	defer func() {
		r := recover()
		if r != nil {
			require.Fail(tb, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	testSetup := prepareSimulatedSetup(tb)
	log.Info(fmt.Sprintf(logStepMarker, "calling setupFunc"))
	setupFunc(tb, testSetup)

	testSetup.startRelayersAndScModule()
	defer testSetup.close()

	log.Info(fmt.Sprintf(logStepMarker, "running and continously call processLoopFunc"))
	interrupt := make(chan os.Signal, 1)
	for {
		select {
		case <-interrupt:
			require.Fail(tb, "signal interrupted")
			return
		case <-time.After(timeout):
			require.Fail(tb, "time out")
			return
		default:
			testDone := processLoopFunc(tb, testSetup)
			if testDone {
				return
			}
		}
	}
}

// TODO: next PRs: fix these tests
//func TestRelayersShouldExecuteTransfers(t *testing.T) {
//	t.Run("ETH->MVX and back, ethNative = true, ethMintBurn = false, mvxNative = false, mvxMintBurn = true", func(t *testing.T) {
//		args := argSimulatedSetup{
//			mvxIsMintBurn:        true,
//			mvxIsNative:          false,
//			ethIsMintBurn:        false,
//			ethIsNative:          true,
//			transferBackAndForth: true,
//		}
//		testRelayersShouldExecuteTransfersEthToMVX(t, args)
//	})
//	t.Run("MVX->ETH, ethNative = false, ethMintBurn = true, mvxNative = true, mvxMintBurn = false", func(t *testing.T) {
//		args := argSimulatedSetup{
//			mvxIsMintBurn:        false,
//			mvxIsNative:          true,
//			ethIsMintBurn:        true,
//			ethIsNative:          false,
//			transferBackAndForth: true,
//		}
//		testRelayersShouldExecuteTransfersMVXToETH(t, args)
//	})
//	t.Run("ETH->MVX with SC call that works, ethNative = true, ethMintBurn = false, mvxNative = false, mvxMintBurn = true", func(t *testing.T) {
//		args := argSimulatedSetup{
//			mvxIsMintBurn:        true,
//			mvxIsNative:          false,
//			ethIsMintBurn:        false,
//			ethIsNative:          true,
//			ethSCCallMethod:      "callPayable",
//			ethSCCallGasLimit:    50000000,
//			ethSCCallArguments:   nil,
//			transferBackAndForth: false,
//		}
//		testRelayersShouldExecuteTransfersEthToMVX(t, args)
//	})
//}
//
//func testRelayersShouldExecuteTransfersEthToMVX(t *testing.T, argsSimulatedSetup argSimulatedSetup) {
//	defer func() {
//		r := recover()
//		if r != nil {
//			require.Fail(t, "should have not panicked")
//		}
//	}()
//
//	argsSimulatedSetup.t = t
//	testSetup := prepareSimulatedSetup(argsSimulatedSetup)
//	defer testSetup.close()
//
//	testSetup.checkESDTBalance(testSetup.mvxReceiverAddress, testSetup.mvxUniversalToken, "0", true)
//
//	testSetup.createBatch(batchProcessor.ToMultiversX)
//
//	interrupt := make(chan os.Signal, 1)
//	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
//	ethToMVXDone := false
//	mvxToETHDone := false
//
//	safeAddr, err := data.NewAddressFromBech32String(testSetup.mvxSafeAddress)
//	require.NoError(t, err)
//
//	// send half of the amount back to ETH
//	valueToSendFromMVX := big.NewInt(0).Div(mintAmount, big.NewInt(2))
//	initialSafeValue, err := testSetup.mvxChainSimulator.GetESDTBalance(testSetup.testContext, safeAddr, testSetup.mvxChainSpecificToken)
//	require.NoError(t, err)
//	initialSafeValueInt, _ := big.NewInt(0).SetString(initialSafeValue, 10)
//	expectedFinalValueOnMVXSafe := initialSafeValueInt.Add(initialSafeValueInt, feeInt)
//	expectedFinalValueOnETH := big.NewInt(0).Sub(valueToSendFromMVX, feeInt)
//	for {
//		select {
//		case <-interrupt:
//			require.Fail(t, "signal interrupted")
//			return
//		case <-time.After(timeout):
//			require.Fail(t, "time out")
//			return
//		default:
//			receiverToCheckBalance := testSetup.mvxReceiverAddress
//			if len(testSetup.ethSCCallMethod) > 0 {
//				receiverToCheckBalance = testSetup.mvxTestCallerAddress
//			}
//
//			isTransferDoneFromETH := testSetup.checkESDTBalance(receiverToCheckBalance, testSetup.mvxUniversalToken, mintAmount.String(), false)
//			if !ethToMVXDone && isTransferDoneFromETH {
//				ethToMVXDone = true
//
//				if argsSimulatedSetup.transferBackAndForth {
//					log.Info("ETH->MvX transfer finished, now sending back to ETH...")
//
//					testSetup.sendMVXToEthTransaction(valueToSendFromMVX.Bytes())
//				} else {
//					log.Info("ETH->MvX transfers done")
//					return
//				}
//			}
//
//			isTransferDoneFromMVX := testSetup.checkETHStatus(testSetup.ethOwnerAddress, expectedFinalValueOnETH.Uint64())
//			safeSavedFee := testSetup.checkESDTBalance(safeAddr, testSetup.mvxChainSpecificToken, expectedFinalValueOnMVXSafe.String(), false)
//			if !mvxToETHDone && isTransferDoneFromMVX && safeSavedFee {
//				mvxToETHDone = true
//			}
//
//			if ethToMVXDone && mvxToETHDone {
//				log.Info("MvX<->ETH transfers done")
//				return
//			}
//
//			// commit blocks in order to execute incoming txs from relayers
//			testSetup.simulatedETHChain.Commit()
//
//			testSetup.mvxChainSimulator.GenerateBlocks(testSetup.testContext, 1)
//
//		case <-interrupt:
//			require.Fail(t, "signal interrupted")
//			return
//		case <-time.After(timeout):
//			require.Fail(t, "time out")
//			return
//		}
//	}
//}
//
//func testRelayersShouldExecuteTransfersMVXToETH(t *testing.T, argsSimulatedSetup argSimulatedSetup) {
//	defer func() {
//		r := recover()
//		if r != nil {
//			require.Fail(t, "should have not panicked")
//		}
//	}()
//
//	argsSimulatedSetup.t = t
//	testSetup := prepareSimulatedSetup(argsSimulatedSetup)
//	defer testSetup.close()
//
//	testSetup.checkESDTBalance(testSetup.mvxReceiverAddress, testSetup.mvxUniversalToken, "0", true)
//
//	safeAddr, err := data.NewAddressFromBech32String(testSetup.mvxSafeAddress)
//	require.NoError(t, err)
//
//	initialSafeValue, err := testSetup.mvxChainSimulator.GetESDTBalance(testSetup.testContext, safeAddr, testSetup.mvxChainSpecificToken)
//	require.NoError(t, err)
//
//	testSetup.createBatch(batchProcessor.FromMultiversX)
//
//	// wait for signal interrupt or time out
//	interrupt := make(chan os.Signal, 1)
//	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
//
//	// send half of the amount back to ETH
//	valueSentFromETH := big.NewInt(0).Div(mintAmount, big.NewInt(2))
//	initialSafeValueInt, _ := big.NewInt(0).SetString(initialSafeValue, 10)
//	expectedFinalValueOnMVXSafe := initialSafeValueInt.Add(initialSafeValueInt, valueSentFromETH)
//	expectedFinalValueOnETH := big.NewInt(0).Sub(valueSentFromETH, feeInt)
//	expectedFinalValueOnETH = expectedFinalValueOnETH.Mul(expectedFinalValueOnETH, big.NewInt(1000000))
//	for {
//		select {
//		case <-interrupt:
//			require.Fail(t, "signal interrupted")
//			return
//		case <-time.After(timeout):
//			require.Fail(t, "time out")
//			return
//		default:
//			isTransferDoneFromMVX := testSetup.checkETHStatus(testSetup.ethOwnerAddress, expectedFinalValueOnETH.Uint64())
//			safeSavedFunds := testSetup.checkESDTBalance(safeAddr, testSetup.mvxChainSpecificToken, expectedFinalValueOnMVXSafe.String(), false)
//			if isTransferDoneFromMVX && safeSavedFunds {
//				log.Info("MVX->ETH transfer finished")
//
//				return
//			}
//
//			// commit blocks in order to execute incoming txs from relayers
//			testSetup.simulatedETHChain.Commit()
//
//			testSetup.mvxChainSimulator.GenerateBlocks(testSetup.testContext, 1)
//		}
//	}
//}
//
//func TestRelayersShouldNotExecuteTransfers(t *testing.T) {
//	t.Run("ETH->MVX, ethNative = true, ethMintBurn = false, mvxNative = true, mvxMintBurn = false", func(t *testing.T) {
//		args := argSimulatedSetup{
//			mvxIsMintBurn: false,
//			mvxIsNative:   true,
//			ethIsMintBurn: false,
//			ethIsNative:   true,
//		}
//		expectedStringInLogs := "error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true"
//		testRelayersShouldNotExecuteTransfers(t, args, expectedStringInLogs, batchProcessor.ToMultiversX)
//	})
//	t.Run("ETH->MVX, ethNative = true, ethMintBurn = false, mvxNative = true, mvxMintBurn = true", func(t *testing.T) {
//		args := argSimulatedSetup{
//			mvxIsMintBurn: true,
//			mvxIsNative:   true,
//			ethIsMintBurn: false,
//			ethIsNative:   true,
//		}
//		expectedStringInLogs := "error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true"
//		testRelayersShouldNotExecuteTransfers(t, args, expectedStringInLogs, batchProcessor.ToMultiversX)
//	})
//	t.Run("ETH->MVX, ethNative = true, ethMintBurn = true, mvxNative = true, mvxMintBurn = false", func(t *testing.T) {
//		args := argSimulatedSetup{
//			mvxIsMintBurn: false,
//			mvxIsNative:   true,
//			ethIsMintBurn: true,
//			ethIsNative:   true,
//		}
//		testEthContractsShouldError(t, args)
//	})
//	t.Run("ETH->MVX, ethNative = true, ethMintBurn = true, mvxNative = true, mvxMintBurn = true", func(t *testing.T) {
//		args := argSimulatedSetup{
//			mvxIsMintBurn: true,
//			mvxIsNative:   true,
//			ethIsMintBurn: true,
//			ethIsNative:   true,
//		}
//		testEthContractsShouldError(t, args)
//	})
//	t.Run("ETH->MVX, ethNative = false, ethMintBurn = true, mvxNative = false, mvxMintBurn = true", func(t *testing.T) {
//		args := argSimulatedSetup{
//			mvxIsMintBurn: true,
//			mvxIsNative:   false,
//			ethIsMintBurn: true,
//			ethIsNative:   false,
//		}
//		testEthContractsShouldError(t, args)
//	})
//	t.Run("MVX->ETH, ethNative = true, ethMintBurn = false, mvxNative = true, mvxMintBurn = false", func(t *testing.T) {
//		args := argSimulatedSetup{
//			mvxIsMintBurn: false,
//			mvxIsNative:   true,
//			ethIsMintBurn: false,
//			ethIsNative:   true,
//		}
//		expectedStringInLogs := "error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true"
//		testRelayersShouldNotExecuteTransfers(t, args, expectedStringInLogs, batchProcessor.FromMultiversX)
//	})
//}
//
//func testRelayersShouldNotExecuteTransfers(
//	t *testing.T,
//	argsSimulatedSetup argSimulatedSetup,
//	expectedStringInLogs string,
//	direction batchProcessor.Direction,
//) {
//	defer func() {
//		r := recover()
//		if r != nil {
//			require.Fail(t, "should have not panicked")
//		}
//	}()
//
//	argsSimulatedSetup.t = t
//	testSetup := prepareSimulatedSetup(argsSimulatedSetup)
//	defer testSetup.close()
//
//	testSetup.checkESDTBalance(testSetup.mvxReceiverAddress, testSetup.mvxUniversalToken, "0", true)
//
//	testSetup.createBatch(direction)
//
//	// start a mocked log observer that is looking for a specific relayer error
//	chanCnt := 0
//	mockLogObserver := mock.NewMockLogObserver(expectedStringInLogs)
//	err := logger.AddLogObserver(mockLogObserver, &logger.PlainFormatter{})
//	require.NoError(t, err)
//	defer func() {
//		require.NoError(t, logger.RemoveLogObserver(mockLogObserver))
//	}()
//
//	numOfTimesToRepeatErrorForRelayer := 10
//	numOfErrorsToWait := numOfTimesToRepeatErrorForRelayer * numRelayers
//
//	// wait for signal interrupt or time out
//	roundDuration := time.Second
//	roundTimer := time.NewTimer(roundDuration)
//	interrupt := make(chan os.Signal, 1)
//	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
//
//	for {
//		roundTimer.Reset(roundDuration)
//		select {
//		case <-interrupt:
//			require.Fail(t, "signal interrupted")
//			return
//		case <-time.After(timeout):
//			require.Fail(t, "time out")
//			return
//		case <-mockLogObserver.LogFoundChan():
//			chanCnt++
//			if chanCnt >= numOfErrorsToWait {
//				testSetup.checkESDTBalance(testSetup.mvxReceiverAddress, testSetup.mvxUniversalToken, "0", true)
//
//				log.Info(fmt.Sprintf("test passed, relayers are stuck, expected string `%s` found in all relayers' logs for %d times", expectedStringInLogs, numOfErrorsToWait))
//
//				return
//			}
//		case <-roundTimer.C:
//			// commit blocks
//			testSetup.simulatedETHChain.Commit()
//
//			testSetup.mvxChainSimulator.GenerateBlocks(testSetup.testContext, 1)
//		}
//	}
//}
//
//func testEthContractsShouldError(t *testing.T, argsSimulatedSetup argSimulatedSetup) {
//	defer func() {
//		r := recover()
//		if r != nil {
//			require.Fail(t, "should have not panicked")
//		}
//	}()
//
//	testSetup := &simulatedSetup{}
//	testSetup.T = t
//
//	// create a test context
//	testSetup.testContext, testSetup.testContextCancel = context.WithCancel(context.Background())
//
//	testSetup.workingDir = t.TempDir()
//
//	testSetup.generateKeys()
//
//	receiverKeys := generateMvxPrivatePublicKey(t)
//	mvxReceiverAddress, err := data.NewAddressFromBech32String(receiverKeys.pk)
//	require.NoError(t, err)
//
//	testSetup.ethOwnerAddress = crypto.PubkeyToAddress(ethOwnerSK.PublicKey)
//	ethDepositorAddr := crypto.PubkeyToAddress(ethDepositorSK.PublicKey)
//
//	// create ethereum simulator
//	testSetup.createEthereumSimulatorAndDeployContracts(ethDepositorAddr, argsSimulatedSetup.ethIsMintBurn, argsSimulatedSetup.ethIsNative)
//
//	// add allowance for the sender
//	auth, _ := bind.NewKeyedTransactorWithChainID(ethDepositorSK, testSetup.ethChainID)
//	tx, err := testSetup.ethGenericTokenContract.Approve(auth, testSetup.ethSafeAddress, mintAmount)
//	require.NoError(t, err)
//	testSetup.simulatedETHChain.Commit()
//	testSetup.checkEthTxResult(tx.Hash())
//
//	// deposit on ETH safe should fail due to bad setup
//	auth, _ = bind.NewKeyedTransactorWithChainID(ethDepositorSK, testSetup.ethChainID)
//	_, err = testSetup.ethSafeContract.Deposit(auth, testSetup.ethGenericTokenAddress, mintAmount, mvxReceiverAddress.AddressSlice())
//	require.Error(t, err)
//}
