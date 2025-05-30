package framework

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/require"
)

const (
	minRelayerStake          = "10000000000000000000" // 10 EGLD
	kdaIssueCost             = "50000000000000000"    // 0.05 EGLD
	emptyAddress             = "erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu"
	kdaSystemSCAddress       = "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"
	slashAmount              = "00"
	zeroStringValue          = "0"
	canAddSpecialRoles       = "canAddSpecialRoles"
	trueStr                  = "true"
	kdaRoleLocalMint         = "KDARoleLocalMint"
	kdaRoleLocalBurn         = "KDARoleLocalBurn"
	hexTrue                  = "01"
	hexFalse                 = "00"
	gwei                     = "GWEI"
	maxBridgedAmountForToken = "500000"
	deployGasLimit           = 150000000 // 150 million
	setCallsGasLimit         = 80000000  // 80 million
	issueTokenGasLimit       = 70000000  // 70 million
	createDepositGasLimit    = 20000000  // 20 million
	generalSCCallGasLimit    = 50000000  // 50 million
	gasLimitPerDataByte      = 1500

	aggregatorContractPath    = "testdata/contracts/kda/multiversx-price-aggregator-sc.wasm"
	wrapperContractPath       = "testdata/contracts/kda/bridged-tokens-wrapper.wasm"
	multiTransferContractPath = "testdata/contracts/kda/multi-transfer-kda.wasm"
	safeContractPath          = "testdata/contracts/kda/kda-safe.wasm"
	multisigContractPath      = "testdata/contracts/kda/multisig.wasm"
	bridgeProxyContractPath   = "testdata/contracts/kda/bridge-proxy.wasm"
	testCallerContractPath    = "testdata/contracts/kda/test-caller.wasm"

	setBridgeProxyContractAddressFunction                = "setBridgeProxyContractAddress"
	setWrappingContractAddressFunction                   = "setWrappingContractAddress"
	changeOwnerAddressFunction                           = "ChangeOwnerAddress"
	setEsdtSafeOnMultiTransferFunction                   = "setEsdtSafeOnMultiTransfer"
	setEsdtSafeOnWrapperFunction                         = "setEsdtSafeContractAddress"
	setEsdtSafeAddressFunction                           = "setEsdtSafeAddress"
	stakeFunction                                        = "stake"
	unpauseFunction                                      = "unpause"
	unpauseEsdtSafeFunction                              = "unpauseEsdtSafe"
	unpauseProxyFunction                                 = "unpauseProxy"
	pauseEsdtSafeFunction                                = "pauseEsdtSafe"
	pauseFunction                                        = "pause"
	issueFunction                                        = "issue"
	setSpecialRoleFunction                               = "setSpecialRole"
	kdaTransferFunction                                  = "KDATransfer"
	setPairDecimalsFunction                              = "setPairDecimals"
	addWrappedTokenFunction                              = "addWrappedToken"
	depositLiquidityFunction                             = "depositLiquidity"
	whitelistTokenFunction                               = "whitelistToken"
	addMappingFunction                                   = "addMapping"
	kdaSafeAddTokenToWhitelistFunction                   = "kdaSafeAddTokenToWhitelist"
	kdaSafeSetMaxBridgedAmountForTokenFunction           = "kdaSafeSetMaxBridgedAmountForToken"
	multiTransferEsdtSetMaxBridgedAmountForTokenFunction = "multiTransferEsdtSetMaxBridgedAmountForToken"
	submitBatchFunction                                  = "submitBatch"
	unwrapTokenCreateTransactionFunction                 = "unwrapTokenCreateTransaction"
	createTransactionFunction                            = "createTransaction"
	setBridgedTokensWrapperAddressFunction               = "setBridgedTokensWrapperAddress"
	setMultiTransferAddressFunction                      = "setMultiTransferAddress"
	withdrawRefundFeesForEthereumFunction                = "withdrawRefundFeesForEthereum"
	getRefundFeesForEthereumFunction                     = "getRefundFeesForEthereum"
	withdrawTransactionFeesFunction                      = "withdrawTransactionFees"
	getTransactionFeesFunction                           = "getTransactionFees"
	initSupplyMintBurnEsdtSafe                           = "initSupplyMintBurnEsdtSafe"
	initSupplyEsdtSafe                                   = "initSupplyEsdtSafe"
)

var (
	feeInt = big.NewInt(50)
)

// KleverchainHandler will handle all the operations on the Kleverchain side
type KleverchainHandler struct {
	testing.TB
	*KeysStore
	Quorum         string
	TokensRegistry TokensRegistry
	ChainSimulator ChainSimulatorWrapper

	AggregatorAddress        *KlvAddress
	WrapperAddress           *KlvAddress
	SafeAddress              *KlvAddress
	MultisigAddress          *KlvAddress
	MultiTransferAddress     *KlvAddress
	ScProxyAddress           *KlvAddress
	TestCallerAddress        *KlvAddress
	KDASystemContractAddress *KlvAddress
}

// NewKleverchainHandler will create the handler that will adapt all test operations on Kleverchain
func NewKleverchainHandler(
	tb testing.TB,
	ctx context.Context,
	keysStore *KeysStore,
	tokensRegistry TokensRegistry,
	chainSimulator ChainSimulatorWrapper,
	quorum string,
) *KleverchainHandler {
	handler := &KleverchainHandler{
		TB:             tb,
		KeysStore:      keysStore,
		TokensRegistry: tokensRegistry,
		ChainSimulator: chainSimulator,
		Quorum:         quorum,
	}

	handler.KDASystemContractAddress = NewKlvAddressFromBech32(handler, kdaSystemSCAddress)

	handler.ChainSimulator.GenerateBlocksUntilEpochReached(ctx, 1)

	handler.ChainSimulator.FundWallets(ctx, handler.WalletsToFundOnKleverchain())
	handler.ChainSimulator.GenerateBlocks(ctx, 1)

	return handler
}

// DeployAndSetContracts will deploy all required contracts on Kleverchain side and do the proper wiring
func (handler *KleverchainHandler) DeployAndSetContracts(ctx context.Context) {
	handler.deployContracts(ctx)

	handler.wireMultiTransfer(ctx)
	handler.wireSCProxy(ctx)
	handler.wireSafe(ctx)

	handler.changeOwners(ctx)
	handler.finishSettings(ctx)
}

func (handler *KleverchainHandler) deployContracts(ctx context.Context) {
	// deploy aggregator
	stakeValue, _ := big.NewInt(0).SetString(minRelayerStake, 10)
	aggregatorDeployParams := []string{
		hex.EncodeToString([]byte("EGLD")),
		hex.EncodeToString(stakeValue.Bytes()),
		"01",
		"02",
		"03",
	}

	for _, oracleKey := range handler.OraclesKeys {
		aggregatorDeployParams = append(aggregatorDeployParams, oracleKey.KlvAddress.Hex())
	}

	hash := ""
	handler.AggregatorAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		aggregatorContractPath,
		handler.OwnerKeys.KlvSk,
		deployGasLimit,
		aggregatorDeployParams,
	)
	require.NotEqual(handler, emptyAddress, handler.AggregatorAddress)
	log.Info("Deploy: aggregator contract", "address", handler.AggregatorAddress, "transaction hash", hash, "num oracles", len(handler.OraclesKeys))

	// deploy wrapper
	handler.WrapperAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		wrapperContractPath,
		handler.OwnerKeys.KlvSk,
		deployGasLimit,
		[]string{},
	)
	require.NotEqual(handler, emptyAddress, handler.WrapperAddress)
	log.Info("Deploy: wrapper contract", "address", handler.WrapperAddress, "transaction hash", hash)

	// deploy multi-transfer
	handler.MultiTransferAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		multiTransferContractPath,
		handler.OwnerKeys.KlvSk,
		deployGasLimit,
		[]string{},
	)
	require.NotEqual(handler, emptyAddress, handler.MultiTransferAddress)
	log.Info("Deploy: multi-transfer contract", "address", handler.MultiTransferAddress, "transaction hash", hash)

	// deploy safe
	handler.SafeAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		safeContractPath,
		handler.OwnerKeys.KlvSk,
		deployGasLimit,
		[]string{
			handler.AggregatorAddress.Hex(),
			handler.MultiTransferAddress.Hex(),
			"01",
		},
	)
	require.NotEqual(handler, emptyAddress, handler.SafeAddress)
	log.Info("Deploy: safe contract", "address", handler.SafeAddress, "transaction hash", hash)

	// deploy bridge proxy
	handler.ScProxyAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		bridgeProxyContractPath,
		handler.OwnerKeys.KlvSk,
		deployGasLimit,
		[]string{
			handler.MultiTransferAddress.Hex(),
		},
	)
	require.NotEqual(handler, emptyAddress, handler.ScProxyAddress)
	log.Info("Deploy: SC proxy contract", "address", handler.ScProxyAddress, "transaction hash", hash)

	// deploy multisig
	minRelayerStakeInt, _ := big.NewInt(0).SetString(minRelayerStake, 10)
	minRelayerStakeHex := hex.EncodeToString(minRelayerStakeInt.Bytes())
	params := []string{
		handler.SafeAddress.Hex(),
		handler.MultiTransferAddress.Hex(),
		handler.ScProxyAddress.Hex(),
		minRelayerStakeHex,
		slashAmount,
		handler.Quorum}
	for _, relayerKeys := range handler.RelayersKeys {
		params = append(params, relayerKeys.KlvAddress.Hex())
	}
	handler.MultisigAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		multisigContractPath,
		handler.OwnerKeys.KlvSk,
		deployGasLimit,
		params,
	)
	require.NotEqual(handler, emptyAddress, handler.MultisigAddress)
	log.Info("Deploy: multisig contract", "address", handler.MultisigAddress, "transaction hash", hash)

	// deploy test-caller
	handler.TestCallerAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		testCallerContractPath,
		handler.OwnerKeys.KlvSk,
		deployGasLimit,
		[]string{},
	)
	require.NotEqual(handler, emptyAddress, handler.TestCallerAddress)
	log.Info("Deploy: test-caller contract", "address", handler.TestCallerAddress, "transaction hash", hash)
}

func (handler *KleverchainHandler) wireMultiTransfer(ctx context.Context) {
	// setBridgeProxyContractAddress
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.MultiTransferAddress,
		zeroStringValue,
		setCallsGasLimit,
		setBridgeProxyContractAddressFunction,
		[]string{
			handler.ScProxyAddress.Hex(),
		},
	)
	log.Info("Set in multi-transfer contract the SC proxy contract", "transaction hash", hash, "status", txResult.Status)

	// setWrappingContractAddress
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.MultiTransferAddress,
		zeroStringValue,
		setCallsGasLimit,
		setWrappingContractAddressFunction,
		[]string{
			handler.WrapperAddress.Hex(),
		},
	)
	log.Info("Set in multi-transfer contract the wrapper contract", "transaction hash", hash, "status", txResult.Status)
}

func (handler *KleverchainHandler) wireSCProxy(ctx context.Context) {
	// setBridgedTokensWrapper in SC bridge proxy
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.ScProxyAddress,
		zeroStringValue,
		setCallsGasLimit,
		setBridgedTokensWrapperAddressFunction,
		[]string{
			handler.WrapperAddress.Hex(),
		},
	)
	log.Info("Set in SC proxy contract the wrapper contract", "transaction hash", hash, "status", txResult.Status)

	// setMultiTransferAddress in SC bridge proxy
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.ScProxyAddress,
		zeroStringValue,
		setCallsGasLimit,
		setMultiTransferAddressFunction,
		[]string{
			handler.MultiTransferAddress.Hex(),
		},
	)
	log.Info("Set in SC proxy contract the multi-transfer contract", "transaction hash", hash, "status", txResult.Status)

	// setEsdtSafeAddress on bridge proxy
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.ScProxyAddress,
		zeroStringValue,
		setCallsGasLimit,
		setEsdtSafeAddressFunction,
		[]string{
			handler.SafeAddress.Hex(),
		},
	)
	log.Info("Set in SC proxy contract the safe contract", "transaction hash", hash, "status", txResult.Status)
}

func (handler *KleverchainHandler) wireSafe(ctx context.Context) {
	// setBridgedTokensWrapperAddress
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.SafeAddress,
		zeroStringValue,
		setCallsGasLimit,
		setBridgedTokensWrapperAddressFunction,
		[]string{
			handler.WrapperAddress.Hex(),
		},
	)
	log.Info("Set in safe contract the wrapper contract", "transaction hash", hash, "status", txResult.Status)

	//setBridgeProxyContractAddress
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.SafeAddress,
		zeroStringValue,
		setCallsGasLimit,
		setBridgeProxyContractAddressFunction,
		[]string{
			handler.ScProxyAddress.Hex(),
		},
	)
	log.Info("Set in safe contract the SC proxy contract", "transaction hash", hash, "status", txResult.Status)
}

func (handler *KleverchainHandler) changeOwners(ctx context.Context) {
	// ChangeOwnerAddress for safe
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.SafeAddress,
		zeroStringValue,
		setCallsGasLimit,
		changeOwnerAddressFunction,
		[]string{
			handler.MultisigAddress.Hex(),
		},
	)
	log.Info("ChangeOwnerAddress for safe contract", "transaction hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for multi-transfer
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.MultiTransferAddress,
		zeroStringValue,
		setCallsGasLimit,
		changeOwnerAddressFunction,
		[]string{
			handler.MultisigAddress.Hex(),
		},
	)
	log.Info("ChangeOwnerAddress for multi-transfer contract", "transaction hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for bridge proxy
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.ScProxyAddress,
		zeroStringValue,
		setCallsGasLimit,
		changeOwnerAddressFunction,
		[]string{
			handler.MultisigAddress.Hex(),
		},
	)
	log.Info("ChangeOwnerAddress for SC proxy contract", "transaction hash", hash, "status", txResult.Status)
}

func (handler *KleverchainHandler) finishSettings(ctx context.Context) {
	// unpause sc proxy
	hash, txResult := handler.callContractNoParams(ctx, handler.MultisigAddress, unpauseProxyFunction)
	log.Info("Un-paused SC proxy contract", "transaction hash", hash, "status", txResult.Status)

	// setEsdtSafeOnMultiTransfer
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		setEsdtSafeOnMultiTransferFunction,
		[]string{},
	)
	log.Info("Set in multisig contract the safe contract (automatically)", "transaction hash", hash, "status", txResult.Status)

	// stake relayers on multisig
	handler.stakeAddressesOnContract(ctx, handler.MultisigAddress, handler.RelayersKeys)

	// stake relayers on price aggregator
	handler.stakeAddressesOnContract(ctx, handler.AggregatorAddress, handler.OraclesKeys)

	// unpause multisig
	hash, txResult = handler.callContractNoParams(ctx, handler.MultisigAddress, unpauseFunction)
	log.Info("Un-paused multisig contract", "transaction hash", hash, "status", txResult.Status)

	handler.UnPauseContractsAfterTokenChanges(ctx)
}

// CheckForZeroBalanceOnReceivers will check that the balances for all provided tokens are 0 for the test address and the test SC call address
func (handler *KleverchainHandler) CheckForZeroBalanceOnReceivers(ctx context.Context, tokens ...TestTokenParams) {
	for _, params := range tokens {
		handler.CheckForZeroBalanceOnReceiversForToken(ctx, params)
	}
}

// CheckForZeroBalanceOnReceiversForToken will check that the balance for the test address and the test SC call address is 0
func (handler *KleverchainHandler) CheckForZeroBalanceOnReceiversForToken(ctx context.Context, token TestTokenParams) {
	balance := handler.GetKDAUniversalTokenBalance(ctx, handler.TestKeys.KlvAddress, token.AbstractTokenIdentifier)
	require.Equal(handler, big.NewInt(0).String(), balance.String())

	balance = handler.GetKDAUniversalTokenBalance(ctx, handler.TestCallerAddress, token.AbstractTokenIdentifier)
	require.Equal(handler, big.NewInt(0).String(), balance.String())
}

// GetKDAUniversalTokenBalance will return the universal KDA token's balance
func (handler *KleverchainHandler) GetKDAUniversalTokenBalance(
	ctx context.Context,
	address *KlvAddress,
	abstractTokenIdentifier string,
) *big.Int {
	token := handler.TokensRegistry.GetTokenData(abstractTokenIdentifier)
	require.NotNil(handler, token)

	balanceString := handler.ChainSimulator.GetKDABalance(ctx, address, token.KlvUniversalToken)

	balance, ok := big.NewInt(0).SetString(balanceString, 10)
	require.True(handler, ok)

	return balance
}

// GetKDAChainSpecificTokenBalance will return the chain specific KDA token's balance
func (handler *KleverchainHandler) GetKDAChainSpecificTokenBalance(
	ctx context.Context,
	address *KlvAddress,
	abstractTokenIdentifier string,
) *big.Int {
	token := handler.TokensRegistry.GetTokenData(abstractTokenIdentifier)
	require.NotNil(handler, token)

	balanceString := handler.ChainSimulator.GetKDABalance(ctx, address, token.KlvChainSpecificToken)

	balance, ok := big.NewInt(0).SetString(balanceString, 10)
	require.True(handler, ok)

	return balance
}

func (handler *KleverchainHandler) callContractNoParams(ctx context.Context, contract *KlvAddress, endpoint string) (string, *data.TransactionOnNetwork) {
	return handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		contract,
		zeroStringValue,
		setCallsGasLimit,
		endpoint,
		[]string{},
	)
}

// UnPauseContractsAfterTokenChanges can unpause contracts after token changes
func (handler *KleverchainHandler) UnPauseContractsAfterTokenChanges(ctx context.Context) {
	// unpause safe
	hash, txResult := handler.callContractNoParams(ctx, handler.MultisigAddress, unpauseEsdtSafeFunction)
	log.Info("unpaused safe executed", "hash", hash, "status", txResult.Status)

	// unpause wrapper
	hash, txResult = handler.callContractNoParams(ctx, handler.WrapperAddress, unpauseFunction)
	log.Info("unpaused wrapper executed", "hash", hash, "status", txResult.Status)

	// unpause aggregator
	hash, txResult = handler.callContractNoParams(ctx, handler.AggregatorAddress, unpauseFunction)
	log.Info("unpaused aggregator executed", "hash", hash, "status", txResult.Status)
}

// PauseContractsForTokenChanges can pause contracts for token changes
func (handler *KleverchainHandler) PauseContractsForTokenChanges(ctx context.Context) {
	// pause safe
	hash, txResult := handler.callContractNoParams(ctx, handler.MultisigAddress, pauseEsdtSafeFunction)
	log.Info("paused safe executed", "hash", hash, "status", txResult.Status)

	// pause aggregator
	hash, txResult = handler.callContractNoParams(ctx, handler.AggregatorAddress, pauseFunction)
	log.Info("paused aggregator executed", "hash", hash, "status", txResult.Status)

	// pause wrapper
	hash, txResult = handler.callContractNoParams(ctx, handler.WrapperAddress, pauseFunction)
	log.Info("paused wrapper executed", "hash", hash, "status", txResult.Status)
}

func (handler *KleverchainHandler) stakeAddressesOnContract(ctx context.Context, contract *KlvAddress, allKeys []KeysHolder) {
	for _, keys := range allKeys {
		hash, txResult := handler.ChainSimulator.SendTx(
			ctx,
			keys.KlvSk,
			contract,
			minRelayerStake,
			setCallsGasLimit,
			[]byte(stakeFunction),
		)
		log.Info(fmt.Sprintf("Address %s staked on contract %s with transaction hash %s, status %s", keys.KlvAddress, contract, hash, txResult.Status))
	}
}

// IssueAndWhitelistToken will issue and whitelist the token on Kleverchain
func (handler *KleverchainHandler) IssueAndWhitelistToken(ctx context.Context, params IssueTokenParams) {
	if params.HasChainSpecificToken {
		handler.issueAndWhitelistTokensWithChainSpecific(ctx, params)
	} else {
		handler.issueAndWhitelistTokens(ctx, params)
	}
}

func (handler *KleverchainHandler) issueAndWhitelistTokensWithChainSpecific(ctx context.Context, params IssueTokenParams) {
	handler.issueUniversalToken(ctx, params)
	handler.issueChainSpecificToken(ctx, params)
	handler.setLocalRolesForUniversalTokenOnWrapper(ctx, params)
	handler.transferChainSpecificTokenToSCs(ctx, params)
	handler.addUniversalTokenToWrapper(ctx, params)
	handler.whitelistTokenOnWrapper(ctx, params)
	handler.setRolesForSpecificTokenOnSafe(ctx, params)
	handler.addMappingInMultisig(ctx, params)
	handler.whitelistTokenOnMultisig(ctx, params)
	handler.setInitialSupply(ctx, params)
	handler.setPairDecimalsOnAggregator(ctx, params)
	handler.setMaxBridgeAmountOnSafe(ctx, params)
	handler.setMaxBridgeAmountOnMultitransfer(ctx, params)
}

func (handler *KleverchainHandler) issueAndWhitelistTokens(ctx context.Context, params IssueTokenParams) {
	handler.issueUniversalToken(ctx, params)

	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	handler.TokensRegistry.RegisterChainSpecificToken(params.AbstractTokenIdentifier, tkData.KlvUniversalToken)

	handler.setRolesForSpecificTokenOnSafe(ctx, params)
	handler.addMappingInMultisig(ctx, params)
	handler.whitelistTokenOnMultisig(ctx, params)
	handler.setInitialSupply(ctx, params)
	handler.setPairDecimalsOnAggregator(ctx, params)
	handler.setMaxBridgeAmountOnSafe(ctx, params)
	handler.setMaxBridgeAmountOnMultitransfer(ctx, params)
}

func (handler *KleverchainHandler) issueUniversalToken(ctx context.Context, params IssueTokenParams) {
	token := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(handler, token)

	valueToMintInt, ok := big.NewInt(0).SetString(params.ValueToMintOnKlv, 10)
	require.True(handler, ok)

	// issue universal token
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.KDASystemContractAddress,
		kdaIssueCost,
		issueTokenGasLimit,
		issueFunction,
		[]string{
			hex.EncodeToString([]byte(params.KlvUniversalTokenDisplayName)),
			hex.EncodeToString([]byte(params.KlvUniversalTokenTicker)),
			hex.EncodeToString(valueToMintInt.Bytes()),
			fmt.Sprintf("%02x", params.NumOfDecimalsUniversal),
			hex.EncodeToString([]byte(canAddSpecialRoles)),
			hex.EncodeToString([]byte(trueStr))})
	kdaUniversalToken := handler.getTokenNameFromResult(*txResult)
	require.Greater(handler, len(kdaUniversalToken), 0)
	handler.TokensRegistry.RegisterUniversalToken(params.AbstractTokenIdentifier, kdaUniversalToken)
	log.Info("issue universal token tx executed", "hash", hash, "status", txResult.Status, "token", kdaUniversalToken, "owner", handler.OwnerKeys.KlvAddress)
}

func (handler *KleverchainHandler) issueChainSpecificToken(ctx context.Context, params IssueTokenParams) {
	valueToMintInt, ok := big.NewInt(0).SetString(params.ValueToMintOnKlv, 10)
	require.True(handler, ok)

	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.KDASystemContractAddress,
		kdaIssueCost,
		issueTokenGasLimit,
		issueFunction,
		[]string{
			hex.EncodeToString([]byte(params.KlvChainSpecificTokenDisplayName)),
			hex.EncodeToString([]byte(params.KlvChainSpecificTokenTicker)),
			hex.EncodeToString(valueToMintInt.Bytes()),
			fmt.Sprintf("%02x", params.NumOfDecimalsChainSpecific),
			hex.EncodeToString([]byte(canAddSpecialRoles)),
			hex.EncodeToString([]byte(trueStr))})
	kdaChainSpecificToken := handler.getTokenNameFromResult(*txResult)
	require.Greater(handler, len(kdaChainSpecificToken), 0)
	handler.TokensRegistry.RegisterChainSpecificToken(params.AbstractTokenIdentifier, kdaChainSpecificToken)
	log.Info("issue chain specific token tx executed", "hash", hash, "status", txResult.Status, "token", kdaChainSpecificToken, "owner", handler.OwnerKeys.KlvAddress)
}

func (handler *KleverchainHandler) setLocalRolesForUniversalTokenOnWrapper(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// set local roles bridged tokens wrapper
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.KDASystemContractAddress,
		zeroStringValue,
		setCallsGasLimit,
		setSpecialRoleFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.KlvUniversalToken)),
			handler.WrapperAddress.Hex(),
			hex.EncodeToString([]byte(kdaRoleLocalMint)),
			hex.EncodeToString([]byte(kdaRoleLocalBurn))})
	log.Info("set local roles bridged tokens wrapper tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *KleverchainHandler) transferChainSpecificTokenToSCs(ctx context.Context, params IssueTokenParams) {
	valueToMintInt, ok := big.NewInt(0).SetString(params.ValueToMintOnKlv, 10)
	require.True(handler, ok)

	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// transfer to wrapper sc
	initialMintValue := valueToMintInt.Div(valueToMintInt, big.NewInt(3))
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.WrapperAddress,
		zeroStringValue,
		setCallsGasLimit,
		kdaTransferFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.KlvChainSpecificToken)),
			hex.EncodeToString(initialMintValue.Bytes()),
			hex.EncodeToString([]byte(depositLiquidityFunction))})
	log.Info("transfer to wrapper sc tx executed", "hash", hash, "status", txResult.Status)

	// transfer to safe sc
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.SafeAddress,
		zeroStringValue,
		setCallsGasLimit,
		kdaTransferFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.KlvChainSpecificToken)),
			hex.EncodeToString(initialMintValue.Bytes())})
	log.Info("transfer to safe sc tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *KleverchainHandler) addUniversalTokenToWrapper(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// add wrapped token
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.WrapperAddress,
		zeroStringValue,
		setCallsGasLimit,
		addWrappedTokenFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.KlvUniversalToken)),
			fmt.Sprintf("%02x", params.NumOfDecimalsUniversal),
		})
	log.Info("add wrapped token tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *KleverchainHandler) whitelistTokenOnWrapper(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// wrapper whitelist token
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.WrapperAddress,
		zeroStringValue,
		setCallsGasLimit,
		whitelistTokenFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.KlvChainSpecificToken)),
			fmt.Sprintf("%02x", params.NumOfDecimalsChainSpecific),
			hex.EncodeToString([]byte(tkData.KlvUniversalToken))})
	log.Info("wrapper whitelist token tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *KleverchainHandler) setRolesForSpecificTokenOnSafe(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// set local roles kda safe
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.KDASystemContractAddress,
		zeroStringValue,
		setCallsGasLimit,
		setSpecialRoleFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.KlvChainSpecificToken)),
			handler.SafeAddress.Hex(),
			hex.EncodeToString([]byte(kdaRoleLocalMint)),
			hex.EncodeToString([]byte(kdaRoleLocalBurn))})
	log.Info("set local roles kda safe tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *KleverchainHandler) addMappingInMultisig(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// add mapping
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		addMappingFunction,
		[]string{
			hex.EncodeToString(tkData.EthErc20Address.Bytes()),
			hex.EncodeToString([]byte(tkData.KlvChainSpecificToken))})
	log.Info("add mapping tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *KleverchainHandler) whitelistTokenOnMultisig(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// whitelist token
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		kdaSafeAddTokenToWhitelistFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.KlvChainSpecificToken)),
			hex.EncodeToString([]byte(params.KlvChainSpecificTokenTicker)),
			getHexBool(params.IsMintBurnOnKlv),
			getHexBool(params.IsNativeOnKlv),
			hex.EncodeToString(zeroValueBigInt.Bytes()), // total_balance
			hex.EncodeToString(zeroValueBigInt.Bytes()), // mint_balance
			hex.EncodeToString(zeroValueBigInt.Bytes()), // burn_balance
		})
	log.Info("whitelist token tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *KleverchainHandler) setInitialSupply(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// set initial supply
	if len(params.InitialSupplyValue) > 0 {
		initialSupply, okConvert := big.NewInt(0).SetString(params.InitialSupplyValue, 10)
		require.True(handler, okConvert)

		if params.IsMintBurnOnKlv {
			hash, txResult := handler.ChainSimulator.ScCall(
				ctx,
				handler.OwnerKeys.KlvSk,
				handler.MultisigAddress,
				zeroStringValue,
				setCallsGasLimit,
				initSupplyMintBurnEsdtSafe,
				[]string{
					hex.EncodeToString([]byte(tkData.KlvChainSpecificToken)),
					hex.EncodeToString(initialSupply.Bytes()),
					hex.EncodeToString([]byte{0}),
				},
			)
			log.Info("initial supply tx executed", "hash", hash, "status", txResult.Status,
				"initial mint", params.InitialSupplyValue, "initial burned", "0")
		} else {
			hash, txResult := handler.ChainSimulator.ScCall(
				ctx,
				handler.OwnerKeys.KlvSk,
				handler.MultisigAddress,
				zeroStringValue,
				setCallsGasLimit,
				kdaTransferFunction,
				[]string{
					hex.EncodeToString([]byte(tkData.KlvChainSpecificToken)),
					hex.EncodeToString(initialSupply.Bytes()),
					hex.EncodeToString([]byte(initSupplyEsdtSafe)),
					hex.EncodeToString([]byte(tkData.KlvChainSpecificToken)),
					hex.EncodeToString(initialSupply.Bytes()),
				})

			log.Info("initial supply tx executed", "hash", hash, "status", txResult.Status,
				"initial value", params.InitialSupplyValue)
		}
	}
}

func (handler *KleverchainHandler) setPairDecimalsOnAggregator(ctx context.Context, params IssueTokenParams) {
	// setPairDecimals on aggregator
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.AggregatorAddress,
		zeroStringValue,
		setCallsGasLimit,
		setPairDecimalsFunction,
		[]string{
			hex.EncodeToString([]byte(gwei)),
			hex.EncodeToString([]byte(params.KlvChainSpecificTokenTicker)),
			fmt.Sprintf("%02x", params.NumOfDecimalsChainSpecific)})
	log.Info("setPairDecimals tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *KleverchainHandler) setMaxBridgeAmountOnSafe(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// safe set max bridge amount for token
	maxBridgedAmountForTokenInt, _ := big.NewInt(0).SetString(maxBridgedAmountForToken, 10)
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		kdaSafeSetMaxBridgedAmountForTokenFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.KlvChainSpecificToken)),
			hex.EncodeToString(maxBridgedAmountForTokenInt.Bytes())})
	log.Info("safe set max bridge amount for token tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *KleverchainHandler) setMaxBridgeAmountOnMultitransfer(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// multi-transfer set max bridge amount for token
	maxBridgedAmountForTokenInt, _ := big.NewInt(0).SetString(maxBridgedAmountForToken, 10)
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		multiTransferEsdtSetMaxBridgedAmountForTokenFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.KlvChainSpecificToken)),
			hex.EncodeToString(maxBridgedAmountForTokenInt.Bytes())})
	log.Info("multi-transfer set max bridge amount for token tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *KleverchainHandler) getTokenNameFromResult(txResult data.TransactionOnNetwork) string {
	for _, event := range txResult.Logs.Events {
		if event.Identifier == issueFunction {
			require.Greater(handler, len(event.Topics), 1)

			return string(event.Topics[0])
		}
	}

	require.Fail(handler, "did not find the event with the issue identifier")
	return ""
}

// SubmitAggregatorBatch will submit the aggregator batch
func (handler *KleverchainHandler) SubmitAggregatorBatch(ctx context.Context, params IssueTokenParams) {
	txHashes := make([]string, 0, len(handler.OraclesKeys))
	for _, key := range handler.OraclesKeys {
		hash := handler.submitAggregatorBatchForKey(ctx, key, params)
		txHashes = append(txHashes, hash)
	}

	for _, hash := range txHashes {
		txResult := handler.ChainSimulator.GetTransactionResult(ctx, hash)
		log.Info("submit aggregator batch tx", "hash", hash, "status", txResult.Status)
	}
}

func (handler *KleverchainHandler) submitAggregatorBatchForKey(ctx context.Context, key KeysHolder, params IssueTokenParams) string {
	timestamp := handler.ChainSimulator.GetBlockchainTimeStamp(ctx)
	require.Greater(handler, timestamp, uint64(0), "something went wrong and the chain simulator returned 0 for the current timestamp")

	timestampAsBigInt := big.NewInt(0).SetUint64(timestamp)

	hash := handler.ChainSimulator.ScCallWithoutGenerateBlocks(
		ctx,
		key.KlvSk,
		handler.AggregatorAddress,
		zeroStringValue,
		setCallsGasLimit,
		submitBatchFunction,
		[]string{
			hex.EncodeToString([]byte(gwei)),
			hex.EncodeToString([]byte(params.KlvChainSpecificTokenTicker)),
			hex.EncodeToString(timestampAsBigInt.Bytes()),
			hex.EncodeToString(feeInt.Bytes()),
			fmt.Sprintf("%02x", params.NumOfDecimalsChainSpecific)})

	log.Info("submit aggregator batch tx sent", "transaction hash", hash, "submitter", key.KlvAddress.Bech32())

	return hash
}

// SendDepositTransactionFromKleverchain will send the deposit transaction from Kleverchain
func (handler *KleverchainHandler) SendDepositTransactionFromKleverchain(ctx context.Context, token *TokenData, params TestTokenParams, value *big.Int) {
	if params.HasChainSpecificToken {
		handler.unwrapCreateTransaction(ctx, token, value)
		return
	}

	handler.createTransactionWithoutUnwrap(ctx, token, value)
}

func (handler *KleverchainHandler) createTransactionWithoutUnwrap(ctx context.Context, token *TokenData, value *big.Int) {
	// create transaction params
	params := []string{
		hex.EncodeToString([]byte(token.KlvUniversalToken)),
		hex.EncodeToString(value.Bytes()),
		hex.EncodeToString([]byte(createTransactionFunction)),
		hex.EncodeToString(handler.TestKeys.EthAddress.Bytes()),
	}
	dataField := strings.Join(params, "@")

	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.TestKeys.KlvSk,
		handler.SafeAddress,
		zeroStringValue,
		createDepositGasLimit+gasLimitPerDataByte*uint64(len(dataField)),
		kdaTransferFunction,
		params,
	)
	log.Info("Kleverchain->Ethereum createTransaction sent", "hash", hash, "token", token.KlvUniversalToken, "status", txResult.Status)
}

func (handler *KleverchainHandler) unwrapCreateTransaction(ctx context.Context, token *TokenData, value *big.Int) {
	// create transaction params
	params := []string{
		hex.EncodeToString([]byte(token.KlvUniversalToken)),
		hex.EncodeToString(value.Bytes()),
		hex.EncodeToString([]byte(unwrapTokenCreateTransactionFunction)),
		hex.EncodeToString([]byte(token.KlvChainSpecificToken)),
		hex.EncodeToString(handler.SafeAddress.Bytes()),
		hex.EncodeToString(handler.TestKeys.EthAddress.Bytes()),
	}
	dataField := strings.Join(params, "@")

	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.TestKeys.KlvSk,
		handler.WrapperAddress,
		zeroStringValue,
		createDepositGasLimit+gasLimitPerDataByte*uint64(len(dataField)),
		kdaTransferFunction,
		params,
	)
	log.Info("Kleverchain->Ethereum unwrapCreateTransaction sent", "hash", hash, "token", token.KlvUniversalToken, "status", txResult.Status)
}

// TestWithdrawFees will try to withdraw the fees for the provided token from the safe contract to the owner
func (handler *KleverchainHandler) TestWithdrawFees(
	ctx context.Context,
	token string,
	expectedDeltaForRefund *big.Int,
	expectedDeltaForAccumulated *big.Int,
) {
	handler.withdrawFees(ctx, token, expectedDeltaForRefund, getRefundFeesForEthereumFunction, withdrawRefundFeesForEthereumFunction)
	handler.withdrawFees(ctx, token, expectedDeltaForAccumulated, getTransactionFeesFunction, withdrawTransactionFeesFunction)
}

func (handler *KleverchainHandler) withdrawFees(ctx context.Context,
	token string,
	expectedDelta *big.Int,
	getFunction string,
	withdrawFunction string,
) {
	queryParams := []string{
		hex.EncodeToString([]byte(token)),
	}
	responseData := handler.ChainSimulator.ExecuteVMQuery(ctx, handler.SafeAddress, getFunction, queryParams)
	value := big.NewInt(0).SetBytes(responseData[0])
	require.Equal(handler, expectedDelta.String(), value.String())
	if expectedDelta.Cmp(zeroValueBigInt) == 0 {
		return
	}

	handler.ChainSimulator.GenerateBlocks(ctx, 5) // ensure block finality
	initialBalanceStr := handler.ChainSimulator.GetKDABalance(ctx, handler.OwnerKeys.KlvAddress, token)
	initialBalance, ok := big.NewInt(0).SetString(initialBalanceStr, 10)
	require.True(handler, ok)

	handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.KlvSk,
		handler.MultisigAddress,
		zeroStringValue,
		generalSCCallGasLimit,
		withdrawFunction,
		[]string{
			hex.EncodeToString([]byte(token)),
		},
	)

	handler.ChainSimulator.GenerateBlocks(ctx, 5) // ensure block finality
	finalBalanceStr := handler.ChainSimulator.GetKDABalance(ctx, handler.OwnerKeys.KlvAddress, token)
	finalBalance, ok := big.NewInt(0).SetString(finalBalanceStr, 10)
	require.True(handler, ok)

	require.Equal(handler, expectedDelta, finalBalance.Sub(finalBalance, initialBalance),
		fmt.Sprintf("mismatch on balance check after the call to %s: initial balance: %s, final balance %s, expected delta: %s",
			withdrawFunction, initialBalanceStr, finalBalanceStr, expectedDelta.String()))
}

// TransferToken is able to create an KDA transfer
func (handler *KleverchainHandler) TransferToken(ctx context.Context, source KeysHolder, receiver KeysHolder, amount *big.Int, params TestTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// transfer to the test key, so it will have funds to carry on with the deposits
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		source.KlvSk,
		receiver.KlvAddress,
		zeroStringValue,
		createDepositGasLimit,
		kdaTransferFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.KlvUniversalToken)),
			hex.EncodeToString(amount.Bytes())})

	log.Info("transfer to tx executed",
		"source address", source.KlvAddress.Bech32(),
		"receiver", receiver.KlvAddress.Bech32(),
		"token", tkData.KlvUniversalToken,
		"amount", amount.String(),
		"hash", hash, "status", txResult.Status)
}

func getHexBool(input bool) string {
	if input {
		return hexTrue
	}

	return hexFalse
}
