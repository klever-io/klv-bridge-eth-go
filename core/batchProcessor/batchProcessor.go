package batchProcessor

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
)

// Direction is the direction of the transfer
type Direction string

const (
	// FromKleverchain is the direction of the transfer
	FromKleverchain Direction = "FromKleverchain"
	// ToKleverchain is the direction of the transfer
	ToKleverchain Direction = "ToKleverchain"
)

// ArgListsBatch is a struct that contains the batch data in a format that is easy to use
type ArgListsBatch struct {
	EthTokens     []common.Address
	Recipients    []common.Address
	KdaTokenBytes [][]byte
	Amounts       []*big.Int
	Nonces        []*big.Int
	Direction     Direction
}

// ExtractListKlvToEth will extract the batch data into a format that is easy to use
// The transfer is from Kleverchain to Ethereum
func ExtractListKlvToEth(batch *bridgeCore.TransferBatch) *ArgListsBatch {
	arg := &ArgListsBatch{
		Direction: FromKleverchain,
	}

	for _, dt := range batch.Deposits {
		recipient := common.BytesToAddress(dt.ToBytes)
		arg.Recipients = append(arg.Recipients, recipient)

		token := common.BytesToAddress(dt.DestinationTokenBytes)
		arg.EthTokens = append(arg.EthTokens, token)

		amount := big.NewInt(0).Set(dt.Amount)
		arg.Amounts = append(arg.Amounts, amount)

		nonce := big.NewInt(0).SetUint64(dt.Nonce)
		arg.Nonces = append(arg.Nonces, nonce)

		arg.KdaTokenBytes = append(arg.KdaTokenBytes, dt.SourceTokenBytes)
	}

	return arg
}

// ExtractListEthToKlv will extract the batch data into a format that is easy to use
// The transfer is from Ehtereum to Kleverchain
func ExtractListEthToKlv(batch *bridgeCore.TransferBatch) *ArgListsBatch {
	arg := &ArgListsBatch{
		Direction: ToKleverchain,
	}

	for _, dt := range batch.Deposits {
		recipient := common.BytesToAddress(dt.ToBytes)
		arg.Recipients = append(arg.Recipients, recipient)

		token := common.BytesToAddress(dt.SourceTokenBytes)
		arg.EthTokens = append(arg.EthTokens, token)

		amount := big.NewInt(0).Set(dt.Amount)
		arg.Amounts = append(arg.Amounts, amount)

		nonce := big.NewInt(0).SetUint64(dt.Nonce)
		arg.Nonces = append(arg.Nonces, nonce)

		arg.KdaTokenBytes = append(arg.KdaTokenBytes, dt.DestinationTokenBytes)
	}

	return arg
}
