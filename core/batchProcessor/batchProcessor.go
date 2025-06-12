package batchProcessor

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	bridgeCore "github.com/klever-io/klv-bridge-eth-go/core"
)

// Direction is the direction of the transfer
type Direction string

const (
	// FromKc is the direction of the transfer
	FromKc Direction = "FromKc"
	// ToKc is the direction of the transfer
	ToKc Direction = "ToKc"
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
// The transfer is from Klever Blockchain to Ethereum
func ExtractListKlvToEth(batch *bridgeCore.TransferBatch) *ArgListsBatch {
	arg := &ArgListsBatch{
		Direction: FromKc,
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
// The transfer is from Ehtereum to Klever Blockchain
func ExtractListEthToKlv(batch *bridgeCore.TransferBatch) *ArgListsBatch {
	arg := &ArgListsBatch{
		Direction: ToKc,
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
