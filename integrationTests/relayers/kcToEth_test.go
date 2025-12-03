//go:build !slow

package relayers

import (
	"context"
	"fmt"
	"math/big"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/factory"
	"github.com/klever-io/klv-bridge-eth-go/integrationTests"
	"github.com/klever-io/klv-bridge-eth-go/integrationTests/mock"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var zero = big.NewInt(0)
var relayerEthBalance = big.NewInt(1000000000)

func asyncCancelCall(cancelHandler func(), delay time.Duration) {
	go func() {
		time.Sleep(delay)
		cancelHandler()
	}()
}

func TestRelayersShouldExecuteSimpleTransfersFromKCToEth(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numTransactions := 2
	deposits, tokensAddresses, erc20Map := createTransactions(numTransactions)

	tokens, availableBalances := availableTokensMapToSlices(erc20Map)
	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()
	erc20ContractsHolder := createMockErc20ContractsHolder(tokens, safeContractEthAddress, availableBalances)

	numRelayers := 3
	ethereumChainMock := mock.NewEthereumChainMock()
	ethereumChainMock.SetQuorum(numRelayers)
	expectedStatuses := []byte{bridgeCore.Executed, bridgeCore.Rejected}
	ethereumChainMock.GetStatusesAfterExecutionHandler = func() ([]byte, bool) {
		if callIsFromBalanceValidator() {
			// statuses can not be final at this point as the batch was not executed yet
			return expectedStatuses, false
		}

		return expectedStatuses, true
	}
	ethereumChainMock.BalanceAtCalled = func(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
		return relayerEthBalance, nil
	}
	kcChainMock := mock.NewKleverBlockchainMock()
	for i := 0; i < len(deposits); i++ {
		ethereumChainMock.UpdateNativeTokens(tokensAddresses[i], false)
		ethereumChainMock.UpdateMintBurnTokens(tokensAddresses[i], true)
		ethereumChainMock.UpdateMintBalances(tokensAddresses[i], zero)
		ethereumChainMock.UpdateBurnBalances(tokensAddresses[i], zero)

		kcChainMock.AddTokensPair(tokensAddresses[i], deposits[i].Ticker, true, true, zero, zero, deposits[i].Amount)
	}
	pendingBatch := mock.KleverBlockchainPendingBatch{
		Nonce:                    big.NewInt(1),
		KleverBlockchainDeposits: deposits,
	}

	kcChainMock.SetPendingBatch(&pendingBatch)
	kcChainMock.SetQuorum(numRelayers)

	relayers := make([]bridgeComponents, 0, numRelayers)
	defer closeRelayers(relayers)

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1200)
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

	// let all transactions propagate
	time.Sleep(time.Second * 5)

	checkTestStatus(t, kcChainMock, ethereumChainMock, numTransactions, deposits, tokensAddresses)
}

func callIsFromBalanceValidator() bool {
	callStack := string(debug.Stack())
	return strings.Contains(callStack, "(*balanceValidator).getTotalTransferAmountInPendingKlvBatches")
}

func TestRelayersShouldExecuteTransfersFromKCToEthIfTransactionsAppearInBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("simple tokens transfers", func(t *testing.T) {
		testRelayersShouldExecuteTransfersFromKCToEthIfTransactionsAppearInBatch(t, false)
	})
	t.Run("native tokens transfers", func(t *testing.T) {
		testRelayersShouldExecuteTransfersFromKCToEthIfTransactionsAppearInBatch(t, true)
	})
}

func testRelayersShouldExecuteTransfersFromKCToEthIfTransactionsAppearInBatch(t *testing.T, withNativeTokens bool) {
	numTransactions := 2
	deposits, tokensAddresses, erc20Map := createTransactions(numTransactions)

	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()
	tokens, availableBalances := availableTokensMapToSlices(erc20Map)
	erc20ContractsHolder := createMockErc20ContractsHolder(tokens, safeContractEthAddress, availableBalances)

	numRelayers := 3
	ethereumChainMock := mock.NewEthereumChainMock()
	ethereumChainMock.SetQuorum(numRelayers)
	expectedStatuses := []byte{bridgeCore.Executed, bridgeCore.Rejected}
	ethereumChainMock.GetStatusesAfterExecutionHandler = func() ([]byte, bool) {
		if callIsFromBalanceValidator() {
			// statuses can not be final at this point as the batch was not executed yet
			return expectedStatuses, false
		}

		return expectedStatuses, true
	}
	ethereumChainMock.BalanceAtCalled = func(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
		return relayerEthBalance, nil
	}
	kcChainMock := mock.NewKleverBlockchainMock()
	for i := 0; i < len(deposits); i++ {
		nativeBalanceValue := deposits[i].Amount

		if !withNativeTokens {
			ethereumChainMock.UpdateNativeTokens(tokensAddresses[i], true)
			ethereumChainMock.UpdateMintBurnTokens(tokensAddresses[i], false)
			ethereumChainMock.UpdateTotalBalances(tokensAddresses[i], nativeBalanceValue)

			kcChainMock.AddTokensPair(tokensAddresses[i], deposits[i].Ticker, withNativeTokens, true, zero, nativeBalanceValue, nativeBalanceValue)
		} else {
			ethereumChainMock.UpdateNativeTokens(tokensAddresses[i], false)
			ethereumChainMock.UpdateMintBurnTokens(tokensAddresses[i], true)
			ethereumChainMock.UpdateBurnBalances(tokensAddresses[i], zero)
			ethereumChainMock.UpdateMintBalances(tokensAddresses[i], zero)

			kcChainMock.AddTokensPair(tokensAddresses[i], deposits[i].Ticker, withNativeTokens, true, zero, zero, nativeBalanceValue)
		}
	}
	pendingBatch := mock.KleverBlockchainPendingBatch{
		Nonce:                    big.NewInt(1),
		KleverBlockchainDeposits: deposits,
	}
	kcChainMock.SetPendingBatch(&pendingBatch)
	kcChainMock.SetQuorum(numRelayers)

	ethereumChainMock.ProposeMultiTransferKdaBatchCalled = func() {
		deposit := deposits[0]

		kcChainMock.AddDepositToCurrentBatch(deposit)
	}

	relayers := make([]bridgeComponents, 0, numRelayers)
	defer closeRelayers(relayers)

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1200)
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

	// let all transactions propagate
	time.Sleep(time.Second * 5)

	checkTestStatus(t, kcChainMock, ethereumChainMock, numTransactions, deposits, tokensAddresses)
}

func createTransactions(n int) ([]mock.KleverBlockchainDeposit, []common.Address, map[common.Address]*big.Int) {
	tokensAddresses := make([]common.Address, 0, n)
	deposits := make([]mock.KleverBlockchainDeposit, 0, n)
	erc20 := make(map[common.Address]*big.Int)
	for i := 0; i < n; i++ {
		deposit, tokenAddress := createTransaction(i)
		tokensAddresses = append(tokensAddresses, tokenAddress)
		deposits = append(deposits, deposit)

		val, found := erc20[tokenAddress]
		if !found {
			val = big.NewInt(0)
			erc20[tokenAddress] = val
		}
		val.Add(val, deposit.Amount)
	}

	return deposits, tokensAddresses, erc20
}

func createTransaction(index int) (mock.KleverBlockchainDeposit, common.Address) {
	tokenAddress := testsCommon.CreateRandomEthereumAddress()
	amount := big.NewInt(int64(index*1000) + 500) // 0 as amount is not relevant

	return mock.KleverBlockchainDeposit{
		From:            testsCommon.CreateRandomKCAddress(),
		To:              testsCommon.CreateRandomEthereumAddress(),
		Ticker:          fmt.Sprintf("tck-00000%d", index+1),
		Amount:          amount,
		ConvertedAmount: amount, // default to same as Amount for now
	}, tokenAddress
}

func checkTestStatus(
	t *testing.T,
	kcChainMock *mock.KleverBlockchainMock,
	ethereumChainMock *mock.EthereumChainMock,
	numTransactions int,
	deposits []mock.KleverBlockchainDeposit,
	tokensAddresses []common.Address,
) {
	transactions := kcChainMock.GetAllSentTransactions(context.Background())
	assert.Equal(t, 5, len(transactions))
	assert.Nil(t, kcChainMock.ProposedTransfer())
	assert.NotNil(t, kcChainMock.PerformedActionID())

	transfer := ethereumChainMock.GetLastProposedTransfer()
	require.NotNil(t, transfer)

	require.Equal(t, numTransactions, len(transfer.Amounts))

	for i := 0; i < len(transfer.Amounts); i++ {
		assert.Equal(t, deposits[i].To, transfer.Recipients[i])
		assert.Equal(t, tokensAddresses[i], transfer.Tokens[i])
		assert.Equal(t, deposits[i].Amount, transfer.Amounts[i])
	}
}

// TestRelayersShouldExecuteTransfersFromKCToEthWithDecimalConversion tests the KC→ETH flow
// when tokens have different decimal configurations. In this direction:
// - Amount field contains the ETH-side amount (up to 18 decimals) - used for Ethereum execution
// - ConvertedAmount field contains the KDA-side amount (max 8 decimals) - original amount, used for refunds
func TestRelayersShouldExecuteTransfersFromKCToEthWithDecimalConversion(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("8 to 18 decimals conversion", func(t *testing.T) {
		// KDA 8 decimals → ETH 18 decimals (scale up by 10^10)
		// Example: 1.5 tokens = 150,000,000 (KDA 8 decimals) → 1,500,000,000,000,000,000 wei (ETH 18 decimals)
		testRelayersShouldExecuteTransfersFromKCToEthWithDecimalConversion(t, kcToEthDecimalConversionTestArgs{
			kdaDecimals:    8,
			ethDecimals:    18,
			kdaAmount:      big.NewInt(150_000_000),                             // 1.5 in KDA 8 decimals
			expectedEthAmt: big.NewInt(0).Mul(big.NewInt(15), big.NewInt(1e17)), // 1.5 ETH in wei
		})
	})
	t.Run("6 to 18 decimals conversion", func(t *testing.T) {
		// KDA 6 decimals → ETH 18 decimals (scale up by 10^12)
		// Example: 2.5 tokens = 2,500,000 (KDA 6 decimals) → 2,500,000,000,000,000,000 (ETH 18 decimals)
		testRelayersShouldExecuteTransfersFromKCToEthWithDecimalConversion(t, kcToEthDecimalConversionTestArgs{
			kdaDecimals:    6,
			ethDecimals:    18,
			kdaAmount:      big.NewInt(2_500_000),                               // 2.5 in KDA 6 decimals
			expectedEthAmt: big.NewInt(0).Mul(big.NewInt(25), big.NewInt(1e17)), // 2.5 ETH in wei
		})
	})
	t.Run("same decimals no conversion", func(t *testing.T) {
		// 8 decimals → 8 decimals (no conversion needed)
		// Example: WKLV already has 8 decimals on both sides
		testRelayersShouldExecuteTransfersFromKCToEthWithDecimalConversion(t, kcToEthDecimalConversionTestArgs{
			kdaDecimals:    8,
			ethDecimals:    8,
			kdaAmount:      big.NewInt(100_000_000), // 1.0 token in 8 decimals
			expectedEthAmt: big.NewInt(100_000_000), // same amount
		})
	})
	t.Run("6 to 6 decimals no conversion USDC style", func(t *testing.T) {
		// USDC: 6 decimals on both chains
		testRelayersShouldExecuteTransfersFromKCToEthWithDecimalConversion(t, kcToEthDecimalConversionTestArgs{
			kdaDecimals:    6,
			ethDecimals:    6,
			kdaAmount:      big.NewInt(1_000_000), // 1.0 USDC
			expectedEthAmt: big.NewInt(1_000_000), // same amount
		})
	})
}

type kcToEthDecimalConversionTestArgs struct {
	kdaDecimals    uint8
	ethDecimals    uint8
	kdaAmount      *big.Int
	expectedEthAmt *big.Int
}

func testRelayersShouldExecuteTransfersFromKCToEthWithDecimalConversion(t *testing.T, args kcToEthDecimalConversionTestArgs) {
	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()

	tokenErc20 := testsCommon.CreateRandomEthereumAddress()
	ticker := "tck-dec001"

	from := testsCommon.CreateRandomKCAddress()
	to := testsCommon.CreateRandomEthereumAddress()

	// In KC→ETH flow:
	// - Amount is the ETH-side amount (scaled up if needed)
	// - ConvertedAmount is the original KDA-side amount
	deposit := mock.KleverBlockchainDeposit{
		From:            from,
		To:              to,
		Ticker:          ticker,
		Amount:          args.expectedEthAmt, // ETH-side amount (already converted by KdaSafe.createTransaction)
		ConvertedAmount: args.kdaAmount,      // Original KDA amount (for refunds)
	}

	deposits := []mock.KleverBlockchainDeposit{deposit}
	tokens := []common.Address{tokenErc20}
	availableBalances := []*big.Int{args.expectedEthAmt}

	erc20ContractsHolder := createMockErc20ContractsHolder(tokens, safeContractEthAddress, availableBalances)

	numRelayers := 3
	ethereumChainMock := mock.NewEthereumChainMock()
	ethereumChainMock.SetQuorum(numRelayers)
	expectedStatuses := []byte{bridgeCore.Executed}
	ethereumChainMock.GetStatusesAfterExecutionHandler = func() ([]byte, bool) {
		if callIsFromBalanceValidator() {
			return expectedStatuses, false
		}
		return expectedStatuses, true
	}
	ethereumChainMock.BalanceAtCalled = func(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
		return relayerEthBalance, nil
	}

	// Configure as native token on Klever (mint/burn on Ethereum)
	// For mint/burn tokens, both balances start at 0
	ethereumChainMock.UpdateNativeTokens(tokenErc20, false)
	ethereumChainMock.UpdateMintBurnTokens(tokenErc20, true)
	ethereumChainMock.UpdateMintBalances(tokenErc20, zero)
	ethereumChainMock.UpdateBurnBalances(tokenErc20, zero)

	kcChainMock := mock.NewKleverBlockchainMock()
	// For mint/burn tokens native on Klever: totalBalance=0, mintBalances=0, burnBalances=kdaAmount (KC-side)
	// The last parameter is burnBalances which should be in KDA decimals
	kcChainMock.AddTokensPair(tokenErc20, ticker, true, true, zero, zero, args.kdaAmount)

	// Set up decimal conversion so balance validator can convert ETH amounts to KDA amounts
	// For KC→ETH: ETH decimals = kdaAmount * multiplier / divisor when converting from ETH to KDA
	// The conversion mock does: convertedAmount = (amount * multiplier) / divisor
	// To convert from ETH to KDA: we need divisor to scale down (if ETH > KDA decimals)
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

	pendingBatch := mock.KleverBlockchainPendingBatch{
		Nonce:                    big.NewInt(1),
		KleverBlockchainDeposits: deposits,
	}
	kcChainMock.SetPendingBatch(&pendingBatch)
	kcChainMock.SetQuorum(numRelayers)

	relayers := make([]bridgeComponents, 0, numRelayers)
	defer closeRelayers(relayers)

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*300)
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

	// Verify the transfer was processed
	transactions := kcChainMock.GetAllSentTransactions(context.Background())
	assert.Equal(t, 5, len(transactions))
	assert.Nil(t, kcChainMock.ProposedTransfer())
	assert.NotNil(t, kcChainMock.PerformedActionID())

	transfer := ethereumChainMock.GetLastProposedTransfer()
	require.NotNil(t, transfer)
	require.Equal(t, 1, len(transfer.Amounts))

	// Verify the transfer to Ethereum uses the ETH-side amount (scaled up)
	assert.Equal(t, to, transfer.Recipients[0])
	assert.Equal(t, tokenErc20, transfer.Tokens[0])

	// This is the key validation: Ethereum receives the ETH-decimals amount
	assert.Equal(t, args.expectedEthAmt, transfer.Amounts[0],
		"Ethereum should receive the ETH-side amount (KDA %d → ETH %d decimals)",
		args.kdaDecimals, args.ethDecimals)
}
