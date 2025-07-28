package nonceHandlerV2

import (
	"context"
	"crypto/rand"
	"errors"
	"sync"
	"testing"

	"github.com/klever-io/klever-go/data/transaction"
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

		tx := createDefaultTx(t)

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

		tx := createDefaultTx(t)

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

		tx1 := createDefaultTx(t)
		tx2 := createDefaultTx(t)
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
					Nonce: blockchainNonce,
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
					Nonce: blockchainNonce,
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
		tx := createDefaultTx(t)

		err := anh.ApplyNonceAndGasPrice(context.Background(), tx)
		require.Equal(t, expectedErr, err)
	})
	t.Run("getNonceUpdatingCurrent computed nonce concurrent usage should work", func(t *testing.T) {
		t.Parallel()

		proxy := &testsCommon.ProxyStub{}
		anh, _ := NewAddressNonceHandlerWithPrivateAccess(proxy, testAddress)

		numberOfWorkers := 10
		var wg sync.WaitGroup
		var mu sync.Mutex

		nonces := make(map[uint64]bool)
		for i := 0; i < numberOfWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				tx := createDefaultTx(t)
				err := anh.ApplyNonceAndGasPrice(context.Background(), tx)
				assert.Nil(t, err)

				mu.Lock()
				nonces[tx.RawData.Nonce] = true
				mu.Unlock()
			}()
		}
		wg.Wait()

		require.Equal(t, numberOfWorkers, len(nonces))
		for i := 0; i < numberOfWorkers; i++ {
			_, exists := nonces[uint64(i)]
			require.True(t, exists)
		}
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
					Nonce: blockchainNonce - 1,
				}, nil
			},
			SendTransactionsCalled: func(_ context.Context, txs []*transaction.Transaction) ([]string, error) {
				return make([]string, 0), expectedErr
			},
		}
		anh, _ := NewAddressNonceHandlerWithPrivateAccess(proxy, testAddress)
		tx := createDefaultTx(t)
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
					Nonce: blockchainNonce,
				}, nil
			},
		}
		anh, _ := NewAddressNonceHandlerWithPrivateAccess(proxy, testAddress)
		tx := createDefaultTx(t)
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
		tx := createDefaultTx(t)
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
					Nonce: blockchainNonce - 1,
				}, nil
			},
		}
		anh, _ := NewAddressNonceHandlerWithPrivateAccess(proxy, testAddress)
		tx := createDefaultTx(t)
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
					Nonce: blockchainNonce - 1,
				}, nil
			},
			SendTransactionsCalled: func(_ context.Context, txs []*transaction.Transaction) ([]string, error) {
				return make([]string, 0), nil
			},
		}
		anh, _ := NewAddressNonceHandlerWithPrivateAccess(proxy, testAddress)
		tx := createDefaultTx(t)
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

func createDefaultTx(t testing.TB) *transaction.Transaction {
	tx := transaction.NewBaseTransaction(testAddress.Bytes(), 0, nil, 0, 0)
	contractRequest := &transaction.TransferContract{
		ToAddress: testAddress.Bytes(),
		AssetID:   []byte("KLV"),
		Amount:    10,
	}

	err := tx.PushContract(transaction.TXContract_SmartContractType, contractRequest)
	require.Nil(t, err)

	return tx
}
