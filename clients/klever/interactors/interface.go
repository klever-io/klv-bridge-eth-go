package interactors

import (
	"context"

	"github.com/klever-io/klever-go/data/transaction"
	sdkAddress "github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
)

// Proxy holds the primitive functions that the multiversx proxy engine supports & implements
// dependency inversion: blockchain package is considered inner business logic, this package is considered "plugin"
type Proxy interface {
	GetNetworkConfig(ctx context.Context) (*models.NetworkConfig, error)
	GetAccount(ctx context.Context, address sdkAddress.Address) (*models.Account, error)
	SendTransaction(ctx context.Context, tx *transaction.Transaction) (string, error)
	SendTransactions(ctx context.Context, txs []*transaction.Transaction) ([]string, error)
	EstimateTransactionFees(ctx context.Context, txs *transaction.Transaction) (*transaction.FeesResponse, error)
	IsInterfaceNil() bool
}

// AddressNonceHandler defines the component able to handler address nonces
type AddressNonceHandler interface {
	ApplyNonceAndGasPrice(ctx context.Context, tx *transaction.Transaction) error
	ReSendTransactionsIfRequired(ctx context.Context) error
	SendTransaction(ctx context.Context, tx *transaction.Transaction) (string, error)
	IsInterfaceNil() bool
}

// TransactionNonceHandlerV2 defines the component able to apply nonce for a given frontend transaction.
type TransactionNonceHandlerV2 interface {
	ApplyNonceAndGasPrice(ctx context.Context, address sdkAddress.Address, tx *transaction.Transaction) error
	SendTransaction(ctx context.Context, tx *transaction.Transaction) (string, error)
	Close() error
	IsInterfaceNil() bool
}
