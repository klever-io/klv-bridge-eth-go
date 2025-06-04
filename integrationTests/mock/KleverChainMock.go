package mock

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	factoryHasher "github.com/klever-io/klever-go/crypto/hashing/factory"
	"github.com/klever-io/klever-go/data/transaction"
	"github.com/klever-io/klever-go/tools"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	"github.com/klever-io/klv-bridge-eth-go/integrationTests"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("integrationTests/mock")

var _ proxy.Proxy = (*KleverBlockchainMock)(nil)

// KleverBlockchainMock -
type KleverBlockchainMock struct {
	*kleverBlockchainContractStateMock
	mutState         sync.RWMutex
	sentTransactions map[string]*transaction.Transaction
	accounts         *kleverBlockchainAccountsMock
}

// NewKleverBlockchainMock -
func NewKleverBlockchainMock() *KleverBlockchainMock {
	return &KleverBlockchainMock{
		kleverBlockchainContractStateMock: newKleverBlockchainContractStateMock(),
		sentTransactions:                  make(map[string]*transaction.Transaction),
		accounts:                          newKleverBlockchainAccountsMock(),
	}
}

// EstimateTransactionFees implements proxy.Proxy.
func (mock *KleverBlockchainMock) EstimateTransactionFees(ctx context.Context, txs *transaction.Transaction) (*transaction.FeesResponse, error) {
	return &transaction.FeesResponse{
		CostResponse: &transaction.CostResponse{
			KAppFee:       1,
			BandwidthFee:  1,
			GasEstimated:  1,
			GasMultiplier: 1,
			RetMessage:    "OK",
		},
	}, nil
}

// ExecuteVMQuery implements proxy.Proxy.
func (mock *KleverBlockchainMock) ExecuteVMQuery(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	return mock.processVmRequests(vmRequest)
}

// GetAccount implements proxy.Proxy.
func (mock *KleverBlockchainMock) GetAccount(ctx context.Context, address address.Address) (*models.Account, error) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	return mock.accounts.getOrCreate(address), nil
}

// GetKDATokenData implements proxy.Proxy.
func (mock *KleverBlockchainMock) GetKDATokenData(ctx context.Context, address address.Address, tokenIdentifier string) (*models.KDAFungibleTokenData, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	isMintBurn, found := mock.mintBurnTokens[tokenIdentifier]
	balance := mock.totalBalances[tokenIdentifier]
	if found && isMintBurn {
		balance = big.NewInt(0)
	}

	return &models.KDAFungibleTokenData{
		TokenIdentifier: tokenIdentifier,
		Balance:         balance.String(),
	}, nil
}

// GetNetworkConfig implements proxy.Proxy.
func (mock *KleverBlockchainMock) GetNetworkConfig(ctx context.Context) (*models.NetworkConfig, error) {
	return &models.NetworkConfig{
		ChainID: "t",
	}, nil
}

// GetNetworkStatus implements proxy.Proxy.
func (mock *KleverBlockchainMock) GetNetworkStatus(ctx context.Context) (*models.NodeOverview, error) {
	return &models.NodeOverview{}, nil
}

// GetTransactionInfoWithResults implements proxy.Proxy.
func (mock *KleverBlockchainMock) GetTransactionInfoWithResults(ctx context.Context, hash string) (*models.TransactionData, error) {
	return &models.TransactionData{}, nil
}

// IsInterfaceNil implements proxy.Proxy.
func (mock *KleverBlockchainMock) IsInterfaceNil() bool {
	return mock == nil
}

// SendTransaction implements proxy.Proxy.
func (mock *KleverBlockchainMock) SendTransaction(ctx context.Context, transaction *transaction.Transaction) (string, error) {
	if transaction == nil {
		panic("nil transaction")
	}

	addrAsBech32 := transaction.GetRawData().GetSender()
	addressHandler, err := address.NewAddress(string(addrAsBech32))
	if err != nil {
		panic(fmt.Sprintf("%v while creating address handler for string %s", err, addrAsBech32))
	}

	hasher, err := factoryHasher.NewHasher("blake2b")
	if err != nil {
		return "", err
	}

	hash, err := tools.CalculateHash(integrationTests.TestMarshalizer, hasher, transaction)
	if err != nil {
		panic(err)
	}

	log.Info("sent Klever Blockchain transaction", "sender", addrAsBech32, "data", transaction.String())

	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.sentTransactions[string(hash)] = transaction
	mock.accounts.updateNonce(addressHandler, transaction.GetRawData().GetNonce())

	mock.processTransaction(transaction)

	return hex.EncodeToString(hash), nil
}

// SendTransactions implements proxy.Proxy.
func (mock *KleverBlockchainMock) SendTransactions(ctx context.Context, txs []*transaction.Transaction) ([]string, error) {
	hashes := make([]string, 0, len(txs))
	for _, tx := range txs {
		hash, _ := mock.SendTransaction(ctx, tx)
		hashes = append(hashes, hash)
	}

	return hashes, nil
}

// AddTokensPair -
func (mock *KleverBlockchainMock) AddTokensPair(erc20 common.Address, ticker string, isNativeToken, isMintBurnToken bool, totalBalance, mintBalances, burnBalances *big.Int) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.addTokensPair(erc20, ticker, isNativeToken, isMintBurnToken, totalBalance, mintBalances, burnBalances)
}

// SetLastExecutedEthBatchID -
func (mock *KleverBlockchainMock) SetLastExecutedEthBatchID(lastExecutedEthBatchId uint64) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.lastExecutedEthBatchId = lastExecutedEthBatchId
}

// SetLastExecutedEthTxId -
func (mock *KleverBlockchainMock) SetLastExecutedEthTxId(lastExecutedEthTxId uint64) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.lastExecutedEthTxId = lastExecutedEthTxId
}

// SetQuorum -
func (mock *KleverBlockchainMock) SetQuorum(quorum int) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.quorum = quorum
}

// AddRelayer -
func (mock *KleverBlockchainMock) AddRelayer(address address.Address) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.relayers = append(mock.relayers, address.Bytes())
}

// PerformedActionID returns the performed action ID
func (mock *KleverBlockchainMock) PerformedActionID() *big.Int {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	return mock.performedAction
}

// ProposedTransfer returns the proposed transfer that matches the performed action ID
func (mock *KleverBlockchainMock) ProposedTransfer() *kleverBlockchainProposedTransfer {
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
func (mock *KleverBlockchainMock) SetPendingBatch(pendingBatch *KleverBlockchainPendingBatch) {
	mock.mutState.Lock()
	mock.setPendingBatch(pendingBatch)
	mock.mutState.Unlock()
}

// AddDepositToCurrentBatch -
func (mock *KleverBlockchainMock) AddDepositToCurrentBatch(deposit KleverBlockchainDeposit) {
	mock.mutState.Lock()
	mock.pendingBatch.KleverBlockchainDeposits = append(mock.pendingBatch.KleverBlockchainDeposits, deposit)
	mock.mutState.Unlock()
}

// GetAllSentTransactions -
func (mock *KleverBlockchainMock) GetAllSentTransactions(_ context.Context) map[string]*transaction.Transaction {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	txs := make(map[string]*transaction.Transaction)
	for hash, tx := range mock.sentTransactions {
		txs[hash] = tx
	}

	return txs
}
