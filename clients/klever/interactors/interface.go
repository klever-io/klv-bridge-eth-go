package interactors

import (
	"context"

	"github.com/klever-io/klever-go/data/transaction"
	sdkAddress "github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
)

// AddressNonceHandler defines the component able to handler address nonces
type AddressNonceHandler interface {
	ApplyNonceAndGasPrice(ctx context.Context, tx *transaction.Transaction) error
	ReSendTransactionsIfRequired(ctx context.Context) error
	SendTransaction(ctx context.Context, tx *transaction.Transaction) (string, error)
	DropTransactions()
	IsInterfaceNil() bool
}

// TransactionNonceHandlerV2 defines the component able to apply nonce for a given frontend transaction.
type TransactionNonceHandlerV2 interface {
	ApplyNonceAndGasPrice(ctx context.Context, address sdkAddress.Address, tx *transaction.Transaction) error
	SendTransaction(ctx context.Context, tx *transaction.Transaction) (string, error)
	Close() error
	IsInterfaceNil() bool
}
