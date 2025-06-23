package framework

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/klever-io/klv-bridge-eth-go/config"
	"github.com/klever-io/klv-bridge-eth-go/executors/kleverBlockchain/module"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/stretchr/testify/require"
)

// framework constants
const (
	LogStepMarker                = "#################################### %s ####################################"
	proxyCacherExpirationSeconds = 600
	proxyMaxNoncesDelta          = 7
	NumRelayers                  = 3
	NumOracles                   = 3
	quorum                       = "03"
)

// TestSetup is the struct that holds all subcomponents for the testing infrastructure
type TestSetup struct {
	testing.TB
	TokensRegistry
	*KeysStore
	Bridge                 *BridgeComponents
	EthereumHandler        *EthereumHandler
	KCHandler              *KCHandler
	WorkingDir             string
	ChainSimulator         ChainSimulatorWrapper
	ScCallerKeys           KeysHolder
	ScCallerModuleInstance SCCallerModule

	ctxCancel             func()
	Ctx                   context.Context
	mutBalances           sync.RWMutex
	kdaBalanceForSafe     map[string]*big.Int
	ethBalanceTestAddress map[string]*big.Int
	numScCallsInTest      uint32
}

// NewTestSetup creates a new e2e test setup
func NewTestSetup(tb testing.TB) *TestSetup {
	log.Info(fmt.Sprintf(LogStepMarker, "starting setup"))

	setup := &TestSetup{
		TB:                    tb,
		TokensRegistry:        NewTokenRegistry(tb),
		WorkingDir:            tb.TempDir(),
		kdaBalanceForSafe:     make(map[string]*big.Int),
		ethBalanceTestAddress: make(map[string]*big.Int),
	}
	setup.KeysStore = NewKeysStore(tb, setup.WorkingDir, NumRelayers, NumOracles)

	// create a test context
	setup.Ctx, setup.ctxCancel = context.WithCancel(context.Background())

	setup.EthereumHandler = NewEthereumHandler(tb, setup.Ctx, setup.KeysStore, setup.TokensRegistry, quorum)
	setup.EthereumHandler.DeployContracts(setup.Ctx)

	setup.createChainSimulatorWrapper()
	setup.KCHandler = NewKCHandler(tb, setup.Ctx, setup.KeysStore, setup.TokensRegistry, setup.ChainSimulator, quorum)
	setup.KCHandler.DeployAndSetContracts(setup.Ctx)

	return setup
}

func (setup *TestSetup) createChainSimulatorWrapper() {
	// create a new working directory
	tmpDir := path.Join(setup.TempDir(), "test")
	err := os.MkdirAll(tmpDir, os.ModePerm)
	require.NoError(setup, err)

	// start the chain simulator
	args := ArgChainSimulatorWrapper{
		TB:                           setup.TB,
		ProxyCacherExpirationSeconds: proxyCacherExpirationSeconds,
		ProxyMaxNoncesDelta:          proxyMaxNoncesDelta,
	}
	setup.ChainSimulator = CreateChainSimulatorWrapper(args)
	require.NoError(setup, err)
}

// StartRelayersAndScModule will start the bridge and the SC execution module
func (setup *TestSetup) StartRelayersAndScModule() {
	log.Info(fmt.Sprintf(LogStepMarker, "starting relayers & sc execution module"))

	// start relayers
	setup.Bridge = NewBridgeComponents(
		setup.TB,
		setup.WorkingDir,
		setup.ChainSimulator,
		setup.EthereumHandler.EthChainWrapper,
		setup.EthereumHandler.Erc20ContractsHolder,
		setup.EthereumHandler.SimulatedChain,
		NumRelayers,
		setup.EthereumHandler.SafeAddress.Hex(),
		setup.KCHandler.SafeAddress,
		setup.KCHandler.MultisigAddress,
	)

	setup.startScCallerModule()
}

func (setup *TestSetup) startScCallerModule() {
	cfg := config.ScCallsModuleConfig{
		ScProxyBech32Address:            setup.KCHandler.ScProxyAddress.Bech32(),
		ExtraGasToExecute:               60_000_000,  // 60 million: this ensures that a SC call with 0 gas limit is refunded
		MaxGasLimitToUse:                249_999_999, // max cross shard limit
		GasLimitForOutOfGasTransactions: 30_000_000,  // gas to use when a higher than max allowed is encountered
		NetworkAddress:                  setup.ChainSimulator.GetNetworkAddress(),
		ProxyMaxNoncesDelta:             5,
		ProxyFinalityCheck:              false,
		ProxyCacherExpirationSeconds:    60, // 1 minute
		ProxyRestAPIEntityType:          string(sdkCore.Proxy),
		IntervalToResendTxsInSeconds:    1,
		PrivateKeyFile:                  path.Join(setup.WorkingDir, SCCallerFilename),
		PollingIntervalInMillis:         1000, // 1 second
		Filter: config.PendingOperationsFilterConfig{
			AllowedEthAddresses: []string{"*"},
			AllowedKlvAddresses: []string{"*"},
			AllowedTokens:       []string{"*"},
		},
		TransactionChecks: config.TransactionChecksConfig{
			CheckTransactionResults:    true,
			CloseAppOnError:            false,
			ExecutionTimeoutInSeconds:  2,
			TimeInSecondsBetweenChecks: 1,
		},
	}

	var err error
	setup.ScCallerModuleInstance, err = module.NewScCallsModule(cfg, log, nil)
	require.Nil(setup, err)
	log.Info("started SC calls module", "monitoring SC proxy address", setup.KCHandler.ScProxyAddress)
}

// IssueAndConfigureTokens will issue and configure the provided tokens on both chains
func (setup *TestSetup) IssueAndConfigureTokens(tokens ...TestTokenParams) {
	log.Info(fmt.Sprintf(LogStepMarker, fmt.Sprintf("issuing %d tokens", len(tokens))))

	require.Greater(setup, len(tokens), 0)

	setup.EthereumHandler.PauseContractsForTokenChanges(setup.Ctx)
	setup.KCHandler.PauseContractsForTokenChanges(setup.Ctx)

	for _, token := range tokens {
		setup.processNumScCallsOperations(token)
		setup.AddToken(token.IssueTokenParams)
		setup.EthereumHandler.IssueAndWhitelistToken(setup.Ctx, token.IssueTokenParams)
		setup.KCHandler.IssueAndWhitelistToken(setup.Ctx, token.IssueTokenParams)

		kdaBalanceForSafe := setup.KCHandler.GetKDAChainSpecificTokenBalance(setup.Ctx, setup.KCHandler.SafeAddress, token.AbstractTokenIdentifier)
		ethBalanceForTestAddr := setup.EthereumHandler.GetBalance(setup.TestKeys.EthAddress, token.AbstractTokenIdentifier)

		setup.mutBalances.Lock()
		setup.kdaBalanceForSafe[token.AbstractTokenIdentifier] = kdaBalanceForSafe
		setup.ethBalanceTestAddress[token.AbstractTokenIdentifier] = ethBalanceForTestAddr
		setup.mutBalances.Unlock()

		log.Info("recorded the KDA balance for safe contract", "token", token.AbstractTokenIdentifier, "balance", kdaBalanceForSafe.String())
		log.Info("recorded the ETH balance for test address", "token", token.AbstractTokenIdentifier, "balance", ethBalanceForTestAddr.String())
	}

	setup.EthereumHandler.UnPauseContractsAfterTokenChanges(setup.Ctx)
	setup.KCHandler.UnPauseContractsAfterTokenChanges(setup.Ctx)

	for _, token := range tokens {
		setup.KCHandler.SubmitAggregatorBatch(setup.Ctx, token.IssueTokenParams)
	}
}

func (setup *TestSetup) processNumScCallsOperations(token TestTokenParams) {
	for _, op := range token.TestOperations {
		if len(op.KlvSCCallData) > 0 || op.KlvForceSCCall {
			atomic.AddUint32(&setup.numScCallsInTest, 1)
		}
	}
}

// GetNumScCallsOperations returns the number of SC calls in this test setup
func (setup *TestSetup) GetNumScCallsOperations() uint32 {
	return atomic.LoadUint32(&setup.numScCallsInTest)
}

// IsTransferDoneFromEthereum returns true if all provided tokens are bridged from Ethereum towards KC
func (setup *TestSetup) IsTransferDoneFromEthereum(tokens ...TestTokenParams) bool {
	isDone := true
	for _, params := range tokens {
		isDone = isDone && setup.isTransferDoneFromEthereumForToken(params)
	}

	return isDone
}

func (setup *TestSetup) isTransferDoneFromEthereumForToken(params TestTokenParams) bool {
	expectedValueOnReceiver := big.NewInt(0)
	expectedValueOnContract := big.NewInt(0)
	for _, operation := range params.TestOperations {
		if operation.ValueToTransferToKlv == nil {
			continue
		}

		if len(operation.KlvSCCallData) > 0 || operation.KlvForceSCCall {
			if !operation.KlvFaultySCCall {
				expectedValueOnContract.Add(expectedValueOnContract, operation.ValueToTransferToKlv)
			}
		} else {
			expectedValueOnReceiver.Add(expectedValueOnReceiver, operation.ValueToTransferToKlv)
		}
	}

	receiverBalance := setup.KCHandler.GetKDAUniversalTokenBalance(setup.Ctx, setup.TestKeys.KlvAddress, params.AbstractTokenIdentifier)
	if receiverBalance.String() != expectedValueOnReceiver.String() {
		return false
	}

	contractBalance := setup.KCHandler.GetKDAUniversalTokenBalance(setup.Ctx, setup.KCHandler.TestCallerAddress, params.AbstractTokenIdentifier)
	return contractBalance.String() == expectedValueOnContract.String()
}

// IsTransferDoneFromEthereumWithRefund returns true if all provided tokens are bridged from Ethereum towards Klever Blockchain including refunds
func (setup *TestSetup) IsTransferDoneFromEthereumWithRefund(tokens ...TestTokenParams) bool {
	isDone := true
	for _, params := range tokens {
		isDone = isDone && setup.isTransferDoneFromEthereumWithRefundForToken(params)
	}

	return isDone
}

func (setup *TestSetup) isTransferDoneFromEthereumWithRefundForToken(params TestTokenParams) bool {
	expectedValueOnReceiver := big.NewInt(0)
	for _, operation := range params.TestOperations {
		valueToTransferToKlv := big.NewInt(0)
		if operation.ValueToTransferToKlv != nil {
			valueToTransferToKlv.Set(operation.ValueToTransferToKlv)
		}

		valueToSendFromKlv := big.NewInt(0)
		if operation.ValueToSendFromKlv != nil {
			valueToSendFromKlv.Set(operation.ValueToSendFromKlv)
			// we subtract the fee also
			expectedValueOnReceiver.Sub(expectedValueOnReceiver, feeInt)
		}

		expectedValueOnReceiver.Add(expectedValueOnReceiver, big.NewInt(0).Sub(valueToSendFromKlv, valueToTransferToKlv))
		if len(operation.KlvSCCallData) > 0 || operation.KlvForceSCCall {
			if operation.KlvFaultySCCall {
				// the balance should be bridged back to the receiver on Ethereum - fee
				expectedValueOnReceiver.Add(expectedValueOnReceiver, valueToTransferToKlv)
				expectedValueOnReceiver.Sub(expectedValueOnReceiver, feeInt)
			}
		}
	}

	receiverBalance := setup.EthereumHandler.GetBalance(setup.TestKeys.EthAddress, params.AbstractTokenIdentifier)
	return receiverBalance.String() == expectedValueOnReceiver.String()
}

// IsTransferDoneFromKlever Blockchain returns true if all provided tokens are bridged from Klever Blockchain towards Ethereum
func (setup *TestSetup) IsTransferDoneFromKC(tokens ...TestTokenParams) bool {
	isDone := true
	for _, params := range tokens {
		isDone = isDone && setup.isTransferDoneFromKCForToken(params)
	}

	return isDone
}

func (setup *TestSetup) isTransferDoneFromKCForToken(params TestTokenParams) bool {
	setup.mutBalances.Lock()
	initialBalanceForSafe := setup.kdaBalanceForSafe[params.AbstractTokenIdentifier]
	expectedReceiver := big.NewInt(0).Set(setup.ethBalanceTestAddress[params.AbstractTokenIdentifier])
	expectedReceiver.Add(expectedReceiver, params.EthTestAddrExtraBalance)
	setup.mutBalances.Unlock()

	ethTestBalance := setup.EthereumHandler.GetBalance(setup.TestKeys.EthAddress, params.AbstractTokenIdentifier)
	isTransferDoneFromKC := ethTestBalance.String() == expectedReceiver.String()

	expectedKdaSafe := big.NewInt(0).Add(initialBalanceForSafe, params.KDASafeExtraBalance)
	balanceForSafe := setup.KCHandler.GetKDAChainSpecificTokenBalance(setup.Ctx, setup.KCHandler.SafeAddress, params.AbstractTokenIdentifier)
	isSafeContractOnCorrectBalance := expectedKdaSafe.String() == balanceForSafe.String()

	return isTransferDoneFromKC && isSafeContractOnCorrectBalance
}

// CreateBatchOnKC will create deposits that will be gathered in a batch on KC
func (setup *TestSetup) CreateBatchOnKC(tokensParams ...TestTokenParams) {
	for _, params := range tokensParams {
		setup.createBatchOnKCForToken(params)
	}
}

func (setup *TestSetup) createBatchOnKCForToken(params TestTokenParams) {
	token := setup.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(setup, token)

	setup.transferTokensToTestKey(params)
	valueToMintOnEthereum := setup.sendFromKCToEthereumForToken(params)
	setup.EthereumHandler.Mint(setup.Ctx, params, valueToMintOnEthereum)
}

func (setup *TestSetup) transferTokensToTestKey(params TestTokenParams) {
	depositValue := big.NewInt(0)
	for _, operation := range params.TestOperations {
		if operation.ValueToSendFromKlv == nil {
			continue
		}

		depositValue.Add(depositValue, operation.ValueToSendFromKlv)
	}

	setup.KCHandler.TransferToken(
		setup.Ctx,
		setup.OwnerKeys,
		setup.TestKeys,
		depositValue,
		params,
	)
}

// SendFromKCToEthereum will create the deposits that will be gathered in a batch on Klever Blockchain (without mint on Ethereum)
func (setup *TestSetup) SendFromKCToEthereum(tokensParams ...TestTokenParams) {
	for _, params := range tokensParams {
		_ = setup.sendFromKCToEthereumForToken(params)
	}
}

func (setup *TestSetup) sendFromKCToEthereumForToken(params TestTokenParams) *big.Int {
	token := setup.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(setup, token)

	depositValue := big.NewInt(0)
	for _, operation := range params.TestOperations {
		if operation.ValueToSendFromKlv == nil {
			continue
		}

		depositValue.Add(depositValue, operation.ValueToSendFromKlv)
		setup.KCHandler.SendDepositTransactionFromKC(setup.Ctx, token, params, operation.ValueToSendFromKlv)
	}

	return depositValue
}

// TestWithdrawTotalFeesOnEthereumForTokens will test the withdrawal functionality for the provided test tokens
func (setup *TestSetup) TestWithdrawTotalFeesOnEthereumForTokens(tokensParams ...TestTokenParams) {
	for _, param := range tokensParams {
		token := setup.TokensRegistry.GetTokenData(param.AbstractTokenIdentifier)

		expectedAccumulated := big.NewInt(0)
		for _, operation := range param.TestOperations {
			if operation.ValueToSendFromKlv == nil {
				continue
			}
			if operation.ValueToSendFromKlv.Cmp(zeroValueBigInt) == 0 {
				continue
			}

			expectedAccumulated.Add(expectedAccumulated, feeInt)
		}

		setup.KCHandler.TestWithdrawFees(setup.Ctx, token.KlvChainSpecificToken, zeroValueBigInt, expectedAccumulated)
	}
}

// Close will close the test subcomponents
func (setup *TestSetup) Close() {
	log.Info(fmt.Sprintf(LogStepMarker, "closing relayers & sc execution module"))

	setup.Bridge.CloseRelayers()
	require.NoError(setup, setup.EthereumHandler.Close())

	setup.ctxCancel()
	_ = setup.ScCallerModuleInstance.Close()
}
