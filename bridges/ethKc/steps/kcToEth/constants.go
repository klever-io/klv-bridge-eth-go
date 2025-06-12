package kctoeth

const (
	// GettingPendingBatchFromKc is the step identifier for fetching the pending batch from the Kc
	GettingPendingBatchFromKc = "get pending batch from Kc"

	// SigningProposedTransferOnEthereum is the step identifier for signing proposed transfer
	SigningProposedTransferOnEthereum = "sign proposed transfer"

	// WaitingForQuorumOnTransfer is the step identifier for waiting until the quorum is reached
	WaitingForQuorumOnTransfer = "wait for quorum on transfer"

	// PerformingTransfer is the step identifier for performing the transfer on Ethereum
	PerformingTransfer = "perform transfer"

	// WaitingTransferConfirmation is the step identifier for waiting the transfer confirmation on Ethereum
	WaitingTransferConfirmation = "wait transfer confirmating"

	// ResolvingSetStatusOnKc is the step identifier for resolving set status on Kc
	ResolvingSetStatusOnKc = "resolve set status"

	// ProposingSetStatusOnKc is the step idetifier for proposing set status action on Kc
	ProposingSetStatusOnKc = "propose set status"

	// SigningProposedSetStatusOnKc is the step identifier for signing proposed set status action
	SigningProposedSetStatusOnKc = "sign proposed set status"

	// WaitingForQuorumOnSetStatus is the step identifier for waiting until the quorum is reached
	WaitingForQuorumOnSetStatus = "wait for quorum on set status"

	// PerformingSetStatus is the step identifier for performing the set status action on Kc
	PerformingSetStatus = "perform set status"

	// NumSteps indicates how many steps the state machine for Klever Blockchain -> Ethereum flow has
	NumSteps = 10
)
