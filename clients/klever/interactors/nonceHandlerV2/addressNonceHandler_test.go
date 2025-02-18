package nonceHandlerV2

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"testing"

	"github.com/klever-io/klever-go/data/transaction"
	idata "github.com/klever-io/klever-go/indexer/data"
	chainModels "github.com/klever-io/klever-go/network/api/models"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/interactors"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
	testsCommon "github.com/klever-io/klv-bridge-eth-go/testsCommon/interactors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testAddressAsBech32String = "klv12e0kqcvqsrayj8j0c4dqjyvnv4ep253m5anx4rfj4jeq34lxsg8s84ec9j"
var testAddress, _ = address.NewAddress(testAddressAsBech32String)
var expectedErr = errors.New("expected error")

func TestAddressNonceHandler_NewAddressNonceHandlerWithPrivateAccess(t *testing.T) {
	t.Parallel()

	t.Run("nil proxy", func(t *testing.T) {
		t.Parallel()

		anh, err := NewAddressNonceHandlerWithPrivateAccess(nil, nil)
		assert.Nil(t, anh)
		assert.Equal(t, interactors.ErrNilProxy, err)
	})
	t.Run("nil addressHandler", func(t *testing.T) {
		t.Parallel()

		anh, err := NewAddressNonceHandlerWithPrivateAccess(&testsCommon.ProxyStub{}, nil)
		assert.Nil(t, anh)
		assert.Equal(t, interactors.ErrNilAddress, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		pubkey := make([]byte, 32)
		_, _ = rand.Read(pubkey)
		addressHandler, err := address.NewAddressFromBytes(pubkey)
		assert.Nil(t, err)

		_, err = NewAddressNonceHandlerWithPrivateAccess(&testsCommon.ProxyStub{}, addressHandler)
		assert.Nil(t, err)
	})
}

func TestAddressNonceHandler_ApplyNonceAndGasPrice(t *testing.T) {
	t.Parallel()
	t.Run("tx already sent; oldTx.BandwidthFee == txArgs.BandwidthFee", func(t *testing.T) {
		t.Parallel()

		tx := createDefaultTx()

		anh, err := NewAddressNonceHandlerWithPrivateAccess(&testsCommon.ProxyStub{}, testAddress)
		require.Nil(t, err)

		err = anh.ApplyNonceAndGasPrice(context.Background(), tx)
		require.Nil(t, err)

		_, err = anh.SendTransaction(context.Background(), tx)
		require.Nil(t, err)

		err = anh.ApplyNonceAndGasPrice(context.Background(), tx)
		require.Equal(t, interactors.ErrTxWithSameNonceAndGasPriceAlreadySent, err)
	})
	t.Run("tx already sent; oldTx.BandwidthFee < txArgs.BandwidthFee", func(t *testing.T) {
		t.Parallel()

		tx := createDefaultTx()

		estimateTransactionFeesTimesCalled := 0
		proxy := &testsCommon.ProxyStub{
			EstimateTransactionFeesCalled: func(_ context.Context, tx *transaction.Transaction) (*transaction.FeesResponse, error) {
				estimateTransactionFeesTimesCalled++
				fees := &transaction.FeesResponse{
					CostResponse: &transaction.CostResponse{
						BandwidthFee: 100 * int64(estimateTransactionFeesTimesCalled),
						KAppFee:      100,
					},
				}

				return fees, nil
			},
		}

		anh, err := NewAddressNonceHandlerWithPrivateAccess(proxy, testAddress)
		require.Nil(t, err)

		err = anh.ApplyNonceAndGasPrice(context.Background(), tx)
		require.Nil(t, err)

		_, err = anh.SendTransaction(context.Background(), tx)
		require.Nil(t, err)

		err = anh.ApplyNonceAndGasPrice(context.Background(), tx)
		require.Nil(t, err)

		_, err = anh.SendTransaction(context.Background(), tx)
		require.Nil(t, err)
	})
	t.Run("same transaction but with different nonce should work", func(t *testing.T) {
		t.Parallel()

		tx1 := createDefaultTx()
		tx2 := createDefaultTx()
		tx2.RawData.Nonce++

		anh, err := NewAddressNonceHandlerWithPrivateAccess(&testsCommon.ProxyStub{}, testAddress)
		require.Nil(t, err)

		err = anh.ApplyNonceAndGasPrice(context.Background(), tx1)
		require.Nil(t, err)

		_, err = anh.SendTransaction(context.Background(), tx1)
		require.Nil(t, err)

		err = anh.ApplyNonceAndGasPrice(context.Background(), tx2)
		require.Nil(t, err)

		_, err = anh.SendTransaction(context.Background(), tx2)
		require.Nil(t, err)
	})
}

func TestAddressNonceHandler_getNonceUpdatingCurrent(t *testing.T) {
	t.Parallel()

	t.Run("proxy returns error shall return error", func(t *testing.T) {
		t.Parallel()

		proxy := &testsCommon.ProxyStub{
			GetAccountCalled: func(_ context.Context, address address.Address) (*models.Account, error) {
				return nil, expectedErr
			},
		}

		anh, _ := NewAddressNonceHandlerWithPrivateAccess(proxy, testAddress)
		nonce, err := anh.getNonceUpdatingCurrent(context.Background())
		require.Equal(t, expectedErr, err)
		require.Equal(t, uint64(0), nonce)
	})
	t.Run("gap nonce detected", func(t *testing.T) {
		t.Parallel()

		blockchainNonce := uint64(100)
		proxy := &testsCommon.ProxyStub{
			GetAccountCalled: func(_ context.Context, address address.Address) (*models.Account, error) {
				return &models.Account{
					AccountInfo: &idata.AccountInfo{
						Nonce: blockchainNonce,
					},
				}, nil
			},
		}

		anh, _ := NewAddressNonceHandlerWithPrivateAccess(proxy, testAddress)
		anh.lowestNonce = blockchainNonce + 1

		nonce, err := anh.getNonceUpdatingCurrent(context.Background())
		require.Equal(t, interactors.ErrGapNonce, err)
		require.Equal(t, nonce, blockchainNonce)
	})
	t.Run("when computedNonce already set, getNonceUpdatingCurrent shall increase it", func(t *testing.T) {
		t.Parallel()

		blockchainNonce := uint64(100)
		proxy := &testsCommon.ProxyStub{
			GetAccountCalled: func(_ context.Context, address address.Address) (*models.Account, error) {
				return &models.Account{
					AccountInfo: &idata.AccountInfo{
						Nonce: blockchainNonce,
					},
				}, nil
			},
		}

		anh, _ := NewAddressNonceHandlerWithPrivateAccess(proxy, testAddress)
		anh.computedNonceWasSet = true
		computedNonce := uint64(105)
		anh.computedNonce = computedNonce

		nonce, err := anh.getNonceUpdatingCurrent(context.Background())
		require.Nil(t, err)
		require.Equal(t, nonce, computedNonce+1)
	})
	t.Run("getNonceUpdatingCurrent returns error should error", func(t *testing.T) {
		t.Parallel()

		proxy := &testsCommon.ProxyStub{
			GetAccountCalled: func(_ context.Context, address address.Address) (*models.Account, error) {
				return nil, expectedErr
			},
		}
		anh, _ := NewAddressNonceHandlerWithPrivateAccess(proxy, testAddress)
		tx := createDefaultTx()

		err := anh.ApplyNonceAndGasPrice(context.Background(), tx)
		require.Equal(t, expectedErr, err)
	})
}

func TestAddressNonceHandler_ReSendTransactionsIfRequired(t *testing.T) {
	t.Parallel()

	t.Run("proxy returns error shall error", func(t *testing.T) {
		t.Parallel()

		proxy := &testsCommon.ProxyStub{
			GetAccountCalled: func(_ context.Context, address address.Address) (*models.Account, error) {
				return nil, expectedErr
			},
		}

		anh, _ := NewAddressNonceHandlerWithPrivateAccess(proxy, testAddress)
		err := anh.ReSendTransactionsIfRequired(context.Background())
		require.Equal(t, expectedErr, err)
	})
	t.Run("proxy sendTransaction returns error shall error", func(t *testing.T) {
		t.Parallel()

		blockchainNonce := uint64(100)
		proxy := &testsCommon.ProxyStub{
			GetAccountCalled: func(_ context.Context, address address.Address) (*models.Account, error) {
				return &models.Account{
					AccountInfo: &idata.AccountInfo{
						Nonce: blockchainNonce - 1,
					},
				}, nil
			},
			SendTransactionsCalled: func(_ context.Context, txs []*transaction.Transaction) ([]string, error) {
				return make([]string, 0), expectedErr
			},
		}
		anh, _ := NewAddressNonceHandlerWithPrivateAccess(proxy, testAddress)
		tx := createDefaultTx()
		tx.RawData.Nonce = blockchainNonce
		_, err := anh.SendTransaction(context.Background(), tx)
		require.Nil(t, err)
		require.Equal(t, 1, len(anh.transactions))

		anh.computedNonce = blockchainNonce

		err = anh.ReSendTransactionsIfRequired(context.Background())
		require.Equal(t, 1, len(anh.transactions))
		require.Equal(t, expectedErr, err)
	})
	t.Run("account.Nonce == anh.computedNonce", func(t *testing.T) {
		t.Parallel()

		blockchainNonce := uint64(100)
		proxy := &testsCommon.ProxyStub{
			GetAccountCalled: func(_ context.Context, address address.Address) (*models.Account, error) {
				return &models.Account{
					AccountInfo: &idata.AccountInfo{
						Nonce: blockchainNonce,
					},
				}, nil
			},
		}
		anh, _ := NewAddressNonceHandlerWithPrivateAccess(proxy, testAddress)
		tx := createDefaultTx()
		_, err := anh.SendTransaction(context.Background(), tx)
		require.Nil(t, err)
		require.Equal(t, 1, len(anh.transactions))

		anh.computedNonce = blockchainNonce
		anh.lowestNonce = 80
		err = anh.ReSendTransactionsIfRequired(context.Background())
		require.Equal(t, anh.computedNonce, anh.lowestNonce)
		require.Equal(t, 0, len(anh.transactions))
		require.Nil(t, err)
	})
	t.Run("len(anh.transactions) == 0", func(t *testing.T) {
		t.Parallel()

		anh, _ := NewAddressNonceHandlerWithPrivateAccess(&testsCommon.ProxyStub{}, testAddress)
		tx := createDefaultTx()
		_, err := anh.SendTransaction(context.Background(), tx)
		require.Nil(t, err)
		require.Equal(t, 1, len(anh.transactions))

		anh.computedNonce = 100
		anh.lowestNonce = 80
		err = anh.ReSendTransactionsIfRequired(context.Background())
		require.Equal(t, anh.computedNonce, anh.lowestNonce)
		require.Equal(t, 0, len(anh.transactions))
		require.Nil(t, err)
	})
	t.Run("lowestNonce should be recalculated each time", func(t *testing.T) {
		t.Parallel()

		blockchainNonce := uint64(100)
		proxy := &testsCommon.ProxyStub{
			GetAccountCalled: func(_ context.Context, address address.Address) (*models.Account, error) {
				return &models.Account{
					AccountInfo: &idata.AccountInfo{
						Nonce: blockchainNonce - 1,
					},
				}, nil
			},
		}
		anh, _ := NewAddressNonceHandlerWithPrivateAccess(proxy, testAddress)
		tx := createDefaultTx()
		tx.RawData.Nonce = blockchainNonce + 1
		_, err := anh.SendTransaction(context.Background(), tx)
		require.Nil(t, err)
		require.Equal(t, 1, len(anh.transactions))

		anh.computedNonce = blockchainNonce + 2
		anh.lowestNonce = blockchainNonce
		err = anh.ReSendTransactionsIfRequired(context.Background())
		require.Equal(t, blockchainNonce+1, anh.lowestNonce)
		require.Equal(t, 1, len(anh.transactions))
		require.Nil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		blockchainNonce := uint64(100)
		proxy := &testsCommon.ProxyStub{
			GetAccountCalled: func(_ context.Context, address address.Address) (*models.Account, error) {
				return &models.Account{
					AccountInfo: &idata.AccountInfo{
						Nonce: blockchainNonce - 1,
					},
				}, nil
			},
			SendTransactionsCalled: func(_ context.Context, txs []*transaction.Transaction) ([]string, error) {
				return make([]string, 0), nil
			},
		}
		anh, _ := NewAddressNonceHandlerWithPrivateAccess(proxy, testAddress)
		tx := createDefaultTx()
		tx.RawData.Nonce = blockchainNonce
		_, err := anh.SendTransaction(context.Background(), tx)
		require.Nil(t, err)
		require.Equal(t, 1, len(anh.transactions))

		anh.computedNonce = blockchainNonce

		err = anh.ReSendTransactionsIfRequired(context.Background())
		require.Equal(t, 1, len(anh.transactions))
		require.Nil(t, err)
	})
}

func createDefaultTx() *transaction.Transaction {
	tx := transaction.NewBaseTransaction(testAddress.Bytes(), 0, nil, 0, 0)
	contractRequest := chainModels.TransferTXRequest{
		Receiver: testAddressAsBech32String,
		Amount:   10,
		KDA:      "KLV",
	}

	contractBytes, _ := json.Marshal(contractRequest)

	txArgs := transaction.TXArgs{
		Type:     uint32(transaction.TXContract_TransferContractType),
		Sender:   testAddress.Bytes(),
		Contract: json.RawMessage(contractBytes),
	}

	tx.AddTransaction(txArgs)

	return tx
}
