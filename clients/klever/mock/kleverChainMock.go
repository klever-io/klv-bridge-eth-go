package mock

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klever-go/data/transaction"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	"github.com/klever-io/klv-bridge-eth-go/integrationTests"
	"github.com/multiversx/mx-chain-core-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/data"
)

var log = logger.GetOrCreate("integrationTests/mock")

// KleverChainMock -
type KleverChainMock struct {
	*kleverContractStateMock
	mutState         sync.RWMutex
	sentTransactions map[string]*transaction.Transaction
	accounts         *kleverAccountsMock
}

// NewKleverChainMock -
func NewKleverChainMock() *KleverChainMock {
	return &KleverChainMock{
		kleverContractStateMock: newKleverContractStateMock(),
		sentTransactions:        make(map[string]*transaction.Transaction),
		accounts:                newKleverAccountsMock(),
	}
}

// GetNetworkConfig -
func (mock *KleverChainMock) GetNetworkConfig(_ context.Context) (*models.NetworkConfig, error) {
	return &models.NetworkConfig{
		ChainID: "t",
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
func (mock *KleverChainMock) SendTransaction(_ context.Context, transaction *transaction.Transaction) (string, error) {
	if transaction == nil {
		panic("nil transaction")
	}

	addrAsBech32 := transaction.GetSender()
	addressHandler, err := address.NewAddressFromBytes(addrAsBech32)
	if err != nil {
		panic(fmt.Sprintf("%v while creating address handler for string %s", err, addrAsBech32))
	}

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, transaction)
	if err != nil {
		panic(err)
	}

	var data []byte
	if len(transaction.GetData()) > 0 {
		data = transaction.GetData()[0]
	}

	log.Info("sent Klever transaction", "sender", addrAsBech32, "data", string(data))

	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.sentTransactions[string(hash)] = transaction
	mock.accounts.updateNonce(addressHandler, transaction.GetNonce())

	mock.processTransaction(transaction)

	return hex.EncodeToString(hash), nil
}

// SendTransactions -
func (mock *KleverChainMock) SendTransactions(ctx context.Context, txs []*transaction.Transaction) ([]string, error) {
	hashes := make([]string, 0, len(txs))
	for _, tx := range txs {
		hash, _ := mock.SendTransaction(ctx, tx)
		hashes = append(hashes, hash)
	}

	return hashes, nil
}

// GetAllSentTransactions -
func (mock *KleverChainMock) GetAllSentTransactions(_ context.Context) map[string]*transaction.Transaction {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	txs := make(map[string]*transaction.Transaction)
	for hash, tx := range mock.sentTransactions {
		txs[hash] = tx
	}

	return txs
}

// ExecuteVMQuery -
func (mock *KleverChainMock) ExecuteVMQuery(_ context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	return mock.processVmRequests(vmRequest)
}

// GetAccount -
func (mock *KleverChainMock) GetAccount(_ context.Context, address address.Address) (*models.Account, error) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	return mock.accounts.getOrCreate(address), nil
}

// GetTransactionInfoWithResults -
func (mock *KleverChainMock) GetTransactionInfoWithResults(_ context.Context, _ string) (*data.TransactionInfo, error) {
	return &data.TransactionInfo{}, nil
}

// ProcessTransactionStatus -
func (mock *KleverChainMock) ProcessTransactionStatus(_ context.Context, _ string) (transaction.Transaction_TXResult, error) {
	return transaction.Transaction_SUCCESS, nil
}

// AddRelayer -
func (mock *KleverChainMock) AddRelayer(address address.Address) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.relayers = append(mock.relayers, address.Bytes())
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
func (mock *KleverChainMock) GetESDTTokenData(_ context.Context, _ address.Address, tokenIdentifier string) (*data.ESDTFungibleTokenData, error) {
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

// EstimateTransactionFees -
func (mock *KleverChainMock) EstimateTransactionFees(_ context.Context, txs *transaction.Transaction) (*transaction.FeesResponse, error) {
	return nil, nil
}

// IsInterfaceNil -
func (mock *KleverChainMock) IsInterfaceNil() bool {
	return mock == nil
}
