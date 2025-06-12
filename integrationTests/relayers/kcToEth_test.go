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

func TestRelayersShouldExecuteSimpleTransfersFromKcToEth(t *testing.T) {
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

func TestRelayersShouldExecuteTransfersFromKcToEthIfTransactionsAppearInBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("simple tokens transfers", func(t *testing.T) {
		testRelayersShouldExecuteTransfersFromKcToEthIfTransactionsAppearInBatch(t, false)
	})
	t.Run("native tokens transfers", func(t *testing.T) {
		testRelayersShouldExecuteTransfersFromKcToEthIfTransactionsAppearInBatch(t, true)
	})
}

func testRelayersShouldExecuteTransfersFromKcToEthIfTransactionsAppearInBatch(t *testing.T, withNativeTokens bool) {
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

	ethereumChainMock.ProposeMultiTransferEsdtBatchCalled = func() {
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

	return mock.KleverBlockchainDeposit{
		From:   testsCommon.CreateRandomKcAddress(),
		To:     testsCommon.CreateRandomEthereumAddress(),
		Ticker: fmt.Sprintf("tck-00000%d", index+1),
		Amount: big.NewInt(int64(index*1000) + 500), // 0 as amount is not relevant
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
