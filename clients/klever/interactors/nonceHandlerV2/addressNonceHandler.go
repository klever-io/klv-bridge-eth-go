package nonceHandlerV2

import (
	"context"
	"sync"

	"github.com/klever-io/klever-go/data/transaction"
	"github.com/klever-io/klever-go/tools"
	"github.com/klever-io/klever-go/tools/check"

	sdkAddress "github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/interactors"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy"
)

// addressNonceHandler is the handler used for one address. It is able to handle the current
// nonce as max(current_stored_nonce, account_nonce). After each call of the getNonce function
// the current_stored_nonce is incremented. This will prevent "nonce too low in transaction"
// errors on the node interceptor. To prevent the "nonce too high in transaction" error,
// a retrial mechanism is implemented. This struct is able to store all sent transactions,
// having a function that sweeps the map in order to resend a transaction or remove them
// because they were executed. This struct is concurrent safe.
type addressNonceHandler struct {
	mut                 sync.RWMutex
	address             sdkAddress.Address
	proxy               proxy.Proxy
	computedNonceWasSet bool
	computedNonce       uint64
	lowestNonce         uint64
	transactions        map[uint64]*transaction.Transaction
}

// NewAddressNonceHandler returns a new instance of a addressNonceHandler
func NewAddressNonceHandler(proxy proxy.Proxy, address sdkAddress.Address) (interactors.AddressNonceHandler, error) {
	if check.IfNil(proxy) {
		return nil, interactors.ErrNilProxy
	}
	if check.IfNil(address) {
		return nil, interactors.ErrNilAddress
	}
	return &addressNonceHandler{
		address:      address,
		proxy:        proxy,
		transactions: make(map[uint64]*transaction.Transaction),
	}, nil
}

// ApplyNonceAndGasPrice will apply the computed nonce to the given FrontendTransaction
func (anh *addressNonceHandler) ApplyNonceAndGasPrice(ctx context.Context, tx *transaction.Transaction) error {
	fees, err := anh.proxy.EstimateTransactionFees(ctx, tx)
	if err != nil {
		return err
	}

	// clone tx to not modify the original one before validating transaction with same nonce
	txClone := tx.Clone()
	txClone.GetRawData().BandwidthFee = fees.BandwidthFee
	txClone.GetRawData().KAppFee = fees.KAppFee

	oldTx := anh.getOlderTxWithSameNonce(txClone)
	if oldTx != nil {
		err := anh.handleTxWithSameNonce(oldTx, txClone)
		if err != nil {
			return err
		}
	}

	nonce, err := anh.getNonceUpdatingCurrent(ctx)
	tx.GetRawData().Nonce = nonce
	if err != nil {
		return err
	}

	tx.GetRawData().BandwidthFee = fees.BandwidthFee
	tx.GetRawData().KAppFee = fees.KAppFee
	return nil
}

func (anh *addressNonceHandler) handleTxWithSameNonce(oldTx *transaction.Transaction, tx *transaction.Transaction) error {
	if oldTx.GetRawData().BandwidthFee != tx.GetRawData().BandwidthFee {
		return nil
	}

	return interactors.ErrTxWithSameNonceAndGasPriceAlreadySent
}

func (anh *addressNonceHandler) getNonceUpdatingCurrent(ctx context.Context) (uint64, error) {
	account, err := anh.proxy.GetAccount(ctx, anh.address)
	if err != nil {
		return 0, err
	}

	if anh.lowestNonce > account.Nonce {
		return account.Nonce, interactors.ErrGapNonce
	}

	anh.mut.Lock()
	defer anh.mut.Unlock()

	if !anh.computedNonceWasSet {
		anh.computedNonce = account.Nonce
		anh.computedNonceWasSet = true

		return anh.computedNonce, nil
	}

	anh.computedNonce++

	return tools.MaxUint64(anh.computedNonce, account.Nonce), nil
}

// ReSendTransactionsIfRequired will resend the cached transactions that still have a nonce greater that the one fetched from the blockchain
func (anh *addressNonceHandler) ReSendTransactionsIfRequired(ctx context.Context) error {
	account, err := anh.proxy.GetAccount(ctx, anh.address)
	if err != nil {
		return err
	}

	anh.mut.Lock()
	if account.Nonce == anh.computedNonce {
		anh.lowestNonce = anh.computedNonce
		anh.transactions = make(map[uint64]*transaction.Transaction)
		anh.mut.Unlock()

		return nil
	}

	resendableTxs := make([]*transaction.Transaction, 0, len(anh.transactions))
	minNonce := anh.computedNonce
	for txNonce, tx := range anh.transactions {
		if txNonce <= account.Nonce {
			delete(anh.transactions, txNonce)
			continue
		}
		minNonce = tools.MinUint64(txNonce, minNonce)
		resendableTxs = append(resendableTxs, tx)
	}
	anh.lowestNonce = minNonce
	anh.mut.Unlock()

	if len(resendableTxs) == 0 {
		return nil
	}

	hashes, err := anh.proxy.SendTransactions(ctx, resendableTxs)
	if err != nil {
		return err
	}

	addressAsBech32String := anh.address.Bech32()
	log.Debug("resent transactions", "address", addressAsBech32String, "total txs", len(resendableTxs), "received hashes", len(hashes))

	return nil
}

// SendTransaction will save and propagate a transaction to the network
func (anh *addressNonceHandler) SendTransaction(ctx context.Context, tx *transaction.Transaction) (string, error) {
	anh.mut.Lock()
	anh.transactions[tx.GetRawData().GetNonce()] = tx
	anh.mut.Unlock()

	return anh.proxy.SendTransaction(ctx, tx)
}

func (anh *addressNonceHandler) getOlderTxWithSameNonce(tx *transaction.Transaction) *transaction.Transaction {
	anh.mut.RLock()
	defer anh.mut.RUnlock()

	return anh.transactions[tx.GetRawData().GetNonce()]
}

// IsInterfaceNil returns true if there is no value under the interface
func (anh *addressNonceHandler) IsInterfaceNil() bool {
	return anh == nil
}
