package interactors

import (
	"context"
	"fmt"

	"github.com/klever-io/klever-go/data/transaction"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
)

// ProxyStub -
type ProxyStub struct {
	GetNetworkConfigCalled              func(ctx context.Context) (*models.NetworkConfig, error)
	SendTransactionCalled               func(ctx context.Context, transaction *transaction.Transaction) (string, error)
	SendTransactionsCalled              func(ctx context.Context, txs []*transaction.Transaction) ([]string, error)
	ExecuteVMQueryCalled                func(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error)
	GetAccountCalled                    func(ctx context.Context, address address.Address) (*models.Account, error)
	GetNetworkStatusCalled              func(ctx context.Context) (*models.NodeOverview, error)
	GetShardOfAddressCalled             func(ctx context.Context, bech32Address string) (uint32, error)
	GetKDATokenDataCalled               func(ctx context.Context, address address.Address, tokenIdentifier string) (*models.KDAFungibleTokenData, error)
	GetTransactionInfoWithResultsCalled func(_ context.Context, _ string) (*models.TransactionData, error)
	ProcessTransactionStatusCalled      func(ctx context.Context, hexTxHash string) (transaction.Transaction_TXResult, error)
	EstimateTransactionFeesCalled       func(ctx context.Context, txs *transaction.Transaction) (*transaction.FeesResponse, error)
}

// GetNetworkConfig -
func (eps *ProxyStub) GetNetworkConfig(ctx context.Context) (*models.NetworkConfig, error) {
	if eps.GetNetworkConfigCalled != nil {
		return eps.GetNetworkConfigCalled(ctx)
	}

	return &models.NetworkConfig{}, nil
}

// SendTransaction -
func (eps *ProxyStub) SendTransaction(ctx context.Context, transaction *transaction.Transaction) (string, error) {
	if eps.SendTransactionCalled != nil {
		return eps.SendTransactionCalled(ctx, transaction)
	}

	return "", nil
}

// SendTransactions -
func (eps *ProxyStub) SendTransactions(ctx context.Context, txs []*transaction.Transaction) ([]string, error) {
	if eps.SendTransactionsCalled != nil {
		return eps.SendTransactionsCalled(ctx, txs)
	}

	return make([]string, 0), nil
}

// ExecuteVMQuery -
func (eps *ProxyStub) ExecuteVMQuery(ctx context.Context, vmRequest *models.VmValueRequest) (*models.VmValuesResponseData, error) {
	if eps.ExecuteVMQueryCalled != nil {
		return eps.ExecuteVMQueryCalled(ctx, vmRequest)
	}

	return &models.VmValuesResponseData{}, nil
}

// GetAccount -
func (eps *ProxyStub) GetAccount(ctx context.Context, address address.Address) (*models.Account, error) {
	if eps.GetAccountCalled != nil {
		return eps.GetAccountCalled(ctx, address)
	}

	return &models.Account{}, nil
}

// GetNetworkStatus -
func (eps *ProxyStub) GetNetworkStatus(ctx context.Context) (*models.NodeOverview, error) {
	if eps.GetNetworkStatusCalled != nil {
		return eps.GetNetworkStatusCalled(ctx)
	}

	return nil, fmt.Errorf("not implemented")
}

// GetShardOfAddress -
func (eps *ProxyStub) GetShardOfAddress(ctx context.Context, bech32Address string) (uint32, error) {
	if eps.GetShardOfAddressCalled != nil {
		return eps.GetShardOfAddressCalled(ctx, bech32Address)
	}

	return 0, fmt.Errorf("not implemented")
}

// GetKDATokenData -
func (eps *ProxyStub) GetKDATokenData(ctx context.Context, address address.Address, tokenIdentifier string) (*models.KDAFungibleTokenData, error) {
	if eps.GetKDATokenDataCalled != nil {
		return eps.GetKDATokenDataCalled(ctx, address, tokenIdentifier)
	}

	return nil, fmt.Errorf("not implemented")
}

// GetTransactionInfoWithResults -
func (eps *ProxyStub) GetTransactionInfoWithResults(ctx context.Context, hash string) (*models.TransactionData, error) {
	if eps.GetTransactionInfoWithResultsCalled != nil {
		return eps.GetTransactionInfoWithResultsCalled(ctx, hash)
	}

	return nil, fmt.Errorf("not implemented")
}

// ProcessTransactionStatus -
func (eps *ProxyStub) ProcessTransactionStatus(ctx context.Context, hexTxHash string) (transaction.Transaction_TXResult, error) {
	if eps.ProcessTransactionStatusCalled != nil {
		return eps.ProcessTransactionStatusCalled(ctx, hexTxHash)
	}

	return transaction.Transaction_FAILED, nil
}

// GetTransactionInfoWithResults -
func (eps *ProxyStub) EstimateTransactionFees(ctx context.Context, txs *transaction.Transaction) (*transaction.FeesResponse, error) {
	if eps.EstimateTransactionFeesCalled != nil {
		return eps.EstimateTransactionFeesCalled(ctx, txs)
	}

	return &transaction.FeesResponse{
		CostResponse: &transaction.CostResponse{},
		KDAFees:      &transaction.Transaction_KDAFee{},
	}, nil
}

// IsInterfaceNil -
func (eps *ProxyStub) IsInterfaceNil() bool {
	return eps == nil
}
