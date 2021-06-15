package bridge

import (
	"context"
)

type Broadcaster interface {
	Signatures() [][]byte
	SendSignature(signature []byte)
}

type Mapper interface {
	GetTokenId(string) string
	GetErc20Address(string) string
}

type Bridge interface {
	GetPending(context.Context) *Batch
	ProposeSetStatus(context.Context, *Batch)
	ProposeTransfer(context.Context, *Batch) (string, error)
	WasProposedTransfer(context.Context, BatchId) bool
	GetActionIdForProposeTransfer(context.Context, BatchId) ActionId
	WasProposedSetStatus(context.Context, *Batch) bool
	GetActionIdForSetStatusOnPendingTransfer(context.Context) ActionId
	WasExecuted(context.Context, ActionId, BatchId) bool
	Sign(context.Context, ActionId) (string, error)
	Execute(context.Context, ActionId, BatchId) (string, error)
	SignersCount(context.Context, ActionId) uint
}
