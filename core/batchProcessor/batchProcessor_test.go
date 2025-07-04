package batchProcessor

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
	"github.com/stretchr/testify/assert"
)

func TestExtractListEthToKlv(t *testing.T) {
	t.Parallel()

	testBatch := &bridgeCore.TransferBatch{
		ID: 37,
		Deposits: []*bridgeCore.DepositTransfer{
			{
				Nonce:                 1,
				ToBytes:               []byte("to 1"),
				FromBytes:             []byte("from 1"),
				SourceTokenBytes:      []byte("source token 1"),
				DestinationTokenBytes: []byte("destination token 1"),
				Amount:                big.NewInt(11),
			},
			{
				Nonce:                 2,
				ToBytes:               []byte("to 2"),
				FromBytes:             []byte("from 2"),
				SourceTokenBytes:      []byte("source token 2"),
				DestinationTokenBytes: []byte("destination token 2"),
				Amount:                big.NewInt(22),
			},
		},
		Statuses: nil,
	}

	args := ExtractListEthToKlv(testBatch)

	expectedEthTokens := []common.Address{
		common.BytesToAddress([]byte("source token 1")),
		common.BytesToAddress([]byte("source token 2")),
	}
	assert.Equal(t, expectedEthTokens, args.EthTokens)

	expectedRecipients := []common.Address{
		common.BytesToAddress([]byte("to 1")),
		common.BytesToAddress([]byte("to 2")),
	}
	assert.Equal(t, expectedRecipients, args.Recipients)

	expectedKdaTokenBytes := [][]byte{
		[]byte("destination token 1"),
		[]byte("destination token 2"),
	}
	assert.Equal(t, expectedKdaTokenBytes, args.KdaTokenBytes)

	expectedAmounts := []*big.Int{
		big.NewInt(11),
		big.NewInt(22),
	}
	assert.Equal(t, expectedAmounts, args.Amounts)

	expectedNonces := []*big.Int{
		big.NewInt(1),
		big.NewInt(2),
	}
	assert.Equal(t, expectedNonces, args.Nonces)
}

func TestExtractListKlvToEth(t *testing.T) {
	t.Parallel()

	testBatch := &bridgeCore.TransferBatch{
		ID: 37,
		Deposits: []*bridgeCore.DepositTransfer{
			{
				Nonce:                 1,
				ToBytes:               []byte("to 1"),
				FromBytes:             []byte("from 1"),
				SourceTokenBytes:      []byte("source token 1"),
				DestinationTokenBytes: []byte("destination token 1"),
				Amount:                big.NewInt(11),
			},
			{
				Nonce:                 2,
				ToBytes:               []byte("to 2"),
				FromBytes:             []byte("from 2"),
				SourceTokenBytes:      []byte("source token 2"),
				DestinationTokenBytes: []byte("destination token 2"),
				Amount:                big.NewInt(22),
			},
		},
		Statuses: nil,
	}

	args := ExtractListKlvToEth(testBatch)

	expectedEthTokens := []common.Address{
		common.BytesToAddress([]byte("destination token 1")),
		common.BytesToAddress([]byte("destination token 2")),
	}
	assert.Equal(t, expectedEthTokens, args.EthTokens)

	expectedRecipients := []common.Address{
		common.BytesToAddress([]byte("to 1")),
		common.BytesToAddress([]byte("to 2")),
	}
	assert.Equal(t, expectedRecipients, args.Recipients)

	expectedKdaTokenBytes := [][]byte{
		[]byte("source token 1"),
		[]byte("source token 2"),
	}
	assert.Equal(t, expectedKdaTokenBytes, args.KdaTokenBytes)

	expectedAmounts := []*big.Int{
		big.NewInt(11),
		big.NewInt(22),
	}
	assert.Equal(t, expectedAmounts, args.Amounts)

	expectedNonces := []*big.Int{
		big.NewInt(1),
		big.NewInt(2),
	}
	assert.Equal(t, expectedNonces, args.Nonces)
}
