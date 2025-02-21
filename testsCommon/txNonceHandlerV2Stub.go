package testsCommon

import (
	"context"

	"github.com/klever-io/klever-go/data/transaction"
	sdkAddress "github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
)

// TxNonceHandlerV2Stub -
type TxNonceHandlerV2Stub struct {
	ApplyNonceAndGasPriceCalled func(ctx context.Context, address sdkAddress.Address, tx *transaction.Transaction) error
	SendTransactionCalled       func(ctx context.Context, tx *transaction.Transaction) (string, error)
	ForceNonceReFetchCalled     func(address sdkAddress.Address) error
	CloseCalled                 func() error
}

// ApplyNonceAndGasPrice -
func (stub *TxNonceHandlerV2Stub) ApplyNonceAndGasPrice(ctx context.Context, address sdkAddress.Address, tx *transaction.Transaction) error {
	if stub.ApplyNonceAndGasPriceCalled != nil {
		return stub.ApplyNonceAndGasPriceCalled(ctx, address, tx)
	}

	return nil
}

// SendTransaction -
func (stub *TxNonceHandlerV2Stub) SendTransaction(ctx context.Context, tx *transaction.Transaction) (string, error) {
	if stub.SendTransactionCalled != nil {
		return stub.SendTransactionCalled(ctx, tx)
	}

	return "", nil
}

// Close -
func (stub *TxNonceHandlerV2Stub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (stub *TxNonceHandlerV2Stub) IsInterfaceNil() bool {
	return stub == nil
}
