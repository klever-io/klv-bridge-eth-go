package ethtokc

const (
	// GettingPendingBatchFromEthereum is the step identifier for fetching the pending batch from the Ethereum chain
	GettingPendingBatchFromEthereum = "get pending batch from Ethereum"

	// ProposingTransferOnKC is the step identifier for proposing transfer on KC
	ProposingTransferOnKC = "propose transfer"

	// SigningProposedTransferOnKC is the step identifier for signing proposed transfer
	SigningProposedTransferOnKC = "sign proposed transfer"

	// WaitingForQuorum is the step identifier for waiting until the quorum is reached
	WaitingForQuorum = "wait for quorum"

	// PerformingActionID is the step identifier for performing the ActionID on KC
	PerformingActionID = "perform action"

	// NumSteps indicates how many steps the state machine for Ethereum -> KC flow has
	NumSteps = 5
)
