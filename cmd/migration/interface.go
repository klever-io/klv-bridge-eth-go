package main

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/klever-io/klv-bridge-eth-go/executors/ethereum"
)

// BatchCreator defines the operations implemented by an entity that can create an Ethereum batch message that can be used
// in signing or transfer execution
type BatchCreator interface {
	CreateBatchInfo(ctx context.Context, newSafeAddress common.Address, partialMigration map[string]*big.Float) (*ethereum.BatchInfo, error)
}
