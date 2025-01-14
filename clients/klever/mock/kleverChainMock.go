package mock

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klever-go-sdk/core/address"
	"github.com/klever-io/klever-go-sdk/provider"
	"github.com/klever-io/klv-bridge-eth-go/integrationTests"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	logger "github.com/multiversx/mx-chain-logger-go"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

var log = logger.GetOrCreate("integrationTests/mock")

// KleverChainMock -
type KleverChainMock struct {
	*kleverContractStateMock
	mutState         sync.RWMutex
	sentTransactions map[string]*transaction.FrontendTransaction
	accounts         *kleverAccountsMock
}

// NewKleverChainMock -
func NewKleverChainMock() *KleverChainMock {
	return &KleverChainMock{
		kleverContractStateMock: newKleverContractStateMock(),
		sentTransactions:        make(map[string]*transaction.FrontendTransaction),
		accounts:                newKleverAccountsMock(),
	}
}

// GetNetworkConfig -
func (mock *KleverChainMock) GetNetworkConfig(_ context.Context) (*data.NetworkConfig, error) {
	return &data.NetworkConfig{
		ChainID:                  "t",
		LatestTagSoftwareVersion: "",
		MinGasPrice:              1000000000,
		MinTransactionVersion:    1,
	}, nil
}

// GetNetworkStatus -
func (mock *KleverChainMock) GetNetworkStatus(_ context.Context, _ uint32) (*data.NetworkStatus, error) {
	return &data.NetworkStatus{}, nil
}

// GetShardOfAddress -
func (mock *KleverChainMock) GetShardOfAddress(_ context.Context, _ string) (uint32, error) {
	return 0, nil
}

// SendTransaction -
func (mock *KleverChainMock) SendTransaction(_ context.Context, transaction *transaction.FrontendTransaction) (string, error) {
	if transaction == nil {
		panic("nil transaction")
	}

	addrAsBech32 := transaction.Sender
	addressHandler, err := address.NewAddress(addrAsBech32)
	if err != nil {
		panic(fmt.Sprintf("%v while creating address handler for string %s", err, addrAsBech32))
	}

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, transaction)
	if err != nil {
		panic(err)
	}

	log.Info("sent Klever transaction", "sender", addrAsBech32, "data", string(transaction.Data))

	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.sentTransactions[string(hash)] = transaction
	mock.accounts.updateNonce(addressHandler, transaction.Nonce)

	mock.processTransaction(transaction)

	return hex.EncodeToString(hash), nil
}

// SendTransactions -
func (mock *KleverChainMock) SendTransactions(ctx context.Context, txs []*transaction.FrontendTransaction) ([]string, error) {
	hashes := make([]string, 0, len(txs))
	for _, tx := range txs {
		hash, _ := mock.SendTransaction(ctx, tx)
		hashes = append(hashes, hash)
	}

	return hashes, nil
}

// GetAllSentTransactions -
func (mock *KleverChainMock) GetAllSentTransactions(_ context.Context) map[string]*transaction.FrontendTransaction {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	txs := make(map[string]*transaction.FrontendTransaction)
	for hash, tx := range mock.sentTransactions {
		txs[hash] = tx
	}

	return txs
}

// ExecuteVMQuery -
func (mock *KleverChainMock) ExecuteVMQuery(_ context.Context, vmRequest *provider.VmValueRequest) (*provider.VmValuesResponseData, error) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	return mock.processVmRequests(vmRequest)
}

// GetAccount -
func (mock *KleverChainMock) GetAccount(_ context.Context, address address.Address) (*data.Account, error) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	return mock.accounts.getOrCreate(address), nil
}

// GetTransactionInfoWithResults -
func (mock *KleverChainMock) GetTransactionInfoWithResults(_ context.Context, _ string) (*data.TransactionInfo, error) {
	return &data.TransactionInfo{}, nil
}

// ProcessTransactionStatus -
func (mock *KleverChainMock) ProcessTransactionStatus(_ context.Context, _ string) (transaction.TxStatus, error) {
	return "", nil
}

// AddRelayer -
func (mock *KleverChainMock) AddRelayer(address sdkCore.AddressHandler) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.relayers = append(mock.relayers, address.AddressBytes())
}

// SetLastExecutedEthBatchID -
func (mock *KleverChainMock) SetLastExecutedEthBatchID(lastExecutedEthBatchId uint64) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.lastExecutedEthBatchId = lastExecutedEthBatchId
}

// SetLastExecutedEthTxId -
func (mock *KleverChainMock) SetLastExecutedEthTxId(lastExecutedEthTxId uint64) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.lastExecutedEthTxId = lastExecutedEthTxId
}

// AddTokensPair -
func (mock *KleverChainMock) AddTokensPair(erc20 common.Address, ticker string, isNativeToken, isMintBurnToken bool, totalBalance, mintBalances, burnBalances *big.Int) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.addTokensPair(erc20, ticker, isNativeToken, isMintBurnToken, totalBalance, mintBalances, burnBalances)
}

// SetQuorum -
func (mock *KleverChainMock) SetQuorum(quorum int) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.quorum = quorum
}

// PerformedActionID returns the performed action ID
func (mock *KleverChainMock) PerformedActionID() *big.Int {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	return mock.performedAction
}

// ProposedTransfer returns the proposed transfer that matches the performed action ID
func (mock *KleverChainMock) ProposedTransfer() *kleverProposedTransfer {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	if mock.performedAction == nil {
		return nil
	}

	for hash, transfer := range mock.proposedTransfers {
		if HashToActionID(hash).String() == mock.performedAction.String() {
			return transfer
		}
	}

	return nil
}

// SetPendingBatch -
func (mock *KleverChainMock) SetPendingBatch(pendingBatch *KleverPendingBatch) {
	mock.mutState.Lock()
	mock.setPendingBatch(pendingBatch)
	mock.mutState.Unlock()
}

// AddDepositToCurrentBatch -
func (mock *KleverChainMock) AddDepositToCurrentBatch(deposit KleverDeposit) {
	mock.mutState.Lock()
	mock.pendingBatch.KleverDeposits = append(mock.pendingBatch.KleverDeposits, deposit)
	mock.mutState.Unlock()
}

// GetESDTTokenData -
func (mock *KleverChainMock) GetESDTTokenData(_ context.Context, _ address.Address, tokenIdentifier string, _ api.AccountQueryOptions) (*data.ESDTFungibleTokenData, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	isMintBurn, found := mock.mintBurnTokens[tokenIdentifier]
	balance := mock.totalBalances[tokenIdentifier]
	if found && isMintBurn {
		balance = big.NewInt(0)
	}

	return &data.ESDTFungibleTokenData{
		TokenIdentifier: tokenIdentifier,
		Balance:         balance.String(),
	}, nil
}

// IsInterfaceNil -
func (mock *KleverChainMock) IsInterfaceNil() bool {
	return mock == nil
}
