package balanceValidator

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klv-bridge-eth-go/clients"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/klever-io/klv-bridge-eth-go/core/batchProcessor"
	"github.com/klever-io/klv-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

var (
	ethToken = common.BytesToAddress([]byte("eth token"))
	kdaToken = []byte("kda token")
	amount   = big.NewInt(100)
	amount2  = big.NewInt(200)
)

func createMockArgsBalanceValidator() ArgsBalanceValidator {
	return ArgsBalanceValidator{
		Log:            &testscommon.LoggerStub{},
		KcClient:       &bridge.KcClientStub{},
		EthereumClient: &bridge.EthereumClientStub{},
	}
}

type testConfiguration struct {
	isNativeOnEth      bool
	isMintBurnOnEth    bool
	totalBalancesOnEth *big.Int
	burnBalancesOnEth  *big.Int
	mintBalancesOnEth  *big.Int

	isNativeOnKlv      bool
	isMintBurnOnKlv    bool
	totalBalancesOnKlv *big.Int
	burnBalancesOnKlv  *big.Int
	mintBalancesOnKlv  *big.Int

	errorsOnCalls map[string]error

	ethToken  common.Address
	kdaToken  []byte
	amount    *big.Int
	direction batchProcessor.Direction

	lastExecutedEthBatch       uint64
	pendingKlvBatchId          uint64
	amountsOnKlvPendingBatches map[uint64][]*big.Int
	amountsOnEthPendingBatches map[uint64][]*big.Int
}

func (cfg *testConfiguration) deepClone() testConfiguration {
	result := testConfiguration{
		isNativeOnEth:              cfg.isNativeOnEth,
		isMintBurnOnEth:            cfg.isMintBurnOnEth,
		isNativeOnKlv:              cfg.isNativeOnKlv,
		isMintBurnOnKlv:            cfg.isMintBurnOnKlv,
		errorsOnCalls:              make(map[string]error),
		ethToken:                   common.HexToAddress(cfg.ethToken.Hex()),
		kdaToken:                   make([]byte, len(cfg.kdaToken)),
		direction:                  cfg.direction,
		lastExecutedEthBatch:       cfg.lastExecutedEthBatch,
		pendingKlvBatchId:          cfg.pendingKlvBatchId,
		amountsOnKlvPendingBatches: make(map[uint64][]*big.Int),
		amountsOnEthPendingBatches: make(map[uint64][]*big.Int),
	}
	if cfg.totalBalancesOnEth != nil {
		result.totalBalancesOnEth = big.NewInt(0).Set(cfg.totalBalancesOnEth)
	}
	if cfg.burnBalancesOnEth != nil {
		result.burnBalancesOnEth = big.NewInt(0).Set(cfg.burnBalancesOnEth)
	}
	if cfg.mintBalancesOnEth != nil {
		result.mintBalancesOnEth = big.NewInt(0).Set(cfg.mintBalancesOnEth)
	}
	if cfg.totalBalancesOnKlv != nil {
		result.totalBalancesOnKlv = big.NewInt(0).Set(cfg.totalBalancesOnKlv)
	}
	if cfg.burnBalancesOnKlv != nil {
		result.burnBalancesOnKlv = big.NewInt(0).Set(cfg.burnBalancesOnKlv)
	}
	if cfg.mintBalancesOnKlv != nil {
		result.mintBalancesOnKlv = big.NewInt(0).Set(cfg.mintBalancesOnKlv)
	}
	if cfg.amount != nil {
		result.amount = big.NewInt(0).Set(cfg.amount)
	}

	for key, err := range cfg.errorsOnCalls {
		result.errorsOnCalls[key] = err
	}
	copy(result.kdaToken, cfg.kdaToken)
	for nonce, values := range cfg.amountsOnKlvPendingBatches {
		result.amountsOnKlvPendingBatches[nonce] = make([]*big.Int, 0, len(values))
		for _, value := range values {
			result.amountsOnKlvPendingBatches[nonce] = append(result.amountsOnKlvPendingBatches[nonce], big.NewInt(0).Set(value))
		}
	}
	for nonce, values := range cfg.amountsOnEthPendingBatches {
		result.amountsOnEthPendingBatches[nonce] = make([]*big.Int, 0, len(values))
		for _, value := range values {
			result.amountsOnEthPendingBatches[nonce] = append(result.amountsOnEthPendingBatches[nonce], big.NewInt(0).Set(value))
		}
	}

	return result
}

type testResult struct {
	checkRequiredBalanceOnEthCalled bool
	checkRequiredBalanceOnKlvCalled bool
	error                           error
}

func TestNewBalanceValidator(t *testing.T) {
	t.Parallel()

	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBalanceValidator()
		args.Log = nil
		instance, err := NewBalanceValidator(args)
		assert.Nil(t, instance)
		assert.Equal(t, ErrNilLogger, err)
	})
	t.Run("nil Klever Blockchain client should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBalanceValidator()
		args.KcClient = nil
		instance, err := NewBalanceValidator(args)
		assert.Nil(t, instance)
		assert.Equal(t, ErrNilKcClient, err)
	})
	t.Run("nil Ethereum client should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBalanceValidator()
		args.EthereumClient = nil
		instance, err := NewBalanceValidator(args)
		assert.Nil(t, instance)
		assert.Equal(t, ErrNilEthereumClient, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBalanceValidator()
		instance, err := NewBalanceValidator(args)
		assert.NotNil(t, instance)
		assert.Nil(t, err)
	})
}

func TestBalanceValidator_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *balanceValidator
	assert.True(t, instance.IsInterfaceNil())

	instance = &balanceValidator{}
	assert.False(t, instance.IsInterfaceNil())
}

func TestBridgeExecutor_CheckToken(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	t.Run("unknown direction should error", func(t *testing.T) {
		t.Parallel()

		cfg := testConfiguration{
			direction: "",
		}
		result := validatorTester(cfg)
		assert.ErrorIs(t, result.error, ErrInvalidDirection)
	})
	t.Run("query operations error", func(t *testing.T) {
		t.Parallel()

		t.Run("on isMintBurnOnEthereum", func(t *testing.T) {
			cfg := testConfiguration{
				direction: batchProcessor.FromKc,
				errorsOnCalls: map[string]error{
					"MintBurnTokensEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on isMintBurnOnKc", func(t *testing.T) {
			cfg := testConfiguration{
				direction: batchProcessor.ToKc,
				errorsOnCalls: map[string]error{
					"IsMintBurnTokenKlv": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on isNativeOnEthereum", func(t *testing.T) {
			cfg := testConfiguration{
				direction: batchProcessor.ToKc,
				errorsOnCalls: map[string]error{
					"NativeTokensEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on isNativeOnKc", func(t *testing.T) {
			cfg := testConfiguration{
				direction: batchProcessor.FromKc,
				errorsOnCalls: map[string]error{
					"IsNativeTokenKlv": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeEthAmount, TotalBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.FromKc,
				isMintBurnOnKlv: true,
				isNativeOnEth:   true,
				errorsOnCalls: map[string]error{
					"TotalBalancesEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeEthAmount, BurnBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.FromKc,
				isNativeOnKlv:   true,
				isMintBurnOnEth: true,
				errorsOnCalls: map[string]error{
					"BurnBalancesEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeEthAmount, MintBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.FromKc,
				isNativeOnKlv:   true,
				isMintBurnOnEth: true,
				errorsOnCalls: map[string]error{
					"MintBalancesEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeEthAmount, GetLastExecutedEthBatchID", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.FromKc,
				isNativeOnKlv:   true,
				isMintBurnOnEth: true,
				errorsOnCalls: map[string]error{
					"GetLastExecutedEthBatchIDKlv": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeEthAmount, GetBatch", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.FromKc,
				isNativeOnKlv:   true,
				isMintBurnOnEth: true,
				errorsOnCalls: map[string]error{
					"GetBatchEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.True(t, result.checkRequiredBalanceOnEthCalled)
			assert.False(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeKlvAmount, TotalBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToKc,
				isNativeOnKlv:   true,
				isMintBurnOnEth: true,
				errorsOnCalls: map[string]error{
					"TotalBalancesKlv": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeKlvAmount, BurnBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToKc,
				isMintBurnOnKlv: true,
				isNativeOnEth:   true,
				errorsOnCalls: map[string]error{
					"BurnBalancesKlv": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeKlvAmount, MintBalances", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToKc,
				isMintBurnOnKlv: true,
				isNativeOnEth:   true,
				errorsOnCalls: map[string]error{
					"MintBalancesKlv": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeKlvAmount, GetLastKlvBatchID", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToKc,
				isMintBurnOnKlv: true,
				isNativeOnEth:   true,
				errorsOnCalls: map[string]error{
					"GetLastKlvBatchID": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeKlvAmount, GetBatch", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToKc,
				isMintBurnOnKlv: true,
				isNativeOnEth:   true,
				errorsOnCalls: map[string]error{
					"GetBatchKlv": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on computeKlvAmount, WasExecuted", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToKc,
				isMintBurnOnKlv: true,
				isNativeOnEth:   true,
				errorsOnCalls: map[string]error{
					"WasExecutedEth": expectedError,
				},
			}
			result := validatorTester(cfg)
			assert.Equal(t, expectedError, result.error)
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
	})
	t.Run("invalid setup", func(t *testing.T) {
		t.Parallel()

		t.Run("on Ethereum is not native nor is mint/burn, should error", func(t *testing.T) {
			cfg := testConfiguration{
				direction:       batchProcessor.ToKc,
				isMintBurnOnKlv: true,
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrInvalidSetup)
			assert.Contains(t, result.error.Error(), "isNativeOnEthereum = false, isMintBurnOnEthereum = false")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on Klever Blockchain is not native nor is mint/burn, should error", func(t *testing.T) {
			cfg := testConfiguration{
				direction:     batchProcessor.ToKc,
				isNativeOnEth: true,
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrInvalidSetup)
			assert.Contains(t, result.error.Error(), "isNativeOnKc = false, isMintBurnOnKc = false")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("native on both chains, should error", func(t *testing.T) {
			cfg := testConfiguration{
				direction:     batchProcessor.ToKc,
				isNativeOnEth: true,
				isNativeOnKlv: true,
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrInvalidSetup)
			assert.Contains(t, result.error.Error(), "isNativeOnEthereum = true, isNativeOnKc = true")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
	})
	t.Run("bad values on mint & burn", func(t *testing.T) {
		t.Parallel()

		t.Run("on Ethereum, native", func(t *testing.T) {
			t.Parallel()

			cfg := testConfiguration{
				direction:         batchProcessor.ToKc,
				isMintBurnOnEth:   true,
				isNativeOnEth:     true,
				isMintBurnOnKlv:   true,
				burnBalancesOnEth: big.NewInt(37),
				mintBalancesOnEth: big.NewInt(38),
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrNegativeAmount)
			assert.Contains(t, result.error.Error(), "ethAmount: -1")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on Ethereum, non-native", func(t *testing.T) {
			t.Parallel()

			cfg := testConfiguration{
				direction:         batchProcessor.ToKc,
				isMintBurnOnEth:   true,
				isNativeOnKlv:     true,
				burnBalancesOnEth: big.NewInt(38),
				mintBalancesOnEth: big.NewInt(37),
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrNegativeAmount)
			assert.Contains(t, result.error.Error(), "ethAmount: -1")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on Kc, native", func(t *testing.T) {
			t.Parallel()

			cfg := testConfiguration{
				direction:         batchProcessor.ToKc,
				isMintBurnOnEth:   true,
				isMintBurnOnKlv:   true,
				isNativeOnKlv:     true,
				burnBalancesOnKlv: big.NewInt(37),
				mintBalancesOnKlv: big.NewInt(38),
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrNegativeAmount)
			assert.Contains(t, result.error.Error(), "kdaAmount: -1")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
		t.Run("on Kc, non-native", func(t *testing.T) {
			t.Parallel()

			cfg := testConfiguration{
				direction:         batchProcessor.ToKc,
				isNativeOnEth:     true,
				isMintBurnOnKlv:   true,
				burnBalancesOnKlv: big.NewInt(38),
				mintBalancesOnKlv: big.NewInt(37),
			}
			result := validatorTester(cfg)
			assert.ErrorIs(t, result.error, ErrNegativeAmount)
			assert.Contains(t, result.error.Error(), "kdaAmount: -1")
			assert.False(t, result.checkRequiredBalanceOnEthCalled)
			assert.True(t, result.checkRequiredBalanceOnKlvCalled)
		})
	})
	t.Run("scenarios", func(t *testing.T) {
		t.Parallel()

		t.Run("Eth -> Klv", func(t *testing.T) {
			t.Parallel()

			t.Run("native on Klv, mint-burn on Eth, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.ToKc,
					isMintBurnOnEth:    true,
					isNativeOnKlv:      true,
					burnBalancesOnEth:  big.NewInt(1100),  // initial burn (1000) + burn from this transfer (100)
					mintBalancesOnEth:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnKlv: big.NewInt(10000),
					amount:             amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Klv, mint-burn on Eth, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.ToKc,
					isMintBurnOnEth:    true,
					isNativeOnKlv:      true,
					burnBalancesOnEth:  big.NewInt(1220),  // initial burn (1000) + burn from this transfer (100) + burn from next batches (120)
					mintBalancesOnEth:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnKlv: big.NewInt(10000),
					amount:             amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Klv but with mint-burn, mint-burn on Eth, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.ToKc,
					isMintBurnOnEth:   true,
					isNativeOnKlv:     true,
					isMintBurnOnKlv:   true,
					burnBalancesOnEth: big.NewInt(1100),  // initial burn (1000) + burn from this transfer (100)
					mintBalancesOnEth: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnKlv: big.NewInt(12000),
					mintBalancesOnKlv: big.NewInt(2000), // burn - mint on Klv === mint - burn on Eth
					amount:            amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Klv but with mint-burn, mint-burn on Eth, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.ToKc,
					isMintBurnOnEth:   true,
					isNativeOnKlv:     true,
					isMintBurnOnKlv:   true,
					burnBalancesOnEth: big.NewInt(1220),  // initial burn (1000) + burn from this transfer (100) + next batches (120)
					mintBalancesOnEth: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnKlv: big.NewInt(12000),
					mintBalancesOnKlv: big.NewInt(2000), // burn - mint on Klv === mint - burn on Eth
					amount:            amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Eth, mint-burn on Klv, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.ToKc,
					isMintBurnOnKlv:    true,
					isNativeOnEth:      true,
					burnBalancesOnKlv:  big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnKlv:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnEth: big.NewInt(10100), // initial (10000) + locked from this transfer (100)
					amount:             amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnKlv.Add(cfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Eth, mint-burn on Klv, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.ToKc,
					isMintBurnOnKlv:    true,
					isNativeOnEth:      true,
					burnBalancesOnKlv:  big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnKlv:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnEth: big.NewInt(10220), // initial (10000) + locked from this transfer (100) + next batches (120)
					amount:             amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnKlv.Add(cfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Eth but with mint-burn, mint-burn on Klv, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.ToKc,
					isMintBurnOnKlv:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnKlv: big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnKlv: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnEth: big.NewInt(12100),
					mintBalancesOnEth: big.NewInt(2000), // burn - mint - transfer on Eth === mint - burn on Klv
					amount:            amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnKlv.Add(cfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Eth but with mint-burn, mint-burn on Klv, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.ToKc,
					isMintBurnOnKlv:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnKlv: big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnKlv: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnEth: big.NewInt(12220),
					mintBalancesOnEth: big.NewInt(2000), // burn - mint - transfer on Eth - next transfers === mint - burn on Klv
					amount:            amount,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnKlv.Add(cfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
		})

		t.Run("Klv -> Eth", func(t *testing.T) {
			t.Parallel()

			t.Run("native on Klv, mint-burn on Eth, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.FromKc,
					isMintBurnOnEth:    true,
					isNativeOnKlv:      true,
					burnBalancesOnEth:  big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnEth:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnKlv: big.NewInt(10100), // initial (10000) + transfer from this batch (100)
					amount:             amount,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Klv, mint-burn on Eth, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.FromKc,
					isMintBurnOnEth:    true,
					isNativeOnKlv:      true,
					burnBalancesOnEth:  big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnEth:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnKlv: big.NewInt(10220), // initial (10000) + transfer from this batch (100) + next batches (120)
					amount:             amount,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Klv but with mint-burn, mint-burn on Eth, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.FromKc,
					isMintBurnOnEth:   true,
					isNativeOnKlv:     true,
					isMintBurnOnKlv:   true,
					burnBalancesOnEth: big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnEth: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnKlv: big.NewInt(12100),
					mintBalancesOnKlv: big.NewInt(2000), // burn - mint - transfer on Klv === mint - burn on Eth
					amount:            amount,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Klv but with mint-burn, mint-burn on Eth, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.FromKc,
					isMintBurnOnEth:   true,
					isNativeOnKlv:     true,
					isMintBurnOnKlv:   true,
					burnBalancesOnEth: big.NewInt(1000),  // initial burn (1000)
					mintBalancesOnEth: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnKlv: big.NewInt(12220),
					mintBalancesOnKlv: big.NewInt(2000), // burn - mint - transfer - next batches on Klv === mint - burn on Eth
					amount:            amount,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnEth.Add(cfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Eth, mint-burn on Klv, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.FromKc,
					isMintBurnOnKlv:    true,
					isNativeOnEth:      true,
					burnBalancesOnKlv:  big.NewInt(1100),  // initial burn (1000) + transfer from this batch (100)
					mintBalancesOnKlv:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnEth: big.NewInt(10000), // initial (10000)
					amount:             amount,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnKlv.Add(cfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Eth, mint-burn on Klv, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:          batchProcessor.FromKc,
					isMintBurnOnKlv:    true,
					isNativeOnEth:      true,
					burnBalancesOnKlv:  big.NewInt(1220),  // initial burn (1000) + transfer from this batch (100) + next batches (120)
					mintBalancesOnKlv:  big.NewInt(11000), // minted (10000) + initial burn (1000)
					totalBalancesOnEth: big.NewInt(10000), // initial (10000)
					amount:             amount,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnKlv.Add(cfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Eth but with mint-burn, mint-burn on Klv, ok values, no next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.FromKc,
					isMintBurnOnKlv:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnKlv: big.NewInt(1100),  // initial burn (1000) + transfer from this batch (100)
					mintBalancesOnKlv: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnEth: big.NewInt(12000),
					mintBalancesOnEth: big.NewInt(2000), // burn - mint on Eth === mint - burn - transfer on Klv
					amount:            amount,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnKlv.Add(cfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("native on Eth but with mint-burn, mint-burn on Klv, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				cfg := testConfiguration{
					direction:         batchProcessor.FromKc,
					isMintBurnOnKlv:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnKlv: big.NewInt(1220),  // initial burn (1000) + transfer from this batch (100) + transfer from next batches
					mintBalancesOnKlv: big.NewInt(11000), // minted (10000) + initial burn (1000)
					burnBalancesOnEth: big.NewInt(12000),
					mintBalancesOnEth: big.NewInt(2000), // burn - mint on Eth === mint - burn - transfer - next batches on Klv
					amount:            amount,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					cfg.burnBalancesOnKlv.Add(cfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(cfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
		})

		t.Run("Klv <-> Eth", func(t *testing.T) {
			t.Parallel()

			t.Run("from Eth: native on Klv, mint-burn on Eth, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				existingNativeBalanceKlv := int64(10000)
				existingBurnEth := int64(150000)
				existingMintEth := int64(160000)

				cfg := testConfiguration{
					direction:          batchProcessor.ToKc,
					isMintBurnOnEth:    true,
					isNativeOnKlv:      true,
					burnBalancesOnEth:  big.NewInt(existingBurnEth + 100 + 30 + 40 + 50),
					mintBalancesOnEth:  big.NewInt(existingMintEth),
					totalBalancesOnKlv: big.NewInt(existingNativeBalanceKlv + 60 + 80 + 100 + 200),
					amount:             amount,
					pendingKlvBatchId:  1,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					copiedCfg := cfg.deepClone()
					copiedCfg.burnBalancesOnEth.Add(copiedCfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(copiedCfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from Eth: native on Klv but with mint-burn, mint-burn on Eth, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				existingBurnKlv := int64(370000) // burn > mint because the token is native
				existingMintKlv := int64(360000)
				existingBurnEth := int64(150000)
				existingMintEth := int64(160000)

				cfg := testConfiguration{
					direction:         batchProcessor.ToKc,
					isMintBurnOnEth:   true,
					isNativeOnKlv:     true,
					isMintBurnOnKlv:   true,
					burnBalancesOnEth: big.NewInt(existingBurnEth + 100 + 30 + 40 + 50),
					mintBalancesOnEth: big.NewInt(existingMintEth),
					burnBalancesOnKlv: big.NewInt(existingBurnKlv + 60 + 80 + 100 + 200),
					mintBalancesOnKlv: big.NewInt(existingMintKlv),
					amount:            amount,
					pendingKlvBatchId: 1,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					copiedCfg := cfg.deepClone()
					copiedCfg.burnBalancesOnEth.Add(copiedCfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(copiedCfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from Eth: native on Eth, mint-burn on Klv, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				existingBurnKlv := int64(360000)
				existingMintKlv := int64(370000)
				existingNativeBalanceEth := int64(10000)

				cfg := testConfiguration{
					direction:          batchProcessor.ToKc,
					isMintBurnOnKlv:    true,
					isNativeOnEth:      true,
					burnBalancesOnKlv:  big.NewInt(existingBurnKlv + 200 + 60 + 80 + 100),
					mintBalancesOnKlv:  big.NewInt(existingMintKlv),
					totalBalancesOnEth: big.NewInt(existingNativeBalanceEth + 100 + 30 + 40 + 50),
					amount:             amount,
					pendingKlvBatchId:  1,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					copiedCfg := cfg.deepClone()
					copiedCfg.burnBalancesOnKlv.Add(copiedCfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(copiedCfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from Eth: native on Eth but with mint-burn, mint-burn on Klv, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				existingBurnKlv := int64(360000)
				existingMintKlv := int64(370000)
				existingBurnEth := int64(160000) // burn > mint because the token is native
				existingMintEth := int64(150000)

				cfg := testConfiguration{
					direction:         batchProcessor.ToKc,
					isMintBurnOnKlv:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnKlv: big.NewInt(existingBurnKlv + 200 + 60 + 80 + 100),
					mintBalancesOnKlv: big.NewInt(existingMintKlv),
					burnBalancesOnEth: big.NewInt(existingBurnEth + 100 + 30 + 40 + 50),
					mintBalancesOnEth: big.NewInt(existingMintEth),
					amount:            amount,
					pendingKlvBatchId: 1,
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.False(t, result.checkRequiredBalanceOnEthCalled)
				assert.True(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					copiedCfg := cfg.deepClone()
					copiedCfg.burnBalancesOnKlv.Add(copiedCfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(copiedCfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from Klv: native on Klv, mint-burn on Eth, ok values, with next pending batches on both chains", func(t *testing.T) {
				t.Parallel()

				existingNativeBalanceKlv := int64(10000)
				existingBurnEth := int64(150000)
				existingMintEth := int64(160000)

				cfg := testConfiguration{
					direction:          batchProcessor.FromKc,
					isMintBurnOnEth:    true,
					isNativeOnKlv:      true,
					burnBalancesOnEth:  big.NewInt(existingBurnEth + 200 + 60 + 80 + 100),
					mintBalancesOnEth:  big.NewInt(existingMintEth),
					totalBalancesOnKlv: big.NewInt(existingNativeBalanceKlv + 30 + 40 + 50 + 100),
					amount:             amount,
					pendingKlvBatchId:  1,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					copiedCfg := cfg.deepClone()
					copiedCfg.burnBalancesOnEth.Add(copiedCfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(copiedCfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from Klv: native on Klv but with mint-burn, mint-burn on Eth, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				existingBurnKlv := int64(370000) // burn > mint because the token is native
				existingMintKlv := int64(360000)
				existingBurnEth := int64(150000)
				existingMintEth := int64(160000)

				cfg := testConfiguration{
					direction:         batchProcessor.FromKc,
					isMintBurnOnEth:   true,
					isNativeOnKlv:     true,
					isMintBurnOnKlv:   true,
					burnBalancesOnEth: big.NewInt(existingBurnEth + 200 + 60 + 80 + 100),
					mintBalancesOnEth: big.NewInt(existingMintEth),
					burnBalancesOnKlv: big.NewInt(existingBurnKlv + 30 + 40 + 50 + 100),
					mintBalancesOnKlv: big.NewInt(existingMintKlv),
					amount:            amount,
					pendingKlvBatchId: 1,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					copiedCfg := cfg.deepClone()
					copiedCfg.burnBalancesOnEth.Add(copiedCfg.burnBalancesOnEth, big.NewInt(1))
					result = validatorTester(copiedCfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from Klv: native on Eth, mint-burn on Klv, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				existingBurnKlv := int64(360000)
				existingMintKlv := int64(370000)
				existingNativeBalanceEth := int64(10000)

				cfg := testConfiguration{
					direction:          batchProcessor.FromKc,
					isMintBurnOnKlv:    true,
					isNativeOnEth:      true,
					burnBalancesOnKlv:  big.NewInt(existingBurnKlv + 100 + 30 + 40 + 50),
					mintBalancesOnKlv:  big.NewInt(existingMintKlv),
					totalBalancesOnEth: big.NewInt(existingNativeBalanceEth + 200 + 60 + 80 + 100),
					amount:             amount,
					pendingKlvBatchId:  1,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					copiedCfg := cfg.deepClone()
					copiedCfg.burnBalancesOnKlv.Add(copiedCfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(copiedCfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
			t.Run("from Klv: native on Eth but with mint-burn, mint-burn on Klv, ok values, with next pending batches", func(t *testing.T) {
				t.Parallel()

				existingBurnKlv := int64(360000)
				existingMintKlv := int64(370000)
				existingBurnEth := int64(160000) // burn > mint because the token is native
				existingMintEth := int64(150000)

				cfg := testConfiguration{
					direction:         batchProcessor.FromKc,
					isMintBurnOnKlv:   true,
					isNativeOnEth:     true,
					isMintBurnOnEth:   true,
					burnBalancesOnKlv: big.NewInt(existingBurnKlv + 100 + 30 + 40 + 50),
					mintBalancesOnKlv: big.NewInt(existingMintKlv),
					burnBalancesOnEth: big.NewInt(existingBurnEth + 200 + 60 + 80 + 100),
					mintBalancesOnEth: big.NewInt(existingMintEth),
					amount:            amount,
					pendingKlvBatchId: 1,
					amountsOnKlvPendingBatches: map[uint64][]*big.Int{
						1: {amount},
						2: {big.NewInt(30), big.NewInt(40)},
						3: {big.NewInt(50)},
					},
					amountsOnEthPendingBatches: map[uint64][]*big.Int{
						1: {amount2},
						2: {big.NewInt(60), big.NewInt(80)},
						3: {big.NewInt(100)},
					},
					kdaToken: kdaToken,
					ethToken: ethToken,
				}

				result := validatorTester(cfg)
				assert.Nil(t, result.error)
				assert.True(t, result.checkRequiredBalanceOnEthCalled)
				assert.False(t, result.checkRequiredBalanceOnKlvCalled)

				t.Run("mismatch should error", func(t *testing.T) {
					copiedCfg := cfg.deepClone()
					copiedCfg.burnBalancesOnKlv.Add(copiedCfg.burnBalancesOnKlv, big.NewInt(1))
					result = validatorTester(copiedCfg)
					assert.ErrorIs(t, result.error, ErrBalanceMismatch)
				})
			})
		})
	})
}

func validatorTester(cfg testConfiguration) testResult {
	args := createMockArgsBalanceValidator()

	result := testResult{}

	lastKlvBatchID := uint64(0)
	for key := range cfg.amountsOnKlvPendingBatches {
		if key > lastKlvBatchID {
			lastKlvBatchID = key
		}
	}

	args.KcClient = &bridge.KcClientStub{
		CheckRequiredBalanceCalled: func(ctx context.Context, token []byte, value *big.Int) error {
			result.checkRequiredBalanceOnKlvCalled = true
			return nil
		},
		IsMintBurnTokenCalled: func(ctx context.Context, token []byte) (bool, error) {
			err := cfg.errorsOnCalls["IsMintBurnTokenKlv"]
			if err != nil {
				return false, err
			}

			return cfg.isMintBurnOnKlv, nil
		},
		IsNativeTokenCalled: func(ctx context.Context, token []byte) (bool, error) {
			err := cfg.errorsOnCalls["IsNativeTokenKlv"]
			if err != nil {
				return false, err
			}

			return cfg.isNativeOnKlv, nil
		},
		TotalBalancesCalled: func(ctx context.Context, token []byte) (*big.Int, error) {
			err := cfg.errorsOnCalls["TotalBalancesKlv"]
			if err != nil {
				return nil, err
			}

			return returnBigIntOrZeroIfNil(cfg.totalBalancesOnKlv), nil
		},
		MintBalancesCalled: func(ctx context.Context, token []byte) (*big.Int, error) {
			err := cfg.errorsOnCalls["MintBalancesKlv"]
			if err != nil {
				return nil, err
			}

			return returnBigIntOrZeroIfNil(cfg.mintBalancesOnKlv), nil
		},
		BurnBalancesCalled: func(ctx context.Context, token []byte) (*big.Int, error) {
			err := cfg.errorsOnCalls["BurnBalancesKlv"]
			if err != nil {
				return nil, err
			}

			return returnBigIntOrZeroIfNil(cfg.burnBalancesOnKlv), nil
		},
		GetPendingBatchCalled: func(ctx context.Context) (*bridgeCore.TransferBatch, error) {
			err := cfg.errorsOnCalls["GetPendingBatchKlv"]
			if err != nil {
				return nil, err
			}

			batch := &bridgeCore.TransferBatch{
				ID: cfg.pendingKlvBatchId,
			}
			applyDummyFromKlvDepositsToBatch(cfg, batch)

			return batch, nil
		},
		GetBatchCalled: func(ctx context.Context, batchID uint64) (*bridgeCore.TransferBatch, error) {
			err := cfg.errorsOnCalls["GetBatchKlv"]
			if err != nil {
				return nil, err
			}

			if batchID > getMaxKlvPendingBatchID(cfg) {
				return nil, clients.ErrNoBatchAvailable
			}
			batch := &bridgeCore.TransferBatch{
				ID: batchID,
			}
			applyDummyFromKlvDepositsToBatch(cfg, batch)

			return batch, nil
		},
		GetLastExecutedEthBatchIDCalled: func(ctx context.Context) (uint64, error) {
			err := cfg.errorsOnCalls["GetLastExecutedEthBatchIDKlv"]
			if err != nil {
				return 0, err
			}

			return cfg.lastExecutedEthBatch, nil
		},
		GetLastKlvBatchIDCalled: func(ctx context.Context) (uint64, error) {
			err := cfg.errorsOnCalls["GetLastKlvBatchID"]
			if err != nil {
				return 0, err
			}

			return lastKlvBatchID, nil
		},
	}
	args.EthereumClient = &bridge.EthereumClientStub{
		CheckRequiredBalanceCalled: func(ctx context.Context, erc20Address common.Address, value *big.Int) error {
			result.checkRequiredBalanceOnEthCalled = true
			return nil
		},
		MintBurnTokensCalled: func(ctx context.Context, account common.Address) (bool, error) {
			err := cfg.errorsOnCalls["MintBurnTokensEth"]
			if err != nil {
				return false, err
			}

			return cfg.isMintBurnOnEth, nil
		},
		NativeTokensCalled: func(ctx context.Context, account common.Address) (bool, error) {
			err := cfg.errorsOnCalls["NativeTokensEth"]
			if err != nil {
				return false, err
			}

			return cfg.isNativeOnEth, nil
		},
		TotalBalancesCalled: func(ctx context.Context, account common.Address) (*big.Int, error) {
			err := cfg.errorsOnCalls["TotalBalancesEth"]
			if err != nil {
				return nil, err
			}

			return returnBigIntOrZeroIfNil(cfg.totalBalancesOnEth), nil
		},
		MintBalancesCalled: func(ctx context.Context, account common.Address) (*big.Int, error) {
			err := cfg.errorsOnCalls["MintBalancesEth"]
			if err != nil {
				return nil, err
			}

			return returnBigIntOrZeroIfNil(cfg.mintBalancesOnEth), nil
		},
		BurnBalancesCalled: func(ctx context.Context, account common.Address) (*big.Int, error) {
			err := cfg.errorsOnCalls["BurnBalancesEth"]
			if err != nil {
				return nil, err
			}

			return returnBigIntOrZeroIfNil(cfg.burnBalancesOnEth), nil
		},
		GetBatchCalled: func(ctx context.Context, nonce uint64) (*bridgeCore.TransferBatch, bool, error) {
			err := cfg.errorsOnCalls["GetBatchEth"]
			if err != nil {
				return nil, false, err
			}

			batch := &bridgeCore.TransferBatch{
				ID: nonce,
			}
			applyDummyFromEthDepositsToBatch(cfg, batch)

			return batch, false, nil
		},
		WasExecutedCalled: func(ctx context.Context, batchID uint64) (bool, error) {
			err := cfg.errorsOnCalls["WasExecutedEth"]
			if err != nil {
				return false, err
			}

			_, found := cfg.amountsOnKlvPendingBatches[batchID]
			return !found, nil
		},
	}

	validator, err := NewBalanceValidator(args)
	if err != nil {
		result.error = err
		return result
	}

	result.error = validator.CheckToken(context.Background(), cfg.ethToken, cfg.kdaToken, cfg.amount, cfg.direction)

	return result
}

func applyDummyFromKlvDepositsToBatch(cfg testConfiguration, batch *bridgeCore.TransferBatch) {
	if cfg.amountsOnKlvPendingBatches != nil {
		values, found := cfg.amountsOnKlvPendingBatches[batch.ID]
		if found {
			depositCounter := uint64(0)

			for _, deposit := range values {
				batch.Deposits = append(batch.Deposits, &bridgeCore.DepositTransfer{
					Nonce:            depositCounter,
					Amount:           big.NewInt(0).Set(deposit),
					SourceTokenBytes: kdaToken,
				})
			}
		}
	}
}

func applyDummyFromEthDepositsToBatch(cfg testConfiguration, batch *bridgeCore.TransferBatch) {
	if cfg.amountsOnEthPendingBatches != nil {
		values, found := cfg.amountsOnEthPendingBatches[batch.ID]
		if found {
			depositCounter := uint64(0)

			for _, deposit := range values {
				batch.Deposits = append(batch.Deposits, &bridgeCore.DepositTransfer{
					Nonce:            depositCounter,
					Amount:           big.NewInt(0).Set(deposit),
					SourceTokenBytes: ethToken.Bytes(),
				})
			}
		}
	}
}

func getMaxKlvPendingBatchID(cfg testConfiguration) uint64 {
	if cfg.amountsOnKlvPendingBatches == nil {
		return 0
	}

	maxBatchIDFound := uint64(0)
	for batchID := range cfg.amountsOnKlvPendingBatches {
		if batchID > maxBatchIDFound {
			maxBatchIDFound = batchID
		}
	}

	return maxBatchIDFound
}

func returnBigIntOrZeroIfNil(value *big.Int) *big.Int {
	if value == nil {
		return big.NewInt(0)
	}

	return big.NewInt(0).Set(value)
}
