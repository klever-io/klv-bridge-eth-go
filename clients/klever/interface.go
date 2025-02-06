package klever

import (
	"context"

	"github.com/klever-io/klever-go/data/transaction"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/builders"
)

// NonceTransactionsHandler represents the interface able to handle the current nonce and the transactions resend mechanism
type NonceTransactionsHandler interface {
	ApplyNonceAndGasPrice(ctx context.Context, address address.Address, tx *transaction.Transaction) error
	SendTransaction(ctx context.Context, tx *transaction.Transaction) (string, error)
	Close() error
}

// TokensMapper can convert a token bytes from one chain to another
type TokensMapper interface {
	ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error)
	IsInterfaceNil() bool
}

type txHandler interface {
	SendTransactionReturnHash(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (string, error)
	Close() error
}

type roleProvider interface {
	IsWhitelisted(address address.Address) bool
	IsInterfaceNil() bool
}
