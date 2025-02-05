package builders

import (
	"math/big"

	"github.com/klever-io/klv-bridge-eth-go/clients/klever/blockchain/address"
	"github.com/klever-io/klv-bridge-eth-go/clients/klever/proxy/models"
)

// TxDataBuilder defines the behavior of a transaction data builder
type TxDataBuilder interface {
	Function(function string) TxDataBuilder

	ArgHexString(hexed string) TxDataBuilder
	ArgAddress(address address.Address) TxDataBuilder
	ArgBigInt(value *big.Int) TxDataBuilder
	ArgInt64(value int64) TxDataBuilder
	ArgBytes(bytes []byte) TxDataBuilder
	ArgBytesList(list [][]byte) TxDataBuilder

	ToDataString() (string, error)
	ToDataBytes() ([]byte, error)

	IsInterfaceNil() bool
}

// VMQueryBuilder defines the behavior of a vm query builder
type VMQueryBuilder interface {
	Function(function string) VMQueryBuilder
	CallerAddress(address address.Address) VMQueryBuilder
	Address(address address.Address) VMQueryBuilder

	ArgHexString(hexed string) VMQueryBuilder
	ArgAddress(address address.Address) VMQueryBuilder
	ArgBigInt(value *big.Int) VMQueryBuilder
	ArgInt64(value int64) VMQueryBuilder
	ArgBytes(bytes []byte) VMQueryBuilder

	ToVmValueRequest() (*models.VmValueRequest, error)

	IsInterfaceNil() bool
}
