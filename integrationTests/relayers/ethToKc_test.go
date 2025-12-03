//go:build !slow

package relayers

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/klever-io/klv-bridge-eth-go/clients/ethereum/contract"
	"github.com/klever-io/klv-bridge-eth-go/config"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/factory"
	"github.com/klever-io/klv-bridge-eth-go/integrationTests"
	"github.com/klever-io/klv-bridge-eth-go/integrationTests/mock"
	"github.com/klever-io/klv-bridge-eth-go/status"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const noGasStationURL = ""

type argsForSCCallsTest struct {
	providedScCallData []byte
	expectedScCallData []byte
}

func TestRelayersShouldExecuteTransfersFromEthToKC(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("simple tokens transfers", func(t *testing.T) {
		testRelayersShouldExecuteTransfersFromEthToKC(t, false)
	})
	t.Run("native tokens transfers", func(t *testing.T) {
		testRelayersShouldExecuteTransfersFromEthToKC(t, true)
	})
}

func testRelayersShouldExecuteTransfersFromEthToKC(t *testing.T, withNativeTokens bool) {
	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()
	token1Erc20 := testsCommon.CreateRandomEthereumAddress()
	ticker1 := "tck-000001"

	token2Erc20 := testsCommon.CreateRandomEthereumAddress()
	ticker2 := "tck-000002"

	value1 := big.NewInt(111111111)
	destination1 := testsCommon.CreateRandomKCAddress()
	depositor1 := testsCommon.CreateRandomEthereumAddress()

	value2 := big.NewInt(222222222)
	destination2 := testsCommon.CreateRandomKCAddress()
	depositor2 := testsCommon.CreateRandomEthereumAddress()

	tokens := []common.Address{token1Erc20, token2Erc20}
	availableBalances := []*big.Int{value1, value2}

	erc20ContractsHolder := createMockErc20ContractsHolder(tokens, safeContractEthAddress, availableBalances)

	batchNonceOnEthereum := uint64(345)
	txNonceOnEthereum := uint64(772634)
	batch := contract.Batch{
		Nonce:                  big.NewInt(int64(batchNonceOnEthereum) + 1),
		BlockNumber:            0,
		LastUpdatedBlockNumber: 0,
		DepositsCount:          2,
	}

	numRelayers := 3
	ethereumChainMock := mock.NewEthereumChainMock()
	ethereumChainMock.AddBatch(batch)
	ethereumChainMock.AddDepositToBatch(batchNonceOnEthereum+1, contract.Deposit{
		Nonce:        big.NewInt(int64(txNonceOnEthereum) + 1),
		TokenAddress: token1Erc20,
		Amount:       value1,
		Depositor:    depositor1,
		Recipient:    destination1.AddressSlice(),
		Status:       0,
	})
	ethereumChainMock.AddDepositToBatch(batchNonceOnEthereum+1, contract.Deposit{
		Nonce:        big.NewInt(int64(txNonceOnEthereum) + 2),
		TokenAddress: token2Erc20,
		Amount:       value2,
		Depositor:    depositor2,
		Recipient:    destination2.AddressSlice(),
		Status:       0,
	})
	ethereumChainMock.AddBatch(batch)
	ethereumChainMock.SetQuorum(numRelayers)
	ethereumChainMock.SetFinalNonce(batchNonceOnEthereum + 1)

	kcChainMock := mock.NewKleverBlockchainMock()

	if !withNativeTokens {
		ethereumChainMock.UpdateNativeTokens(token1Erc20, true)
		ethereumChainMock.UpdateMintBurnTokens(token1Erc20, false)
		ethereumChainMock.UpdateTotalBalances(token1Erc20, value1)

		ethereumChainMock.UpdateNativeTokens(token2Erc20, true)
		ethereumChainMock.UpdateMintBurnTokens(token2Erc20, false)
		ethereumChainMock.UpdateTotalBalances(token2Erc20, value2)

		kcChainMock.AddTokensPair(token1Erc20, ticker1, withNativeTokens, true, zero, zero, zero)
		kcChainMock.AddTokensPair(token2Erc20, ticker2, withNativeTokens, true, zero, zero, zero)
	} else {
		ethereumChainMock.UpdateNativeTokens(token1Erc20, false)
		ethereumChainMock.UpdateMintBurnTokens(token1Erc20, true)
		ethereumChainMock.UpdateBurnBalances(token1Erc20, value1)
		ethereumChainMock.UpdateMintBalances(token1Erc20, value1)

		ethereumChainMock.UpdateNativeTokens(token2Erc20, false)
		ethereumChainMock.UpdateMintBurnTokens(token2Erc20, true)
		ethereumChainMock.UpdateBurnBalances(token2Erc20, value2)
		ethereumChainMock.UpdateMintBalances(token2Erc20, value2)

		kcChainMock.AddTokensPair(token1Erc20, ticker1, withNativeTokens, true, zero, zero, value1)
		kcChainMock.AddTokensPair(token2Erc20, ticker2, withNativeTokens, true, zero, zero, value2)
	}

	kcChainMock.SetLastExecutedEthBatchID(batchNonceOnEthereum)
	kcChainMock.SetLastExecutedEthTxId(txNonceOnEthereum)
	kcChainMock.GetStatusesAfterExecutionHandler = func() []byte {
		return []byte{bridgeCore.Executed, bridgeCore.Rejected}
	}
	kcChainMock.SetQuorum(numRelayers)

	relayers := make([]bridgeComponents, 0, numRelayers)
	defer closeRelayers(relayers)

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()
	kcChainMock.ProcessFinishedHandler = func() {
		log.Info("kcChainMock.ProcessFinishedHandler called")
		asyncCancelCall(cancel, time.Second*5)
	}

	for i := 0; i < numRelayers; i++ {
		argsBridgeComponents := createMockBridgeComponentsArgs(i, messengers[i], kcChainMock, ethereumChainMock)
		argsBridgeComponents.Configs.GeneralConfig.Eth.SafeContractAddress = safeContractEthAddress.Hex()
		argsBridgeComponents.Erc20ContractsHolder = erc20ContractsHolder
		relayer, err := factory.NewEthKleverBridgeComponents(argsBridgeComponents)
		require.Nil(t, err)

		kcChainMock.AddRelayer(relayer.KleverRelayerAddress())
		ethereumChainMock.AddRelayer(relayer.EthereumRelayerAddress())

		go func() {
			err = relayer.Start()
			integrationTests.Log.LogIfError(err)
			require.Nil(t, err)
		}()

		relayers = append(relayers, relayer)
	}

	<-ctx.Done()
	time.Sleep(time.Second * 5)

	assert.NotNil(t, kcChainMock.PerformedActionID())
	transfer := kcChainMock.ProposedTransfer()
	require.NotNil(t, transfer)
	require.Equal(t, 2, len(transfer.Transfers))
	assert.Equal(t, batchNonceOnEthereum+1, transfer.BatchId.Uint64())

	assert.Equal(t, destination1.Bytes(), transfer.Transfers[0].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker1)), transfer.Transfers[0].Token)
	assert.Equal(t, value1, transfer.Transfers[0].Amount)
	assert.Equal(t, depositor1, common.BytesToAddress(transfer.Transfers[0].From))
	assert.Equal(t, txNonceOnEthereum+1, transfer.Transfers[0].Nonce.Uint64())
	assert.Equal(t, []byte{bridgeCore.MissingDataProtocolMarker}, transfer.Transfers[0].Data)

	assert.Equal(t, destination2.Bytes(), transfer.Transfers[1].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker2)), transfer.Transfers[1].Token)
	assert.Equal(t, value2, transfer.Transfers[1].Amount)
	assert.Equal(t, depositor2, common.BytesToAddress(transfer.Transfers[1].From))
	assert.Equal(t, txNonceOnEthereum+2, transfer.Transfers[1].Nonce.Uint64())
	assert.Equal(t, []byte{bridgeCore.MissingDataProtocolMarker}, transfer.Transfers[1].Data)
}

func TestRelayersShouldExecuteTransferFromEthToKCHavingTxsWithSCcalls(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("correct SC call", func(t *testing.T) {
		testArgs := argsForSCCallsTest{
			providedScCallData: bridge.EthCallDataMock,
			expectedScCallData: bridge.CallDataMock,
		}

		testRelayersShouldExecuteTransferFromEthToKCHavingTxsWithSCcalls(t, testArgs)
	})
	t.Run("no SC call", func(t *testing.T) {
		testArgs := argsForSCCallsTest{
			providedScCallData: []byte{bridgeCore.MissingDataProtocolMarker},
			expectedScCallData: []byte{bridgeCore.MissingDataProtocolMarker},
		}

		testRelayersShouldExecuteTransferFromEthToKCHavingTxsWithSCcalls(t, testArgs)
	})
}

func testRelayersShouldExecuteTransferFromEthToKCHavingTxsWithSCcalls(t *testing.T, args argsForSCCallsTest) {
	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()

	token1Erc20 := testsCommon.CreateRandomEthereumAddress()
	ticker1 := "tck-000001"

	token2Erc20 := testsCommon.CreateRandomEthereumAddress()
	ticker2 := "tck-000002"

	token3Erc20 := testsCommon.CreateRandomEthereumAddress()
	ticker3 := "tck-000003"

	value1 := big.NewInt(111111111)
	destination1 := testsCommon.CreateRandomKCAddress()
	depositor1 := testsCommon.CreateRandomEthereumAddress()

	value2 := big.NewInt(222222222)
	destination2 := testsCommon.CreateRandomKCAddress()
	depositor2 := testsCommon.CreateRandomEthereumAddress()

	depositor3 := testsCommon.CreateRandomEthereumAddress()

	value3 := big.NewInt(333333333)
	destination3Sc := testsCommon.CreateRandomKCSCAddress()

	tokens := []common.Address{token1Erc20, token2Erc20, token3Erc20}
	availableBalances := []*big.Int{value1, value2, value3}

	erc20ContractsHolder := createMockErc20ContractsHolder(tokens, safeContractEthAddress, availableBalances)

	batchNonceOnEthereum := uint64(345)
	txNonceOnEthereum := uint64(772634)
	batch := contract.Batch{
		Nonce:                  big.NewInt(int64(batchNonceOnEthereum) + 1),
		BlockNumber:            0,
		LastUpdatedBlockNumber: 0,
		DepositsCount:          3,
	}

	numRelayers := 3
	ethereumChainMock := mock.NewEthereumChainMock()
	ethereumChainMock.AddBatch(batch)
	ethereumChainMock.AddDepositToBatch(batchNonceOnEthereum+1, contract.Deposit{
		Nonce:        big.NewInt(int64(txNonceOnEthereum) + 1),
		TokenAddress: token1Erc20,
		Amount:       value1,
		Depositor:    depositor1,
		Recipient:    destination1.AddressSlice(),
		Status:       0,
	})
	ethereumChainMock.AddDepositToBatch(batchNonceOnEthereum+1, contract.Deposit{
		Nonce:        big.NewInt(int64(txNonceOnEthereum) + 2),
		TokenAddress: token2Erc20,
		Amount:       value2,
		Depositor:    depositor2,
		Recipient:    destination2.AddressSlice(),
		Status:       0,
	})
	ethereumChainMock.AddDepositToBatch(batchNonceOnEthereum+1, contract.Deposit{
		Nonce:        big.NewInt(int64(txNonceOnEthereum) + 3),
		TokenAddress: token3Erc20,
		Amount:       value3,
		Depositor:    depositor3,
		Recipient:    destination3Sc.AddressSlice(),
		Status:       0,
	})
	ethereumChainMock.AddBatch(batch)
	ethereumChainMock.SetQuorum(numRelayers)
	ethereumChainMock.SetFinalNonce(batchNonceOnEthereum + 1)

	ethereumChainMock.UpdateNativeTokens(token1Erc20, true)
	ethereumChainMock.UpdateMintBurnTokens(token1Erc20, false)
	ethereumChainMock.UpdateTotalBalances(token1Erc20, value1)

	ethereumChainMock.UpdateNativeTokens(token2Erc20, true)
	ethereumChainMock.UpdateMintBurnTokens(token2Erc20, false)
	ethereumChainMock.UpdateTotalBalances(token2Erc20, value2)

	ethereumChainMock.UpdateNativeTokens(token3Erc20, true)
	ethereumChainMock.UpdateMintBurnTokens(token3Erc20, false)
	ethereumChainMock.UpdateTotalBalances(token3Erc20, value3)

	ethereumChainMock.FilterLogsCalled = func(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
		expectedBatchNonceHash := []common.Hash{
			common.BytesToHash(big.NewInt(int64(batchNonceOnEthereum + 1)).Bytes()),
		}
		require.Equal(t, 2, len(q.Topics))
		assert.Equal(t, expectedBatchNonceHash, q.Topics[1])

		scExecAbi, err := contract.ERC20SafeMetaData.GetAbi()
		require.Nil(t, err)

		eventInputs := scExecAbi.Events["ERC20SCDeposit"].Inputs.NonIndexed()
		packedArgs, err := eventInputs.Pack(big.NewInt(0).SetUint64(txNonceOnEthereum+3), args.providedScCallData)
		require.Nil(t, err)

		scLog := types.Log{
			Data: packedArgs,
		}

		return []types.Log{scLog}, nil
	}

	kcChainMock := mock.NewKleverBlockchainMock()
	kcChainMock.AddTokensPair(token1Erc20, ticker1, false, true, zero, zero, zero)
	kcChainMock.AddTokensPair(token2Erc20, ticker2, false, true, zero, zero, zero)
	kcChainMock.AddTokensPair(token3Erc20, ticker3, false, true, zero, zero, zero)
	kcChainMock.SetLastExecutedEthBatchID(batchNonceOnEthereum)
	kcChainMock.SetLastExecutedEthTxId(txNonceOnEthereum)
	kcChainMock.GetStatusesAfterExecutionHandler = func() []byte {
		return []byte{bridgeCore.Executed, bridgeCore.Rejected, bridgeCore.Executed}
	}
	kcChainMock.SetQuorum(numRelayers)

	relayers := make([]bridgeComponents, 0, numRelayers)
	defer func() {
		for _, r := range relayers {
			_ = r.Close()
		}
	}()

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()
	kcChainMock.ProcessFinishedHandler = func() {
		log.Info("kcChainMock.ProcessFinishedHandler called")
		asyncCancelCall(cancel, time.Second*5)
	}

	for i := 0; i < numRelayers; i++ {
		argsBridgeComponents := createMockBridgeComponentsArgs(i, messengers[i], kcChainMock, ethereumChainMock)
		argsBridgeComponents.Configs.GeneralConfig.Eth.SafeContractAddress = safeContractEthAddress.Hex()
		argsBridgeComponents.Erc20ContractsHolder = erc20ContractsHolder
		relayer, err := factory.NewEthKleverBridgeComponents(argsBridgeComponents)
		require.Nil(t, err)

		kcChainMock.AddRelayer(relayer.KleverRelayerAddress())
		ethereumChainMock.AddRelayer(relayer.EthereumRelayerAddress())

		go func() {
			err = relayer.Start()
			integrationTests.Log.LogIfError(err)
			require.Nil(t, err)
		}()

		relayers = append(relayers, relayer)
	}

	<-ctx.Done()
	time.Sleep(time.Second * 5)

	assert.NotNil(t, kcChainMock.PerformedActionID())
	transfer := kcChainMock.ProposedTransfer()
	require.NotNil(t, transfer)
	require.Equal(t, 3, len(transfer.Transfers))
	assert.Equal(t, batchNonceOnEthereum+1, transfer.BatchId.Uint64())

	assert.Equal(t, destination1.Bytes(), transfer.Transfers[0].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker1)), transfer.Transfers[0].Token)
	assert.Equal(t, value1, transfer.Transfers[0].Amount)
	assert.Equal(t, depositor1, common.BytesToAddress(transfer.Transfers[0].From))
	assert.Equal(t, txNonceOnEthereum+1, transfer.Transfers[0].Nonce.Uint64())
	assert.Equal(t, []byte{bridgeCore.MissingDataProtocolMarker}, transfer.Transfers[0].Data)

	assert.Equal(t, destination2.Bytes(), transfer.Transfers[1].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker2)), transfer.Transfers[1].Token)
	assert.Equal(t, value2, transfer.Transfers[1].Amount)
	assert.Equal(t, depositor2, common.BytesToAddress(transfer.Transfers[1].From))
	assert.Equal(t, txNonceOnEthereum+2, transfer.Transfers[1].Nonce.Uint64())
	assert.Equal(t, []byte{bridgeCore.MissingDataProtocolMarker}, transfer.Transfers[1].Data)

	assert.Equal(t, destination3Sc.Bytes(), transfer.Transfers[2].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker3)), transfer.Transfers[2].Token)
	assert.Equal(t, value3, transfer.Transfers[2].Amount)
	assert.Equal(t, depositor3, common.BytesToAddress(transfer.Transfers[2].From))
	assert.Equal(t, txNonceOnEthereum+3, transfer.Transfers[2].Nonce.Uint64())
	assert.Equal(t, args.expectedScCallData, transfer.Transfers[2].Data)
}

// TestRelayersShouldExecuteTransfersFromEthToKCWithDecimalConversion tests the ETH→KC flow
// when tokens have different decimal configurations. Ethereum tokens commonly use 18 decimals,
// while Klever has a maximum of 8 decimals. The smart contracts handle the conversion at proposal time.
func TestRelayersShouldExecuteTransfersFromEthToKCWithDecimalConversion(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("18 to 8 decimals conversion", func(t *testing.T) {
		// ETH 18 decimals → KDA 8 decimals (scale down by 10^10)
		// Example: 1.5 ETH = 1,500,000,000,000,000,000 wei → 150,000,000 (KDA 8 decimals)
		testRelayersShouldExecuteTransfersFromEthToKCWithDecimalConversion(t, decimalConversionTestArgs{
			ethDecimals:    18,
			kdaDecimals:    8,
			ethAmount:      big.NewInt(0).Mul(big.NewInt(15), big.NewInt(1e17)), // 1.5 ETH in wei
			expectedKdaAmt: big.NewInt(150_000_000),                             // 1.5 in KDA 8 decimals
		})
	})
	t.Run("18 to 6 decimals conversion", func(t *testing.T) {
		// ETH 18 decimals → KDA 6 decimals (scale down by 10^12)
		// Example: 2.5 tokens = 2,500,000,000,000,000,000 → 2,500,000 (KDA 6 decimals)
		testRelayersShouldExecuteTransfersFromEthToKCWithDecimalConversion(t, decimalConversionTestArgs{
			ethDecimals:    18,
			kdaDecimals:    6,
			ethAmount:      big.NewInt(0).Mul(big.NewInt(25), big.NewInt(1e17)), // 2.5 tokens in ETH 18 decimals
			expectedKdaAmt: big.NewInt(2_500_000),                               // 2.5 in KDA 6 decimals
		})
	})
	t.Run("same decimals no conversion", func(t *testing.T) {
		// 8 decimals → 8 decimals (no conversion needed)
		// Example: WKLV already has 8 decimals on both sides
		testRelayersShouldExecuteTransfersFromEthToKCWithDecimalConversion(t, decimalConversionTestArgs{
			ethDecimals:    8,
			kdaDecimals:    8,
			ethAmount:      big.NewInt(100_000_000), // 1.0 token in 8 decimals
			expectedKdaAmt: big.NewInt(100_000_000), // same amount
		})
	})
	t.Run("6 to 6 decimals no conversion USDC style", func(t *testing.T) {
		// USDC: 6 decimals on both chains
		testRelayersShouldExecuteTransfersFromEthToKCWithDecimalConversion(t, decimalConversionTestArgs{
			ethDecimals:    6,
			kdaDecimals:    6,
			ethAmount:      big.NewInt(1_000_000), // 1.0 USDC
			expectedKdaAmt: big.NewInt(1_000_000), // same amount
		})
	})
}

type decimalConversionTestArgs struct {
	ethDecimals    uint8
	kdaDecimals    uint8
	ethAmount      *big.Int
	expectedKdaAmt *big.Int
}

func testRelayersShouldExecuteTransfersFromEthToKCWithDecimalConversion(t *testing.T, args decimalConversionTestArgs) {
	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()

	tokenErc20 := testsCommon.CreateRandomEthereumAddress()
	ticker := "tck-dec001"

	destination := testsCommon.CreateRandomKCAddress()
	depositor := testsCommon.CreateRandomEthereumAddress()

	tokens := []common.Address{tokenErc20}
	availableBalances := []*big.Int{args.ethAmount}

	erc20ContractsHolder := createMockErc20ContractsHolder(tokens, safeContractEthAddress, availableBalances)

	batchNonceOnEthereum := uint64(100)
	txNonceOnEthereum := uint64(500)
	batch := contract.Batch{
		Nonce:                  big.NewInt(int64(batchNonceOnEthereum) + 1),
		BlockNumber:            0,
		LastUpdatedBlockNumber: 0,
		DepositsCount:          1,
	}

	numRelayers := 3
	ethereumChainMock := mock.NewEthereumChainMock()
	ethereumChainMock.AddBatch(batch)
	ethereumChainMock.AddDepositToBatch(batchNonceOnEthereum+1, contract.Deposit{
		Nonce:        big.NewInt(int64(txNonceOnEthereum) + 1),
		TokenAddress: tokenErc20,
		Amount:       args.ethAmount,
		Depositor:    depositor,
		Recipient:    destination.AddressSlice(),
		Status:       0,
	})
	ethereumChainMock.AddBatch(batch)
	ethereumChainMock.SetQuorum(numRelayers)
	ethereumChainMock.SetFinalNonce(batchNonceOnEthereum + 1)

	// Configure as native token on Ethereum (not mint/burn)
	ethereumChainMock.UpdateNativeTokens(tokenErc20, true)
	ethereumChainMock.UpdateMintBurnTokens(tokenErc20, false)
	ethereumChainMock.UpdateTotalBalances(tokenErc20, args.ethAmount)

	kcChainMock := mock.NewKleverBlockchainMock()
	kcChainMock.AddTokensPair(tokenErc20, ticker, false, true, zero, zero, zero)

	// Set up decimal conversion: convertedAmount = (amount * multiplier) / divisor
	// For 18→8: multiplier=1, divisor=10^10
	// For 18→6: multiplier=1, divisor=10^12
	// For same decimals: multiplier=1, divisor=1
	var multiplier, divisor *big.Int
	if args.ethDecimals > args.kdaDecimals {
		multiplier = big.NewInt(1)
		decimalDiff := args.ethDecimals - args.kdaDecimals
		divisor = new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimalDiff)), nil)
	} else if args.ethDecimals < args.kdaDecimals {
		decimalDiff := args.kdaDecimals - args.ethDecimals
		multiplier = new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimalDiff)), nil)
		divisor = big.NewInt(1)
	} else {
		multiplier = big.NewInt(1)
		divisor = big.NewInt(1)
	}
	kcChainMock.SetDecimalConversion(ticker, multiplier, divisor)

	kcChainMock.SetLastExecutedEthBatchID(batchNonceOnEthereum)
	kcChainMock.SetLastExecutedEthTxId(txNonceOnEthereum)
	kcChainMock.GetStatusesAfterExecutionHandler = func() []byte {
		return []byte{bridgeCore.Executed}
	}
	kcChainMock.SetQuorum(numRelayers)

	relayers := make([]bridgeComponents, 0, numRelayers)
	defer closeRelayers(relayers)

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()
	kcChainMock.ProcessFinishedHandler = func() {
		log.Info("kcChainMock.ProcessFinishedHandler called - decimal conversion test")
		asyncCancelCall(cancel, time.Second*5)
	}

	for i := 0; i < numRelayers; i++ {
		argsBridgeComponents := createMockBridgeComponentsArgs(i, messengers[i], kcChainMock, ethereumChainMock)
		argsBridgeComponents.Configs.GeneralConfig.Eth.SafeContractAddress = safeContractEthAddress.Hex()
		argsBridgeComponents.Erc20ContractsHolder = erc20ContractsHolder

		relayer, err := factory.NewEthKleverBridgeComponents(argsBridgeComponents)
		require.NoError(t, err)

		kcChainMock.AddRelayer(relayer.KleverRelayerAddress())
		ethereumChainMock.AddRelayer(relayer.EthereumRelayerAddress())

		relayers = append(relayers, relayer)
		go func(r bridgeComponents) {
			if err := r.Start(); err != nil {
				integrationTests.Log.LogIfError(err)
				assert.NoError(t, err)
			}
		}(relayer)
	}

	<-ctx.Done()
	time.Sleep(time.Second * 5)

	// Verify the transfer was proposed and executed
	assert.NotNil(t, kcChainMock.PerformedActionID())
	transfer := kcChainMock.ProposedTransfer()
	require.NotNil(t, transfer)
	require.Equal(t, 1, len(transfer.Transfers))
	assert.Equal(t, batchNonceOnEthereum+1, transfer.BatchId.Uint64())

	// Verify transfer details
	assert.Equal(t, destination.Bytes(), transfer.Transfers[0].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker)), transfer.Transfers[0].Token)
	assert.Equal(t, depositor, common.BytesToAddress(transfer.Transfers[0].From))
	assert.Equal(t, txNonceOnEthereum+1, transfer.Transfers[0].Nonce.Uint64())
	assert.Equal(t, []byte{bridgeCore.MissingDataProtocolMarker}, transfer.Transfers[0].Data)

	// Verify amount fields - this is the key validation for decimal conversion
	// Amount should be the original ETH amount (in ETH decimals)
	assert.Equal(t, args.ethAmount, transfer.Transfers[0].Amount,
		"Amount should be the original ETH amount in ETH decimals")

	// ConvertedAmount should be the converted KDA amount (in KDA decimals)
	assert.Equal(t, args.expectedKdaAmt, transfer.Transfers[0].ConvertedAmount,
		"ConvertedAmount should be the converted amount in KDA decimals (ETH %d → KDA %d decimals)",
		args.ethDecimals, args.kdaDecimals)
}

func createMockBridgeComponentsArgs(
	index int,
	messenger p2p.Messenger,
	kcMock *mock.KleverBlockchainMock,
	ethereumChainMock *mock.EthereumChainMock,
) factory.ArgsEthereumToKleverBridge {

	generalConfigs := CreateBridgeComponentsConfig(index, "testdata", noGasStationURL)
	return factory.ArgsEthereumToKleverBridge{
		Configs: config.Configs{
			GeneralConfig:   generalConfigs,
			ApiRoutesConfig: config.ApiRoutesConfig{},
			FlagsConfig: config.ContextFlagsConfig{
				RestApiInterface: bridgeCore.WebServerOffString,
			},
		},
		Proxy:                     kcMock,
		ClientWrapper:             ethereumChainMock,
		Messenger:                 messenger,
		StatusStorer:              testsCommon.NewStorerMock(),
		TimeForBootstrap:          time.Second * 5,
		TimeBeforeRepeatJoin:      time.Second * 30,
		MetricsHolder:             status.NewMetricsHolder(),
		AppStatusHandler:          &statusHandler.AppStatusHandlerStub{},
		KleverClientStatusHandler: &testsCommon.StatusHandlerStub{},
	}
}
